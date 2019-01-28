package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	scerrors "github.com/aphistic/softcopy/internal/pkg/errors"
)

type OptionLoader struct {
	envLoader EnvLoader

	options Options
}

func NewOptionLoader(options Options) *OptionLoader {
	return &OptionLoader{
		envLoader: &realEnvLoader{},

		options: options,
	}
}

func (ol *OptionLoader) getValue(name string) (interface{}, error) {
	opt, ok := ol.options.Option(name)
	if !ok {
		return "", scerrors.ErrNotFound
	}

	if opt.Value != "" && opt.ValueFrom != nil {
		return nil, fmt.Errorf("options cannot have a value and a value from")
	}

	if opt.ValueFrom != nil {
		if opt.ValueFrom.EnvRef != nil {
			val, ok := os.LookupEnv(opt.ValueFrom.EnvRef.Key)
			if !ok {
				return nil, scerrors.ErrNotFound
			}

			return val, nil
		}
		return nil, fmt.Errorf("could not load external value")
	}

	return opt.Value, nil
}

func (ol *OptionLoader) GetString(name string) (string, error) {
	val, err := ol.getValue(name)
	if err != nil {
		return "", err
	}

	strVal, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("unhandled type: %s", reflect.TypeOf(val))
	}

	return strVal, nil
}

func (ol *OptionLoader) GetStringOrDefault(name string, defaultVal string) (string, error) {
	val, err := ol.GetString(name)
	if err == scerrors.ErrNotFound {
		return defaultVal, nil
	} else if err != nil {
		return "", err
	}

	return val, nil
}

func (ol *OptionLoader) GetInt(name string) (int, error) {
	val, err := ol.getValue(name)
	if err != nil {
		return 0, err
	}

	switch v := val.(type) {
	case string:
		val, err := strconv.ParseInt(v, 0, 0)
		if err != nil {
			return 0, err
		}
		return int(val), nil
	default:
		return 0, fmt.Errorf("unhandled type: %s", reflect.TypeOf(val))
	}
}

func (ol *OptionLoader) GetIntOrDefault(name string, defaultVal int) (int, error) {
	val, err := ol.GetInt(name)
	if err == scerrors.ErrNotFound {
		return defaultVal, nil
	} else if err != nil {
		return 0, err
	}

	return val, err
}

type Options []*Option

func (o Options) Option(name string) (*Option, bool) {
	for _, opt := range o {
		if opt.Name == name {
			return opt, true
		}
	}

	return nil, false
}

type Option struct {
	Name      string           `yaml:"name"`
	Value     string           `yaml:"value"`
	ValueFrom *OptionValueFrom `yaml:"valueFrom"`
}

type OptionValueFrom struct {
	EnvRef *OptionValueEnvRef `yaml:"envRef"`
}

type OptionValueEnvRef struct {
	Key string `yaml:"key"`
}
