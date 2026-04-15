package api

import (
	_ "embed"
	"encoding/json"
	"net/http"
)

// openAPISpec é a especificação OpenAPI 3.0 completa da API
func openAPISpec(version string) map[string]any {
	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "uniface-linter API",
			"version":     version,
			"description": "Analisador estático de componentes Uniface para as convenções COMLOG.\n\nRecebe XMLs exportados do Uniface e retorna um relatório detalhado de violações organizadas por categoria (Nomenclatura, Complexidade, Documentação, Tratamento de Erros, Variáveis Globais).\n\n**Parâmetros de query disponíveis em todos os endpoints POST:**\n- `max_lines` — threshold de linhas para COMP001 (padrão: 100)\n- `disable` — IDs de regras separados por vírgula (ex: `GLB002,DOC003`)",
			"contact": map[string]any{
				"name": "COMLOG Team",
			},
		},
		"servers": []map[string]any{
			{"url": "/", "description": "Servidor atual"},
		},
		"tags": []map[string]any{
			{"name": "Utilitários", "description": "Health check e informações da API"},
			{"name": "Análise", "description": "Análise de componentes Uniface"},
		},
		"paths": map[string]any{
			"/health": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Utilitários"},
					"summary":     "Health check",
					"description": "Retorna o status da API, versão e timestamp atual.",
					"operationId": "getHealth",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "API funcionando normalmente",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"$ref": "#/components/schemas/HealthResponse"},
									"example": map[string]any{
										"status":    "ok",
										"version":   version,
										"timestamp": "2026-04-15T00:00:00Z",
									},
								},
							},
						},
					},
				},
			},
			"/rules": map[string]any{
				"get": map[string]any{
					"tags":        []string{"Utilitários"},
					"summary":     "Listar regras",
					"description": "Retorna todas as 17 regras de validação disponíveis com ID, categoria e descrição.",
					"operationId": "getRules",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Lista de regras",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type":  "array",
										"items": map[string]any{"$ref": "#/components/schemas/RuleInfo"},
									},
									"example": []map[string]any{
										{"id": "NOM001", "category": "Nomenclatura", "description": "Operations devem usar lowerCamelCase com prefixo semântico"},
										{"id": "ERR001", "category": "Tratamento de Erros", "description": "activate/call devem ser seguidos de #include g_vld_erro"},
									},
								},
							},
						},
					},
				},
			},
			"/analyze": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Análise"},
					"summary":     "Analisar um componente",
					"description": "Analisa um único arquivo XML de componente Uniface.\n\nAceita dois formatos:\n1. **multipart/form-data** com o campo `file` contendo o XML\n2. **application/xml** com o XML diretamente no body (use o header `X-Filename` para nomear o arquivo nos logs)\n\nO resultado inclui todos os issues encontrados, contadores por severidade e um resumo por categoria.",
					"operationId": "analyzeComponent",
					"parameters":  queryParams(),
					"requestBody": map[string]any{
						"required":    true,
						"description": "Arquivo XML do componente Uniface",
						"content": map[string]any{
							"multipart/form-data": map[string]any{
								"schema": map[string]any{
									"type":     "object",
									"required": []string{"file"},
									"properties": map[string]any{
										"file": map[string]any{
											"type":        "string",
											"format":      "binary",
											"description": "Arquivo XML exportado do Uniface",
										},
									},
								},
							},
							"application/xml": map[string]any{
								"schema": map[string]any{
									"type":        "string",
									"format":      "binary",
									"description": "Conteúdo XML diretamente no body",
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Análise concluída com sucesso",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema":  map[string]any{"$ref": "#/components/schemas/AnalysisResult"},
									"example": exampleResult(),
								},
							},
						},
						"400": map[string]any{"$ref": "#/components/responses/BadRequest"},
						"422": map[string]any{"$ref": "#/components/responses/UnprocessableEntity"},
					},
				},
			},
			"/analyze/batch": map[string]any{
				"post": map[string]any{
					"tags":        []string{"Análise"},
					"summary":     "Analisar múltiplos componentes",
					"description": "Analisa vários arquivos XML em uma única requisição via multipart/form-data.\n\nUse o campo `files` para enviar múltiplos arquivos. Os resultados de cada componente são retornados em um array, junto com totalizadores globais.\n\nArquivos que falharem no parsing são registrados no campo `errors` sem interromper a análise dos demais.",
					"operationId": "analyzeBatch",
					"parameters":  queryParams(),
					"requestBody": map[string]any{
						"required":    true,
						"description": "Múltiplos arquivos XML de componentes Uniface",
						"content": map[string]any{
							"multipart/form-data": map[string]any{
								"schema": map[string]any{
									"type":     "object",
									"required": []string{"files"},
									"properties": map[string]any{
										"files": map[string]any{
											"type":        "array",
											"items":       map[string]any{"type": "string", "format": "binary"},
											"description": "Lista de arquivos XML do Uniface",
										},
									},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "Análise em lote concluída",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"$ref": "#/components/schemas/BatchResult"},
								},
							},
						},
						"400": map[string]any{"$ref": "#/components/responses/BadRequest"},
					},
				},
			},
		},
		"components": map[string]any{
			"schemas":   schemas(),
			"responses": responses(),
		},
	}
}

