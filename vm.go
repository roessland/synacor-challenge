package main

import "bytes"

import "io"
import "bufio"
import "fmt"
import "encoding/binary"
import "os"
import "log"
import "io/ioutil"

type Byte struct {
	High, Low uint16
}

type VM struct {
	Register []uint16
	Memory   []uint16
	Stack    []uint16
	IP       uint16
}

func NewVM() *VM {
	return &VM{
		Register: make([]uint16, 8),
		Memory:   make([]uint16, 32768),
		Stack:    make([]uint16, 0),
	}
}

func IsRegister(address uint16) bool {
	if 0 <= address && address <= 32767 {
		return false
	} else if 32768 <= address && address <= 32775 {
		return true
	} else {
		log.Print("VM Get: Invalid memory value %v\n", address)
		return false
	}
}

// At returns the literal value if it is in [0,32767], but if the value
// is in [32768, 32775] the value in the corresponding register is returned.
// ONLY for getting a, b, c. Not for memory!
func (vm *VM) Enhance(address uint16) uint16 {
	value := vm.Memory[address]
	if 0 <= value && value <= 32767 {
		return value
	} else if 32768 <= value && value <= 32775 {
		return vm.Register[value-32768]
	} else {
		log.Print("VM Enhance: Invalid memory value %v\n", value)
		return 50000
	}
}

func (vm *VM) At(address uint16) uint16 {
	value := vm.Memory[address]
	if 0 <= value && value <= 32767 {
		return vm.Memory[value]
	} else if 32768 <= value && value <= 32775 {
		return vm.Register[value-32768]
	} else {
		log.Print("VM At: Invalid memory value %v\n", value)
		return 50000
	}
}

func (vm *VM) Set(address uint16, value uint16) {
	if 0 <= address && address <= 32767 {
		vm.Memory[address] = value
	} else if 32768 <= address && address <= 32775 {
		vm.Register[address-32768] = value
	} else {
		log.Print("VM Get: Invalid memory write %v\n", address)
	}
}

/*
Current op: 4
Current op: 8
Current op: 10
Current op: 4
Current op: 8
Current op: 6
Current op: 15
Current op: 4
Current op: 8
Current op: 19
nCurrent op: 19
*/

