package main

import (
	"context"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/eteu-technologies/near-api-go/client"
	"github.com/eteu-technologies/near-api-go/types"
	"github.com/eteu-technologies/near-api-go/types/action"
	"github.com/eteu-technologies/near-api-go/types/key"
	"github.com/eteu-technologies/near-api-go/types/transaction"
)

var (
	accID       = "mikroskeem.testnet"
	secretKey   = os.Getenv("NEAR_PRIV_KEY")
	targetAccID = "dev-1621263077598-74843909627468"
)

func main() {
	keyPair, err := key.NewBase58KeyPair(secretKey)
	if err != nil {
		log.Fatal("failed to load private key: ", err)
	}

	addr := "https://rpc.testnet.near.org"

	rpc, err := client.NewClient(addr)
	if err != nil {
		log.Fatal("failed to create rpc client: ", err)
	}

	log.Printf("near network: %s", rpc.NetworkAddr())

	// Create a transaction
	txn := transaction.Transaction{
		SignerID:   accID,
		ReceiverID: targetAccID,
		Actions: []action.Action{
			action.NewFunctionCall("increment", nil, types.DefaultFunctionCallGas, types.NEARToYocto(0)),
		},
	}

	res, err := rpc.TransactionSendAwait(context.Background(), txn, client.WithLatestBlock(), client.WithKeyPair(keyPair))
	if err != nil {
		log.Fatal("failed to do txn: ", err)
	}

	spew.Dump(res)
}
