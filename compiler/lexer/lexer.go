package lexer

import "fmt"

type TokenType string

const (
	// --- Literais ---
	TOKEN_INT_LIT TokenType = "INT_LIT"
	TOKEN_TRUE    TokenType = "TRUE"
	TOKEN_FALSE   TokenType = "FALSE"

	// --- Identificador ---
	TOKEN_IDENT TokenType = "IDENT" // nome de variável, função, etc

	// --- Palavras reservadas ---
	TOKEN_VAR   TokenType = "VAR"
	TOKEN_INT   TokenType = "INT"
	TOKEN_BOOL  TokenType = "BOOL"
	TOKEN_IF    TokenType = "IF"
	TOKEN_ELSE  TokenType = "ELSE"
	TOKEN_WHILE TokenType = "WHILE"
	TOKEN_PRINT TokenType = "PRINT"
	TOKEN_READ  TokenType = "READ"

	// --- Operadores aritméticos ---
	TOKEN_PLUS  TokenType = "+"
	TOKEN_MINUS TokenType = "-"
	TOKEN_STAR  TokenType = "*"
	TOKEN_SLASH TokenType = "/"

	// --- Operadores de relacionais ---
	TOKEN_EQ  TokenType = "=="
	TOKEN_NEQ TokenType = "!="
	TOKEN_LT  TokenType = "<"
	TOKEN_GT  TokenType = ">"

	// --- Operadore lógico unário ---
	TOKEN_BANG TokenType = "!"

	// --- Pontuação ---
	TOKEN_ASSIGN    TokenType = "="
	TOKEN_SEMICOLON TokenType = ";"
	TOKEN_COLON     TokenType = ":"
	TOKEN_LPAREN    TokenType = "("
	TOKEN_RPAREN    TokenType = ")"
	TOKEN_LBRACE    TokenType = "{"
	TOKEN_RBRACE    TokenType = "}"

	// --- Controle ---
	TOKEN_EOF     TokenType = "EOF"
	TOKEN_ILLEGAL TokenType = "ILLEGAL"
)

// estrutura de dados para mapeamento de palavras reservas
var keywords = map[string]TokenType{
	"var":   TOKEN_VAR,
	"int":   TOKEN_INT,
	"bool":  TOKEN_BOOL,
	"true":  TOKEN_TRUE,
	"false": TOKEN_FALSE,
	"if":    TOKEN_IF,
	"else":  TOKEN_ELSE,
	"while": TOKEN_WHILE,
	"print": TOKEN_PRINT,
	"read":  TOKEN_READ,
}

// Estrutura do Token
type Token struct {
	Type   TokenType
	Lexeme string // texto original do código-fonte
	Line   int    // linha de encontro do token para mensagens de erro
}

func (t Token) String() string {
	return fmt.Sprintf("Token{%-10s %-12q linha: %d}", t.Type, t.Lexeme, t.Line)
}

// --- Estrutura do Lexer ---
// Mantem dois cursores para rastrear o início e a posição atual do lexeme, além de um contador de linha para mensagens de erro.
type Lexer struct {
	source  string
	start   int // inicio do lexeme
	current int // posição atual de leitura
	line    int // linha atual para mensagens de erro
	Errors  []string
}

// Função para criar um novo lexer a partir do código-fonte
func New(source string) *Lexer {
	return &Lexer{
		source: source,
		line:   1,
	}
}

// Funções auxiliares

// isAtEnd retorna true quando todos os caracteres foram consumidos
func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

// advance() consome o caráter atual e avança o cursor, retornando o caractere consumido
func (l *Lexer) advance() byte {
	ch := l.source[l.current]
	l.current++
	return ch
}

// peek retorna o caracte atual sem consumir, ou 0 se estiver no final do arquivo (lookahead de 1)
func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

// ppekNext retorna o próximo caractere sem consumir, ou 0 se estiver no final do arquivo (lookahead de 2)
func (l *Lexer) peekNext() byte {
	if l.current+1 >= len(l.source) {
		return 0
	}
	return l.source[l.current+1]
}

// match consome o próximo caractere se ele for igual a expected
func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.current++
	return true
}

// currentLexeme retorna o trecho do código-fonte do início ao cursor atual.
func (l *Lexer) currentLexeme() string {
	return l.source[l.start:l.current]
}

