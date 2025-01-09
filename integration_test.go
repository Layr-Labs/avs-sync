package main

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	walletsdk "github.com/Layr-Labs/eigensdk-go/chainio/clients/wallet"
	"github.com/Layr-Labs/eigensdk-go/chainio/txmgr"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	rpccalls "github.com/Layr-Labs/eigensdk-go/metrics/collectors/rpc_calls"
	"github.com/Layr-Labs/eigensdk-go/signerv2"
	"github.com/Layr-Labs/eigensdk-go/types"

	"github.com/Layr-Labs/avs-sync/avssync"
	contractreg "github.com/Layr-Labs/avs-sync/bindings/ContractsRegistry"
)

type ContractAddresses struct {
	RegistryCoordinator    common.Address
	OperatorStateRetriever common.Address
	DelegationManager      common.Address
	Erc20MockStrategy      common.Address
}

// there are 2 ways to call avsSync, either with a list of operators (meant to be run by operator teams)
// or without a list of operators (meant to be run by the avs team to update the entire quorum of operators)

// here we test the case where we call avsSync with a list of operators
func TestIntegrationUpdateSingleOperatorPath(t *testing.T) {

	/* Start the anvil chain */
	anvilC := startAnvilTestContainer()
	// Not sure why but deferring anvilC.Terminate() causes a panic when the test finishes...
	// so letting it terminate silently for now
	anvilHttpEndpoint, err := anvilC.Endpoint(context.Background(), "http")
	if err != nil {
		t.Fatal(err)
	}

	contractAddresses := getContractAddressesFromContractRegistry(anvilHttpEndpoint)
	operatorEcdsaPrivKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	operatorEcdsaPrivKey, err := crypto.HexToECDSA(operatorEcdsaPrivKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	operatorAddr := crypto.PubkeyToAddress(operatorEcdsaPrivKey.PublicKey)
	operatorBlsPrivKey := "0x1"
	c := NewAvsSyncComponents(t, anvilHttpEndpoint, contractAddresses, []common.Address{operatorAddr}, 0)
	avsSync := c.avsSync

	// first register operator into avs. at this point, the operator will have whatever stake it had registered in eigenlayer in the avs
	registerOperatorWithAvs(c.wallet, anvilHttpEndpoint, contractAddresses, operatorEcdsaPrivKeyHex, operatorBlsPrivKey, true)

	// get stake of operator before sync
	operatorsPerQuorumBeforeSync, err := c.avsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}
	operatorStakeBeforeSync := operatorsPerQuorumBeforeSync[0][0].Stake

	// deposit into strategy to create a diff between eigenlayer and avs stakes
	depositAmount := big.NewInt(100)
	depositErc20IntoStrategyForOperator(c.wallet, anvilHttpEndpoint, contractAddresses.DelegationManager, contractAddresses.Erc20MockStrategy, operatorEcdsaPrivKeyHex, operatorAddr.Hex(), depositAmount, false)

	// run avsSync
	go avsSync.Start(context.Background())
	time.Sleep(5 * time.Second)

	// get stake of operator after sync
	operatorsPerQuorumAfterSync, err := c.avsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}
	operatorStakeAfterSync := operatorsPerQuorumAfterSync[0][0].Stake
	operatorStakeDiff := new(big.Int).Sub(operatorStakeAfterSync, operatorStakeBeforeSync)

	// we just check that the diff is equal to the deposited amount
	if operatorStakeDiff.Cmp(depositAmount) != 0 {
		t.Errorf("expected operator stake diff to be equal to deposit amount, got %v", operatorStakeDiff)
	}

}

