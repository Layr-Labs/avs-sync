package main

import (
	"time"

	"github.com/urfave/cli"
)

var envVarPrefix = "AVS_SYNC_"

var (
	/* Required Flags */
	EcdsaPrivateKeyFlag = cli.StringFlag{
		Name:     "ecdsa-private-key",
		Required: true,
		Usage:    "Ethereum ecdsa private key",
		EnvVar:   envVarPrefix + "ECDSA_PRIVATE_KEY",
	}
	RegistryCoordinatorAddrFlag = cli.StringFlag{
		Name:     "registry-coordinator-addr",
		Required: true,
		Usage:    "AVS Registry coordinator address",
		EnvVar:   envVarPrefix + "REGISTRY_COORDINATOR_ADDR",
	}
	OperatorStateRetrieverAddrFlag = cli.StringFlag{
		Name:     "operator-state-retriever-addr",
		Required: true,
		Usage:    "AVS Operator state retriever address",
		EnvVar:   envVarPrefix + "OPERATOR_STATE_RETRIEVER_ADDR",
	}
	EthHttpUrlFlag = cli.StringFlag{
		Name:     "eth-http-url",
		Required: true,
		Usage:    "Ethereum http url",
		EnvVar:   envVarPrefix + "ETH_HTTP_URL",
	}
	SyncIntervalFlag = cli.DurationFlag{
		Name:     "sync-interval",
		Required: true,
		Usage:    "Interval at which to sync with the chain (e.g. 24h). If set to 0, will only sync once and then exit.",
		Value:    24 * time.Hour,
		EnvVar:   envVarPrefix + "SYNC_INTERVAL",
	}
	/* Optional Flags */
	FirstSyncTimeFlag = cli.StringFlag{
		Name:     "first-sync-time",
		Required: false,
		Usage:    "Set the HH:MI:SS time at which to run the first sync update",
		EnvVar:   envVarPrefix + "FIRST_SYNC_TIME",
	}
	OperatorListFlag = cli.StringSliceFlag{
		Name:   "operators",
		Usage:  "List of operators to update stakes for",
		EnvVar: envVarPrefix + "OPERATORS",
	}
	QuorumListFlag = cli.IntSliceFlag{
		Name:   "quorums",
		Usage:  "List of quorums to update stakes for (only needs to be present if operators list not present and fetch-quorums-dynamically is false)",
		EnvVar: envVarPrefix + "QUORUMS",
	}
	NoFetchQuorumDynamicallyFlag = cli.BoolFlag{
		Name:   "fetch-quorums-dynamically",
		Usage:  "If set to true, will use the quorumList argument instead of fetching the list of quorums registered in the contract dynamically and updating all of them",
		EnvVar: envVarPrefix + "FETCH_QUORUMS_DYNAMICALLY",
	}
	ReaderTimeoutDurationFlag = cli.DurationFlag{
		Name:   "reader-timeout-duration",
		Usage:  "Timeout duration for rpc calls to read from chain in `SECONDS`",
		Value:  5 * time.Second,
		EnvVar: envVarPrefix + "READER_TIMEOUT_DURATION",
	}
	WriterTimeoutDurationFlag = cli.DurationFlag{
		Name:   "writer-timeout-duration",
		Usage:  "Timeout duration for transactions to be confirmed in `SECONDS`",
		Value:  90 * time.Second,
		EnvVar: envVarPrefix + "WRITER_TIMEOUT_DURATION",
	}
	retrySyncNTimes = cli.IntFlag{
		Name:   "retry-sync-n-times",
		Usage:  "Number of times to retry syncing before giving up",
		Value:  3,
		EnvVar: envVarPrefix + "RETRY_SYNC_N_TIMES",
	}
)

var RequiredFlags = []cli.Flag{
	EcdsaPrivateKeyFlag,
	RegistryCoordinatorAddrFlag,
	OperatorStateRetrieverAddrFlag,
	EthHttpUrlFlag,
	SyncIntervalFlag,
}

var OptionalFlags = []cli.Flag{
	FirstSyncTimeFlag,
	OperatorListFlag,
	QuorumListFlag,
	NoFetchQuorumDynamicallyFlag,
	ReaderTimeoutDurationFlag,
	WriterTimeoutDurationFlag,
	retrySyncNTimes,
}

func init() {
	Flags = append(RequiredFlags, OptionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
