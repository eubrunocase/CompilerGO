package parser

import "fmt"

// Nós da AST

//   - Node      -> qualquer nó
//   - Statement -> nó que representa uma declaração (stmt)
//   - Expr      -> nó que representa uma expressão (expr)

// Node é a interface raiz de todos os nós da AST.
type Node interface {
	nodeType() string
}

// representa qualquer instrução da linguagem
type Statement interface {
	Node
	stmtNode()
}

// representa qualquer expressão que produz um valor
type Expr interface {
	Node
	exprNode()
}

// NÓS DE STATEMENT

// Raiz da AST = Lista de statements
type Program struct {
	Stmts []Statement
}

func (p *Program) nodeType() string { return "Program" }

// VarDecl -> var IDENT : type = expr ;
type VarDecl struct {
	Line  int
	Name  string
	VType string
	Init  Expr
}

func (v *VarDecl) nodeType() string { return "VarDecl" }
func (v *VarDecl) stmtNode()        {}

// AssignStmt -> IDENT = expr ;
type AssignStmt struct {
	Line int
	Name string
	Val  Expr
}

func (a *AssignStmt) nodeType() string { return "AssignStmt" }
func (a *AssignStmt) stmtNode()        {}

// IfStmt → if ( expr ) block [ else block ]
type IfStmt struct {
	Line int
	Cond Expr
	Then *BlockStmt
	Else *BlockStmt
}

func (i *IfStmt) nodeType() string { return "IfStmt" }
func (i *IfStmt) stmtNode()        {}

// WhileStmt → while ( expr ) block
type WhileStmt struct {
	Line int
	Cond Expr
	Body *BlockStmt
}

func (w *WhileStmt) nodeType() string { return "WhileStmt" }
func (w *WhileStmt) stmtNode()        {}

// PrintStmt → print ( expr ) ;
type PrintStmt struct {
	Line int
	Val  Expr
}

func (p *PrintStmt) nodeType() string { return "PrintStmt" }
func (p *PrintStmt) stmtNode()        {}

// ReadStmt → read ( IDENT ) ;
type ReadStmt struct {
	Line int
	Name string
}

func (r *ReadStmt) nodeType() string { return "ReadStmt" }
func (r *ReadStmt) stmtNode()        {}

// BlockStmt → { stmt* }
type BlockStmt struct {
	Stmts []Statement
}

func (b *BlockStmt) nodeType() string { return "BlockStmt" }
func (b *BlockStmt) stmtNode()        {}

//  NÓS DE EXPRESSÃO

// BinaryExpr → expr op expr
// Cobre: +, -, *, /, ==, !=, <, >
type BinaryExpr struct {
	Line  int
	Left  Expr
	Op    string // lexeme do operador
	Right Expr
}

func (b *BinaryExpr) nodeType() string { return "BinaryExpr" }
func (b *BinaryExpr) exprNode()        {}

// UnaryExpr → op expr
// Cobre: - (negativo), ! (not lógico)
type UnaryExpr struct {
	Line  int
	Op    string
	Right Expr
}

func (u *UnaryExpr) nodeType() string { return "UnaryExpr" }
func (u *UnaryExpr) exprNode()        {}

// IntLiteral → INT_LIT
type IntLiteral struct {
	Line  int
	Value int
}

func (i *IntLiteral) nodeType() string { return "IntLiteral" }
func (i *IntLiteral) exprNode()        {}

// BoolLiteral → true | false
type BoolLiteral struct {
	Line  int
	Value bool
}

func (b *BoolLiteral) nodeType() string { return "BoolLiteral" }
func (b *BoolLiteral) exprNode()        {}

// Identifier → IDENT
type Identifier struct {
	Line int
	Name string
}

func (i *Identifier) nodeType() string { return "Identifier" }
func (i *Identifier) exprNode()        {}

//  ERROS DE PARSE

// ParseError representa um erro sintático com linha e mensagem.
type ParseError struct {
	Line    int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("erro sintático linha %d: %s", e.Line, e.Message)
}
