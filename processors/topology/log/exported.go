package log

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

// WithError creates an entry from the standard logger and adds an error to it, using the value defined in "error" as key.
func WithError(err error) *logrus.Entry {
	return WithFields(fileAndLine()).WithField("error", err)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	WithFields(fileAndLine()).Debug(args...)
}

func DebugWithContext(args ...interface{}) {
	WithFields(contextFields()).Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	WithFields(fileAndLine()).Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	WithFields(fileAndLine()).Info(args...)
}

// InfoWithoutContext logs a message at level Info on the standard logger
// Without the fileAndLine
// Info logs a message at level Info on the standard logger.
func InfoWithoutContextOrFormatting(args ...interface{}) {
	if Logr.Level != logrus.ErrorLevel {
		fmt.Fprintln(Logr.Out, args...)
	}
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	WithFields(fileAndLine()).Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	WithFields(fileAndLine()).Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	WithFields(fileAndLine()).Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	WithFields(fileAndLine()).Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	WithFields(fileAndLine()).Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Debugf(format, args...)
}

func DebugfWithContext(format string, args ...interface{}) {
	WithFields(contextFields()).Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(format string, args ...interface{}) {
	WithFields(fileAndLine()).Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	WithFields(fileAndLine()).Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(args ...interface{}) {
	WithFields(fileAndLine()).Debugln(args...)
}

func DebuglnWithContext(args ...interface{}) {
	WithFields(contextFields()).Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(args ...interface{}) {
	WithFields(fileAndLine()).Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(args ...interface{}) {
	WithFields(fileAndLine()).Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(args ...interface{}) {
	WithFields(fileAndLine()).Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(args ...interface{}) {
	WithFields(fileAndLine()).Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(args ...interface{}) {
	WithFields(fileAndLine()).Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(args ...interface{}) {
	WithFields(fileAndLine()).Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(args ...interface{}) {
	WithFields(fileAndLine()).Fatalln(args...)
}
