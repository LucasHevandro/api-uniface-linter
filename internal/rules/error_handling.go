package rules

import (
	"fmt"
	"regexp"
	"strings"

	"uniface-linter/internal/parser"
)

// ============================================================
// RULE: ERR001 - #include de tratamento de erro após activate/call
// ============================================================

type ErrorHandlingRule struct{}

func (r *ErrorHandlingRule) ID() string { return "ERR001" }
func (r *ErrorHandlingRule) Description() string {
	return "Todo activate ou call deve ser seguido de #include lib_coamo:g_vld_erro"
}

var (
	reActivateCall = regexp.MustCompile(`(?i)^\s*(activate|call)\s+`)
	reErrInclude   = regexp.MustCompile(`(?i)^\s*#include\s+lib_coamo:g_vld_erro`)
)

func (r *ErrorHandlingRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		lines := strings.Split(unit.Body, "\n")

		for i, line := range lines {
			if !reActivateCall.MatchString(line) {
				continue
			}

			// Próxima linha não-vazia deve ser o #include de erro
			hasErrInclude := false
			for j := i + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" {
					continue
				}
				hasErrInclude = reErrInclude.MatchString(next)
				break
			}

			if !hasErrInclude {
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityWarning,
					Category:   "Tratamento de Erros",
					Message:    fmt.Sprintf("%s '%s': '%s' sem #include lib_coamo:g_vld_erro na linha seguinte", unit.Kind, unit.Name, strings.TrimSpace(line)),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart + i,
					Suggestion: "Adicione '#include lib_coamo:g_vld_erro' imediatamente após cada activate ou call",
				})
			}
		}
	}
	return issues
}
