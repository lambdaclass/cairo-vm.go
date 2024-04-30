package hints_test

import (
	"testing"

	"github.com/lambdaclass/cairo-vm.go/pkg/builtins"
	. "github.com/lambdaclass/cairo-vm.go/pkg/hints"
	. "github.com/lambdaclass/cairo-vm.go/pkg/hints/hint_codes"
	. "github.com/lambdaclass/cairo-vm.go/pkg/hints/hint_utils"
	"github.com/lambdaclass/cairo-vm.go/pkg/lambdaworks"
	. "github.com/lambdaclass/cairo-vm.go/pkg/lambdaworks"
	. "github.com/lambdaclass/cairo-vm.go/pkg/types"
	. "github.com/lambdaclass/cairo-vm.go/pkg/vm"
	"github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
	. "github.com/lambdaclass/cairo-vm.go/pkg/vm/memory"
)

func TestAssertNNHintOk(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromUint64(17))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NN,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("ASSERT_NN hint test failed with error %s", err)
	}
}

func TestAssertNNHintFail(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromDecString("-1"))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NN,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_NN hint should have failed")
	}
}

func TestIsPositiveOkPositive(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value":       {NewMaybeRelocatableFelt(FeltFromUint64(17))},
			"is_positive": {nil},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: IS_POSITIVE,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("IS_POSITIVE hint test failed with error %s", err)
	}
	// Check ids.is_positive
	is_positive, err := idsManager.GetFelt("is_positive", vm)
	if err != nil || is_positive != FeltFromUint64(1) {
		t.Errorf("IS_POSITIVE hint test incorrect value for ids.is_positive")
	}
}

func TestIsPositiveOkNegative(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value":       {NewMaybeRelocatableFelt(FeltFromDecString("-1"))},
			"is_positive": {nil},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: IS_POSITIVE,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("IS_POSITIVE hint test failed with error %s", err)
	}
	// Check ids.is_positive
	is_positive, err := idsManager.GetFelt("is_positive", vm)
	if err != nil || is_positive != FeltFromUint64(0) {
		t.Errorf("IS_POSITIVE hint test incorrect value for ids.is_positive")
	}
}

func TestIsPositiveOutOfRange(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value":       {NewMaybeRelocatableFelt(FeltFromDecString("340282366920938463463374607431768211457"))},
			"is_positive": {nil},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: IS_POSITIVE,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("IS_POSITIVE hint test should have failed")
	}
}
func TestAssertNotZeroHintOk(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(17))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_ZERO,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("ASSERT_NOT_ZERO hint test failed with error %s", err)
	}
}

func TestAssertNotZeroHintFail(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(0))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_ZERO,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_NOT_ZERO hint should have failed")
	}
}

func TestVerifyValidSignature(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	signature_builtin := builtins.NewSignatureBuiltinRunner(2048)
	vm.BuiltinRunners = append(vm.BuiltinRunners, signature_builtin)

	hintProcessor := CairoVmHintProcessor{}
	vm.Segments.AddSegment()

	r_felt := lambdaworks.FeltFromDecString("3086480810278599376317923499561306189851900463386393948998357832163236918254")
	s_felt := lambdaworks.FeltFromDecString("598673427589502599949712887611119751108407514580626464031881322743364689811")
	r := memory.NewMaybeRelocatableFelt(r_felt)
	s := memory.NewMaybeRelocatableFelt(s_felt)

	vm.RunContext.Fp = memory.NewRelocatable(1, 3)

	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"ecdsa_ptr":   {nil},
			"signature_r": {r},
			"signature_s": {s},
		},
		vm,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: VERIFY_ECDSA_SIGNATURE,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)

	if err != nil {
		t.Errorf("Verify signature hint for correct signature failed with error: %s", err)
	}
}

