// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package provider

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/amazon-cloudwatch-agent/cfg/envconfig"
	"github.com/aws/amazon-cloudwatch-agent/extension/agenthealth/handler/stats/agent"
)

const (
	flagGetInterval = 5 * time.Minute
)

var (
	flagSingleton *flagStats
	flagOnce      sync.Once
)

type flagStats struct {
	*intervalStats
	describeTagsStatusCounts [2]int

	flagSet agent.FlagSet
}

// StatusCounters holds counters for success and failure
type StatusCounters struct {
	Counters [2]int
}

var (
	describeTagsCounters StatusCounters
	counterMutex         sync.Mutex
)

// resetDescribeTagsCounter resets the global counters every 5 minutes
func resetDescribeTagsCounter() {
	for {
		time.Sleep(5 * time.Minute)
		counterMutex.Lock()
		describeTagsCounters = StatusCounters{}
		counterMutex.Unlock()
	}
}

// IncrementDescribeTagsCounter increments the counter for DescribeTags API calls
// based on whether the call was a success or failure
func IncrementDescribeTagsCounter(isSuccess bool) {
	counterMutex.Lock()
	defer counterMutex.Unlock()

	if isSuccess {
		describeTagsCounters.Counters[0]++
	} else {
		describeTagsCounters.Counters[1]++
	}
	fmt.Println(describeTagsCounters.Counters[0])
	fmt.Println(describeTagsCounters.Counters[1])
	fmt.Println("Above are the counters")

}

// GetDescribeTagsCounters retrieves the current values of the counters
func GetDescribeTagsCounters() [2]int {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	return describeTagsCounters.Counters
}

// Update the flagStats with current counter values
func (p *flagStats) update() {
	counters := GetDescribeTagsCounters()
	p.stats.Store(agent.Stats{
		ImdsFallbackSucceed:       boolToSparseInt(p.flagSet.IsSet(agent.FlagIMDSFallbackSuccess)),
		SharedConfigFallback:      boolToSparseInt(p.flagSet.IsSet(agent.FlagSharedConfigFallback)),
		AppSignals:                boolToSparseInt(p.flagSet.IsSet(agent.FlagAppSignal)),
		EnhancedContainerInsights: boolToSparseInt(p.flagSet.IsSet(agent.FlagEnhancedContainerInsights)),
		RunningInContainer:        boolToInt(p.flagSet.IsSet(agent.FlagRunningInContainer)),
		Mode:                      p.flagSet.GetString(agent.FlagMode),
		RegionType:                p.flagSet.GetString(agent.FlagRegionType),
		DescribeTagsApiCounts:     counters,
	})
}

func boolToInt(value bool) *int {
	result := boolToSparseInt(value)
	if result != nil {
		return result
	}
	return aws.Int(0)
}

func boolToSparseInt(value bool) *int {
	if value {
		return aws.Int(1)
	}
	return nil
}

func newFlagStats(flagSet agent.FlagSet, interval time.Duration) *flagStats {
	stats := &flagStats{
		flagSet:       flagSet,
		intervalStats: newIntervalStats(interval),
	}
	stats.flagSet.OnChange(stats.update)
	if envconfig.IsRunningInContainer() {
		stats.flagSet.Set(agent.FlagRunningInContainer)
	} else {
		stats.update()
	}
	return stats
}

func GetFlagsStats() agent.StatsProvider {
	flagOnce.Do(func() {
		flagSingleton = newFlagStats(agent.UsageFlags(), flagGetInterval)
		go resetDescribeTagsCounter()
	})
	return flagSingleton
}