// Simulating an operator registered between the moment we read the operator set and the moment we try to update the operator set to ensure this behaves as expected
func TestIntegrationFullOperatorSetWithRaceConditionFailsToUpdate(t *testing.T) {
	/* Start the anvil chain with no mining and FIFO transaction ordering to be able to force retries */
	anvilC := startAnvilTestContainer("--order", "fifo", "--no-mining")
	anvilHttpEndpoint, err := anvilC.Endpoint(context.Background(), "http")
	require.NoError(t, err)

	contractAddresses := getContractAddressesFromContractRegistry(anvilHttpEndpoint)

	ethClient, err := ethclient.Dial(anvilHttpEndpoint)
	require.NoError(t, err)

	c := NewAvsSyncComponents(t, anvilHttpEndpoint, contractAddresses, []common.Address{}, 0)
	c.avsSync.RetrySyncNTimes = 1

	operator1EcdsaPrivKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	operator1Addr := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(operator1EcdsaPrivKeyHex)).PublicKey)
	operator1BlsPrivKey := "0x1"

	operator2EcdsaPrivKeyHex := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	operator2BlsPrivKey := "0x2"
	operator2Addr := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(operator2EcdsaPrivKeyHex)).PublicKey)
	operator2Wallet := createWalletForOperator(t, operator2EcdsaPrivKeyHex, ethClient)

	// Register first operator
	registerOperatorWithAvs(c.wallet, anvilHttpEndpoint, contractAddresses, operator1EcdsaPrivKeyHex, operator1BlsPrivKey, false)

	// mine block
	advanceChainByNBlocks(1, anvilC)

	// get state pre sync
	operatorsPerQuorumBeforeSync, err := c.avsSync.AvsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}

	operatorStakeBeforeSync := operatorsPerQuorumBeforeSync[0][0].Stake

	// deposit into strategy to create a diff between eigenlayer and avs stakes
	depositAmount := big.NewInt(100)
	depositErc20IntoStrategyForOperator(c.wallet, anvilHttpEndpoint, contractAddresses.DelegationManager, contractAddresses.Erc20MockStrategy, operator1EcdsaPrivKeyHex, operator1Addr.Hex(), depositAmount, false)

	// mine block
	advanceChainByNBlocks(1, anvilC)

	// Register the second operator. Recall that because we are running anvil in FIFO mode
	// this transaction will be included before the call to UpdateStakesOfOperatorSubsetForAllQuorums
	registerOperatorWithAvs(operator2Wallet, anvilHttpEndpoint, contractAddresses, operator2EcdsaPrivKeyHex, operator2BlsPrivKey, false)

	// Start the sync
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go c.avsSync.Start(ctx)

	// Mine another block to include operator2's registration
	advanceChainByNBlocks(1, anvilC)

	// Wait for sync process to complete
	time.Sleep(2 * time.Second)

	// Verify the final state
	operators, err := c.avsReader.GetOperatorAddrsInQuorumsAtCurrentBlock(
		&bind.CallOpts{},
		[]types.QuorumNum{0},
	)
	require.NoError(t, err)

	// We expect both operators to be registered
	require.Len(t, operators[0], 2, "Expected both operators to be registered")

	// Verify each operator is in the set
	operatorAddrs := make(map[common.Address]bool)
	for _, op := range operators[0] {
		operatorAddrs[op] = true
	}

	require.True(t, operatorAddrs[operator1Addr], "Operator 1 should be registered")
	require.True(t, operatorAddrs[operator2Addr], "Operator 2 should be registered")

	// get stake of operator after sync
	operatorsPerQuorumAfterSync, err := c.avsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}
	operatorStakeAfterSync := operatorsPerQuorumAfterSync[0][0].Stake

	fmt.Printf("Stake before sync: %v\n", operatorStakeBeforeSync)
	fmt.Printf("Stake after sync: %v\n", operatorStakeAfterSync)

	// we check that the stake before and after the sync are the same despite the deposit happening
	if operatorStakeBeforeSync.Cmp(operatorStakeAfterSync) != 0 {
		fmt.Printf("%v", operatorStakeAfterSync)
		fmt.Printf("%v", depositAmount)
		t.Errorf("expected operator stake before and after sync to be equal, got operatorStakeBeforeSync %v operatorStakeAfterSync %v ", operatorStakeBeforeSync, operatorStakeAfterSync)
	}

}

