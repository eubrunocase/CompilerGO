// Compilador Completo
//
// Uso:
//
//	go run main.go <arquivo.shei>          # compila e executa
//	go run main.go --tokens <arquivo.shei> # mostra tokens (debug)
//	go run main.go --ir <arquivo.shei>     # mostra código IR (debug)
package main

import (
	"fmt"
	"os"

	"compilador.com/m/ir"
	"compilador.com/m/lexer"
	"compilador.com/m/parser"
	"compilador.com/m/semantic"
	"compilador.com/m/vm"
)

func main() {

	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	showTokens := false
	showIR := false
	filename := ""

	for _, arg := range args {
		switch arg {
		case "--tokens":
			showTokens = true
		case "--ir":
			showIR = true
		default:
			filename = arg
		}
	}

	if filename == "" {
		fmt.Fprintln(os.Stderr, "erro: nenhum arquivo fornecido")
		printUsage()
		os.Exit(1)
	}

	// Lê o arquivo fonte
	source, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro ao ler arquivo: %v\n", err)
		os.Exit(1)
	}

	printHeader("SheiLang Compiler", filename)

	// Análise Léxica
	printPhase("1", "Análise Léxica")

	lex := lexer.New(string(source))
	tokens := lex.Tokenize()

	if lex.HasErrors() {
		for _, e := range lex.Errors {
			fmt.Fprintln(os.Stderr, "  ✗", e)
		}
		os.Exit(1)
	}
	fmt.Printf("  ✓ %d tokens reconhecidos\n", len(tokens))

	if showTokens {
		fmt.Println()
		for _, tok := range tokens {
			fmt.Println(" ", tok)
		}
	}

	// Análise Sintática
	printPhase("2", "Análise Sintática")

	par := parser.New(tokens)
	ast := par.Parse()

	if par.HasErrors() {
		for _, e := range par.Errors {
			fmt.Fprintln(os.Stderr, "  ✗", e)
		}
		os.Exit(1)
	}
	fmt.Printf("  ✓ AST construída com %d statements\n", len(ast.Stmts))

	//  Análise Semântica
	printPhase("3", "Análise Semântica")

	analyzer := semantic.New()
	analyzer.Analyze(ast)

	if analyzer.HasErrors() {
		for _, e := range analyzer.Errors {
			fmt.Fprintln(os.Stderr, "  ✗", e)
		}
		os.Exit(1)
	}
	fmt.Println("  ✓ Tipos e escopos verificados")

	//  Geração de IR (TAC)
	printPhase("4", "Geração de Código Intermediário (TAC)")

	gen := ir.New()
	gen.Generate(ast)

	fmt.Printf("  ✓ %d instruções TAC geradas\n", len(gen.Instructions))

	if showIR {
		fmt.Println()
		for _, instr := range gen.Instructions {
			fmt.Println(instr)
		}
	}

	// Execução na VM
	printPhase("5", "Execução na VM")
	fmt.Println()

	machine := vm.New(gen.Instructions)
	if err := machine.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "  ✗", err)
		os.Exit(1)
	}
}

// UTILITÁRIOS DE SAÍDA

func printHeader(title, filename string) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════╗")
	fmt.Printf("║  %-36s║\n", title)
	fmt.Printf("║  arquivo: %-26s║\n", truncate(filename, 26))
	fmt.Println("╚══════════════════════════════════════╝")
	fmt.Println()
}

func printPhase(num, name string) {
	fmt.Printf("\n[Fase %s] %s\n", num, name)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-max+3:]
}

func printUsage() {
	fmt.Println(`
Uso: sheilang [opções] <arquivo.shei>

Opções:
  --tokens   exibe os tokens reconhecidos pelo lexer
  --ir       exibe o código intermediário (TAC) gerado

Exemplos:
  go run main.go programa.shei
  go run main.go --tokens programa.shei
  go run main.go --ir programa.shei`)
}
