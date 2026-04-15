package rules_test

import (
	"testing"

	"uniface-linter/internal/parser"
	"uniface-linter/internal/rules"
)

// helpers para construir contexto de teste
func makeCtx(ops []parser.ProcUnit, entries []parser.ProcUnit) *rules.RuleContext {
	return &rules.RuleContext{
		ComponentName: "TEST001",
		ComponentType: "SERVICE",
		Description:   "Componente de teste",
		Comment:       "Autor: teste\nData: 01/01/2024\nFunção: testes unitários",
		Script:        "",
		Operations:    ops,
		Entries:       entries,
	}
}

func op(name, body string, lineStart int, params ...parser.Param) parser.ProcUnit {
	return parser.ProcUnit{
		Name:          name,
		Kind:          "operation",
		Body:          body,
		LineStart:     lineStart,
		Params:        params,
		HeaderComment: ";|\n;Descrição: Teste\n;Autor: teste\n;Criação: 01/01/2024\n;Projeto: COMLOG-1",
	}
}

func entry(name, body string, lineStart int, params ...parser.Param) parser.ProcUnit {
	return parser.ProcUnit{
		Name:          name,
		Kind:          "entry",
		Body:          body,
		LineStart:     lineStart,
		Params:        params,
		HeaderComment: ";|\n;Descrição: Teste\n;Autor: teste\n;Criação: 01/01/2024\n;Projeto: COMLOG-1",
	}
}

func param(typ, name, dir string) parser.Param {
	return parser.Param{Type: typ, Name: name, Direction: dir}
}

// ============================================================
// NOM001 - Nomenclatura de Operations
// ============================================================

func TestNOM001_ValidOperation(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacaoMi", "operation postFixacaoMi\nparams\nendparams\nreturn 0\nend;", 1),
		op("getClientePk", "operation getClientePk\nparams\nendparams\nreturn 0\nend;", 10),
		op("findFixacoesMi", "operation findFixacoesMi\nparams\nendparams\nreturn 0\nend;", 20),
	}, nil)
	ctx.Entries = []parser.ProcUnit{}

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.Severity == rules.SeverityError {
			t.Errorf("Não esperava ERROR para operation válida: %s", i.Message)
		}
	}
}

func TestNOM001_InvalidCasing(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("PostFixacaoMi", "operation PostFixacaoMi\nreturn 0\nend;", 1),
		op("GET_CLIENTE", "operation GET_CLIENTE\nreturn 0\nend;", 10),
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	errorCount := 0
	for _, i := range issues {
		if i.RuleID == "NOM001" && i.Severity == rules.SeverityError {
			errorCount++
		}
	}
	if errorCount != 2 {
		t.Errorf("Esperava 2 ERRORs de NOM001, got %d", errorCount)
	}
}

func TestNOM001_MissingPrefix(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("gerarLoteFaturamento", "operation gerarLoteFaturamento\nreturn 0\nend;", 1),
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "NOM001" && i.Severity == rules.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Error("Esperava WARNING para operation sem prefixo semântico")
	}
}

// ============================================================
// NOM002 - Nomenclatura de Entries
// ============================================================

func TestNOM002_ValidEntry(t *testing.T) {
	r := &rules.NamingEntriesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		entry("plPostFixacaoMi", "entry plPostFixacaoMi\nreturn 0\nend;", 1),
		entry("plGetClientePk", "entry plGetClientePk\nreturn 0\nend;", 10),
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.Severity == rules.SeverityError {
			t.Errorf("Não esperava ERROR para entry válida: %s", i.Message)
		}
	}
}

func TestNOM002_MissingPLPrefix(t *testing.T) {
	r := &rules.NamingEntriesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		entry("getClientePk", "entry getClientePk\nreturn 0\nend;", 1),
		entry("postFixacao", "entry postFixacao\nreturn 0\nend;", 10),
	})

	issues := r.Run(ctx)
	errCount := 0
	for _, i := range issues {
		if i.RuleID == "NOM002" && i.Severity == rules.SeverityError {
			errCount++
		}
	}
	if errCount != 2 {
		t.Errorf("Esperava 2 ERRORs de NOM002, got %d", errCount)
	}
}

