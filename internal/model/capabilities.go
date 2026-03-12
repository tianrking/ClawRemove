package model

type ProviderSkill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Inputs      []string `json:"inputs,omitempty"`
}

type ProviderTool struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ReadOnly    bool     `json:"readOnly"`
	Targets     []string `json:"targets,omitempty"`
}

type ProviderCapabilities struct {
	Skills []ProviderSkill `json:"skills"`
	Tools  []ProviderTool  `json:"tools"`
}
