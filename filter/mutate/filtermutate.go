package filtermutate

import (
	"context"
	"errors"
	"strings"

	"github.com/tsaikd/gogstash/config"
	"github.com/tsaikd/gogstash/config/logevent"
)

// ModuleName is the name used in config file
const ModuleName = "mutate"

// errors
var (
	ErrNotConfigured = errors.New("filter mutate not configured")
)

// FilterConfig holds the configuration json fields and internal objects
type FilterConfig struct {
	config.FilterConfig

	Split   [2]string `yaml:"split"`
	Replace [3]string `yaml:"replace"`
	Merge   [2]string `yaml:"merge"` // merge string value into existing string slice field
}

// DefaultFilterConfig returns an FilterConfig struct with default values
func DefaultFilterConfig() FilterConfig {
	return FilterConfig{
		FilterConfig: config.FilterConfig{
			CommonConfig: config.CommonConfig{
				Type: ModuleName,
			},
		},
	}
}

// InitHandler initialize the filter plugin
func InitHandler(ctx context.Context, raw *config.ConfigRaw) (config.TypeFilterConfig, error) {
	conf := DefaultFilterConfig()
	err := config.ReflectConfig(raw, &conf)
	if err != nil {
		return nil, err
	}

	if conf.Split[0] == "" && conf.Replace[0] == "" && conf.Merge[0] == "" && len(conf.AddTags) == 0 && len(conf.RemoveTags) == 0 && len(conf.RemoveFields) == 0 && len(conf.AddFields) == 0 {
		return nil, ErrNotConfigured
	}

	return &conf, nil
}

// Event the main filter event
func (f *FilterConfig) Event(ctx context.Context, event logevent.LogEvent) logevent.LogEvent {
	if f.Split[0] != "" {
		event.SetValue(f.Split[0], strings.Split(event.GetString(f.Split[0]), f.Split[1]))
	}
	if f.Replace[0] != "" {
		event.SetValue(f.Replace[0], strings.Replace(event.GetString(f.Replace[0]), f.Replace[1], f.Replace[2], -1))
	}
	if f.Merge[0] != "" {
		event = mergeField(event, f.Merge[0], f.Merge[1])
	}
	return event
}

func mergeField(event logevent.LogEvent, destinationName, source string) logevent.LogEvent {
	destinationValue := event.Get(destinationName)
	value := event.Format(source)
	if destinationValue == nil {
		destinationValue = []string{value}
		event.SetValue(destinationName, destinationValue)
		return event
	}
	switch currentDestination := destinationValue.(type) {
	case string:
		var newDestination []string
		if currentDestination != "" {
			newDestination = append(newDestination, currentDestination)
		}
		newDestination = append(newDestination, value)
		event.SetValue(destinationName, newDestination)
	case []string:
		currentDestination = append(currentDestination, value)
		event.SetValue(destinationName, currentDestination)
	}
	return event
}
