package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
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

	operatorEcdsaPrivKeyHexStr := cliCtx.String(EcdsaPrivateKeyFlag.Name)
	ecdsaPrivKey, err := crypto.HexToECDSA(operatorEcdsaPrivKeyHexStr)
	if err != nil {
		logger.Errorf("Cannot create ecdsa private key", "err", err)
		return err
	}
	ecdsaAddr := crypto.PubkeyToAddress(ecdsaPrivKey.PublicKey)
	logger.Debug("ECDSA Address", "address", ecdsaAddr.Hex())

	ethHttpClient, err := eth.NewClient(cliCtx.String(EthHttpUrlFlag.Name))
	if err != nil {
		logger.Fatalf("Cannot create eth client", "err", err)
	}

	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	chainid, err := ethHttpClient.ChainID(rpcCtx)
	if err != nil {
		logger.Fatalf("Cannot get chain id", "err", err)
	}
	// confusing interface, see https://github.com/Layr-Labs/eigensdk-go/issues/90
	signerFn, _, err := signerv2.SignerFromConfig(signerv2.Config{
		PrivateKey: ecdsaPrivKey,
	}, chainid)
	if err != nil {
		logger.Errorf("Cannot create signer", "err", err)
		return err
	}

	txMgr := txmgr.NewSimpleTxManager(ethHttpClient, logger, signerFn, ecdsaAddr)

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
	avsSync := NewAvsSync(
		logger,
		avsReader,
		avsWriter,
		cliCtx.Duration(SleepBeforeFirstSyncDurationFlag.Name),
		cliCtx.Duration(SyncIntervalFlag.Name),
		operators,
		quorums,
		cliCtx.Bool(FetchQuorumDynamicallyFlag.Name),
		cliCtx.Int(retrySyncNTimes.Name),
		cliCtx.Duration(ReaderTimeoutDurationFlag.Name),
		cliCtx.Duration(WriterTimeoutDurationFlag.Name),
	)
	avsSync.Start()

	return nil
}
