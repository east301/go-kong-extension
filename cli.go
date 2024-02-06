package kongext

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/goccy/go-yaml"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func Parse[T any](configs []ConfigFunc, options ...kong.Option) (*kong.Kong, *kong.Context, T) {
	//
	var wrapped struct {
		Result  T        `embed:""`
		Configs []string `name:"config"`
	}

	parser := lo.Must(kong.New(&wrapped, append([]kong.Option{ConfigResolver(configs...)}, options...)...))
	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	if len(wrapped.Configs) == 0 {
		return parser, ctx, wrapped.Result
	}

	//
	for _, path := range lo.Reverse(wrapped.Configs) {
		config, err := CreateYAMLConfigFromFile(path)
		if err != nil {
			parser.FatalIfErrorf(err)
		}

		configs = append([]ConfigFunc{config}, configs...)
	}

	parser = lo.Must(kong.New(&wrapped, append([]kong.Option{ConfigResolver(configs...)}, options...)...))
	ctx, err = parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	return parser, ctx, wrapped.Result
}

type ConfigDumpCommand[T any] struct{}

func (c ConfigDumpCommand[T]) Run(config T) error {
	content, err := yaml.MarshalWithOptions(config, yaml.Indent(2))
	if err != nil {
		return errors.Wrap(err, "could not marshal config as YAML")
	} else {
		fmt.Println(string(content))
	}

	return nil
}
