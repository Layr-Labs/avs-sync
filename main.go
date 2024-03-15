package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	walletsdk "github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Flags = Flags
	app.Name = "AvsSync"
	app.Usage = "Updates stakes of operators"
	app.Description = "Service that runs a cron job which updates the stakes of the specified operators for the specified AVS' stake registry"

	app.Action = avsSyncMain

	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Application failed. Message:", err)
	}
}

func avsSyncMain(cliCtx *cli.Context) error {
	log.Println("Registering Node")
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		return err
	}

	writerTimeout := cliCtx.Duration(WriterTimeoutDurationFlag.Name)
	readerTimeout := cliCtx.Duration(ReaderTimeoutDurationFlag.Name)

	ethHttpClient, err := eth.NewClient(cliCtx.String(EthHttpUrlFlag.Name))
	if err != nil {
		logger.Fatalf("Cannot create eth client", "err", err)
	}

	rpcCtx, cancel := context.WithTimeout(context.Background(), readerTimeout)
	defer cancel()
	chainid, err := ethHttpClient.ChainID(rpcCtx)
	if err != nil {
		logger.Fatalf("Cannot get chain id", "err", err)
	}

	var wallet walletsdk.Wallet
	var sender common.Address
	operatorEcdsaPrivKeyHexStr := cliCtx.String(EcdsaPrivateKeyFlag.Name)
	if len(operatorEcdsaPrivKeyHexStr) > 0 {
		ecdsaPrivKey, err := crypto.HexToECDSA(operatorEcdsaPrivKeyHexStr)
		if err != nil {
			return fmt.Errorf("Cannot create ecdsa private key: %w", err)
		}

		signerV2, address, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivKey}, chainid)
		if err != nil {
			return err
		}
		wallet, err = walletsdk.NewPrivateKeyWallet(ethHttpClient, signerV2, address, logger)
		if err != nil {
			return err
		}
		sender = address
	} else {
		fbAPIKey := cliCtx.String(FireblocksAPIKeyFlag.Name)
		fbSecret := cliCtx.String(FireblocksAPISecretFlag.Name)
		fbBaseURL := cliCtx.String(FireblocksBaseURLFlag.Name)
		fbVaultAccountName := cliCtx.String(FireblocksVaultAccountNameFlag.Name)
		if fbAPIKey == "" {
			return errors.New("Fireblocks API key is not set")
		}
		if fbSecret == "" {
			return errors.New("Fireblocks API secret is not set")
		}
		if fbBaseURL == "" {
			return errors.New("Fireblocks base URL is not set")
		}
		if fbVaultAccountName == "" {
			return errors.New("Fireblocks vault account name is not set")
		}

		fireblocksClient, err := fireblocks.NewClient(
			fbAPIKey,
			[]byte(fbSecret),
			fbBaseURL,
			writerTimeout,
			logger,
		)
		if err != nil {
			return err
		}
		wallet, err = walletsdk.NewFireblocksWallet(fireblocksClient, ethHttpClient, fbVaultAccountName, logger)
		if err != nil {
			return err
		}
		// TODO: read this from wallet
		// sender, err = wallet.SenderAddress()
		sender = common.HexToAddress("0x0000000000000000000000000000000000000123")
	}

	txMgr := txmgr.NewSimpleTxManager(wallet, ethHttpClient, logger, sender)

	avsWriter, err := avsregistry.BuildAvsRegistryChainWriter(
		common.HexToAddress(cliCtx.String(RegistryCoordinatorAddrFlag.Name)),
		common.HexToAddress(cliCtx.String(OperatorStateRetrieverAddrFlag.Name)),
		logger,
		ethHttpClient,
		txMgr,
	)
	if err != nil {
		logger.Fatalf("Cannot create avs writer", "err", err)
	}
	avsReader, err := avsregistry.BuildAvsRegistryChainReader(
		common.HexToAddress(cliCtx.String(RegistryCoordinatorAddrFlag.Name)),
		common.HexToAddress(cliCtx.String(OperatorStateRetrieverAddrFlag.Name)),
		ethHttpClient,
		logger,
	)
	if err != nil {
		logger.Fatalf("Cannot create avs reader", "err", err)
	}

	operatorsList := cliCtx.StringSlice(OperatorListFlag.Name)
	var operators []common.Address
	for _, operator := range operatorsList {
		operators = append(operators, common.HexToAddress(operator))
	}
	var quorums []byte
	for _, quorum := range cliCtx.IntSlice(QuorumListFlag.Name) {
		quorums = append(quorums, byte(quorum))
	}

	firstSyncTimeStr := cliCtx.String(FirstSyncTimeFlag.Name)
	var sleepBeforeFirstSyncDuration time.Duration
	if firstSyncTimeStr == "" {
		sleepBeforeFirstSyncDuration = 0 * time.Second
	} else {
		now := time.Now()
		firstSyncTime, err := time.Parse(time.TimeOnly, firstSyncTimeStr)
		firstSyncTime = time.Date(now.Year(), now.Month(), now.Day(), firstSyncTime.Hour(), firstSyncTime.Minute(), firstSyncTime.Second(), 0, now.Location())
		if err != nil {
			return err
		}
		if now.After(firstSyncTime) {
			// If the set time is before the current time, add a day to the set time
			firstSyncTime = firstSyncTime.Add(24 * time.Hour)
		}
		sleepBeforeFirstSyncDuration = firstSyncTime.Sub(now)
	}
	logger.Infof("Sleeping for %v before first sync, so that it happens at %v", sleepBeforeFirstSyncDuration, time.Now().Add(sleepBeforeFirstSyncDuration))
	avsSync := NewAvsSync(
		logger,
		avsReader,
		avsWriter,
		sleepBeforeFirstSyncDuration,
		cliCtx.Duration(SyncIntervalFlag.Name),
		operators,
		quorums,
		cliCtx.Bool(FetchQuorumDynamicallyFlag.Name),
		cliCtx.Int(retrySyncNTimes.Name),
		readerTimeout,
		writerTimeout,
	)

	avsSync.Start()
	return nil
}
