package main

import (
	"context"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/client/v2/common"
	"github.com/algorand/go-algorand-sdk/crypto"
	transaction "github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/mnemonic"
)

//GulliCoin example
// based on purestake's submit-a-txn example

const algodAddress = "https://testnet-algorand.api.purestake.io/ps2"
const psToken = "..." //apikey

// Initalize throw-away account for this example - check that is has funds before running the program.
const mn = "..." // enter your 25 word phrase
// Initalize throw-away account for this example - check that is has funds before running the program.
const fromAddr = "..." // enter your address

// Function from Algorand Inc. - utility for waiting on a transaction confirmation
func waitForConfirmation(txID string, client *algod.Client) {
	status, err := client.Status().Do(context.Background())
	if err != nil {
		fmt.Printf("error getting algod status: %s\n", err)
		return
	}
	lastRound := status.LastRound
	for {
		pt, _, err := client.PendingTransactionInformation(txID).Do(context.Background())
		if err != nil {
			fmt.Printf("error getting pending transaction: %s\n", err)
			return
		}
		if pt.ConfirmedRound > 0 {
			fmt.Printf("Transaction "+txID+" confirmed in round %d\n", pt.ConfirmedRound)
			break
		}
		fmt.Printf("Waiting for confirmation...\n")
		lastRound++
		status, err = client.StatusAfterBlock(lastRound).Do(context.Background())
	}
}

func main() {
	commonClient, err := common.MakeClient(algodAddress, "X-API-Key", psToken)
	if err != nil {
		fmt.Printf("failed to make common client: %s\n", err)
		return
	}
	algodClient := (*algod.Client)(commonClient)
	fmt.Println("Algod client created")

	// Recover private key from the mnemonic
	fromAddrPvtKey, err := mnemonic.ToPrivateKey(mn)
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return
	}
	fmt.Println("Private key recovered from mnemonic")

	// Get the suggested transaction parameters
	txParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		fmt.Printf("error getting suggested tx params: %s\n", err)
		return
	}

	// Make transaction
	note := []byte(nil)
	coinTotalIssuance := uint64(1000000)
	coinDecimalsForDisplay := uint32(0) // i.e. 1 accounting unit in a transfer == 1 coin; we are not working in microCoins or anything
	accountsAreDefaultFrozen := false // if you have this coin, you can transact, the freeze manager doesn't need to unfreeze you first
	managerAddress := fromAddr // the account issuing this is also the account in charge of managing this
	assetReserveAddress := "" // there is no asset reserve (the reserve is for informational purposes only anyways)
	addressWithFreezingPrivileges := fromAddr // this account can blacklist others from receiving or sending assets, freezing their account
	addressWithClawbackPrivileges := fromAddr // this account is allowed to clawback coins from others
	assetUnitName := "gullies"
	assetName := "Gullicoin"
	assetUrl := "github.com/gunnlaugurcalvi"
	assetMetadataHash := "" // I am not making any hash commitments related to this, it's just a fun coin
	txn, err := transaction.MakeAssetCreateTxn(fromAddr, note, txParams, coinTotalIssuance, coinDecimalsForDisplay, accountsAreDefaultFrozen, managerAddress, assetReserveAddress, addressWithFreezingPrivileges, addressWithClawbackPrivileges, assetUnitName, assetName, assetUrl, assetMetadataHash)
	if err != nil {
		fmt.Printf("Error creating transaction: %s\n", err)
		return
	}

	// Sign the Transaction
	_, bytes, err := crypto.SignTransaction(fromAddrPvtKey, txn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}

	// Broadcast the transaction to the network
	sendResponse, err := algodClient.SendRawTransaction(bytes).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return
	}
	fmt.Printf("Transaction successfull with ID: %s\n", sendResponse)
	fmt.Printf("Waiting for confirmation...\n")
	waitForConfirmation(sendResponse, algodClient)
}