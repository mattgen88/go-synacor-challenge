package virtualmachine

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// https://stackoverflow.com/questions/28541609/looking-for-reasonable-stack-implementation-in-golang
type stack []uint16

func (s *stack) Push(v uint16) {
	*s = append(*s, v)
}

func (s *stack) Pop() uint16 {
	result := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return result
}

const (
	_Halt = iota
	_Set
	_Push
	_Pop
	_Eq
	_Gt
	_Jmp
	_Jt
	_Jf
	_Add
	_Mult
	_Mod
	_And
	_Or
	_Not
	_Rmem
	_Wmem
	_Call
	_Ret
	_Out
	_In
	_Noop

	// MaxValue is the max value a value can take before considered special
	MaxValue = 32768
)

// VM is a vm
type VM struct {
	memory    []uint16
	DebugFD   *os.File      `json:"-"`
	StateFD   *os.File      `json:"-"`
	Registers []uint16      `json:"registers"`
	Stack     *stack        `json:"stack"`
	Pointer   uint16        `json:"pointer"`
	In        *bufio.Reader `json:"-"`
}

// Debug Writes current state out to debug log
func (v *VM) Debug() {
	// table := tablewriter.NewWriter(v.DebugFD)
	// table.SetHeader([]string{"Instruction #", "Value", "0", "1", "2", "3", "4", "5", "6", "7", "Stack"})
	// table.Append([]string{
	// 	fmt.Sprintf("%d", v.Pointer),
	// 	fmt.Sprintf("%d", v.memory[v.Pointer]),
	// 	fmt.Sprintf("%d", v.Registers[0]),
	// 	fmt.Sprintf("%d", v.Registers[1]),
	// 	fmt.Sprintf("%d", v.Registers[2]),
	// 	fmt.Sprintf("%d", v.Registers[3]),
	// 	fmt.Sprintf("%d", v.Registers[4]),
	// 	fmt.Sprintf("%d", v.Registers[5]),
	// 	fmt.Sprintf("%d", v.Registers[6]),
	// 	fmt.Sprintf("%d", v.Registers[7]),
	// 	spew.Sprint(v.Stack),
	// })
	// table.Render()
}

// Start starts the VM
func (v *VM) Start() {
	for {
		opCode := v.ReadValue()
		switch opCode {
		case _Halt:
			fmt.Printf("halt\n")
			os.Exit(0)

		case _Set:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			v.Registers[a] = b

		case _Push:
			a := v.ReadValue()
			v.Stack.Push(a)

		case _Pop:
			a := v.ReadAsRegNum()
			v.Registers[a] = v.Stack.Pop()

		case _Eq:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = 0
			if b == c {
				v.Registers[a] = 1
			}

		case _Gt:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = 0
			if b > c {
				v.Registers[a] = 1
			}

		case _Jmp:
			a := v.ReadValue()
			v.Pointer = a

		case _Jt:
			a := v.ReadValue()
			b := v.ReadValue()
			if a != 0 {
				v.Pointer = b
			}

		case _Jf:
			a := v.ReadValue()
			b := v.ReadValue()
			if a == 0 {
				v.Pointer = b
			}

		case _Add:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = (b + c) % MaxValue

		case _Mult:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = (b * c) % MaxValue

		case _Mod:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = b % c

		case _And:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = b & c

		case _Or:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			c := v.ReadValue()
			v.Registers[a] = b | c

		case _Not:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			// b0111111111111111 == 0x7FFF or 15bit bitwise inverse
			v.Registers[a] = ^b & 0x7FFF

		case _Rmem:
			a := v.ReadAsRegNum()
			b := v.ReadValue()
			v.Registers[a] = v.memory[b]

		case _Wmem:
			a := v.ReadValue()
			b := v.ReadValue()
			v.memory[a] = b

		case _Call:
			a := v.ReadValue()
			v.Stack.Push(v.Pointer)
			v.Pointer = a

		case _Ret:
			v.Pointer = v.Stack.Pop()

		case _Out:
			a := v.ReadValue()
			fmt.Print(string(a))

		case _In:
			a := v.ReadAsRegNum()

			if v.In.Size() != 0 {
				char, _ := v.In.ReadByte()
				v.Registers[a] = uint16(char)
			}

		case _Noop:
		default:
			panic(fmt.Sprintf("Read an instruction of %d", opCode))
		}
	}
}

// ReadValue reads the current memory location and returns dereferenced value
func (v *VM) ReadValue() uint16 {
	v.Debug()
	val := uint16(v.memory[v.Pointer])
	v.Pointer++
	return v.Dereference(val)
}

// ReadAsRegNum reads the current memory location and returns the register number referenced
func (v *VM) ReadAsRegNum() uint16 {
	v.Debug()
	val := uint16(v.memory[v.Pointer])
	v.Pointer++
	return val - MaxValue
}

// Dereference will return a value or get the value from the appropriate register from val
func (v *VM) Dereference(val uint16) uint16 {
	if val >= MaxValue {
		reg := val - MaxValue
		return v.GetRegister(reg)
	}
	return val
}

// GetRegister fetches a value from a register
func (v *VM) GetRegister(reg uint16) uint16 {
	return uint16(v.Registers[reg])
}

// Load loads the state
func (v *VM) Load(f *os.File) {
	state, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(state, v)
}

// Save saves the state
func (v *VM) Save(f *os.File) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	v.StateFD.Write(b)
}

// New Constructs and initializes a new VM and sets its instruction set
func New(instructionSet []uint16) *VM {
	vm := &VM{}
	vm.memory = instructionSet
	vm.Registers = make([]uint16, 8)
	vm.Pointer = 0
	vm.Stack = &stack{}
	vm.In = bufio.NewReader(os.Stdin)
	return vm
}
