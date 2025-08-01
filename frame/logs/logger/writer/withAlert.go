package writer

import (
	"github.com/aiden2048/pkg/frame/logs/logger"
)

func NewWithAlertWriter(realWriter logger.Writer, cfgLoader *logger.ConfLoader, alertFunc AlertFunc) *WithAlertWriter {
	return &WithAlertWriter{
		cfgLoader:  cfgLoader,
		alertFunc:  alertFunc,
		realWriter: realWriter,
	}
}

var _ logger.Writer = (*WithAlertWriter)(nil)

type AlertFunc func(msg *logger.Msg)

type WithAlertWriter struct {
	realWriter logger.Writer
	cfgLoader  *logger.ConfLoader
	alertFunc  AlertFunc
}

func (w *WithAlertWriter) GetAlertLevel() logger.Level {
	levelStr := w.cfgLoader.GetConf().AlertLevel
	if levelStr == "" {
		return logger.LevelWarn
	}
	return logger.TransStrToLevel(levelStr)
}

func (w *WithAlertWriter) WriteMsg(msg *logger.Msg) error {
	if err := w.realWriter.WriteMsg(msg); err != nil {
		return err
	}

	if w.alertFunc == nil {
		return nil
	}

	if w.GetAlertLevel() > msg.Level {
		return nil
	}

	w.alertFunc(msg)

	return nil
}

func (w *WithAlertWriter) GetMsg(level logger.Level, format string, args ...interface{}) (*logger.Msg, error) {
	return w.realWriter.GetMsg(level, format, args...)
}

func (w *WithAlertWriter) GetMsgBySkipCall(level logger.Level, skipCall int, format string, args ...interface{}) (*logger.Msg, error) {
	return w.realWriter.GetMsgBySkipCall(level, skipCall, format, args...)
}

func (w *WithAlertWriter) GetSkipCall() int {
	return w.realWriter.GetSkipCall() - 1
}

func (w *WithAlertWriter) Write(level logger.Level, format string, args ...interface{}) error {
	if w.alertFunc == nil {
		if err := w.realWriter.Write(level, format, args...); err != nil {
			return err
		}
		return nil
	}

	if w.GetAlertLevel() > level {
		if err := w.realWriter.Write(level, format, args...); err != nil {
			return err
		}
		return nil
	}

	msg, err := w.realWriter.GetMsg(level, format, args...)
	if err != nil {
		return err
	}

	if err = w.realWriter.WriteMsg(msg); err != nil {
		return err
	}

	w.alertFunc(msg)

	return nil
}

func (w *WithAlertWriter) WriteBySkipCall(level logger.Level, skipCall int, format string, args ...interface{}) error {
	if w.alertFunc == nil {
		if err := w.realWriter.WriteBySkipCall(level, skipCall, format, args...); err != nil {
			return err
		}
		return nil
	}

	if w.GetAlertLevel() > level {
		if err := w.realWriter.WriteBySkipCall(level, skipCall, format, args...); err != nil {
			return err
		}
		return nil
	}

	msg, err := w.realWriter.GetMsgBySkipCall(level, skipCall, format, args...)
	if err != nil {
		return err
	}

	if err = w.realWriter.WriteMsg(msg); err != nil {
		return err
	}

	w.alertFunc(msg)

	return nil
}

func (w *WithAlertWriter) Flush() error {
	return w.realWriter.Flush()
}
