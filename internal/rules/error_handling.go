package rules

import (
	"fmt"
	"regexp"
	"strings"

	"uniface-linter/internal/parser"
)

// ============================================================
// RULE: ERR001 - Uso de include de erro após activate/call
// ============================================================

type ErrorHandlingRule struct{}

func (r *ErrorHandlingRule) ID() string          { return "ERR001" }
func (r *ErrorHandlingRule) Description() string  { return "Chamadas activate/call devem ser seguidas de #include de tratamento de erro" }

var (
	reActivate    = regexp.MustCompile(`(?m)^\s*activate\s+`)
	reCall        = regexp.MustCompile(`(?m)^\s*call\s+`)
	reErrInclude  = regexp.MustCompile(`(?m)^\s*#include\s+lib_coamo:g_vld_err`)
	reComponentTo = regexp.MustCompile(`(?m)^\s*componentToStruct\s+`)
)

func (r *ErrorHandlingRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		lines := strings.Split(unit.Body, "\n")
		uncheckedCalls := 0

		for i, line := range lines {
			trimmed := strings.TrimSpace(line)

			isCall := reActivate.MatchString(line) || reCall.MatchString(line) || reComponentTo.MatchString(line)
			if !isCall {
				continue
			}

			// Verificar se a próxima linha não-vazia contém o #include de erro
			hasErrCheck := false
			for j := i + 1; j < len(lines) && j < i+4; j++ {
				next := strings.TrimSpace(lines[j])
				if next == "" {
					continue
				}
				if reErrInclude.MatchString(next) {
					hasErrCheck = true
				}
				break
			}

			if !hasErrCheck {
				_ = trimmed
				uncheckedCalls++
			}
		}

		if uncheckedCalls > 0 {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Tratamento de Erros",
				Message:    fmt.Sprintf("%s '%s' tem %d chamada(s) activate/call sem #include de verificação de erro logo após", unit.Kind, unit.Name, uncheckedCalls),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Adicione '#include lib_coamo:g_vld_erro' após cada activate ou call",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: ERR002 - Verificação de status raw ($status < 0)
// ============================================================

type RawStatusCheckRule struct{}

func (r *RawStatusCheckRule) ID() string          { return "ERR002" }
func (r *RawStatusCheckRule) Description() string  { return "Evitar verificação manual de $status; usar includes padronizados" }

var reRawStatus = regexp.MustCompile(`(?m)if\s*\(\s*\$status\s*<\s*0`)

func (r *RawStatusCheckRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		matches := reRawStatus.FindAllStringIndex(unit.Body, -1)
		if len(matches) > 3 {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityInfo,
				Category:   "Tratamento de Erros",
				Message:    fmt.Sprintf("%s '%s' tem %d verificações manuais de $status < 0 (prefira includes padronizados)", unit.Kind, unit.Name, len(matches)),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Prefira: #include lib_coamo:g_vld_erro ou #include lib_coamo:g_vld_erroMsg",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: ERR003 - $t_ds_erro$ deve ser parâmetro de saída
// ============================================================

type ErrorParamRule struct{}

func (r *ErrorParamRule) ID() string          { return "ERR003" }
func (r *ErrorParamRule) Description() string  { return "$t_ds_erro$ deve ser sempre parâmetro 'out' e o último parâmetro" }

func (r *ErrorParamRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		// Procurar se o corpo usa $t_ds_erro$ mas não tem como parâmetro
		if strings.Contains(unit.Body, "$t_ds_erro$") {
			hasErrParam := false
			errParamIsLast := false

			params := unit.Params
			for i, p := range params {
				if strings.Contains(p.Name, "t_ds_erro") || strings.Contains(p.Name, "erro") {
					hasErrParam = true
					if i == len(params)-1 {
						errParamIsLast = true
					}
					if p.Direction != "out" {
						issues = append(issues, Issue{
							RuleID:     r.ID(),
							Severity:   SeverityWarning,
							Category:   "Tratamento de Erros",
							Message:    fmt.Sprintf("%s '%s': parâmetro de erro '%s' deveria ser 'out'", unit.Kind, unit.Name, p.Name),
							Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
							Line:       unit.LineStart,
						})
					}
				}
			}

			if hasErrParam && !errParamIsLast {
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityWarning,
					Category:   "Tratamento de Erros",
					Message:    fmt.Sprintf("%s '%s': $t_ds_erro$ deve ser o último parâmetro", unit.Kind, unit.Name),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart,
					Suggestion: "Convenção: $t_ds_erro$ sempre o último quando necessário",
				})
			}
		}
	}
	return issues
}

// ============================================================
// RULE: ERR004 - Uso de números mágicos sem #define
// ============================================================

type MagicNumberRule struct{}

func (r *MagicNumberRule) ID() string          { return "ERR004" }
func (r *MagicNumberRule) Description() string  { return "Evitar números mágicos; usar constantes #define" }

// Detecta literais numéricas > 1 em comparações/atribuições que não sejam 0, 1, -1
var reMagicNum = regexp.MustCompile(`(?m)[=!<>]\s*([2-9]\d{1,}|[2-9])\b`)

func (r *MagicNumberRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		matches := reMagicNum.FindAllString(unit.Body, -1)
		// Filtrar falsos positivos (comparações com -2, 100, etc.)
		suspicious := 0
		for _, m := range matches {
			if !strings.Contains(m, "-2") { // -2 é status do Uniface
				suspicious++
			}
		}

		if suspicious > 3 {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityInfo,
				Category:   "Tratamento de Erros",
				Message:    fmt.Sprintf("%s '%s' pode conter números mágicos (%d ocorrências suspeitas)", unit.Kind, unit.Name, suspicious),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Defina constantes com #define: ex. '#define TpNegocioVenda = 1;Tipo Venda'",
			})
		}
	}
	return issues
}
