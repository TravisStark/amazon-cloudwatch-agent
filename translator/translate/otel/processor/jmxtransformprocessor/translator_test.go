package jmxtransformprocessor

import (
	"fmt"
	"github.com/aws/amazon-cloudwatch-agent/internal/util/testutil"
	"github.com/aws/amazon-cloudwatch-agent/translator/translate/otel/common"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"go.opentelemetry.io/collector/component"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap"
)

func TestTranslator(t *testing.T) {
	factory := transformprocessor.NewFactory()

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
				gotCfg, ok := got.(*transformprocessor.Config)
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

				//This gets the list of statements we got from translate
				gotMetricContextStatements, ok := gotMap["metricstatements"].([]interface{})[0].(map[string]interface{})["statements"].([]interface{})
				gotDatapointContextStatement, ok := gotMap["metricstatements"].([]interface{})[1].(map[string]interface{})["statements"].([]interface{})

				wantMetricContextStatements, ok := wantMap["metricstatements"].([]interface{})[0].(map[string]interface{})["statements"].([]interface{})
				wantDatapointContextStatement, ok := wantMap["metricstatements"].([]interface{})[1].(map[string]interface{})["statements"].([]interface{})

				// Compare 'nameList' with another list
				if !containsSameElements(gotMetricContextStatements, wantMetricContextStatements) {
					t.Fatal("statements in 'metricstatements[]' do not match")
				}
				if !containsSameElements(gotDatapointContextStatement, wantDatapointContextStatement) {
					t.Fatal("statements in 'metricstatements[]' do not match")
				}

			}
		})
	}
}

func containsSameElements(list1, list2 []interface{}) bool {
	if len(list1) != len(list2) {
		fmt.Printf("Length mismatch: list1 has %d elements, list2 has %d elements\n", len(list1), len(list2))
		return false
	}

	map1 := make(map[interface{}]int)
	map2 := make(map[interface{}]int)

	for _, item := range list1 {
		map1[item]++
	}

	for _, item := range list2 {
		map2[item]++
	}

	// Compare maps and print differences
	mismatch := false
	for key, count1 := range map1 {
		count2, exists := map2[key]
		if !exists {
			fmt.Printf("Key '%v' is missing in list2\n", key)
			mismatch = true
		} else if count1 != count2 {
			fmt.Printf("Key '%v' count mismatch: list1 has %d, list2 has %d\n", key, count1, count2)
			mismatch = true
		}
	}

	for key := range map2 {
		if _, exists := map1[key]; !exists {
			fmt.Printf("Key '%v' is missing in list1\n", key)
			mismatch = true
		}
	}

	return !mismatch
}
