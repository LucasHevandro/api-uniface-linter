package rules

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"uniface-linter/internal/parser"
)

// ============================================================
// RULE: NOM001 - Nomenclatura de Operations (lowerCamelCase)
// ============================================================

type NamingOperationsRule struct{}

func (r *NamingOperationsRule) ID() string         { return "NOM001" }
func (r *NamingOperationsRule) Description() string { return "Operations devem usar lowerCamelCase com prefixo semântico" }

var validOpPrefixes = []string{
	"post", "get", "find", "put", "del", "cancel",
	"efetivar", "calcular", "atualizar", "in", "findOrInit", "getComUp",
}

func (r *NamingOperationsRule) Run(ctx *RuleContext) []Issue {
	ops := ctx.Operations.([]parser.ProcUnit)
	var issues []Issue

	for _, op := range ops {
		if !isLowerCamelCase(op.Name) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityError,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Operation '%s' não está em lowerCamelCase", op.Name),
				Location:   fmt.Sprintf("operation %s", op.Name),
				Line:       op.LineStart,
				Suggestion: "Use lowerCamelCase: ex. 'postFixacaoMi', 'getClientePk'",
			})
		}

		if !hasValidPrefix(op.Name) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Operation '%s' não inicia com prefixo semântico reconhecido", op.Name),
				Location:   fmt.Sprintf("operation %s", op.Name),
				Line:       op.LineStart,
				Suggestion: fmt.Sprintf("Prefixos válidos: %s", strings.Join(validOpPrefixes, ", ")),
			})
		}
	}
	return issues
}

// ============================================================
// RULE: NOM002 - Nomenclatura de Entries (pl + lowerCamelCase)
// ============================================================

type NamingEntriesRule struct{}

func (r *NamingEntriesRule) ID() string         { return "NOM002" }
func (r *NamingEntriesRule) Description() string { return "Entries devem iniciar com 'pl' e usar lowerCamelCase" }

func (r *NamingEntriesRule) Run(ctx *RuleContext) []Issue {
	entries := ctx.Entries.([]parser.ProcUnit)
	var issues []Issue

	for _, e := range entries {
		if !strings.HasPrefix(e.Name, "pl") {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityError,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Entry '%s' deve iniciar com 'pl'", e.Name),
				Location:   fmt.Sprintf("entry %s", e.Name),
				Line:       e.LineStart,
				Suggestion: fmt.Sprintf("Renomeie para 'pl%s%s'", strings.ToUpper(e.Name[:1]), e.Name[1:]),
			})
			continue
		}
		rest := e.Name[2:]
		if rest != "" && !startsWithUpper(rest) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Entry '%s': após 'pl' esperava-se letra maiúscula (ex: plPostXxx)", e.Name),
				Location:   fmt.Sprintf("entry %s", e.Name),
				Line:       e.LineStart,
				Suggestion: "Padrão correto: plPostFixacaoMi, plGetClientePk",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: NOM003 - #define com UpperCamelCase
// ============================================================

type NamingDefinesRule struct{}

func (r *NamingDefinesRule) ID() string         { return "NOM003" }
func (r *NamingDefinesRule) Description() string { return "#define deve usar UpperCamelCase ou código de mensagem (ESxxx/ENxxx)" }

var reDefineExtract = regexp.MustCompile(`(?m)^#define\s+(\S+)\s*=`)
var reMsgCode = regexp.MustCompile(`^(ES|EN|EW)\d+$`)

// Defines internos do Uniface que devem ser ignorados
var unifaceInternalDefines = map[string]bool{
	"$triggerAbbr":             true,
	"BreakInheritance_cpt_GENERAL": true,
	"BreakInheritance_cpt_INIT":    true,
}

