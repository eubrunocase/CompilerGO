// Package ir implementa a Geração de Código Intermediário da SheiLang.
//
// Responsabilidade: traduzir a AST para Código de Três Endereços (TAC).
//
// O TAC é uma representação linear de instruções simples no formato:
//
//	result = operand1 op operand2
//
// Ele é independente de máquina e facilita a geração de código final.
package ir

import (
	"fmt"

	"compilador.com/m/parser"
)

//  INSTRUÇÕES DO TAC

// OpCode representa a operação de uma instrução TAC.
type OpCode string

const (
	// Atribuição simples: result = arg1
	OP_ASSIGN OpCode = "ASSIGN"

	// Operações binárias: result = arg1 op arg2
	OP_ADD OpCode = "ADD"
	OP_SUB OpCode = "SUB"
	OP_MUL OpCode = "MUL"
	OP_DIV OpCode = "DIV"
	OP_EQ  OpCode = "EQ"
	OP_NEQ OpCode = "NEQ"
	OP_LT  OpCode = "LT"
	OP_GT  OpCode = "GT"

	// Operações unárias: result = op arg1
	OP_NEG OpCode = "NEG" // negativo aritmético
	OP_NOT OpCode = "NOT" // negação lógica

	// Desvios (controle de fluxo)
	OP_JUMP  OpCode = "JUMP"  // desvio incondicional
	OP_JUMPF OpCode = "JUMPF" // desvio se arg1 == false (0)

	// Label (ponto de destino de desvios)
	OP_LABEL OpCode = "LABEL"

	// E/S
	OP_PRINT OpCode = "PRINT"
	OP_READ  OpCode = "READ"
)

// Instruction representa uma instrução TAC.
//
// Formato geral:
//
//	Result = Arg1 Op Arg2
//
// Exemplos:
//
//	t0 = a + b        → {OP_ADD, "t0", "a", "b"}
//	t1 = t0           → {OP_ASSIGN, "t1", "t0", ""}
//	JUMPF t1 L2       → {OP_JUMPF, "", "t1", "L2"}
//	L1:               → {OP_LABEL, "L1", "", ""}
//	PRINT x           → {OP_PRINT, "", "x", ""}
type Instruction struct {
	Op     OpCode
	Result string // destino da operação
	Arg1   string // primeiro operando
	Arg2   string // segundo operando (pode ser vazio)
}

func (i Instruction) String() string {
	switch i.Op {
	case OP_LABEL:
		return fmt.Sprintf("%s:", i.Result)
	case OP_JUMP:
		return fmt.Sprintf("  JUMP %s", i.Arg1)
	case OP_JUMPF:
		return fmt.Sprintf("  JUMPF %s %s", i.Arg1, i.Arg2)
	case OP_PRINT:
		return fmt.Sprintf("  PRINT %s", i.Arg1)
	case OP_READ:
		return fmt.Sprintf("  READ %s", i.Arg1)
	case OP_ASSIGN:
		return fmt.Sprintf("  %s = %s", i.Result, i.Arg1)
	case OP_NEG:
		return fmt.Sprintf("  %s = -%s", i.Result, i.Arg1)
	case OP_NOT:
		return fmt.Sprintf("  %s = !%s", i.Result, i.Arg1)
	default:
		return fmt.Sprintf("  %s = %s %s %s", i.Result, i.Arg1, i.Op, i.Arg2)
	}
}

// =========================================================
// 2. GERADOR DE IR
// =========================================================

// Generator traduz a AST em TAC.
type Generator struct {
	Instructions []Instruction
	tempCount    int // contador de temporários: t0, t1, t2, ...
	labelCount   int // contador de labels: L0, L1, L2, ...
}

func New() *Generator {
	return &Generator{}
}

// newTemp gera um novo nome de variável temporária.
func (g *Generator) newTemp() string {
	name := fmt.Sprintf("t%d", g.tempCount)
	g.tempCount++
	return name
}

// newLabel gera um novo nome de label.
func (g *Generator) newLabel() string {
	name := fmt.Sprintf("L%d", g.labelCount)
	g.labelCount++
	return name
}

// emit adiciona uma instrução ao programa TAC.
func (g *Generator) emit(op OpCode, result, arg1, arg2 string) {
	g.Instructions = append(g.Instructions, Instruction{op, result, arg1, arg2})
}

// =========================================================
// 3. GERAÇÃO A PARTIR DO PROGRAMA
// =========================================================

// Generate é o ponto de entrada: gera TAC para o programa inteiro.
func (g *Generator) Generate(prog *parser.Program) {
	for _, stmt := range prog.Stmts {
		g.genStmt(stmt)
	}
}

// =========================================================
// 4. GERAÇÃO DE STATEMENTS
// =========================================================

