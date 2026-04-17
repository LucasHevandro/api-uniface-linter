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

// PartialResult é o resultado de uma análise parcial
// (proc específica pelo nome ou código avulso sem XML)
type PartialResult struct {
	Mode          string        `json:"mode"` // "proc_filter" | "raw_code"
	ComponentName string        `json:"component_name"`
	Filter        string        `json:"filter,omitempty"` // preenchido no modo proc_filter
	ProcsAnalyzed []string      `json:"procs_analyzed"`   // nomes das procs analisadas
	TotalProcs    int           `json:"total_procs"`
	TotalIssues   int           `json:"total_issues"`
	ErrorCount    int           `json:"error_count"`
	WarningCount  int           `json:"warning_count"`
	InfoCount     int           `json:"info_count"`
	Issues        []rules.Issue `json:"issues"`
	Summary       Summary       `json:"summary"`
}

// AnalyzeProc analisa apenas as ProcUnits cujo nome contém o filtro.
// Recebe o componente já parseado — ideal para code review de uma proc alterada.
func AnalyzeProc(comp *parser.Component, filter string, cfg Config) *PartialResult {
	filtered := parser.FilterByName(comp, filter)

	result := &PartialResult{
		Mode:          "proc_filter",
		ComponentName: comp.Name,
		Filter:        filter,
	}

	// Coletar nomes das procs que serão analisadas
	for _, op := range filtered.Operations {
		result.ProcsAnalyzed = append(result.ProcsAnalyzed, "operation "+op.Name)
	}
	for _, e := range filtered.Entries {
		result.ProcsAnalyzed = append(result.ProcsAnalyzed, "entry "+e.Name)
	}
	result.TotalProcs = len(result.ProcsAnalyzed)

	if result.TotalProcs == 0 {
		return result // nenhuma proc encontrada com esse nome
	}

	// Reusar o Analyze normal, mas com o componente filtrado
	full := Analyze(filtered, comp.Name, cfg)
	result.Issues = full.Issues
	result.TotalIssues = full.TotalIssues
	result.ErrorCount = full.ErrorCount
	result.WarningCount = full.WarningCount
	result.InfoCount = full.InfoCount
	result.Summary = full.Summary

	return result
}

// AnalyzeRawCode analisa código Uniface avulso colado diretamente,
// sem precisar do XML completo do componente.
func AnalyzeRawCode(code string, componentName string, cfg Config) *PartialResult {
	if componentName == "" {
		componentName = "componente-avulso"
	}

	comp := parser.ParseRawCode(code, componentName)

	result := &PartialResult{
		Mode:          "raw_code",
		ComponentName: componentName,
	}

	for _, op := range comp.Operations {
		result.ProcsAnalyzed = append(result.ProcsAnalyzed, "operation "+op.Name)
	}
	for _, e := range comp.Entries {
		result.ProcsAnalyzed = append(result.ProcsAnalyzed, "entry "+e.Name)
	}
	result.TotalProcs = len(result.ProcsAnalyzed)

	if result.TotalProcs == 0 {
		return result
	}

	full := Analyze(comp, componentName, cfg)
	result.Issues = full.Issues
	result.TotalIssues = full.TotalIssues
	result.ErrorCount = full.ErrorCount
	result.WarningCount = full.WarningCount
	result.InfoCount = full.InfoCount
	result.Summary = full.Summary

	return result
}
