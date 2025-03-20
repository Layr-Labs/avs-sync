package main

import (
	"time"

	"github.com/urfave/cli"
)

var envVarPrefix = "AVS_SYNC_"

var (
	/* Required Flags */
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
	ServiceManagerAddrFlag = cli.StringFlag{
		Name:     "service-manager-addr",
		Required: true,
		Usage:    "AVS Service Manager address",
		EnvVar:   envVarPrefix + "SERVICE_MANAGER_ADDR",
	}
	DontUseAllocationManagerFlag = cli.BoolFlag{
		Name:   "dont-use-allocation-manager",
		Usage:  "If set to true, will not use the allocation manager",
		EnvVar: envVarPrefix + "DONT_USE_ALLOCATION_MANAGER",
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
	MetricsAddrFlag = cli.StringFlag{
		Name:   "metrics-addr",
		Usage:  "Prometheus server address (ip:port)",
		Value:  ":9090",
		EnvVar: envVarPrefix + "PROMETHEUS_SERVER_ADDR",
	}
	FirstSyncTimeFlag = cli.StringFlag{
		Name:     "first-sync-time",
		Required: false,
		Usage:    "Set the HH:MI:SS time at which to run the first sync update (in UTC)",
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
	FetchQuorumDynamicallyFlag = cli.BoolTFlag{
		Name:   "fetch-quorums-dynamically",
		Usage:  "If set to true (default), will fetch the list of quorums registered in the contract and update all of them",
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
	UseFireblocksFlag = cli.BoolTFlag{
		Name:     "use-fireblocks",
		Usage:    "Use Fireblocks to sign transactions. Ignores ecdsa-private-key. Fireblocks credentials must be provided.",
		Required: false,
		EnvVar:   envVarPrefix + "USE_FIREBLOCKS",
	}
	SecretManagerRegionFlag = cli.StringFlag{
		Name:   "secret-manager-region",
		Usage:  "Region of the secret manager",
		EnvVar: envVarPrefix + "SECRET_MANAGER_REGION",
	}
	SecretManagerEcdsaPrivateKeyNameFlag = cli.StringFlag{
		Name:   "secret-manager-ecdsa-private-key-name",
		Usage:  "Name of the secret in the secret manager that contains the Ethereum ecdsa private key. If not set, ecdsa-private-key must be specified.",
		EnvVar: envVarPrefix + "SECRET_MANAGER_ECDSA_PRIVATE_KEY_NAME",
	}
	EcdsaPrivateKeyFlag = cli.StringFlag{
		Name:   "ecdsa-private-key",
		Usage:  "Ethereum ecdsa private key. If not set, Fireblocks credentials must be set.",
		EnvVar: envVarPrefix + "ECDSA_PRIVATE_KEY",
	}
	// Fireblocks flags
	SecretManagerFireblocksAPIKeyNameFlag = cli.StringFlag{
		Name:   "secret-manager-fireblocks-api-key-name",
		Usage:  "Name of the secret in the secret manager that contains the Fireblocks API Key.",
		EnvVar: envVarPrefix + "SECRET_MANAGER_FIREBLOCKS_API_KEY_NAME",
	}
	FireblocksAPIKeyFlag = cli.StringFlag{
		Name:   "fireblocks-api-key",
		Usage:  "Fireblocks API Key. Ignored if ecdsa-private-key is set.",
		EnvVar: envVarPrefix + "FIREBLOCKS_API_KEY",
	}
	SecretManagerFireblocksAPISecretNameFlag = cli.StringFlag{
		Name:   "secret-manager-fireblocks-api-secret-name",
		Usage:  "Name of the secret in the secret manager that contains the Fireblocks API Secret.",
		EnvVar: envVarPrefix + "SECRET_MANAGER_FIREBLOCKS_API_SECRET_NAME",
	}
	FireblocksAPISecretPathFlag = cli.StringFlag{
		Name:   "fireblocks-api-secret-path",
		Usage:  "Path to Fireblocks API Secret. Ignored if ecdsa-private-key is set.",
		EnvVar: envVarPrefix + "FIREBLOCKS_API_SECRET_PATH",
	}
	FireblocksBaseURLFlag = cli.StringFlag{
		Name:   "fireblocks-api-url",
		Usage:  "Fireblocks API URL. Ignored if ecdsa-private-key is set.",
		EnvVar: envVarPrefix + "FIREBLOCKS_API_URL",
		Value:  "https://api.fireblocks.io",
	}
	FireblocksVaultAccountNameFlag = cli.StringFlag{
		Name:   "fireblocks-vault-account-name",
		Usage:  "Fireblocks Vault Account Name. Ignored if ecdsa-private-key is set.",
		EnvVar: envVarPrefix + "FIREBLOCKS_VAULT_ACCOUNT_NAME",
	}
)

var RequiredFlags = []cli.Flag{
	RegistryCoordinatorAddrFlag,
	OperatorStateRetrieverAddrFlag,
	ServiceManagerAddrFlag,
	DontUseAllocationManagerFlag,
	EthHttpUrlFlag,
	SyncIntervalFlag,
}

var OptionalFlags = []cli.Flag{
	MetricsAddrFlag,
	FirstSyncTimeFlag,
	OperatorListFlag,
	QuorumListFlag,
	FetchQuorumDynamicallyFlag,
	ReaderTimeoutDurationFlag,
	WriterTimeoutDurationFlag,
	retrySyncNTimes,
	UseFireblocksFlag,
	SecretManagerRegionFlag,
	SecretManagerEcdsaPrivateKeyNameFlag,
	EcdsaPrivateKeyFlag,
	SecretManagerFireblocksAPIKeyNameFlag,
	FireblocksAPIKeyFlag,
	SecretManagerFireblocksAPISecretNameFlag,
	FireblocksAPISecretPathFlag,
	FireblocksBaseURLFlag,
	FireblocksVaultAccountNameFlag,
}

func init() {
	Flags = append(RequiredFlags, OptionalFlags...)
	Flags = append(Flags, loggerFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag
