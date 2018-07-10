// Package log implements a logger. Leverages the logrus package.
package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
)

// Config : Struct containing fields required to start an instance of logr
type Config struct {
	Level    string `desc:"One of the logrus log-levels- (info, debug, error, warn)"`
	Location string `desc:"stdout, stderr, or the path to a file"`
	Type     string `desc:"Log output type (text or json)"`
}

// Logr is the instance of logrus
var (
	Logr *logrus.Logger
)

// sync used for singleton pattern in Newlogr
var once sync.Once

// Fields are logrus.Fields. This lets us add our own fields to the logs and enabled to only import this file, not logrus.
type Fields map[string]interface{}

func init() {
	var err error
	cfg := NewLogrConfig()
	if os.Getenv("SIGMA_LOG_TYPE") != "" {
		cfg.Type = os.Getenv("SIGMA_LOG_TYPE")
	}
	if os.Getenv("SIGMA_LOG_LEVEL") != "" {
		cfg.Level = os.Getenv("SIGMA_LOG_LEVEL")
	}

	Logr, err = initLogr(cfg)
	if err != nil {
		fmt.Printf("ERROR: Could not init default logr")
	}
}

// NewLogrConfig : config factory setting defaults for the logr
func NewLogrConfig() Config {
	return Config{
		Level:    "info",
		Location: "stdout",
		Type:     "text",
	}
}

func Field(k string, v interface{}) Fields {
	f := make(Fields)
	f[k] = v
	return f
}

// Only would be called on top level log.WithField, otherwise, it uses logrus Entry.WithField()
func WithField(k string, v interface{}) *logrus.Entry {
	return Logr.WithField(k, v)
}

func WithFields(f ...Fields) *logrus.Entry {
	if len(f) == 0 {
		return Logr.WithFields(logrus.Fields{})
	}
	e := Logr.WithFields(logrus.Fields(f[0]))
	for i := 1; i < len(f); i++ {
		e = e.WithFields(logrus.Fields(f[i]))
	}
	return e
}

// Newlogr is the initializer for a new logr. Sets the level
func NewLogr(conf Config) error {
	// Set up the logger
	var makeErr error
	makelogr := func() {
		Logr, makeErr = initLogr(conf)
		if makeErr != nil {
			fmt.Println(makeErr)
		}
	}
	once.Do(makelogr)

	if Logr == nil || makeErr != nil {
		return fmt.Errorf("Could not instantiate the logr")
	}
	return nil
}

func initLogr(conf Config) (*logrus.Logger, error) {
	// Set up the logger
	initLogr := logrus.New()

	//LOG LEVEL
	lvl, err := logrus.ParseLevel(conf.Level)
	if err != nil {
		fmt.Printf("Invalid log-level: %v", err)
	}
	initLogr.Level = lvl

	//LOG LOCATION
	switch strings.ToLower(conf.Location) {
	case "stdout":
		initLogr.Out = os.Stdout
	case "stderr":
		initLogr.Out = os.Stderr
	default:
		var f *os.File
		f, err = os.OpenFile(conf.Location, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			return initLogr, fmt.Errorf("Could not open file descriptor for logr")
		}
		initLogr.Out = f
	}

	//LOG TYPE
	switch strings.ToLower(conf.Type) {
	case "json":
		initLogr.Formatter = new(logrus.JSONFormatter)
	default:
		custom := new(logrus.TextFormatter)
		custom.FullTimestamp = true
		initLogr.Formatter = custom
	}

	if initLogr == nil {
		return nil, fmt.Errorf("Could not instantiate the logr")
	}

	return initLogr, nil
}

// ContextFields returns packageName, fileName, funcName and linenumber of the caller.
func fileAndLine(lvl ...int) Fields {
	level := 2
	if len(lvl) == 1 {
		level = lvl[0]
	}
	_, file, line, _ := runtime.Caller(level)
	_, fileName := path.Split(file)

	return Fields{
		"file": fileName,
		"line": line,
	}
}

// ContextFields returns packageName, fileName, funcName and linenumber of the caller.
func contextFields(lvl ...int) Fields {
	level := 2
	if len(lvl) == 1 {
		level = lvl[0]
	}
	pc, file, line, _ := runtime.Caller(level)
	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	packageName := ""
	funcName := parts[pl-1]

	if len(parts) >= 0 && pl-2 < len(parts) {
		if parts[pl-2][0] == '(' {
			funcName = parts[pl-2] + "." + funcName
			packageName = strings.Join(parts[0:pl-2], ".")
		} else {
			packageName = strings.Join(parts[0:pl-1], ".")
		}

		pkgs := strings.Split(packageName, "/sigma/")
		if len(pkgs) > 1 {
			packageName = pkgs[1]
		}
	}

	return Fields{
		"package": packageName,
		"file":    fileName,
		"func":    funcName,
		"line":    line,
	}
}

// Closes the logr file
func Close() {
	if Logr.Out.(*os.File) != os.Stderr && Logr.Out.(*os.File) != os.Stdout {
		err := Logr.Out.(*os.File).Close()
		if err != nil {
			fmt.Printf("Could not close logr file descriptor: %v\n", err)
		}
	}
}
