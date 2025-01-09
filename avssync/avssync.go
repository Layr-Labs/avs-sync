package avssync

import (
	"context"
	"sort"
	"strconv"
	"time"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/types"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/prometheus/client_golang/prometheus"
)

type AvsSync struct {
	AvsReader       *avsregistry.ChainReader
	AvsWriter       *avsregistry.ChainWriter
	RetrySyncNTimes int

	logger                       sdklogging.Logger
	sleepBeforeFirstSyncDuration time.Duration
	syncInterval                 time.Duration
	operators                    []common.Address // empty means we update all operators
	quorums                      []byte
	fetchQuorumsDynamically      bool

	readerTimeoutDuration time.Duration
	writerTimeoutDuration time.Duration
	prometheusServerAddr  string
	Metrics               *Metrics
}

// NewAvsSync creates a new AvsSync object
//
//	operators - list of operators to update. If empty, update all operators (must give a list of quorums, or set fetchQuorumsDynamically to true)
//	quorums - list of quorums to update (only needed if operators is not empty)
//	fetchQuorumsDynamically - if true, fetch the list of quorums registered in the contract and update all of them (only needed if operators is not empty)
func NewAvsSync(
	logger sdklogging.Logger,
	avsReader *avsregistry.ChainReader, avsWriter *avsregistry.ChainWriter,
	sleepBeforeFirstSyncDuration time.Duration, syncInterval time.Duration, operators []common.Address,
	quorums []byte, fetchQuorumsDynamically bool, retrySyncNTimes int,
	readerTimeoutDuration time.Duration, writerTimeoutDuration time.Duration,
	prometheusServerAddr string,
	prometheusRegistry *prometheus.Registry,
) *AvsSync {
	metrics := NewMetrics(prometheusRegistry)

	return &AvsSync{
		AvsReader:                    avsReader,
		AvsWriter:                    avsWriter,
		RetrySyncNTimes:              retrySyncNTimes,
		logger:                       logger,
		sleepBeforeFirstSyncDuration: sleepBeforeFirstSyncDuration,
		syncInterval:                 syncInterval,
		operators:                    operators,
		quorums:                      quorums,
		fetchQuorumsDynamically:      fetchQuorumsDynamically,
		readerTimeoutDuration:        readerTimeoutDuration,
		writerTimeoutDuration:        writerTimeoutDuration,
		prometheusServerAddr:         prometheusServerAddr,
		Metrics:                      metrics,
	}
}

func (a *AvsSync) Start(ctx context.Context) {
	// TODO: should prob put all of these in a config struct, to make sure we don't forget to print any of them
	//       when we add new ones.
	a.logger.Info("Avssync config",
		"sleepBeforeFirstSyncDuration", a.sleepBeforeFirstSyncDuration,
		"syncInterval", a.syncInterval,
		"operators", a.operators,
		"quorums", a.quorums,
		"fetchQuorumsDynamically", a.fetchQuorumsDynamically,
		"readerTimeoutDuration", a.readerTimeoutDuration,
		"writerTimeoutDuration", a.writerTimeoutDuration,
		"prometheusServerAddr", a.prometheusServerAddr,
	)

	if a.prometheusServerAddr != "" {
		a.Metrics.Start(a.prometheusServerAddr)
	} else {
		a.logger.Info("Prometheus server address not set, not starting metrics server")
	}

	// ticker doesn't tick immediately, so we send a first updateStakes here
	// see https://github.com/golang/go/issues/17601
	// we first sleep some amount of time before the first sync, which allows the syncs to happen at some preferred time
	// for eg midnight every night, without needing to schedule the start of avssync outside of this program
	time.Sleep(a.sleepBeforeFirstSyncDuration)
	a.updateStakes()

	if a.syncInterval == 0 {
		a.logger.Infof("Sync interval is 0, running updateStakes once and exiting")
		return // only run once
	}

	// update stakes every syncInterval
	ticker := time.NewTicker(a.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Context done, exiting")
			return
		case <-ticker.C:
			a.updateStakes()
			a.logger.Infof("Sleeping for %s", a.syncInterval)
		}
	}
}

