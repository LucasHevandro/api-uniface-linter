package parser

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Component representa um componente Uniface parseado
type Component struct {
	Name         string
	Description  string
	Type         string // FORM, SERVICE, REPORT, etc
	Comment      string
	Declarations string
	Script       string
	Operations   []ProcUnit
	Entries      []ProcUnit
}

// ProcUnit representa uma operation ou entry
type ProcUnit struct {
	Name        string
	Kind        string // "operation" ou "entry"
	Params      []Param
	Variables   []string
	Body        string
	LineStart   int
	LineEnd     int
	HeaderComment string
}

// Param representa um parâmetro de operation/entry
type Param struct {
	Name      string
	Type      string
	Direction string // in, out, inout
}

// Parse lê e interpreta o XML de um componente Uniface
func Parse(filePath string) (*Component, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}
	defer f.Close()

	raw, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo: %w", err)
	}

	// O XML do Uniface usa entidades não-padrão e DOCTYPE customizado
	// Removemos a declaração DOCTYPE e substituímos entidades problemáticas
	content := string(raw)
	content = stripDoctype(content)
	content = sanitizeEntities(content)

	comp, err := parseXML([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("erro ao parsear XML: %w", err)
	}

	return comp, nil
}

func stripDoctype(s string) string {
	// Remove <!DOCTYPE ...>
	re := regexp.MustCompile(`<!DOCTYPE[^>]*>`)
	return re.ReplaceAllString(s, "")
}

func sanitizeEntities(s string) string {
	// Substitui entidades Uniface por placeholder vazio, 
	// mas preserva as entidades padrões do XML para o unmarshal não quebrar '<' e '>'
	re := regexp.MustCompile(`&[a-zA-Z][a-zA-Z0-9_]*;`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		switch match {
		case "&lt;", "&gt;", "&amp;", "&quot;", "&apos;":
			return match
		default:
			return ""
		}
	})
}

type unifaceXML struct {
	XMLName xml.Name `xml:"UNIFACE"`
	Table   struct {
		OCCs []struct {
			DATs []struct {
				Name    string `xml:"name,attr"`
				Content string `xml:",chardata"`
			} `xml:"DAT"`
		} `xml:"OCC"`
	} `xml:"TABLE"`
}

func parseXML(data []byte) (*Component, error) {
	var u unifaceXML
	if err := xml.Unmarshal(data, &u); err != nil {
		return nil, err
	}

	if len(u.Table.OCCs) == 0 {
		return nil, fmt.Errorf("nenhum componente encontrado no XML")
	}

	occ := u.Table.OCCs[0]
	datMap := make(map[string]string)
	for _, dat := range occ.DATs {
		datMap[dat.Name] = dat.Content
	}

	comp := &Component{
		Name:         strings.TrimSpace(datMap["ULABEL"]),
		Description:  strings.TrimSpace(datMap["UDESCR"]),
		Comment:      strings.TrimSpace(datMap["UCOMMENT"]),
		Declarations: strings.TrimSpace(datMap["UDECLARATIONS"]),
		Script:       datMap["USCRIPT"],
	}

	// Detectar tipo pelo FTYP ou pelo conteúdo
	comp.Type = detectType(datMap)

	// Parsear operations e entries do script
	comp.Operations, comp.Entries = parseScript(comp.Script)

	return comp, nil
}

func detectType(datMap map[string]string) string {
	ftyp := strings.TrimSpace(datMap["FTYP"])
	switch ftyp {
	case "F":
		return "FORM"
	case "R":
		return "REPORT"
	case "S":
		return "SERVICE"
	case "P":
		return "PROCESS"
	default:
		// Tentar inferir pelo nome ou descrição
		label := strings.ToUpper(datMap["ULABEL"])
		switch {
		case strings.Contains(label, "FRM") || strings.Contains(label, "CPT"):
			return "FORM"
		case strings.Contains(label, "REL") || strings.Contains(label, "RPT"):
			return "REPORT"
		default:
			return "SERVICE"
		}
	}
}

