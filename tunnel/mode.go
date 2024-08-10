package tunnel

import (
	"encoding/json"
	"errors"
	"strings"
)

type Mode int

// ModeMapping is a mapping for Mode enum
var ModeMapping = map[string]Mode{
	Global.String(): Global,
	Rule.String():   Rule,
	Direct.String(): Direct,
}

const (
	Global Mode = iota
	Rule
	Direct
)

// UnmarshalJSON unserialize Mode
func (m *Mode) UnmarshalJSON(data []byte) error {
	var tp string
	json.Unmarshal(data, &tp)
	mode, exist := ModeMapping[strings.ToLower(tp)]
	if !exist {
		return errors.New("invalid mode")
	}
	*m = mode
	return nil
}

// UnmarshalYAML unserialize Mode with yaml
func (m *Mode) UnmarshalYAML(unmarshal func(any) error) error {
	var tp string
	unmarshal(&tp)
	mode, exist := ModeMapping[strings.ToLower(tp)]
	if !exist {
		return errors.New("invalid mode")
	}
	*m = mode
	return nil
}

// MarshalJSON serialize Mode
func (m Mode) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

// MarshalYAML serialize Mode with yaml
func (m Mode) MarshalYAML() (any, error) {
	return m.String(), nil
}

func (m Mode) String() string {
	switch m {
	case Global:
		return "global"
	case Rule:
		return "rule"
	case Direct:
		return "direct"
	default:
		return "Unknown"
	}
}
