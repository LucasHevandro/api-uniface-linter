package parser_test

import (
	"os"
	"strings"
	"testing"

	"uniface-linter/internal/parser"
)

const minimalXML = `<?xml version='1.0' encoding='UTF-8' ?>
<UNIFACE release="10.4" repversion="8" xmlengine="2.0">
<TABLE>
<DSC name="UFORM" model="DICT" system="S" pseudo="73" level="1" noupdate="0" rbk="0" ffsql="0" transnr="0" segsize="0" ufocc="0" charset=".U">
<FLD name="ULABEL" seqno="2" type="S" level="2" />
<FLD name="UDESCR" seqno="7" type="S" level="2" />
<FLD name="UCOMMENT" seqno="35" type="S" level="2" />
<FLD name="UDECLARATIONS" seqno="38" type="S" level="2" />
<FLD name="USCRIPT" seqno="39" type="S" level="2" />
</DSC>
<OCC>
<DAT name="ULABEL">CCNO142</DAT>
<DAT name="UDESCR">AGG CRUD CN</DAT>
<DAT name="UCOMMENT">Autor    : mpereira
Data      : 01/01/2024
Funcao : CRUD da CN</DAT>
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
		struct pStFixacao : in
		struct pStResult : out
		$t_ds_erro$ : out
	endparams
	variables
		string vDsMetodo
	endvariables
	$t_ds_erro$ = ""
	call plPostFixacaoMi(pCdOperador, pStFixacao, pStResult, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	return 0
end;postFixacaoMi

;|
;Descricao: PL de criacao
;Autor: mpereira
;Criacao: 01/01/2024
;Projeto: COMLOG-241
entry plPostFixacaoMi
	params
		numeric pCdOperador : in
		struct pStFixacao : in
		struct pStResult : out
		$t_ds_erro$ : out
	endparams
	$t_ds_erro$ = ""
	activate "gsiso032".montarListaParaMensagem(pCdOperador, $t_ds_erro$)
	#include lib_coamo:g_vld_erro
	return 0
end;plPostFixacaoMi
</DAT>
</OCC>
</TABLE>
</UNIFACE>`

func writeTempXML(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}

func TestParse_ComponentName(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("Parse falhou: %v", err)
	}
	if comp.Name != "CCNO142" {
		t.Errorf("Nome errado: got %q, want %q", comp.Name, "CCNO142")
	}
}

func TestParse_ComponentDescription(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if comp.Description != "AGG CRUD CN" {
		t.Errorf("Descrição errada: %q", comp.Description)
	}
}

func TestParse_Comment(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(comp.Comment, "mpereira") {
		t.Errorf("Comment não contém autor: %q", comp.Comment)
	}
}

func TestParse_Operations(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(comp.Operations) != 1 {
		t.Fatalf("Esperava 1 operation, got %d", len(comp.Operations))
	}
	op := comp.Operations[0]
	if op.Name != "postFixacaoMi" {
		t.Errorf("Nome da operation errado: %q", op.Name)
	}
	if op.Kind != "operation" {
		t.Errorf("Kind errado: %q", op.Kind)
	}
}

func TestParse_Entries(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(comp.Entries) != 1 {
		t.Fatalf("Esperava 1 entry, got %d", len(comp.Entries))
	}
	e := comp.Entries[0]
	if e.Name != "plPostFixacaoMi" {
		t.Errorf("Nome da entry errado: %q", e.Name)
	}
}

func TestParse_Params(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(comp.Operations) == 0 {
		t.Fatal("Nenhuma operation")
	}
	op := comp.Operations[0]
	if len(op.Params) != 4 {
		t.Fatalf("Esperava 4 params, got %d: %+v", len(op.Params), op.Params)
	}
	if op.Params[0].Name != "pCdOperador" || op.Params[0].Type != "numeric" || op.Params[0].Direction != "in" {
		t.Errorf("Param[0] errado: %+v", op.Params[0])
	}
	if op.Params[2].Type != "struct" || op.Params[2].Direction != "out" {
		t.Errorf("Param[2] deveria ser struct out: %+v", op.Params[2])
	}
}

func TestParse_HeaderComment(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(comp.Operations) == 0 {
		t.Fatal("Nenhuma operation")
	}
	hdr := comp.Operations[0].HeaderComment
	if !strings.Contains(hdr, "Descricao") && !strings.Contains(hdr, "Descrição") {
		t.Errorf("HeaderComment deveria conter descrição: %q", hdr)
	}
}

func TestParse_LineNumbers(t *testing.T) {
	path := writeTempXML(t, minimalXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(comp.Operations) == 0 {
		t.Fatal("Nenhuma operation")
	}
	op := comp.Operations[0]
	if op.LineStart <= 0 {
		t.Errorf("LineStart inválido: %d", op.LineStart)
	}
	if op.LineEnd <= op.LineStart {
		t.Errorf("LineEnd (%d) deve ser > LineStart (%d)", op.LineEnd, op.LineStart)
	}
}

func TestExtractDefines(t *testing.T) {
	script := `#define TpNegocioVenda = 1;Tipo Venda
#define TpNegocioCompra = 2;Tipo Compra
#define ES227="ES227";Parametros obrigatorios`

	defines := parser.ExtractDefines(script)
	if len(defines) != 3 {
		t.Errorf("Esperava 3 defines, got %d: %v", len(defines), defines)
	}
}

func TestExtractIncludes(t *testing.T) {
	script := `#include lib_coamo:g_vld_erro
call algumaCoisa
#include lib_coamo:g_vld_erroMsg`

	includes := parser.ExtractIncludes(script)
	if len(includes) != 2 {
		t.Errorf("Esperava 2 includes, got %d: %v", len(includes), includes)
	}
}

func TestHasGlobalVarUsage(t *testing.T) {
	if !parser.HasGlobalVarUsage("$t_ls_contexto$ = valor") {
		t.Error("Deveria detectar variável global $t_ls_contexto$")
	}
	if parser.HasGlobalVarUsage("nenhuma variavel global aqui") {
		t.Error("Não deveria detectar variável global onde não há")
	}
}

func TestParse_InvalidFile(t *testing.T) {
	_, err := parser.Parse("/caminho/que/nao/existe.xml")
	if err == nil {
		t.Error("Esperava erro para arquivo inexistente")
	}
}

func TestParse_MalformedXML(t *testing.T) {
	path := writeTempXML(t, `<?xml version='1.0'?><INVALIDO>`)
	_, err := parser.Parse(path)
	if err == nil {
		t.Error("Esperava erro para XML malformado sem componente")
	}
}
