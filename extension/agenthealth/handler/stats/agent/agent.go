// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/aws/amazon-cloudwatch-agent/internal/util/collections"
)

const (
	AllowAllOperations = "*"
)

type Stats struct {
	CpuPercent                *float64         `json:"cpu,omitempty"`
	MemoryBytes               *uint64          `json:"mem,omitempty"`
	FileDescriptorCount       *int32           `json:"fd,omitempty"`
	ThreadCount               *int32           `json:"th,omitempty"`
	LatencyMillis             *int64           `json:"lat,omitempty"`
	PayloadBytes              *int             `json:"load,omitempty"`
	StatusCode                *int             `json:"code,omitempty"`
	StatusCodesByAPI          map[string][]int `json:"status_codes_by_api,omitempty"`
	SharedConfigFallback      *int             `json:"scfb,omitempty"`
	ImdsFallbackSucceed       *int             `json:"ifs,omitempty"`
	AppSignals                *int             `json:"as,omitempty"`
	EnhancedContainerInsights *int             `json:"eci,omitempty"`
	RunningInContainer        *int             `json:"ric,omitempty"`
	RegionType                *string          `json:"rt,omitempty"`
	Mode                      *string          `json:"m,omitempty"`
}

// Merge the other Stats into the current. If the field is not nil,
// then it'll overwrite the existing one.
func (s *Stats) Merge(other Stats) {
	if other.CpuPercent != nil {
		s.CpuPercent = other.CpuPercent
	}
	if other.MemoryBytes != nil {
		s.MemoryBytes = other.MemoryBytes
	}
	if other.FileDescriptorCount != nil {
		s.FileDescriptorCount = other.FileDescriptorCount
	}
	if other.ThreadCount != nil {
		s.ThreadCount = other.ThreadCount
	}
	if other.LatencyMillis != nil {
		s.LatencyMillis = other.LatencyMillis
	}
	if other.PayloadBytes != nil {
		s.PayloadBytes = other.PayloadBytes
	}
	if other.StatusCode != nil {
		s.StatusCode = other.StatusCode
	}
	if other.SharedConfigFallback != nil {
		s.SharedConfigFallback = other.SharedConfigFallback
	}
	if other.ImdsFallbackSucceed != nil {
		s.ImdsFallbackSucceed = other.ImdsFallbackSucceed
	}
	if other.AppSignals != nil {
		s.AppSignals = other.AppSignals
	}
	if other.EnhancedContainerInsights != nil {
		s.EnhancedContainerInsights = other.EnhancedContainerInsights
	}
	if other.RunningInContainer != nil {
		s.RunningInContainer = other.RunningInContainer
	}
	if other.RegionType != nil {
		s.RegionType = other.RegionType
	}
	if other.Mode != nil {
		s.Mode = other.Mode
	}
	if other.StatusCodesByAPI != nil {
		if s.StatusCodesByAPI == nil {
			s.StatusCodesByAPI = make(map[string][]int)
		}
		for api, codes := range other.StatusCodesByAPI {
			// Ensure the target slice is initialized if not present
			if _, exists := s.StatusCodesByAPI[api]; !exists {
				s.StatusCodesByAPI[api] = make([]int, len(codes)) // Initialize with the same length
			}
			// Ensure both slices are the same length before merging
			if len(s.StatusCodesByAPI[api]) == len(codes) {
				for i := range codes {
					s.StatusCodesByAPI[api][i] += codes[i] // Sum the counts for each status code
				}
			} else {
				// Optionally handle different lengths, such as logging an error or adjusting sizes
				// For now, we can just log a warning
				log.Printf("Warning: Status code arrays for API %s have different lengths. Skipping merge for this API.", api)
			}
		}
	}
}

func (s *Stats) Marshal() (string, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	content := strings.TrimPrefix(string(raw), "{")
	return strings.TrimSuffix(content, "}"), nil
}

type StatsProvider interface {
	Stats(operation string) Stats
}

type OperationsFilter struct {
	operations collections.Set[string]
	allowAll   bool
}

func (of OperationsFilter) IsAllowed(operationName string) bool {
	return of.allowAll || of.operations.Contains(operationName)
}

func NewOperationsFilter(operations ...string) OperationsFilter {
	allowed := collections.NewSet[string](operations...)
	return OperationsFilter{
		operations: allowed,
		allowAll:   allowed.Contains(AllowAllOperations),
	}
}

type StatsConfig struct {
	// Operations are the allowed operation names to gather stats for.
	Operations []string `mapstructure:"operations,omitempty"`
	// UsageFlags are the usage flags to set on start up.
	UsageFlags map[Flag]any `mapstructure:"usage_flags,omitempty"`
}
