// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package jmxtransformprocessor

import (
	_ "embed"
	"fmt"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

//go:embed testdata/config.yaml
var transformJmxConfig string

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
	if !(conf != nil && conf.IsSet(common.ContainerInsightsConfigKey)) {
		return nil, nil //returning nil for now !!!!!!!
	}
	cfg := t.factory.CreateDefaultConfig().(*transformprocessor.Config)
	clusterName := conf.Get(common.ConfigKey(common.ContainerInsightsConfigKey, "cluster_name"))

	if clusterName == nil {
		return common.GetYamlFileToYamlConfig(cfg, transformJmxConfig)
	}
	c := confmap.NewFromStringMap(map[string]interface{}{
		"metric_statements": map[string]any{
			"context": "resource",
			"statements": []string{
				"keep_keys(attributes, [\"ClusterName\", \"Namespace\"])",
				fmt.Sprintf("set(attributes[\"ClusterName\"], \"%s\")", clusterName),
			},
		},
	})

	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal jmx filter processor (%s): %w", t.ID(), err)
	}

	return cfg, nil
}
