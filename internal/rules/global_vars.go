package rules

import (
	"fmt"
	"regexp"
	"strings"

	"uniface-linter/internal/parser"
)

// ============================================================
// RULE: GLB001 - Evitar variáveis globais ($var$) em PLs
// ============================================================

type GlobalVarInPLRule struct{}

func (r *GlobalVarInPLRule) ID() string          { return "GLB001" }
func (r *GlobalVarInPLRule) Description() string  { return "PLs devem evitar uso de variáveis globais ($var$); preferir parâmetros" }

var reGlobalVarUsage = regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)\$`)

// Variáveis globais aceitas nas PLs (padrão do framework)
var allowedGlobals = map[string]bool{
	"t_ds_erro":        true,
	"status":           true,
	"procerrorcontext": true,
	"componentname":    true,
}

func (r *GlobalVarInPLRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	entries := ctx.Entries.([]parser.ProcUnit)

	for _, entry := range entries {
		vars := reGlobalVarUsage.FindAllStringSubmatch(entry.Body, -1)
		problematic := make(map[string]int)

		for _, v := range vars {
			varName := strings.ToLower(v[1])
			if allowedGlobals[varName] {
				continue
			}
			problematic[v[0]]++
		}

		if len(problematic) > 0 {
			varNames := make([]string, 0, len(problematic))
			for k := range problematic {
				varNames = append(varNames, k)
			}
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Variáveis Globais",
				Message:    fmt.Sprintf("Entry '%s' usa variáveis globais em PL: %s", entry.Name, strings.Join(varNames, ", ")),
				Location:   fmt.Sprintf("entry %s", entry.Name),
				Line:       entry.LineStart,
				Suggestion: "Nas PLs, prefira passar e receber informações por parâmetros em vez de variáveis globais",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: GLB002 - Variáveis de componente não devem ser mantidas em PLs
// ============================================================

type ComponentVarInPLRule struct{}

func (r *ComponentVarInPLRule) ID() string          { return "GLB002" }
func (r *ComponentVarInPLRule) Description() string  { return "PLs não devem atualizar variáveis de componente (T_VAR)" }

var reComponentVar = regexp.MustCompile(`(?i)\bT_[A-Z_]+\b`)

// Variáveis T_xxx aceitas em PLs (padrão do framework)
var acceptedTVars = map[string]bool{
	"T_DS_ERRO": true,
}

func (r *ComponentVarInPLRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	entries := ctx.Entries.([]parser.ProcUnit)

	for _, entry := range entries {
		compVarsInPL := reComponentVar.FindAllString(entry.Body, -1)

		var problematic []string
		for _, v := range compVarsInPL {
			if acceptedTVars[strings.ToUpper(v)] {
				continue
			}
			problematic = append(problematic, v)
		}

		if len(problematic) > 0 {
			unique := uniqueStrings(problematic)
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityInfo,
				Category:   "Variáveis Globais",
				Message:    fmt.Sprintf("Entry '%s' referencia possíveis variáveis de componente: %s", entry.Name, strings.Join(unique, ", ")),
				Location:   fmt.Sprintf("entry %s", entry.Name),
				Line:       entry.LineStart,
				Suggestion: "Evite manter ou atualizar variáveis de componente (T_xxx) em PLs; use parâmetros",
			})
		}
	}
	return issues
}

func uniqueStrings(sl []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range sl {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