// here we test the case where we call avsSync without a list of operators
// although the operator set here consists of a single operator, the code path is different
// we force a retry by making the first updateOperator call faill
// this would for eg happen if an operator registered between the moment we read the operator set and the moment we try to update the operator set
// since the contract makes sure we are updating the full operator set
func TestIntegrationFullOperatorSetWithRetry(t *testing.T) {
	/* Start the anvil chain with no mining and FIFO transaction ordering to be able to force retries */
	anvilC := startAnvilTestContainer("--order", "fifo", "--no-mining")
	anvilHttpEndpoint, err := anvilC.Endpoint(context.Background(), "http")
	require.NoError(t, err)

	contractAddresses := getContractAddressesFromContractRegistry(anvilHttpEndpoint)

	ethClient, err := ethclient.Dial(anvilHttpEndpoint)
	require.NoError(t, err)

	c := NewAvsSyncComponents(t, anvilHttpEndpoint, contractAddresses, []common.Address{}, 0)
	c.avsSync.RetrySyncNTimes = 10

	operator1EcdsaPrivKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	operator1Addr := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(operator1EcdsaPrivKeyHex)).PublicKey)
	operator1BlsPrivKey := "0x1"

	operator2EcdsaPrivKeyHex := "59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d"
	operator2BlsPrivKey := "0x2"
	operator2Addr := crypto.PubkeyToAddress(crypto.ToECDSAUnsafe(common.FromHex(operator2EcdsaPrivKeyHex)).PublicKey)
	operator2Wallet := createWalletForOperator(t, operator2EcdsaPrivKeyHex, ethClient)

	// Register first operator
	registerOperatorWithAvs(c.wallet, anvilHttpEndpoint, contractAddresses, operator1EcdsaPrivKeyHex, operator1BlsPrivKey, false)

	// mine block
	advanceChainByNBlocks(1, anvilC)

	// get state pre sync
	operatorsPerQuorumBeforeSync, err := c.avsSync.AvsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}

	operatorStakeBeforeSync := operatorsPerQuorumBeforeSync[0][0].Stake

	// deposit into strategy to create a diff between eigenlayer and avs stakes
	depositAmount := big.NewInt(100)
	depositErc20IntoStrategyForOperator(c.wallet, anvilHttpEndpoint, contractAddresses.DelegationManager, contractAddresses.Erc20MockStrategy, operator1EcdsaPrivKeyHex, operator1Addr.Hex(), depositAmount, false)
	advanceChainByNBlocks(1, anvilC)

	// Register the second operator. Recall that because we are running anvil in FIFO mode
	// this transaction will be included before the call to UpdateStakesOfOperatorSubsetForAllQuorums
	registerOperatorWithAvs(operator2Wallet, anvilHttpEndpoint, contractAddresses, operator2EcdsaPrivKeyHex, operator2BlsPrivKey, false)

	// Start the sync
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go c.avsSync.Start(ctx)

	// Mine another block to include operator2's registration then wait for update
	advanceChainByNBlocks(1, anvilC)
	time.Sleep(500 * time.Millisecond)

	// Mine Block to include update
	advanceChainByNBlocks(1, anvilC)

	// Wait for sync process to complete
	time.Sleep(10 * time.Second)

	// Verify the final state
	operators, err := c.avsReader.GetOperatorAddrsInQuorumsAtCurrentBlock(
		&bind.CallOpts{},
		[]types.QuorumNum{0},
	)
	require.NoError(t, err)

	// We expect both operators to be registered
	require.Len(t, operators[0], 2, "Expected both operators to be registered")

	// Verify each operator is in the set
	operatorAddrs := make(map[common.Address]bool)
	for _, op := range operators[0] {
		operatorAddrs[op] = true
	}

	require.True(t, operatorAddrs[operator1Addr], "Operator 1 should be registered")
	require.True(t, operatorAddrs[operator2Addr], "Operator 2 should be registered")

	// get stake of operator after sync
	operatorsPerQuorumAfterSync, err := c.avsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}
	operatorStakeAfterSync := operatorsPerQuorumAfterSync[0][0].Stake
	operatorStakeDiff := new(big.Int).Sub(operatorStakeAfterSync, operatorStakeBeforeSync)

	fmt.Printf("Stake before sync: %v\n", operatorStakeBeforeSync)
	fmt.Printf("Stake after sync: %v\n", operatorStakeAfterSync)

	// we just check that the diff is equal to the deposited amount
	if operatorStakeDiff.Cmp(depositAmount) != 0 {
		t.Errorf("expected operator stake diff to be equal to deposit amount, got %v", operatorStakeDiff)
	}

}

