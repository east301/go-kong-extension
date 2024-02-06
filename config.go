package kongext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

type ConfigFunc func(string) mo.Option[any]

func CreateYAMLConfigFromFile(path string) (ConfigFunc, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "could not read a file")
	}

	var parsed map[string]any
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		return nil, errors.Wrap(err, "could not parse YAML")
	}

	return CreateMapConfig(parsed), nil
}

func CreateMapConfig(target map[string]any) ConfigFunc {
	return func(path string) mo.Option[any] {
		var current any = target

		for _, key := range strings.Split(path, ".") {
			if current == nil {
				return mo.None[any]()
			} else if reflect.TypeOf(current).Kind() != reflect.Map {
				return mo.None[any]()
			}

			child := reflect.ValueOf(current).MapIndex(reflect.ValueOf(key))
			if !child.IsValid() {
				return mo.None[any]()
			}

			current = child.Interface()
		}

		return mo.Some(current)
	}
}

func CreateStructConfig(target any) ConfigFunc {
	return func(path string) mo.Option[any] {
		var current any = target

		for _, key := range strings.Split(path, ".") {
			//
			if current == nil {
				return mo.None[any]()
			}

			currentType := reflect.TypeOf(current)
			if currentType.Kind() != reflect.Struct {
				return mo.None[any]()
			}

			var matchedField *reflect.Value
			currentValue := reflect.ValueOf(current)
			for i := 0; i < currentType.NumField(); i++ {
				fieldValue := currentValue.Field(i)
				fieldType := currentType.Field(i)

				if fieldValue.IsValid() && GetStructFieldName(fieldType) == key {
					matchedField = &fieldValue
					break
				}
			}

			if matchedField == nil {
				return mo.None[any]()
			}

			current = matchedField.Interface()
		}

		return mo.Some(current)
	}
}

func GetStructFieldName(field reflect.StructField) string {
	if v := field.Tag.Get("json"); v != "" {
		return strings.Split(v, ",")[0]
	} else if v := field.Tag.Get("name"); v != "" {
		return v
	}

	return field.Name
}

func ConfigResolver(configs ...ConfigFunc) kong.Option {
	return kong.Resolvers(kong.ResolverFunc(func(_ *kong.Context, _ *kong.Path, flag *kong.Flag) (any, error) {
		for _, config := range configs {
			//
			v, ok := config(flag.Name).Get()
			if !ok {
				continue
			}

			//
			var result any = v
			if t := reflect.TypeOf(result); t.Kind() == reflect.Slice || t.Kind() == reflect.Map {
				if serialized, err := json.Marshal(result); err == nil {
					result = string(serialized)
				} else {
					return nil, errors.Wrap(err, "could not serialize value")
				}
			}

			return result, nil
		}

		return nil, nil
	}))
}

func CollectConfigFiles(candidatePaths ...string) []string {
	var result []string
	for _, path := range candidatePaths {
		if _, err := os.Stat(path); err == nil {
			result = append(result, path)
		}
	}

	return result
}

func LoadConfigs(appName string, defaultConfigObject any) ([]ConfigFunc, error) {
	//
	var result []ConfigFunc

	//
	configPaths := CollectConfigFiles(
		filepath.Join(lo.Must(os.Getwd()), "%s.config.yaml"),
		"/opt/%s/etc/%s.config.yaml",
		"/etc/%s.config.yaml",
	)

	for _, path := range configPaths {
		config, err := CreateYAMLConfigFromFile(fmt.Sprintf(path, appName))
		if err != nil {
			return nil, err
		}

		result = append(result, config)
	}

	//
	if defaultConfigObject != nil {
		result = append(result, CreateStructConfig(defaultConfigObject))
	}

	//
	return result, nil
}
