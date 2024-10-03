// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package jmxattributeprocessor

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
)

func TestTranslator(t *testing.T) {
	factory := attributesprocessor.NewFactory()
	testCases := map[string]struct {
		input   map[string]any
		name    string
		wantID  string
		want    *confmap.Conf
		wantErr error
	}{
		"ConfigWithNoContainerInsights": {
			input:  map[string]any{},
			name:   "",
			wantID: "attributes",
			wantErr: &common.MissingKeyError{
				ID:      component.NewIDWithName(factory.Type(), ""),
				JsonKey: common.ContainerInsightsConfigKey,
			},
		},
		"ConfigWithValidGeneralConfig": {
			input: map[string]any{
				common.ContainerInsightsConfigKey: true,
			},
			name:   "General",
			wantID: "attributes/General",
			want: confmap.NewFromStringMap(map[string]any{
				"include": map[string]any{
					"metric_names": []any{
						"jvm.classes.loaded",
						"jvm.memory.bytes.used",
						"jvm.memory.pool.bytes.used",
						"jvm.operating.system.total.swap.space.size",
						"jvm.operating.system.system.cpu.load",
						"jvm.operating.system.process.cpu.load",
						"jvm.operating.system.free.swap.space.size",
						"jvm.operating.system.total.physical.memory.size",
						"jvm.operating.system.free.physical.memory.size",
						"jvm.operating.system.open.file.descriptor.count",
						"jvm.operating.system.available.processors",
						"jvm.threads.count",
						"jvm.threads.daemon",
						"tomcat.sessions",
						"tomcat.rejected_sessions",
						"tomcat.traffic.received",
						"tomcat.traffic.sent",
						"tomcat.request_count",
						"tomcat.errors",
						"tomcat.processing_time",
					},
					"attributes": []map[string]any{
						{"key": "ClusterName"},
						{"key": "Namespace"},
					},
				},
			}),
		},
		"ConfigWithValidJvmMemoryBytesUsedConfig": {
			input: map[string]any{
				common.ContainerInsightsConfigKey: true,
			},
			name:   "JvmMemoryBytesUsed",
			wantID: "attributes/JvmMemoryBytesUsed",
			want: confmap.NewFromStringMap(map[string]any{
				"include": map[string]any{
					"metric_names": []any{"jvm.memory.bytes.used"},
					"attributes": []map[string]any{
						{"key": "ClusterName"},
						{"key": "Namespace"},
					},
				},
			}),
		},
		"ConfigWithValidJvmMemoryPoolBytesUsedConfig": {
			input: map[string]any{
				common.ContainerInsightsConfigKey: true,
			},
			name:   "JvmMemoryPoolBytesUsed",
			wantID: "attributes/JvmMemoryPoolBytesUsed",
			want: confmap.NewFromStringMap(map[string]any{
				"include": map[string]any{
					"metric_names": []any{"jvm_memory_pool_bytes_used"},
					"attributes": []map[string]any{
						{"key": "ClusterName"},
						{"key": "Namespace"},
					},
				},
				"actions":
			}),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tt := NewTranslatorWithName(testCase.name)
			require.EqualValues(t, testCase.wantID, tt.ID().String())
			conf := confmap.NewFromStringMap(testCase.input)
			got, err := tt.Translate(conf)
			require.Equal(t, testCase.wantErr, err)
			if err == nil {
				require.NotNil(t, got)
				gotCfg, ok := got.(*attributesprocessor.Config)
				require.True(t, ok)
				wantCfg := factory.CreateDefaultConfig()
				require.NoError(t, testCase.want.Unmarshal(wantCfg))
				require.Equal(t, wantCfg, gotCfg)
			}
		})
	}
}
