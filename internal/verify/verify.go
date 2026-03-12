package verify

import "github.com/tianrking/ClawRemove/internal/model"

func Classify(evidence model.EvidenceSet) model.Verification {
	var all []model.Residual
	for _, item := range evidence.Items {
		all = append(all, model.Residual{
			Kind:     item.Kind,
			Target:   item.Target,
			Evidence: item.Strength,
			Reason:   item.Reason,
			Risk:     item.Risk,
		})
	}

	verification := model.Verification{
		Verified:  true,
		Residuals: all,
		Summary: model.VerificationSummary{
			Exact:     evidence.Summary.Exact,
			Strong:    evidence.Summary.Strong,
			Heuristic: evidence.Summary.Heuristic,
		},
	}
	for _, residual := range all {
		switch residual.Evidence {
		case "exact", "strong":
			verification.Confirmed = append(verification.Confirmed, residual)
		default:
			verification.Investigate = append(verification.Investigate, residual)
		}
	}
	return verification
}
