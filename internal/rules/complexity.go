package rules

import (
	"fmt"
	"regexp"
	"strings"

	"uniface-linter/internal/parser"
)

const defaultMaxLines = 100

// ============================================================
// RULE: COMP001 - Tamanho máximo de PROCs
// ============================================================

type ProcSizeRule struct {
	MaxLines int
}

func (r *ProcSizeRule) ID() string          { return "COMP001" }
func (r *ProcSizeRule) Description() string  { return fmt.Sprintf("PROCs não devem exceder %d linhas de código", r.MaxLines) }

func (r *ProcSizeRule) Run(ctx *RuleContext) []Issue {
	if r.MaxLines == 0 {
		r.MaxLines = defaultMaxLines
	}
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		lines := strings.Split(unit.Body, "\n")
		// Contar apenas linhas não-vazias e não-comentário
		codeLines := 0
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if trimmed != "" && !strings.HasPrefix(trimmed, ";") && !strings.HasPrefix(trimmed, "#") {
				codeLines++
			}
		}

		if codeLines > r.MaxLines {
			sev := SeverityWarning
			if codeLines > r.MaxLines*2 {
				sev = SeverityError
			}
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   sev,
				Category:   "Complexidade",
				Message:    fmt.Sprintf("%s '%s' tem %d linhas de código (máximo: %d)", unit.Kind, unit.Name, codeLines, r.MaxLines),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Extraia lógica em PLs menores para melhor modularização",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: COMP002 - Operation deve delegar para PL
// ============================================================

type OperationDelegatesRule struct{}

func (r *OperationDelegatesRule) ID() string          { return "COMP002" }
func (r *OperationDelegatesRule) Description() string  { return "Operations devem delegar lógica para PLs (entries)" }

var reCallPl = regexp.MustCompile(`(?i)\bcall\s+pl[A-Z]`)

func (r *OperationDelegatesRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	ops := ctx.Operations.([]parser.ProcUnit)

	for _, op := range ops {
		body := op.Body
		codeLines := countCodeLines(body)

		// Se a operation tem mais de 10 linhas e não chama nenhuma PL
		if codeLines > 10 && !reCallPl.MatchString(body) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Complexidade",
				Message:    fmt.Sprintf("Operation '%s' tem %d linhas mas não delega para nenhuma PL (entry pl...)", op.Name, codeLines),
				Location:   fmt.Sprintf("operation %s", op.Name),
				Line:       op.LineStart,
				Suggestion: "Operations devem ter o mínimo de código, delegando o serviço para PLs",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: COMP003 - Número de parâmetros excessivo
// ============================================================

type ParamCountRule struct{}

func (r *ParamCountRule) ID() string          { return "COMP003" }
func (r *ParamCountRule) Description() string  { return "PROCs com muitos parâmetros devem usar struct" }

func (r *ParamCountRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)
	const maxParams = 7

	for _, unit := range all {
		// Contar parâmetros não-struct
		nonStructParams := 0
		hasStruct := false
		for _, p := range unit.Params {
			if p.Type == "struct" {
				hasStruct = true
			} else {
				nonStructParams++
			}
		}

		totalParams := len(unit.Params)
		if totalParams > maxParams && !hasStruct {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Complexidade",
				Message:    fmt.Sprintf("%s '%s' tem %d parâmetros sem uso de struct", unit.Kind, unit.Name, totalParams),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Agrupe parâmetros relacionados em structs para reduzir o acoplamento",
			})
		}
		_ = hasStruct
		_ = nonStructParams
	}
	return issues
}

// ============================================================
// Helpers
// ============================================================

func countCodeLines(body string) int {
	count := 0
	for _, l := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" && !strings.HasPrefix(trimmed, ";") && !strings.HasPrefix(trimmed, "#") {
			count++
		}
	}
	return count
}
