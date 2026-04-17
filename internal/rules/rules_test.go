package rules_test

import (
	"fmt"
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

func TestNOM001_ValidCRUD(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacaoMi", "", 1),
		op("getClientePk", "", 5),
		op("findFixacoesMi", "", 10),
		op("putFixacao", "", 15),
		op("delFixacao", "", 20),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.Severity == rules.SeverityError {
			t.Errorf("não esperava ERROR para operation CRUD válida: %s", i.Message)
		}
	}
}

func TestNOM001_ValidService(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("efetivarProcesso", "", 1),
		op("calcularTotalCN", "", 5),
		op("validarParametros", "", 10),
		op("gerarHistorico", "", 15),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.Severity == rules.SeverityError {
			t.Errorf("não esperava ERROR para operation de serviço válida: %s", i.Message)
		}
	}
}

func TestNOM001_NotLowerCamelCase(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("PostFixacaoMi", "", 1),
		op("GET_CLIENTE", "", 5),
	}, []parser.ProcUnit{})
	errCount := 0
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM001" && i.Severity == rules.SeverityError {
			errCount++
		}
	}
	if errCount != 2 {
		t.Errorf("esperava 2 ERRORs NOM001 (casing inválido), got %d", errCount)
	}
}

func TestNOM001_UnknownPrefix(t *testing.T) {
	r := &rules.NamingOperationsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("executarAlgo", "", 1),
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM001" && i.Severity == rules.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Error("esperava WARNING NOM001 para operation sem prefixo reconhecido")
	}
}

// ============================================================
// NOM002 - Nomenclatura de Entries
// ============================================================

func TestNOM002_ValidEntry(t *testing.T) {
	r := &rules.NamingEntriesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		entry("plPostFixacaoMi", "", 1),
		entry("plGetClientePk", "", 5),
	})
	for _, i := range r.Run(ctx) {
		if i.Severity == rules.SeverityError {
			t.Errorf("não esperava ERROR para entry válida: %s", i.Message)
		}
	}
}

func TestNOM002_MissingPLPrefix(t *testing.T) {
	r := &rules.NamingEntriesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		entry("postFixacao", "", 1),
		entry("getCliente", "", 5),
	})
	errCount := 0
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM002" && i.Severity == rules.SeverityError {
			errCount++
		}
	}
	if errCount != 2 {
		t.Errorf("esperava 2 ERRORs NOM002 (sem prefixo pl), got %d", errCount)
	}
}

func TestNOM002_PLPrefixLowercaseAfter(t *testing.T) {
	r := &rules.NamingEntriesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		entry("plpostFixacao", "", 1), // pl ok, mas 'p' minúsculo depois
	})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM002" && i.Severity == rules.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Error("esperava WARNING NOM002 para entry com letra minúscula após pl")
	}
}

// ============================================================
// NOM003 - Nomenclatura de #define
// ============================================================

func TestNOM003_ValidUpperCamelCase(t *testing.T) {
	r := &rules.NamingDefinesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{})
	ctx.Script = "#define TpNegocioVenda = 1;Tipo\n#define CdStatusAtivo = 2;Status"
	for _, i := range r.Run(ctx) {
		t.Errorf("não esperava issue para #define UpperCamelCase: %s", i.Message)
	}
}

func TestNOM003_ValidMsgCode(t *testing.T) {
	r := &rules.NamingDefinesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{})
	ctx.Script = `#define ES227="ES227";Parametros obrigatorios.
#define EN1720="EN1720";Falha ao ajustar quantidade.
#define EW100="EW100";Aviso.`
	for _, i := range r.Run(ctx) {
		t.Errorf("não esperava issue para #define com código de mensagem: %s", i.Message)
	}
}

func TestNOM003_InvalidLowerCase(t *testing.T) {
	r := &rules.NamingDefinesRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{})
	ctx.Script = "#define tpNegocioVenda = 1;Tipo\n#define cdStatus = 2;Status"
	count := 0
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM003" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("esperava 2 WARNINGs NOM003 para #define sem UpperCamelCase, got %d", count)
	}
}

