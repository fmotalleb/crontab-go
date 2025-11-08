package abstraction

import "go.uber.org/zap"

type (
	Validatable interface {
		Validate(log *zap.Logger) error
	}
)
