# SheiLang Compiler

Compilador completo para a linguagem **SheiLang**, implementado em Go. O projeto cobre todas as fases clássicas de compilação da análise léxica até a execução em máquina virtual, servindo como referência didática para o estudo de compiladores.

---

## Sumário

- [Visão Geral](#visão-geral)
- [Arquitetura do Compilador](#arquitetura-do-compilador)
- [A Linguagem SheiLang](#a-linguagem-sheilang)
- [Pré-requisitos](#pré-requisitos)
- [Como Rodar](#como-rodar)
- [Opções de Debug](#opções-de-debug)
- [Exemplo de Uso](#exemplo-de-uso)
- [Estrutura do Projeto](#estrutura-do-projeto)
- [Pipeline Detalhado](#pipeline-detalhado)

---

## Visão Geral

O compilador processa arquivos `.shei` em cinco fases sequenciais:

```
código-fonte (.shei)
       │
       ▼
 [1] Análise Léxica     → lista de tokens
       │
       ▼
 [2] Análise Sintática  → AST (Árvore Sintática Abstrata)
       │
       ▼
 [3] Análise Semântica  → verificação de tipos e escopos
       │
       ▼
 [4] Geração de IR/TAC  → instruções de três endereços
       │
       ▼
 [5] Execução na VM     → resultado no terminal
```

---

## Arquitetura do Compilador

| Fase | Pacote | Arquivo | Responsabilidade |
|------|--------|---------|-----------------|
| 1 – Léxica | `lexer` | `lexer/lexer.go` | Tokenização do código-fonte |
| 2 – Sintática | `parser` | `parser/parser.go` + `parser/ast.go` | Construção da AST por descida recursiva |
| 3 – Semântica | `semantic` | `semantic/semantic.go` | Tabela de símbolos e verificação de tipos |
| 4 – IR | `ir` | `ir/ir.go` | Geração de Código de Três Endereços (TAC) |
| 5 – VM | `vm` | `vm/vm.go` | Máquina virtual que executa as instruções TAC |
| Ponto de entrada | — | `main.go` | Orquestra as cinco fases e a CLI |

---

## A Linguagem SheiLang

SheiLang é uma linguagem imperativa tipada com dois tipos primitivos e suporte a fluxo de controle.

### Tipos

| Tipo | Descrição |
|------|-----------|
| `int` | Inteiros (ex: `42`, `-7`) |
| `bool` | Booleanos (`true` ou `false`) |

### Sintaxe

```
// Declaração de variável
var nome: tipo = expressão;

// Atribuição
nome = expressão;

// Condicional
if (condição) {
    ...
} else {
    ...
}

// Laço
while (condição) {
    ...
}

// Saída
print(expressão);

// Entrada (somente int)
read(variavel);

// Comentário de linha
// este é um comentário
```

### Operadores

| Categoria | Operadores |
|-----------|-----------|
| Aritméticos | `+`  `-`  `*`  `/` |
| Relacionais | `<`  `>` |
| Igualdade | `==`  `!=` |
| Lógico unário | `!` |
| Negação aritmética | `-` (unário) |

### Precedência (maior para menor)

```
primário  (literais, identificadores, (expr))
  unário  (-, !)
  fator   (*, /)
  termo   (+, -)
  comparação (<, >)
  igualdade  (==, !=)
```

---

## Pré-requisitos

- **Go 1.21** ou superior  
  Verifique com: `go version`

Se não tiver o Go instalado, baixe em: https://go.dev/dl/

---

## Como Rodar

### 1. Clone o repositório

```bash
git clone <url-do-repositorio>
cd CompilerGO-main/compiler
```

### 2. Execute um arquivo `.shei`

```bash
go run main.go <arquivo.shei>
```

**Exemplo com o arquivo incluso:**

```bash
go run main.go exemplo.shei
```

Saída esperada:

```
╔══════════════════════════════════════╗
║  SheiLang Compiler                   ║
║  arquivo: exemplo.shei               ║
╚══════════════════════════════════════╝

[Fase 1] Análise Léxica
  ✓ 67 tokens reconhecidos

[Fase 2] Análise Sintática
  ✓ AST construída com 8 statements

[Fase 3] Análise Semântica
  ✓ Tipos e escopos verificados

[Fase 4] Geração de Código Intermediário (TAC)
  ✓ 38 instruções TAC geradas

[Fase 5] Execução na VM

120
1
0
1
2
3
4
```

### 3. Compilar o binário (opcional)

Para gerar um executável em vez de usar `go run`:

```bash
go build -o sheilang main.go
./sheilang exemplo.shei
```

---

## Opções de Debug

O compilador oferece duas flags para inspecionar fases internas:

### `--tokens` — exibe os tokens do Lexer

```bash
go run main.go --tokens exemplo.shei
```

Mostra cada token reconhecido com seu tipo, lexeme e linha:

```
  Token{INT_LIT    "5"          linha: 1}
  Token{IDENT      "n"          linha: 1}
  Token{VAR        "var"        linha: 1}
  ...
```

### `--ir` — exibe o código intermediário (TAC)

```bash
go run main.go --ir exemplo.shei
```

Mostra as instruções de Três Endereços geradas antes da execução:

```
  n = 5
  resultado = 1
  i = 1
L0:
  t0 = i LT n
  JUMPF t0 L1
  i = i + 1
  ...
```

As duas flags podem ser combinadas:

```bash
go run main.go --tokens --ir exemplo.shei
```

---

## Exemplo de Uso

O arquivo `exemplo.shei` incluído no projeto demonstra as principais funcionalidades da linguagem:

```sheilang
// Calcula fatorial de 5
var n: int = 5;
var resultado: int = 1;
var i: int = 1;

while (i < n) {
    i = i + 1;
    resultado = resultado * i;
}

print(resultado);  // esperado: 120

// Testa if-else e booleanos
var x: int = 10;
var dobro: int = x * 2;

if (dobro > 15) {
    print(1);   // esperado: 1
} else {
    print(0);
}

// Conta de 0 a 4
var contador: int = 0;
while (contador < 5) {
    print(contador);   // esperado: 0 1 2 3 4
    contador = contador + 1;
}
```

### Usando `read()` para entrada interativa

```sheilang
var n: int = 0;
read(n);

var dobro: int = n * 2;
print(dobro);
```

Ao executar, o programa aguarda você digitar um número:

```
entrada> 7
14
```

---

## Estrutura do Projeto

```
compiler/
├── main.go              # Ponto de entrada e orquestração das fases
├── go.mod               # Módulo Go: compilador.com/m
├── exemplo.shei         # Programa de exemplo SheiLang
│
├── lexer/
│   └── lexer.go         # Análise léxica: tokeniza o código-fonte
│
├── parser/
│   ├── ast.go           # Definição dos nós da AST
│   └── parser.go        # Parser por descida recursiva
│
├── semantic/
│   └── semantic.go      # Tabela de símbolos e verificação de tipos
│
├── ir/
│   └── ir.go            # Geração de código TAC (Três Endereços)
│
└── vm/
    └── vm.go            # Máquina virtual: executa as instruções TAC
```

---

## Pipeline Detalhado

### Fase 1 — Análise Léxica (`lexer`)

O `Lexer` percorre o código-fonte caractere por caractere usando dois cursores (`start` e `current`) e um contador de linha. Implementa um AFD (Autômato Finito Determinístico) reconhecendo:

- **Literais inteiros**: sequências de `[0-9]+`
- **Identificadores e palavras reservadas**: `var`, `int`, `bool`, `if`, `else`, `while`, `print`, `read`, `true`, `false`
- **Operadores**: `+`, `-`, `*`, `/`, `<`, `>`, `==`, `!=`, `!`, `=`
- **Pontuação**: `;`, `:`, `(`, `)`, `{`, `}`
- **Comentários de linha**: `//` até fim da linha (descartados)
- **Lookahead de 2 caracteres** para diferenciar `=` de `==`, `!` de `!=`

Erros léxicos (caracteres inválidos) são coletados e exibidos antes de encerrar.

---

### Fase 2 — Análise Sintática (`parser`)

O `Parser` implementa um **parser por descida recursiva** consumindo a lista de tokens. Produz uma **AST** (Árvore Sintática Abstrata) com os seguintes nós:

**Statements:** `VarDecl`, `AssignStmt`, `IfStmt`, `WhileStmt`, `PrintStmt`, `ReadStmt`, `BlockStmt`

**Expressões:** `BinaryExpr`, `UnaryExpr`, `IntLiteral`, `BoolLiteral`, `Identifier`

A precedência de operadores é codificada pela hierarquia de funções de parsing (da menor para a maior precedência):

```
parseExpr → parseEquality → parseComparison → parseTerm
         → parseFactor  → parseUnary      → parsePrimary
```

---

### Fase 3 — Análise Semântica (`semantic`)

O `Analyzer` percorre a AST verificando:

- **Tabela de símbolos com escopos**: cada bloco `{ }` abre um novo frame; variáveis são buscadas do escopo mais interno para o mais externo.
- **Declaração prévia**: uso de variável não declarada gera erro.
- **Redeclaração**: declarar a mesma variável no mesmo escopo gera erro.
- **Compatibilidade de tipos**:
  - Operadores aritméticos (`+`, `-`, `*`, `/`) exigem `int` → produzem `int`
  - Operadores relacionais (`<`, `>`) exigem `int` → produzem `bool`
  - Operadores de igualdade (`==`, `!=`) exigem tipos iguais → produzem `bool`
  - Operador `!` exige `bool` → produz `bool`
  - Condições de `if` e `while` devem ser `bool`
  - `read()` aceita somente variáveis `int`

---

### Fase 4 — Geração de IR/TAC (`ir`)

O `Generator` traduz a AST para **Código de Três Endereços (TAC)**, uma representação linear no formato:

```
resultado = operando1  op  operando2
```

Variáveis temporárias são geradas automaticamente (`t0`, `t1`, `t2`, ...) e labels de desvio também (`L0`, `L1`, ...).

**Padrão TAC para `if-else`:**
```
  JUMPF <cond> Lelse
  <bloco then>
  JUMP Lend
Lelse:
  <bloco else>
Lend:
```

**Padrão TAC para `while`:**
```
Lstart:
  <cond>
  JUMPF <cond> Lend
  <corpo>
  JUMP Lstart
Lend:
```

---

### Fase 5 — Máquina Virtual (`vm`)

A `VM` executa as instruções TAC em um ciclo **fetch → decode → execute**. Usa:

- **`memory`** (`map[string]int`): armazena variáveis e temporários
- **`labels`** (`map[string]int`): índice de labels para desvios em O(1)
- **`ip`**: ponteiro de instrução

Booleanos são representados internamente como inteiros (`0` = false, `1` = true). A VM detecta divisão por zero e entrada inválida em `read()`, reportando erros em tempo de execução.

---

## Mensagens de Erro

O compilador reporta erros em cada fase com indicação de linha:

```
// Erro léxico
linha 3: caractere inesperado: '@'

// Erro sintático
erro sintático linha 5: esperado ";", encontrado "}" (RBRACE)

// Erro semântico
erro semântico linha 8: variável "x" não declarada
erro semântico linha 12: atribuição inválida: variável "flag" é "bool", mas expressão é "int"

// Erro em tempo de execução
erro em tempo de execução: divisão por zero
```
