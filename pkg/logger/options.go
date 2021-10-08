package logger

import (
	"fmt"
)

const (
	defaultJSONOutput   = false
	defaultReportCaller = false
	defaultOutputLevel  = "debug"
	undefinedID         = ""
)

// Options defines the sets of options for Keel logging
type Options struct {
	// ID is the unique id of Keel Application
	ID string

	// JSONFormatEnabled is the flag to enable JSON formatted log
	JSONFormatEnabled bool

	// OutputLevel is the level of logging
	OutputLevel string
}

// SetOutputLevel sets the log output level
func (o *Options) SetOutputLevel(outputLevel string) error {
	if toLogLevel(outputLevel) == UndefinedLevel {
		return fmt.Errorf("undefined Log Output Level: %s", outputLevel)
	}
	o.OutputLevel = outputLevel
	return nil
}

// SetID sets Keel Application ID
func (o *Options) SetID(id string) {
	o.ID = id
}

// AttachCmdFlags attaches log options to command flags
func (o *Options) AttachCmdFlags(
	stringVar func(p *string, name string, value string, usage string),
	boolVar func(p *bool, name string, value bool, usage string)) {
	if stringVar != nil {
		stringVar(
			&o.OutputLevel,
			"log-level",
			defaultOutputLevel,
			"Options are debug, info, warn, error, or fatal (default info)")
	}
	if boolVar != nil {
		boolVar(
			&o.JSONFormatEnabled,
			"log-as-json",
			defaultJSONOutput,
			"print log as JSON (default false)")
	}
}

// DefaultOptions returns default values of Options
func DefaultOptions() Options {
	return Options{
		JSONFormatEnabled: defaultJSONOutput,
		ID:                undefinedID,
		OutputLevel:       defaultOutputLevel,
	}
}

// ApplyOptionsToLoggers applys options to all registered loggers
func ApplyOptionsToLoggers(options *Options) error {
	internalLoggers := getLoggers()

	// Apply formatting options first
	for _, v := range internalLoggers {
		v.EnableJSONOutput(options.JSONFormatEnabled)

		if options.ID != undefinedID {
			v.SetID(options.ID)
		}
	}

	pluginLogLevel := toLogLevel(options.OutputLevel)
	if pluginLogLevel == UndefinedLevel {
		return fmt.Errorf("invalid value for --log-level: %s", options.OutputLevel)
	}

	for _, v := range internalLoggers {
		v.SetOutputLevel(pluginLogLevel)
	}
	return nil
}
