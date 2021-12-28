package eureka_client

import "log"

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string, err error)
	Error(msg string, err error)
}

type DefaultLogger struct {
}

func (*DefaultLogger) Debug(msg string) {
	log.Println(msg)
}

func (*DefaultLogger) Info(msg string) {
	log.Println(msg)
}

func (*DefaultLogger) Warn(msg string, err error) {
	log.Printf("%s, error: %v", msg, err)
}

func (*DefaultLogger) Error(msg string, err error) {
	log.Printf("%s, error: %v", msg, err)
}

func NewLogger() Logger {
	return &DefaultLogger{}
}
