package rules

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"uniface-linter/internal/parser"
)

// ============================================================
// Helpers compartilhados
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

// ============================================================
// RULE: NOM001 - Nomenclatura de Operations (lowerCamelCase + prefixo semântico)
// ============================================================

// Prefixos de services CRUD
var crudPrefixes = []string{"post", "get", "find", "put", "del", "cancel"}

// Prefixos de services de orquestração (verbos no infinitivo)
var servicePrefixes = []string{
	"efetivar", "cancelar", "gravar", "buscar", "calcular", "atualizar",
	"processar", "validar", "gerar", "enviar", "in", "findOrInit", "getComUp",
}

var allOpPrefixes = func() []string {
	all := make([]string, 0, len(crudPrefixes)+len(servicePrefixes))
	all = append(all, crudPrefixes...)
	all = append(all, servicePrefixes...)
	return all
}()

type NamingOperationsRule struct{}

func (r *NamingOperationsRule) ID() string { return "NOM001" }
func (r *NamingOperationsRule) Description() string {
	return "Operations devem usar lowerCamelCase com prefixo semântico (CRUD ou verbo no infinitivo)"
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
				Suggestion: "Use lowerCamelCase: ex. 'postFixacaoMi', 'efetivarProcesso'",
			})
		}

		if !hasValidOpPrefix(op.Name) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Operation '%s' não inicia com prefixo semântico reconhecido", op.Name),
				Location:   fmt.Sprintf("operation %s", op.Name),
				Line:       op.LineStart,
				Suggestion: "CRUD: post/get/find/put/del/cancel | Serviço: efetivar/cancelar/calcular/buscar/gerar/validar/processar (verbos no infinitivo)",
			})
		}
	}
	return issues
}

func hasValidOpPrefix(name string) bool {
	lower := strings.ToLower(name)
	for _, p := range allOpPrefixes {
		if strings.HasPrefix(lower, p) {
			return true
		}
	}
	return false
}

// ============================================================
// RULE: NOM002 - Nomenclatura de Entries (pl + UpperCamelCase)
// ============================================================

type NamingEntriesRule struct{}

func (r *NamingEntriesRule) ID() string { return "NOM002" }
func (r *NamingEntriesRule) Description() string {
	return "Entries devem iniciar com 'pl' seguido de letra maiúscula"
}

func (r *NamingEntriesRule) Run(ctx *RuleContext) []Issue {
	entries := ctx.Entries.([]parser.ProcUnit)
	var issues []Issue

	for _, e := range entries {
		if !strings.HasPrefix(strings.ToLower(e.Name), "pl") {
			suggestion := e.Name
			if len(suggestion) > 0 {
				suggestion = fmt.Sprintf("pl%s%s", strings.ToUpper(string(suggestion[0])), suggestion[1:])
			}
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityError,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Entry '%s' deve iniciar com 'pl'", e.Name),
				Location:   fmt.Sprintf("entry %s", e.Name),
				Line:       e.LineStart,
				Suggestion: fmt.Sprintf("Renomeie para '%s'", suggestion),
			})
			continue
		}
		rest := e.Name[2:]
		if rest != "" && !startsWithUpper(rest) {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("Entry '%s': após 'pl' esperava-se letra maiúscula", e.Name),
				Location:   fmt.Sprintf("entry %s", e.Name),
				Line:       e.LineStart,
				Suggestion: "Padrão correto: plPostFixacaoMi, plGetClientePk",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: NOM003 - #define com UpperCamelCase ou código de mensagem
// ============================================================

type NamingDefinesRule struct{}

func (r *NamingDefinesRule) ID() string { return "NOM003" }
func (r *NamingDefinesRule) Description() string {
	return "#define deve usar UpperCamelCase ou código de mensagem (ESxxx/ENxxx/EWxxx)"
}

var (
	reDefineExtract = regexp.MustCompile(`(?m)^#define\s+(\S+)\s*=`)
	reMsgCode       = regexp.MustCompile(`^(ES|EN|EW)\d+$`)
)

var unifaceInternalDefines = map[string]bool{
	"$triggerAbbr":                true,
	"BreakInheritance_cpt_GENERAL": true,
	"BreakInheritance_cpt_INIT":    true,
}

func (r *NamingDefinesRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	seen := make(map[string]bool)
	matches := reDefineExtract.FindAllStringSubmatch(ctx.Script, -1)

	for _, m := range matches {
		name := m[1]
		if unifaceInternalDefines[name] || strings.HasPrefix(name, "$") || seen[name] {
			continue
		}
		seen[name] = true

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
// RULE: NOM004 - Parâmetros com prefixo 'p', variáveis com prefixo 'v'
// ============================================================

type NamingParamsVarsRule struct{}

func (r *NamingParamsVarsRule) ID() string { return "NOM004" }
func (r *NamingParamsVarsRule) Description() string {
	return "Parâmetros devem iniciar com 'p' e variáveis locais com 'v'"
}

var reVarDecl = regexp.MustCompile(`(?i)^\s*(?:numeric|string|boolean|struct|date|datetime|float|integer)\s+(.+)`)

func extractVarNames(body string) []string {
	var names []string
	inVars := false

	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, "variables") {
			inVars = true
			continue
		}
		if strings.EqualFold(trimmed, "endvariables") {
			inVars = false
			continue
		}
		if !inVars {
			continue
		}
		m := reVarDecl.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		// m[1] pode ser lista separada por vírgula: "vDsMetodo, vDsMsg"
		for _, raw := range strings.Split(m[1], ",") {
			fields := strings.Fields(strings.TrimSpace(raw))
			if len(fields) > 0 && fields[0] != "" {
				names = append(names, fields[0])
			}
		}
	}
	return names
}

func (r *NamingParamsVarsRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		// Verificar parâmetros
		for _, p := range unit.Params {
			name := p.Name
			if strings.Contains(name, "$") {
				continue // ignora $t_ds_erro$ e similares
			}
			if !strings.HasPrefix(name, "p") {
				suggestion := ""
				if len(name) > 0 {
					suggestion = fmt.Sprintf("Renomeie para 'p%s%s'", strings.ToUpper(string(name[0])), name[1:])
				}
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityWarning,
					Category:   "Nomenclatura",
					Message:    fmt.Sprintf("Parâmetro '%s' em %s '%s' deve iniciar com 'p'", name, unit.Kind, unit.Name),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart,
					Suggestion: suggestion,
				})
			}
		}

		// Verificar variáveis locais
		for _, varName := range extractVarNames(unit.Body) {
			if len(varName) == 0 {
				continue
			}
			if !strings.HasPrefix(varName, "v") {
				suggestion := fmt.Sprintf("Renomeie para 'v%s%s'", strings.ToUpper(string(varName[0])), varName[1:])
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityWarning,
					Category:   "Nomenclatura",
					Message:    fmt.Sprintf("Variável '%s' em %s '%s' deve iniciar com 'v'", varName, unit.Kind, unit.Name),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart,
					Suggestion: suggestion,
				})
			}
		}
	}
	return issues
}

