package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config representa o arquivo de configuração do linter
type Config struct {
	// Threshold de linhas para complexidade (padrão: 100)
	MaxProcLines int `json:"max_proc_lines"`

	// Regras desabilitadas por ID (ex: ["GLB002", "DOC003"])
	DisabledRules []string `json:"disabled_rules"`

	// Onde salvar o relatório JSON (padrão: linter-report.json)
	OutputFile string `json:"output_file"`

	// Se true, imprime resumo no terminal também
	PrintSummary bool `json:"print_summary"`

	// Severidade mínima para retornar exit code 1 ("ERROR", "WARNING", "INFO")
	FailOn string `json:"fail_on"`
}

var defaults = Config{
	MaxProcLines: 100,
	OutputFile:   "linter-report.json",
	PrintSummary: true,
	FailOn:       "ERROR",
}

// Load carrega configuração do arquivo. Se o arquivo não existir, retorna os defaults.
func Load(path string) (*Config, error) {
	cfg := defaults

	if path == "" {
		path = ".uniface-linter.json"
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("erro ao ler config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("erro ao parsear config: %w", err)
	}

	return &cfg, nil
}

// Save grava a configuração atual em arquivo
func Save(cfg *Config, path string) error {
	if path == "" {
		path = ".uniface-linter.json"
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// DefaultConfigJSON retorna o JSON de configuração padrão como string (para o comando init)
func DefaultConfigJSON() string {
	d, _ := json.MarshalIndent(defaults, "", "  ")
	return string(d)
}
