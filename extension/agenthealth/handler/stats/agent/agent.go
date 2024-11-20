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
	CpuPercent                *float64          `json:"cpu,omitempty"`
	MemoryBytes               *uint64           `json:"mem,omitempty"`
	FileDescriptorCount       *int32            `json:"fd,omitempty"`
	ThreadCount               *int32            `json:"th,omitempty"`
	LatencyMillis             *int64            `json:"lat,omitempty"`
	PayloadBytes              *int              `json:"load,omitempty"`
	StatusCode                *int              `json:"code,omitempty"`
	SharedConfigFallback      *int              `json:"scfb,omitempty"`
	ImdsFallbackSucceed       *int              `json:"ifs,omitempty"`
	AppSignals                *int              `json:"as,omitempty"`
	EnhancedContainerInsights *int              `json:"eci,omitempty"`
	RunningInContainer        *int              `json:"ric,omitempty"`
	RegionType                *string           `json:"rt,omitempty"`
	Mode                      *string           `json:"m,omitempty"`
	EntityRejected            *int              `json:"ent,omitempty"`
	StatusCodes               map[string][5]int `json:"codes,omitempty"` //represents status codes 200,400,408,413,429,
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
	if other.EntityRejected != nil {
		s.EntityRejected = other.EntityRejected
	}
	if other.StatusCodes != nil {
		log.Println("Merging status codes from another source.")

		if s.StatusCodes == nil {
			log.Println("Initializing status codes map as it was nil.")
			s.StatusCodes = make(map[string][5]int)
		}

		for key, value := range other.StatusCodes {
			log.Printf("Processing key: %s with value: 200=%d, 400=%d, 408=%d, 413=%d, 429=%d", key, value[0], value[1], value[2], value[3], value[4])

			if existing, ok := s.StatusCodes[key]; ok {
				log.Printf(
					"Key %s already exists. Existing: 200=%d, 400=%d, 408=%d, 413=%d, 429=%d. Merging with: 200=%d, 400=%d, 408=%d, 413=%d, 429=%d",
					key, existing[0], existing[1], existing[2], existing[3], existing[4],
					value[0], value[1], value[2], value[3], value[4],
				)

				//Merge the values for each status code
				s.StatusCodes[key] = [5]int{
					existing[0] + value[0], // 200
					existing[1] + value[1], // 400
					existing[2] + value[2], // 408
					existing[3] + value[3], // 413
					existing[4] + value[4], // 429
				}

				log.Printf(
					"Updated key %s: 200=%d, 400=%d, 408=%d, 413=%d, 429=%d",
					key, s.StatusCodes[key][0], s.StatusCodes[key][1], s.StatusCodes[key][2], s.StatusCodes[key][3], s.StatusCodes[key][4],
				)
			} else {
				log.Printf("Key %s does not exist. Adding it with: 200=%d, 400=%d, 408=%d, 413=%d, 429=%d", key, value[0], value[1], value[2], value[3], value[4])
				s.StatusCodes[key] = value
			}
		}
		log.Println("Merging of status codes completed.")
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

type StatusCodeConfig struct {
	// Operations are the allowed operation names to gather stats for.
	Operations []string `json:"operations,omitempty"`
}

var StatusCodeOperations = []string{ // all the operations that are allowed
	"DescribeInstances",
	"DescribeTags",
	"DescribeVolumes",
	"DescribeContainerInstances",
	"DescribeServices",
	"DescribeTaskDefinition",
	"ListServices",
	"ListTasks",
	"CreateLogGroup",
	"CreateLogStream",
}

var StatusCodeAndOtherOperations = []string{ // all the operations that are allowed
	"PutMetricData",
	"PutLogEvents",
	"PutTraceSegments",
	"DescribeInstances",
	"DescribeTags",
	"DescribeVolumes",
	"DescribeContainerInstances",
	"DescribeServices",
	"DescribeTaskDefinition",
	"ListServices",
	"ListTasks",
	"CreateLogGroup",
	"CreateLogStream",
}

// NewStatusCodeOperationsFilter creates a new filter for allowed operations and status codes.
func NewStatusCodeOperationsFilter() OperationsFilter {
	allowed := collections.NewSet[string](StatusCodeOperations...)

	return OperationsFilter{
		operations: allowed,
		allowAll:   allowed.Contains(AllowAllOperations),
	}
}

func NewStatusCodeAndOtherOperationsFilter() OperationsFilter {
	allowed := collections.NewSet[string](StatusCodeAndOtherOperations...)

	return OperationsFilter{
		operations: allowed,
		allowAll:   allowed.Contains(AllowAllOperations),
	}
}
