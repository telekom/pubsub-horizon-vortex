// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strings"
)

func Flatten(data map[string]any, prefix string) map[string]any {
	var flatMap = make(map[string]any)

	for key, value := range data {
		var fullKey string

		var keyBuilder = strings.Builder{}
		if len(prefix) > 0 {
			keyBuilder.WriteString(prefix)
			keyBuilder.WriteRune('.')
		}
		keyBuilder.WriteString(key)
		fullKey = keyBuilder.String()

		if subMap, ok := value.(map[string]any); ok {
			var subFlatMap = Flatten(subMap, fullKey)
			for subKey, subValue := range subFlatMap {
				flatMap[subKey] = subValue
			}
		} else {
			flatMap[fullKey] = value
		}
	}

	return flatMap
}