func (g *Generator) genStmt(stmt parser.Statement) {
	switch s := stmt.(type) {
	case *parser.VarDecl:
		g.genVarDecl(s)
	case *parser.AssignStmt:
		g.genAssignStmt(s)
	case *parser.IfStmt:
		g.genIfStmt(s)
	case *parser.WhileStmt:
		g.genWhileStmt(s)
	case *parser.PrintStmt:
		g.genPrintStmt(s)
	case *parser.ReadStmt:
		g.genReadStmt(s)
	case *parser.BlockStmt:
		g.genBlock(s)
	}
}

// genVarDecl → result = expr
func (g *Generator) genVarDecl(s *parser.VarDecl) {
	src := g.genExpr(s.Init)
	g.emit(OP_ASSIGN, s.Name, src, "")
}

// genAssignStmt → result = expr
func (g *Generator) genAssignStmt(s *parser.AssignStmt) {
	src := g.genExpr(s.Val)
	g.emit(OP_ASSIGN, s.Name, src, "")
}

// genIfStmt implementa o padrão TAC para if-else:
//
//	  <cond>
//	  JUMPF cond Lelse
//	  <then>
//	  JUMP Lend
//	Lelse:
//	  <else>      (opcional)
//	Lend:
func (g *Generator) genIfStmt(s *parser.IfStmt) {
	cond := g.genExpr(s.Cond)

	lElse := g.newLabel()
	lEnd := g.newLabel()

	g.emit(OP_JUMPF, "", cond, lElse)
	g.genBlock(s.Then)
	g.emit(OP_JUMP, "", lEnd, "")

	g.emit(OP_LABEL, lElse, "", "")
	if s.Else != nil {
		g.genBlock(s.Else)
	}
	g.emit(OP_LABEL, lEnd, "", "")
}

// genWhileStmt implementa o padrão TAC para while:
//
//	Lstart:
//	  <cond>
//	  JUMPF cond Lend
//	  <body>
//	  JUMP Lstart
//	Lend:
func (g *Generator) genWhileStmt(s *parser.WhileStmt) {
	lStart := g.newLabel()
	lEnd := g.newLabel()

	g.emit(OP_LABEL, lStart, "", "")
	cond := g.genExpr(s.Cond)
	g.emit(OP_JUMPF, "", cond, lEnd)
	g.genBlock(s.Body)
	g.emit(OP_JUMP, "", lStart, "")
	g.emit(OP_LABEL, lEnd, "", "")
}

// genPrintStmt → PRINT expr
func (g *Generator) genPrintStmt(s *parser.PrintStmt) {
	val := g.genExpr(s.Val)
	g.emit(OP_PRINT, "", val, "")
}

// genReadStmt → READ ident
func (g *Generator) genReadStmt(s *parser.ReadStmt) {
	g.emit(OP_READ, "", s.Name, "")
}

func (g *Generator) genBlock(b *parser.BlockStmt) {
	for _, stmt := range b.Stmts {
		g.genStmt(stmt)
	}
}

// =========================================================
// 5. GERAÇÃO DE EXPRESSÕES
// =========================================================

// genExpr gera o TAC para uma expressão e retorna o nome
// do operando onde o resultado foi armazenado (variável ou temporário).
func (g *Generator) genExpr(expr parser.Expr) string {
	switch e := expr.(type) {
	case *parser.IntLiteral:
		return fmt.Sprintf("%d", e.Value)

	case *parser.BoolLiteral:
		if e.Value {
			return "1"
		}
		return "0"

	case *parser.Identifier:
		return e.Name

	case *parser.UnaryExpr:
		return g.genUnary(e)

	case *parser.BinaryExpr:
		return g.genBinary(e)
	}
	return "0"
}

// genUnary → t = op arg
func (g *Generator) genUnary(e *parser.UnaryExpr) string {
	arg := g.genExpr(e.Right)
	t := g.newTemp()
	switch e.Op {
	case "-":
		g.emit(OP_NEG, t, arg, "")
	case "!":
		g.emit(OP_NOT, t, arg, "")
	}
	return t
}

// genBinary → t = arg1 op arg2
func (g *Generator) genBinary(e *parser.BinaryExpr) string {
	left := g.genExpr(e.Left)
	right := g.genExpr(e.Right)
	t := g.newTemp()

	opMap := map[string]OpCode{
		"+": OP_ADD, "-": OP_SUB, "*": OP_MUL, "/": OP_DIV,
		"==": OP_EQ, "!=": OP_NEQ, "<": OP_LT, ">": OP_GT,
	}
	if op, ok := opMap[e.Op]; ok {
		g.emit(op, t, left, right)
	}
	return t
}
