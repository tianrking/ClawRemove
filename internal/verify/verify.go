package verify

import "github.com/tianrking/ClawRemove/internal/model"

type Rule interface {
	Evaluate(residual *model.Residual)
}

func Classify(evidence model.EvidenceSet, rules []Rule) model.Verification {
	var all []model.Residual
	for _, item := range evidence.Items {
		residual := model.Residual{
			Kind:       item.Kind,
			Target:     item.Target,
			Evidence:   item.Strength,
			Reason:     item.Reason,
			Risk:       item.Risk,
			Rule:       item.Rule,
			Source:     item.Source,
			Confidence: item.Confidence,
		}
		for _, rule := range rules {
			rule.Evaluate(&residual)
		}
		all = append(all, residual)
	}

	verification := model.Verification{
		Verified:  true,
		Residuals: all,
	}

	for _, residual := range all {
		switch residual.Evidence {
		case "exact":
			verification.Confirmed = append(verification.Confirmed, residual)
			verification.Summary.Exact++
		case "strong":
			verification.Confirmed = append(verification.Confirmed, residual)
			verification.Summary.Strong++
		default:
			verification.Investigate = append(verification.Investigate, residual)
			verification.Summary.Heuristic++
		}
	}
	return verification
}
