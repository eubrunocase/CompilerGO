package parser

import (
	"fmt"
	"strconv"

	"compilador.com/m/lexer"
)

// =========================================================
// 3. ESTRUTURA DO PARSER
// =========================================================

// Parser consome tokens produzidos pelo Lexer e constrói a AST.
// Mantém um cursor 'pos' na lista de tokens.
type Parser struct {
	tokens []lexer.Token
	pos    int
	Errors []*ParseError
}

// New cria um novo Parser a partir da lista de tokens.
func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens}
}

// =========================================================
// 4. MÉTODOS AUXILIARES
// =========================================================

// current retorna o token na posição atual sem avançar.
func (p *Parser) current() lexer.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	// Retorna EOF sintético se ultrapassar o fim
	return lexer.Token{Type: lexer.TOKEN_EOF, Line: p.tokens[len(p.tokens)-1].Line}
}

// peek retorna o token seguinte (lookahead de 1) sem avançar.
func (p *Parser) peek() lexer.Token {
	next := p.pos + 1
	if next < len(p.tokens) {
		return p.tokens[next]
	}
	return lexer.Token{Type: lexer.TOKEN_EOF}
}

// advance consome e retorna o token atual.
func (p *Parser) advance() lexer.Token {
	tok := p.current()
	if tok.Type != lexer.TOKEN_EOF {
		p.pos++
	}
	return tok
}

// check retorna true se o token atual for do tipo esperado.
func (p *Parser) check(t lexer.TokenType) bool {
	return p.current().Type == t
}

// match consome o token atual se for do tipo esperado.
// Retorna true em caso de sucesso.
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

// expect consome o token atual se for do tipo esperado.
// Em caso de falha, registra erro e retorna token vazio.
func (p *Parser) expect(t lexer.TokenType) lexer.Token {
	if p.check(t) {
		return p.advance()
	}
	cur := p.current()
	p.addError(cur.Line, "esperado %q, encontrado %q (%s)", string(t), cur.Lexeme, cur.Type)
	return lexer.Token{}
}

// addError registra um erro sintático.
func (p *Parser) addError(line int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	p.Errors = append(p.Errors, &ParseError{Line: line, Message: msg})
}

// HasErrors retorna true se houve erros sintáticos.
func (p *Parser) HasErrors() bool {
	return len(p.Errors) > 0
}

// =========================================================
// 5. PONTO DE ENTRADA
// =========================================================

// Parse inicia o parsing e retorna a AST completa.
// Regra: program → stmt* EOF
func (p *Parser) Parse() *Program {
	prog := &Program{}
	for !p.check(lexer.TOKEN_EOF) {
		stmt := p.parseStmt()
		if stmt != nil {
			prog.Stmts = append(prog.Stmts, stmt)
		}
	}
	return prog
}

// =========================================================
// 6. PARSING DE STATEMENTS
// =========================================================

// parseStmt → var_decl | assign_stmt | if_stmt | while_stmt
//
//	| print_stmt | read_stmt | block
//
// Cada alternativa é escolhida pelo token atual (token de lookahead).
func (p *Parser) parseStmt() Statement {
	cur := p.current()

	switch cur.Type {
	case lexer.TOKEN_VAR:
		return p.parseVarDecl()
	case lexer.TOKEN_IF:
		return p.parseIfStmt()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStmt()
	case lexer.TOKEN_PRINT:
		return p.parsePrintStmt()
	case lexer.TOKEN_READ:
		return p.parseReadStmt()
	case lexer.TOKEN_LBRACE:
		return p.parseBlock()
	case lexer.TOKEN_IDENT:
		// Identificador no início → atribuição
		return p.parseAssignStmt()
	default:
		p.addError(cur.Line, "statement inválido: token inesperado %q", cur.Lexeme)
		p.advance() // consome token problemático para evitar loop infinito
		return nil
	}
}

// parseVarDecl → var IDENT : type = expr ;
func (p *Parser) parseVarDecl() *VarDecl {
	tok := p.expect(lexer.TOKEN_VAR)

	nameTok := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_COLON)

	// tipo: int | bool
	typeTok := p.current()
	if typeTok.Type != lexer.TOKEN_INT && typeTok.Type != lexer.TOKEN_BOOL {
		p.addError(typeTok.Line, "tipo inválido %q: esperado 'int' ou 'bool'", typeTok.Lexeme)
	} else {
		p.advance()
	}

	p.expect(lexer.TOKEN_ASSIGN)
	init := p.parseExpr()
	p.expect(lexer.TOKEN_SEMICOLON)

	return &VarDecl{
		Line:  tok.Line,
		Name:  nameTok.Lexeme,
		VType: typeTok.Lexeme,
		Init:  init,
	}
}

// parseAssignStmt → IDENT = expr ;
func (p *Parser) parseAssignStmt() *AssignStmt {
	nameTok := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_ASSIGN)
	val := p.parseExpr()
	p.expect(lexer.TOKEN_SEMICOLON)

	return &AssignStmt{
		Line: nameTok.Line,
		Name: nameTok.Lexeme,
		Val:  val,
	}
}

// parseIfStmt → if ( expr ) block [ else block ]
func (p *Parser) parseIfStmt() *IfStmt {
	tok := p.expect(lexer.TOKEN_IF)
	p.expect(lexer.TOKEN_LPAREN)
	cond := p.parseExpr()
	p.expect(lexer.TOKEN_RPAREN)
	then := p.parseBlock()

	var elseBlock *BlockStmt
	if p.match(lexer.TOKEN_ELSE) {
		elseBlock = p.parseBlock()
	}

	return &IfStmt{
		Line: tok.Line,
		Cond: cond,
		Then: then,
		Else: elseBlock,
	}
}

