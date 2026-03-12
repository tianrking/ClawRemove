package skills

import "github.com/tianrking/ClawRemove/internal/model"

func Catalog(capabilities model.ProviderCapabilities) []model.ProviderSkill {
	return capabilities.Skills
}
