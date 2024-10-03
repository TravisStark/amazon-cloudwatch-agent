// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package jmxattributeprocessor

import (
	"fmt"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
)

type translator struct {
	name    string
	factory processor.Factory
}

type Option func(any)

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslatorWithName(name string) common.Translator[component.Config] {
	return &translator{name: name, factory: attributesprocessor.NewFactory()}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	if conf == nil || !conf.IsSet(common.ContainerInsightsConfigKey) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: common.ContainerInsightsConfigKey}
	}

	includeMetricNames := []string{}
	cfg := t.factory.CreateDefaultConfig().(*attributesprocessor.Config)
	switch t.name {
	case "General":
		includeMetricNames = []string{
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
		}
	case "JvmMemoryBytesUsed":
		includeMetricNames = []string{
			"jvm.memory.bytes.used",
		}
	case "JvmMemoryPoolBytesUsed":
		includeMetricNames = []string{
			"jvm_memory_pool_bytes_used",
		}
	default:
		includeMetricNames = []string{}
	}

	//general include metrics
	c := confmap.NewFromStringMap(map[string]interface{}{
		"include": map[string]interface{}{
			"metric_names": includeMetricNames,
			"attributes": []map[string]interface{}{
				{
					"key": "ClusterName",
				},
				{
					"key": "Namespace",
				},
			},
		},
	})

	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal attribute processor (%s): %w", t.ID(), err)
	}

	return cfg, nil
}