// ============================================================
// NOM004 - Parâmetros com 'p' e variáveis com 'v'
// ============================================================

func TestNOM004_ValidParams(t *testing.T) {
	r := &rules.NamingParamsVarsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("numeric", "pCdOperador", "in"),
			param("struct", "pStFixacao", "in"),
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		t.Errorf("não esperava issue para parâmetros válidos: %s", i.Message)
	}
}

func TestNOM004_InvalidParams(t *testing.T) {
	r := &rules.NamingParamsVarsRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("numeric", "cdOperador", "in"),
			param("struct", "stFixacao", "in"),
		),
	}, []parser.ProcUnit{})
	count := 0
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM004" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("esperava 2 WARNINGs NOM004 (parâmetros sem 'p'), got %d", count)
	}
}

func TestNOM004_ValidVariables(t *testing.T) {
	r := &rules.NamingParamsVarsRule{}
	body := `operation postFixacao
variables
  string vDsMetodo, vDsMsg
  numeric vNrCN
  struct vStResult
endvariables
return 0
end;`
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM004" {
			t.Errorf("não esperava issue para variáveis válidas: %s", i.Message)
		}
	}
}

func TestNOM004_InvalidVariables(t *testing.T) {
	r := &rules.NamingParamsVarsRule{}
	body := `operation postFixacao
variables
  string dsMetodo
  numeric nrCN
endvariables
return 0
end;`
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	count := 0
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM004" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("esperava 2 WARNINGs NOM004 (variáveis sem 'v'), got %d", count)
	}
}

// ============================================================
// NOM005 - Ordem dos parâmetros
// ============================================================

func TestNOM005_PostWithCdOperadorFirst(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("numeric", "pCdOperador", "in"),
			param("struct", "pStFixacao", "in"),
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityError {
			t.Errorf("não esperava ERROR NOM005 com ordem correta: %s", i.Message)
		}
	}
}

func TestNOM005_PostWithoutCdOperadorFirst(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("struct", "pStFixacao", "in"), // pCdOperador ausente no 1º lugar
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityError {
			found = true
		}
	}
	if !found {
		t.Error("esperava ERROR NOM005: postFixacao sem pCdOperador como 1º parâmetro")
	}
}

func TestNOM005_GetDoesNotRequireCdOperador(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("getFixacaoPk", "", 1,
			param("numeric", "pNrCN", "in"),
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityError {
			t.Errorf("get não deve exigir pCdOperador, mas gerou ERROR: %s", i.Message)
		}
	}
}

func TestNOM005_EntryPlPostRequiresCdOperador(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		entry("plPostFixacao", "", 1,
			param("struct", "pStFixacao", "in"), // pCdOperador ausente
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityError {
			found = true
		}
	}
	if !found {
		t.Error("esperava ERROR NOM005: plPostFixacao (entry) sem pCdOperador como 1º parâmetro")
	}
}

func TestNOM005_ErrNotLast(t *testing.T) {
	r := &rules.ParamOrderRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("numeric", "pCdOperador", "in"),
			param("string", "$t_ds_erro$", "out"), // deve ser o último
			param("struct", "pStResult", "out"),
		),
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "NOM005" && i.Severity == rules.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Error("esperava WARNING NOM005: $t_ds_erro$ não é o último parâmetro")
	}
}

// ============================================================
// COMP001 - Tamanho máximo de PROCs
// ============================================================

func TestCOMP001_SmallProc(t *testing.T) {
	r := &rules.ProcSizeRule{MaxLines: 100}
	body := "operation postFixacao\n"
	for i := 0; i < 20; i++ {
		body += "  call plPostFixacao(pCdOperador)\n"
	}
	body += "return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP001" {
			t.Errorf("não esperava COMP001 para PROC com 20 linhas (máx 100): %s", i.Message)
		}
	}
}

