package accrual

import "go.uber.org/zap"

// zapLoggerAdapter адаптирует zap.Logger к интерфейсу resty.Logger
type zapLoggerAdapter struct {
	logger *zap.Logger
}

func (z *zapLoggerAdapter) Errorf(format string, v ...any) {
	z.logger.Sugar().Errorf(format, v...)
}

func (z *zapLoggerAdapter) Warnf(format string, v ...any) {
	z.logger.Sugar().Warnf(format, v...)
}

func (z *zapLoggerAdapter) Debugf(format string, v ...any) {
	z.logger.Sugar().Debugf(format, v...)
}