// ============================================================
// NOM004 - Nomenclatura de parâmetros
// ============================================================

func TestNOM004_ValidParams(t *testing.T) {
	r := &rules.NamingParamsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "operation postFixacao\nreturn 0\nend;", 1,
			param("numeric", "pCdOperador", "in"),
			param("struct", "pStFixacao", "in"),
			param("struct", "pStResult", "out"),
		),
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	for _, i := range issues {
		t.Errorf("Não esperava issues para params válidos: %s", i.Message)
	}
}

func TestNOM004_InvalidParams(t *testing.T) {
	r := &rules.NamingParamsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "operation postFixacao\nreturn 0\nend;", 1,
			param("numeric", "cdOperador", "in"), // sem 'p'
			param("struct", "stFixacao", "in"),    // sem 'p'
		),
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	warnCount := 0
	for _, i := range issues {
		if i.RuleID == "NOM004" {
			warnCount++
		}
	}
	if warnCount != 2 {
		t.Errorf("Esperava 2 warnings de NOM004, got %d", warnCount)
	}
}

// ============================================================
// COMP001 - Tamanho de PROCs
// ============================================================

func TestCOMP001_SmallProc(t *testing.T) {
	r := &rules.ProcSizeRule{MaxLines: 100}
	body := "operation postFixacao\n"
	for i := 0; i < 20; i++ {
		body += "  call plPostFixacao(pCdOperador, pStFixacao, pStResult, $t_ds_erro$)\n"
	}
	body += "return 0\nend;"

	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1},
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "COMP001" {
			t.Errorf("Não esperava issue de tamanho para PROC pequena: %s", i.Message)
		}
	}
}

func TestCOMP001_LargeProc(t *testing.T) {
	r := &rules.ProcSizeRule{MaxLines: 100}
	body := "operation procedimentoGigante\n"
	for i := 0; i < 150; i++ {
		body += "  someCode" + string(rune('a'+i%26)) + " = someValue\n"
	}
	body += "return 0\nend;"

	ctx := makeCtx([]parser.ProcUnit{
		{Name: "procedimentoGigante", Kind: "operation", Body: body, LineStart: 1},
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "COMP001" {
			found = true
		}
	}
	if !found {
		t.Error("Esperava issue COMP001 para PROC com 150 linhas")
	}
}

// ============================================================
// ERR001 - Tratamento de erros após activate/call
// ============================================================

func TestERR001_WithErrorCheck(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		numeric pCdOperador : in
		$t_ds_erro$ : out
	endparams
	call plValidate(pCdOperador, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	activate "serv001".postFixacao(pCdOperador, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			t.Errorf("Não esperava ERR001 quando #include está presente: %s", i.Message)
		}
	}
}

func TestERR001_MissingErrorCheck(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		numeric pCdOperador : in
		$t_ds_erro$ : out
	endparams
	call plValidate(pCdOperador, $t_ds_erro$)
	activate "serv001".postFixacao(pCdOperador, $t_ds_erro$)
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			found = true
		}
	}
	if !found {
		t.Error("Esperava ERR001 quando #include está ausente após activate/call")
	}
}

// ============================================================
// DOC001 - Cabeçalho de PROCs
// ============================================================

func TestDOC001_CompleteHeader(t *testing.T) {
	r := &rules.ProcHeaderRule{}
	unit := parser.ProcUnit{
		Name:          "postFixacao",
		Kind:          "operation",
		LineStart:     10,
		HeaderComment: ";|\n;Descrição: Cria fixação\n;Autor: mpereira\n;Criação: 01/01/2024\n;Projeto: COMLOG-241",
	}
	ctx := makeCtx([]parser.ProcUnit{unit}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "DOC001" && i.Severity == rules.SeverityError {
			t.Errorf("Não esperava ERROR de DOC001 com cabeçalho completo: %s", i.Message)
		}
	}
}