func TestVerifySignatureInvalidEcdsaPointer(t *testing.T) {
	vm := NewVirtualMachine()
	signature_builtin := builtins.NewSignatureBuiltinRunner(2048)
	vm.BuiltinRunners = append(vm.BuiltinRunners, signature_builtin)

	hintProcessor := CairoVmHintProcessor{}
	vm.Segments.AddSegment()

	r_felt := lambdaworks.FeltFromDecString("3086480810278599376317923499561306189851900463386393948998357832163236918254")
	s_felt := lambdaworks.FeltFromDecString("598673427589502599949712887611119751108407514580626464031881322743364689811")
	three := memory.NewMaybeRelocatableFelt(lambdaworks.FeltFromUint64(3))
	r := memory.NewMaybeRelocatableFelt(r_felt)
	s := memory.NewMaybeRelocatableFelt(s_felt)

	vm.RunContext.Fp = memory.NewRelocatable(1, 3)

	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"ecdsa_ptr":   {three},
			"signature_r": {r},
			"signature_s": {s},
		},
		vm,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: VERIFY_ECDSA_SIGNATURE,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)

	if err == nil {
		t.Errorf("Verified a signature with an invalid pointer")
	}
}

func TestAssertNotEqualHintNonComparableDiffType(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromUint64(0))},
			"b": {NewMaybeRelocatableRelocatable(NewRelocatable(0, 0))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_EQUAL,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_NOT_EQUAL hint should have failed")
	}
}

func TestAssertNotEqualHintNonComparableDiffIndex(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableRelocatable(NewRelocatable(1, 0))},
			"b": {NewMaybeRelocatableRelocatable(NewRelocatable(0, 0))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_EQUAL,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_NOT_EQUAL hint should have failed")
	}
}

func TestAssertNotEqualHintEqualRelocatables(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableRelocatable(NewRelocatable(0, 0))},
			"b": {NewMaybeRelocatableRelocatable(NewRelocatable(0, 0))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_EQUAL,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_NOT_EQUAL hint should have failed")
	}
}

func TestAssertNotEqualHintEqualFelts(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromUint64(9))},
			"b": {NewMaybeRelocatableFelt(FeltFromUint64(9))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_EQUAL,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_NOT_EQUAL hint should have failed")
	}
}

func TestAssertNotEqualHintOkFelts(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromUint64(9))},
			"b": {NewMaybeRelocatableFelt(FeltFromUint64(7))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_EQUAL,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("ASSERT_NOT_EQUAL hint failed with error: %s", err)
	}
}

func TestAssertNotEqualHintOkRelocatables(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableRelocatable(NewRelocatable(1, 9))},
			"b": {NewMaybeRelocatableRelocatable(NewRelocatable(1, 7))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_NOT_EQUAL,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("ASSERT_NOT_EQUAL hint failed with error: %s", err)
	}
}

func TestSqrtOk(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromDecString("9"))},
			"root":  {nil},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SQRT,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("SQRT hint failed with error: %s", err)
	}

	root, err := idsManager.GetFelt("root", vm)
	if err != nil {
		t.Errorf("failed to get root: %s", err)
	}
	if root != FeltFromUint64(3) {
		t.Errorf("Expected sqrt(9) == 3. Got: %v", root)
	}
}

func TestUnsignedDivRemHintSuccess(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"div":   {NewMaybeRelocatableFelt(FeltFromDecString("7"))},
			"value": {NewMaybeRelocatableFelt(FeltFromDecString("15"))},
			"r":     {nil},
			"q":     {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: UNSIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("UNSIGNED_DIV_REM hint failed with error: %s", err)
	}

	q, err := idsManager.GetFelt("q", vm)
	if err != nil {
		t.Errorf("failed to get `q`: %s", err)
	}

	r, err := idsManager.GetFelt("r", vm)
	if err != nil {
		t.Errorf("failed to get `r`: %s", err)
	}

	if q != FeltFromUint64(2) {
		t.Errorf("Expected q=3, got: %v", q)
	}

	if r != FeltFromUint64(1) {
		t.Errorf("Expected r=1, got: %v", r)
	}
}

