// Package semantic implementa a Análise Semântica da SheiLang.
//
// Responsabilidades:
//  1. Tabela de Símbolos: rastreia variáveis declaradas e seus tipos.
//  2. Verificação de Tipos (Type Checking): garante que operações
//     sejam aplicadas a tipos compatíveis.
//  3. Escopo: garante declaração prévia de variáveis antes do uso.
//
// Teoria aplicada: Visitor Pattern sobre a AST.
package semantic

import (
	"fmt"

	"compilador.com/m/parser"
)

// =========================================================
// 1. TABELA DE SÍMBOLOS
// =========================================================

// Symbol armazena informações sobre uma variável declarada.
type Symbol struct {
	Name  string
	VType string // "int" ou "bool"
	Line  int    // linha da declaração
}

// SymbolTable implementa uma tabela de símbolos com suporte a escopos.
// Cada escopo é uma camada (frame) empilhada sobre a anterior.
type SymbolTable struct {
	frames []map[string]*Symbol
}

func NewSymbolTable() *SymbolTable {
	st := &SymbolTable{}
	st.pushScope() // escopo global
	return st
}

// pushScope abre um novo escopo (ex: ao entrar em um bloco { }).
func (st *SymbolTable) pushScope() {
	st.frames = append(st.frames, make(map[string]*Symbol))
}

// popScope fecha o escopo mais recente (ex: ao sair de um bloco { }).
func (st *SymbolTable) popScope() {
	if len(st.frames) > 1 {
		st.frames = st.frames[:len(st.frames)-1]
	}
}

// Define declara uma variável no escopo atual.
// Retorna erro se já foi declarada neste escopo.
func (st *SymbolTable) Define(name, vtype string, line int) error {
	current := st.frames[len(st.frames)-1]
	if _, exists := current[name]; exists {
		return fmt.Errorf("variável %q já declarada neste escopo (linha %d)", name, line)
	}
	current[name] = &Symbol{Name: name, VType: vtype, Line: line}
	return nil
}

// Lookup busca uma variável do escopo mais interno para o mais externo.
// Retorna nil se não encontrada (variável não declarada).
func (st *SymbolTable) Lookup(name string) *Symbol {
	for i := len(st.frames) - 1; i >= 0; i-- {
		if sym, ok := st.frames[i][name]; ok {
			return sym
		}
	}
	return nil
}

// =========================================================
// 2. ANALISADOR SEMÂNTICO
// =========================================================

// Analyzer percorre a AST verificando semântica.
type Analyzer struct {
	table  *SymbolTable
	Errors []string
}

func New() *Analyzer {
	return &Analyzer{table: NewSymbolTable()}
}

func (a *Analyzer) addError(line int, format string, args ...any) {
	msg := fmt.Sprintf("erro semântico linha %d: "+format, append([]any{line}, args...)...)
	a.Errors = append(a.Errors, msg)
}

func (a *Analyzer) HasErrors() bool {
	return len(a.Errors) > 0
}

// Analyze é o ponto de entrada: analisa o programa inteiro.
func (a *Analyzer) Analyze(prog *parser.Program) {
	for _, stmt := range prog.Stmts {
		a.checkStmt(stmt)
	}
}

// =========================================================
// 3. VERIFICAÇÃO DE STATEMENTS
// =========================================================

func (a *Analyzer) checkStmt(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		a.checkVarDecl(s)
	case *parser.AssignStmt:
		a.checkAssignStmt(s)
	case *parser.IfStmt:
		a.checkIfStmt(s)
	case *parser.WhileStmt:
		a.checkWhileStmt(s)
	case *parser.PrintStmt:
		a.checkPrintStmt(s)
	case *parser.ReadStmt:
		a.checkReadStmt(s)
	case *parser.BlockStmt:
		a.checkBlock(s)
	}
}

// checkVarDecl verifica: tipo válido, inicialização compatível.
func (a *Analyzer) checkVarDecl(s *parser.VarDecl) {
	initType := a.checkExpr(s.Init)

	// Verifica compatibilidade entre tipo declarado e tipo da expressão
	if initType != "" && s.VType != "" && initType != s.VType {
		a.addError(s.Line,
			"variável %q declarada como %q, mas inicializada com valor do tipo %q",
			s.Name, s.VType, initType)
	}

	// Registra na tabela de símbolos
	if err := a.table.Define(s.Name, s.VType, s.Line); err != nil {
		a.addError(s.Line, "%s", err.Error())
	}
}

// checkAssignStmt verifica: variável declarada, tipo compatível.
func (a *Analyzer) checkAssignStmt(s *parser.AssignStmt) {
	sym := a.table.Lookup(s.Name)
	if sym == nil {
		a.addError(s.Line, "variável %q não declarada", s.Name)
		return
	}

	valType := a.checkExpr(s.Val)
	if valType != "" && valType != sym.VType {
		a.addError(s.Line,
			"atribuição inválida: variável %q é %q, mas expressão é %q",
			s.Name, sym.VType, valType)
	}
}