func queryParams() []map[string]any {
	return []map[string]any{
		{
			"name":        "max_lines",
			"in":          "query",
			"required":    false,
			"description": "Threshold máximo de linhas de código por PROC. Acima desse valor a regra COMP001 é acionada.",
			"schema":      map[string]any{"type": "integer", "default": 100, "minimum": 1, "example": 150},
		},
		{
			"name":        "disable",
			"in":          "query",
			"required":    false,
			"description": "IDs de regras a desabilitar, separados por vírgula. Case-insensitive.",
			"schema":      map[string]any{"type": "string", "example": "GLB002,DOC003"},
		},
	}
}

func schemas() map[string]any {
	return map[string]any{
		"HealthResponse": map[string]any{
			"type":     "object",
			"required": []string{"status", "version", "timestamp"},
			"properties": map[string]any{
				"status":    map[string]any{"type": "string", "example": "ok"},
				"version":   map[string]any{"type": "string", "example": "1.1.0"},
				"timestamp": map[string]any{"type": "string", "format": "date-time"},
			},
		},
		"RuleInfo": map[string]any{
			"type":     "object",
			"required": []string{"id", "category", "description"},
			"properties": map[string]any{
				"id":          map[string]any{"type": "string", "example": "NOM001"},
				"category":    map[string]any{"type": "string", "example": "Nomenclatura"},
				"description": map[string]any{"type": "string", "example": "Operations devem usar lowerCamelCase com prefixo semântico"},
			},
		},
		"Severity": map[string]any{
			"type":        "string",
			"enum":        []string{"ERROR", "WARNING", "INFO"},
			"description": "ERROR — violação obrigatória; WARNING — prática não recomendada; INFO — observação",
		},
		"Issue": map[string]any{
			"type":     "object",
			"required": []string{"rule_id", "severity", "category", "message", "location"},
			"properties": map[string]any{
				"rule_id":    map[string]any{"type": "string", "example": "ERR001"},
				"severity":   map[string]any{"$ref": "#/components/schemas/Severity"},
				"category":   map[string]any{"type": "string", "example": "Tratamento de Erros"},
				"message":    map[string]any{"type": "string", "example": "entry 'plGetLoteFatPK' tem 1 chamada(s) activate/call sem #include de verificação de erro"},
				"location":   map[string]any{"type": "string", "example": "entry plGetLoteFatPK"},
				"line":       map[string]any{"type": "integer", "example": 13},
				"suggestion": map[string]any{"type": "string", "example": "Adicione '#include lib_coamo:g_vld_erro' após cada activate ou call"},
			},
		},
		"Summary": map[string]any{
			"type":        "object",
			"description": "Contagem de issues por categoria",
			"properties": map[string]any{
				"nomenclatura":      map[string]any{"type": "integer", "example": 2},
				"complexidade":      map[string]any{"type": "integer", "example": 0},
				"documentacao":      map[string]any{"type": "integer", "example": 1},
				"tratamento_erros":  map[string]any{"type": "integer", "example": 3},
				"variaveis_globais": map[string]any{"type": "integer", "example": 3},
			},
		},
		"AnalysisResult": map[string]any{
			"type":     "object",
			"required": []string{"component_name", "component_type", "file_path", "total_issues", "error_count", "warning_count", "info_count", "issues", "summary"},
			"properties": map[string]any{
				"component_name":        map[string]any{"type": "string", "example": "CESTO145"},
				"component_type":        map[string]any{"type": "string", "example": "SERVICE"},
				"component_description": map[string]any{"type": "string", "example": "SIS-CEST AGG CRUD LOT FAT"},
				"file_path":             map[string]any{"type": "string", "example": "cpt_cesto145.xml"},
				"total_issues":          map[string]any{"type": "integer", "example": 10},
				"error_count":           map[string]any{"type": "integer", "example": 0},
				"warning_count":         map[string]any{"type": "integer", "example": 6},
				"info_count":            map[string]any{"type": "integer", "example": 4},
				"issues": map[string]any{
					"type":     "array",
					"items":    map[string]any{"$ref": "#/components/schemas/Issue"},
					"nullable": true,
				},
				"summary": map[string]any{"$ref": "#/components/schemas/Summary"},
			},
		},
		"BatchResult": map[string]any{
			"type":     "object",
			"required": []string{"generated_at", "total_files", "total_issues", "error_count", "warning_count", "info_count", "components"},
			"properties": map[string]any{
				"generated_at":  map[string]any{"type": "string", "format": "date-time"},
				"total_files":   map[string]any{"type": "integer", "example": 2},
				"total_issues":  map[string]any{"type": "integer", "example": 28},
				"error_count":   map[string]any{"type": "integer", "example": 0},
				"warning_count": map[string]any{"type": "integer", "example": 24},
				"info_count":    map[string]any{"type": "integer", "example": 4},
				"components": map[string]any{
					"type":  "array",
					"items": map[string]any{"$ref": "#/components/schemas/AnalysisResult"},
				},
				"errors": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"nullable":    true,
					"description": "Arquivos que falharam no parsing (não interrompem o batch)",
				},
			},
		},
		"ErrorResponse": map[string]any{
			"type":     "object",
			"required": []string{"error"},
			"properties": map[string]any{
				"error": map[string]any{"type": "string", "example": "arquivo XML vazio ou não enviado (field: 'file')"},
			},
		},
	}
}