func TestAssertLeFeltOk(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a":               {NewMaybeRelocatableFelt(FeltOne())},
			"b":               {NewMaybeRelocatableFelt(FeltFromUint64(2))},
			"range_check_ptr": {NewMaybeRelocatableRelocatable(NewRelocatable(1, 0))},
		},
		vm,
	)
	constants := SetupConstantsForTest(map[string]Felt{
		"PRIME_OVER_3_HIGH": FeltFromHex("4000000000000088000000000000001"),
		"PRIME_OVER_2_HIGH": FeltFromHex("2AAAAAAAAAAAAB05555555555555556"),
	},
		&idsManager,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_LE_FELT,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, scopes)
	if err != nil {
		t.Errorf("ASSERT_LE_FELT hint failed with error: %s", err)
	}
}

func TestAssertLeFeltExcluded0Zero(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	scopes.AssignOrUpdateVariable("excluded", int(0))
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Code: ASSERT_LE_FELT_EXCLUDED_0,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, scopes)
	if err != nil {
		t.Errorf("ASSERT_LE_FELT_EXCLUDED_0 hint test failed with error %s", err)
	}
	// Check the value of memory[ap]
	val, err := vm.Segments.Memory.GetFelt(vm.RunContext.Ap)
	if err != nil || !val.IsZero() {
		t.Error("Wrong/No value inserted into ap")
	}
}

func TestAssertLeFeltExcluded0One(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	scopes.AssignOrUpdateVariable("excluded", int(1))
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Code: ASSERT_LE_FELT_EXCLUDED_0,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, scopes)
	if err != nil {
		t.Errorf("ASSERT_LE_FELT_EXCLUDED_0 hint test failed with error %s", err)
	}
	// Check the value of memory[ap]
	val, err := vm.Segments.Memory.GetFelt(vm.RunContext.Ap)
	if err != nil || val != FeltOne() {
		t.Error("Wrong/No value inserted into ap")
	}
}

func TestAssertLeFeltExcluded1Zero(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	scopes.AssignOrUpdateVariable("excluded", int(1))
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Code: ASSERT_LE_FELT_EXCLUDED_1,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, scopes)
	if err != nil {
		t.Errorf("ASSERT_LE_FELT_EXCLUDED_1 hint test failed with error %s", err)
	}
	// Check the value of memory[ap]
	val, err := vm.Segments.Memory.GetFelt(vm.RunContext.Ap)
	if err != nil || !val.IsZero() {
		t.Error("Wrong/No value inserted into ap")
	}
}

func TestAssertLeFeltExcluded1One(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	scopes.AssignOrUpdateVariable("excluded", int(0))
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Code: ASSERT_LE_FELT_EXCLUDED_1,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, scopes)
	if err != nil {
		t.Errorf("ASSERT_LE_FELT_EXCLUDED_1 hint test failed with error %s", err)
	}
	// Check the value of memory[ap]
	val, err := vm.Segments.Memory.GetFelt(vm.RunContext.Ap)
	if err != nil || val != FeltOne() {
		t.Error("Wrong/No value inserted into ap")
	}
}

func TestAssertLeFeltExcluded2Ok(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	scopes.AssignOrUpdateVariable("excluded", int(2))
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Code: ASSERT_LE_FELT_EXCLUDED_2,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, scopes)
	if err != nil {
		t.Errorf("ASSERT_LE_FELT_EXCLUDED_2 hint test failed with error %s", err)
	}
}

func TestAssertLeFeltExcluded2Err(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	scopes := NewExecutionScopes()
	scopes.AssignOrUpdateVariable("excluded", int(0))
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Code: ASSERT_LE_FELT_EXCLUDED_2,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, scopes)
	if err == nil {
		t.Errorf("ASSERT_LE_FELT_EXCLUDED_2 hint test should have failed")
	}
}

