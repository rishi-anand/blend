package blend

import (
	"reflect"

	"gopkg.in/yaml.v2"
)

func Blend(value, override []byte) ([]byte, error) {
	valueSlice := yaml.MapSlice{}
	if err := yaml.Unmarshal(value, &valueSlice); err != nil {
		return nil, err
	}

	overrideSlice := yaml.MapSlice{}
	if err := yaml.Unmarshal(override, &overrideSlice); err != nil {
		return nil, err
	}

	for _, overrideItem := range overrideSlice {
		for k, valueItem := range valueSlice {
			if overrideItem.Key == valueItem.Key {
				valueSlice[k] = getValue(valueItem, overrideItem)
			}
		}
	}

	if valueOverride, err := yaml.Marshal(valueSlice); err != nil {
		return nil, err
	} else {
		return valueOverride, nil
	}
}

func getValue(value, override yaml.MapItem) yaml.MapItem {
	if reflect.TypeOf(override.Value) == reflect.TypeOf(yaml.MapSlice{}) &&
		reflect.TypeOf(value.Value) == reflect.TypeOf(yaml.MapSlice{}) {
		override.Value = getMapSliceValue(value.Value.(yaml.MapSlice), override.Value.(yaml.MapSlice))
	}
	return override
}

func getMapSliceValue(valueSlice, overrideSlice yaml.MapSlice) yaml.MapSlice {
	for _, overrideItem := range overrideSlice {
		for k, valueItem := range valueSlice {
			if overrideItem.Key == valueItem.Key {
				valueSlice[k] = getValue(valueItem, overrideItem)
			}
		}
	}
	return valueSlice
}