func (r *NamingDefinesRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	seen := make(map[string]bool) // evitar duplicatas
	matches := reDefineExtract.FindAllStringSubmatch(ctx.Script, -1)

	for _, m := range matches {
		name := m[1]

		// Ignorar defines internos do Uniface
		if unifaceInternalDefines[name] {
			continue
		}
		// Ignorar defines de trigger abreviado do Uniface (padrão $xxx)
		if strings.HasPrefix(name, "$") {
			continue
		}
		// Ignorar já reportados
		if seen[name] {
			continue
		}
		seen[name] = true

		// Aceitar: ESxxx, ENxxx (códigos de mensagem) ou UpperCamelCase
		if reMsgCode.MatchString(name) {
			continue
		}
		if !startsWithUpper(name) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("#define '%s' não está em UpperCamelCase", name),
				Location:   "script global",
				Suggestion: "Use UpperCamelCase: ex. 'TpNegocioVenda', 'CdStatusAtivo'",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: NOM004 - Parâmetros devem ter prefixo p
// ============================================================

type NamingParamsRule struct{}

func (r *NamingParamsRule) ID() string         { return "NOM004" }
func (r *NamingParamsRule) Description() string { return "Parâmetros devem iniciar com 'p' (pCdOperador, pStResult)" }

func (r *NamingParamsRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		for _, param := range unit.Params {
			name := param.Name
			// $t_ds_erro$ e similares são especiais
			if strings.Contains(name, "$") {
				continue
			}
			if !strings.HasPrefix(name, "p") {
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityWarning,
					Category:   "Nomenclatura",
					Message:    fmt.Sprintf("Parâmetro '%s' na %s '%s' não inicia com 'p'", name, unit.Kind, unit.Name),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart,
					Suggestion: fmt.Sprintf("Renomeie para 'p%s%s'", strings.ToUpper(name[:1]), name[1:]),
				})
			}
		}
	}
	return issues
}

// ============================================================
// RULE: NOM005 - Ordem de parâmetros padrão
// ============================================================

type ParamOrderRule struct{}

func (r *ParamOrderRule) ID() string         { return "NOM005" }
func (r *ParamOrderRule) Description() string { return "Ordem dos parâmetros: pCdOperador primeiro, $t_ds_erro$ sempre último" }

func (r *ParamOrderRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		params := unit.Params
		if len(params) < 2 {
			continue
		}

		// $t_ds_erro$ deve ser o último
		for i, p := range params {
			if strings.Contains(p.Name, "t_ds_erro") || strings.Contains(p.Name, "ds_erro") {
				if i != len(params)-1 {
					issues = append(issues, Issue{
						RuleID:     r.ID(),
						Severity:   SeverityWarning,
						Category:   "Nomenclatura",
						Message:    fmt.Sprintf("%s '%s': $t_ds_erro$ deve ser o último parâmetro (está na posição %d de %d)", unit.Kind, unit.Name, i+1, len(params)),
						Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
						Line:       unit.LineStart,
						Suggestion: "Convenção: pCdOperador (1º), ..., pInExiste, pStResult, $t_ds_erro$ (último)",
					})
				}
			}
		}

		// pResult/pStResult deve vir antes de $t_ds_erro$, depois de inputs
		// (verificação básica: se existir pStResult e vier depois de $t_ds_erro$, reportar)
		errIdx := -1
		resultIdx := -1
		for i, p := range params {
			name := strings.ToLower(p.Name)
			if strings.Contains(name, "t_ds_erro") {
				errIdx = i
			}
			if (strings.Contains(name, "result") || strings.Contains(name, "presult")) && p.Direction == "out" {
				resultIdx = i
			}
		}
		if errIdx >= 0 && resultIdx >= 0 && resultIdx > errIdx {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityInfo,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("%s '%s': pStResult deveria vir antes de $t_ds_erro$", unit.Kind, unit.Name),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Ordem esperada: ..., pStResult : out, $t_ds_erro$ : out",
			})
		}
	}
	return issues
}

// ============================================================
// Helpers
// ============================================================

func isLowerCamelCase(s string) bool {
	if s == "" {
		return false
	}
	return unicode.IsLower(rune(s[0]))
}

func startsWithUpper(s string) bool {
	if s == "" {
		return false
	}
	return unicode.IsUpper(rune(s[0]))
}

func hasValidPrefix(name string) bool {
	nameLower := strings.ToLower(name)
	for _, p := range validOpPrefixes {
		if strings.HasPrefix(nameLower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
