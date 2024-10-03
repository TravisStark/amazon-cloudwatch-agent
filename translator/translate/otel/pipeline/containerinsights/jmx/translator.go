// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package containerinsightsjmx

import (
	"fmt"
	awsemfjmx "github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/exporter/awsemf/jmx"
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

var (
	baseKey = common.ConfigKey(common.LogsKey, common.MetricsCollectedKey)
	eksKey  = common.ConfigKey(baseKey, common.KubernetesKey)
	ecsKey  = common.ConfigKey(baseKey, common.ECSKey)
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
	if conf == nil || (!conf.IsSet(ecsKey) && !conf.IsSet(eksKey)) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: fmt.Sprint(ecsKey, " or ", eksKey)}
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
	translators.Exporters.Set(awsemfjmx.NewTranslatorWithName(common.JmxKey)) //this might need to be changed?

	return &translators, nil

}