func TestIntegrationFullOperatorSet(t *testing.T) {
	/* Start the anvil chain */
	anvilC := startAnvilTestContainer()
	// Not sure why but deferring anvilC.Terminate() causes a panic when the test finishes...
	// so letting it terminate silently for now
	anvilHttpEndpoint, err := anvilC.Endpoint(context.Background(), "http")
	if err != nil {
		t.Fatal(err)
	}

	contractAddresses := getContractAddressesFromContractRegistry(anvilHttpEndpoint)
	operatorEcdsaPrivKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	operatorEcdsaPrivKey, err := crypto.HexToECDSA(operatorEcdsaPrivKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	operatorAddr := crypto.PubkeyToAddress(operatorEcdsaPrivKey.PublicKey)
	operatorBlsPrivKey := "0x1"
	// set sync interval to 0 so that we only run once
	c := NewAvsSyncComponents(t, anvilHttpEndpoint, contractAddresses, []common.Address{}, 0)
	avsSync := c.avsSync

	// first register operator into avs. at this point, the operator will have whatever stake it had registered in eigenlayer in the avs
	registerOperatorWithAvs(c.wallet, anvilHttpEndpoint, contractAddresses, operatorEcdsaPrivKeyHex, operatorBlsPrivKey, true)

	// get stake of operator before sync
	operatorsPerQuorumBeforeSync, err := avsSync.AvsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}
	// TODO: should be checking all operators, not just the first one
	operatorStakeBeforeSync := operatorsPerQuorumBeforeSync[0][0].Stake

	// deposit into strategy to create a diff between eigenlayer and avs stakes
	depositAmount := big.NewInt(100)
	depositErc20IntoStrategyForOperator(c.wallet, anvilHttpEndpoint, contractAddresses.DelegationManager, contractAddresses.Erc20MockStrategy, operatorEcdsaPrivKeyHex, operatorAddr.Hex(), depositAmount, true)

	avsSync.Start(context.Background())

	// get stake of operator after sync
	operatorsPerQuorumAfterSync, err := avsSync.AvsReader.GetOperatorsStakeInQuorumsAtCurrentBlock(&bind.CallOpts{}, []types.QuorumNum{0})
	if err != nil {
		t.Fatal(err)
	}
	operatorStakeAfterSync := operatorsPerQuorumAfterSync[0][0].Stake
	operatorStakeDiff := new(big.Int).Sub(operatorStakeAfterSync, operatorStakeBeforeSync)

	// we just check that the diff is equal to the deposited amount
	if operatorStakeDiff.Cmp(depositAmount) != 0 {
		t.Errorf("expected operator stake diff to be equal to deposit amount, got %v", operatorStakeDiff)
	}
}

type AvsSyncComponents struct {
	avsSync   *avssync.AvsSync
	wallet    walletsdk.Wallet
	avsReader *avsregistry.ChainReader
	avsWriter *avsregistry.ChainWriter
}

func NewAvsSyncComponents(t *testing.T, anvilHttpEndpoint string, contractAddresses ContractAddresses, operators []common.Address, syncInterval time.Duration) *AvsSyncComponents {
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		panic(err)
	}
	ecdsaPrivKey, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		panic(err)
	}
	ecdsaAddr := crypto.PubkeyToAddress(ecdsaPrivKey.PublicKey)

	reg := prometheus.NewRegistry()
	rpcCollector := rpccalls.NewCollector("", reg)

	ethInstrumentedHttpClient, err := eth.NewInstrumentedClient(anvilHttpEndpoint, rpcCollector)
	if err != nil {
		panic(err)
	}

	rpcCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	chainid, err := ethInstrumentedHttpClient.ChainID(rpcCtx)
	if err != nil {
		panic(err)
	}
	// confusing interface, see https://github.com/Layr-Labs/eigensdk-go/issues/90
	signerFn, _, err := signerv2.SignerFromConfig(signerv2.Config{
		PrivateKey: ecdsaPrivKey,
	}, chainid)
	if err != nil {
		panic(err)
	}
	wallet, err := walletsdk.NewPrivateKeyWallet(ethInstrumentedHttpClient, signerFn, ecdsaAddr, logger)
	if err != nil {
		panic(err)
	}

	txMgr := txmgr.NewSimpleTxManager(wallet, ethInstrumentedHttpClient, logger, ecdsaAddr)

	avsWriter, err := avsregistry.BuildAvsRegistryChainWriter(
		contractAddresses.RegistryCoordinator,
		contractAddresses.OperatorStateRetriever,
		logger,
		ethInstrumentedHttpClient,
		txMgr,
	)
	if err != nil {
		panic(err)
	}
	avsReader, err := avsregistry.BuildAvsRegistryChainReader(
		contractAddresses.RegistryCoordinator,
		contractAddresses.OperatorStateRetriever,
		ethInstrumentedHttpClient,
		logger,
	)
	if err != nil {
		logger.Fatalf("Cannot create avs reader", "err", err)
	}

	avsSync := avssync.NewAvsSync(
		logger,
		*avsReader,
		*avsWriter,
		0*time.Second,
		syncInterval,
		operators,
		// we only test with one quorum
		[]byte{0},
		false,
		1, // 1 retry
		5*time.Second,
		5*time.Second,
		"", // no metrics server (otherwise parallel tests all try to start server at same endpoint and error out)
		reg,
	)
	return &AvsSyncComponents{
		avsSync:   avsSync,
		wallet:    wallet,
		avsReader: avsReader,
		avsWriter: avsWriter,
	}
}

