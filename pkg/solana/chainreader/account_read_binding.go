package chainreader

import (
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/goplugin/plugin-common/pkg/types"
)

// BinaryDataReader provides an interface for reading bytes from a source. This is likely a wrapper
// for a solana client.
type BinaryDataReader interface {
	ReadAll(context.Context, solana.PublicKey, *rpc.GetAccountInfoOpts) ([]byte, error)
}

// accountReadBinding provides decoding and reading Solana Account data using a defined codec. The
// `idlAccount` refers to the account name in the IDL for which the codec has a type mapping.
type accountReadBinding struct {
	idlAccount string
	codec      types.RemoteCodec
	reader     BinaryDataReader
	opts       *rpc.GetAccountInfoOpts
}

func newAccountReadBinding(acct string, codec types.RemoteCodec, reader BinaryDataReader, opts *rpc.GetAccountInfoOpts) *accountReadBinding {
	return &accountReadBinding{
		idlAccount: acct,
		codec:      codec,
		reader:     reader,
		opts:       opts,
	}
}

var _ readBinding = &accountReadBinding{}

func (b *accountReadBinding) PreLoad(ctx context.Context, address string, result *loadedResult) {
	if result == nil {
		return
	}

	account, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		result.err <- err

		return
	}

	bts, err := b.reader.ReadAll(ctx, account, b.opts)
	if err != nil {
		result.err <- fmt.Errorf("%w: failed to get binary data", err)

		return
	}

	select {
	case <-ctx.Done():
		result.err <- ctx.Err()
	default:
		result.value <- bts
	}
}

func (b *accountReadBinding) GetLatestValue(ctx context.Context, address string, _ any, outVal any, result *loadedResult) error {
	var (
		bts []byte
		err error
	)

	if result != nil {
		// when preloading, the process will wait for one of three conditions:
		// 1. the context ends and returns an error
		// 2. bytes were loaded in the bytes channel
		// 3. an error was loaded in the err channel
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case bts = <-result.value:
		case err = <-result.err:
		}

		if err != nil {
			return err
		}
	} else {
		account, err := solana.PublicKeyFromBase58(address)
		if err != nil {
			return err
		}

		if bts, err = b.reader.ReadAll(ctx, account, b.opts); err != nil {
			return fmt.Errorf("%w: failed to get binary data", err)
		}
	}

	return b.codec.Decode(ctx, bts, outVal, b.idlAccount)
}

func (b *accountReadBinding) CreateType(_ bool) (any, error) {
	return b.codec.CreateType(b.idlAccount, false)
}
