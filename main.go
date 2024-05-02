package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Layr-Labs/avs-sync/avssync"
	"github.com/Layr-Labs/eigensdk-go/aws/secretmanager"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/fireblocks"
	walletsdk "github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
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
	loggerConfig, err := ReadLoggerCLIConfig(cliCtx)
	if err != nil {
		return err
	}
	logger, err := NewLogger(*loggerConfig)
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
	if cliCtx.Bool(UseFireblocksFlag.Name) {
		var apiKey string
		var secretKey []byte
		var err error

		region := cliCtx.String(SecretManagerRegionFlag.Name)
		if len(region) >= 0 {
			logger.Info("Using secret manager to read fireblocks api key and secret")
			smFireblocksAPIKeyName := cliCtx.String(SecretManagerFireblocksAPIKeyNameFlag.Name)
			smFireblockAPISecretName := cliCtx.String(SecretManagerFireblocksAPISecretNameFlag.Name)
			if len(smFireblocksAPIKeyName) > 0 && len(smFireblockAPISecretName) > 0 {
				apiKey, err = secretmanager.ReadStringFromSecretManager(context.Background(), smFireblocksAPIKeyName, region)
				if err != nil {
					return fmt.Errorf("Cannot read fireblocks api key from secret manager: %w", err)
				}
				secretKeyStr, err := secretmanager.ReadStringFromSecretManager(context.Background(), smFireblockAPISecretName, region)
				if err != nil {
					return fmt.Errorf("Cannot read fireblocks secret from secret manager: %w", err)
				}
				secretKey = []byte(secretKeyStr)
			}
		}

		// If the secret manager values are not set, try to read from flags
		if len(apiKey) == 0 || len(secretKey) == 0 {
			logger.Info("Reading fireblocks api key and secret from flags")
			apiKey = cliCtx.String(FireblocksAPIKeyFlag.Name)
			secretPath := cliCtx.String(FireblocksAPISecretPathFlag.Name)
			secretKey, err = os.ReadFile(secretPath)
			if err != nil {
				return fmt.Errorf("Cannot read fireblocks secret from %s: %w", secretPath, err)
			}
		}

		fbBaseURL := cliCtx.String(FireblocksBaseURLFlag.Name)
		fbVaultAccountName := cliCtx.String(FireblocksVaultAccountNameFlag.Name)
		if apiKey == "" {
			return errors.New("Fireblocks API key is not set")
		}
		if len(secretKey) == 0 {
			return errors.New("Fireblocks API secret is not set")
		}
		if fbBaseURL == "" {
			return errors.New("Fireblocks base URL is not set")
		}
		if fbVaultAccountName == "" {
			return errors.New("Fireblocks vault account name is not set")
		}

		fireblocksClient, err := fireblocks.NewClient(
			apiKey,
			secretKey,
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
	} else {
		logger.Info("Using ecdsa private key to create wallet")
		var ecdsaPrivKey *ecdsa.PrivateKey
		smOperatorEcdsaPrivKeyHexStr := cliCtx.String(SecretManagerEcdsaPrivateKeyNameFlag.Name)
		if len(smOperatorEcdsaPrivKeyHexStr) > 0 {
			ecdsaPrivKey, err = crypto.HexToECDSA(smOperatorEcdsaPrivKeyHexStr)
			if err != nil {
				return fmt.Errorf("Cannot create ecdsa private key: %w", err)
			}
		} else {
			operatorEcdsaPrivKeyHexStr := cliCtx.String(EcdsaPrivateKeyFlag.Name)
			ecdsaPrivKey, err = crypto.HexToECDSA(operatorEcdsaPrivKeyHexStr)
			if err != nil {
				return fmt.Errorf("Cannot create ecdsa private key: %w", err)
			}
		}
		signerV2, address, err := signerv2.SignerFromConfig(signerv2.Config{PrivateKey: ecdsaPrivKey}, chainid)
		if err != nil {
			return err
		}
		wallet, err = walletsdk.NewPrivateKeyWallet(ethHttpClient, signerV2, address, logger)
		if err != nil {
			return err
		}
	}

	sender, err := wallet.SenderAddress(context.Background())
	if err != nil {
		return fmt.Errorf("Cannot get sender address: %w", err)
	}
	logger.Infof("Sender address: %s", sender.Hex())
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
	avsSync := avssync.NewAvsSync(
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
		cliCtx.String(MetricsAddrFlag.Name),
	)

	avsSync.Start(context.Background())
	return nil
}
