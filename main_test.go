package kongext

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

type AppConfig struct {
	Group1 Group1Config     `prefix:"group1." embed:"" json:"group1"`
	Group2 Group2ConfigList `name:"group2"             json:"group2"`
}

type Group1Config struct {
	Key string `name:"key" json:"key"`
}

type Group2Config struct {
	Entry string `name:"entry" json:"entry"`
}

type Group2ConfigList []Group2Config

func (c Group2ConfigList) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Group2Config(c))
}

func (c *Group2ConfigList) UnmarshalJSON(data []byte) error {
	var result []Group2Config
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	*c = result
	return nil
}

func Test_main_YAML_config(t *testing.T) {
	//
	path := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(
		path,
		[]byte(strings.Join([]string{
			"group1:",
			"  key: value",
			"group2:",
			"  - entry: value1",
			"  - entry: value2",
		}, "\n")),
		0640,
	)

	yamlConfigResolver, err := CreateYAMLConfigFromFile(path)
	if err != nil {
		t.Errorf("could not create YAML config resolver: %+v", err)
	}

	//
	options := parse(
		t,
		[]kong.Option{ConfigResolver(yamlConfigResolver)},
		[]string{},
	)

	if options.Group1.Key != "value" {
		t.Error()
	}
	if len(options.Group2) != 2 {
		t.Error()
	}
	if options.Group2[0].Entry != "value1" {
		t.Error()
	}
	if options.Group2[1].Entry != "value2" {
		t.Error()
	}
}

func Test_main_JSON_in_command_line(t *testing.T) {
	//
	options := parse(t, nil, []string{"--group2", `[{ "entry": "value" }]`})

	if len(options.Group2) != 1 {
		t.Error()
	}
	if options.Group2[0].Entry != "value" {
		t.Error()
	}
}

func Test_main_struct_default(t *testing.T) {
	//
	defaults := AppConfig{
		Group1: Group1Config{
			Key: "expected",
		},
	}

	//
	options := parse(
		t,
		[]kong.Option{ConfigResolver(CreateStructConfig(defaults))},
		[]string{},
	)

	if options.Group1.Key != "expected" {
		t.Error()
	}
	if len(options.Group2) != 0 {
		t.Error()
	}
}

func Test_main_mix(t *testing.T) {
	//
	defaults := AppConfig{
		Group1: Group1Config{
			Key: "default",
		},
	}

	path := filepath.Join(t.TempDir(), "config.yaml")
	os.WriteFile(
		path,
		[]byte(strings.Join([]string{
			"group1:",
			"  key: config",
		}, "\n")),
		0640,
	)

	yamlConfigResolver, err := CreateYAMLConfigFromFile(path)
	if err != nil {
		t.Error()
	}

	//
	options := parse(
		t,
		[]kong.Option{ConfigResolver(yamlConfigResolver, CreateStructConfig(defaults))},
		[]string{"--group1.key=commandline"},
	)
	if options.Group1.Key != "commandline" {
		t.Error()
	}

	options = parse(
		t,
		[]kong.Option{ConfigResolver(yamlConfigResolver, CreateStructConfig(defaults))},
		[]string{},
	)
	if options.Group1.Key != "config" {
		t.Error()
	}

	options = parse(
		t,
		[]kong.Option{ConfigResolver(CreateStructConfig(defaults))},
		[]string{},
	)
	if options.Group1.Key != "default" {
		t.Error()
	}

	options = parse(
		t,
		[]kong.Option{},
		[]string{},
	)
	if options.Group1.Key != "" {
		t.Error()
	}
}

func parse(t *testing.T, kongOptions []kong.Option, args []string) AppConfig {
	//
	t.Helper()

	//
	var options AppConfig

	parser, err := kong.New(&options, kongOptions...)
	if err != nil {
		t.Errorf("could not construct a parser: %+v", err)
	}
	if _, err := parser.Parse(args); err != nil {
		t.Errorf("could not parse command line options: %+v", err)
	}

	return options
}
