package tools

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
)

type Tool interface {
	Info() model.ProviderTool
	Execute(ctx context.Context, report model.Report, input map[string]any) (any, error)
}