// parseWhileStmt → while ( expr ) block
func (p *Parser) parseWhileStmt() *WhileStmt {
	tok := p.expect(lexer.TOKEN_WHILE)
	p.expect(lexer.TOKEN_LPAREN)
	cond := p.parseExpr()
	p.expect(lexer.TOKEN_RPAREN)
	body := p.parseBlock()

	return &WhileStmt{
		Line: tok.Line,
		Cond: cond,
		Body: body,
	}
}

// parsePrintStmt → print ( expr ) ;
func (p *Parser) parsePrintStmt() *PrintStmt {
	tok := p.expect(lexer.TOKEN_PRINT)
	p.expect(lexer.TOKEN_LPAREN)
	val := p.parseExpr()
	p.expect(lexer.TOKEN_RPAREN)
	p.expect(lexer.TOKEN_SEMICOLON)

	return &PrintStmt{Line: tok.Line, Val: val}
}

// parseReadStmt → read ( IDENT ) ;
func (p *Parser) parseReadStmt() *ReadStmt {
	tok := p.expect(lexer.TOKEN_READ)
	p.expect(lexer.TOKEN_LPAREN)
	nameTok := p.expect(lexer.TOKEN_IDENT)
	p.expect(lexer.TOKEN_RPAREN)
	p.expect(lexer.TOKEN_SEMICOLON)

	return &ReadStmt{Line: tok.Line, Name: nameTok.Lexeme}
}

// parseBlock → { stmt* }
func (p *Parser) parseBlock() *BlockStmt {
	p.expect(lexer.TOKEN_LBRACE)
	block := &BlockStmt{}
	for !p.check(lexer.TOKEN_RBRACE) && !p.check(lexer.TOKEN_EOF) {
		stmt := p.parseStmt()
		if stmt != nil {
			block.Stmts = append(block.Stmts, stmt)
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return block
}

// =========================================================
// 7. PARSING DE EXPRESSÕES
// =========================================================
//
// A hierarquia de métodos abaixo codifica a precedência:
//
//   parseExpr        (menor precedência)
//     └─ parseEquality        (==, !=)
//          └─ parseComparison      (<, >)
//               └─ parseTerm           (+, -)
//                    └─ parseFactor         (*, /)
//                         └─ parseUnary          (-, !)
//                              └─ parsePrimary    (maior precedência)
//
// Essa estrutura garante que * tem maior precedência que +,
// que tem maior que <, e assim por diante.

// parseExpr é o ponto de entrada das expressões.
func (p *Parser) parseExpr() Expr {
	return p.parseEquality()
}

// parseEquality → comparison ( ( == | != ) comparison )*
func (p *Parser) parseEquality() Expr {
	left := p.parseComparison()
	for p.check(lexer.TOKEN_EQ) || p.check(lexer.TOKEN_NEQ) {
		op := p.advance()
		right := p.parseComparison()
		left = &BinaryExpr{Line: op.Line, Left: left, Op: op.Lexeme, Right: right}
	}
	return left
}

// parseComparison → term ( ( < | > ) term )*
func (p *Parser) parseComparison() Expr {
	left := p.parseTerm()
	for p.check(lexer.TOKEN_LT) || p.check(lexer.TOKEN_GT) {
		op := p.advance()
		right := p.parseTerm()
		left = &BinaryExpr{Line: op.Line, Left: left, Op: op.Lexeme, Right: right}
	}
	return left
}

// parseTerm → factor ( ( + | - ) factor )*
func (p *Parser) parseTerm() Expr {
	left := p.parseFactor()
	for p.check(lexer.TOKEN_PLUS) || p.check(lexer.TOKEN_MINUS) {
		op := p.advance()
		right := p.parseFactor()
		left = &BinaryExpr{Line: op.Line, Left: left, Op: op.Lexeme, Right: right}
	}
	return left
}

// parseFactor → unary ( ( * | / ) unary )*
func (p *Parser) parseFactor() Expr {
	left := p.parseUnary()
	for p.check(lexer.TOKEN_STAR) || p.check(lexer.TOKEN_SLASH) {
		op := p.advance()
		right := p.parseUnary()
		left = &BinaryExpr{Line: op.Line, Left: left, Op: op.Lexeme, Right: right}
	}
	return left
}

// parseUnary → ( - | ! ) unary | primary
func (p *Parser) parseUnary() Expr {
	if p.check(lexer.TOKEN_MINUS) || p.check(lexer.TOKEN_BANG) {
		op := p.advance()
		right := p.parseUnary()
		return &UnaryExpr{Line: op.Line, Op: op.Lexeme, Right: right}
	}
	return p.parsePrimary()
}

// parsePrimary → INT_LIT | true | false | IDENT | ( expr )
func (p *Parser) parsePrimary() Expr {
	cur := p.current()

	switch cur.Type {
	case lexer.TOKEN_INT_LIT:
		p.advance()
		val, _ := strconv.Atoi(cur.Lexeme)
		return &IntLiteral{Line: cur.Line, Value: val}

	case lexer.TOKEN_TRUE:
		p.advance()
		return &BoolLiteral{Line: cur.Line, Value: true}

	case lexer.TOKEN_FALSE:
		p.advance()
		return &BoolLiteral{Line: cur.Line, Value: false}

	case lexer.TOKEN_IDENT:
		p.advance()
		return &Identifier{Line: cur.Line, Name: cur.Lexeme}

	case lexer.TOKEN_LPAREN:
		p.advance() // consome '('
		expr := p.parseExpr()
		p.expect(lexer.TOKEN_RPAREN)
		return expr
	}

	p.addError(cur.Line, "expressão inválida: token inesperado %q", cur.Lexeme)
	p.advance()
	return &IntLiteral{Line: cur.Line, Value: 0} // nó de recuperação
}