func startAnvilTestContainer(additionalFlags ...string) testcontainers.Container {
	integrationDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmdArgs := []string{
		"--host", "0.0.0.0",
		"--load-state", "/root/.anvil/state.json",
	}

	cmdArgs = append(cmdArgs, additionalFlags...)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "ghcr.io/foundry-rs/foundry:stable@sha256:daeeaaf4383ee0cbfc9f31f079a04ffb0123e49e5f67f2a20b5ce1ac1959a4d6",
		Mounts: testcontainers.ContainerMounts{
			testcontainers.ContainerMount{
				Source: testcontainers.GenericBindMountSource{
					HostPath: filepath.Join(integrationDir, "tests/eigenlayer-eigencert-eigenda-strategies-deployed-operators-registered-with-eigenlayer-anvil-state.json"),
				},
				Target: "/root/.anvil/state.json",
			},
		},
		Entrypoint:   []string{"anvil"},
		Cmd:          cmdArgs,
		ExposedPorts: []string{"8545/tcp"},
		WaitingFor:   wait.ForLog("Listening on"),
	}
	anvilC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	// this is needed temporarily because anvil restarts at 0 block when we load a state...
	// see comment in start-anvil-chain-with-el-and-avs-deployed.sh
	// 25 is arbitrary, but I think it's enough (not sure at which block exactly deployment happened)
	// this is still needed as of the latest stable anvil
	advanceChainByNBlocks(25, anvilC)

	return anvilC
}

func advanceChainByNBlocks(n int, anvilC testcontainers.Container) {
	anvilEndpoint, err := anvilC.Endpoint(context.Background(), "")
	if err != nil {
		panic(err)
	}
	rpcUrl := "http://" + anvilEndpoint
	// this is just the first anvil address, which is funded so can send ether
	// we just send a transaction to ourselves to advance the chain
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(
			`cast rpc anvil_mine %d --rpc-url %s`,
			n, rpcUrl),
	)
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}

// TODO(samlaf): move this function to eigensdk
func registerOperatorWithAvs(wallet walletsdk.Wallet, ethHttpUrl string, contractAddresses ContractAddresses, ecdsaPrivKeyHex string, blsPrivKeyHex string, waitForMine bool) {
	ethHttpClient, err := NewEthHttpClient(ethHttpUrl)
	if err != nil {
		panic(err)
	}
	blsKeyPair, err := bls.NewKeyPairFromString(blsPrivKeyHex)
	if err != nil {
		panic(err)
	}
	ecdsaPrivKey, err := crypto.HexToECDSA(ecdsaPrivKeyHex)
	if err != nil {
		panic(err)
	}
	ecdsaAddr := crypto.PubkeyToAddress(ecdsaPrivKey.PublicKey)

	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		panic(err)
	}
	txMgr := txmgr.NewSimpleTxManager(wallet, ethHttpClient, logger, ecdsaAddr)

	avsWriter, err := avsregistry.BuildAvsRegistryChainWriter(
		contractAddresses.RegistryCoordinator,
		contractAddresses.OperatorStateRetriever,
		logger,
		ethHttpClient,
		txMgr,
	)
	if err != nil {
		panic(err)
	}

	quorumNumbers := []types.QuorumNum{0}
	socket := "Not Needed"
	operatorToAvsRegistrationSigSalt := [32]byte{123}
	curBlockNum, err := ethHttpClient.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	curBlock, err := ethHttpClient.BlockByNumber(context.Background(), big.NewInt(int64(curBlockNum)))
	if err != nil {
		panic(err)
	}
	sigValidForSeconds := int64(1_000_000)
	operatorToAvsRegistrationSigExpiry := big.NewInt(int64(curBlock.Time()) + sigValidForSeconds)
	_, err = avsWriter.RegisterOperatorInQuorumWithAVSRegistryCoordinator(
		context.Background(),
		ecdsaPrivKey, operatorToAvsRegistrationSigSalt, operatorToAvsRegistrationSigExpiry,
		blsKeyPair, quorumNumbers, socket, waitForMine,
	)
	if err != nil {
		panic(err)
	}
}

