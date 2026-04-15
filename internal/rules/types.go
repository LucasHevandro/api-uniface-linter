package rules

import "fmt"

// Severity indica o nível de severidade de um problema
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

// Issue representa um problema encontrado
type Issue struct {
	RuleID      string   `json:"rule_id"`
	Severity    Severity `json:"severity"`
	Category    string   `json:"category"`
	Message     string   `json:"message"`
	Location    string   `json:"location"`    // ex: "operation postFixacaoMi"
	Line        int      `json:"line,omitempty"`
	Suggestion  string   `json:"suggestion,omitempty"`
}

func (i Issue) String() string {
	loc := i.Location
	if i.Line > 0 {
		loc = fmt.Sprintf("%s (linha %d)", loc, i.Line)
	}
	return fmt.Sprintf("[%s] %s - %s | %s", i.Severity, i.RuleID, loc, i.Message)
}

// Rule interface para todas as regras de validação
type Rule interface {
	ID() string
	Description() string
	Run(ctx *RuleContext) []Issue
}

// RuleContext passa o componente e metadados para as regras
type RuleContext struct {
	ComponentName string
	ComponentType string
	Description   string
	Comment       string
	Declarations  string
	Script        string
	Operations    interface{} // parser.ProcUnit slice
	Entries       interface{} // parser.ProcUnit slice
}
