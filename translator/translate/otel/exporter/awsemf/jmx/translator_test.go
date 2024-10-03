package awsemfjmx

import (
	"github.com/aws/amazon-cloudwatch-agent/cfg/envconfig"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/agent"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsemfexporter"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/confmap"
	"testing"
)

func TestTranslateAWSEMF(t *testing.T) {
	translator := NewTranslatorWithName("awsemfjmx")
	t.Setenv(envconfig.AWS_CA_BUNDLE, "/ca/bundle")
	agent.Global_Config.Region = "us-east-1"
	agent.Global_Config.Role_arn = "global_arn"
	t.Setenv(envconfig.IMDS_NUMBER_RETRY, "0")

	// Sample input configuration map
	input := map[string]interface{}{
		"logs": map[string]interface{}{
			"metrics_collected": map[string]interface{}{
				"kubernetes": map[string]interface{}{
					"metric_namespace": "ContainerInsights",
					"metric_declaration": []map[string]interface{}{
						{
							"dimensions": []interface{}{
								[]string{"ClusterName", "Namespace"},
							},
							"metric_name_selectors": []string{
								"jvm_threads_daemon",
								"catalina_globalrequestprocessor_bytessent",
								"jvm_classes_loaded",
							},
						},
						{
							"dimensions": []interface{}{
								[]string{"ClusterName", "Namespace", "area"},
							},
							"metric_name_selectors": []string{
								"jvm_memory_bytes_used",
							},
						},
					},
				},
			},
		},
	}

	// Convert input to confmap.Conf
	conf := confmap.NewFromStringMap(input)

	// Translate configuration using the translator
	got, err := translator.Translate(conf)

	// Assert no error occurred during translation
	if err != nil {
		t.Fatalf("Unexpected error during translation: %v", err)
	}

	// Expected output configuration
	expectedConfig := &awsemfexporter.Config{
		Namespace:             "ContainerInsights",
		DimensionRollupOption: "ZeroAndSingleDimensionRollup",
		OutputDestination:     "cloudwatch",
		Version:               "1",
		MetricDeclarations: []*awsemfexporter.MetricDeclaration{
			{
				Dimensions: [][]string{
					{"ClusterName", "Namespace"},
				},
				MetricNameSelectors: []string{
					"jvm_threads_daemon",
					"catalina_globalrequestprocessor_bytessent",
					"jvm_classes_loaded",
				},
				LabelMatchers: nil, // Assuming no label matchers in this example
			},
			{
				Dimensions: [][]string{
					{"ClusterName", "Namespace", "area"},
				},
				MetricNameSelectors: []string{
					"jvm_memory_bytes_used",
				},
				LabelMatchers: nil, // Assuming no label matchers in this example
			},
		},
	}

	// Assert the translated configuration matches the expected configuration
	if gotCfg, ok := got.(*awsemfexporter.Config); ok {
		if !compareConfigs(t, gotCfg, expectedConfig) {
			t.Errorf("Translated configuration does not match expected configuration")
		}
	} else {
		t.Fatalf("Translated config is not of type *awsemfexporter.Config")
	}
}

// compareConfigs compares two awsemfexporter.Config instances and logs the differences.
func compareConfigs(t *testing.T, gotCfg, wantCfg *awsemfexporter.Config) bool {

	assert.Equal(t, wantCfg.Namespace, gotCfg.Namespace)
	assert.Equal(t, wantCfg.LogGroupName, gotCfg.LogGroupName)
	assert.Equal(t, wantCfg.LogStreamName, gotCfg.LogStreamName)
	assert.Equal(t, wantCfg.DimensionRollupOption, gotCfg.DimensionRollupOption)
	assert.Equal(t, wantCfg.DisableMetricExtraction, gotCfg.DisableMetricExtraction)
	assert.Equal(t, wantCfg.EnhancedContainerInsights, gotCfg.EnhancedContainerInsights)
	assert.Equal(t, wantCfg.ParseJSONEncodedAttributeValues, gotCfg.ParseJSONEncodedAttributeValues)
	assert.Equal(t, wantCfg.OutputDestination, gotCfg.OutputDestination)
	assert.Equal(t, wantCfg.EKSFargateContainerInsightsEnabled, gotCfg.EKSFargateContainerInsightsEnabled)
	assert.Equal(t, wantCfg.ResourceToTelemetrySettings, gotCfg.ResourceToTelemetrySettings)
	assert.ElementsMatch(t, wantCfg.MetricDeclarations, gotCfg.MetricDeclarations)
	assert.ElementsMatch(t, wantCfg.MetricDescriptors, gotCfg.MetricDescriptors)
	assert.Equal(t, wantCfg.LocalMode, gotCfg.LocalMode)
	assert.Equal(t, "/ca/bundle", gotCfg.CertificateFilePath)
	assert.Equal(t, "global_arn", gotCfg.RoleARN)
	assert.Equal(t, "us-east-1", gotCfg.Region)
	assert.NotNil(t, gotCfg.MiddlewareID)
	assert.Equal(t, "agenthealth/logs", gotCfg.MiddlewareID.String())

	return true
}
