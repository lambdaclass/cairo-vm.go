# cairo-vm.go

This is a work in progress implementation of the [Cairo VM](https://github.com/lambdaclass/cairo-vm) in `Go`. The reasons for doing this include:

- Having a diversity of implementations helps find bugs and make the whole ecosystem more resilient.
- It's a good opportunity to extensively document the VM in general, as currently the documentation on its internals is very scarce and mostly lives on the minds of a few people.

## Other docs

- [Project layout](docs/layout.md)
- [Rust/lambdaworks integration](docs/rust-integration.md)

## Installation

Go needs to be installed. For mac computers, run

```shell
brew install go
```

We also use [pyenv](https://github.com/pyenv/pyenv) to install testing-related dependencies

## Compiling, running, testing

To compile, run:

```shell
make build
```

To run the main example file, run:

```shell
make run
```

Before running the tests, install the testing dependencies:

```
make deps
```

To run all tests, activate the venv created by make deps and run the test target:

```shell
. cairo-vm-env/bin/activate
make test
```

## Running the demo

This project currently has two demo targets, one for running a fibonacci programs and one for running a factorial program. Both of them output their corresponding trace files.
The demo uses cairo_lang to compile both cairo programs, you can install it by running `make deps` (or `make deps-macos` if you are on macos)

To run the fibonacci demo:

```shell
    make demo_fib
```

To run the factorial demo:

```shell
    make demo_factorial
```

## Project Guidelines

- PRs addressing performance are forbidden. We are currently concerned with making it work without bugs and nothing more.
- All PRs must contain tests. Code coverage has to be above 98%.
- To check for security and other types of bugs, the code will be fuzzed extensively.
- PRs must be accompanied by its corresponding documentation. A book will be written documenting the entire inner workings of it, so anyone can dive in to a Cairo VM codebase and follow it along.

# Roadmap

## First milestone: Fibonacci/Factorial

The first milestone for Cairo VM in Go is completed! :tada:

The milestone includes:

- Parsing of `json` programs
- Decoding of instructions
- Memory relocation
- Instruction execution.
- Writing of the trace into files with the correct format.
- Make the fibonacci and factorial tests pass, comparing our own trace with the Rust VM one, making sure they match.
- Writing of the memory into files with the correct format.
- Make the fibonacci and factorial tests pass, comparing our own memory with the Rust VM one, making sure they match.

## Cairo 0/Cairo 1

The above will work for Cairo 0 programs. Cairo 1 has the following extra issues to address:

- There is no `Json` representation of a Cairo 1 Program, so we can only run contracts. This means we will have to add functions to run cairo contracts (aka implement run_from_entrypoint).
- Cairo 1 contracts use the `range_check` builtin by default, so we need to implement it.

## Full VM implementation

To have a full implementation, we will need the following:

- Builtins. Add the `BuiltinRunner` logic, then implement each builtin:
    - `Range check (RC)`
    - `Output`
    - `Bitwise`
    - `Ec_op`
    - `Pedersen`
    - `Keccak`
    - `Poseidon`
    - `Signature`
    - `Segment Arena`
- Memory layouts. This is related to builtins but we will do it after implementing them.
- Hints. Add the `HintProcessor` logic, then implement each hint. Hints need to be documented extensively, implementing them is super easy since it's just porting code; what's not so clear is what they are used for. Having explanations and examples for each is important. List of hints below:
    - Parsing of references https://github.com/lambdaclass/cairo-vm/tree/main/docs/references_parsing
    - `CommonLib` https://github.com/starkware-libs/cairo-lang/tree/master/src/starkware/cairo/common
    - `Secp`
    - `Vrf`
    - `BigInt`
    - `Blake2`
    - `DictManager`
    - `EcRecover`
    - `Field Arithmetic`
    - `Garaga`
    - `set_add`
    - `sha256 utils`
    - `ECDSA verification`
    - `uint384` and `uint384 extension`
    - `uint512 utils`
    - `Cairo 1` hints.
- Proof mode. It's important to explain in detail what this is when we do it. It's one of the most obscure parts of the VM in my experience.
- Air Public inputs. (Tied to Proof-mode)
- Temporary segments.
- Program tests from Cairo VM in Rust.
- Fuzzing/Differential fuzzing.

# Documentation

## High Level Overview

The Cairo virtual machine is meant to be used in the context of STARK validity proofs. What this means is that the point of Cairo is not just to execute some code and get a result, but to *prove* to someone else that said execution was done correctly, without them having to re-execute the entire thing. The rough flow for it looks like this:

- A user writes a Cairo program.
- The program is compiled into Cairo's VM bytecode.
- The VM executes said code and provides a *trace* of execution, i.e. a record of the state of the machine and its memory *at every step of the computation*.
- This trace is passed on to a STARK prover, which creates a cryptographic proof from it, attesting to the correct execution of the program.
- The proof is passed to a verifier, who checks that the proof is valid in a fraction of a second, without re-executing.

The main three components of this flow are:

- A Cairo compiler to turn a program written in the [Cairo programming language](https://www.cairo-lang.org/) into bytecode.
- A Cairo VM to then execute it and generate a trace.
- [A STARK prover and verifier](https://github.com/lambdaclass/starknet_stack_prover_lambdaworks) so one party can prove correct execution, while another can verify it.

While this repo is only concerned with the second component, it's important to keep in mind the other two; especially important are the prover and verifier that this VM feeds its trace to, as a lot of its design decisions come from them. This virtual machine is designed to make proving and verifying both feasible and fast, and that makes it quite different from most other VMs you are probably used to.

## Basic VM flow

Our virtual machine has a very simple flow:

- Take a compiled cairo program as input. You can check out an example program [here](https://github.com/lambdaclass/cairo_vm.go/blob/main/cairo_programs/fibonacci.cairo), and its corresponding compiled version [here](https://github.com/lambdaclass/cairo_vm.go/blob/main/cairo_programs/fibonacci.json).
- Run the bytecode from the compiled program, doing the usual `fetch->decode->execute` loop, running until program termination.
- On every step of the execution, record the values of each register.
- Take the register values and memory at every step and write them to a file, called the `execution trace`.

Barring some simplifications we made, this is all the Cairo VM does. The two main things that stand out as radically different are the memory model and the use of `Field Elements` to perform arithmetic. Below we go into more detail on each step, and in the process explain the ommisions we made.

## Architecture

The Cairo virtual machine uses a Von Neumann architecture with a Non-deterministic read-only memory. What this means, roughly, is that memory is immutable after you've written to it (i.e. you can only write to it once); this is to make the STARK proving easier, but we won't go into that here.

### Memory Segments and Relocation

The process of memory allocation in a contiguous write-once memory region can get pretty complicated. Imagine you want to have a regular call stack, with a stack pointer pointing to the top of it and allocation and deallocation of stack frames and local variables happening throughout execution. Because memory is immutable, this cannot be done the usual way; once you allocate a new stack frame that memory is set, it can't be reused for another one later on.

Because of this, memory in Cairo is divided into `segments`. This is just a way of organizing memory more conveniently for this write-once model. Each segment is nothing more than a contiguous memory region. Segments are identified by an `index`, an integer value that uniquely identifies them.

Memory `cells` (i.e. values in memory) are identified by the index of the segment they belong to and an `offset` into said segment. Thus, the memory cell `{2,0}` is the first cell of segment number `2`.

Even though this segment model is extremely convenient for the VM's execution, the STARK prover needs to have the memory as just one contiguous region. Because of this, once execution of a Cairo program finishes, all the memory segments are collapsed into one; this process is called `Relocation`. We will go into more detail on all of this below.

### Registers

There are only three registers in the Cairo VM:

- The program counter `pc`, which points to the next instruction to be executed.
- The allocation pointer `ap`, pointing to the next unused memory cell.
- The frame pointer `fp`, pointing to the base of the current stack frame. When a new function is called, `fp` is set to the current `ap`. When the function returns, `fp` goes back to its previous value. The VM creates new segments whenever dynamic allocation is needed, so for example the cairo analog to a Rust `Vec` will have its own segment. Relocation at the end meshes everything together.

### Instruction Decoding/Execution

TODO: explain the components of an instruction (`dst_reg`, `op0_reg`, etc), what each one is used for and how they're encoded/decoded.

### Felts

Felts, or Field Elements, are cairo's basic integer type. Every variable in a cairo vm that is not a pointer is a felt. From our point of view we could say a felt in cairo is an unsigned integer in the range [0, CAIRO_PRIME). This means that all operations are done modulo CAIRO_PRIME. The CAIRO_PRIME is 0x800000000000011000000000000000000000000000000000000000000000001, which means felts can be quite big (up to 252 bits), luckily, we have the [Lambdaworks](https://github.com/lambdaclass/lambdaworks) library to help with handling these big integer values and providing fast and efficient modular arithmetic.

### Lambdaworks library wrapper 

[Lambdaworks](https://github.com/lambdaclass/lambdaworks) is a custom performance-focused library that aims to ease programming for developers. It provides essential mathematical and cryptographic methods required for this project, enabling arithmetic operations between `felts` and type conversions efficiently.
We've developed a C wrapper to expose the library's functions and enable easy usage from Go. This allows seamless integration of the library's features within Go projects, enhancing performance and functionality.

### More on memory

The cairo memory is made up of contiguous segments of variable length identified by their index. The first segment (index 0) is the program segment, which stores the instructions of a cairo program. The following segment (index 1) is the execution segment, which holds the values that are created along the execution of the vm, for example, when we call a function, a pointer to the next instruction after the call instruction will be stored in the execution segment which will then be used to find the next instruction after the function returns. The following group of segments are the builtin segments, one for each builtin used by the program, and which hold values used by the builtin runners. The last group of segments are the user segments, which represent data structures created by the user, for example, when creating an array on a cairo program, that array will be represented in memory as its own segment.

An address (or pointer) in cairo is represented as a `relocatable` value, which is made up of a `segment_index` and an `offset`, the `segment_index` tells us which segment the value is stored in and the `offset` tells us how many values exist between the start of the segment and the value.

As the cairo memory can hold both felts and pointers, the basic memory unit is a `maybe_relocatable`, a variable that can be either a `relocatable` or a `felt`.

While memory is continous, some gaps may be present. These gaps can be created on purpose by the user, for example by running:

```
[ap + 1] = 2;
```

Where a gap is created at ap. But they may also be created indireclty by diverging branches, as for example one branch may declare a variable that the other branch doesn't, as memory needs to be allocated for both cases if the second case is ran then a gap is left where the variable should have been written.

#### Memory API

The memory can perform the following basic operations:

- `memory_add_segment`: Creates a new, empty segment in memory and returns a pointer to its start. Values cannot be inserted into a memory segment that hasn't been previously created.

- `memory_insert`: Inserts a `maybe_relocatable` value at an address indicated by a `relocatable` pointer. For this operation to succeed, the pointer's segment_index must be an existing segment (created using `memory_add_segment`), and there mustn't be a value stored at that address, as the memory is immutable after its been written once. If there is a value already stored at that address but it is equal to the value to be inserted then the operation will be successful.

- `memory_get`: Fetches a `maybe_relocatable` value from a memory address indicated by a `relocatable` pointer.

Other operations:

- `memory_load_data`: This is a convenience method, which takes an array of `maybe_relocatable` and inserts them contiguosuly in memory by calling `memory_insert` and advancing the pointer by one after each insertion. Returns a pointer to the next free memory slot after the inserted data.

#### Memory Relocation

During execution, the memory consists of segments of varying length, and they can be accessed by indicating their segment index, and the offset within that segment. When the run is finished, a relocation process takes place, which transforms this segmented memory into a contiguous list of values. The relocation process works as follows:

1- The size of each segment is calculated (The size is equal to the highest offset within the segment + 1, and not the amount of `maybe_relocatable` values, as there can be gaps)
2- A base is assigned to each segment by accumulating the size of the previous segment. The first segment's base is set to 1.
3- All `relocatable` values are converted into a single integer by adding their `offset` value to their segment's base calculated in the previous step

For example, if we have this memory represented by address, value pairs:

    0:0 -> 1
    0:1 -> 4
    0:2 -> 7
    1:0 -> 8
    1:1 -> 0:2
    1:4 -> 0:1
    2:0 -> 1

Step 1: Calculate segment sizes:

    0 --(has size)--> 3
    1 --(has size)--> 5
    2 --(has size)--> 1

Step 2: Assign a base to each segment:

    0 --(has base value)--> 1
    1 --(has base value)--> 4 (that is: 1 + 3)
    2 --(has base value)--> 9 (that is: 4 + 5)

Step 3: Convert relocatables to integers

    1 (base[0] + 0) -> 1
    2 (base[0] + 1) -> 4
    3 (base[0] + 2) -> 7
    4 (base[1] + 0) -> 8
    5 (base[1] + 1) -> 3 (that is: base[0] + 2)
    .... (memory gaps)
    8 (base[1] + 4) -> 2 (that is: base[0] + 1)
    9 (base[2] + 0) -> 1

### Program parsing

The input of the Virtual Machine is a compiled Cairo program in Json format. The main part of the file are listed below:

- data: List of hexadecimal values that represent the instructions and immediate values defined in the cairo program. Each hexadecimal value is stored as a maybe_relocatable element in memory, but they can only be felts because the decoder has to be able to get the instruction fields in its bit representation.

- debug_info: This field provides information about the instructions defined in the data list. Each one is identified with its index inside the data list. For each one it contains information about the cairo variables in scope, the hints executed before that instruction if any, and its location inside the cairo program.

- hints: All the hints used in the program, ordered by the pc offset at which they should be executed.

- identifiers: User-defined symbols in the Cairo code representing variables, functions, classes, etc. with unique names. The expected offset, type and its corresponding information is provided for each identifier

    For example, the identifier representing the main function (usually the entrypoint of the program) is of `function` type, and a list of decorators wrappers (if there are any) are provided as additional information.
    Another example is a user defined struct, is of `struct` type, it provides its size, the members it contains (with its information) and more.

- main_scope: Usually something like __main__. All the identifiers associated with main function will be identified as __main__.identifier_name. Useful to identify the entrypoint of the program.

- prime: The cairo prime in hexadecimal format. As explained above, all arithmetic operations are done over a base field, modulo this primer number.

- reference_manager: Contains information about cairo variables. This information is useful to access to variables when executing cairo hints.

In this project, we use a C++ library called [simdjson](https://github.com/simdjson/simdjson), the json is stored in a custom structure  which the vm can use to run the program and create a trace of its execution.

### Code walkthrough/Write your own Cairo VM

Let's begin by creating the basic types and structures for our VM:

### Felt

As anyone who has ever written a cairo program will know, everything in cairo is a Felt. We can think of it as our unsigned integer. In this project, we use the `Lambdaworks` library to abstract ourselves from modular arithmetic.

TODO: Instructions on how to use Lambdaworks felt from Go

### Relocatable

This is how cairo represents pointers, they are made up of `SegmentIndex`, which segment the variable is in, and `Offset`, how many values exist between the start of a segment and the variable. We represent them like this:

```go
type Relocatable struct {
	SegmentIndex int
	Offset       uint
}
```

### MaybeRelocatable

As the cairo memory can hold both felts and relocatables, we need a data type that can represent both in order to represent a basic memory unit.
We would normally use enums or unions to represent this type, but as go lacks both, we will instead hold a non-typed inner value and rely on the api to make sure we can only create MaybeRelocatable values with either Felt or Relocatable as inner type.

```go
type MaybeRelocatable struct {
	inner any
}

// Creates a new MaybeRelocatable with an Int inner value
func NewMaybeRelocatableInt(felt uint) *MaybeRelocatable {
	return &MaybeRelocatable{inner: Int{felt}}
}

// Creates a new MaybeRelocatable with a Relocatable inner value
func NewMaybeRelocatableRelocatable(relocatable Relocatable) *MaybeRelocatable {
	return &MaybeRelocatable{inner: relocatable}
}
```

We will also add some methods that will allow us access `MaybeRelocatable` inner values:

```go
// If m is Int, returns the inner value + true, if not, returns zero + false
func (m *MaybeRelocatable) GetInt() (Int, bool) {
	int, is_type := m.inner.(Int)
	return int, is_type
}

// If m is Relocatable, returns the inner value + true, if not, returns zero + false
func (m *MaybeRelocatable) GetRelocatable() (Relocatable, bool) {
	rel, is_type := m.inner.(Relocatable)
	return rel, is_type
}
```

These will allow us to safely discern between `Felt` and `Relocatable` values later on.

#### Memory

As we previously described, the memory is made up of a series of segments of variable length, each containing a continuous sequence of `MaybeRelocatable` elements. Memory is also immutable, which means that once we have written a value into memory, it can't be changed.
There are multiple valid ways to represent this memory structure, but the simplest way to represent it is by using a map, maping a `Relocatable` address to a `MaybeRelocatable` value.
As we don't have an actual representation of segments, we have to keep track of the number of segments.

```go
type Memory struct {
	data         map[Relocatable]MaybeRelocatable
	num_segments uint
}
```

Now we can define the basic memory operations:

*Insert*

Here we need to make perform some checks to make sure that the memory remains consistent with its rules:
- We must check that insertions are performed on previously-allocated segments, by checking that the address's segment_index is lower than our segment counter
- We must check that we are not mutating memory we have previously written, by checking that the memory doesn't already contain a value at that address that is not equal to the one we are inserting

```go
func (m *Memory) Insert(addr Relocatable, val *MaybeRelocatable) error {
	// Check that insertions are preformed within the memory bounds
	if addr.segmentIndex >= int(m.num_segments) {
		return errors.New("Error: Inserting into a non allocated segment")
	}

	// Check for possible overwrites
	prev_elem, ok := m.data[addr]
	if ok && prev_elem != *val {
		return errors.New("Memory is write-once, cannot overwrite memory value")
	}

	m.data[addr] = *val

	return nil
}
```

*Get*

This is the easiest operation, as we only need to fetch the value from our map:

```go
// Gets some value stored in the memory address `addr`.
func (m *Memory) Get(addr Relocatable) (*MaybeRelocatable, error) {
	value, ok := m.data[addr]

	if !ok {
		return nil, errors.New("Memory Get: Value not found")
	}

	return &value, nil
}
```

### MemorySegmentManager

In our `Memory` implementation, it looks like we need to have segments allocated before performing any valid memory operation, but we can't do so from the `Memory` api. To do so, we need to use the `MemorySegmentManager`.
The `MemorySegmentManager` is in charge of creating new segments and calculating their size during the relocation process, it has the following structure:

```go
type MemorySegmentManager struct {
	segmentSizes map[uint]uint
	Memory       Memory
}
```

And the following methods:

*Add Segment*

As we are using a map, we dont have to allocate memory for the new segment, so we only have to raise our segment counter and return the first address of the new segment:

```go
func (m *MemorySegmentManager) AddSegment() Relocatable {
	ptr := Relocatable{int(m.Memory.num_segments), 0}
	m.Memory.num_segments += 1
	return ptr
}
```

*Load Data*

This method inserts a contiguous array of values starting from a certain addres in memory, and returns the next address after the inserted values. This is useful when inserting the program's instructions in memory.
In order to perform this operation, we only need to iterate over the array, inserting each value at the address indicated by `ptr` while advancing the ptr with each iteration and then return the final ptr.

```go
func (m *MemorySegmentManager) LoadData(ptr Relocatable, data *[]MaybeRelocatable) (Relocatable, error) {
	for _, val := range *data {
		err := m.Memory.Insert(ptr, &val)
		if err != nil {
			return Relocatable{0, 0}, err
		}
		ptr.offset += 1
	}
	return ptr, nil
}
```

### RunContext

The RunContext keeps track of the vm's registers. Cairo VM only has 3 registers:

- The program counter `Pc`, which points to the next instruction to be executed.
- The allocation pointer `Ap`, pointing to the next unused memory cell.
- The frame pointer `Fp`, pointing to the base of the current stack frame. When a new function is called, `Fp` is set to the current `Ap` value. When the function returns, `Fp` goes back to its previous value.

We can represent it like this:

```go
type RunContext struct {
	Pc memory.Relocatable
	Ap memory.Relocatable
	Fp memory.Relocatable
}
```

### VirtualMachine
With all of these types and structures defined, we can build our VM:

```go
type VirtualMachine struct {
	RunContext     RunContext
	currentStep    uint
	Segments       memory.MemorySegmentManager
}
```

To begin coding the basic execution functionality of our VM, we only need these basic fields, we will be adding more fields as we dive deeper into this guide.

### Instruction Decoding and Execution


*Compute operands*

Once the instruction has been decoded, it is executed by the main loop `run_instruction` whose first function is compute operands. This function is in charge of
calculating the addresses of the operands and fetch them from memory. If the function could not fetch them from the memory then they are deduced from the other operands and
taking in consideration what kind of opcode is going to be executed. 

```go
func (vm *VirtualMachine) ComputeOperands(instruction Instruction) (Operands, error) {
	var res *memory.MaybeRelocatable

	dst_addr, err := vm.RunContext.ComputeDstAddr(instruction)
	if err != nil {
		return Operands{}, errors.New("FailedToComputeDstAddr")
	}
	dst, _ := vm.Segments.Memory.Get(dst_addr)

	op0_addr, err := vm.RunContext.ComputeOp0Addr(instruction)
	if err != nil {
		return Operands{}, fmt.Errorf("FailedToComputeOp0Addr: %s", err)
	}
	op0, _ := vm.Segments.Memory.Get(op0_addr)

	op1_addr, err := vm.RunContext.ComputeOp1Addr(instruction, op0)
	if err != nil {
		return Operands{}, fmt.Errorf("FailedToComputeOp1Addr: %s", err)
	}
	op1, _ := vm.Segments.Memory.Get(op1_addr)

	if op0 == nil {
		deducedOp0, deducedRes, err := vm.DeduceOp0(&instruction, dst, op1)
		if err != nil {
			return Operands{}, err
		}
		op0 = deducedOp0
		if op0 != nil {
			vm.Segments.Memory.Insert(op0_addr, op0)
		}
		res = deducedRes
	}

	if op1 == nil {
		deducedOp1, deducedRes, err := vm.DeduceOp1(instruction, dst, op0)
		if err != nil {
			return Operands{}, err
		}
		op1 = deducedOp1
		if op1 != nil {
			vm.Segments.Memory.Insert(op1_addr, op1)
		}
		if res == nil {
			res = deducedRes
		}
	}

	if res == nil {
		res, err = vm.ComputeRes(instruction, *op0, *op1)

		if err != nil {
			return Operands{}, err
		}
	}

	if dst == nil {
		deducedDst := vm.DeduceDst(instruction, res)
		dst = deducedDst
		if dst != nil {
			vm.Segments.Memory.Insert(dst_addr, dst)
		}
	}

	operands := Operands{
		Dst: *dst,
		Op0: *op0,
		Op1: *op1,
		Res: res,
	}
	return operands, nil
}
```

*ComputeDstAddr*
The function `ComputeDstAddress` computes the address of value that will be stored in the Destination register. It checks which is the destination register (wether AP or FP) and gets the direction from the run_context then if the instruction offset is negativa it substract from the address the offset otherwise it adds the offset.

```go
func (run_context RunContext) ComputeDstAddr(instruction Instruction) (memory.Relocatable, error) {
	var base_addr memory.Relocatable
	switch instruction.DstReg {
	case AP:
		base_addr = run_context.Ap
	case FP:
		base_addr = run_context.Fp
	}

	if instruction.Off0 < 0 {
		return base_addr.SubUint(uint(math.Abs(float64(instruction.Off0))))
	} else {
		return base_addr.AddUint(uint(instruction.Off0))
	}

}
```
*ComputeOp0Addr*

The process is similar to compute the dst addr

```go
func (run_context RunContext) ComputeOp0Addr(instruction Instruction) (memory.Relocatable, error) {
	var base_addr memory.Relocatable
	switch instruction.Op0Reg {
	case AP:
		base_addr = run_context.Ap
	case FP:
		base_addr = run_context.Fp
	}

	if instruction.Off1 < 0 {
		return base_addr.SubUint(uint(math.Abs(float64(instruction.Off1))))
	} else {
		return base_addr.AddUint(uint(instruction.Off1))
	}
}

```

*ComputeOp1Addr*

It computes the address of `Op1` based on  the `Op0` register and the kind of Address the instruction has for `Op1`, If its address is `Op1SrcFp` it calcualtes the direction from Fp register, if it is `Op1SrcAp` then if calculates it if from Ap register. If it is an immediate then checks if the offset 2 is 1 and calculates it from the `Pc`. Finally if its an `Op1SrcOp0` it checks the `Op0` and calculates the direction from it. Then if performs and addition or a substraction if the `Off2` is negative or positive.

```go
func (run_context RunContext) ComputeOp1Addr(instruction Instruction, op0 *memory.MaybeRelocatable) (memory.Relocatable, error) {
	var base_addr memory.Relocatable

	switch instruction.Op1Addr {
	case Op1SrcFP:
		base_addr = run_context.Fp
	case Op1SrcAP:
		base_addr = run_context.Ap
	case Op1SrcImm:
		if instruction.Off2 == 1 {
			base_addr = run_context.Pc
		} else {
			base_addr = memory.NewRelocatable(0, 0)
			return memory.Relocatable{}, &VirtualMachineError{Msg: "UnknownOp0"}
		}
	case Op1SrcOp0:
		if op0 == nil {
			return memory.Relocatable{}, errors.New("Unknown Op0")
		}
		rel, is_rel := op0.GetRelocatable()
		if is_rel {
			base_addr = rel
		} else {
			return memory.Relocatable{}, errors.New("AddressNotRelocatable")
		}
	}

	if instruction.Off2 < 0 {
		return base_addr.SubUint(uint(math.Abs(float64(instruction.Off2))))
	} else {
		return base_addr.AddUint(uint(instruction.Off2))
	}
}
```

*DeduceOp0*

deduces the value of `Op0` if possible (based on `dst` and `Op1`) otherwise it returns a nil.
if res is deduced in th process returns its deduced value as well

```go
func (vm *VirtualMachine) DeduceOp0(instruction *Instruction, dst *memory.MaybeRelocatable, op1 *memory.MaybeRelocatable) (deduced_op0 *memory.MaybeRelocatable, deduced_res *memory.MaybeRelocatable, error error) {
	switch instruction.Opcode {
	case Call:
		deduced_op0 := vm.RunContext.Pc
		deduced_op0.Offset += instruction.Size()
		return memory.NewMaybeRelocatableRelocatable(deduced_op0), nil, nil
	case AssertEq:
		switch instruction.ResLogic {
		case ResAdd:
			if dst != nil && op1 != nil {
				deduced_op0, err := dst.Sub(*op1)
				if err != nil {
					return nil, nil, err
				}
				return &deduced_op0, dst, nil
			}
		case ResMul:
			if dst != nil && op1 != nil {
				dst_felt, dst_is_felt := dst.GetFelt()
				op1_felt, op1_is_felt := op1.GetFelt()
				if dst_is_felt && op1_is_felt && !op1_felt.IsZero() {
					return memory.NewMaybeRelocatableFelt(dst_felt.Div(op1_felt)), dst, nil

				}
			}
		}
	}
	return nil, nil, nil
}
```

*DeduceOp1*

Like `DeducedOp0` it tries to deduce the value of op1 if it is possible (based on dst and op0), it does so only if the Opcode is and AssertEq otherwise it returns a nil.
if res is deduced in th process returns its deduced value as well

```go
func (vm *VirtualMachine) DeduceOp1(instruction Instruction, dst *memory.MaybeRelocatable, op0 *memory.MaybeRelocatable) (*memory.MaybeRelocatable, *memory.MaybeRelocatable, error) {
	if instruction.Opcode == AssertEq {
		switch instruction.ResLogic {
		case ResOp1:
			return dst, dst, nil
		case ResAdd:
			if op0 != nil && dst != nil {
				dst_rel, err := dst.Sub(*op0)
				if err != nil {
					return nil, nil, err
				}
				return &dst_rel, dst, nil
			}
		case ResMul:
			dst_felt, dst_is_felt := dst.GetFelt()
			op0_felt, op0_is_felt := op0.GetFelt()
			if dst_is_felt && op0_is_felt && !op0_felt.IsZero() {
				res := memory.NewMaybeRelocatableFelt(dst_felt.Div(op0_felt))
				return res, dst, nil
			}
		}
	}
	return nil, nil, nil
}

```
*ComputeRes*

If the Res value has not been deduced in the previous steps then it is computed based on the `Op0` and `Op1` values and the res logic stored in the instruction.
 if the value is `Unscontrained` then a nil is returned. 

```go
func (vm *VirtualMachine) ComputeRes(instruction Instruction, op0 memory.MaybeRelocatable, op1 memory.MaybeRelocatable) (*memory.MaybeRelocatable, error) {
	switch instruction.ResLogic {
	case ResOp1:
		return &op1, nil

	case ResAdd:
		maybe_rel, err := op0.Add(op1)
		if err != nil {
			return nil, err
		}
		return &maybe_rel, nil

	case ResMul:
		num_op0, m_type := op0.GetFelt()
		num_op1, other_type := op1.GetFelt()
		if m_type && other_type {
			result := memory.NewMaybeRelocatableFelt(num_op0.Mul(num_op1))
			return result, nil
		} else {
			return nil, errors.New("ComputeResRelocatableMul")
		}

	case ResUnconstrained:
		return nil, nil
	}
	return nil, nil
}
```

*DeduceDst*

if the destination value has not been calculated before then it is deduced based on the Res value. If the opcode is an `AssertEq` then dst is similar to res.
If it is a `Call` then is value is calculated in base of the `Fp` register 

```go
func (vm *VirtualMachine) DeduceDst(instruction Instruction, res *memory.MaybeRelocatable) *memory.MaybeRelocatable {
	switch instruction.Opcode {
	case AssertEq:
		return res
	case Call:
		return memory.NewMaybeRelocatableRelocatable(vm.RunContext.Fp)

	}
	return nil
}
```


### CairoRunner

Now that can can execute cairo steps, lets look at the VM's initialization step.
We will begin by creating our `CairoRunnerc`:

```go
type CairoRunner struct {
	Program       vm.Program
	Vm            vm.VirtualMachine
	ProgramBase   memory.Relocatable
	executionBase memory.Relocatable
	initialPc     memory.Relocatable
	initialAp     memory.Relocatable
	initialFp     memory.Relocatable
	finalPc       memory.Relocatable
	mainOffset    uint
}

func NewCairoRunner(program vm.Program) *CairoRunner {
    // TODO line below is fake
	main_offset := program.identifiers["__main__.main"]
	return &CairoRunner{Program: program, Vm: *vm.NewVirtualMachine(), mainOffset: main_offset}

}
```

Now we will create our `Initialize` method step by step:

```go
// Performs the initialization step, returns the end pointer (pc upon which execution should stop)
func (r *CairoRunner) Initialize() (memory.Relocatable, error) {
	r.initializeSegments()
	end, err := r.initializeMainEntrypoint()
	r.initializeVM()
	return end, err
}
```

*InitializeSegments*

This method will create our program and execution segments

```go
func (r *CairoRunner) initializeSegments() {
	// Program Segment
	r.ProgramBase = r.Vm.Segments.AddSegment()
	// Execution Segment
	r.executionBase = r.Vm.Segments.AddSegment()
}
```

*initializeMainEntrypoint*

This method will initialize the memory and initial register values to begin execution from the main entrypoint, and return the final pc

```go
func (r *CairoRunner) initializeMainEntrypoint() (memory.Relocatable, error) {
	stack := make([]memory.MaybeRelocatable, 0, 2)
	return_fp := r.Vm.Segments.AddSegment()
	return r.initializeFunctionEntrypoint(r.mainOffset, &stack, return_fp)
}
```

*initializeFunctionEntrypoint*

This method will initialize the memory and initial register values to execute a cairo function given its offset within the program segment (aka entrypoint) and return the final pc. In our case, this function will be the main entrypoint, but later on we will be able to use this method to run starknet contract entrypoints. 
The stack will then be loaded into the execution segment in the next method. For now, the stack will be empty, but later on it will contain the builtin bases (which are the arguments for the main function), and the function arguments when running a function from a starknet contract.

```go
func (r *CairoRunner) initializeFunctionEntrypoint(entrypoint uint, stack *[]memory.MaybeRelocatable, return_fp memory.Relocatable) (memory.Relocatable, error) {
	end := r.Vm.Segments.AddSegment()
	*stack = append(*stack, *memory.NewMaybeRelocatableRelocatable(end), *memory.NewMaybeRelocatableRelocatable(return_fp))
	r.initialFp = r.executionBase
	r.initialFp.Offset += uint(len(*stack))
	r.initialAp = r.initialFp
	r.finalPc = end
	return end, r.initializeState(entrypoint, stack)
}
```

*InitializeState*

This method will be in charge of loading the program data into the program segment and the stack into the execution segment

```go
func (r *CairoRunner) initializeState(entrypoint uint, stack *[]memory.MaybeRelocatable) error {
	r.initialPc = r.ProgramBase
	r.initialPc.Offset += entrypoint
	// Load program data
	_, err := r.Vm.Segments.LoadData(r.ProgramBase, &r.Program.Data)
	if err == nil {
		_, err = r.Vm.Segments.LoadData(r.executionBase, stack)
	}
	return err
}
```

*initializeVm*

This method will set the values of the VM's `RunContext` with our `CairoRunner`'s initial values

```go
func (r *CairoRunner) initializeVM() {
	r.Vm.RunContext.Ap = r.initialAp
	r.Vm.RunContext.Fp = r.initialFp
	r.Vm.RunContext.Pc = r.initialPc
}
```

With `CairoRunner.Initialize()` now complete we can move on to the execution step:

*RunUntilPc*

This method will continuously execute cairo steps until the end pc, returned by 'CairoRunner.Initialize()' is reached

```go
    //TODO
```

*Step*

```go
    //TODO
```

*Decode instruction*

```go
    //TODO
```

*Run instruction*

```go
    //TODO
```

*Opcode assertions*

Once we have the instruction's operands to work with, we have to ensure the correctness of them. The first thing we need to differentiate is which type of instruction are we running, we do this by looking at the instruction's opcode. 

The posible opcodes we want to perform assertions on are: 
	1. AssertEq instruction 
	2. Call instruction 

In the first option, we need to ensure the result operand is not null (nil in this case) and also that the result operand is equal to the dst operand. If any of those things fail, we throw an error. 

On the other hand, the Call instruction, what we do first is define our return pc register, we do that adding the size of the instruction to the current pc. Then, we check our operand op0 is equal to the return pc and our dst operand is the same as the return fp register. If any of those things fail, we throw an error. 

If this method returns a nil error, it means operands were computed correctly and we are good to go!

```go
func (vm *VirtualMachine) OpcodeAssertions(instruction Instruction, operands Operands) error {
	switch instruction.Opcode {
	case AssertEq:
		if operands.Res == nil {
			return &VirtualMachineError{"UnconstrainedResAssertEq"}
		}
		if !operands.Res.IsEqual(&operands.Dst) {
			return &VirtualMachineError{"DiffAssertValues"}
		}
	case Call:
		new_rel, err := vm.RunContext.Pc.AddUint(instruction.Size())
		if err != nil {
			return err
		}
		returnPC := memory.NewMaybeRelocatableRelocatable(new_rel)

		if !operands.Op0.IsEqual(returnPC) {
			return &VirtualMachineError{"CantWriteReturnPc"}
		}

		returnFP := vm.RunContext.Fp
		dstRelocatable, _ := operands.Dst.GetRelocatable()
		if !returnFP.IsEqual(&dstRelocatable) {
			return &VirtualMachineError{"CantWriteReturnFp"}
		}
	}

	return nil
}
```

*Update registers*

```go
    //TODO
```


Once we are done executing, we can relocate our memory & trace and output them into files

### Memory Relocation
TODO

### Builtins

Now that we are able to run a basic fibonacci program, lets step up our game by adding builtins to our VM. A builtin is a low level optimization integrated into the core loop of the VM that allows otherwise expensive computation to be performed more efficiently. Builtins have two ways to operate: via validation rules and via auto-deduction rules. Validation rules are applied to every element that is inserted into a builtin's segment. For example, if I want to verify an ecdsa signature, I can insert it into the ecdsa builtin's segment and let a validation rule take care of verifying the signature. Auto-deduction rules take over during instruction execution, when we can't compute the value of an operand who's address belongs to a builtin segment, we can use that builtin's auto-deduction rule to calculate the value of the operand. For example, If I want to calculate the pedersen hash of two values, I can write the values into the pedersen builtin's segment and then ask for the next memory cell, without builtins, this instruction would have failed, as there is no value stored in that cell, but now we can use auto-deduction rules to calculate the hash and fill in that memory cell.

We will define a basic interface to generalize all of our builtin's behaviour:

```go
type BuiltinRunner interface {
	// Returns the first address of the builtin's memory segment
	Base() memory.Relocatable
	// Returns the name of the builtin
	Name() string
	// Creates a memory segment for the builtin and initializes its base
	InitializeSegments(*memory.MemorySegmentManager)
	// Returns the builtin's initial stack
	InitialStack() []memory.MaybeRelocatable
	// Attempts to deduce the value of a memory cell given by its address. Can return either a nil pointer and an error, if an error arises during the deduction,
	// a valid pointer and nil if the deduction was succesful, or a nil pointer and nil if there is no deduction for the memory cell
	DeduceMemoryCell(memory.Relocatable, *memory.Memory) (*memory.MaybeRelocatable, error)
	// Adds a validation rule to the memory
	// Validation rules are applied when a value is inserted into the builtin's segment
	AddValidationRule(*memory.Memory)
}
```

And now lets integrate this into our existing codebase:

First we will make some modifications to our basic structures:

We will add our builtin runners to the VM:

```go
type VirtualMachine struct {
	RunContext     RunContext
	currentStep    uint
	Segments       memory.MemorySegmentManager
	BuiltinRunners []builtins.BuiltinRunner
}
```

Then we will create two new types to handle validation rules in the `Memory`:
 
*ValidationRule* 

This will represent our builtin's validation rules, they take a memory address and a referenece to the memory, and return a list of validated addresses, for most builtins, this list will contain the address it received if the validation was succesful, but some builtins may return additional addresses.

```go
// A function that validates a memory address and returns a list of validated addresses
type ValidationRule func(*Memory, Relocatable) ([]Relocatable, error)
```

*AddressSet*

As go doesn't have a set type, we created our own really basic set for `Relocatable`s. This will hold the values returned by the validation rules, so that we don't have to run them more than once for each memory cell.

```go
// A Set to store Relocatable values
type AddressSet map[Relocatable]bool

func NewAddressSet() AddressSet {
	return make(map[Relocatable]bool)
}

func (set AddressSet) Add(element Relocatable) {
	set[element] = true
}

func (set AddressSet) Contains(element Relocatable) bool {
	return set[element]
}
```

And we will add them to our `Memory` stuct:

``` go
type Memory struct {
	data                map[Relocatable]MaybeRelocatable
	num_segments        uint
	validation_rules    map[uint]ValidationRule
	validated_addresses AddressSet
}
```

Now we only need to add a way to create this validation rules:

```go
// Adds a validation rule for a given segment
func (m *Memory) AddValidationRule(segment_index uint, rule ValidationRule) {
	m.validation_rules[segment_index] = rule
}
```

And a method that runs validations on a memory address:

```go
// Applies the validation rule for the addr's segment if any
// Skips validation if the address is temporary or if it has been previously validated
func (m *Memory) validateAddress(addr Relocatable) error {
	if addr.SegmentIndex < 0 || m.validated_addresses.Contains(addr) {
		return nil
	}
	rule, ok := m.validation_rules[uint(addr.SegmentIndex)]
	if !ok {
		return nil
	}
	validated_addresses, error := rule(m, addr)
	if error != nil {
		return error
	}
	for _, validated_address := range validated_addresses {
		m.validated_addresses.Add(validated_address)
	}
	return nil
}
```

And we are all set to integrate this new logic into our `Memory`'s `Insert` operation:

```go
// Inserts a value in some memory address, given by a Relocatable value.
func (m *Memory) Insert(addr Relocatable, val *MaybeRelocatable) error {
	// Check that insertions are preformed within the memory bounds
	if addr.SegmentIndex >= int(m.num_segments) {
		return errors.New("Error: Inserting into a non allocated segment")
	}
	// Check for possible overwrites
	prev_elem, ok := m.data[addr]
	if ok && prev_elem != *val {
		return errors.New("Memory is write-once, cannot overwrite memory value")
	}

	m.data[addr] = *val

	return m.validateAddress(addr)
}
```

Now we will initialize the builtins from our `CairoRunner`:

*NewCairoRunner*

Here we will have to iterate over the `Builtins` field of the `Program`, and add the corresponding builtin to the `VirtualMachine`'s `BuiltinRunner` field. We don't have any builtins yet, so we wil add a comment as placeholder and just leave a default case. As we implement more builtins, we will add a case for each of them.

```go
func NewCairoRunner(program vm.Program) (*CairoRunner, error) {
	main_offset := program.identifiers["__main__.main"]
	runner := CairoRunner{Program: program, Vm: *vm.NewVirtualMachine(), mainOffset: main_offset}
	for _, builtin_name := range program.Builtins {
		switch builtin_name {
		// Add a case for each builtin here, example:
		// case "range_check":
		// 	runner.Vm.BuiltinRunners = append(runner.Vm.BuiltinRunners, RangeCheckBuiltin{})
		default:
			return nil, errors.New("Invalid builtin")
		}
	}
	return &runner, nil
}
```

*initializeSegments*

Here we will also initialize the builtin segments by calling each builtin's `InitializeSegments` method

```go
func (r *CairoRunner) initializeSegments() {
	// Program Segment
	r.ProgramBase = r.Vm.Segments.AddSegment()
	// Execution Segment
	r.executionBase = r.Vm.Segments.AddSegment()
	// Builtin Segments
	for i := range r.Vm.BuiltinRunners {
		r.Vm.BuiltinRunners[i].InitializeSegments(&r.Vm.Segments)
	}
}
```

*InitializeMainEntryPoint*

Here we will add the builtin's initial_stack to our stack. The builtin's initial_stack is generally made up of the builtin's base, and is what allows the main function to write into the builtin's segment.

```go
func (r *CairoRunner) initializeMainEntrypoint() (memory.Relocatable, error) {
	// When running from main entrypoint, only up to 11 values will be written (9 builtin bases + end + return_fp)
	stack := make([]memory.MaybeRelocatable, 0, 11)
	// Append builtins initial stack to stack
	for i := range r.Vm.BuiltinRunners {
		for _, val := range r.Vm.BuiltinRunners[i].InitialStack() {
			stack = append(stack, val)
		}
	}
	return_fp := r.Vm.Segments.AddSegment()
	return r.initializeFunctionEntrypoint(r.mainOffset, &stack, return_fp)
}
```

*initializeVm*

Here we will add our builtin's validation rules to the `Memory` and use them to validate the meory cells we loaded before

```go
func (r *CairoRunner) initializeVM() error {
	r.Vm.RunContext.Ap = r.initialAp
	r.Vm.RunContext.Fp = r.initialFp
	r.Vm.RunContext.Pc = r.initialPc
	// Add validation rules
	for i := range r.Vm.BuiltinRunners {
		r.Vm.BuiltinRunners[i].AddValidationRule(&r.Vm.Segments.Memory)
	}
	// Apply validation rules to memory
	return r.Vm.Segments.Memory.ValidateExistingMemory()
}
```

For this we will add the method `Memory.ValidateExistingMemory`:

```go
func (m *Memory) ValidateExistingMemory() error {
	for addr := range m.data {
		err := m.validateAddress(addr)
		if err != nil {
			return err
		}
	}
	return nil
}
```


[TODO (for builtins section): BuiltinRunners in Compute Operands]

[Next sections: Implementing each builtin runner]

### Hints

TODO
