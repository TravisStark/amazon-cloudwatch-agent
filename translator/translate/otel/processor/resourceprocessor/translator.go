// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package resourceprocessor

import (
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/processor"

	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
)

type translator struct {
	name    string
	factory processor.Factory
}

var (
	baseKey = common.ConfigKey(common.LogsKey, common.MetricsCollectedKey)
	eksKey  = common.ConfigKey(baseKey, common.KubernetesKey)
)

type Option func(any)

func WithName(name string) Option {
	return func(a any) {
		if t, ok := a.(*translator); ok {
			t.name = name
		}
	}
}

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslator(opts ...Option) common.Translator[component.Config] {
	t := &translator{factory: resourceprocessor.NewFactory()}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

var _ common.Translator[component.Config] = (*translator)(nil)

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

// Translate creates a processor config based on the fields in the
// Metrics section of the JSON config.
func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	if conf == nil || !conf.IsSet(common.JmxConfigKey) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: common.JmxConfigKey}
	}

	cfg := t.factory.CreateDefaultConfig().(*resourceprocessor.Config)

	c := confmap.NewFromStringMap(map[string]any{})
	if t.name == "jmxResource" {
		clusterName, ok := common.GetString(conf, common.ConfigKey(eksKey, "cluster_name"))

		if ok {
			c = confmap.NewFromStringMap(map[string]any{
				//from config
				"attributes": []any{
					map[string]any{
						"key":            "Namespace",
						"from_attribute": "k8s.namespace.name",
						"action":         "insert",
					},
					map[string]any{
						"key":    "ClusterName",
						"value":  clusterName,
						"action": "upsert",
					},
					map[string]any{
						"key":            "NodeName",
						"from_attribute": "host.name",
						"action":         "insert",
					},
				},
			})
		} else {
			c = confmap.NewFromStringMap(map[string]any{
				//from resource detection processor
				"attributes": []any{
					map[string]any{
						"key":            "ClusterName",
						"from_attribute": "k8s.cluster.name",
						"action":         "insert",
					},
					map[string]any{
						"key":            "Namespace",
						"from_attribute": "k8s.namespace.name",
						"action":         "insert",
					},
					map[string]any{
						"key":            "NodeName",
						"from_attribute": "host.name",
						"action":         "insert",
					},
				},
			})
		}

	} else {
		c = confmap.NewFromStringMap(map[string]any{
			"attributes": []any{
				map[string]any{
					"action":  "delete",
					"pattern": "telemetry.sdk.*",
				},
				map[string]any{
					"action": "delete",
					"key":    "service.name",
					"value":  "unknown_service:java",
				},
			},
		})
	}

	if err := c.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal resource processor: %w", err)
	}

	return cfg, nil
}
