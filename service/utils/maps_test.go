package utils_test

import (
	"log"
	"testing"
	"vortex/service/utils"
)

func TestFlatten(t *testing.T) {
	var input = map[string]any{
		"hello": "world",
		"foo": map[string]any{
			"bar": map[string]any{
				"fizz": "buzz",
			},
		},
	}

	var expectation = map[string]any{
		"hello":        "world",
		"foo.bar.fizz": "buzz",
	}

	log.Printf("%+v\n", input)
	log.Printf("%+v\n", expectation)
	log.Printf("%+v\n", utils.Flatten(input, ""))
}
