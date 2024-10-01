// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package jmxtransformprocessor

import (
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

// Mapping from JMX metrics to transform processor metrics
var metricMapping = map[string]string{
	"jvm.operating.system.total.swap.space.size":      "java_lang_operatingsystem_totalswapspacesize",
	"jvm.operating.system.system.cpu.load":            "java_lang_operatingsystem_systemcpuload",
	"jvm.operating.system.process.cpu.load":           "java_lang_operatingsystem_processcpuload",
	"jvm.operating.system.free.swap.space.size":       "java_lang_operatingsystem_freeswapspacesize",
	"jvm.operating.system.total.physical.memory.size": "java_lang_operatingsystem_totalphysicalmemorysize",
	"jvm.operating.system.free.physical.memory.size":  "java_lang_operatingsystem_freephysicalmemorysize",
	"jvm.operating.system.open.file.descriptor.count": "java_lang_operatingsystem_openfiledescriptorcount",
	"jvm.operating.system.available.processors":       "java_lang_operatingsystem_availableprocessors",
	"jvm.threads.count":                               "jvm_threads_current",
	"jvm.threads.daemon":                              "jvm_threads_daemon",
	"catalina_manager_activesessions":                 "catalina_manager_activesessions",
	"catalina_manager_rejectedsessions":               "catalina_manager_rejectedsessions",
	"catalina_globalrequestprocessor_bytesreceived":   "catalina_globalrequestprocessor_bytesreceived",
	"catalina_globalrequestprocessor_bytessent":       "catalina_globalrequestprocessor_bytessent",
	"catalina_globalrequestprocessor_requestcount":    "catalina_globalrequestprocessor_requestcount",
	"catalina_globalrequestprocessor_errorcount":      "catalina_globalrequestprocessor_errorcount",
	"catalina_globalrequestprocessor_processingtime":  "catalina_globalrequestprocessor_processingtime",
	"jvm.classes.loaded":                              "jvm_classes_loaded",
	"jvm.memory.bytes.used":                           "jvm_memory_bytes_used",
	"jvm.memory.pool.bytes.used":                      "jvm_memory_pool_bytes_used",
}

type translator struct {
	name    string
	factory processor.Factory
}

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslatorWithName(name string) common.Translator[component.Config] {
	return &translator{name, metricstransformprocessor.NewFactory()}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	cfg := t.factory.CreateDefaultConfig().(*metricstransformprocessor.Config)
	transformRules := []map[string]interface{}{}

	for oldName, newName := range metricMapping {
		transformRules = append(transformRules, map[string]interface{}{
			"include":  oldName,
			"action":   "update",
			"new_name": newName,
		})
	}

	c := confmap.NewFromStringMap(map[string]interface{}{
		"transforms": transformRules,
	})
	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal into metricstransform config: %w", err)
	}

	return cfg, nil
}
