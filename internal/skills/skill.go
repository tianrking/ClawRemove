package skills

import (
	"context"

	"github.com/tianrking/ClawRemove/internal/model"
)

type Skill interface {
	Info() model.ProviderSkill
	Analyze(ctx context.Context, report model.Report) (any, error)
}
