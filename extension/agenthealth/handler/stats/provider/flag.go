// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package provider

import (
	"log"
	"sync"
	"time"

	"github.com/aws/amazon-cloudwatch-agent/cfg/envconfig"
	"github.com/aws/amazon-cloudwatch-agent/extension/agenthealth/handler/stats/agent"
	"github.com/aws/aws-sdk-go/aws"
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
	describeTagsStatusCounts []int // Changed to a slice
	flagSet                  agent.FlagSet
}

// StatusCounters holds counters for success and failure
type StatusCounters struct {
	Counters []int // Changed to a slice
}

var (
	describeTagsCounters = StatusCounters{Counters: make([]int, 2)} // Initialize with a slice
	counterMutex         sync.Mutex
)

// resetDescribeTagsCounter resets the global counters every 5 minutes
func resetDescribeTagsCounter() {
	log.Println("Starting resetDescribeTagsCounter loop")
	for {
		time.Sleep(5 * time.Minute)
		counterMutex.Lock()
		describeTagsCounters.Counters = make([]int, 2) // Reset using a new slice
		log.Println("Reset describeTagsCounters to:", describeTagsCounters.Counters)
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
		log.Printf("Incremented success counter: %d", describeTagsCounters.Counters[0])
	} else {
		describeTagsCounters.Counters[1]++
		log.Printf("Incremented failure counter: %d", describeTagsCounters.Counters[1])
	}
	log.Println("Current counters:", describeTagsCounters.Counters)
}

// Update the flagStats with current counter values
func (p *flagStats) update() {
	log.Println("Updating flagStats with current counters")
	p.stats.Store(agent.Stats{
		ImdsFallbackSucceed:       boolToSparseInt(p.flagSet.IsSet(agent.FlagIMDSFallbackSuccess)),
		SharedConfigFallback:      boolToSparseInt(p.flagSet.IsSet(agent.FlagSharedConfigFallback)),
		AppSignals:                boolToSparseInt(p.flagSet.IsSet(agent.FlagAppSignal)),
		EnhancedContainerInsights: boolToSparseInt(p.flagSet.IsSet(agent.FlagEnhancedContainerInsights)),
		RunningInContainer:        boolToInt(p.flagSet.IsSet(agent.FlagRunningInContainer)),
		Mode:                      p.flagSet.GetString(agent.FlagMode),
		RegionType:                p.flagSet.GetString(agent.FlagRegionType),
	})
	log.Printf("Updated flagStats: %+v", p.stats.Load())
}

func boolToInt(value bool) *int {
	log.Printf("Converting bool to int: %v", value)
	result := boolToSparseInt(value)
	if result != nil {
		return result
	}
	return aws.Int(0)
}

func boolToSparseInt(value bool) *int {
	if value {
		log.Println("Converting bool to sparse int: true")
		return aws.Int(1)
	}
	log.Println("Converting bool to sparse int: false")
	return nil
}

func newFlagStats(flagSet agent.FlagSet, interval time.Duration) *flagStats {
	log.Println("Creating new flagStats")
	stats := &flagStats{
		flagSet:       flagSet,
		intervalStats: newIntervalStats(interval),
	}
	stats.flagSet.OnChange(stats.update)
	if envconfig.IsRunningInContainer() {
		log.Println("Running in container, setting flagRunningInContainer")
		stats.flagSet.Set(agent.FlagRunningInContainer)
	} else {
		log.Println("Not running in container, updating stats")
		stats.update()
	}
	return stats
}

func GetFlagsStats() agent.StatsProvider {
	flagOnce.Do(func() {
		log.Println("Initializing flagSingleton with newFlagStats")
		flagSingleton = newFlagStats(agent.UsageFlags(), flagGetInterval)
		go resetDescribeTagsCounter()
	})
	return flagSingleton
}
