package client

import (
	"context"
	"errors"

	"github.com/eteu-technologies/near-api-go/client/block"
	"github.com/eteu-technologies/near-api-go/jsonrpc"
	"github.com/eteu-technologies/near-api-go/types"
	"github.com/eteu-technologies/near-api-go/types/hash"
	"github.com/eteu-technologies/near-api-go/types/key"
	"github.com/eteu-technologies/near-api-go/types/transaction"
)

type transactionCtx struct {
	txn         transaction.Transaction
	keyPair     *key.KeyPair
	keyNonceSet bool
}

type TransactionOpt func(context.Context, *transactionCtx) error

func (c *Client) prepareTransaction(ctx context.Context, txn transaction.Transaction, txnOpts ...TransactionOpt) (ctx2 context.Context, blob string, err error) {
	ctx2 = context.WithValue(ctx, clientCtx, c)

	txn2 := txn // Make a copy, need to mutate the txn
	txnCtx := transactionCtx{
		txn:         txn2,
		keyPair:     getKeyPair(ctx2),
		keyNonceSet: false,
	}

	for _, opt := range txnOpts {
		if err = opt(ctx2, &txnCtx); err != nil {
			return
		}
	}

	if txnCtx.keyPair == nil {
		err = errors.New("no keypair specified")
		return
	}

	txnCtx.txn.PublicKey = txnCtx.keyPair.PublicKey.ToPublicKey()

	// Query the access key nonce, if not specified
	if !txnCtx.keyNonceSet {
		var accessKey AccessKeyView
		accessKey, err = c.AccessKeyView(ctx2, txnCtx.txn.SignerID, *&txnCtx.keyPair.PublicKey, block.FinalityFinal())
		if err != nil {
			return
		}

		nonce := accessKey.Nonce

		// Increment nonce by 1
		txnCtx.txn.Nonce = nonce + 1
		txnCtx.keyNonceSet = true
	}

	blob, err = transaction.SignAndSerializeTransaction(*txnCtx.keyPair, txnCtx.txn)
	return
}

// TODO: decode response
// https://docs.near.org/docs/develop/front-end/rpc#send-transaction-async
func (c *Client) TransactionSend(ctx context.Context, txn transaction.Transaction, txnOpts ...TransactionOpt) (res jsonrpc.JSONRPCResponse, err error) {
	ctx2, blob, err := c.prepareTransaction(ctx, txn, txnOpts...)
	if err != nil {
		return
	}
	return c.RPCTransactionSend(ctx2, blob)
}

// TODO: decode response
// https://docs.near.org/docs/develop/front-end/rpc#send-transaction-await
func (c *Client) TransactionSendAwait(ctx context.Context, txn transaction.Transaction, txnOpts ...TransactionOpt) (res jsonrpc.JSONRPCResponse, err error) {
	ctx2, blob, err := c.prepareTransaction(ctx, txn, txnOpts...)
	if err != nil {
		return
	}
	return c.RPCTransactionSendAwait(ctx2, blob)
}

func WithBlockCharacteristic(block block.BlockCharacteristic) TransactionOpt {
	return func(ctx context.Context, txnCtx *transactionCtx) (err error) {
		client := ctx.Value(clientCtx).(*Client)

		var res BlockView
		if res, err = client.BlockDetails(ctx, block); err != nil {
			return
		}

		txnCtx.txn.BlockHash = res.Header.Hash
		return
	}

}

// WithBlockHash sets block hash to attach this transaction to
func WithBlockHash(hash hash.CryptoHash) TransactionOpt {
	return func(_ context.Context, txnCtx *transactionCtx) (err error) {
		txnCtx.txn.BlockHash = hash
		return
	}
}

// WithLatestBlock is alias to `WithBlockCharacteristic(block.FinalityFinal())`
func WithLatestBlock() TransactionOpt {
	return WithBlockCharacteristic(block.FinalityFinal())
}

// WithKeyPair sets key pair to use sign this transaction with
func WithKeyPair(keyPair key.KeyPair) TransactionOpt {
	return func(_ context.Context, txnCtx *transactionCtx) (err error) {
		kp := keyPair
		txnCtx.keyPair = &kp
		return
	}
}

// WithKeyNonce sets key nonce to use with this transaction. If not set via this function, a RPC query will be done to query current nonce and
// (nonce+1) will be used
func WithKeyNonce(nonce types.Nonce) TransactionOpt {
	return func(_ context.Context, txnCtx *transactionCtx) (err error) {
		txnCtx.txn.Nonce = nonce
		txnCtx.keyNonceSet = true
		return
	}
}