func TestAssertLeFeltHintOk(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromUint64(17))},
			"b": {NewMaybeRelocatableFelt(FeltFromUint64(18))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_LT_FELT,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("ASSERT_LT_FELT hint test failed with error %s", err)
	}
}

func TestAssertLeFeltHintErr(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"a": {NewMaybeRelocatableFelt(FeltFromUint64(17))},
			"b": {NewMaybeRelocatableFelt(FeltFromUint64(16))},
		},
		vm,
	)
	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_LT_FELT,
	})
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("ASSERT_LT_FELT hint test should have failed")
	}
}
func TestAssert250BitHintSuccess(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(3))},
			"high":  {nil},
			"low":   {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	constants := SetupConstantsForTest(map[string]Felt{
		"UPPER_BOUND": lambdaworks.FeltFromUint64(10),
		"SHIFT":       lambdaworks.FeltFromUint64(1),
	},
		&idsManager,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_250_BITS,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, nil)
	if err != nil {
		t.Errorf("ASSERT_250_BIT hint failed with error %s", err)
	}

	high, err := idsManager.GetFelt("high", vm)
	if err != nil {
		t.Errorf("failed to get high: %s", err)
	}

	low, err := idsManager.GetFelt("low", vm)
	if err != nil {
		t.Errorf("failed to get low: %s", err)
	}

	if high != FeltFromUint64(3) {
		t.Errorf("Expected high == 3. Got: %v", high)
	}

	if low != FeltFromUint64(0) {
		t.Errorf("Expected low == 0. Got: %v", low)
	}
}

func TestUnsignedDivRemHintDivZeroError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			// This is the condition that should make the hint execution error.
			"div":   {NewMaybeRelocatableFelt(FeltFromDecString("0"))},
			"value": {NewMaybeRelocatableFelt(FeltFromDecString("15"))},
			"r":     {nil},
			"q":     {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: UNSIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Error("UNSIGNED_DIV_REM should have failed")
	}
}

func TestAssert250BitHintFail(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(20))},
			"high":  {nil},
			"low":   {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	constants := SetupConstantsForTest(map[string]Felt{
		"UPPER_BOUND": lambdaworks.FeltFromUint64(10),
		"SHIFT":       lambdaworks.FeltFromUint64(1),
	},
		&idsManager,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: ASSERT_250_BITS,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, nil)
	if err == nil {
		t.Errorf("ASSERT_250_BIT hint should have failed with Value outside of 250 bit error")
	}
}

func TestUnsignedDivRemHintOutOfBoundsError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			// This is the condition that should make the hint execution error.
			"div":   {NewMaybeRelocatableFelt(FeltFromDecString("10633823966279327296825105735305134081"))},
			"value": {NewMaybeRelocatableFelt(FeltFromDecString("15"))},
			"r":     {nil},
			"q":     {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: UNSIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("UNSIGNED_DIV_REM should have failed")
	}
}

func TestSplitFeltAssertPrimeFailure(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(1))},
			"high":  {nil},
			"low":   {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	constants := SetupConstantsForTest(map[string]Felt{
		"MAX_HIGH": lambdaworks.FeltFromHex("0xffffffffffffffffffffffffffffffff"),
		"MAX_LOW":  lambdaworks.FeltFromHex("0xffffffffffffffffffffffffffffffff"),
	},
		&idsManager,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_FELT,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, nil)
	if err == nil {
		t.Errorf("SPLIT_FELT hint should have failed with assert PRIME - 1 == ids.MAX_HIGH * 2**128 + ids.MAX_LOW error")
	}
}

func TestSignedDivRemHintSuccess(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"div":      {NewMaybeRelocatableFelt(FeltFromDecString("3"))},
			"value":    {NewMaybeRelocatableFelt(FeltFromDecString("10"))},
			"bound":    {NewMaybeRelocatableFelt(FeltFromDecString("29"))},
			"r":        {nil},
			"biased_q": {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("UNSIGNED_DIV_REM hint failed with error: %s", err)
	}

	biasedQ, err := idsManager.GetFelt("biased_q", vm)
	if err != nil {
		t.Errorf("failed to get `biased_q`: %s", err)
	}

	r, err := idsManager.GetFelt("r", vm)
	if err != nil {
		t.Errorf("failed to get `r`: %s", err)
	}

	if biasedQ != FeltFromUint64(32) {
		t.Errorf("Expected biased_q=32, got: %v", biasedQ)
	}

	if r != FeltFromUint64(1) {
		t.Errorf("Expected r=1, got: %v", r)
	}
}