func TestDOC001_MissingHeader(t *testing.T) {
	r := &rules.ProcHeaderRule{}
	unit := parser.ProcUnit{
		Name:          "postFixacao",
		Kind:          "operation",
		LineStart:     10,
		HeaderComment: "", // sem cabeçalho
	}
	ctx := makeCtx([]parser.ProcUnit{unit}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "DOC001" {
			found = true
		}
	}
	if !found {
		t.Error("Esperava DOC001 para PROC sem cabeçalho")
	}
}

// ============================================================
// NOM005 - Ordem de parâmetros
// ============================================================

func TestNOM005_CorrectOrder(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "body", 1,
			param("numeric", "pCdOperador", "in"),
			param("struct", "pStFixacao", "in"),
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityWarning {
			t.Errorf("Não esperava WARNING de NOM005 com ordem correta: %s", i.Message)
		}
	}
}

func TestNOM005_ErroroNotLast(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "body", 1,
			param("numeric", "pCdOperador", "in"),
			param("string", "$t_ds_erro$", "out"), // erro: deve ser o último
			param("struct", "pStResult", "out"),   // errado: está depois do erro
		),
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Error("Esperava WARNING NOM005 quando $t_ds_erro$ não é o último")
	}
}

// ============================================================
// COMP002 - Operation deve delegar para PL
// ============================================================

func TestCOMP002_DelegatesCorrectly(t *testing.T) {
	r := &rules.OperationDelegatesRule{}
	body := "operation postFixacao\n"
	for i := 0; i < 15; i++ {
		body += "  ; comentario\n"
	}
	body += "  call plPostFixacao(pCdOperador, $t_ds_erro$)\n"
	body += "  #include lib_coamo:g_vld_erro\n"
	body += "  return 0\nend;"

	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1},
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "COMP002" {
			t.Errorf("Não esperava COMP002 quando operation delega para PL: %s", i.Message)
		}
	}
}

func TestCOMP002_NotDelegating(t *testing.T) {
	r := &rules.OperationDelegatesRule{}
	body := "operation postFixacao\n"
	for i := 0; i < 15; i++ {
		body += "  someDirectCode" + string(rune('a'+i)) + " = value\n"
	}
	body += "  return 0\nend;"

	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1},
	}, []parser.ProcUnit{})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "COMP002" {
			found = true
		}
	}
	if !found {
		t.Error("Esperava COMP002 para operation com código direto sem delegar para PL")
	}
}

// ============================================================
// GLB001 - Variáveis globais em PLs
// ============================================================

func TestGLB001_AllowedGlobals(t *testing.T) {
	r := &rules.GlobalVarInPLRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	$t_ds_erro$ = ""
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1, Params: []parser.Param{}},
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "GLB001" {
			t.Errorf("$t_ds_erro$ é global permitido, não deveria gerar GLB001: %s", i.Message)
		}
	}
}

func TestGLB001_ForbiddenGlobal(t *testing.T) {
	r := &rules.GlobalVarInPLRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	$t_ls_contexto$ = "algum contexto"
	$t_ds_erro$ = ""
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1, Params: []parser.Param{}},
	})

	issues := r.Run(ctx)
	found := false
	for _, i := range issues {
		if i.RuleID == "GLB001" {
			found = true
		}
	}
	if !found {
		t.Error("Esperava GLB001 para uso de $t_ls_contexto$ em PL")
	}
}

// ============================================================
// ERR003 - $t_ds_erro$ como parâmetro out e último
// ============================================================

func TestERR003_ValidErrorParam(t *testing.T) {
	r := &rules.ErrorParamRule{}
	body := `entry plPostFixacao
	params
		numeric pCdOperador : in
		$t_ds_erro$ : out
	endparams
	$t_ds_erro$ = ""
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{{
		Name:      "plPostFixacao",
		Kind:      "entry",
		Body:      body,
		LineStart: 1,
		Params: []parser.Param{
			{Name: "pCdOperador", Type: "numeric", Direction: "in"},
			{Name: "$t_ds_erro$", Type: "string", Direction: "out"},
		},
	}})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "ERR003" {
			t.Errorf("Não esperava ERR003 com parâmetro de erro correto: %s", i.Message)
		}
	}
}
