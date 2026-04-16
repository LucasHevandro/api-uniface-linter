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
