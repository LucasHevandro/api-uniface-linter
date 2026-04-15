# uniface-linter

Analisador estático de componentes **Uniface** para as convenções COMLOG.  
Inspirado no SonarQube, valida automaticamente nomenclatura, complexidade, documentação, tratamento de erros e uso de variáveis globais.

---

## Instalação

```bash
go build -o uniface-linter ./cmd/main.go
```

Ou para instalar globalmente:
```bash
go install uniface-linter/cmd@latest
```

---

## Uso

```bash
# Analisar um arquivo
./uniface-linter cpt_cesto145.xml

# Analisar múltiplos arquivos
./uniface-linter *.xml

# Analisar um diretório inteiro
./uniface-linter ./componentes/

# Com configuração customizada
./uniface-linter --config meu-projeto.json ./componentes/

# Salvar relatório em arquivo específico
./uniface-linter -o resultado.json *.xml

# Silencioso (só gera JSON, sem output no terminal)
./uniface-linter -q *.xml

# Falhar no CI se houver WARNINGs
./uniface-linter --fail-on WARNING *.xml
```

---

## Configuração

Gere um arquivo de configuração padrão:

```bash
./uniface-linter --init
```

Isso cria `.uniface-linter.json`:

```json
{
  "max_proc_lines": 100,
  "disabled_rules": [],
  "output_file": "linter-report.json",
  "print_summary": true,
  "fail_on": "ERROR"
}
```

| Campo            | Padrão               | Descrição                                              |
|------------------|----------------------|--------------------------------------------------------|
| `max_proc_lines` | `100`                | Máximo de linhas de código por PROC (COMP001)          |
| `disabled_rules` | `[]`                 | Lista de IDs de regras a desabilitar                   |
| `output_file`    | `linter-report.json` | Caminho do relatório JSON de saída                     |
| `print_summary`  | `true`               | Imprime resumo no terminal                             |
| `fail_on`        | `ERROR`              | Nível mínimo para retornar exit code 1 (para CI/CD)   |

---

## Regras disponíveis

```bash
./uniface-linter --rules
```

| ID      | Categoria           | Descrição                                                        |
|---------|---------------------|------------------------------------------------------------------|
| NOM001  | Nomenclatura        | Operations devem usar lowerCamelCase com prefixo semântico       |
| NOM002  | Nomenclatura        | Entries devem iniciar com `pl` + UpperCamelCase                  |
| NOM003  | Nomenclatura        | `#define` deve usar UpperCamelCase ou código de mensagem (ESxxx) |
| NOM004  | Nomenclatura        | Parâmetros devem iniciar com `p`                                 |
| COMP001 | Complexidade        | PROCs não devem exceder o limite de linhas de código             |
| COMP002 | Complexidade        | Operations devem delegar lógica para PLs (entries)               |
| COMP003 | Complexidade        | PROCs com muitos parâmetros devem usar struct                    |
| DOC001  | Documentação        | PROCs devem ter cabeçalho (Descrição, Autor, Criação, Projeto)   |
| DOC002  | Documentação        | Componente deve ter cabeçalho documentado                        |
| ERR001  | Tratamento de Erros | `activate`/`call` devem ser seguidos de `#include g_vld_erro`    |
| ERR002  | Tratamento de Erros | Evitar verificação manual de `$status`; usar includes            |
| ERR003  | Tratamento de Erros | `$t_ds_erro$` deve ser parâmetro `out` e o último               |
| ERR004  | Tratamento de Erros | Evitar números mágicos; usar constantes `#define`                |
| GLB001  | Variáveis Globais   | PLs devem evitar variáveis globais (`$var$`)                     |
| GLB002  | Variáveis Globais   | PLs não devem atualizar variáveis de componente (`T_xxx`)        |

---

## Formato do relatório JSON

```json
{
  "generated_at": "2026-04-14T23:38:57",
  "total_files": 2,
  "total_issues": 32,
  "error_count": 0,
  "warning_count": 28,
  "info_count": 4,
  "components": [
    {
      "component_name": "CESTO145",
      "component_type": "SERVICE",
      "component_description": "SIS-CEST AGG CRUD LOT FAT",
      "file_path": "cpt_cesto145.xml",
      "total_issues": 12,
      "error_count": 0,
      "warning_count": 9,
      "info_count": 3,
      "issues": [
        {
          "rule_id": "ERR001",
          "severity": "WARNING",
          "category": "Tratamento de Erros",
          "message": "entry 'plGetLoteFatPK' tem 1 chamada(s) activate/call sem #include...",
          "location": "entry plGetLoteFatPK",
          "line": 13,
          "suggestion": "Adicione '#include lib_coamo:g_vld_erro' após cada activate ou call"
        }
      ],
      "summary": {
        "nomenclatura": 2,
        "complexidade": 0,
        "documentacao": 1,
        "tratamento_erros": 3,
        "variaveis_globais": 6
      }
    }
  ]
}
```

---

## Integração com CI/CD

O linter retorna **exit code 1** quando encontra issues do nível configurado em `fail_on`.

Exemplo para pipeline:
```yaml
- name: Análise Uniface
  run: ./uniface-linter --fail-on WARNING ./exportados/
```

---

## Testes

```bash
go test ./...
```

---

## Estrutura do projeto

```
uniface-linter/
├── cmd/
│   └── main.go              # Entrada CLI
├── internal/
│   ├── parser/
│   │   └── parser.go        # Parser do XML Uniface
│   ├── rules/
│   │   ├── types.go         # Tipos base (Issue, Rule, RuleContext)
│   │   ├── naming.go        # Regras NOM001-NOM004
│   │   ├── complexity.go    # Regras COMP001-COMP003
│   │   ├── documentation.go # Regras DOC001-DOC002
│   │   ├── error_handling.go# Regras ERR001-ERR004
│   │   ├── global_vars.go   # Regras GLB001-GLB002
│   │   └── rules_test.go    # Testes unitários
│   ├── analyzer/
│   │   └── analyzer.go      # Orquestra as regras
│   ├── reporter/
│   │   └── reporter.go      # Saída JSON e terminal
│   └── config/
│       └── config.go        # Configuração
└── README.md
```

---

## Changelog

### v1.1.0
- **NOM005** — Nova regra: ordem dos parâmetros (`pCdOperador` 1º, `$t_ds_erro$` último)
- **NOM003** — Corrigido falso positivo: `$triggerAbbr` (variável interna do Uniface) não é mais reportado
- **Parser** — Agora captura corretamente `$t_ds_erro$` como parâmetro tipado
- **Testes** — 32 testes unitários e de integração (parser, rules, analyzer)

### v1.0.0
- Versão inicial com 16 regras em 5 categorias
