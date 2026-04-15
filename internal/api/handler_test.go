package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"uniface-linter/internal/api"
)

func newServer() http.Handler {
	return api.Chain(
		api.New("test"),
		api.Logger,
		api.CORS,
	)
}

// ─── /health ─────────────────────────────────────────────────

func TestHealth(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("resposta inválida: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status esperado 'ok', got %q", resp["status"])
	}
	if resp["version"] != "test" {
		t.Errorf("version esperada 'test', got %q", resp["version"])
	}
}

// ─── / (root) ────────────────────────────────────────────────

func TestRoot(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if _, ok := resp["endpoints"]; !ok {
		t.Error("resposta deveria conter 'endpoints'")
	}
}

// ─── /rules ──────────────────────────────────────────────────

func TestRules(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("GET", "/rules", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rr.Code)
	}

	var rules []map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&rules); err != nil {
		t.Fatalf("resposta inválida: %v", err)
	}
	if len(rules) < 10 {
		t.Errorf("esperava ao menos 10 regras, got %d", len(rules))
	}
	// Verificar campos obrigatórios
	for _, r := range rules {
		if r["id"] == "" || r["category"] == "" || r["description"] == "" {
			t.Errorf("regra com campos faltando: %+v", r)
		}
	}
}

// ─── /analyze ────────────────────────────────────────────────

const minimalXML = `<?xml version='1.0' encoding='UTF-8' ?>
<UNIFACE release="10.4" repversion="8" xmlengine="2.0">
<TABLE>
<DSC name="UFORM" model="DICT" system="S" pseudo="73" level="1" noupdate="0"
 rbk="0" ffsql="0" transnr="0" segsize="0" ufocc="0" charset=".U">
<FLD name="ULABEL" seqno="2" type="S" level="2" />
<FLD name="UDESCR" seqno="7" type="S" level="2" />
<FLD name="UCOMMENT" seqno="35" type="S" level="2" />
<FLD name="UDECLARATIONS" seqno="38" type="S" level="2" />
<FLD name="USCRIPT" seqno="39" type="S" level="2" />
</DSC>
<OCC>
<DAT name="ULABEL">CCNO142</DAT>
<DAT name="UDESCR">AGG CRUD CN</DAT>
<DAT name="UCOMMENT">Autor: mpereira
Data: 01/01/2024</DAT>
<DAT name="UDECLARATIONS">variables
  numeric T_CD_OPERADOR
endvariables</DAT>
<DAT name="USCRIPT">;|
;Descricao: Cria fixacao
;Autor: mpereira
;Criacao: 01/01/2024
;Projeto: COMLOG-241
operation postFixacaoMi
	params
		numeric pCdOperador : in
		struct pStResult : out
		$t_ds_erro$ : out
	endparams
	$t_ds_erro$ = ""
	call plPostFixacaoMi(pCdOperador, pStResult, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	return 0
end;

;|
;Descricao: PL de criacao
;Autor: mpereira
;Criacao: 01/01/2024
;Projeto: COMLOG-241
entry plPostFixacaoMi
	params
		numeric pCdOperador : in
		struct pStResult : out
		$t_ds_erro$ : out
	endparams
	$t_ds_erro$ = ""
	return 0
end;
</DAT>
</OCC>
</TABLE>
</UNIFACE>`

func makeMultipart(t *testing.T, field, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(fw, content)
	w.Close()
	return body, w.FormDataContentType()
}

func TestAnalyze_Multipart(t *testing.T) {
	srv := newServer()
	body, ct := makeMultipart(t, "file", "ccno142.xml", minimalXML)

	req := httptest.NewRequest("POST", "/analyze", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d\nbody: %s", rr.Code, rr.Body)
	}

	var result map[string]any
	if err := json.NewDecoder(rr.Body).Decode(&result); err != nil {
		t.Fatalf("resposta inválida: %v", err)
	}
	if result["component_name"] != "CCNO142" {
		t.Errorf("component_name errado: %v", result["component_name"])
	}
	if _, ok := result["issues"]; !ok {
		t.Error("resposta deveria conter 'issues'")
	}
	if _, ok := result["summary"]; !ok {
		t.Error("resposta deveria conter 'summary'")
	}
}

