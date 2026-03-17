package inventory

import (
	"fmt"
	"strings"

	"github.com/tianrking/ClawRemove/internal/model"
)

// BuildRuntimeSection converts inventory runtime data into a report section.
func BuildRuntimeSection(inv AIInventory) model.RuntimeSection {
	var items []model.RuntimeItem
	for _, rt := range inv.Runtimes {
		var port int
		if len(rt.Ports) > 0 {
			port = rt.Ports[0]
		}
		items = append(items, model.RuntimeItem{
			Name:    rt.Name,
			Version: rt.Version,
			Path:    rt.Path,
			Running: rt.Running,
			Port:    port,
		})
	}

	summary := "No AI runtimes detected"
	if len(items) > 0 {
		running := 0
		for _, item := range items {
			if item.Running {
				running++
			}
		}
		summary = fmt.Sprintf("%d runtime(s) found, %d running", len(items), running)
	}

	return model.RuntimeSection{
		Detected: items,
		Summary:  summary,
	}
}

// BuildAgentsSection converts inventory agent data into a report section.
func BuildAgentsSection(inv AIInventory) model.AgentsSection {
	var apps []model.AgentItem
	for _, agent := range inv.Agents {
		apps = append(apps, model.AgentItem{
			Name:    agent.Name,
			Path:    agent.Path,
			Version: agent.Version,
		})
	}

	var frameworks []model.AgentItem
	for _, fw := range inv.Frameworks {
		frameworks = append(frameworks, model.AgentItem{
			Name:    fw.Name,
			Version: fw.Version,
			Path:    fw.Path,
			Manager: fw.Manager,
		})
	}

	summary := "No AI agents detected"
	if len(apps) > 0 || len(frameworks) > 0 {
		parts := []string{}
		if len(apps) > 0 {
			parts = append(parts, fmt.Sprintf("%d application(s)", len(apps)))
		}
		if len(frameworks) > 0 {
			parts = append(parts, fmt.Sprintf("%d framework(s)", len(frameworks)))
		}
		summary = strings.Join(parts, ", ") + " found"
	}

	return model.AgentsSection{
		Applications: apps,
		Frameworks:   frameworks,
		Summary:      summary,
	}
}

// BuildArtifactsSection converts inventory cache data into a report section.
func BuildArtifactsSection(inv AIInventory) model.ArtifactsSection {
	var models []model.ArtifactItem
	var caches []model.ArtifactItem
	var vectorDBs []model.ArtifactItem
	var totalSize int64

	for _, cache := range inv.ModelCaches {
		item := model.ArtifactItem{
			Name: cache.Name,
			Path: cache.Path,
			Size: cache.Size,
		}
		// Classify by type field if available, otherwise by name
		cacheType := strings.ToLower(cache.Type)
		cacheName := strings.ToLower(cache.Name)
		if cacheType == "model" || strings.Contains(cacheName, "model") {
			models = append(models, item)
		} else {
			caches = append(caches, item)
		}
		totalSize += cache.Size
	}

	for _, vs := range inv.VectorStores {
		vectorDBs = append(vectorDBs, model.ArtifactItem{
			Name: vs.Name,
			Path: vs.Path,
		})
	}

	summary := "No AI artifacts detected"
	if totalSize > 0 {
		summary = fmt.Sprintf("Total: %s", FormatSize(totalSize))
	}

	return model.ArtifactsSection{
		Models:    models,
		VectorDBs: vectorDBs,
		Caches:    caches,
		TotalSize: totalSize,
		Summary:   summary,
	}
}

// BuildHygieneSection computes storage hygiene analysis from inventory data.
func BuildHygieneSection(inv AIInventory, artifacts model.ArtifactsSection) model.HygieneSection {
	var modelsSize, cacheSize, vectorDBSize, logSize int64
	var recommendations []string

	for _, cache := range inv.ModelCaches {
		cacheType := strings.ToLower(cache.Type)
		cacheName := strings.ToLower(cache.Name)

		// Classify by type first, then by name patterns
		switch {
		case cacheType == "model":
			modelsSize += cache.Size
		case cacheType == "cache":
			cacheSize += cache.Size
		case strings.Contains(cacheName, "model") || strings.Contains(cacheName, "ollama") || strings.Contains(cacheName, "lmstudio"):
			modelsSize += cache.Size
		case strings.Contains(cacheName, "cache"):
			cacheSize += cache.Size
		default:
			cacheSize += cache.Size
		}
	}

	if modelsSize > 10*1024*1024*1024 {
		recommendations = append(recommendations, "Consider cleaning old model versions to free up space (use 'claw-remove cleanup --category model_version')")
	}
	if cacheSize > 5*1024*1024*1024 {
		recommendations = append(recommendations, "Large cache detected - review and clean unused caches (use 'claw-remove cleanup --category orphaned_cache')")
	}

	totalSize := modelsSize + cacheSize + vectorDBSize + logSize

	summary := "Storage healthy"
	if totalSize > 50*1024*1024*1024 {
		summary = fmt.Sprintf("High storage usage: %s", FormatSize(totalSize))
	} else if totalSize > 10*1024*1024*1024 {
		summary = fmt.Sprintf("Moderate storage usage: %s", FormatSize(totalSize))
	}

	return model.HygieneSection{
		ModelsSize:      modelsSize,
		CacheSize:       cacheSize,
		VectorDBSize:    vectorDBSize,
		LogSize:         logSize,
		TotalSize:       totalSize,
		Recommendations: recommendations,
		Summary:         summary,
	}
}

// FormatSize formats byte count to human-readable string.
func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1fTB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1fGB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1fMB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1fKB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}
