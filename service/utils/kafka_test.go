package utils_test

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"vortex/service/utils"
)

func TestGetFieldsFromMessage(t *testing.T) {
	var assertions = assert.New(t)
	var kafkaMessage = mustReadJson("../../testdata/kafka_msg.json")
	var dummy = &sarama.ConsumerMessage{}
	dummy.Key = []byte(kafkaMessage["uuid"].(string))
	dummy.Topic = "subscribed"

	if bytes, err := json.Marshal(kafkaMessage); err != nil {
		assertions.Nil(err, "expected no error")
		dummy.Value = bytes
	}

	dummy.Offset = 1
	dummy.Partition = 0
	dummy.Headers = []*sarama.RecordHeader{
		{Key: []byte("type"), Value: []byte("MESSAGE")},
	}

	var expected = map[string]any{
		"topic":     "subscribed",
		"offset":    int64(1),
		"partition": int32(0),
		"key":       kafkaMessage["uuid"].(string),
		"type":      "MESSAGE",
	}

	var actual = utils.GetFieldsFromMessage(dummy)
	assertions.Equal(expected, actual, "expected fields to be equal")
}

func TestGetHeader(t *testing.T) {
	var assertions = assert.New(t)
	var dummy = []*sarama.RecordHeader{
		{Key: []byte("type"), Value: []byte("dummy")},
	}

	var expected = "dummy"
	var actual = utils.GetHeader(dummy, "type")

	assertions.Equal(expected, actual)
}

func TestGetFieldsFromClaims(t *testing.T) {
	var assertions = assert.New(t)
	var dummy = map[string][]int32{
		"dummy": {0, 1, 2},
	}

	var expected = map[string]any{"dummy.partitions": "0 1 2"}
	var actual = utils.GetFieldsFromClaims(dummy)

	assertions.Equal(expected, actual)
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
