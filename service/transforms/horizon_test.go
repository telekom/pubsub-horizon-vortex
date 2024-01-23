// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package transforms_test

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
	"vortex/service/transforms"
)

var (
	testMessage        = createWorkingCopy(KafkaMessage)
	droppedTestMessage = createWorkingCopy(KafkaDroppedMessage)
)

func TestRenameAdditionalFields(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.RenameAdditionalFields()

	_, ok := testMessage["properties"]
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.False(ok, "expected field 'properties' to not exist yet")

	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	_, ok = transformed["properties"]
	assertions.True(ok, "expected field 'properties' to exist")
}

func TestAddEventUnderscoreIdField(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.AddEventUnderscoreIdField()

	var event, ok = testMessage["event"].(map[string]interface{})
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.True(ok, "expected field 'event' to exist")

	_, ok = event["_id"]
	assertions.False(ok, "expected field '_id' to not exist yet")

	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	transformedEvent, ok := transformed["event"].(map[string]any)
	_, ok = transformedEvent["_id"]
	assertions.True(ok, "expected field '_id' to exist")
}

func TestMoveTimestamp(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.MoveTimestamp()

	var event, ok = testMessage["event"].(map[string]interface{})
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.True(ok, "expected field 'event' to exist")

	_, ok = event["time"]
	assertions.True(ok, "expected field 'time' to exist in 'event'")

	_, ok = testMessage["timestamp"]
	assertions.False(ok, "expected field 'timestamp' to not exist yet")

	var transformed, err = transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	_, ok = transformed["timestamp"]
	assertions.True(ok, "expected field 'timestamp' to exist")
}

func TestUpdateModifiedTime(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.UpdateModifiedTime()

	_, ok := testMessage["modified"]
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.False(ok, "expected field 'modified' to not exist yet")

	_, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")

	_, ok = testMessage["modified"]
	assertions.True(ok, "expected field 'modified' to exist")

	lastModified := testMessage["modified"]
	time.Sleep(1 * time.Second)
	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)
	assertions.NotEqual(lastModified, transformed["modified"], "expected field 'modified' to be different from its previous value")
}

func TestAddTimestampIfDropped(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.AddTimestampIfDropped()

	_, ok := droppedTestMessage["timestamp"]
	log.Printf("droppedTestMessage: %+v\n", droppedTestMessage)
	assertions.False(ok, "expected field 'timestamp' to not exist yet")

	transformed, err := transformFunc(droppedTestMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	_, ok = transformed["timestamp"]
	assertions.True(ok, "expected field 'timestamp' to exist")
}

func TestEnrichPropertiesFromHttpHeaders(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.EnrichPropertiesFromHttpHeaders()

	_, ok := testMessage["httpHeaders"]
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.True(ok, "expected field 'httpHeaders' to exist")

	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	properties, ok := transformed["properties"].(map[string]any)
	assertions.True(ok, "expected field 'properties' to exist")

	businessContext, ok := properties["x-business-context"]
	assertions.True(ok, "expected field 'x-business-context' to exist")
	assertions.Equal("somecontext", businessContext, "expected field 'x-business-context' to be 'somecontext'")

	correlationId, ok := properties["x-correlation-id"]
	assertions.True(ok, "expected field 'x-correlation-id' to exist")
	assertions.Equal("somecorrelation", correlationId, "expected field 'x-correlation-id' to be 'somecorrelation'")
}

func TestDropHttpHeaders(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.DropHttpHeaders()

	_, ok := testMessage["httpHeaders"]
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.True(ok, "expected field 'httpHeaders' to exist")

	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	_, ok = transformed["httpHeaders"]
	assertions.False(ok, "expected field 'httpHeaders' to not exist")
}

func TestDropEventData(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.DropEventData()

	_, ok := testMessage["event"].(map[string]any)["data"]
	log.Printf("testMessage: %+v\n", testMessage)
	assertions.True(ok, "expected field 'event' to exist")

	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	_, ok = transformed["event"].(map[string]any)["data"]
	assertions.False(ok, "expected field 'event' to not exist")
}

func TestFlatten(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.Flatten()

	log.Printf("testMessage: %+v\n", testMessage)
	transformed, err := transformFunc(testMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	for key, value := range transformed {
		_, ok := value.(map[string]any)
		assertions.False(ok, "expected value of field '%s' to not be a map", key)
	}
}

func TestDeleteFlatKeys(t *testing.T) {
	var assertions = assert.New(t)
	var transformFunc = transforms.DeleteFlatKeys("event.id")
	var flatFunc = transforms.Flatten()
	var flattenedMessage, _ = flatFunc(testMessage)

	log.Printf("testMessage: %+v\n", testMessage)
	log.Printf("flattenedMessage: %+v\n", flattenedMessage)

	_, ok := flattenedMessage["event.id"]
	assertions.True(ok, "expected field 'event.id' to exist")

	transformed, err := transformFunc(flattenedMessage)
	assertions.Nil(err, "expected no error")
	log.Printf("transformed: %+v\n", transformed)

	_, ok = transformed["event.id"]
	assertions.False(ok, "expected field 'event.id' to not exist")
}
