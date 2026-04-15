package rules

import (
	"fmt"
	"regexp"
	"strings"

	"uniface-linter/internal/parser"
)

// ============================================================
// RULE: DOC001 - Cabeçalho obrigatório nas PROCs
// ============================================================

type ProcHeaderRule struct{}

func (r *ProcHeaderRule) ID() string          { return "DOC001" }
func (r *ProcHeaderRule) Description() string  { return "PROCs devem ter cabeçalho com Descrição, Autor, Criação e Projeto" }

var (
	reHasDescricao = regexp.MustCompile(`(?i);.*descri`)
	reHasAutor     = regexp.MustCompile(`(?i);.*autor`)
	reHasCriacao   = regexp.MustCompile(`(?i);.*(cria|data)`)
	reHasProjeto   = regexp.MustCompile(`(?i);.*(projeto|item|comlog)`)
)

func (r *ProcHeaderRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	all := append(ctx.Operations.([]parser.ProcUnit), ctx.Entries.([]parser.ProcUnit)...)

	for _, unit := range all {
		hdr := unit.HeaderComment
		missing := []string{}

		if !reHasDescricao.MatchString(hdr) {
			missing = append(missing, "Descrição")
		}
		if !reHasAutor.MatchString(hdr) {
			missing = append(missing, "Autor")
		}
		if !reHasCriacao.MatchString(hdr) {
			missing = append(missing, "Criação/Data")
		}
		if !reHasProjeto.MatchString(hdr) {
			missing = append(missing, "Projeto/Item")
		}

		if len(missing) > 0 {
			sev := SeverityWarning
			if len(missing) >= 3 {
				sev = SeverityError
			}
			issues = append(issues, Issue{
				RuleID:     r.ID(),
				Severity:   sev,
				Category:   "Documentação",
				Message:    fmt.Sprintf("%s '%s' faltam no cabeçalho: %s", unit.Kind, unit.Name, strings.Join(missing, ", ")),
				Location:   fmt.Sprintf("%s %s", unit.Kind, unit.Name),
				Line:       unit.LineStart,
				Suggestion: "Adicione cabeçalho padrão:\n;|\n;Descrição: ...\n;Autor: ...\n;Criação: dd/mm/aaaa\n;Projeto: COMLOG-XXX",
			})
		}
	}
	return issues
}

// ============================================================
// RULE: DOC002 - Cabeçalho do componente
// ============================================================

type ComponentHeaderRule struct{}

func (r *ComponentHeaderRule) ID() string          { return "DOC002" }
func (r *ComponentHeaderRule) Description() string  { return "Componente deve ter cabeçalho documentado (Autor, Data, Função)" }

func (r *ComponentHeaderRule) Run(ctx *RuleContext) []Issue {
	var issues []Issue
	comment := ctx.Comment

	if strings.TrimSpace(comment) == "" {
		issues = append(issues, Issue{
			RuleID:     r.ID(),
			Severity:   SeverityError,
			Category:   "Documentação",
			Message:    "Componente não possui cabeçalho de documentação",
			Location:   fmt.Sprintf("componente %s", ctx.ComponentName),
			Suggestion: "Adicione comentário de cabeçalho com Autor, Data e Função do componente",
		})
		return issues
	}

	if !regexp.MustCompile(`(?i)autor`).MatchString(comment) {
		issues = append(issues, Issue{
			RuleID:     r.ID(),
			Severity:   SeverityWarning,
			Category:   "Documentação",
			Message:    "Cabeçalho do componente não indica o Autor",
			Location:   fmt.Sprintf("componente %s", ctx.ComponentName),
		})
	}

	if !regexp.MustCompile(`(?i)(data|cria)`).MatchString(comment) {
		issues = append(issues, Issue{
			RuleID:     r.ID(),
			Severity:   SeverityWarning,
			Category:   "Documentação",
			Message:    "Cabeçalho do componente não indica a Data de criação",
			Location:   fmt.Sprintf("componente %s", ctx.ComponentName),
		})
	}

	return issues
}

// ============================================================
// RULE: DOC003 - Comentário em modificações
// ============================================================

type ModificationCommentRule struct{}

func (r *ModificationCommentRule) ID() string          { return "DOC003" }
func (r *ModificationCommentRule) Description() string  { return "Modificações devem ter comentário com data, autor e item" }

// Detecta linhas que parecem modificações sem comentário próximo
// Heurística: bloco de código seguido de bloco comentado (o antigo) sem header de modificação
var reModifComment = regexp.MustCompile(`(?i);.*\d{2}/\d{2}/\d{4}.*comlog`)

func (r *ModificationCommentRule) Run(ctx *RuleContext) []Issue {
	// Verificamos se o script como um todo tem algum comentário de modificação
	// Componentes com mais de 100 linhas e sem nenhum comentário de modificação são suspeitos
	lines := strings.Split(ctx.Script, "\n")
	if len(lines) < 100 {
		return nil
	}

	if !reModifComment.MatchString(ctx.Script) {
		return []Issue{{
			RuleID:     r.ID(),
			Severity:   SeverityInfo,
			Category:   "Documentação",
			Message:    "Componente não possui comentários de modificação com referência ao item (ex: ;mpereira 06/07/2021 comlog-397)",
			Location:   fmt.Sprintf("componente %s", ctx.ComponentName),
			Suggestion: "Comente modificações com: ;autor dd/mm/aaaa comlog-XXX",
		}}
	}
	return nil
}
