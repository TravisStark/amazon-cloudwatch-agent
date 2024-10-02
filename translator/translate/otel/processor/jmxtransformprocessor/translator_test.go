package jmxtransformprocessor

import (
	"github.com/aws/amazon-cloudwatch-agent/internal/util/testutil"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"go.opentelemetry.io/collector/component"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"reflect"
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
		"NoContainerInsights": {
			input: map[string]any{},
			wantErr: &common.MissingKeyError{
				ID:      component.NewIDWithName(factory.Type(), "jmx"),
				JsonKey: common.ContainerInsightsConfigKey,
			},
		},
		"WithContainerInsights": {
			input:  testutil.GetJson(t, filepath.Join("testdata", "config.json")),
			index:  0,
			wantID: "filter/jmx",
			want:   testutil.GetConf(t, filepath.Join("testdata", "config.yaml")),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			tt := NewTranslatorWithName("jmx")
			//require.EqualValues(t, testCase.wantID, tt.ID().String())

			conf := confmap.NewFromStringMap(testCase.input)
			got, err := tt.Translate(conf)
			require.Equal(t, testCase.wantErr, err)

			if err == nil {
				require.NotNil(t, got)
				gotCfg, ok := got.(*metricstransformprocessor.Config)
				require.True(t, ok)

				wantCfg := factory.CreateDefaultConfig()
				require.NoError(t, testCase.want.Unmarshal(wantCfg))

				// Convert gotCfg to YAML and unmarshal into a map
				gotYAML, err := yaml.Marshal(gotCfg)
				require.NoError(t, err)

				var gotMap map[string]interface{}
				err = yaml.Unmarshal(gotYAML, &gotMap)
				require.NoError(t, err)

				// Convert wantCfg to YAML and unmarshal into a map
				wantYAML, err := yaml.Marshal(wantCfg)
				require.NoError(t, err)

				var wantMap map[string]interface{}
				err = yaml.Unmarshal(wantYAML, &wantMap)
				require.NoError(t, err)

				// Compare the maps
				require.True(t, reflect.DeepEqual(gotMap, wantMap), "YAML contents do not match")

			}
		})
	}
}
