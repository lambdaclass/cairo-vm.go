package hint_utils_test

import (
	"testing"

	. "github.com/lambdaclass/cairo-vm.go/pkg/hints/hint_utils"
	"github.com/lambdaclass/cairo-vm.go/pkg/lambdaworks"
	"github.com/lambdaclass/cairo-vm.go/pkg/vm"
	"github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
)

func TestIdsManagerGetAddressSimpleReference(t *testing.T) {
	ids := IdsManager{
		References: map[string]HintReference{
			"val": {
				Offset1: OffsetValue{
					Register:  vm.FP,
					ValueType: Reference,
				},
			},
		},
	}
	vm := vm.NewVirtualMachine()
	addr, err := ids.GetAddr("val", vm)
	if err != nil {
		t.Errorf("Error in test: %s", err)
	}
	if addr != vm.RunContext.Fp {
		t.Errorf("IdsManager.GetAddr returned wrong value")
	}
}

func TestIdsManagerGetAddressUnknownIdentifier(t *testing.T) {
	ids := IdsManager{
		References: map[string]HintReference{
			"value": {
				Offset1: OffsetValue{
					Register: vm.FP,
				},
			},
		},
	}
	vm := vm.NewVirtualMachine()
	_, err := ids.GetAddr("val", vm)
	if err == nil {
		t.Errorf("IdsManager.GetAddress should have failed")
	}
}

func TestIdsManagerGetAddressComplexReferenceDoubleDeref(t *testing.T) {
	// reference: [ap + 1] + [fp + 2] = (1, 0) + 3 = (1, 3)
	ids := IdsManager{
		References: map[string]HintReference{
			"val": {
				Offset1: OffsetValue{
					Register:    vm.AP,
					Value:       1,
					ValueType:   Reference,
					Dereference: true,
				},
				Offset2: OffsetValue{
					Register:    vm.FP,
					Value:       2,
					ValueType:   Reference,
					Dereference: true,
				},
			},
		},
	}
	vm := vm.NewVirtualMachine()
	vm.Segments.AddSegment()
	// [ap + 1] = (1, 0)
	vm.Segments.Memory.Insert(vm.RunContext.Ap.AddUint(1), memory.NewMaybeRelocatableRelocatable(memory.Relocatable{SegmentIndex: 1, Offset: 0}))
	// [fp + 2] = 3
	vm.Segments.Memory.Insert(vm.RunContext.Fp.AddUint(2), memory.NewMaybeRelocatableFelt(lambdaworks.FeltFromUint64(3)))
	addr, err := ids.GetAddr("val", vm)
	if err != nil {
		t.Errorf("Error in test: %s", err)
	}
	expected_addr := memory.Relocatable{SegmentIndex: 1, Offset: 3}
	if addr != expected_addr {
		t.Errorf("IdsManager.GetAddr returned wrong value")
	}
}

func TestIdsManagerGetAddressComplexReferenceOneDeref(t *testing.T) {
	// reference: [ap + 1] + 2 = (1, 0) + 2 = (1, 2)
	ids := IdsManager{
		References: map[string]HintReference{
			"val": {
				Offset1: OffsetValue{
					Register:    vm.AP,
					Value:       1,
					ValueType:   Reference,
					Dereference: true,
				},
				Offset2: OffsetValue{
					Value:     2,
					ValueType: Value,
				},
			},
		},
	}
	vm := vm.NewVirtualMachine()
	vm.Segments.AddSegment()
	// [ap + 1] = (1, 0)
	vm.Segments.Memory.Insert(vm.RunContext.Ap.AddUint(1), memory.NewMaybeRelocatableRelocatable(memory.Relocatable{SegmentIndex: 1, Offset: 0}))
	addr, err := ids.GetAddr("val", vm)
	if err != nil {
		t.Errorf("Error in test: %s", err)
	}
	expected_addr := memory.Relocatable{SegmentIndex: 1, Offset: 2}
	if addr != expected_addr {
		t.Errorf("IdsManager.GetAddr returned wrong value")
	}
}
