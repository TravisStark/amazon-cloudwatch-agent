// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package jmxtransformprocessor

import (
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

type translator struct {
	name    string
	factory processor.Factory
}
type Context struct {
	name       string
	statements []string
}

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslatorWithName(name string) common.Translator[component.Config] {
	return &translator{name, transformprocessor.NewFactory()}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	if conf == nil || !conf.IsSet(common.ContainerInsightsConfigKey) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: common.ContainerInsightsConfigKey}

	}
	cfg := t.factory.CreateDefaultConfig().(*transformprocessor.Config)
	contexts := getContexts()
	statementsList := []map[string]interface{}{}
	for _, context := range contexts {
		statementsList = append(statementsList, map[string]interface{}{
			"context":    context.name,
			"statements": context.statements,
		})
	}
	c := confmap.NewFromStringMap(map[string]interface{}{
		"metric_statements": statementsList,
	})

	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal into metricstransform config: %w", err)
	}

	return cfg, nil
}

func getContexts() []Context {
	metricMapping := map[string]string{
		"jvm.classes.loaded":                              "jvm_classes_loaded",
		"jvm.memory.bytes.used":                           "jvm_memory_bytes_used",
		"jvm.memory.pool.bytes.used":                      "jvm_memory_pool_bytes_used",
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
		"tomcat.sessions":                                 "catalina_manager_activesessions",
		"tomcat.rejected_sessions":                        "catalina_manager_rejectedsessions",
		"tomcat.traffic.received":                         "catalina_globalrequestprocessor_bytesreceived",
		"tomcat.traffic.sent":                             "catalina_globalrequestprocessor_bytessent",
		"tomcat.request_count":                            "catalina_globalrequestprocessor_requestcount",
		"tomcat.errors":                                   "catalina_globalrequestprocessor_errorcount",
		"tomcat.processing_time":                          "catalina_globalrequestprocessor_processingtime",
	}

	var contexts []Context
	metricStatements := []string{}
	datapointStatements := []string{}

	for metricName, newName := range metricMapping {
		metricStatement := fmt.Sprintf("set(name, \"%s\") where name == \"%s\"", newName, metricName)
		metricStatements = append(metricStatements, metricStatement)

		// Add keep_keys statements with specific attributes
		var datapointStatement string
		switch newName {
		case "jvm_memory_pool_bytes_used":
			datapointStatement = fmt.Sprintf("keep_keys(attributes, [\"ClusterName\", \"Namespace\", \"pool\"]) where metric.name == \"%s\"", newName)
		case "jvm_memory_bytes_used":
			datapointStatement = fmt.Sprintf("keep_keys(attributes, [\"ClusterName\", \"Namespace\", \"area\"]) where metric.name == \"%s\"", newName)
		default:
			datapointStatement = fmt.Sprintf("keep_keys(attributes, [\"ClusterName\", \"Namespace\"]) where metric.name == \"%s\"", newName)
		}
		datapointStatements = append(datapointStatements, datapointStatement)
	}

	// Create contexts for metrics and datapoints
	contexts = append(contexts, Context{
		name:       "metric",
		statements: metricStatements,
	})

	contexts = append(contexts, Context{
		name:       "datapoint",
		statements: datapointStatements,
	})

	return contexts
}
