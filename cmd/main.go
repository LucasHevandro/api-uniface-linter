package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"uniface-linter/internal/analyzer"
	"uniface-linter/internal/config"
	"uniface-linter/internal/parser"
	"uniface-linter/internal/reporter"
	"uniface-linter/internal/rules"
)

const version = "1.1.0"

const usage = `
uniface-linter v` + version + `
Analisador estático de componentes Uniface para convenções COMLOG

Uso:
  uniface-linter [opções] <arquivo.xml> [arquivo2.xml ...]
  uniface-linter [opções] <diretório/>

Opções:
  -c, --config <arquivo>   Arquivo de configuração (padrão: .uniface-linter.json)
  -o, --output <arquivo>   Arquivo de saída JSON (padrão: linter-report.json)
  -q, --quiet              Não imprime resumo no terminal
  --fail-on <nivel>        Retorna exit code 1 se houver issues deste nível (ERROR|WARNING|INFO)
  --init                   Gera arquivo de configuração padrão no diretório atual
  --rules                  Lista todas as regras disponíveis
  --version                Exibe a versão
  -h, --help               Exibe esta ajuda

Exemplos:
  uniface-linter cpt_cesto145.xml
  uniface-linter -o resultado.json *.xml
  uniface-linter --config meu-projeto.json ./componentes/
  uniface-linter --init
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Print(usage)
		os.Exit(0)
	}

	// Flags especiais sem arquivo
	switch args[0] {
	case "-h", "--help":
		fmt.Print(usage)
		return
	case "--version":
		fmt.Printf("uniface-linter v%s\n", version)
		return
	case "--init":
		runInit()
		return
	case "--rules":
		printRules()
		return
	}

	// Parsear flags
	var configPath, outputPath, failOn string
	var quiet bool
	var files []string
	var procFilter string
	var rawMode bool

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-c", "--config":
			i++
			if i < len(args) {
				configPath = args[i]
			}
		case "-o", "--output":
			i++
			if i < len(args) {
				outputPath = args[i]
			}
		case "-q", "--quiet":
			quiet = true
		case "--fail-on":
			i++
			if i < len(args) {
				failOn = strings.ToUpper(args[i])
			}
		case "--proc":
			i++
			if i < len(args) {
				procFilter = args[i]
			}
		case "--raw":
			rawMode = true
		default:
			// Pode ser arquivo ou diretório
			expanded, err := expandPaths(args[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao expandir path '%s': %v\n", args[i], err)
				os.Exit(1)
			}
			files = append(files, expanded...)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Erro: nenhum arquivo XML informado")
		fmt.Print(usage)
		os.Exit(1)
	}

	// Carregar configuração
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao carregar configuração: %v\n", err)
		os.Exit(1)
	}

	// Flags de linha de comando sobrescrevem config
	if outputPath != "" {
		cfg.OutputFile = outputPath
	}
	if quiet {
		cfg.PrintSummary = false
	}
	if failOn != "" {
		cfg.FailOn = failOn
	}

	// Analisar cada arquivo
	var results []*analyzer.Result

	for _, file := range files {
		fmt.Printf("Analisando: %s\n", file)

		analyzerCfg := analyzer.Config{
			MaxProcLines:  cfg.MaxProcLines,
			DisabledRules: cfg.DisabledRules,
		}

		if rawMode {
			data, err := os.ReadFile(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  ✗ Erro ao ler %s: %v\n", file, err)
				continue
			}
			result := analyzer.AnalyzeRawCode(string(data), filepath.Base(file), analyzerCfg)
			if result.TotalProcs == 0 {
				fmt.Printf("  → nenhuma proc encontrada em '%s'\n", file)
				continue
			}
			fmt.Printf("  → %d proc(s) analisada(s), %d problema(s)\n",
				result.TotalProcs, result.TotalIssues)
			results = append(results, partialToResult(result))
			continue
		}

		comp, err := parser.Parse(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ✗ Erro ao parsear %s: %v\n", file, err)
			continue
		}

		if procFilter != "" {
			result := analyzer.AnalyzeProc(comp, procFilter, analyzerCfg)
			if result.TotalProcs == 0 {
				fmt.Printf("  → nenhuma proc encontrada com o filtro '%s'\n", procFilter)
				continue
			}
			fmt.Printf("  → %d proc(s) analisada(s), %d problema(s)\n",
				result.TotalProcs, result.TotalIssues)
			results = append(results, partialToResult(result))
		} else {
			result := analyzer.Analyze(comp, file, analyzerCfg)
			results = append(results, result)
			fmt.Printf("  → %d problema(s) encontrado(s)\n", result.TotalIssues)
		}
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "Nenhum arquivo analisado com sucesso")
		os.Exit(1)
	}

	// Exibir resumo no terminal
	if cfg.PrintSummary {
		fmt.Println()
		reporter.PrintSummary(results)
	}

	// Gravar relatório JSON
	if err := reporter.WriteJSONFile(results, cfg.OutputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao gravar relatório: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nRelatório JSON salvo em: %s\n", cfg.OutputFile)

	// Exit code baseado no fail-on
	os.Exit(exitCode(results, cfg.FailOn))
}

func partialToResult(p *analyzer.PartialResult) *analyzer.Result {
	return &analyzer.Result{
		ComponentName: p.ComponentName + " [parcial: " + p.Filter + "]",
		ComponentType: "—",
		FilePath:      p.ComponentName,
		TotalIssues:   p.TotalIssues,
		ErrorCount:    p.ErrorCount,
		WarningCount:  p.WarningCount,
		InfoCount:     p.InfoCount,
		Issues:        p.Issues,
		Summary:       p.Summary,
	}
}

func expandPaths(path string) ([]string, error) {
	// Verificar se é glob
	if strings.Contains(path, "*") {
		return filepath.Glob(path)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Diretório: listar todos os .xml
	if info.IsDir() {
		var xmlFiles []string
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".xml") {
				xmlFiles = append(xmlFiles, filepath.Join(path, e.Name()))
			}
		}
		if len(xmlFiles) == 0 {
			return nil, fmt.Errorf("nenhum arquivo .xml encontrado em %s", path)
		}
		return xmlFiles, nil
	}

	return []string{path}, nil
}

func exitCode(results []*analyzer.Result, failOn string) int {
	for _, r := range results {
		switch failOn {
		case "ERROR":
			if r.ErrorCount > 0 {
				return 1
			}
		case "WARNING":
			if r.ErrorCount > 0 || r.WarningCount > 0 {
				return 1
			}
		case "INFO":
			if r.TotalIssues > 0 {
				return 1
			}
		}
	}
	return 0
}

func runInit() {
	path := ".uniface-linter.json"
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("Arquivo %s já existe. Sobrescrever? [s/N]: ", path)
		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(resp) != "s" {
			fmt.Println("Operação cancelada.")
			return
		}
	}

	cfg := config.Config{}
	if err := config.Save(&cfg, path); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Arquivo de configuração criado: %s\n", path)
	fmt.Println("\nConteúdo gerado:")
	fmt.Println(config.DefaultConfigJSON())
}

func printRules() {
	ruleList := []struct{ id, cat, desc string }{
		{"NOM001", "Nomenclatura", "Operations devem usar lowerCamelCase com prefixo semântico"},
		{"NOM002", "Nomenclatura", "Entries devem iniciar com 'pl' + UpperCamelCase"},
		{"NOM003", "Nomenclatura", "#define deve usar UpperCamelCase ou código de mensagem (ESxxx)"},
		{"NOM004", "Nomenclatura", "Parâmetros devem iniciar com 'p'"},
		{"COMP001", "Complexidade", "PROCs não devem exceder o limite de linhas de código"},
		{"COMP002", "Complexidade", "Operations devem delegar lógica para PLs (entries)"},
		{"COMP003", "Complexidade", "PROCs com muitos parâmetros devem usar struct"},
		{"DOC001", "Documentação", "PROCs devem ter cabeçalho (Descrição, Autor, Criação, Projeto)"},
		{"DOC002", "Documentação", "Componente deve ter cabeçalho documentado"},
		{"ERR001", "Tratamento de Erros", "activate/call devem ser seguidos de #include g_vld_erro"},
		{"ERR002", "Tratamento de Erros", "Evitar verificação manual de $status; usar includes"},
		{"ERR003", "Tratamento de Erros", "$t_ds_erro$ deve ser parâmetro 'out' e o último"},
		{"ERR004", "Tratamento de Erros", "Evitar números mágicos; usar constantes #define"},
		{"GLB001", "Variáveis Globais", "PLs devem evitar variáveis globais ($var$)"},
		{"GLB002", "Variáveis Globais", "PLs não devem atualizar variáveis de componente (T_xxx)"},
	}

	fmt.Printf("\n%-8s  %-20s  %s\n", "ID", "Categoria", "Descrição")
	fmt.Println(strings.Repeat("-", 80))
	for _, r := range ruleList {
		fmt.Printf("%-8s  %-20s  %s\n", r.id, r.cat, r.desc)
	}

	_ = rules.SeverityError // garantir import usado
	fmt.Println()
}
