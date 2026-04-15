package analyzer_test

import (
	"os"
	"testing"

	"uniface-linter/internal/analyzer"
	"uniface-linter/internal/parser"
	"uniface-linter/internal/rules"
)

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

const goodComponentXML = `<?xml version='1.0' encoding='UTF-8' ?>
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
<DAT name="UCOMMENT">Autor    : mpereira
Data      : 01/01/2024
Funcao : CRUD da Comunicacao de Negocio</DAT>
<DAT name="UDECLARATIONS">variables
  numeric T_CD_OPERADOR
endvariables</DAT>
<DAT name="USCRIPT">#define ES227="ES227";Parametros obrigatorios.
#define ES811="ES811";Falha ao localizar dados.
#define TpNegocioVenda = 1;Tipo de Negocio Venda

;|
;Descricao: Cria uma nova fixacao de CN.
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
	vDsMetodo = "postFixacaoMi_"

	call plPostFixacaoMi(pCdOperador, pStFixacao, pStResult, $t_ds_erro$)
	#include lib_coamo:g_vld_erro

	return 0
end;postFixacaoMi

;|
;Descricao: PL que grava a fixacao no banco.
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
	variables
		string vDsMetodo
	endvariables

	$t_ds_erro$ = ""
	vDsMetodo = "plPostFixacaoMi_"

	activate "gsiso032".montarListaParaMensagem(pCdOperador, $t_ds_erro$)
	#include lib_coamo:g_vld_erro

	return 0
end;plPostFixacaoMi
</DAT>
</OCC>
</TABLE>
</UNIFACE>`

func TestAnalyze_GoodComponent(t *testing.T) {
	path := writeTempXML(t, goodComponentXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatalf("Parse falhou: %v", err)
	}

	result := analyzer.Analyze(comp, path, analyzer.Config{MaxProcLines: 100})

	// Um componente bem escrito deve ter zero ERRORs
	if result.ErrorCount > 0 {
		for _, issue := range result.Issues {
			if issue.Severity == rules.SeverityError {
				t.Errorf("Não esperava ERROR: [%s] %s", issue.RuleID, issue.Message)
			}
		}
	}
}

func TestAnalyze_ComponentName(t *testing.T) {
	path := writeTempXML(t, goodComponentXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	result := analyzer.Analyze(comp, path, analyzer.Config{})
	if result.ComponentName != "CCNO142" {
		t.Errorf("ComponentName errado: %q", result.ComponentName)
	}
}

func TestAnalyze_DisabledRules(t *testing.T) {
	path := writeTempXML(t, goodComponentXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	// Desabilitar todas as regras
	cfg := analyzer.Config{
		DisabledRules: []string{
			"NOM001", "NOM002", "NOM003", "NOM004", "NOM005",
			"COMP001", "COMP002", "COMP003",
			"DOC001", "DOC002",
			"ERR001", "ERR002", "ERR003", "ERR004",
			"GLB001", "GLB002",
		},
	}
	result := analyzer.Analyze(comp, path, cfg)
	if result.TotalIssues != 0 {
		t.Errorf("Com todas as regras desabilitadas, esperava 0 issues, got %d", result.TotalIssues)
	}
}

func TestAnalyze_SummaryCounters(t *testing.T) {
	path := writeTempXML(t, goodComponentXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	result := analyzer.Analyze(comp, path, analyzer.Config{})

	total := result.ErrorCount + result.WarningCount + result.InfoCount
	if total != result.TotalIssues {
		t.Errorf("Soma de contadores (%d) != TotalIssues (%d)", total, result.TotalIssues)
	}
}

func TestAnalyze_SummaryCategorySum(t *testing.T) {
	path := writeTempXML(t, goodComponentXML)
	comp, err := parser.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	result := analyzer.Analyze(comp, path, analyzer.Config{})

	catSum := result.Summary.Nomenclatura +
		result.Summary.Complexidade +
		result.Summary.Documentacao +
		result.Summary.TratamentoErros +
		result.Summary.VariaveisGlobais

	if catSum != result.TotalIssues {
		t.Errorf("Soma por categoria (%d) != TotalIssues (%d)", catSum, result.TotalIssues)
	}
}

func TestAnalyze_MaxProcLines(t *testing.T) {
	// Testa que o analyzer aceita MaxProcLines sem panicar
	comp := &parser.Component{
		Name:       "TEST001",
		Type:       "SERVICE",
		Comment:    "Autor: a\nData: 01/01/2024",
		Operations: []parser.ProcUnit{},
		Entries:    []parser.ProcUnit{},
	}

	bodyGrande := "operation postAlgo\n"
	for i := 0; i < 20; i++ {
		bodyGrande += "  someCode = " + string(rune('a'+i)) + "\n"
	}
	bodyGrande += "return 0\nend;"

	comp.Operations = []parser.ProcUnit{{
		Name:          "postAlgo",
		Kind:          "operation",
		Body:          bodyGrande,
		LineStart:     1,
		LineEnd:       25,
		HeaderComment: ";|\n;Descricao: algo\n;Autor: x\n;Criacao: 01/01/2024\n;Projeto: COMLOG-1",
		Params:        []parser.Param{},
	}}

	result := analyzer.Analyze(comp, "test.xml", analyzer.Config{MaxProcLines: 5})
	_ = result // Regras serão validadas conforme forem adicionadas
}
