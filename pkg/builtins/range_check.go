package builtins

import (
	"math"

	"github.com/lambdaclass/cairo-vm.go/pkg/lambdaworks"
	"github.com/lambdaclass/cairo-vm.go/pkg/utils"
	"github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
	"github.com/pkg/errors"
)

const RANGE_CHECK_BUILTIN_NAME = "range_check"
const INNER_RC_BOUND_SHIFT = 16
const INNER_RC_BOUND_MASK = math.MaxUint16
const CELLS_PER_RANGE_CHECK = 1

const N_PARTS = 8

func RangeCheckError(err error) error {
	return errors.Wrapf(err, "Range check error")
}

func OutsideBoundsError(felt lambdaworks.Felt) error {
	return RangeCheckError(errors.Errorf("Value %d is out of bounds [0, 2^128]", felt))
}

func NotAFeltError(addr memory.Relocatable, val memory.MaybeRelocatable) error {
	rel, _ := val.GetRelocatable()
	return RangeCheckError(errors.Errorf("Value %d found in %d is not a field element", rel, addr))
}

type RangeCheckBuiltinRunner struct {
	base                  memory.Relocatable
	included              bool
	ratio                 uint
	instancesPerComponent uint
}

func NewRangeCheckBuiltinRunner(ratio uint) *RangeCheckBuiltinRunner {
	return &RangeCheckBuiltinRunner{ratio: ratio, instancesPerComponent: 1}
}

func DefaultRangeCheckBuiltinRunner() *RangeCheckBuiltinRunner {
	return NewRangeCheckBuiltinRunner(8)
}

func (r *RangeCheckBuiltinRunner) Base() memory.Relocatable {
	return r.base
}

func (r *RangeCheckBuiltinRunner) Name() string {
	return RANGE_CHECK_BUILTIN_NAME
}

func (r *RangeCheckBuiltinRunner) InitializeSegments(segments *memory.MemorySegmentManager) {
	r.base = segments.AddSegment()
}

func (r *RangeCheckBuiltinRunner) InitialStack() []memory.MaybeRelocatable {
	if r.included {
		return []memory.MaybeRelocatable{*memory.NewMaybeRelocatableRelocatable(r.base)}
	}
	return []memory.MaybeRelocatable{}
}

func (r *RangeCheckBuiltinRunner) DeduceMemoryCell(addr memory.Relocatable, mem *memory.Memory) (*memory.MaybeRelocatable, error) {
	return nil, nil
}

func RangeCheckValidationRule(mem *memory.Memory, address memory.Relocatable) ([]memory.Relocatable, error) {
	res_val, err := mem.Get(address)
	if err != nil {
		return nil, err
	}
	felt, is_felt := res_val.GetFelt()
	if !is_felt {
		return nil, NotAFeltError(address, *res_val)
	}
	if felt.Bits() <= N_PARTS*INNER_RC_BOUND_SHIFT {
		return []memory.Relocatable{address}, nil
	}
	return nil, OutsideBoundsError(felt)
}

func (r *RangeCheckBuiltinRunner) AddValidationRule(mem *memory.Memory) {
	mem.AddValidationRule(uint(r.base.SegmentIndex), RangeCheckValidationRule)
}

func (r *RangeCheckBuiltinRunner) Include(include bool) {
	r.included = include
}

func (r *RangeCheckBuiltinRunner) Ratio() uint {
	return r.ratio
}

func (r *RangeCheckBuiltinRunner) CellsPerInstance() uint {
	return CELLS_PER_RANGE_CHECK
}

func (r *RangeCheckBuiltinRunner) GetAllocatedMemoryUnits(segments *memory.MemorySegmentManager, currentStep uint) (uint, error) {
	// This condition corresponds to an uninitialized ratio for the builtin, which should only
	// happen when layout is `dynamic`
	if r.Ratio() == 0 {
		// Dynamic layout has the exact number of instances it needs (up to a power of 2).
		used, err := segments.GetSegmentUsedSize(uint(r.base.SegmentIndex))
		if err != nil {
			return 0, err
		}
		instances := used / r.CellsPerInstance()
		components := utils.NextPowOf2(instances / r.instancesPerComponent)
		size := r.CellsPerInstance() * r.instancesPerComponent * components

		return size, nil
	}

	minStep := r.ratio * r.instancesPerComponent
	if currentStep < minStep {
		return 0, errors.Errorf("number of steps must be at least %d for the %s builtin", minStep, r.Name())
	}
	value, err := utils.SafeDiv(currentStep, r.ratio)

	if err != nil {
		return 0, errors.Errorf("error calculating builtin memory units: %s", err)
	}

	return r.CellsPerInstance() * value, nil
}

func (r *RangeCheckBuiltinRunner) GetUsedCellsAndAllocatedSizes(segments *memory.MemorySegmentManager, currentStep uint) (uint, uint, error) {
	used, err := segments.GetSegmentUsedSize(uint(r.base.SegmentIndex))
	if err != nil {
		return 0, 0, err
	}

	size, err := r.GetAllocatedMemoryUnits(segments, currentStep)

	if used > size {
		return 0, 0, errors.Errorf("The builtin %s used %d cells but the capacity is %d", r.Name(), used, size)
	}

	return used, size, nil
}