// makeToken cria um novo token com o lexeme acumulado desde o start
func (l *Lexer) makeToken(t TokenType) Token {
	return Token{Type: t, Lexeme: l.currentLexeme(), Line: l.line}
}

// errorf registra um erro léxico e retorna um token ILLEGAL.
func (l *Lexer) errorf(format string, args ...any) Token {
	msg := fmt.Sprintf("linha %d: "+format, append([]any{l.line}, args...)...)
	l.Errors = append(l.Errors, msg)
	return Token{Type: TOKEN_ILLEGAL, Lexeme: l.currentLexeme(), Line: l.line}
}

// Funções de reconhecimento

// skipWhitespaceAndComments avança o cursor ignorando:
//   - espaços, tabs, retornos de carro
//   - quebras de linha (incrementa l.line)
//   - comentários de linha (// até \n)
func (l *Lexer) skipWhitespaceAndComments() {
	for !l.isAtEnd() {
		ch := l.peek()
		switch ch {
		case ' ', '\t', '\r':
			l.advance() // ignora espaços em branco
		case '\n':
			l.line++
			l.advance()
		case '/':
			if l.peekNext() == '/' {
				for !l.isAtEnd() && l.peek() != '\n' {
					l.advance()
				}
			} else {
				return
			}
		default:
			return
		}
	}
}

// readNumber reconhece literais inteiros: [0-9]+
func (l *Lexer) readNumber() Token {
	for !l.isAtEnd() && isDigit(l.peek()) {
		l.advance()
	}
	return l.makeToken(TOKEN_INT_LIT)
}

// readIdentOrKeyword reconhece identificadores e palavras reservadas.
func (l *Lexer) readIdentOrKeyword() Token {
	for !l.isAtEnd() && isAlphaNumeric(l.peek()) {
		l.advance()
	}
	lexeme := l.currentLexeme()
	// Verifica se é palavra reservada
	if tt, ok := keywords[lexeme]; ok {
		return l.makeToken(tt)
	}
	return l.makeToken(TOKEN_IDENT)
}

// PRÓXIMO TOKEN

// nextToken lê o próximo token do código-fonte.
// Cada case representa uma transição do AFD.
func (l *Lexer) nextToken() Token {
	l.skipWhitespaceAndComments()

	if l.isAtEnd() {
		return Token{Type: TOKEN_EOF, Lexeme: "", Line: l.line}
	}

	l.start = l.current
	ch := l.advance()

	// Literais numéricos
	if isDigit(ch) {
		return l.readNumber()
	}

	// Identificadores e palavras reservadas
	if isAlpha(ch) {
		return l.readIdentOrKeyword()
	}
	// Operadores e pontuação
	switch ch {
	case '+':
		return l.makeToken(TOKEN_PLUS)
	case '-':
		return l.makeToken(TOKEN_MINUS)
	case '*':
		return l.makeToken(TOKEN_STAR)
	case '/':
		return l.makeToken(TOKEN_SLASH)
	case '<':
		return l.makeToken(TOKEN_LT)
	case '>':
		return l.makeToken(TOKEN_GT)
	case ';':
		return l.makeToken(TOKEN_SEMICOLON)
	case ':':
		return l.makeToken(TOKEN_COLON)
	case '(':
		return l.makeToken(TOKEN_LPAREN)
	case ')':
		return l.makeToken(TOKEN_RPAREN)
	case '{':
		return l.makeToken(TOKEN_LBRACE)
	case '}':
		return l.makeToken(TOKEN_RBRACE)

	// '=' ou '=='
	case '=':
		if l.match('=') {
			return l.makeToken(TOKEN_EQ)
		}
		return l.makeToken(TOKEN_ASSIGN)

	// '!' ou '!='
	case '!':
		if l.match('=') {
			return l.makeToken(TOKEN_NEQ)
		}
		return l.makeToken(TOKEN_BANG)
	}
	return l.errorf("caractere inesperado: %q", ch)
}

// INTERFACE PÚBLICA

// Tokenize percorre todo o código-fonte e retorna a lista
// completa de tokens (incluindo o EOF final).
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.nextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF || tok.Type == TOKEN_ILLEGAL {
			break
		}
	}
	return tokens
}

// Retorna true caso tenham erros na analise léxica
func (l *Lexer) HasErrors() bool {
	return len(l.Errors) > 0
}

// FUNÇÕES AUXILIARES DE CLASSIFICAÇÃO DE CARACTERES

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_'
}

func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}
