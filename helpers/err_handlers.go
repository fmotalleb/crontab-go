// Package helpers provides helper functions.
package helpers

import "go.uber.org/zap"

func PanicOnErr(log *zap.Logger, errorCatcher func() error, message string) {
	if err := errorCatcher(); err != nil {
		log.Panic(message, zap.Error(err))
	}
}

func FatalOnErr(log *zap.Logger, errorCatcher func() error, message string) {
	if err := errorCatcher(); err != nil {
		log.Fatal(message, zap.Error(err))
	}
}

func WarnOnErr(log *zap.Logger, errorCatcher func() error, message string) error {
	if err := errorCatcher(); err != nil {
		log.Warn(message, zap.Error(err))
		return err
	}
	return nil
}

func WarnOnErrIgnored(log *zap.Logger, errorCatcher func() error, message string) {
	if err := errorCatcher(); err != nil {
		log.Warn(message, zap.Error(err))
	}
}
