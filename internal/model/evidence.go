package model

type Evidence struct {
	Kind       string  `json:"kind"`
	Target     string  `json:"target"`
	Strength   string  `json:"strength"`
	Reason     string  `json:"reason"`
	Risk       string  `json:"risk,omitempty"`
	Rule       string  `json:"rule,omitempty"`
	Source     string  `json:"source,omitempty"`
	Confidence float64 `json:"confidence,omitempty"`
}

type EvidenceSummary struct {
	Exact     int `json:"exact"`
	Strong    int `json:"strong"`
	Heuristic int `json:"heuristic"`
}

type EvidenceSet struct {
	Items   []Evidence      `json:"items"`
	Summary EvidenceSummary `json:"summary"`
}