func TestSignedDivRemHintDivZeroError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"div":      {NewMaybeRelocatableFelt(FeltFromDecString("0"))},
			"value":    {NewMaybeRelocatableFelt(FeltFromDecString("10"))},
			"bound":    {NewMaybeRelocatableFelt(FeltFromDecString("29"))},
			"r":        {nil},
			"biased_q": {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("UNSIGNED_DIV_REM hint should have failed")
	}
}

func TestSignedDivRemHintOutOfRcBoundsError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"div":      {NewMaybeRelocatableFelt(FeltFromDecString("10633823966279327296825105735305134081"))},
			"value":    {NewMaybeRelocatableFelt(FeltFromDecString("10"))},
			"bound":    {NewMaybeRelocatableFelt(FeltFromDecString("29"))},
			"r":        {nil},
			"biased_q": {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("UNSIGNED_DIV_REM hint should have failed")
	}
}

func TestSignedDivRemHintOutOfBoundsError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"div":      {NewMaybeRelocatableFelt(FeltFromDecString("4"))},
			"value":    {NewMaybeRelocatableFelt(FeltFromDecString("16"))},
			"bound":    {NewMaybeRelocatableFelt(FeltFromDecString("2"))},
			"r":        {nil},
			"biased_q": {nil},
		},
		vm,
	)
	rcBuiltin := builtins.DefaultRangeCheckBuiltinRunner()
	vm.BuiltinRunners = []builtins.BuiltinRunner{rcBuiltin}

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SIGNED_DIV_REM,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("UNSIGNED_DIV_REM hint should have failed")
	}
}

func TestSplitFeltAssertMaxHighFailedAssertion(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(1))},
			"high":  {nil},
			"low":   {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	constants := SetupConstantsForTest(map[string]Felt{
		"MAX_HIGH": lambdaworks.FeltFromHex("0xffffffffffffffffffffffffffffffffffff"),
		"MAX_LOW":  lambdaworks.FeltFromHex("0xffffffffffffffffffffffffffffffff"),
	},
		&idsManager,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_FELT,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, nil)
	if err == nil {
		t.Errorf("SPLIT_FELT hint should have failed with assert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128")
	}
}

func TestSplitFeltAssertMaxLowFailedAssertion(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromUint64(1))},
			"high":  {nil},
			"low":   {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	constants := SetupConstantsForTest(map[string]Felt{
		"MAX_HIGH": lambdaworks.FeltFromHex("0xffffffffffffffffffffffffffffffff"),
		"MAX_LOW":  lambdaworks.FeltFromHex("0xffffffffffffffffffffffffffffffffffff"),
	},
		&idsManager,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_FELT,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, nil)
	if err == nil {
		t.Errorf("SPLIT_FELT hint should have failed with assert ids.MAX_HIGH < 2**128 and ids.MAX_LOW < 2**128")
	}
}

