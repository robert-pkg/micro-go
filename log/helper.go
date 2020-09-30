package log

import "os"

type Helper struct {
	Logger
}

func NewHelper(log Logger) *Helper {
	return &Helper{Logger: log}
}

func (h *Helper) Info(msg string, args ...interface{}) {
	h.Log(InfoLevel, msg, args...)
}

func (h *Helper) Infof(template string, args ...interface{}) {
	h.Logf(InfoLevel, template, args...)
}

func (h *Helper) Trace(msg string, args ...interface{}) {
	h.Log(TraceLevel, msg, args...)
}

func (h *Helper) Tracef(template string, args ...interface{}) {
	h.Logf(TraceLevel, template, args...)
}

func (h *Helper) Debug(msg string, args ...interface{}) {
	h.Log(DebugLevel, msg, args...)
}

func (h *Helper) Debugf(template string, args ...interface{}) {
	h.Logf(DebugLevel, template, args...)
}

func (h *Helper) Warn(msg string, args ...interface{}) {
	h.Log(WarnLevel, msg, args...)
}

func (h *Helper) Warnf(template string, args ...interface{}) {
	h.Logf(WarnLevel, template, args...)
}

func (h *Helper) Error(msg string, args ...interface{}) {
	h.Log(ErrorLevel, msg, args...)
}

func (h *Helper) Errorf(template string, args ...interface{}) {
	h.Logf(ErrorLevel, template, args...)
}

func (h *Helper) Fatal(msg string, args ...interface{}) {
	h.Log(FatalLevel, msg, args...)
	os.Exit(1)
}

func (h *Helper) Fatalf(template string, args ...interface{}) {
	h.Logf(FatalLevel, template, args...)
	os.Exit(1)
}
