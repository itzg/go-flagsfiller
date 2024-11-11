package flagsfiller

import (
	"log/slog"
	"reflect"
)

func init() {
	RegisterSimpleType(slogLevelConverter)
}

func slogLevelConverter(s string, _ reflect.StructTag) (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(s))
	if err != nil {
		return slog.LevelInfo, err
	}
	return level, nil
}
