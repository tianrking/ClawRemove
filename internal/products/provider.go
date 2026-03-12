package products

import (
	"fmt"

	"github.com/tianrking/ClawRemove/internal/model"
	"github.com/tianrking/ClawRemove/internal/products/openclaw"
	"github.com/tianrking/ClawRemove/internal/skills"
	"github.com/tianrking/ClawRemove/internal/tools"
	"github.com/tianrking/ClawRemove/internal/verify"
)

type Provider interface {
	ID() string
	DisplayName() string
	Facts() model.ProductFacts
	Capabilities() model.ProviderCapabilities
	Tools() []tools.Tool
	Skills() []skills.Skill
	VerificationRules() []verify.Rule
}

func Registry() []Provider {
	return []Provider{
		openclaw.Provider{},
	}
}

func Resolve(id string) (Provider, error) {
	for _, provider := range Registry() {
		if provider.ID() == id {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("unknown product provider: %s", id)
}
