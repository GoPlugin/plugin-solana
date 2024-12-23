package client

import (
	"fmt"
	"time"
)

// verifyLoop may only be triggered once, on Start, if initial chain ID check
// fails.
//
// It will continue checking until success and then exit permanently.
func (s *sendOnlyNode[CHAIN_ID, RPC]) verifyLoop() {
	defer s.wg.Done()
	ctx, cancel := s.chStop.NewCtx()
	defer cancel()

	backoff := NewRedialBackoff()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff.Duration()):
		}
		chainID, err := s.rpc.ChainID(ctx)
		if err != nil {
			ok := s.IfStarted(func() {
				if changed := s.setState(NodeStateUnreachable); changed {
					promPoolRPCNodeTransitionsToUnreachable.WithLabelValues(s.chainID.String(), s.name).Inc()
				}
			})
			if !ok {
				return
			}
			s.log.Errorw(fmt.Sprintf("Verify failed: %v", err), "err", err)
			continue
		} else if chainID.String() != s.chainID.String() {
			ok := s.IfStarted(func() {
				if changed := s.setState(NodeStateInvalidChainID); changed {
					promPoolRPCNodeTransitionsToInvalidChainID.WithLabelValues(s.chainID.String(), s.name).Inc()
				}
			})
			if !ok {
				return
			}
			s.log.Errorf(
				"sendonly rpc ChainID doesn't match local chain ID: RPC ID=%s, local ID=%s, node name=%s",
				chainID.String(),
				s.chainID.String(),
				s.name,
			)

			continue
		}
		ok := s.IfStarted(func() {
			if changed := s.setState(NodeStateAlive); changed {
				promPoolRPCNodeTransitionsToAlive.WithLabelValues(s.chainID.String(), s.name).Inc()
			}
		})
		if !ok {
			return
		}
		s.log.Infow("Sendonly RPC Node is online", "NodeState", s.state)
		return
	}
}
