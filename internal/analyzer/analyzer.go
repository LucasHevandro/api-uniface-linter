package analyzer

import (
	"uniface-linter/internal/parser"
	"uniface-linter/internal/rules"
)

// Config configura o analyzer
type Config struct {
	MaxProcLines  int
	DisabledRules []string
}

// Result é o resultado da análise de um componente
type Result struct {
	ComponentName string        `json:"component_name"`
	ComponentType string        `json:"component_type"`
	Description   string        `json:"component_description"`
	FilePath      string        `json:"file_path"`
	TotalIssues   int           `json:"total_issues"`
	ErrorCount    int           `json:"error_count"`
	WarningCount  int           `json:"warning_count"`
	InfoCount     int           `json:"info_count"`
	Issues        []rules.Issue `json:"issues"`
	Summary       Summary       `json:"summary"`
}

// Summary agrega os dados por categoria
type Summary struct {
	Nomenclatura     int `json:"nomenclatura"`
	Complexidade     int `json:"complexidade"`
	Documentacao     int `json:"documentacao"`
	TratamentoErros  int `json:"tratamento_erros"`
	VariaveisGlobais int `json:"variaveis_globais"`
}

// Analyze analisa um componente Uniface com base nas convenções COMLOG
func Analyze(comp *parser.Component, filePath string, cfg Config) *Result {
	// Montar contexto
	ctx := &rules.RuleContext{
		ComponentName: comp.Name,
		ComponentType: comp.Type,
		Description:   comp.Description,
		Comment:       comp.Comment,
		Declarations:  comp.Declarations,
		Script:        comp.Script,
		Operations:    comp.Operations,
		Entries:       comp.Entries,
	}

	// Registrar todas as regras
	allRules := []rules.Rule{
		// Nomenclatura
		&rules.NamingOperationsRule{},
		&rules.NamingEntriesRule{},
		&rules.NamingDefinesRule{},
		&rules.NamingParamsVarsRule{},
		&rules.ParamOrderRule{},
		// Complexidade
		&rules.ProcSizeRule{MaxLines: cfg.MaxProcLines},
		&rules.OperationDelegatesRule{},
		&rules.ParamCountRule{},
		// Tratamento de erros
		&rules.ErrorHandlingRule{},
		// Documentação
		&rules.ProcHeaderRule{},
		&rules.ComponentHeaderRule{},
	}

	// Filtrar regras desabilitadas
	disabled := make(map[string]bool)
	for _, id := range cfg.DisabledRules {
		disabled[id] = true
	}

	var allIssues []rules.Issue
	for _, rule := range allRules {
		if disabled[rule.ID()] {
			continue
		}
		issues := rule.Run(ctx)
		allIssues = append(allIssues, issues...)
	}

	// Montar resultado
	result := &Result{
		ComponentName: comp.Name,
		ComponentType: comp.Type,
		Description:   comp.Description,
		FilePath:      filePath,
		Issues:        allIssues,
	}

	for _, issue := range allIssues {
		result.TotalIssues++
		switch issue.Severity {
		case rules.SeverityError:
			result.ErrorCount++
		case rules.SeverityWarning:
			result.WarningCount++
		case rules.SeverityInfo:
			result.InfoCount++
		}

		switch issue.Category {
		case "Nomenclatura":
			result.Summary.Nomenclatura++
		case "Complexidade":
			result.Summary.Complexidade++
		case "Documentação":
			result.Summary.Documentacao++
		case "Tratamento de Erros":
			result.Summary.TratamentoErros++
		case "Variáveis Globais":
			result.Summary.VariaveisGlobais++
		}
	}

	return result
}
