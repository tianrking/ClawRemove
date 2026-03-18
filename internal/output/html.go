package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/tianrking/ClawRemove/internal/model"
)

// PrintEnvironmentHTML prints a full environment report as HTML.
func PrintEnvironmentHTML(w io.Writer, report model.EnvironmentReport) error {
	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Environment Report - ClawRemove</title>
    <style>
        :root {
            --primary: #7D56F4;
            --success: #22C55E;
            --warning: #F59E0B;
            --danger: #EF4444;
            --bg: #0F111A;
            --card-bg: #1A1D2E;
            --text: #E5E7EB;
            --text-muted: #9CA3AF;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg);
            color: var(--text);
            margin: 0;
            padding: 20px;
            line-height: 1.6;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: var(--primary); margin-bottom: 10px; }
        h2 { color: var(--text); border-bottom: 2px solid var(--primary); padding-bottom: 10px; margin-top: 30px; }
        .meta { color: var(--text-muted); font-size: 14px; margin-bottom: 30px; }
        .card {
            background: var(--card-bg);
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .card-header { font-weight: 600; margin-bottom: 15px; color: var(--primary); }
        .item { padding: 10px 0; border-bottom: 1px solid #2D3348; }
        .item:last-child { border-bottom: none; }
        .item-name { font-weight: 500; }
        .item-path { color: var(--text-muted); font-size: 13px; margin-top: 4px; }
        .badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
        }
        .badge-success { background: rgba(34, 197, 94, 0.2); color: var(--success); }
        .badge-warning { background: rgba(245, 158, 11, 0.2); color: var(--warning); }
        .badge-danger { background: rgba(239, 68, 68, 0.2); color: var(--danger); }
        .size { color: var(--text-muted); }
        .summary {
            background: linear-gradient(135deg, var(--primary), #9333EA);
            border-radius: 8px;
            padding: 25px;
            margin-top: 30px;
            text-align: center;
        }
        .summary-title { font-size: 14px; opacity: 0.9; }
        .summary-value { font-size: 36px; font-weight: 700; margin-top: 10px; }
        .recommendations { margin-top: 10px; }
        .recommendation { padding: 8px 12px; background: rgba(34, 197, 94, 0.1); border-left: 3px solid var(--success); margin-bottom: 8px; }
        .footer { text-align: center; color: var(--text-muted); margin-top: 40px; font-size: 13px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🤖 AI Environment Report</h1>
        <div class="meta">
            Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `<br>
            Platform: ` + report.Platform + `
        </div>
`)

	// AI Runtime section
	if len(report.Runtime.Detected) > 0 {
		html.WriteString(`        <h2>⚡ AI Runtime</h2>
        <div class="card">
`)
		for _, rt := range report.Runtime.Detected {
			status := "stopped"
			statusClass := "badge-warning"
			if rt.Running {
				status = "running"
				statusClass = "badge-success"
			}
			html.WriteString(fmt.Sprintf(`            <div class="item">
                <span class="item-name">%s</span>
                <span class="badge %s">%s</span>
                <div class="item-path">%s</div>
            </div>
`, rt.Name, statusClass, status, rt.Path))
		}
		html.WriteString(`        </div>
`)
	}

	// AI Tools section
	if len(report.Agents.Applications) > 0 {
		html.WriteString(`        <h2>🔧 AI Tools</h2>
        <div class="card">
`)
		for _, app := range report.Agents.Applications {
			html.WriteString(fmt.Sprintf(`            <div class="item">
                <span class="item-name">%s</span>
                <div class="item-path">%s</div>
            </div>
`, app.Name, app.Path))
		}
		html.WriteString(`        </div>
`)
	}

	// Models section
	if len(report.Artifacts.Models) > 0 {
		html.WriteString(`        <h2>📦 Models</h2>
        <div class="card">
`)
		for _, m := range report.Artifacts.Models {
			html.WriteString(fmt.Sprintf(`            <div class="item">
                <span class="item-name">%s</span>
                <span class="size">%s</span>
                <div class="item-path">%s</div>
            </div>
`, m.Name, formatSize(m.Size), m.Path))
		}
		html.WriteString(`        </div>
`)
	}

	// Caches section
	if len(report.Artifacts.Caches) > 0 {
		html.WriteString(`        <h2>💾 Caches</h2>
        <div class="card">
`)
		for _, c := range report.Artifacts.Caches {
			html.WriteString(fmt.Sprintf(`            <div class="item">
                <span class="item-name">%s</span>
                <span class="size">%s</span>
                <div class="item-path">%s</div>
            </div>
`, c.Name, formatSize(c.Size), c.Path))
		}
		html.WriteString(`        </div>
`)
	}

	// Security findings
	if len(report.Security.Findings) > 0 {
		html.WriteString(`        <h2>🔒 Security Findings</h2>
        <div class="card">
`)
		for _, f := range report.Security.Findings {
			severityClass := "badge-warning"
			if f.Severity == "high" {
				severityClass = "badge-danger"
			}
			html.WriteString(fmt.Sprintf(`            <div class="item">
                <span class="item-name">%s</span>
                <span class="badge %s">%s</span>
                <div class="item-path">%s</div>
            </div>
`, f.Type, severityClass, f.Severity, f.Location))
		}
		html.WriteString(`        </div>
`)
	}

	// Total AI Storage
	html.WriteString(fmt.Sprintf(`        <div class="summary">
            <div class="summary-title">Total AI Storage</div>
            <div class="summary-value">%s</div>
        </div>
`, formatSize(report.Hygiene.TotalSize)))

	// Recommendations
	if len(report.Hygiene.Recommendations) > 0 {
		html.WriteString(`        <h2>💡 Recommendations</h2>
        <div class="recommendations">
`)
		for _, rec := range report.Hygiene.Recommendations {
			html.WriteString(fmt.Sprintf(`            <div class="recommendation">%s</div>
`, rec))
		}
		html.WriteString(`        </div>
`)
	}

	html.WriteString(`
        <div class="footer">
            Generated by <a href="https://github.com/tianrking/ClawRemove" style="color: var(--primary);">ClawRemove</a>
        </div>
    </div>
</body>
</html>`)

	_, err := io.WriteString(w, html.String())
	return err
}

// PrintCleanupHTML prints a cleanup report as HTML.
func PrintCleanupHTML(w io.Writer, report model.CleanupReport) error {
	var html strings.Builder
	html.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Cleanup Report - ClawRemove</title>
    <style>
        :root {
            --primary: #7D56F4;
            --success: #22C55E;
            --warning: #F59E0B;
            --danger: #EF4444;
            --bg: #0F111A;
            --card-bg: #1A1D2E;
            --text: #E5E7EB;
            --text-muted: #9CA3AF;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: var(--bg);
            color: var(--text);
            margin: 0;
            padding: 20px;
            line-height: 1.6;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: var(--primary); margin-bottom: 10px; }
        h2 { color: var(--text); border-bottom: 2px solid var(--primary); padding-bottom: 10px; margin-top: 30px; }
        .meta { color: var(--text-muted); font-size: 14px; margin-bottom: 30px; }
        .card {
            background: var(--card-bg);
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 20px;
        }
        .item { padding: 12px 0; border-bottom: 1px solid #2D3348; }
        .item:last-child { border-bottom: none; }
        .item-reason { font-weight: 500; }
        .item-path { color: var(--text-muted); font-size: 13px; margin-top: 4px; }
        .badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 500;
        }
        .badge-low { background: rgba(34, 197, 94, 0.2); color: var(--success); }
        .badge-medium { background: rgba(245, 158, 11, 0.2); color: var(--warning); }
        .badge-high { background: rgba(239, 68, 68, 0.2); color: var(--danger); }
        .size { color: var(--text-muted); }
        .summary {
            background: linear-gradient(135deg, var(--primary), #9333EA);
            border-radius: 8px;
            padding: 25px;
            margin-top: 30px;
            text-align: center;
        }
        .summary-title { font-size: 14px; opacity: 0.9; }
        .summary-value { font-size: 36px; font-weight: 700; margin-top: 10px; }
        .empty { text-align: center; padding: 40px; color: var(--text-muted); }
        .footer { text-align: center; color: var(--text-muted); margin-top: 40px; font-size: 13px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🧹 AI Cleanup Report</h1>
        <div class="meta">
            Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `<br>
            Candidates: ` + fmt.Sprintf("%d", len(report.Candidates)) + `
        </div>
`)

	if len(report.Candidates) == 0 {
		html.WriteString(`        <div class="empty">
            <p>No cleanup candidates found. Your system is clean!</p>
        </div>
`)
	} else {
		// Group by category
		categories := map[string][]model.CleanupCandidate{}
		categoryOrder := []string{"model_version", "orphaned_cache", "unused_vectordb", "log_rotation"}
		for _, c := range report.Candidates {
			categories[c.Category] = append(categories[c.Category], c)
		}

		categoryNames := map[string]string{
			"model_version":    "📦 Old Model Versions",
			"orphaned_cache":   "🗑️ Orphaned Caches",
			"unused_vectordb":  "🗄️ Unused Vector Databases",
			"log_rotation":     "📝 Log Files",
		}

		for _, cat := range categoryOrder {
			items, ok := categories[cat]
			if !ok || len(items) == 0 {
				continue
			}

			catName := categoryNames[cat]
			if catName == "" {
				catName = cat
			}

			var catSize int64
			for _, item := range items {
				catSize += item.Size
			}

			html.WriteString(fmt.Sprintf(`        <h2>%s <span class="size">(%s)</span></h2>
        <div class="card">
`, catName, formatSize(catSize)))

			for _, item := range items {
				riskClass := "badge-low"
				if item.Risk == "medium" {
					riskClass = "badge-medium"
				} else if item.Risk == "high" {
					riskClass = "badge-high"
				}
				html.WriteString(fmt.Sprintf(`            <div class="item">
                <span class="item-reason">%s</span>
                <span class="badge %s">%s</span>
                <div class="item-path">%s (%s)</div>
            </div>
`, item.Reason, riskClass, item.Risk, item.Path, formatSize(item.Size)))
			}
			html.WriteString(`        </div>
`)
		}
	}

	// Summary
	html.WriteString(fmt.Sprintf(`        <div class="summary">
            <div class="summary-title">Total Reclaimable</div>
            <div class="summary-value">%s</div>
        </div>
`, formatSize(report.TotalReclaimable)))

	html.WriteString(`
        <div class="footer">
            Generated by <a href="https://github.com/tianrking/ClawRemove" style="color: var(--primary);">ClawRemove</a>
        </div>
    </div>
</body>
</html>`)

	_, err := io.WriteString(w, html.String())
	return err
}