func TestCOMP001_LargeProc_Warning(t *testing.T) {
	r := &rules.ProcSizeRule{MaxLines: 100}
	body := "operation postFixacao\n"
	for i := 0; i < 120; i++ {
		body += fmt.Sprintf("  code%d = value\n", i)
	}
	body += "return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP001" && i.Severity == rules.SeverityWarning {
			found = true
		}
	}
	if !found {
		t.Error("esperava WARNING COMP001 para PROC com 120 linhas (máx 100)")
	}
}

func TestCOMP001_VeryLargeProc_Error(t *testing.T) {
	r := &rules.ProcSizeRule{MaxLines: 100}
	body := "operation postFixacao\n"
	for i := 0; i < 210; i++ {
		body += fmt.Sprintf("  code%d = value\n", i)
	}
	body += "return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP001" && i.Severity == rules.SeverityError {
			found = true
		}
	}
	if !found {
		t.Error("esperava ERROR COMP001 para PROC com 210 linhas (> 2x o máximo de 100)")
	}
}

func TestCOMP001_CommentsAndBlanksNotCounted(t *testing.T) {
	r := &rules.ProcSizeRule{MaxLines: 10}
	// 5 linhas de código real + muitos comentários e brancos
	body := "operation postFixacao\n"
	for i := 0; i < 20; i++ {
		body += "  ; comentario que não conta\n"
		body += "\n"
	}
	for i := 0; i < 5; i++ {
		body += fmt.Sprintf("  code%d = value\n", i)
	}
	body += "return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP001" {
			t.Errorf("comentários e linhas em branco não devem ser contados: %s", i.Message)
		}
	}
}

// ============================================================
// COMP002 - Operation deve delegar para PL
// ============================================================

func TestCOMP002_DelegatesCorrectly(t *testing.T) {
	r := &rules.OperationDelegatesRule{}
	body := "operation postFixacao\n"
	for i := 0; i < 12; i++ {
		body += "  ; comentario\n"
	}
	body += "  call plPostFixacao(pCdOperador, $t_ds_erro$)\n"
	body += "  #include lib_coamo:g_vld_erro\n"
	body += "  return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP002" {
			t.Errorf("não esperava COMP002 quando operation delega para PL: %s", i.Message)
		}
	}
}

func TestCOMP002_NotDelegating(t *testing.T) {
	r := &rules.OperationDelegatesRule{}
	body := "operation postFixacao\n"
	for i := 0; i < 15; i++ {
		body += fmt.Sprintf("  codigo%d = valor\n", i)
	}
	body += "  return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP002" {
			found = true
		}
	}
	if !found {
		t.Error("esperava COMP002 para operation com código direto sem delegar para PL")
	}
}

func TestCOMP002_SmallOpNotChecked(t *testing.T) {
	r := &rules.OperationDelegatesRule{}
	// operation com <= 10 linhas não deve ser verificada
	body := "operation postFixacao\n"
	for i := 0; i < 5; i++ {
		body += fmt.Sprintf("  codigo%d = valor\n", i)
	}
	body += "  return 0\nend;"
	ctx := makeCtx([]parser.ProcUnit{
		{Name: "postFixacao", Kind: "operation", Body: body, LineStart: 1, Params: []parser.Param{}},
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP002" {
			t.Errorf("operations pequenas (<=10 linhas) não devem ser verificadas: %s", i.Message)
		}
	}
}

// ============================================================
// COMP003 - Número excessivo de parâmetros sem struct
// ============================================================

func TestCOMP003_ManyParamsWithStruct(t *testing.T) {
	r := &rules.ParamCountRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("numeric", "pCdOperador", "in"),
			param("numeric", "pNrCN", "in"),
			param("numeric", "pNrFix", "in"),
			param("string", "pDsObs", "in"),
			param("numeric", "pTpNeg", "in"),
			param("struct", "pStFixacao", "in"), // tem struct → sem issue
			param("struct", "pStResult", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP003" {
			t.Errorf("não esperava COMP003 quando usa struct: %s", i.Message)
		}
	}
}

