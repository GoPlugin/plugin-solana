package fees

import (
	"context"
	"fmt"
	"sync"

	"github.com/goplugin/plugin-common/pkg/logger"
	"github.com/goplugin/plugin-common/pkg/services"
	"github.com/goplugin/plugin-common/pkg/utils"
	"github.com/goplugin/plugin-common/pkg/utils/mathutil"

	"github.com/goplugin/plugin-solana/pkg/solana/client"
	"github.com/goplugin/plugin-solana/pkg/solana/config"
)

var _ Estimator = &blockHistoryEstimator{}

type blockHistoryEstimator struct {
	starter services.StateMachine
	chStop  services.StopChan
	done    sync.WaitGroup

	client *utils.LazyLoad[client.ReaderWriter]
	cfg    config.Config
	lgr    logger.Logger

	price uint64
	lock  sync.RWMutex
}

// NewBlockHistoryEstimator creates a new fee estimator that parses historical fees from a fetched block
// Note: getRecentPrioritizationFees is not used because it provides the lowest prioritization fee for an included tx in the block
// which is not effective enough for increasing the chances of block inclusion
func NewBlockHistoryEstimator(c *utils.LazyLoad[client.ReaderWriter], cfg config.Config, lgr logger.Logger) (*blockHistoryEstimator, error) {
	return &blockHistoryEstimator{
		chStop: make(chan struct{}),
		client: c,
		cfg:    cfg,
		lgr:    lgr,
		price:  cfg.ComputeUnitPriceDefault(), // use default value
	}, nil
}

func (bhe *blockHistoryEstimator) Start(ctx context.Context) error {
	return bhe.starter.StartOnce("solana_blockHistoryEstimator", func() error {
		bhe.done.Add(1)
		go bhe.run()
		bhe.lgr.Debugw("BlockHistoryEstimator: started")
		return nil
	})
}

func (bhe *blockHistoryEstimator) run() {
	defer bhe.done.Done()
	ctx, cancel := bhe.chStop.NewCtx()
	defer cancel()

	ticker := services.NewTicker(bhe.cfg.BlockHistoryPollPeriod())
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := bhe.calculatePrice(ctx); err != nil {
				bhe.lgr.Error(fmt.Errorf("BlockHistoryEstimator failed to fetch price: %w", err))
			}
		}
	}
}

func (bhe *blockHistoryEstimator) Close() error {
	close(bhe.chStop)
	bhe.done.Wait()
	bhe.lgr.Debugw("BlockHistoryEstimator: stopped")
	return nil
}

func (bhe *blockHistoryEstimator) BaseComputeUnitPrice() uint64 {
	price := bhe.readRawPrice()
	if price >= bhe.cfg.ComputeUnitPriceMin() && price <= bhe.cfg.ComputeUnitPriceMax() {
		return price
	}

	if price < bhe.cfg.ComputeUnitPriceMin() {
		bhe.lgr.Warnw("BlockHistoryEstimator: estimation below minimum consider lowering ComputeUnitPriceMin", "min", bhe.cfg.ComputeUnitPriceMin(), "calculated", price)
		return bhe.cfg.ComputeUnitPriceMin()
	}

	bhe.lgr.Warnw("BlockHistoryEstimator: estimation above maximum consider increasing ComputeUnitPriceMax", "min", bhe.cfg.ComputeUnitPriceMax(), "calculated", price)
	return bhe.cfg.ComputeUnitPriceMax()
}

func (bhe *blockHistoryEstimator) readRawPrice() uint64 {
	bhe.lock.RLock()
	defer bhe.lock.RUnlock()
	return bhe.price
}

func (bhe *blockHistoryEstimator) calculatePrice(ctx context.Context) error {
	// fetch client
	c, err := bhe.client.Get()
	if err != nil {
		return fmt.Errorf("failed to get client in blockHistoryEstimator.getFee: %w", err)
	}

	// get latest block based on configured confirmation
	block, err := c.GetLatestBlock(ctx)
	if err != nil {
		return fmt.Errorf("failed to get block in blockHistoryEstimator.getFee: %w", err)
	}

	// parse block for fee data
	feeData, err := ParseBlock(block)
	if err != nil {
		return fmt.Errorf("failed to parse block in blockHistoryEstimator.getFee: %w", err)
	}

	// take median of returned fee values
	v, err := mathutil.Median(feeData.Prices...)
	if err != nil {
		return fmt.Errorf("failed to find median in blockHistoryEstimator.getFee: %w", err)
	}

	// set data
	bhe.lock.Lock()
	bhe.price = uint64(v) // ComputeUnitPrice is uint64 underneath
	bhe.lock.Unlock()
	bhe.lgr.Debugw("BlockHistoryEstimator: updated",
		"computeUnitPrice", v,
		"blockhash", block.Blockhash,
		"slot", block.ParentSlot+1,
		"count", len(feeData.Prices),
	)
	return nil
}
