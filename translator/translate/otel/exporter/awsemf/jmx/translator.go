// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT

package awsemfjmx

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsemfexporter"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/exporter"
	"gopkg.in/yaml.v3"

	"github.com/aws/amazon-cloudwatch-agent/cfg/envconfig"
	"github.com/aws/amazon-cloudwatch-agent/internal/retryer"
	"github.com/aws/amazon-cloudwatch-agent/translator/config"
	"github.com/aws/amazon-cloudwatch-agent/translator/context"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/agent"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/extension/agenthealth"
)

//go:embed awsemfjmx_config.yaml
var defaultKubernetesConfig string

var (
	endpointOverrideKey = common.ConfigKey(common.LogsKey, common.EndpointOverrideKey)
	roleARNPathKey      = common.ConfigKey(common.LogsKey, common.CredentialsKey, common.RoleARNKey)
)

type translator struct {
	name    string
	factory exporter.Factory
}

var _ common.Translator[component.Config] = (*translator)(nil)

func NewTranslatorWithName(name string) common.Translator[component.Config] {
	return &translator{name, awsemfexporter.NewFactory()}
}

func (t *translator) ID() component.ID {
	return component.NewIDWithName(t.factory.Type(), t.name)
}

// Translate creates an awsemf exporter config based on the input json config
func (t *translator) Translate(conf *confmap.Conf) (component.Config, error) {
	if conf == nil || !conf.IsSet(common.ContainerInsightsConfigKey) {
		return nil, &common.MissingKeyError{ID: t.ID(), JsonKey: common.ContainerInsightsConfigKey}
	}

	cfg := t.factory.CreateDefaultConfig().(*awsemfexporter.Config)
	cfg.MiddlewareID = &agenthealth.LogsID

	var defaultConfig string
	defaultConfig = defaultKubernetesConfig

	if defaultConfig != "" {
		var rawConf map[string]interface{}
		if err := yaml.Unmarshal([]byte(defaultConfig), &rawConf); err != nil {
			return nil, fmt.Errorf("unable to read default config: %w", err)
		}
		conf := confmap.NewFromStringMap(rawConf)
		if err := conf.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("unable to unmarshal config: %w", err)
		}
	}
	cfg.AWSSessionSettings.CertificateFilePath = os.Getenv(envconfig.AWS_CA_BUNDLE)
	if conf.IsSet(endpointOverrideKey) {
		cfg.AWSSessionSettings.Endpoint, _ = common.GetString(conf, endpointOverrideKey)
	}
	cfg.AWSSessionSettings.IMDSRetries = retryer.GetDefaultRetryNumber()
	if profileKey, ok := agent.Global_Config.Credentials[agent.Profile_Key]; ok {
		cfg.AWSSessionSettings.Profile = fmt.Sprintf("%v", profileKey)
	}
	cfg.AWSSessionSettings.Region = agent.Global_Config.Region
	cfg.AWSSessionSettings.RoleARN = agent.Global_Config.Role_arn
	if conf.IsSet(roleARNPathKey) {
		cfg.AWSSessionSettings.RoleARN, _ = common.GetString(conf, roleARNPathKey)
	}
	if credentialsFileKey, ok := agent.Global_Config.Credentials[agent.CredentialsFile_Key]; ok {
		cfg.AWSSessionSettings.SharedCredentialsFile = []string{fmt.Sprintf("%v", credentialsFileKey)}
	}
	if context.CurrentContext().Mode() == config.ModeOnPrem || context.CurrentContext().Mode() == config.ModeOnPremise {
		cfg.AWSSessionSettings.LocalMode = true
	}

	return cfg, nil
}
