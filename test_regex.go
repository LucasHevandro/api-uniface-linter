package main

import (
	"fmt"
	"regexp"
)

func main() {
	reReturnErroExec := regexp.MustCompile(`(?i)^\s*return\s*<g_erroexec>`)
	fmt.Println("Return:", reReturnErroExec.MatchString("return &lt;g_erroexec&gt;"))
	fmt.Println("Return2:", reReturnErroExec.MatchString("return <g_erroexec>"))
}
