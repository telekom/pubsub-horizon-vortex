package transforms_test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"slices"
	"testing"
	"vortex/service/transforms"
)

var (
	KafkaMessage        = mustReadJson("../../testdata/kafka_msg.json")
	KafkaDroppedMessage = mustReadJson("../../testdata/kafka_dropped_msg.json")
	TransformedMessage  = mustReadJson("../../testdata/transformed_msg.json")
)

func TestRegistry_ApplyTransforms(t *testing.T) {
	var assertions = assert.New(t)
	var original = createWorkingCopy(KafkaMessage)
	var ignoreProperties = []string{
		"modified",
	}

	transformed, err := transforms.GlobalRegistry.ApplyTransforms(original)
	assertions.Nil(err, "expected no error")

	for key, value := range transformed {
		t.Run(key, func(t *testing.T) {
			if slices.Contains(ignoreProperties, key) {
				assert.NotEmptyf(t, key, "expected key '%s' to not be empty", key)
			} else {
				log.Printf("expected: '%+v' actual: '%+v'\n", TransformedMessage[key], value)
				assert.Equal(t, TransformedMessage[key], value, "expected value to match")
			}
		})
	}
}

func createWorkingCopy(data map[string]any) map[string]any {
	var copy = make(map[string]any)
	for k, v := range data {
		copy[k] = v
	}

	return copy
}

func mustReadJson(filename string) map[string]any {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	var data = make(map[string]any)
	if err := json.Unmarshal(bytes, &data); err != nil {
		panic(err)
	}

	return data
}
