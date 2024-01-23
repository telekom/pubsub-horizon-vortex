// Copyright 2024 Deutsche Telekom IT GmbH
//
// SPDX-License-Identifier: Apache-2.0

package transforms

var GlobalRegistry *Registry

type Registry struct {
	transforms []TransformFunc
}

type TransformFunc func(data map[string]any) (map[string]any, error)

func NewRegistry() *Registry {
	return &Registry{
		make([]TransformFunc, 0),
	}
}

func init() {
	GlobalRegistry = NewRegistry()
	GlobalRegistry.Register(
		RenameAdditionalFields(),
		EnrichPropertiesFromHttpHeaders(),
		DropHttpHeaders(),
		DropEventData(),
		AddTimestampIfDropped(),
		UpdateModifiedTime(),
		AddEventUnderscoreIdField(),
		Flatten(),
		DeleteFlatKeys(
			"event.source",
			"event.specversion",
			"event.datacontenttype",
			"event.dataref",
			"uuid",
		),
	)
}

func (r *Registry) Register(transformFuncs ...TransformFunc) {
	for _, transformFunc := range transformFuncs {
		r.transforms = append(r.transforms, transformFunc)
	}
}

func (r *Registry) ApplyTransforms(data map[string]any) (map[string]any, error) {
	var current = data
	var err error
	for _, transform := range r.transforms {
		current, err = transform(current)
		if err != nil {
			return nil, err
		}
	}
	return current, nil
}
