package vm

import (
	"errors"

	"github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
)

// VirtualMachine represents the Cairo VM.
// Runs Cairo assembly and produces an execution trace.
type VirtualMachine struct {
	RunContext  RunContext
	currentStep uint
	Segments    memory.MemorySegmentManager
}

func NewVirtualMachine() *VirtualMachine {
	return &VirtualMachine{Segments: *memory.NewMemorySegmentManager()}
}

type Operands struct {
	Dst memory.MaybeRelocatable
	Res *memory.MaybeRelocatable
	Op0 memory.MaybeRelocatable
	Op1 memory.MaybeRelocatable
}

// Updates the value of PC according to the executed instruction
func (vm *VirtualMachine) UpdatePc(instruction *Instruction, operands *Operands) error {
	switch instruction.PcUpdate {
	case PcUpdateRegular:
		vm.RunContext.Pc.Offset += instruction.Size()
	case PcUpdateJump:
		if operands.Res == nil {
			return errors.New("Res.UNCONSTRAINED cannot be used with PcUpdate.JUMP")
		}
		res, ok := operands.Res.GetRelocatable()
		if !ok {
			return errors.New("An integer value as Res cannot be used with PcUpdate.JUMP")
		}
		vm.RunContext.Pc = res
	case PcUpdateJumpRel:
		if operands.Res == nil {
			return errors.New("Res.UNCONSTRAINED cannot be used with PcUpdate.JUMP_REL")
		}
		res, ok := operands.Res.GetInt()
		if !ok {
			return errors.New("A relocatable value as Res cannot be used with PcUpdate.JUMP_REL")
		}
		new_pc, err := vm.RunContext.Pc.AddFelt(res)
		if err != nil {
			return err
		}
		vm.RunContext.Pc = new_pc
	case PcUpdateJnz:
		if operands.Dst.IsZero() {
			vm.RunContext.Pc.Offset += instruction.Size()
		} else {
			new_pc, err := vm.RunContext.Pc.AddMaybeRelocatable(operands.Op1)
			if err != nil {
				return err
			}
			vm.RunContext.Pc = new_pc
		}

	}
	return nil
}

// Updates the value of AP according to the executed instruction
func (vm *VirtualMachine) UpdateAp(instruction *Instruction, operands *Operands) error {
	switch instruction.ApUpdate {
	case ApUpdateAdd:
		if operands.Res == nil {
			return errors.New("Res.UNCONSTRAINED cannot be used with ApUpdate.ADD")
		}
		new_ap, err := vm.RunContext.Ap.AddMaybeRelocatable(*operands.Res)
		if err != nil {
			return err
		}
		vm.RunContext.Ap = new_ap
	case ApUpdateAdd1:
		vm.RunContext.Ap.Offset += 1
	case ApUpdateAdd2:
		vm.RunContext.Ap.Offset += 2
	}
	return nil
}

// Updates the value of FP according to the executed instruction
func (vm *VirtualMachine) UpdateFp(instruction *Instruction, operands *Operands) error {
	switch instruction.FpUpdate {
	case FpUpdateAPPlus2:
		vm.RunContext.Fp.Offset = vm.RunContext.Ap.Offset + 2
	case FpUpdateDst:
		rel, ok := operands.Dst.GetRelocatable()
		if ok {
			vm.RunContext.Fp = rel
		} else {
			felt, _ := operands.Dst.GetInt()
			new_fp, err := vm.RunContext.Fp.AddFelt(felt)
			if err != nil {
				return err
			}
			vm.RunContext.Fp = new_fp
		}
	}
	return nil
}
