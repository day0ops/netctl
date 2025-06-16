package log

import (
	"fmt"
	"io"
	"os"
)

type FmtMachineLogger struct {
	outWriter io.Writer
	errWriter io.Writer
	debug     bool
}

// NewFmtMachineLogger creates a MachineLogger implementation used by the drivers
func NewFmtMachineLogger() *FmtMachineLogger {
	return &FmtMachineLogger{
		outWriter: os.Stdout,
		errWriter: os.Stderr,
		debug:     false,
	}
}

func (ml *FmtMachineLogger) SetDebug(debug bool) {
	ml.debug = debug
}

func (ml *FmtMachineLogger) SetOutWriter(out io.Writer) {
	ml.outWriter = out
}

func (ml *FmtMachineLogger) SetErrWriter(err io.Writer) {
	ml.errWriter = err
}

func (ml *FmtMachineLogger) Debug(args ...interface{}) {
	if ml.debug {
		fmt.Fprintln(ml.errWriter, args...)
	}
}

func (ml *FmtMachineLogger) Debugf(fmtString string, args ...interface{}) {
	if ml.debug {
		fmt.Fprintf(ml.errWriter, fmtString+"\n", args...)
	}
}

func (ml *FmtMachineLogger) Error(args ...interface{}) {
	fmt.Fprintln(ml.errWriter, args...)
}

func (ml *FmtMachineLogger) Errorf(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.errWriter, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Info(args ...interface{}) {
	fmt.Fprintln(ml.outWriter, args...)
}

func (ml *FmtMachineLogger) Infof(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.outWriter, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Warn(args ...interface{}) {
	fmt.Fprintln(ml.outWriter, args...)
}

func (ml *FmtMachineLogger) Warnf(fmtString string, args ...interface{}) {
	fmt.Fprintf(ml.outWriter, fmtString+"\n", args...)
}