// TODO(samlaf): move this function to eigensdk
func depositErc20IntoStrategyForOperator(
	wallet walletsdk.Wallet,
	ethHttpUrl string,
	delegationManagerAddr common.Address,
	erc20MockStrategyAddr common.Address,
	ecdsaPrivKeyHex string,
	operatorAddressHex string,
	amount *big.Int,
	waitForMined bool,
) {
	ethHttpClient, err := NewEthHttpClient(ethHttpUrl)
	if err != nil {
		panic(err)
	}
	ecdsaPrivKey, err := crypto.HexToECDSA(ecdsaPrivKeyHex)
	if err != nil {
		panic(err)
	}
	ecdsaAddr := crypto.PubkeyToAddress(ecdsaPrivKey.PublicKey)

	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		panic(err)
	}
	noopMetrics := metrics.NewNoopMetrics()

	txMgr := txmgr.NewSimpleTxManager(wallet, ethHttpClient, logger, ecdsaAddr)
	elWriter, err := elcontracts.BuildELChainWriter(
		delegationManagerAddr,
		common.Address{}, // avsDirectory not needed so we just pass an empty address
		ethHttpClient,
		logger,
		noopMetrics,
		txMgr,
	)
	if err != nil {
		panic(err)
	}

	_, err = elWriter.DepositERC20IntoStrategy(context.Background(), erc20MockStrategyAddr, amount, waitForMined)
	if err != nil {
		panic(err)
	}

}

func getContractAddressesFromContractRegistry(ethHttpUrl string) ContractAddresses {
	ethHttpClient, err := NewEthHttpClient(ethHttpUrl)
	if err != nil {
		panic(err)
	}
	// The ContractsRegistry contract should always be deployed at this address on anvil
	// it's the address of the contract created at nonce 0 by the first anvil account (0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266)
	contractsRegistry, err := contractreg.NewContractContractsRegistry(common.HexToAddress("0x5FbDB2315678afecb367f032d93F642f64180aa3"), ethHttpClient)
	if err != nil {
		panic(err)
	}
	registryCoordinatorAddr, err := contractsRegistry.Contracts(&bind.CallOpts{}, "eigencertRegistryCoordinator")
	if err != nil {
		panic(err)
	}
	operatorStateRetrieverAddr, err := contractsRegistry.Contracts(&bind.CallOpts{}, "eigencertOperatorStateRetriever")
	if err != nil {
		panic(err)
	}
	delegationManagerAddr, err := contractsRegistry.Contracts(&bind.CallOpts{}, "delegationManager")
	if err != nil {
		panic(err)
	}
	erc20MockStrategyAddr, err := contractsRegistry.Contracts(&bind.CallOpts{}, "erc20MockStrategy")
	if err != nil {
		panic(err)
	}
	return ContractAddresses{
		RegistryCoordinator:    registryCoordinatorAddr,
		OperatorStateRetriever: operatorStateRetrieverAddr,
		DelegationManager:      delegationManagerAddr,
		Erc20MockStrategy:      erc20MockStrategyAddr,
	}
}

func NewEthHttpClient(rpcAddress string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(rpcAddress)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func createWalletForOperator(t *testing.T, privKeyHex string, ethClient *ethclient.Client) walletsdk.Wallet {
	logger, err := logging.NewZapLogger(logging.Development)
	if err != nil {
		panic(err)
	}

	ecdsaPrivKey, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		t.Fatal(err)
	}
	ecdsaAddr := crypto.PubkeyToAddress(ecdsaPrivKey.PublicKey)

	chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	signerFn, _, err := signerv2.SignerFromConfig(signerv2.Config{
		PrivateKey: ecdsaPrivKey,
	}, chainID)
	if err != nil {
		t.Fatal(err)
	}

	wallet, err := walletsdk.NewPrivateKeyWallet(ethClient, signerFn, ecdsaAddr, logger)
	if err != nil {
		t.Fatal(err)
	}

	return wallet
}