func (vm *VM) Execute() {
	opnames := map[uint16]string{
		0:  "halt",
		1:  "set a b",
		2:  "push a",
		3:  "pop a",
		4:  "eq a b c",
		5:  "gt a b c",
		6:  "jmp a",
		7:  "jt a b",
		8:  "jf a b",
		9:  "add a b c",
		10: "mult a b c",
		11: "mod a b c",
		12: "and a b c",
		13: "or a b c",
		14: "not a b",
		15: "rmem a b",
		16: "wmem a b",
		17: "call a",
		18: "ret",
		19: "out a",
		20: "in a",
		21: "noop",
	}

	reader := bufio.NewReader(os.Stdin)
	charStack := []byte{}
	for {
		op := vm.Memory[vm.IP]
		if false {
			fmt.Printf("----------")
			fmt.Printf("IP: %v\n", vm.IP)
			fmt.Printf("Op: %v\n", opnames[op])
			fmt.Printf("<a> (%v, %v) = %v\n", vm.IP+1, vm.Memory[vm.IP+1], vm.Enhance(vm.IP+1))
			fmt.Printf("<b> (%v, %v) = %v\n", vm.IP+2, vm.Memory[vm.IP+2], vm.Enhance(vm.IP+2))
			fmt.Printf("<c> (%v, %v) = %v\n", vm.IP+3, vm.Memory[vm.IP+3], vm.Enhance(vm.IP+3))
			fmt.Printf("Registers: %v\n", vm.Register)
		}
		switch op {
		case 0:
			// halt: 0
			// stop execution and terminate program
			return
		case 1:
			// set: 1 a b
			// set register <a> to the value of <b>
			vm.Register[vm.Memory[vm.IP+1]-32768] = vm.Enhance(vm.IP + 2)
			vm.IP += 3
		case 2:
			// push: 2 a
			// push <a> onto the stack
			vm.Stack = append(vm.Stack, vm.Enhance(vm.IP+1))
			vm.IP += 2
		case 3:
			// pop: 3 a
			// remove top element from the stack and write it into <a>; empty stack = error
			if len(vm.Stack) == 0 {
				log.Fatal("VM pop: Stack was empty!")
			}
			val := vm.Stack[len(vm.Stack)-1]
			vm.Register[vm.Memory[vm.IP+1]-32768] = val
			vm.Stack = vm.Stack[:len(vm.Stack)-1]
			vm.IP += 2
		case 4:
			// eq: 4 a b c
			// set <a> to 1 if <b> is equal to <c>; set it to 0 otherwise
			if vm.Enhance(vm.IP+2) == vm.Enhance(vm.IP+3) {
				vm.Register[vm.Memory[vm.IP+1]-32768] = 1
			} else {
				vm.Register[vm.Memory[vm.IP+1]-32768] = 0
			}
			vm.IP += 4
		case 5:
			// gt: 5 a b c
			// set <a> to 1 if <b> is greater than <c>; set it to 0 otherwise
			if vm.Enhance(vm.IP+2) > vm.Enhance(vm.IP+3) {
				vm.Register[vm.Memory[vm.IP+1]-32768] = 1
			} else {
				vm.Register[vm.Memory[vm.IP+1]-32768] = 0
			}
			vm.IP += 4
		case 6:
			// jmp: 6 a
			// jump to <a>
			vm.IP = vm.Enhance(vm.IP + 1)
		case 7:
			// jt: 7 a b
			// if <a> is nonzero, jump to <b>
			if vm.Enhance(vm.IP+1) != 0 {
				vm.IP = vm.Enhance(vm.IP + 2)
			} else {
				vm.IP += 3
			}
		case 8:
			// jf: 8 a b
			// if <a> is zero, jump to <b>
			if vm.Enhance(vm.IP+1) == 0 {
				vm.IP = vm.Enhance(vm.IP + 2)
			} else {
				vm.IP += 3
			}
		case 9:
			// add: 9 a b c
			// assign into <a> the sum of <b> and <c> (modulo 32768)
			sum := (int(vm.Enhance(vm.IP+2)) + int(vm.Enhance(vm.IP+3))) % 32768
			vm.Register[vm.Memory[vm.IP+1]-32768] = uint16(sum)
			vm.IP += 4
		case 10:
			// mult: 10 a b c
			// store into <a> the product of <b> and <c> (modulo 32768)
			prod := (int(vm.Enhance(vm.IP+2)) * int(vm.Enhance(vm.IP+3))) % 32768
			vm.Register[vm.Memory[vm.IP+1]-32768] = uint16(prod)
			vm.IP += 4
		case 11:
			// mod 11 a b c
			// store into <a> the remainder of <b> divided by <c>
			rem := vm.Enhance(vm.IP+2) % vm.Enhance(vm.IP+3)
			vm.Register[vm.Memory[vm.IP+1]-32768] = rem
			vm.IP += 4
		case 12:
			// and: 12 a b c
			// stores into <a> the bitwise and of <b> and <c>
			vm.Register[vm.Memory[vm.IP+1]-32768] = vm.Enhance(vm.IP+2) & vm.Enhance(vm.IP+3)
			vm.IP += 4
		case 13:
			// or: 13 a b c
			// stores into <a> the bitwise or of <b> and <c>
			vm.Register[vm.Memory[vm.IP+1]-32768] = vm.Enhance(vm.IP+2) | vm.Enhance(vm.IP+3)
			vm.IP += 4
		case 14:
			// not: 14 a b
			// stores into <a> the 15-bit bitwise not of <b>
			vm.Register[vm.Memory[vm.IP+1]-32768] = (^vm.Enhance(vm.IP + 2)) & 0x7FFF
			vm.IP += 3
		case 15:
			// rmem: 15 a b
			// read memory at address <b> and write it to <a>
			vm.Register[vm.Memory[vm.IP+1]-32768] = vm.Memory[vm.Enhance(vm.IP+2)]
			vm.IP += 3
		case 16:
			// wmem 16 a b
			// write the value from <b> into memory at address <a>
			vm.Memory[vm.Enhance(vm.IP+1)] = vm.Enhance(vm.IP + 2)
			vm.IP += 3
		case 17:
			// call: 17 a
			// write the address of the next instruction to the stack and jump to <a>
			vm.Stack = append(vm.Stack, vm.IP+2)
			vm.IP = vm.Enhance(vm.IP + 1)
		case 18:
			// ret: 18
			// remove the top element from the stack and jump to it; empty stack = halt
			if len(vm.Stack) == 0 {
				fmt.Printf("Empty stack. Halting...\n")
				return
			}
			p := vm.Stack[len(vm.Stack)-1]
			vm.Stack = vm.Stack[:len(vm.Stack)-1]
			vm.IP = p
		case 19:
			// out: 19 a
			fmt.Printf("%c", vm.Enhance(vm.IP+1))
			vm.IP += 2
		case 20:
			// in: 20 a
			// Read a char and write its ascii code to <a>
			for len(charStack) == 0 {
				line, _ := reader.ReadBytes('\n')
				// Reverse line so it is popped properly
				for i, j := 0, len(line)-1; i < j; i, j = i+1, j-1 {
					line[i], line[j] = line[j], line[i]
				}
				charStack = append(charStack, line...)
			}
			c := charStack[len(charStack)-1]
			charStack = charStack[:len(charStack)-1]
			vm.Register[vm.Memory[vm.IP+1]-32768] = uint16(c)
			vm.IP += 2
		case 21:
			// noop: 21
			// no operation
			vm.IP += 1
		default:
			fmt.Printf("VM: Unknown opcode %v\n", op)
			return
		}
	}
}

func (vm *VM) LoadBinary(filename string) *VM {
	log.Printf("Loading file %v\n", filename)
	var err error
	b, err := ioutil.ReadFile(filename)
	buf := bytes.NewReader(b)
	if err != nil {
		log.Print(err)
	}

	for i := 0; ; i++ {
		var value Byte
		err = binary.Read(buf, binary.LittleEndian, &value)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Read failed", err)
		}
		vm.Memory[2*i] = value.High
		vm.Memory[2*i+1] = value.Low
	}
	log.Printf("Finished loading file %v\n", filename)

	return vm
}

func main() {
	log.SetOutput(os.Stdout)
	vm := NewVM()
	vm.LoadBinary("challenge.bin")
	vm.Execute()
}