func TestCOMP003_ManyParamsWithoutStruct(t *testing.T) {
	r := &rules.ParamCountRule{}
	ctx := makeCtx([]parser.ProcUnit{
		op("postFixacao", "", 1,
			param("numeric", "pCdOperador", "in"),
			param("numeric", "pNrCN", "in"),
			param("numeric", "pNrFix", "in"),
			param("string", "pDsObs", "in"),
			param("numeric", "pTpNeg", "in"),
			param("numeric", "pVlPreco", "in"),
			param("numeric", "pCdMoeda", "in"),
			param("string", "$t_ds_erro$", "out"), // 8 params, sem struct
		),
	}, []parser.ProcUnit{})
	found := false
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP003" {
			found = true
		}
	}
	if !found {
		t.Error("esperava COMP003 para PROC com 8 parâmetros sem struct")
	}
}

func TestCOMP003_ExactLimitNoIssue(t *testing.T) {
	r := &rules.ParamCountRule{}
	// exatamente 7 parâmetros → sem issue
	ctx := makeCtx([]parser.ProcUnit{
		op("getFixacao", "", 1,
			param("numeric", "pNrCN", "in"),
			param("numeric", "pNrFix", "in"),
			param("numeric", "pTpDep", "in"),
			param("numeric", "pCdDep", "in"),
			param("numeric", "pNrLote", "in"),
			param("numeric", "pVlPreco", "out"),
			param("string", "$t_ds_erro$", "out"),
		),
	}, []parser.ProcUnit{})
	for _, i := range r.Run(ctx) {
		if i.RuleID == "COMP003" {
			t.Errorf("7 parâmetros não deve gerar COMP003: %s", i.Message)
		}
	}
}

// ============================================================
// ERR001 - #include de tratamento de erro após activate/call
// ============================================================

func TestERR001_ActivateWithInclude(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
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
			t.Errorf("não esperava ERR001 quando #include está presente: %s", i.Message)
		}
	}
}

func TestERR001_CallWithInclude(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	call plValidate(pCdOperador, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			t.Errorf("não esperava ERR001 quando #include está presente: %s", i.Message)
		}
	}
}

func TestERR001_ActivateWithoutInclude(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
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
		t.Error("esperava ERR001 para activate sem #include")
	}
}

func TestERR001_CallWithoutInclude(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	call plValidate(pCdOperador, $t_ds_erro$)
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
		t.Error("esperava ERR001 para call sem #include")
	}
}

func TestERR001_MultipleCallsOneMissing(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	call plValidate(pCdOperador, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	activate "serv001".postFixacao(pCdOperador, $t_ds_erro$)
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	count := 0
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("esperava 1 issue ERR001 (só o activate sem include), got %d", count)
	}
}

func TestERR001_MultipleCallsBothMissing(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
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
	count := 0
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("esperava 2 issues ERR001 (call e activate sem include), got %d", count)
	}
}

func TestERR001_ReturnErroExec(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	activate "serv001".postFixacao(pCdOperador, $t_ds_erro$)
	return<g_erroexec>
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			t.Errorf("não esperava ERR001 quando return<g_erroexec> está presente: %s", i.Message)
		}
	}
}

func TestERR001_MultiLineContinuationWithInclude(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	// activate quebrado em duas linhas com %\ — o include vem após a linha de continuação
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	activate "cesto147".gera_historico("CEST_LOTEFAT", $componentname, 1, %\
	pCdOperador, $datim, vLsOcc, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			t.Errorf("não esperava ERR001 para activate multi-linha com #include presente: %s", i.Message)
		}
	}
}

func TestERR001_MultiLineContinuationWithoutInclude(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	activate "cesto147".gera_historico("CEST_LOTEFAT", $componentname, 1, %\
	pCdOperador, $datim, vLsOcc, $t_ds_erro$)
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
		t.Error("esperava ERR001 para activate multi-linha sem tratamento de erro")
	}
}

