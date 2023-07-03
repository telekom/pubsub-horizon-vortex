package utils

import (
	"fmt"
	"github.com/IBM/sarama"
	"strings"
)

func GetFieldsFromMessage(message *sarama.ConsumerMessage) map[string]any {
	var fields = map[string]any{
		"topic":     message.Topic,
		"offset":    message.Offset,
		"partition": message.Partition,
		"key":       string(message.Key),
		"type":      GetHeader(message.Headers, "type"),
	}
	return fields
}

func GetHeader(headers []*sarama.RecordHeader, key string) string {
	for _, header := range headers {
		if string(header.Key) == key {
			return string(header.Value)
		}
	}
	return ""
}

func GetFieldsFromClaims(claims map[string][]int32) map[string]any {
	var fields = make(map[string]any)
	for topic, partitions := range claims {
		var key = fmt.Sprintf("%s.partitions", topic)
		fields[key] = strings.Trim(fmt.Sprintf("%v", partitions), "[]")
	}
	return fields
}
