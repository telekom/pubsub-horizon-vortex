package transforms

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
	"vortex/service/utils"
)

func RenameAdditionalFields() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		if _, ok := data["additionalFields"]; !ok {
			return data, nil
		}

		data["properties"] = data["additionalFields"]
		delete(data, "additionalFields")
		return data, nil
	}
}

// AddEventUnderscoreIdField add _id to event, because the spring projects read id from _id
func AddEventUnderscoreIdField() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		var event, ok = data["event"].(map[string]any)
		if ok {
			event["_id"] = event["id"]
			data["event"] = event
		}

		return data, nil
	}
}

// MoveTimestamp is currently not used (on purpose)
func MoveTimestamp() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		var event, ok = data["event"].(map[string]any)
		if !ok {
			return nil, errors.New(fmt.Sprintf("could no move timestamp from event.time to timestamp"))
		}

		if _, ok := event["time"]; !ok {
			return data, nil
		}

		timeString, ok := event["time"].(string)
		if !ok {
			return nil, errors.New("could not cast event.time to string!")
		}

		var parsed, err = time.Parse(time.RFC3339, timeString)
		if err != nil {
			return nil, errors.New("could not parse event.time as RFC3339 timestamp. Expected format: " + time.RFC3339)
		}

		data["timestamp"] = parsed
		return data, nil
	}
}

func UpdateModifiedTime() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		data["modified"] = time.Now().UTC()
		return data, nil
	}
}

func AddTimestampIfDropped() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		if status, ok := data["status"]; ok && status == "DROPPED" {
			data["timestamp"] = time.Now().UTC()
		}
		return data, nil
	}
}

func EnrichPropertiesFromHttpHeaders() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		headersToInclude := []string{"x-business-context", "x-correlation-id"}
		log.Debug().Fields(data).Msg("Read data fields")
		properties, isPropertiesOk := data["properties"].(map[string]any)
		httpHeaders, isHttpHeaderOk := data["httpHeaders"].(map[string]interface{})

		log.Debug().Fields(properties).Bool("isPropertiesOk", isPropertiesOk).Msg("Read properties field")
		log.Debug().Fields(httpHeaders).Bool("isHttpHeaderOk", isHttpHeaderOk).Msg("Read httpHeaders field")

		if !isHttpHeaderOk || !isPropertiesOk {
			return data, nil
		}

		for _, headerToInclude := range headersToInclude {
			headerArray, ok := httpHeaders[headerToInclude].([]interface{})
			if !ok {
				continue
			}

			log.Debug().Any("headerArray", headerArray).Msgf("Read headerArray")

			if len(headerArray) == 0 {
				continue
			}

			// Normally length is always = 1, but to ensure possible added fields joining all values
			headerArrayStrings := stringifySlice(headerArray)
			properties[headerToInclude] = strings.Join(headerArrayStrings, ",")
		}

		data["properties"] = properties

		return data, nil
	}
}

func DropHttpHeaders() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		delete(data, "httpHeaders")
		return data, nil
	}
}

func DropEventData() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		var event = data["event"].(map[string]any)
		delete(event, "data")
		data["event"] = event
		return data, nil
	}
}

func Flatten() TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		return utils.Flatten(data, ""), nil
	}
}

func DeleteFlatKeys(keys ...string) TransformFunc {
	return func(data map[string]any) (map[string]any, error) {
		for _, key := range keys {
			delete(data, key)
		}
		return data, nil
	}
}

func stringifySlice(slice []any) []string {
	stringSlice := make([]string, len(slice))
	for i, v := range slice {
		stringSlice[i] = v.(string)
	}
	return stringSlice
}
