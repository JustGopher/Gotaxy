package global

import "context"

var (
	Ctx    context.Context
	Cancel context.CancelFunc
)
