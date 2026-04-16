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
	return "Todo activate ou call deve ser seguido de #include lib_coamo:g_vld_erro ou return<g_erroexec>"
}

var (
	reActivateCall   = regexp.MustCompile(`(?i)^\s*(activate|call)\s+`)
	reErrInclude     = regexp.MustCompile(`(?i)^\s*#include\s+lib_coamo:g_vld_erro`)
	reReturnErroExec = regexp.MustCompile(`(?i)^\s*return\s*<g_erroexec>`)
	reTestStatus     = regexp.MustCompile(`(?i)if\s*\(?\s*\$status\s*<\s*0\s*\)?`)
)

func isErrHandled(line string) bool {
	return reErrInclude.MatchString(line) || reReturnErroExec.MatchString(line) || reTestStatus.MatchString(line)
}

// isContinuationLine reporta se a linha termina com %\ (continuação de linha Uniface)
func isContinuationLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasSuffix(trimmed, `%\`)
}

func (r *ErrorHandlingRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		lines := strings.Split(unit.Body, "\n")

		for i, line := range lines {
			if !reActivateCall.MatchString(line) {
				continue
			}

			// Avançar enquanto a linha atual for continuação (%\)
			// O statement só termina na linha que não tem %\ no final
			end := i
			for end < len(lines) && isContinuationLine(lines[end]) {
				end++
			}

			// Próxima linha não-vazia após o fim do statement deve tratar o erro
			handled := false
			for j := end + 1; j < len(lines); j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" {
					continue
				}
				handled = isErrHandled(next)
				break
			}

			if !handled {
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityWarning,
					Category:   "Tratamento de Erros",
					Message:    fmt.Sprintf("%s '%s': '%s' sem tratamento de erro na linha seguinte", unit.Kind, unit.Name, strings.TrimSpace(line)),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart + i,
					Suggestion: "Adicione '#include lib_coamo:g_vld_erro' ou 'return<g_erroexec>' imediatamente após cada activate ou call",
				})
			}
		}
	}
	return issues
}
