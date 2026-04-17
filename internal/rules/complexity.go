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

func (r *ProcSizeRule) ID() string { return "COMP001" }
func (r *ProcSizeRule) Description() string {
	max := r.MaxLines
	if max == 0 {
		max = defaultMaxLines
	}
	return fmt.Sprintf("PROCs não devem exceder %d linhas de código", max)
}

func (r *ProcSizeRule) Run(ctx *RuleContext) []Issue {
	if r.MaxLines == 0 {
		r.MaxLines = defaultMaxLines
	}
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		codeLines := countCodeLines(unit.Body)
		if codeLines <= r.MaxLines {
			continue
		}

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
	return issues
}

// ============================================================
// RULE: COMP002 - Operation deve delegar para PL
// ============================================================

type OperationDelegatesRule struct{}

func (r *OperationDelegatesRule) ID() string { return "COMP002" }
func (r *OperationDelegatesRule) Description() string {
	return "Operations devem delegar lógica para PLs (entries pl...)"
}

// Detecta chamadas a entries: call plXxx ou activate "xxx".plXxx
var reCallPL = regexp.MustCompile(`(?i)\bcall\s+pl[A-Za-z]`)

func (r *OperationDelegatesRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	ops := ctx.Operations.([]parser.ProcUnit)

	for _, op := range ops {
		codeLines := countCodeLines(op.Body)
		if codeLines <= 10 {
			continue
		}
		if reCallPL.MatchString(op.Body) {
			continue
		}
		issues = append(issues, Issue{
			RuleID:     r.ID(),
			Severity:   SeverityWarning,
			Category:   "Complexidade",
			Message:    fmt.Sprintf("Operation '%s' tem %d linhas mas não delega para nenhuma PL (call pl...)", op.Name, codeLines),
			Location:   fmt.Sprintf("operation %s", op.Name),
			Line:       op.LineStart,
			Suggestion: "Operations devem ter o mínimo de código, delegando a lógica para PLs (entries)",
		})
	}
	return issues
}

// ============================================================
// RULE: COMP003 - Número excessivo de parâmetros sem struct
// ============================================================

type ParamCountRule struct{}

func (r *ParamCountRule) ID() string { return "COMP003" }
func (r *ParamCountRule) Description() string {
	return "PROCs com muitos parâmetros devem usar struct para agrupá-los"
}

const maxParams = 7

func (r *ParamCountRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		hasStruct := false
		for _, p := range unit.Params {
			if strings.EqualFold(p.Type, "struct") {
				hasStruct = true
				break
			}
		}

		total := len(unit.Params)
		if total > maxParams && !hasStruct {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Complexidade",
				Message:    fmt.Sprintf("%s '%s' tem %d parâmetros sem uso de struct", unit.Kind, unit.Name, total),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Agrupe parâmetros relacionados em structs para reduzir o acoplamento",
			})
		}
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