// checkIfStmt verifica: condição deve ser bool.
func (a *Analyzer) checkIfStmt(s *parser.IfStmt) {
	condType := a.checkExpr(s.Cond)
	if condType != "" && condType != "bool" {
		a.addError(s.Line, "condição do 'if' deve ser 'bool', encontrado %q", condType)
	}
	a.checkBlock(s.Then)
	if s.Else != nil {
		a.checkBlock(s.Else)
	}
}

// checkWhileStmt verifica: condição deve ser bool.
func (a *Analyzer) checkWhileStmt(s *parser.WhileStmt) {
	condType := a.checkExpr(s.Cond)
	if condType != "" && condType != "bool" {
		a.addError(s.Line, "condição do 'while' deve ser 'bool', encontrado %q", condType)
	}
	a.checkBlock(s.Body)
}

// checkPrintStmt aceita qualquer tipo (int ou bool).
func (a *Analyzer) checkPrintStmt(s *parser.PrintStmt) {
	a.checkExpr(s.Val)
}

// checkReadStmt verifica: variável declarada e do tipo int.
func (a *Analyzer) checkReadStmt(s *parser.ReadStmt) {
	sym := a.table.Lookup(s.Name)
	if sym == nil {
		a.addError(s.Line, "variável %q não declarada", s.Name)
		return
	}
	if sym.VType != "int" {
		a.addError(s.Line, "read() só suporta variáveis 'int', mas %q é %q", s.Name, sym.VType)
	}
}

// checkBlock abre/fecha escopo e verifica todos os statements internos.
func (a *Analyzer) checkBlock(b *parser.BlockStmt) {
	a.table.pushScope()
	for _, stmt := range b.Stmts {
		a.checkStmt(stmt)
	}
	a.table.popScope()
}

// =========================================================
// 4. VERIFICAÇÃO DE EXPRESSÕES (retorna o tipo resultante)
// =========================================================

// checkExpr verifica a expressão e retorna seu tipo ("int" ou "bool").
// Retorna "" quando o tipo não pode ser determinado (ex: erro anterior).
func (a *Analyzer) checkExpr(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.IntLiteral:
		return "int"

	case *parser.BoolLiteral:
		return "bool"

	case *parser.Identifier:
		sym := a.table.Lookup(e.Name)
		if sym == nil {
			a.addError(e.Line, "variável %q não declarada", e.Name)
			return ""
		}
		return sym.VType

	case *parser.UnaryExpr:
		return a.checkUnary(e)

	case *parser.BinaryExpr:
		return a.checkBinary(e)
	}
	return ""
}

// checkUnary verifica operadores unários.
//   - '-' exige 'int', produz 'int'
//   - '!' exige 'bool', produz 'bool'
func (a *Analyzer) checkUnary(e *parser.UnaryExpr) string {
	rightType := a.checkExpr(e.Right)
	switch e.Op {
	case "-":
		if rightType != "" && rightType != "int" {
			a.addError(e.Line, "operador '-' exige 'int', encontrado %q", rightType)
		}
		return "int"
	case "!":
		if rightType != "" && rightType != "bool" {
			a.addError(e.Line, "operador '!' exige 'bool', encontrado %q", rightType)
		}
		return "bool"
	}
	return ""
}

// checkBinary verifica operadores binários.
// Regras:
//   - Aritméticos (+,-,*,/)  → ambos int → resultado int
//   - Relacionais (<,>)      → ambos int → resultado bool
//   - Igualdade (==,!=)      → ambos mesmo tipo → resultado bool
func (a *Analyzer) checkBinary(e *parser.BinaryExpr) string {
	leftType := a.checkExpr(e.Left)
	rightType := a.checkExpr(e.Right)

	switch e.Op {
	case "+", "-", "*", "/":
		if leftType != "" && leftType != "int" {
			a.addError(e.Line, "operador %q exige 'int' à esquerda, encontrado %q", e.Op, leftType)
		}
		if rightType != "" && rightType != "int" {
			a.addError(e.Line, "operador %q exige 'int' à direita, encontrado %q", e.Op, rightType)
		}
		return "int"

	case "<", ">":
		if leftType != "" && leftType != "int" {
			a.addError(e.Line, "operador %q exige 'int' à esquerda, encontrado %q", e.Op, leftType)
		}
		if rightType != "" && rightType != "int" {
			a.addError(e.Line, "operador %q exige 'int' à direita, encontrado %q", e.Op, rightType)
		}
		return "bool"

	case "==", "!=":
		if leftType != "" && rightType != "" && leftType != rightType {
			a.addError(e.Line,
				"operador %q exige tipos iguais, encontrado %q e %q",
				e.Op, leftType, rightType)
		}
		return "bool"
	}
	return ""
}
