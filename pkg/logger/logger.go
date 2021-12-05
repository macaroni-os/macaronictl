/*
	Copyright © 2021 RockHopper OS Linux
	See AUTHORS and LICENSE for the license details and contributors.
*/
package logger

import (
	"fmt"
	"os"
	"regexp"

	specs "github.com/funtoo/rhctl/pkg/specs"

	"github.com/kyokomi/emoji"
	"github.com/logrusorgru/aurora"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type RhCtlLogger struct {
	Config *specs.RhCtlConfig
	Logger *zap.Logger
	Aurora aurora.Aurora
}

var defaultLogger *RhCtlLogger = nil

func NewRhCtlLogger(config *specs.RhCtlConfig) *RhCtlLogger {
	return &RhCtlLogger{
		Logger: nil,
		Aurora: aurora.NewAurora(config.GetLogging().Color),
		Config: config,
	}
}

func (l *RhCtlLogger) GetAurora() aurora.Aurora {
	return l.Aurora
}

func (l *RhCtlLogger) SetAsDefault() {
	defaultLogger = l
}

func GetDefaultLogger() *RhCtlLogger {
	return defaultLogger
}

func (l *RhCtlLogger) InitLogger2File() error {
	var err error

	// TODO: test permission for open logfile.
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{l.Config.GetLogging().Path}
	cfg.Level = level2AtomicLevel(l.Config.GetLogging().Level)
	cfg.ErrorOutputPaths = []string{}
	if l.Config.GetLogging().JsonFormat {
		cfg.Encoding = "json"
	} else {
		cfg.Encoding = "console"
	}
	cfg.DisableCaller = true
	cfg.DisableStacktrace = true
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	l.Logger, err = cfg.Build()
	if err != nil {
		fmt.Fprint(os.Stderr, "Error on initialize file logger: "+err.Error()+"\n")
		return err
	}

	return nil
}

func level2Number(level string) int {
	switch level {
	case "error":
		return 0
	case "warning":
		return 1
	case "info":
		return 2
	default:
		return 3
	}
}

func (l *RhCtlLogger) log2File(level, msg string) {
	switch level {
	case "error":
		l.Logger.Error(msg)
	case "warning":
		l.Logger.Warn(msg)
	case "info":
		l.Logger.Info(msg)
	default:
		l.Logger.Debug(msg)
	}
}

func level2AtomicLevel(level string) zap.AtomicLevel {
	switch level {
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "warning":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	default:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	}
}

func (l *RhCtlLogger) Msg(level string, withoutColor, ln bool, msg ...interface{}) {
	var message string
	var confLevel, msgLevel int

	if l.Config.GetGeneral().HasDebug() {
		confLevel = 3
	} else {
		confLevel = level2Number(l.Config.GetLogging().Level)
	}
	msgLevel = level2Number(level)
	if msgLevel > confLevel {
		return
	}

	for idx, m := range msg {
		if idx > 0 {
			message += " "
		}
		message += fmt.Sprintf("%v", m)
	}

	var levelMsg string

	if withoutColor || !l.Config.GetLogging().Color {
		levelMsg = message
	} else {
		switch level {
		case "warning":
			levelMsg = l.Aurora.Bold(l.Aurora.Yellow(":construction: " + message)).BgBlack().String()
		case "debug":
			levelMsg = l.Aurora.White(message).BgBlack().String()
		case "info":
			levelMsg = l.Aurora.Bold(l.Aurora.White(message)).BgBlack().String()
		case "error":
			levelMsg = l.Aurora.Bold(l.Aurora.Red(":bomb: " + message + ":fire:")).BgBlack().String()
		}
	}

	if l.Config.GetLogging().EnableEmoji {
		levelMsg = emoji.Sprint(levelMsg)
	} else {
		re := regexp.MustCompile(`[:][\w]+[:]`)
		levelMsg = re.ReplaceAllString(levelMsg, "")
	}

	if l.Logger != nil {
		l.log2File(level, message)
	}

	if ln {
		fmt.Println(levelMsg)
	} else {
		fmt.Print(levelMsg)
	}
}

func (l *RhCtlLogger) Warning(mess ...interface{}) {
	l.Msg("warning", false, true, mess...)
}

func (l *RhCtlLogger) Debug(mess ...interface{}) {
	l.Msg("debug", false, true, mess...)
}

func (l *RhCtlLogger) DebugC(mess ...interface{}) {
	l.Msg("debug", true, true, mess...)
}

func (l *RhCtlLogger) Info(mess ...interface{}) {
	l.Msg("info", false, true, mess...)
}

func (l *RhCtlLogger) InfoC(mess ...interface{}) {
	l.Msg("info", true, true, mess...)
}

func (l *RhCtlLogger) Error(mess ...interface{}) {
	l.Msg("error", false, true, mess...)
}

func (l *RhCtlLogger) Fatal(mess ...interface{}) {
	l.Error(mess)
	os.Exit(1)
}