func (a *AvsSync) updateStakes() {
	if len(a.operators) == 0 {
		a.logger.Info("Updating stakes of entire operator set")
		a.maybeUpdateQuorumSet()
		a.logger.Infof("Current quorum set: %v", convertQuorumsBytesToInts(a.quorums))

		// we update one quorum at a time, just to make sure we don't run into any gas limit issues
		// in case there are a lot of operators in a given quorum
		for _, quorum := range a.quorums {
			a.tryNTimesUpdateStakesOfEntireOperatorSetForQuorum(quorum, a.RetrySyncNTimes)
		}
		a.logger.Info("Completed stake update. Check logs to make sure every quorum update succeeded successfully.")
	} else {
		a.logger.Infof("Updating stakes of operators: %v", a.operators)
		timeoutCtx, cancel := context.WithTimeout(context.Background(), a.writerTimeoutDuration)
		defer cancel()
		// this one we update all quorums at once, since we're only updating a subset of operators (which should be a small number)
		receipt, err := a.AvsWriter.UpdateStakesOfOperatorSubsetForAllQuorums(timeoutCtx, a.operators, true)
		if err != nil {
			// no quorum label means we are updating all quorums
			for _, quorum := range a.quorums {
				a.Metrics.UpdateStakeAttemptInc(UpdateStakeStatusError, strconv.Itoa(int(quorum)))
			}
			a.logger.Error("Error updating stakes of operator subset for all quorums", err)
			return
		} else if receipt.Status == gethtypes.ReceiptStatusFailed {
			a.Metrics.TxRevertedTotalInc()
			a.logger.Error("Update stakes of operator subset for all quorums reverted")
			return
		}
		a.logger.Info("Completed stake update successfully")
		return
	}
}

func (a *AvsSync) maybeUpdateQuorumSet() {
	if !a.fetchQuorumsDynamically {
		return
	}
	a.logger.Info("Fetching quorum set dynamically")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), a.readerTimeoutDuration)
	defer cancel()
	quorumCount, err := a.AvsReader.GetQuorumCount(&bind.CallOpts{Context: timeoutCtx})
	if err != nil {
		a.logger.Error("Error fetching quorum set dynamically", err)
		return
	}
	// quorums are numbered from 0 to quorumCount-1,
	// so we just create a list of bytes from 0 to quorumCount-1
	var quorums []byte
	for i := 0; i < int(quorumCount); i++ {
		quorums = append(quorums, byte(i))
	}
	a.quorums = quorums
}

func (a *AvsSync) tryNTimesUpdateStakesOfEntireOperatorSetForQuorum(quorum byte, retryNTimes int) {
	for i := 0; i < retryNTimes; i++ {
		a.logger.Debug("tryNTimesUpdateStakesOfEntireOperatorSetForQuorum", "quorum", int(quorum), "retryNTimes", retryNTimes, "try", i+1)

		timeoutCtx, cancel := context.WithTimeout(context.Background(), a.readerTimeoutDuration)
		defer cancel()
		// we need to refetch the operator set because one reason for update stakes failing is that the operator set has changed
		// in between us fetching it and trying to update it (the contract makes sure the entire operator set is updated and reverts if not)
		operatorAddrsPerQuorum, err := a.AvsReader.GetOperatorAddrsInQuorumsAtCurrentBlock(&bind.CallOpts{Context: timeoutCtx}, types.QuorumNums{types.QuorumNum(quorum)})
		if err != nil {
			a.logger.Warn("Error fetching operator addresses in quorums", "err", err, "quorum", quorum, "retryNTimes", retryNTimes, "try", i+1)
			continue
		}
		var operators []common.Address
		operators = append(operators, operatorAddrsPerQuorum[0]...)
		sort.Slice(operators, func(i, j int) bool {
			return operators[i].Big().Cmp(operators[j].Big()) < 0
		})
		a.logger.Infof("Updating stakes of operators in quorum %d: %v", int(quorum), operators)
		timeoutCtx, cancel = context.WithTimeout(context.Background(), a.writerTimeoutDuration)
		defer cancel()
		receipt, err := a.AvsWriter.UpdateStakesOfEntireOperatorSetForQuorums(timeoutCtx, [][]common.Address{operators}, types.QuorumNums{types.QuorumNum(quorum)}, true)
		if err != nil {
			a.logger.Warn("Error updating stakes of entire operator set for quorum", "err", err, "quorum", int(quorum), "retryNTimes", retryNTimes, "try", i+1)
			continue
		}
		if receipt.Status == gethtypes.ReceiptStatusFailed {
			a.Metrics.TxRevertedTotalInc()
			a.logger.Error("Update stakes of entire operator set for quorum reverted", "quorum", int(quorum))
			continue
		}

		// Update metrics on success
		a.Metrics.UpdateStakeAttemptInc(UpdateStakeStatusSucceed, strconv.Itoa(int(quorum)))
		a.Metrics.OperatorsUpdatedSet(strconv.Itoa(int(quorum)), len(operators))

		return
	}

	// Update metrics on failure
	a.Metrics.UpdateStakeAttemptInc(UpdateStakeStatusError, strconv.Itoa(int(quorum)))
	a.logger.Error("Giving up after retrying", "retryNTimes", retryNTimes)
}

func convertQuorumsBytesToInts(quorums []byte) []int {
	var quorumsInts []int
	for _, quorum := range quorums {
		quorumsInts = append(quorumsInts, int(quorum))
	}
	return quorumsInts
}
