package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"uniface-linter/internal/analyzer"
	"uniface-linter/internal/parser"
	"uniface-linter/internal/rules"
)

const maxUploadSize = 10 * 1024 * 1024 // 10 MB

// Server encapsula o roteador e configurações da API
type Server struct {
	mux     *http.ServeMux
	version string
}

// New cria um novo servidor da API
func New(version string) *Server {
	s := &Server{
		mux:     http.NewServeMux(),
		version: version,
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /", s.handleRoot)
	s.mux.HandleFunc("POST /analyze", s.handleAnalyze)
	s.mux.HandleFunc("POST /analyze/batch", s.handleAnalyzeBatch)
	s.mux.HandleFunc("GET /rules", s.handleRules)
}

// ─── Respostas padrão ────────────────────────────────────────

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	Status    string `json:"status"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

type rootResponse struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Endpoints   map[string]string `json:"endpoints"`
}

type ruleInfo struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Description string `json:"description"`
}

// ─── Handlers ────────────────────────────────────────────────

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, rootResponse{
		Name:        "uniface-linter API",
		Version:     s.version,
		Description: "Analisador estático de componentes Uniface (convenções COMLOG)",
		Endpoints: map[string]string{
			"GET  /health":        "Status da API",
			"GET  /rules":         "Lista todas as regras disponíveis",
			"POST /analyze":       "Analisa um único arquivo XML (multipart: field 'file')",
			"POST /analyze/batch": "Analisa múltiplos arquivos XML (multipart: field 'files')",
		},
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "ok",
		Version:   s.version,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleRules(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, allRules())
}

// handleAnalyze recebe um único arquivo XML via multipart/form-data (field "file")
// ou via body direto com Content-Type application/xml
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	cfg := configFromQuery(r)

	var xmlData []byte
	var filename string
	var err error

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		xmlData, filename, err = readMultipartSingle(r, "file")
	} else {
		// Aceita body direto como XML
		xmlData, err = io.ReadAll(io.LimitReader(r.Body, maxUploadSize))
		filename = r.Header.Get("X-Filename")
		if filename == "" {
			filename = "componente.xml"
		}
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("erro ao ler arquivo: %v", err))
		return
	}
	if len(xmlData) == 0 {
		writeError(w, http.StatusBadRequest, "arquivo XML vazio ou não enviado (field: 'file')")
		return
	}

	result, err := analyzeXML(xmlData, filename, cfg)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// handleAnalyzeBatch recebe múltiplos XMLs via multipart/form-data (field "files")
func (s *Server) handleAnalyzeBatch(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "erro ao parsear multipart: "+err.Error())
		return
	}

	cfg := configFromQuery(r)
	files := r.MultipartForm.File["files"]

	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "nenhum arquivo enviado (field: 'files')")
		return
	}

	type batchReport struct {
		GeneratedAt  string             `json:"generated_at"`
		TotalFiles   int                `json:"total_files"`
		TotalIssues  int                `json:"total_issues"`
		ErrorCount   int                `json:"error_count"`
		WarningCount int                `json:"warning_count"`
		InfoCount    int                `json:"info_count"`
		Components   []*analyzer.Result `json:"components"`
		Errors       []string           `json:"errors,omitempty"`
	}

	report := batchReport{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		TotalFiles:  len(files),
	}

	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", fh.Filename, err))
			continue
		}
		data, err := io.ReadAll(io.LimitReader(f, maxUploadSize))
		f.Close()
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", fh.Filename, err))
			continue
		}

		result, err := analyzeXML(data, fh.Filename, cfg)
		if err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("%s: %v", fh.Filename, err))
			continue
		}

		report.Components = append(report.Components, result)
		report.TotalIssues += result.TotalIssues
		report.ErrorCount += result.ErrorCount
		report.WarningCount += result.WarningCount
		report.InfoCount += result.InfoCount
	}

	writeJSON(w, http.StatusOK, report)
}

// ─── Helpers ─────────────────────────────────────────────────

func analyzeXML(data []byte, filename string, cfg analyzer.Config) (*analyzer.Result, error) {
	// Gravar em arquivo temporário (o parser precisa de path)
	tmp, err := os.CreateTemp("", "uniface-*.xml")
	if err != nil {
		return nil, fmt.Errorf("erro interno ao criar arquivo temporário")
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("erro interno ao escrever arquivo temporário")
	}
	tmp.Close()

	comp, err := parser.Parse(tmp.Name())
	if err != nil {
		return nil, fmt.Errorf("falha ao parsear XML '%s': %v", filename, err)
	}

	result := analyzer.Analyze(comp, filename, cfg)
	return result, nil
}

func readMultipartSingle(r *http.Request, field string) ([]byte, string, error) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return nil, "", err
	}
	f, fh, err := r.FormFile(field)
	if err != nil {
		return nil, "", fmt.Errorf("field '%s' não encontrado", field)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxUploadSize))
	return data, fh.Filename, err
}

// configFromQuery lê parâmetros de configuração da query string
// Ex: POST /analyze?max_lines=150&disable=GLB002
func configFromQuery(r *http.Request) analyzer.Config {
	cfg := analyzer.Config{MaxProcLines: 100}

	if v := r.URL.Query().Get("max_lines"); v != "" {
		var n int
		fmt.Sscanf(v, "%d", &n)
		if n > 0 {
			cfg.MaxProcLines = n
		}
	}

	if v := r.URL.Query().Get("disable"); v != "" {
		for _, id := range strings.Split(v, ",") {
			id = strings.TrimSpace(strings.ToUpper(id))
			if id != "" {
				cfg.DisabledRules = append(cfg.DisabledRules, id)
			}
		}
	}

	return cfg
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func allRules() []ruleInfo {
	return []ruleInfo{
		{"ERR001", "Tratamento de Erros", "Todo activate ou call deve ser seguido de #include lib_coamo:g_vld_erro"},
	}
}

// Garantir que o import de rules é usado (para severidades nos testes de integração)
var _ = rules.SeverityError