var (
	reOperation = regexp.MustCompile(`(?m)^\s*(operation|entry)\s+(\S+)\s*$`)
	reParams    = regexp.MustCompile(`(?m)^\s*params\s*$`)
	reEndParams = regexp.MustCompile(`(?m)^\s*endparams\s*$`)
	reEnd       = regexp.MustCompile(`(?m)^\s*end;`)
	reVariables = regexp.MustCompile(`(?m)^\s*variables\s*$`)
	reEndVars   = regexp.MustCompile(`(?m)^\s*endvariables\s*$`)
	reParam        = regexp.MustCompile(`(?m)^\s*(numeric|string|boolean|struct|date|datetime|float)\s+(\S+)\s*:\s*(in|out|inout)`)
	reParamSpecial = regexp.MustCompile(`(?m)^\s*(\$[a-zA-Z_][a-zA-Z0-9_]*\$)\s*:\s*(in|out|inout)`)
	reDefine       = regexp.MustCompile(`(?m)^#define\s+(\S+)\s*=`)
	reGlobalVar = regexp.MustCompile(`(?m)^\s*\$[a-zA-Z_][a-zA-Z0-9_]*\$`)
	reInclude   = regexp.MustCompile(`(?m)^#include\s+(\S+)`)
)

func parseScript(script string) (ops []ProcUnit, entries []ProcUnit) {
	lines := strings.Split(script, "\n")

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		// Detectar início de operation ou entry
		var kind, name string
		if strings.HasPrefix(line, "operation ") {
			kind = "operation"
			name = strings.TrimSpace(strings.TrimPrefix(line, "operation"))
		} else if strings.HasPrefix(line, "entry ") {
			kind = "entry"
			name = strings.TrimSpace(strings.TrimPrefix(line, "entry"))
		} else {
			i++
			continue
		}

		// Coletar header comment (linhas ;| antes da declaração)
		headerComment := extractHeaderComment(lines, i)

		unit := ProcUnit{
			Name:          name,
			Kind:          kind,
			LineStart:     i + 1,
			HeaderComment: headerComment,
		}

		// Coletar body até o próximo end;
		bodyLines := []string{line}
		j := i + 1
		inParams := false
		inVars := false

		for j < len(lines) {
			bodyLine := lines[j]
			trimmed := strings.TrimSpace(bodyLine)
			bodyLines = append(bodyLines, bodyLine)

			if reParams.MatchString(trimmed) {
				inParams = true
			} else if reEndParams.MatchString(trimmed) {
				inParams = false
			} else if reVariables.MatchString(trimmed) {
				_ = inVars
				inVars = true
			} else if reEndVars.MatchString(trimmed) {
				inVars = false
				_ = inVars
			}

			// Parsear parâmetros (tipo explícito e especiais como $t_ds_erro$)
			if inParams {
				if m := reParam.FindStringSubmatch(bodyLine); m != nil {
					unit.Params = append(unit.Params, Param{
						Type:      m[1],
						Name:      m[2],
						Direction: m[3],
					})
				} else if m := reParamSpecial.FindStringSubmatch(bodyLine); m != nil {
					unit.Params = append(unit.Params, Param{
						Type:      "string",
						Name:      m[1],
						Direction: m[2],
					})
				}
			}

			// Detectar fim da unit
			if reEnd.MatchString(trimmed) {
				unit.LineEnd = j + 1
				break
			}

			j++
		}

		unit.Body = strings.Join(bodyLines, "\n")
		if unit.LineEnd == 0 {
			unit.LineEnd = j + 1
		}

		if kind == "operation" {
			ops = append(ops, unit)
		} else {
			entries = append(entries, unit)
		}

		i = j + 1
	}

	return
}

func extractHeaderComment(lines []string, procLine int) string {
	// Procura por bloco de comentário ;| nas linhas anteriores
	var commentLines []string
	for k := procLine - 1; k >= 0 && k >= procLine-20; k-- {
		trimmed := strings.TrimSpace(lines[k])
		if strings.HasPrefix(trimmed, ";") {
			commentLines = append([]string{trimmed}, commentLines...)
		} else if trimmed == "" {
			// linha em branco OK
			continue
		} else {
			break
		}
	}
	return strings.Join(commentLines, "\n")
}

// ExtractDefines extrai as constantes #define do script
func ExtractDefines(script string) []string {
	matches := reDefine.FindAllStringSubmatch(script, -1)
	var names []string
	for _, m := range matches {
		names = append(names, m[1])
	}
	return names
}

// HasGlobalVarUsage detecta uso de variáveis globais $var$
func HasGlobalVarUsage(code string) bool {
	return reGlobalVar.MatchString(code)
}

// ExtractIncludes extrai os #include usados no código
func ExtractIncludes(code string) []string {
	matches := reInclude.FindAllStringSubmatch(code, -1)
	var includes []string
	for _, m := range matches {
		includes = append(includes, strings.TrimSpace(m[1]))
	}
	return includes
}
