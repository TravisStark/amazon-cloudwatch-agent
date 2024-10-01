// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package containerinsightsjmx

import (
	"github.com/aws/amazon-cloudwatch-agent/cfg/envconfig"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/exporter/debug"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/processor/jmxfilterprocessor"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/processor/jmxtransformprocessor"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/receiver/otlp"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

const (
	pipelineName = "containerinsightsjmx"
)

type translator struct {
}

var _ common.Translator[*common.ComponentTranslators] = (*translator)(nil)

func NewTranslator() common.Translator[*common.ComponentTranslators] {
	return &translator{}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(component.DataTypeMetrics, pipelineName)
}

// Translate creates a pipeline for container insights if the logs.metrics_collected.ecs or logs.metrics_collected.kubernetes
// section is present.
func (t *translator) Translate(conf *confmap.Conf) (*common.ComponentTranslators, error) {
	//Checking if container and jmx is configured
	if conf == nil || !envconfig.IsRunningInContainer() || !isJMXConfigured(conf) {
		return nil, nil
	}
	translators := common.ComponentTranslators{
		Receivers:  common.NewTranslatorMap[component.Config](),
		Processors: common.NewTranslatorMap[component.Config](),
		Exporters:  common.NewTranslatorMap[component.Config](),
		Extensions: common.NewTranslatorMap[component.Config](),
	}

	translators.Receivers.Set(otlp.NewTranslatorWithName(common.JmxKey))

	translators.Processors.Set(jmxfilterprocessor.NewTranslatorWithName(common.JmxKey))
	translators.Processors.Set(jmxtransformprocessor.NewTranslatorWithName(common.JmxKey))
	translators.Exporters.Set(debug.NewTranslatorWithName(common.JmxKey)) //this might need to be changed?

	return &translators, nil

}

func isJMXConfigured(conf *confmap.Conf) bool {
	boolean, _ := common.GetBool(conf, common.JmxConfigKey)
	return boolean
}
