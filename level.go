package betterslog

import (
	"fmt"
	"log/slog"
	"strings"
)

// Name for common levels.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type (
	Level    = slog.Level
	LevelVar = slog.LevelVar
	Leveler  = slog.Leveler
)

// NewPerLoggerLeveler configures logging level by logger names.
// The format is "loggerName.subLogger=level".
// If a level is configured for a parent logger, but not configured for
// a child logger, the child logger derives from its parent.
//
// Note that per-logger levels only work with handlers in this package.
func NewPerLoggerLeveler(defaultLevel Leveler, perLoggerLevels []string) (Leveler, error) {
	levelFunc, err := buildPerLoggerLevelFunc(perLoggerLevels)
	if err != nil {
		return nil, err
	}
	pll := &perLoggerLeveler{
		defaultLevel: defaultLevel,
		levelFunc:    levelFunc,
	}
	return pll, nil
}

type perLoggerLeveler struct {
	defaultLevel Leveler
	levelFunc    perLoggerLevelFunc
}

func (pll *perLoggerLeveler) Level() Level {
	return pll.defaultLevel.Level()
}

func (pll *perLoggerLeveler) GetLoggerLevel(loggerName string) Level {
	if loggerName != "" && pll.levelFunc != nil {
		if level, ok := pll.levelFunc(loggerName); ok {
			return level
		}
	}
	return pll.defaultLevel.Level()
}

type perLoggerLevelFunc func(name string) (Level, bool)

func buildPerLoggerLevelFunc(levelRules []string) (perLoggerLevelFunc, error) {
	if len(levelRules) == 0 {
		return nil, nil
	}
	tree := &radixTree[Level]{}
	for _, rule := range levelRules {
		tmp := strings.Split(rule, "=")
		if len(tmp) != 2 {
			return nil, fmt.Errorf("invalid per logger level rule: %s", rule)
		}
		loggerName, levelName := tmp[0], tmp[1]
		var level Level
		if err := level.UnmarshalText([]byte(levelName)); err != nil {
			return nil, fmt.Errorf("invalid level: %q: %w", levelName, err)
		}
		tree.root.insert(loggerName, level)
	}
	return tree.search, nil
}
