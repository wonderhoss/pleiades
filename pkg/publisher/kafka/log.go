package kafka

import (
	"github.com/op/go-logging"
)

type crudLogger struct {
	debugEnable bool
}
type crudErrorLogger struct{}

func newCrudLogger() *crudLogger {
	if logging.GetLevel("kafka-client") == logging.DEBUG {
		return &crudLogger{debugEnable: true}
	}
	return &crudLogger{}
}

func (c *crudLogger) Printf(s string, p ...interface{}) {
	if c.debugEnable {
		kafkaLogger.Debugf(s, p...)
	}
}

func (c *crudErrorLogger) Printf(s string, p ...interface{}) {
	kafkaLogger.Errorf(s, p...)
}
