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

// Placeholder para evitar erro de "no test functions"
func TestPlaceholder(t *testing.T) {}