func TestSplitFeltSuccess(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()

	firstLimb := lambdaworks.FeltFromUint64(1)
	secondLimb := lambdaworks.FeltFromUint64(2)
	thirdLimb := lambdaworks.FeltFromUint64(3)
	fourthLimb := lambdaworks.FeltFromUint64(4)
	value := fourthLimb.Or(thirdLimb.Shl(64).Or(secondLimb.Shl(128).Or(firstLimb.Shl(192))))
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(value)},
			"high":  {nil},
			"low":   {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	constants := SetupConstantsForTest(map[string]Felt{
		"MAX_HIGH": lambdaworks.FeltFromDecString("10633823966279327296825105735305134080"),
		"MAX_LOW":  lambdaworks.FeltFromUint64(0),
	},
		&idsManager,
	)

	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_FELT,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, &constants, nil)
	if err != nil {
		t.Errorf("SPLIT_FELT hint failed with error %s", err)
	}

	high, err := idsManager.GetFelt("high", vm)
	if err != nil {
		t.Errorf("failed to get high: %s", err)
	}

	low, err := idsManager.GetFelt("low", vm)
	if err != nil {
		t.Errorf("failed to get low: %s", err)
	}

	if high != firstLimb.Shl(64).Or(secondLimb) {
		t.Errorf("Expected high == 335438970432432812899076431678123043273. Got: %v", high)
	}

	if low != thirdLimb.Shl(64).Or(fourthLimb) {
		t.Errorf("Expected low == 0. Got: %v", low)
	}
}

func TestSplitIntHintSuccess(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value":  {NewMaybeRelocatableFelt(FeltFromDecString("6"))},
			"base":   {NewMaybeRelocatableFelt(FeltFromDecString("4"))},
			"bound":  {NewMaybeRelocatableFelt(FeltFromDecString("58"))},
			"output": {NewMaybeRelocatableRelocatable(NewRelocatable(0, 4))},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_INT,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("SPLIT_INT hint failed with error %s", err)
	}

	res, err := vm.Segments.Memory.GetFelt(NewRelocatable(0, 4))
	if err != nil {
		t.Errorf("SPLIT_INT hint failed, `res` value not inserted")
	}

	if res.Cmp(lambdaworks.FeltFromUint64(2)) != 0 {
		t.Errorf("SPLIT_INT hint failed. Expected 2, got: %d", res.ToBigInt())
	}
}

func TestSplitIntHintOutOfBoundsError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value":  {NewMaybeRelocatableFelt(FeltFromDecString("17"))},
			"base":   {NewMaybeRelocatableFelt(FeltFromDecString("9"))},
			"bound":  {NewMaybeRelocatableFelt(FeltFromDecString("5"))},
			"output": {NewMaybeRelocatableRelocatable(NewRelocatable(0, 4))},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_INT,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("SPLIT_INT hint should have failed")
	}
}

func TestSplitIntAssertRangeHintSuccess(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromDecString("0"))},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_INT_ASSERT_RANGE,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("SPLIT_INT_ASSERT_RANGE hint failed with error: %s", err)
	}
}

func TestSplitIntAssertRangeHintOutOfRangeError(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"value": {NewMaybeRelocatableFelt(FeltFromDecString("3"))},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: SPLIT_INT_ASSERT_RANGE,
	})

	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err == nil {
		t.Errorf("SPLIT_INT_ASSERT_RANGE hint should have failed")
	}
}

func TestIsQuadResidueOk(t *testing.T) {
	vm := NewVirtualMachine()
	vm.Segments.AddSegment()
	idsManager := SetupIdsForTest(
		map[string][]*MaybeRelocatable{
			"x": {NewMaybeRelocatableFelt(FeltFromDecString("151461"))},
			"y": {nil},
		},
		vm,
	)

	hintProcessor := CairoVmHintProcessor{}
	hintData := any(HintData{
		Ids:  idsManager,
		Code: IS_QUAD_RESIDUE,
	})
	
	err := hintProcessor.ExecuteHint(vm, &hintData, nil, nil)
	if err != nil {
		t.Errorf("IS_QUAD_RESIDUE hint failed with error: %s", err)

	}

	root, err := idsManager.GetFelt("y", vm)
	if err != nil {
		t.Errorf("failed to get Y: %s", err)
	}
	if root != FeltFromDecString("724216096429330872433307564225094804656986062203645968681663112868516882638") {
		t.Errorf("Expected is_quad_residue(151461) == 724216096429330872433307564225094804656986062203645968681663112868516882638 Got: %v", root)
	}
}