func responses() map[string]any {
	return map[string]any{
		"BadRequest": map[string]any{
			"description": "Requisição inválida (arquivo ausente, field errado, multipart malformado)",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
				},
			},
		},
		"UnprocessableEntity": map[string]any{
			"description": "XML enviado não é um componente Uniface válido",
			"content": map[string]any{
				"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/ErrorResponse"},
					"example": map[string]any{
						"error": "falha ao parsear XML 'comp.xml': nenhum componente encontrado no XML",
					},
				},
			},
		},
	}
}

func exampleResult() map[string]any {
	return map[string]any{
		"component_name":        "CESTO145",
		"component_type":        "SERVICE",
		"component_description": "SIS-CEST AGG CRUD LOT FAT",
		"file_path":             "cpt_cesto145.xml",
		"total_issues":          3,
		"error_count":           0,
		"warning_count":         2,
		"info_count":            1,
		"issues": []map[string]any{
			{
				"rule_id":    "ERR001",
				"severity":   "WARNING",
				"category":   "Tratamento de Erros",
				"message":    "entry 'plGetLoteFatPK' tem 1 chamada(s) activate/call sem #include de verificação de erro logo após",
				"location":   "entry plGetLoteFatPK",
				"line":       13,
				"suggestion": "Adicione '#include lib_coamo:g_vld_erro' após cada activate ou call",
			},
			{
				"rule_id":  "GLB001",
				"severity": "WARNING",
				"category": "Variáveis Globais",
				"message":  "Entry 'plPostLoteFat' usa variáveis globais em PL: $t_ls_contexto$",
				"location": "entry plPostLoteFat",
				"line":     185,
			},
			{
				"rule_id":  "DOC003",
				"severity": "INFO",
				"category": "Documentação",
				"message":  "Componente não possui comentários de modificação com referência ao item",
				"location": "componente CESTO145",
			},
		},
		"summary": map[string]any{
			"nomenclatura":      0,
			"complexidade":      0,
			"documentacao":      1,
			"tratamento_erros":  1,
			"variaveis_globais": 1,
		},
	}
}

// swaggerUIHTML retorna a página HTML do Swagger UI carregada via CDN
func swaggerUIHTML(version string) string {
	return `<!DOCTYPE html>
<html lang="pt-BR">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>uniface-linter API ` + version + `</title>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui.min.css">
  <style>
    body { margin: 0; background: #f8f9fa; }
    .swagger-ui .topbar { background: #1a3a5c; }
    .swagger-ui .topbar .download-url-wrapper { display: none; }
    .swagger-ui .topbar-wrapper .link { pointer-events: none; }
    .swagger-ui .topbar-wrapper img { display: none; }
    .swagger-ui .topbar-wrapper::after {
      content: 'uniface-linter API ` + version + `';
      color: white;
      font-size: 18px;
      font-weight: 600;
      padding: 0 16px;
    }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-bundle.min.js"></script>
  <script src="https://cdnjs.cloudflare.com/ajax/libs/swagger-ui/5.17.14/swagger-ui-standalone-preset.min.js"></script>
  <script>
    window.onload = () => {
      SwaggerUIBundle({
        url: '/openapi.json',
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: 'StandaloneLayout',
        tryItOutEnabled: true,
        requestInterceptor: (req) => req,
        defaultModelsExpandDepth: 2,
        defaultModelExpandDepth: 2,
      });
    };
  </script>
</body>
</html>`
}

// ── Handlers ────────────────────────────────────────────────────

func (s *Server) handleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := openAPISpec(s.version)
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.Encode(spec)
}

func (s *Server) handleSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerUIHTML(s.version)))
}
