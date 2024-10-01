package jmxtransformprocessor

import (
	"fmt"
	"github.com/aws/amazon-cloudwatch-agent/internal/util/testutil"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"go.opentelemetry.io/collector/component"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestTranslator(t *testing.T) {
	factory := metricstransformprocessor.NewFactory()

	testCases := map[string]struct {
		translator common.Translator[component.Config]
		input      map[string]any
		index      int
		wantID     string
		want       *confmap.Conf
		wantErr    error
	}{
		"TransformMetrics": {
			input: map[string]any{
				"metrics": map[string]any{
					"jvm.operating.system.total.swap.space.size": "value",
					"jvm.threads.count":                          "value",
					"catalina_manager_activesessions":            "value",
				},
			},
			want: confmap.NewFromStringMap(map[string]interface{}{
				"transforms": []map[string]interface{}{
					{
						"include":  "jvm.operating.system.total.swap.space.size",
						"action":   "update",
						"new_name": "java_lang_operatingsystem_totalswapspacesize",
					},
					{
						"include":  "jvm.threads.count",
						"action":   "update",
						"new_name": "jvm_threads_current",
					},
					{
						"include":  "catalina_manager_activesessions",
						"action":   "update",
						"new_name": "catalina_manager_activesessions",
					},
				},
			}),
		},
		"EmptyMetricsSection": {
			input: map[string]any{
				"metrics": map[string]any{},
			},
			want: confmap.NewFromStringMap(map[string]interface{}{
				"transforms": []map[string]interface{}{},
			}),
		},
		"InvalidConfig": {
			input: map[string]any{
				"metrics": "invalid_structure",
			},
			wantErr: fmt.Errorf("unable to unmarshal into metricstransform config"),
		},
		"WithCompleteConfig": {
			input:  testutil.GetJson(t, filepath.Join("testdata", "config.json")),
			index:  0,
			wantID: "filter/jmx",
			want:   testutil.GetConf(t, filepath.Join("testdata", "config.yaml")),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tt := NewTranslatorWithName("jmx")
			require.EqualValues(t, testCase.wantID, tt.ID().String())

			conf := confmap.NewFromStringMap(testCase.input)
			got, err := tt.Translate(conf)
			require.Equal(t, testCase.wantErr, err)

			if err == nil {
				require.NotNil(t, got)

				gotCfg, ok := got.(*filterprocessor.Config)
				require.True(t, ok)

				wantCfg := factory.CreateDefaultConfig()
				require.NoError(t, testCase.want.Unmarshal(wantCfg))

				gotYAML, err := yaml.Marshal(gotCfg)
				require.NoError(t, err)

				wantYAML, err := yaml.Marshal(wantCfg)
				require.NoError(t, err)

				require.Equal(t, wantCfg, gotCfg, "Expected:\n%s\nGot:\n%s", wantYAML, gotYAML)
			}
		})
	}
}
