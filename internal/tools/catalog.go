package tools

import "github.com/tianrking/ClawRemove/internal/model"

func Catalog(capabilities model.ProviderCapabilities) []model.ProviderTool {
	return capabilities.Tools
}
