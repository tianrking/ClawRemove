package security

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Scanner performs AI tool-specific security checks.
// It only scans AI tool configuration files, not general file systems.
type Scanner struct {
	home string
}

// NewScanner creates a new AI security scanner.
func NewScanner() *Scanner {
	home, _ := os.UserHomeDir()
	return &Scanner{home: home}
}

// SecurityFinding represents a security finding.
type SecurityFinding struct {
	Type        string `json:"type"`        // api_key_exposed, token_in_config, etc.
	Provider    string `json:"provider"`    // openai, anthropic, cursor, etc.
	Location    string `json:"location"`    // file path or env var
	Line        int    `json:"line,omitempty"`
	Severity    string `json:"severity"`    // high, medium, low
	Remediation string `json:"remediation"` // how to fix
}

// SecurityReport contains security findings.
type SecurityReport struct {
	Findings []SecurityFinding `json:"findings"`
	Summary  SecuritySummary   `json:"summary"`
}

// SecuritySummary contains summary statistics.
type SecuritySummary struct {
	Total      int `json:"total"`
	HighRisk   int `json:"highRisk"`
	MediumRisk int `json:"mediumRisk"`
	LowRisk    int `json:"lowRisk"`
}

// AI-specific configuration paths to check
var aiConfigPaths = []struct {
	path     string
	provider string
}{
	// OpenAI
	{".openai/api_key.txt", "openai"},
	{".openai/token", "openai"},
	// Anthropic/Claude
	{".config/claude/credentials.json", "anthropic"},
	{".claude/api_key", "anthropic"},
	// Cursor
	{".cursor/api_key.txt", "cursor"},
	{".cursor/config.json", "cursor"},
	// OpenClaw
	{".openclaw/credentials/oauth.json", "openclaw"},
	{".openclaw/.env", "openclaw"},
	// NanoBot
	{".nanobot/.env", "nanobot"},
	// PicoClaw
	{".picoclaw/.env", "picoclaw"},
	// Aider
	{".aider/api_key", "aider"},
	// General AI tool env files
	{".env", "general"},
	{".env.local", "general"},
	{".env.development", "general"},
}

// AI-specific environment variables
var aiEnvVars = []struct {
	name     string
	provider string
}{
	{"OPENAI_API_KEY", "openai"},
	{"ANTHROPIC_API_KEY", "anthropic"},
	{"CLAUDE_API_KEY", "anthropic"},
	{"GOOGLE_API_KEY", "google"},
	{"GEMINI_API_KEY", "google"},
	{"AZURE_OPENAI_API_KEY", "azure"},
	{"HUGGINGFACE_TOKEN", "huggingface"},
	{"HF_TOKEN", "huggingface"},
	{"REPLICATE_API_TOKEN", "replicate"},
	{"COHERE_API_KEY", "cohere"},
	{"OPENROUTER_API_KEY", "openrouter"},
	{"ZHIPU_API_KEY", "zhipu"},
	{"BIGMODEL_API_KEY", "zhipu"},
}

// API key patterns (AI-specific only)
var apiKeyPatterns = []struct {
	name     string
	pattern  *regexp.Regexp
	provider string
}{
	{"OpenAI API Key", regexp.MustCompile(`sk-[a-zA-Z0-9]{48,}`), "openai"},
	{"Anthropic API Key", regexp.MustCompile(`sk-ant-[a-zA-Z0-9-_]{80,}`), "anthropic"},
	{"Google API Key", regexp.MustCompile(`AIza[a-zA-Z0-9_-]{35}`), "google"},
	{"HuggingFace Token", regexp.MustCompile(`hf_[a-zA-Z0-9]{34}`), "huggingface"},
}

// Scan performs AI tool security checks only.
func (s *Scanner) Scan() SecurityReport {
	var findings []SecurityFinding

	// Check AI environment variables
	findings = append(findings, s.scanAIEnvVars()...)

	// Check AI configuration files
	findings = append(findings, s.scanAIConfigFiles()...)

	return SecurityReport{
		Findings: findings,
		Summary:  summarizeFindings(findings),
	}
}

// scanAIEnvVars checks AI-specific environment variables.
func (s *Scanner) scanAIEnvVars() []SecurityFinding {
	var findings []SecurityFinding

	for _, env := range aiEnvVars {
		value := os.Getenv(env.name)
		if value != "" {
			// Check if the value looks like a real API key
			if len(value) >= 20 {
				findings = append(findings, SecurityFinding{
					Type:        "api_key_in_env",
					Provider:    env.provider,
					Location:    env.name + " (environment variable)",
					Severity:    "high",
					Remediation: "Use a secure secret manager or .env file not in shell profile",
				})
			}
		}
	}

	return findings
}

// scanAIConfigFiles scans AI tool configuration files for exposed secrets.
func (s *Scanner) scanAIConfigFiles() []SecurityFinding {
	var findings []SecurityFinding

	for _, config := range aiConfigPaths {
		fullPath := filepath.Join(s.home, config.path)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		// Check for API key patterns
		findings = append(findings, s.scanContent(string(content), fullPath, config.provider)...)
	}

	return findings
}

// scanContent scans file content for API keys.
func (s *Scanner) scanContent(content, path, provider string) []SecurityFinding {
	var findings []SecurityFinding

	lines := strings.Split(content, "\n")
	for lineNum, line := range lines {
		for _, pattern := range apiKeyPatterns {
			if pattern.pattern.MatchString(line) {
				findings = append(findings, SecurityFinding{
					Type:        "api_key_in_file",
					Provider:    pattern.provider,
					Location:    path,
					Line:        lineNum + 1,
					Severity:    "high",
					Remediation: "Remove API key and use environment variable or secret manager",
				})
			}
		}

		// Check for generic API key assignments
		lowerLine := strings.ToLower(line)
		if (strings.Contains(lowerLine, "api_key") ||
			strings.Contains(lowerLine, "api-key") ||
			strings.Contains(lowerLine, "apikey")) &&
			strings.Contains(line, "=") &&
			!strings.Contains(line, "your_") &&
			!strings.Contains(line, "YOUR_") &&
			!strings.HasPrefix(strings.TrimSpace(line), "#") {
			findings = append(findings, SecurityFinding{
				Type:        "potential_api_key",
				Provider:    provider,
				Location:    path,
				Line:        lineNum + 1,
				Severity:    "medium",
				Remediation: "Verify this is not a real API key",
			})
		}
	}

	return findings
}

// summarizeFindings creates a summary.
func summarizeFindings(findings []SecurityFinding) SecuritySummary {
	var summary SecuritySummary
	summary.Total = len(findings)

	for _, f := range findings {
		switch f.Severity {
		case "high":
			summary.HighRisk++
		case "medium":
			summary.MediumRisk++
		default:
			summary.LowRisk++
		}
	}

	return summary
}
