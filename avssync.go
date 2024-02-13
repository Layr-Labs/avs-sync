package main

import (
	"context"
	"sort"
	"time"

	"github.com/Layr-Labs/eigensdk-go/chainio/clients/avsregistry"
	sdklogging "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type AvsSync struct {
	logger    sdklogging.Logger
	avsReader avsregistry.AvsRegistryReader
	avsWriter avsregistry.AvsRegistryWriter

	sleepBeforeFirstSyncDuration time.Duration
	syncInterval                 time.Duration
	operators                    []common.Address // empty means we update all operators
	quorums                      []byte
	fetchQuorumsDynamically      bool
	retrySyncNTimes              int

	readerTimeoutDuration time.Duration
	writerTimeoutDuration time.Duration
}

// NewAvsSync creates a new AvsSync object
//
//	operators - list of operators to update. If empty, update all operators (must give a list of quorums, or set fetchQuorumsDynamically to true)
//	quorums - list of quorums to update (only needed if operators is not empty)
//	fetchQuorumsDynamically - if true, fetch the list of quorums registered in the contract and update all of them (only needed if operators is not empty)
func NewAvsSync(
	logger sdklogging.Logger,
	avsReader avsregistry.AvsRegistryReader, avsWriter avsregistry.AvsRegistryWriter,
	sleepBeforeFirstSyncDuration time.Duration, syncInterval time.Duration, operators []common.Address,
	quorums []byte, fetchQuorumsDynamically bool, retrySyncNTimes int,
	readerTimeoutDuration time.Duration,
	writerTimeoutDuration time.Duration,
) *AvsSync {
	return &AvsSync{
		logger:                       logger,
		avsReader:                    avsReader,
		avsWriter:                    avsWriter,
		sleepBeforeFirstSyncDuration: sleepBeforeFirstSyncDuration,
		syncInterval:                 syncInterval,
		operators:                    operators,
		quorums:                      quorums,
		fetchQuorumsDynamically:      fetchQuorumsDynamically,
		retrySyncNTimes:              retrySyncNTimes,
		readerTimeoutDuration:        readerTimeoutDuration,
		writerTimeoutDuration:        writerTimeoutDuration,
	}
}

func (a *AvsSync) Start() {
	a.logger.Infof("Starting avs sync with sleepBeforeFirstSyncDuration=%s, syncInterval=%s, operators=%v, quorums=%v, fetchQuorumsDynamically=%v, readerTimeoutDuration=%s, writerTimeoutDuration=%s",
		a.sleepBeforeFirstSyncDuration, a.syncInterval, a.operators, convertQuorumsBytesToInts(a.quorums), a.fetchQuorumsDynamically, a.readerTimeoutDuration, a.writerTimeoutDuration)

	// ticker doesn't tick immediately, so we send a first updateStakes here
	// see https://github.com/golang/go/issues/17601
	// we first sleep some amount of time before the first sync, which allows the syncs to happen at some preferred time
	// for eg midnight every night, without needing to schedule the start of avssync outside of this program
	time.Sleep(a.sleepBeforeFirstSyncDuration)
	err := a.updateStakes()
	if err != nil {
		a.logger.Error("Error updating stakes", err)
	}

	if a.syncInterval == 0 {
		a.logger.Infof("Sync interval is 0, running updateStakes once and exiting")
		return // only run once
	}

	// update stakes every syncInterval
	ticker := time.NewTicker(a.syncInterval)
	defer ticker.Stop()

	for range ticker.C {
		err := a.updateStakes()
		if err != nil {
			a.logger.Error("Error updating stakes", err)
		}
		a.logger.Infof("Sleeping for %s\n", a.syncInterval)
	}
}

func (a *AvsSync) updateStakes() error {
	if len(a.operators) == 0 {
		a.logger.Info("Updating stakes of entire operator set")
		a.maybeUpdateQuorumSet()
		a.logger.Infof("Current quorum set: %v", convertQuorumsBytesToInts(a.quorums))

		// we update one quorum at a time, just to make sure we don't run into any gas limit issues
		// in case there are a lot of operators in a given quorum
		for _, quorum := range a.quorums {
			a.tryNTimesUpdateStakesOfEntireOperatorSetForQuorum(quorum, a.retrySyncNTimes)
		}
		a.logger.Info("Completed stake update successfully")
		return nil
	} else {
		a.logger.Infof("Updating stakes of operators: %v", a.operators)
		timeoutCtx, cancel := context.WithTimeout(context.Background(), a.writerTimeoutDuration)
		defer cancel()
		// this one we update all quorums at once, since we're only updating a subset of operators (which should be a small number)
		_, err := a.avsWriter.UpdateStakesOfOperatorSubsetForAllQuorums(timeoutCtx, a.operators)
		if err == nil {
			a.logger.Info("Completed stake update successfully")
		}
		return err
	}
}

func (a *AvsSync) maybeUpdateQuorumSet() {
	if !a.fetchQuorumsDynamically {
		return
	}
	a.logger.Info("Fetching quorum set dynamically")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), a.readerTimeoutDuration)
	defer cancel()
	quorumCount, err := a.avsReader.GetQuorumCount(&bind.CallOpts{Context: timeoutCtx})
	if err != nil {
		a.logger.Error("Error fetching quorum set dynamically", err)
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
		operatorAddrsPerQuorum, err := a.avsReader.GetOperatorAddrsInQuorumsAtCurrentBlock(&bind.CallOpts{Context: timeoutCtx}, []byte{quorum})
		if err != nil {
			a.logger.Error("Error fetching operator addresses in quorums", "err", err)
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
		_, err = a.avsWriter.UpdateStakesOfEntireOperatorSetForQuorums(timeoutCtx, [][]common.Address{operators}, []byte{quorum})
		if err != nil {
			a.logger.Error("Error updating stakes of entire operator set for quorum", "err", err, "quorum", int(quorum))
			continue
		}
		return
	}
	a.logger.Error("Giving up after retrying", "retryNTimes", retryNTimes)
}

func convertQuorumsBytesToInts(quorums []byte) []int {
	var quorumsInts []int
	for _, quorum := range quorums {
		quorumsInts = append(quorumsInts, int(quorum))
	}
	return quorumsInts
}
