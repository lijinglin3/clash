package log

import (
	"encoding/json"
	"errors"
)

// LevelMapping is a mapping for Level enum
var LevelMapping = map[string]Level{
	ERROR.String():   ERROR,
	WARNING.String(): WARNING,
	INFO.String():    INFO,
	DEBUG.String():   DEBUG,
	SILENT.String():  SILENT,
}

const (
	DEBUG Level = iota
	INFO
	WARNING
	ERROR
	SILENT
)

type Level int

// UnmarshalYAML unserialize Level with yaml
func (l *Level) UnmarshalYAML(unmarshal func(any) error) error {
	var tp string
	unmarshal(&tp)
	level, exist := LevelMapping[tp]
	if !exist {
		return errors.New("invalid mode")
	}
	*l = level
	return nil
}

// UnmarshalJSON unserialize Level with json
func (l *Level) UnmarshalJSON(data []byte) error {
	var tp string
	json.Unmarshal(data, &tp)
	level, exist := LevelMapping[tp]
	if !exist {
		return errors.New("invalid mode")
	}
	*l = level
	return nil
}

// MarshalJSON serialize Level with json
func (l Level) MarshalJSON() ([]byte, error) {
	return json.Marshal(l.String())
}

// MarshalYAML serialize Level with yaml
func (l Level) MarshalYAML() (any, error) {
	return l.String(), nil
}

func (l Level) String() string {
	switch l {
	case INFO:
		return "info"
	case WARNING:
		return "warning"
	case ERROR:
		return "error"
	case DEBUG:
		return "debug"
	case SILENT:
		return "silent"
	default:
		return "unknown"
	}
}
