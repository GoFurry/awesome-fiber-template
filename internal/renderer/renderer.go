package renderer

import "github.com/GoFurry/fiberx/internal/planner"

type Result struct {
	PreviewFiles    []string
	Warnings        []string
	ReplaceRuleHits []string
	InjectionHits   []string
}

func Render(plan planner.Plan) Result {
	warnings := []string{
		"Phase 2 wires the generator pipeline only; manifests and generator assets arrive in later phases.",
	}

	if len(plan.Capabilities) == 0 {
		warnings = append(warnings, "No capabilities are wired into the Phase 2 skeleton yet.")
	}

	replaceRuleHits := make([]string, 0, len(plan.ReplaceRules))
	for _, rule := range plan.ReplaceRules {
		replaceRuleHits = append(replaceRuleHits, rule.Name)
	}

	injectionHits := make([]string, 0, len(plan.InjectionRules))
	for _, rule := range plan.InjectionRules {
		injectionHits = append(injectionHits, rule.Name)
	}

	return Result{
		PreviewFiles:    []string{},
		Warnings:        warnings,
		ReplaceRuleHits: replaceRuleHits,
		InjectionHits:   injectionHits,
	}
}