func TestAnalyze_DirectBody(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("POST", "/analyze", strings.NewReader(minimalXML))
	req.Header.Set("Content-Type", "application/xml")
	req.Header.Set("X-Filename", "ccno142.xml")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d\nbody: %s", rr.Code, rr.Body)
	}
}

func TestAnalyze_EmptyBody(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("POST", "/analyze", nil)
	req.Header.Set("Content-Type", "application/xml")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status esperado 400, got %d", rr.Code)
	}
}

func TestAnalyze_InvalidXML(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("POST", "/analyze", strings.NewReader("<invalido>"))
	req.Header.Set("Content-Type", "application/xml")
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status esperado 422, got %d", rr.Code)
	}

	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("resposta de erro deveria ter campo 'error'")
	}
}

func TestAnalyze_QueryConfig(t *testing.T) {
	srv := newServer()
	body, ct := makeMultipart(t, "file", "ccno142.xml", minimalXML)

	// max_lines=5 deve gerar COMP001
	req := httptest.NewRequest("POST", "/analyze?max_lines=5", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rr.Code)
	}

	var result map[string]any
	json.NewDecoder(rr.Body).Decode(&result)

	issues := result["issues"].([]any)
	foundCOMP001 := false
	for _, raw := range issues {
		issue := raw.(map[string]any)
		if issue["rule_id"] == "COMP001" {
			foundCOMP001 = true
			break
		}
	}
	if !foundCOMP001 {
		t.Error("com max_lines=5 esperava encontrar COMP001 nos issues")
	}
}

func TestAnalyze_DisableRule(t *testing.T) {
	srv := newServer()
	body, ct := makeMultipart(t, "file", "ccno142.xml", minimalXML)

	// disable=ERR001 — não deve aparecer no resultado
	req := httptest.NewRequest("POST", "/analyze?disable=ERR001", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d", rr.Code)
	}

	var result map[string]any
	json.NewDecoder(rr.Body).Decode(&result)

	// issues pode ser nil quando nenhum issue é encontrado
	if rawIssues, ok := result["issues"]; ok && rawIssues != nil {
		issues := rawIssues.([]any)
		for _, raw := range issues {
			issue := raw.(map[string]any)
			if issue["rule_id"] == "ERR001" {
				t.Error("ERR001 foi desabilitado mas ainda apareceu no resultado")
			}
		}
	}
}

// ─── /analyze/batch ──────────────────────────────────────────

func TestAnalyzeBatch(t *testing.T) {
	srv := newServer()

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for _, name := range []string{"comp1.xml", "comp2.xml"} {
		fw, _ := w.CreateFormFile("files", name)
		io.WriteString(fw, minimalXML)
	}
	w.Close()

	req := httptest.NewRequest("POST", "/analyze/batch", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status esperado 200, got %d\nbody: %s", rr.Code, rr.Body)
	}

	var result map[string]any
	json.NewDecoder(rr.Body).Decode(&result)

	if result["total_files"].(float64) != 2 {
		t.Errorf("total_files esperado 2, got %v", result["total_files"])
	}
	components := result["components"].([]any)
	if len(components) != 2 {
		t.Errorf("esperava 2 components, got %d", len(components))
	}
}

func TestAnalyzeBatch_NoFiles(t *testing.T) {
	srv := newServer()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	w.Close()

	req := httptest.NewRequest("POST", "/analyze/batch", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status esperado 400, got %d", rr.Code)
	}
}

// ─── CORS ────────────────────────────────────────────────────

func TestCORS_Headers(t *testing.T) {
	srv := newServer()
	req := httptest.NewRequest("OPTIONS", "/analyze", nil)
	rr := httptest.NewRecorder()
	srv.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("OPTIONS esperava 204, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header Access-Control-Allow-Origin ausente")
	}
}