// ============================================================
// RULE: NOM005 - Ordem dos parâmetros
// ============================================================

type ParamOrderRule struct{}

func (r *ParamOrderRule) ID() string { return "NOM005" }
func (r *ParamOrderRule) Description() string {
	return "Ordem dos parâmetros: pCdOperador 1º (em post/put/del), pStResult antes de $t_ds_erro$, $t_ds_erro$ sempre último"
}

// procRequiresCdOperador verifica se a PROC (operation ou entry) é do tipo
// post/put/del e portanto exige pCdOperador como primeiro parâmetro.
func procRequiresCdOperador(unit parser.ProcUnit) bool {
	name := unit.Name
	// Para entries, descarta o prefixo "pl"
	if unit.Kind == "entry" && len(name) > 2 && strings.HasPrefix(strings.ToLower(name), "pl") {
		name = name[2:]
	}
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, "post") ||
		strings.HasPrefix(lower, "put") ||
		strings.HasPrefix(lower, "del")
}

func (r *ParamOrderRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		params := unit.Params
		if len(params) == 0 {
			continue
		}

		// Regra 1: post/put/del devem ter pCdOperador como primeiro parâmetro
		if procRequiresCdOperador(unit) {
			if !strings.EqualFold(params[0].Name, "pCdOperador") {
				issues = append(issues, Issue{
					RuleID:     r.ID(),
					Severity:   SeverityError,
					Category:   "Nomenclatura",
					Message:    fmt.Sprintf("%s '%s': primeiro parâmetro deve ser 'pCdOperador' (encontrado: '%s')", unit.Kind, unit.Name, params[0].Name),
					Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
					Line:       unit.LineStart,
					Suggestion: "Operações post/put/del devem ter 'pCdOperador' como primeiro parâmetro",
				})
			}
		}

		// Localizar índices de $t_ds_erro$ e pResult
		errIdx := -1
		resultIdx := -1
		for i, p := range params {
			lower := strings.ToLower(p.Name)
			if strings.Contains(lower, "t_ds_erro") || strings.Contains(lower, "ds_erro") {
				errIdx = i
			}
			if strings.Contains(lower, "result") && p.Direction == "out" {
				resultIdx = i
			}
		}

		// Regra 2: $t_ds_erro$ deve ser o último
		if errIdx >= 0 && errIdx != len(params)-1 {
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   SeverityWarning,
				Category:   "Nomenclatura",
				Message:    fmt.Sprintf("%s '%s': '$t_ds_erro$' deve ser o último parâmetro (posição %d de %d)", unit.Kind, unit.Name, errIdx+1, len(params)),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Ordem: ..., pStResult : out, $t_ds_erro$ : out",
			})
		}

		// Regra 3: pStResult deve vir antes de $t_ds_erro$
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
