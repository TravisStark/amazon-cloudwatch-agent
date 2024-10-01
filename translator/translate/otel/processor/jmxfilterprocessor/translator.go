// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package jmxfilterprocessor

import (
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

const (
	matchTypeStrict = "strict"
)

type translator struct {
	name    string
	factory processor.Factory
}

type Option func(any)

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslatorWithName(name string) common.Translator[component.Config] {
	return &translator{name: name, factory: filterprocessor.NewFactory()}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	if conf == nil || !conf.IsSet(common.JmxConfigKey) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: common.JmxConfigKey}
	}

	cfg := t.factory.CreateDefaultConfig().(*filterprocessor.Config)

	allowedMetrics := map[string]bool{
		"jvm.classes.loaded":                true,
		"jvm.gc.collections.count":          true,
		"jvm.gc.collections.elapsed":        true,
		"jvm.memory.heap.init":              true,
		"jvm.memory.heap.used":              true,
		"jvm.memory.heap.committed":         true,
		"jvm.memory.heap.max":               true,
		"jvm.memory.nonheap.init":           true,
		"jvm.memory.nonheap.used":           true,
		"jvm.memory.nonheap.committed":      true,
		"jvm.memory.nonheap.max":            true,
		"jvm.memory.pool.init":              true,
		"jvm.memory.pool.used":              true,
		"jvm.memory.pool.committed":         true,
		"jvm.memory.pool.max":               true,
		"jvm.os.total.swap.space.size":      true,
		"jvm.os.system.cpu.load":            true,
		"jvm.os.process.cpu.load":           true,
		"jvm.os.free.swap.space.size":       true,
		"jvm.os.total.physical.memory.size": true,
		"jvm.os.free.physical.memory.size":  true,
		"jvm.os.open.file.descriptor.count": true,
		"jvm.os.available.processors":       true,
		"jvm.threads.count":                 true,
		"jvm.threads.daemon":                true,
		"tomcat.sessions":                   true,
		"tomcat.rejected_sessions":          true,
		"tomcat.errors":                     true,
		"tomcat.request_count":              true,
		"tomcat.max_time":                   true,
		"tomcat.processing_time":            true,
		"tomcat.traffic.received":           true,
		"tomcat.traffic.sent":               true,
		"tomcat.threads.idle":               true,
		"tomcat.threads.busy":               true,
	}

	includeMetricNames := []string{}

	jmxMap := common.GetIndexedMap(conf, common.JmxConfigKey, 0)

	//jvm metrics
	if jvmMap, ok := jmxMap["jvm"].(map[string]any); ok {
		if measurements, ok := jvmMap["measurement"].([]interface{}); ok {
			for _, measurement := range measurements {
				if metricName, ok := measurement.(string); ok {
					if _, allowed := allowedMetrics[metricName]; allowed {
						includeMetricNames = append(includeMetricNames, metricName)
					}
				}
			}
		}
	}

	//tomcat metrics
	if tomcatMap, ok := jmxMap["tomcat"].(map[string]any); ok {
		if measurements, ok := tomcatMap["measurement"].([]interface{}); ok {
			for _, measurement := range measurements {
				if metricName, ok := measurement.(string); ok {
					if _, allowed := allowedMetrics[metricName]; allowed {
						includeMetricNames = append(includeMetricNames, metricName)
					}
				}
			}
		}
	}

	c := confmap.NewFromStringMap(map[string]interface{}{
		"metrics": map[string]any{
			"include": map[string]any{
				"match_type":   matchTypeStrict,
				"metric_names": includeMetricNames,
			},
		},
	})

	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal jmx filter processor (%s): %w", t.ID(), err)
	}

	return cfg, nil
}
