package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"uniface-linter/internal/analyzer"
	"uniface-linter/internal/rules"
)

// Report é o documento JSON final com todos os resultados
type Report struct {
	GeneratedAt string             `json:"generated_at"`
	TotalFiles  int                `json:"total_files"`
	TotalIssues int                `json:"total_issues"`
	ErrorCount  int                `json:"error_count"`
	WarningCount int               `json:"warning_count"`
	InfoCount   int                `json:"info_count"`
	Components  []*analyzer.Result `json:"components"`
}

// WriteJSON gera o relatório em JSON no writer fornecido
func WriteJSON(results []*analyzer.Result, w io.Writer) error {
	report := buildReport(results)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// WriteJSONFile grava o relatório em um arquivo
func WriteJSONFile(results []*analyzer.Result, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de relatório: %w", err)
	}
	defer f.Close()
	return WriteJSON(results, f)
}

// PrintSummary imprime um resumo legível no terminal
func PrintSummary(results []*analyzer.Result) {
	report := buildReport(results)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("  UNIFACE LINTER - RELATÓRIO DE ANÁLISE")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  Gerado em : %s\n", report.GeneratedAt)
	fmt.Printf("  Arquivos  : %d\n", report.TotalFiles)
	fmt.Printf("  Problemas : %d total  (%d ERROR  %d WARNING  %d INFO)\n",
		report.TotalIssues, report.ErrorCount, report.WarningCount, report.InfoCount)
	fmt.Println(strings.Repeat("-", 60))

	for _, comp := range report.Components {
		if comp.TotalIssues == 0 {
			fmt.Printf("\n  ✓ %s (%s) — sem problemas\n", comp.ComponentName, comp.ComponentType)
			continue
		}

		fmt.Printf("\n  ● %s (%s) — %d problema(s)\n", comp.ComponentName, comp.ComponentType, comp.TotalIssues)
		fmt.Printf("    %s\n", comp.Description)

		// Agrupar por categoria
		byCat := make(map[string][]rules.Issue)
		for _, issue := range comp.Issues {
			byCat[issue.Category] = append(byCat[issue.Category], issue)
		}

		for cat, issues := range byCat {
			fmt.Printf("\n    [%s]\n", cat)
			for _, issue := range issues {
				icon := severityIcon(issue.Severity)
				loc := issue.Location
				if issue.Line > 0 {
					loc = fmt.Sprintf("%s (L%d)", loc, issue.Line)
				}
				fmt.Printf("      %s [%s] %s\n", icon, issue.RuleID, issue.Message)
				fmt.Printf("          → %s\n", loc)
				if issue.Suggestion != "" {
					firstLine := strings.Split(issue.Suggestion, "\n")[0]
					fmt.Printf("          💡 %s\n", firstLine)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))

	if report.ErrorCount > 0 {
		fmt.Printf("  ✗ Análise concluída com %d ERRO(S)\n", report.ErrorCount)
	} else if report.WarningCount > 0 {
		fmt.Printf("  ⚠ Análise concluída com %d AVISO(S)\n", report.WarningCount)
	} else {
		fmt.Println("  ✓ Análise concluída sem erros críticos")
	}
	fmt.Println(strings.Repeat("=", 60))
}

func buildReport(results []*analyzer.Result) Report {
	report := Report{
		GeneratedAt: time.Now().Format("2006-01-02T15:04:05"),
		TotalFiles:  len(results),
		Components:  results,
	}
	for _, r := range results {
		report.TotalIssues += r.TotalIssues
		report.ErrorCount += r.ErrorCount
		report.WarningCount += r.WarningCount
		report.InfoCount += r.InfoCount
	}
	return report
}

func severityIcon(s rules.Severity) string {
	switch s {
	case rules.SeverityError:
		return "✗"
	case rules.SeverityWarning:
		return "⚠"
	default:
		return "ℹ"
	}
}
