// Package vm implementa a Máquina Virtual de Pilha da SheiLang.
//
// Responsabilidade: executar as instruções TAC geradas pelo ir.Generator.
//
// A VM usa uma pilha (stack) para operações e um mapa (memory)
// para armazenar variáveis e temporários.
//
// Ciclo de execução:
//  1. Buscar instrução (fetch)
//  2. Decodificar operação (decode)
//  3. Executar (execute)
package vm

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"compilador.com/m/ir"
)

// =========================================================
// 1. ESTRUTURA DA VM
// =========================================================

// VM é a máquina virtual de pilha.
type VM struct {
	instructions []ir.Instruction
	labels       map[string]int // nome do label → índice na lista
	memory       map[string]int // variáveis e temporários
	ip           int            // instruction pointer
	reader       *bufio.Reader
}

// New cria e prepara a VM com as instruções geradas.
func New(instructions []ir.Instruction) *VM {
	vm := &VM{
		instructions: instructions,
		labels:       make(map[string]int),
		memory:       make(map[string]int),
		reader:       bufio.NewReader(os.Stdin),
	}
	vm.buildLabelIndex()
	return vm
}

// buildLabelIndex pré-processa os labels para desvios O(1).
func (vm *VM) buildLabelIndex() {
	for i, instr := range vm.instructions {
		if instr.Op == ir.OP_LABEL {
			vm.labels[instr.Result] = i
		}
	}
}

// =========================================================
// 2. RESOLUÇÃO DE OPERANDOS
// =========================================================

// resolve retorna o valor inteiro de um operando.
// O operando pode ser:
//   - Literal numérico: "42" → 42
//   - Nome de variável/temporário: "x" → vm.memory["x"]
func (vm *VM) resolve(operand string) int {
	if val, err := strconv.Atoi(operand); err == nil {
		return val
	}
	return vm.memory[operand]
}

// boolToInt converte bool → 0 ou 1.
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// =========================================================
// 3. LOOP DE EXECUÇÃO
// =========================================================

// Run executa o programa até o fim das instruções.
func (vm *VM) Run() error {
	for vm.ip < len(vm.instructions) {
		instr := vm.instructions[vm.ip]
		vm.ip++

		if err := vm.execute(instr); err != nil {
			return err
		}
	}
	return nil
}

// execute despacha e executa uma instrução TAC.
func (vm *VM) execute(instr ir.Instruction) error {
	switch instr.Op {

	// --- Labels: ignorados no execute (já indexados) ---
	case ir.OP_LABEL:
		// nada a fazer

	// --- Atribuição ---
	case ir.OP_ASSIGN:
		vm.memory[instr.Result] = vm.resolve(instr.Arg1)

	// --- Operações aritméticas ---
	case ir.OP_ADD:
		vm.memory[instr.Result] = vm.resolve(instr.Arg1) + vm.resolve(instr.Arg2)
	case ir.OP_SUB:
		vm.memory[instr.Result] = vm.resolve(instr.Arg1) - vm.resolve(instr.Arg2)
	case ir.OP_MUL:
		vm.memory[instr.Result] = vm.resolve(instr.Arg1) * vm.resolve(instr.Arg2)
	case ir.OP_DIV:
		divisor := vm.resolve(instr.Arg2)
		if divisor == 0 {
			return fmt.Errorf("erro em tempo de execução: divisão por zero")
		}
		vm.memory[instr.Result] = vm.resolve(instr.Arg1) / divisor

	// --- Operadores relacionais e de igualdade ---
	case ir.OP_EQ:
		vm.memory[instr.Result] = boolToInt(vm.resolve(instr.Arg1) == vm.resolve(instr.Arg2))
	case ir.OP_NEQ:
		vm.memory[instr.Result] = boolToInt(vm.resolve(instr.Arg1) != vm.resolve(instr.Arg2))
	case ir.OP_LT:
		vm.memory[instr.Result] = boolToInt(vm.resolve(instr.Arg1) < vm.resolve(instr.Arg2))
	case ir.OP_GT:
		vm.memory[instr.Result] = boolToInt(vm.resolve(instr.Arg1) > vm.resolve(instr.Arg2))

	// --- Operadores unários ---
	case ir.OP_NEG:
		vm.memory[instr.Result] = -vm.resolve(instr.Arg1)
	case ir.OP_NOT:
		val := vm.resolve(instr.Arg1)
		if val == 0 {
			vm.memory[instr.Result] = 1
		} else {
			vm.memory[instr.Result] = 0
		}

	// --- Desvios ---
	case ir.OP_JUMP:
		idx, ok := vm.labels[instr.Arg1]
		if !ok {
			return fmt.Errorf("label desconhecido: %q", instr.Arg1)
		}
		vm.ip = idx

	case ir.OP_JUMPF:
		cond := vm.resolve(instr.Arg1)
		if cond == 0 { // false → desvia
			idx, ok := vm.labels[instr.Arg2]
			if !ok {
				return fmt.Errorf("label desconhecido: %q", instr.Arg2)
			}
			vm.ip = idx
		}

	// --- E/S ---
	case ir.OP_PRINT:
		val := vm.resolve(instr.Arg1)
		fmt.Println(val)

	case ir.OP_READ:
		fmt.Printf("entrada> ")
		line, _ := vm.reader.ReadString('\n')
		line = strings.TrimSpace(line)
		val, err := strconv.Atoi(line)
		if err != nil {
			return fmt.Errorf("entrada inválida: %q não é um inteiro", line)
		}
		vm.memory[instr.Arg1] = val

	default:
		return fmt.Errorf("instrução desconhecida: %s", instr.Op)
	}

	return nil
}