func TestERR001_IncludeWithBlankLineBetween(t *testing.T) {
	r := &rules.ErrorHandlingRule{}
	// Linha em branco entre call e #include: ainda deve ser aceito
	body := `entry plPostFixacao
	params
		$t_ds_erro$ : out
	endparams
	call plValidate(pCdOperador, $t_ds_erro$)

	#include lib_coamo:g_vld_erro
	return 0
end;`

	ctx := makeCtx([]parser.ProcUnit{}, []parser.ProcUnit{
		{Name: "plPostFixacao", Kind: "entry", Body: body, LineStart: 1},
	})

	issues := r.Run(ctx)
	for _, i := range issues {
		if i.RuleID == "ERR001" {
			t.Errorf("não esperava ERR001 quando #include está presente (com linha em branco): %s", i.Message)
		}
	}
}

// ============================================================
// DOC001 - Cabeçalho obrigatório nas PROCs
// ============================================================

func TestDOC001_ValidHeader(t *testing.T) {
	r := &rules.ProcHeaderRule{}
	unit := parser.ProcUnit{
		Name:          "plValid",
		Kind:          "entry",
		HeaderComment: ";|\n;Descrição: Faz algo\n;Autor: Fulano\n;Criação: 01/01/2024\n;Projeto: 123",
	}
	ctx := makeCtx(nil, []parser.ProcUnit{unit})

	issues := r.Run(ctx)
	if len(issues) > 0 {
		t.Errorf("não esperava issues para cabeçalho válido, got %d", len(issues))
	}
}

func TestDOC001_MissingOneField(t *testing.T) {
	r := &rules.ProcHeaderRule{}
	unit := parser.ProcUnit{
		Name:          "plMissing",
		Kind:          "entry",
		HeaderComment: ";|\n;Descrição: Faz algo\n;Autor: Fulano\n;Criação: 01/01/2024",
	}
	ctx := makeCtx(nil, []parser.ProcUnit{unit})

	issues := r.Run(ctx)
	if len(issues) != 1 {
		t.Fatalf("esperava 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != rules.SeverityWarning {
		t.Errorf("esperava severidade Warning, got %s", issues[0].Severity)
	}
}

func TestDOC001_MissingThreeFields(t *testing.T) {
	r := &rules.ProcHeaderRule{}
	unit := parser.ProcUnit{
		Name:          "plMissing",
		Kind:          "entry",
		HeaderComment: ";|\n;Descrição: Faz algo",
	}
	ctx := makeCtx(nil, []parser.ProcUnit{unit})

	issues := r.Run(ctx)
	if len(issues) != 1 {
		t.Fatalf("esperava 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != rules.SeverityError {
		t.Errorf("esperava severidade Error, got %s", issues[0].Severity)
	}
}

// ============================================================
// DOC002 - Cabeçalho do componente
// ============================================================

func TestDOC002_ValidHeader(t *testing.T) {
	r := &rules.ComponentHeaderRule{}
	ctx := makeCtx(nil, nil)
	ctx.Comment = "Autor: Teste\nData: 01/01/2024\nFunção: Validar"

	issues := r.Run(ctx)
	if len(issues) > 0 {
		t.Errorf("não esperava issues para comentário de componente válido, got %d", len(issues))
	}
}

func TestDOC002_EmptyHeader(t *testing.T) {
	r := &rules.ComponentHeaderRule{}
	ctx := makeCtx(nil, nil)
	ctx.Comment = "   \n"

	issues := r.Run(ctx)
	if len(issues) != 1 {
		t.Fatalf("esperava 1 issue para comentário vazio, got %d", len(issues))
	}
	if issues[0].Severity != rules.SeverityError {
		t.Errorf("esperava severidade Error, got %s", issues[0].Severity)
	}
}

func TestDOC002_MissingData(t *testing.T) {
	r := &rules.ComponentHeaderRule{}
	ctx := makeCtx(nil, nil)
	ctx.Comment = "Autor: Fulano\nFunção: Validar"

	issues := r.Run(ctx)
	if len(issues) != 1 {
		t.Fatalf("esperava 1 issue para data faltando, got %d", len(issues))
	}
	if issues[0].Severity != rules.SeverityWarning {
		t.Errorf("esperava severidade Warning, got %s", issues[0].Severity)
	}
}
