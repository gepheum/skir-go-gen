package goldens_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"e2e-test/skir_client"
	goldens "e2e-test/skirout/external/gepheum/skir_golden_tests/goldens"
)

// typedValueResult holds closures for encoding and re-decoding a typed value.
type typedValueResult struct {
	toDenseJson           func() string
	toReadableJson        func() string
	toBytes               func() []byte
	typeDescriptorJson    func() string
	reDecodeFromDenseJson func(j string) (typedValueResult, error)
	reDecodeFromBytes     func(b []byte) (typedValueResult, error)
}

func newTypedValueResult[T any](v T, s skir_client.Serializer[T]) typedValueResult {
	return typedValueResult{
		toDenseJson:        func() string { return s.ToJson(v) },
		toReadableJson:     func() string { return s.ToJson(v, skir_client.Readable{}) },
		toBytes:            func() []byte { return s.ToBytes(v) },
		typeDescriptorJson: func() string { return s.TypeDescriptor().AsJson() },
		reDecodeFromDenseJson: func(j string) (typedValueResult, error) {
			decoded, err := s.FromJson(j)
			if err != nil {
				return typedValueResult{}, fmt.Errorf("re-decode from dense json: %w", err)
			}
			return newTypedValueResult(decoded, s), nil
		},
		reDecodeFromBytes: func(b []byte) (typedValueResult, error) {
			decoded, err := s.FromBytes(b)
			if err != nil {
				return typedValueResult{}, fmt.Errorf("re-decode from bytes: %w", err)
			}
			return newTypedValueResult(decoded, s), nil
		},
	}
}

func evaluateStringExpression(expr goldens.StringExpression) (string, error) {
	switch expr.Kind() {
	case goldens.StringExpression_kind_LiteralWrapper:
		return expr.UnwrapLiteral(), nil
	case goldens.StringExpression_kind_ToDenseJsonWrapper:
		res, err := evaluateTypedValue(expr.UnwrapToDenseJson())
		if err != nil {
			return "", err
		}
		return res.toDenseJson(), nil
	case goldens.StringExpression_kind_ToReadableJsonWrapper:
		res, err := evaluateTypedValue(expr.UnwrapToReadableJson())
		if err != nil {
			return "", err
		}
		return res.toReadableJson(), nil
	default:
		return "", fmt.Errorf("unsupported StringExpression kind: %v", expr.Kind())
	}
}

func evaluateBytesExpression(expr goldens.BytesExpression) (skir_client.Bytes, error) {
	switch expr.Kind() {
	case goldens.BytesExpression_kind_LiteralWrapper:
		return expr.UnwrapLiteral(), nil
	case goldens.BytesExpression_kind_ToBytesWrapper:
		res, err := evaluateTypedValue(expr.UnwrapToBytes())
		if err != nil {
			return skir_client.Bytes{}, err
		}
		return skir_client.BytesFromSlice(res.toBytes()), nil
	default:
		return skir_client.Bytes{}, fmt.Errorf("unsupported BytesExpression kind: %v", expr.Kind())
	}
}

func parseFromJson[T any](s skir_client.Serializer[T], expr goldens.StringExpression, keepUnrecognized bool) (T, error) {
	j, err := evaluateStringExpression(expr)
	if err != nil {
		var zero T
		return zero, err
	}
	if keepUnrecognized {
		return s.FromJson(j, skir_client.KeepUnrecognizedValues{})
	}
	return s.FromJson(j)
}

func parseFromBytes[T any](s skir_client.Serializer[T], expr goldens.BytesExpression, keepUnrecognized bool) (T, error) {
	b, err := evaluateBytesExpression(expr)
	if err != nil {
		var zero T
		return zero, err
	}
	if keepUnrecognized {
		return s.FromBytes(b.Slice(), skir_client.KeepUnrecognizedValues{})
	}
	return s.FromBytes(b.Slice())
}

func evaluateTypedValue(tv goldens.TypedValue) (typedValueResult, error) {
	switch tv.Kind() {
	case goldens.TypedValue_kind_BoolWrapper:
		return newTypedValueResult(tv.UnwrapBool(), skir_client.BoolSerializer()), nil
	case goldens.TypedValue_kind_Int32Wrapper:
		return newTypedValueResult(tv.UnwrapInt32(), skir_client.Int32Serializer()), nil
	case goldens.TypedValue_kind_Int64Wrapper:
		return newTypedValueResult(tv.UnwrapInt64(), skir_client.Int64Serializer()), nil
	case goldens.TypedValue_kind_Hash64Wrapper:
		return newTypedValueResult(tv.UnwrapHash64(), skir_client.Hash64Serializer()), nil
	case goldens.TypedValue_kind_Float32Wrapper:
		return newTypedValueResult(tv.UnwrapFloat32(), skir_client.Float32Serializer()), nil
	case goldens.TypedValue_kind_Float64Wrapper:
		return newTypedValueResult(tv.UnwrapFloat64(), skir_client.Float64Serializer()), nil
	case goldens.TypedValue_kind_TimestampWrapper:
		return newTypedValueResult(tv.UnwrapTimestamp(), skir_client.TimestampSerializer()), nil
	case goldens.TypedValue_kind_StringWrapper:
		return newTypedValueResult(tv.UnwrapString(), skir_client.StringSerializer()), nil
	case goldens.TypedValue_kind_BytesWrapper:
		return newTypedValueResult(tv.UnwrapBytes(), skir_client.BytesSerializer()), nil
	case goldens.TypedValue_kind_BoolOptionalWrapper:
		return newTypedValueResult(tv.UnwrapBoolOptional(), skir_client.OptionalSerializer(skir_client.BoolSerializer())), nil
	case goldens.TypedValue_kind_IntsWrapper:
		return newTypedValueResult(tv.UnwrapInts(), skir_client.ArraySerializer(skir_client.Int32Serializer())), nil
	case goldens.TypedValue_kind_PointWrapper:
		return newTypedValueResult(*tv.UnwrapPoint(), goldens.Point_serializer()), nil
	case goldens.TypedValue_kind_ColorWrapper:
		return newTypedValueResult(*tv.UnwrapColor(), goldens.Color_serializer()), nil
	case goldens.TypedValue_kind_MyEnumWrapper:
		return newTypedValueResult(tv.UnwrapMyEnum(), goldens.MyEnum_serializer()), nil
	case goldens.TypedValue_kind_KeyedArraysWrapper:
		return newTypedValueResult(*tv.UnwrapKeyedArrays(), goldens.KeyedArrays_serializer()), nil
	case goldens.TypedValue_kind_RecStructWrapper:
		return newTypedValueResult(*tv.UnwrapRecStruct(), goldens.RecStruct_serializer()), nil
	case goldens.TypedValue_kind_RecEnumWrapper:
		return newTypedValueResult(tv.UnwrapRecEnum(), goldens.RecEnum_serializer()), nil

	case goldens.TypedValue_kind_PointFromJsonDropUnrecognizedWrapper:
		v, err := parseFromJson(goldens.Point_serializer(), tv.UnwrapPointFromJsonDropUnrecognized(), false)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Point_serializer()), nil
	case goldens.TypedValue_kind_PointFromJsonKeepUnrecognizedWrapper:
		v, err := parseFromJson(goldens.Point_serializer(), tv.UnwrapPointFromJsonKeepUnrecognized(), true)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Point_serializer()), nil
	case goldens.TypedValue_kind_PointFromBytesDropUnrecognizedWrapper:
		v, err := parseFromBytes(goldens.Point_serializer(), tv.UnwrapPointFromBytesDropUnrecognized(), false)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Point_serializer()), nil
	case goldens.TypedValue_kind_PointFromBytesKeepUnrecognizedWrapper:
		v, err := parseFromBytes(goldens.Point_serializer(), tv.UnwrapPointFromBytesKeepUnrecognized(), true)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Point_serializer()), nil

	case goldens.TypedValue_kind_ColorFromJsonDropUnrecognizedWrapper:
		v, err := parseFromJson(goldens.Color_serializer(), tv.UnwrapColorFromJsonDropUnrecognized(), false)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Color_serializer()), nil
	case goldens.TypedValue_kind_ColorFromJsonKeepUnrecognizedWrapper:
		v, err := parseFromJson(goldens.Color_serializer(), tv.UnwrapColorFromJsonKeepUnrecognized(), true)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Color_serializer()), nil
	case goldens.TypedValue_kind_ColorFromBytesDropUnrecognizedWrapper:
		v, err := parseFromBytes(goldens.Color_serializer(), tv.UnwrapColorFromBytesDropUnrecognized(), false)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Color_serializer()), nil
	case goldens.TypedValue_kind_ColorFromBytesKeepUnrecognizedWrapper:
		v, err := parseFromBytes(goldens.Color_serializer(), tv.UnwrapColorFromBytesKeepUnrecognized(), true)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.Color_serializer()), nil

	case goldens.TypedValue_kind_MyEnumFromJsonDropUnrecognizedWrapper:
		v, err := parseFromJson(goldens.MyEnum_serializer(), tv.UnwrapMyEnumFromJsonDropUnrecognized(), false)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.MyEnum_serializer()), nil
	case goldens.TypedValue_kind_MyEnumFromJsonKeepUnrecognizedWrapper:
		v, err := parseFromJson(goldens.MyEnum_serializer(), tv.UnwrapMyEnumFromJsonKeepUnrecognized(), true)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.MyEnum_serializer()), nil
	case goldens.TypedValue_kind_MyEnumFromBytesDropUnrecognizedWrapper:
		v, err := parseFromBytes(goldens.MyEnum_serializer(), tv.UnwrapMyEnumFromBytesDropUnrecognized(), false)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.MyEnum_serializer()), nil
	case goldens.TypedValue_kind_MyEnumFromBytesKeepUnrecognizedWrapper:
		v, err := parseFromBytes(goldens.MyEnum_serializer(), tv.UnwrapMyEnumFromBytesKeepUnrecognized(), true)
		if err != nil {
			return typedValueResult{}, err
		}
		return newTypedValueResult(v, goldens.MyEnum_serializer()), nil

	case goldens.TypedValue_kind_RoundTripDenseJsonWrapper:
		inner, err := evaluateTypedValue(tv.UnwrapRoundTripDenseJson())
		if err != nil {
			return typedValueResult{}, err
		}
		return inner.reDecodeFromDenseJson(inner.toDenseJson())
	case goldens.TypedValue_kind_RoundTripReadableJsonWrapper:
		inner, err := evaluateTypedValue(tv.UnwrapRoundTripReadableJson())
		if err != nil {
			return typedValueResult{}, err
		}
		return inner.reDecodeFromDenseJson(inner.toReadableJson())
	case goldens.TypedValue_kind_RoundTripBytesWrapper:
		inner, err := evaluateTypedValue(tv.UnwrapRoundTripBytes())
		if err != nil {
			return typedValueResult{}, err
		}
		return inner.reDecodeFromBytes(inner.toBytes())

	default:
		return typedValueResult{}, fmt.Errorf("unsupported TypedValue kind: %v", tv.Kind())
	}
}

func reserializeValueAndVerify(t *testing.T, input *goldens.Assertion_ReserializeValue) {
	t.Helper()
	res, err := evaluateTypedValue(input.Value())
	if err != nil {
		t.Fatalf("evaluateTypedValue failed: %v", err)
		return
	}

	// Check expected type descriptor.
	if td := input.ExpectedTypeDescriptor(); td != nil {
		if got := res.typeDescriptorJson(); got != *td {
			t.Errorf("type descriptor:\n  got:      %s\n  expected: %s", got, *td)
		}
	}

	// Check dense JSON.
	actualDenseJson := res.toDenseJson()
	if exp := input.ExpectedDenseJson(); !exp.IsEmpty() {
		found := false
		for _, e := range exp.ToSlice() {
			if actualDenseJson == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("dense json:\n  got:      %s\n  expected one of: %v", actualDenseJson, exp.ToSlice())
		}
	}

	// Verify alternative JSON representations decode to the canonical dense JSON.
	for i, altExpr := range input.AlternativeJsons().ToSlice() {
		alt, err := evaluateStringExpression(altExpr)
		if err != nil {
			t.Errorf("alternative_json[%d]: %v", i, err)
			continue
		}
		reDecoded, err := res.reDecodeFromDenseJson(alt)
		if err != nil {
			t.Errorf("alternative_json[%d]: re-decode from %q failed: %v", i, alt, err)
			continue
		}
		if got := reDecoded.toDenseJson(); got != actualDenseJson {
			t.Errorf("alternative_json[%d]: round-trip dense json:\n  got:      %s\n  expected: %s", i, got, actualDenseJson)
		}
	}

	// Check readable JSON.
	actualReadableJson := res.toReadableJson()
	if exp := input.ExpectedReadableJson(); !exp.IsEmpty() {
		found := false
		for _, e := range exp.ToSlice() {
			if actualReadableJson == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("readable json:\n  got:      %s\n  expected one of: %v", actualReadableJson, exp.ToSlice())
		}
	}

	// Check bytes.
	actualBytes := skir_client.BytesFromSlice(res.toBytes())
	if exp := input.ExpectedBytes(); !exp.IsEmpty() {
		found := false
		for _, e := range exp.ToSlice() {
			if actualBytes.Equal(e) {
				found = true
				break
			}
		}
		if !found {
			var expHexes []string
			for _, e := range exp.ToSlice() {
				expHexes = append(expHexes, "hex:"+e.Hex())
			}
			t.Errorf("bytes:\n  got:      hex:%s\n  expected one of: %v", actualBytes.Hex(), expHexes)
		}
	}

	// Verify alternative byte representations decode to the canonical dense JSON.
	for i, altExpr := range input.AlternativeBytes().ToSlice() {
		alt, err := evaluateBytesExpression(altExpr)
		if err != nil {
			t.Errorf("alternative_bytes[%d]: %v", i, err)
			continue
		}
		reDecoded, err := res.reDecodeFromBytes(alt.Slice())
		if err != nil {
			t.Errorf("alternative_bytes[%d]: re-decode from hex:%s failed: %v", i, alt.Hex(), err)
			continue
		}
		if got := reDecoded.toDenseJson(); got != actualDenseJson {
			t.Errorf("alternative_bytes[%d]: round-trip dense json:\n  got:      %s\n  expected: %s", i, got, actualDenseJson)
		}
	}
}

func reserializeLargeStringAndVerify(t *testing.T, input *goldens.Assertion_ReserializeLargeString) {
	t.Helper()
	str := strings.Repeat("a", int(input.NumChars()))
	s := skir_client.StringSerializer()

	// Dense JSON round-trip.
	j := s.ToJson(str)
	if rt, err := s.FromJson(j); err != nil || rt != str {
		t.Errorf("large string dense-json round-trip failed (len=%d): %v", len(str), err)
	}

	// Readable JSON round-trip.
	j = s.ToJson(str, skir_client.Readable{})
	if rt, err := s.FromJson(j); err != nil || rt != str {
		t.Errorf("large string readable-json round-trip failed (len=%d): %v", len(str), err)
	}

	// Bytes: check prefix and round-trip.
	b := s.ToBytes(str)
	prefix := input.ExpectedBytePrefix()
	if !bytes.HasPrefix(b, prefix.Slice()) {
		t.Errorf("large string bytes prefix:\n  got:      hex:%s\n  expected prefix: hex:%s",
			skir_client.BytesFromSlice(b).Hex(), prefix.Hex())
	}
	if rt, err := s.FromBytes(b); err != nil || rt != str {
		t.Errorf("large string bytes round-trip failed (len=%d): %v", len(str), err)
	}
}

func reserializeLargeArrayAndVerify(t *testing.T, input *goldens.Assertion_ReserializeLargeArray) {
	t.Helper()
	n := int(input.NumItems())
	elems := make([]int32, n)
	for i := range elems {
		elems[i] = 1
	}
	arr := skir_client.ArrayFromSlice(elems)
	s := skir_client.ArraySerializer(skir_client.Int32Serializer())

	// Dense JSON round-trip.
	j := s.ToJson(arr)
	if rt, err := s.FromJson(j); err != nil || rt.Len() != n {
		t.Errorf("large array dense-json round-trip failed (len=%d): %v", n, err)
	}

	// Readable JSON round-trip.
	j = s.ToJson(arr, skir_client.Readable{})
	if rt, err := s.FromJson(j); err != nil || rt.Len() != n {
		t.Errorf("large array readable-json round-trip failed (len=%d): %v", n, err)
	}

	// Bytes: check prefix and round-trip.
	b := s.ToBytes(arr)
	prefix := input.ExpectedBytePrefix()
	if !bytes.HasPrefix(b, prefix.Slice()) {
		t.Errorf("large array bytes prefix:\n  got:      hex:%s\n  expected prefix: hex:%s",
			skir_client.BytesFromSlice(b).Hex(), prefix.Hex())
	}
	if rt, err := s.FromBytes(b); err != nil || rt.Len() != n {
		t.Errorf("large array bytes round-trip failed (len=%d): %v", n, err)
	}
}

func verifyAssertion(t *testing.T, assertion goldens.Assertion) {
	t.Helper()
	switch assertion.Kind() {
	case goldens.Assertion_kind_StringEqualWrapper:
		input := assertion.UnwrapStringEqual()
		actual, err := evaluateStringExpression(input.Actual())
		if err != nil {
			t.Fatalf("StringEqual.actual: %v", err)
		}
		expected, err := evaluateStringExpression(input.Expected())
		if err != nil {
			t.Fatalf("StringEqual.expected: %v", err)
		}
		if actual != expected {
			t.Errorf("string not equal:\n  got:      %q\n  expected: %q", actual, expected)
		}

	case goldens.Assertion_kind_StringInWrapper:
		input := assertion.UnwrapStringIn()
		actual, err := evaluateStringExpression(input.Actual())
		if err != nil {
			t.Fatalf("StringIn.actual: %v", err)
		}
		found := false
		for _, exp := range input.Expected().ToSlice() {
			if actual == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("string not in expected set:\n  got:      %q\n  expected: %v", actual, input.Expected().ToSlice())
		}

	case goldens.Assertion_kind_BytesEqualWrapper:
		input := assertion.UnwrapBytesEqual()
		actual, err := evaluateBytesExpression(input.Actual())
		if err != nil {
			t.Fatalf("BytesEqual.actual: %v", err)
		}
		expected, err := evaluateBytesExpression(input.Expected())
		if err != nil {
			t.Fatalf("BytesEqual.expected: %v", err)
		}
		if !actual.Equal(expected) {
			t.Errorf("bytes not equal:\n  got:      hex:%s\n  expected: hex:%s", actual.Hex(), expected.Hex())
		}

	case goldens.Assertion_kind_BytesInWrapper:
		input := assertion.UnwrapBytesIn()
		actual, err := evaluateBytesExpression(input.Actual())
		if err != nil {
			t.Fatalf("BytesIn.actual: %v", err)
		}
		found := false
		for _, exp := range input.Expected().ToSlice() {
			if actual.Equal(exp) {
				found = true
				break
			}
		}
		if !found {
			var expHexes []string
			for _, exp := range input.Expected().ToSlice() {
				expHexes = append(expHexes, "hex:"+exp.Hex())
			}
			t.Errorf("bytes not in expected set:\n  got:      hex:%s\n  expected: %v", actual.Hex(), expHexes)
		}

	case goldens.Assertion_kind_ReserializeValueWrapper:
		reserializeValueAndVerify(t, assertion.UnwrapReserializeValue())

	case goldens.Assertion_kind_ReserializeLargeStringWrapper:
		reserializeLargeStringAndVerify(t, assertion.UnwrapReserializeLargeString())

	case goldens.Assertion_kind_ReserializeLargeArrayWrapper:
		reserializeLargeArrayAndVerify(t, assertion.UnwrapReserializeLargeArray())

	default:
		t.Errorf("unsupported assertion kind: %v", assertion.Kind())
	}
}

// TestGoldens runs all unit tests from the goldens schema.
func TestGoldens(t *testing.T) {
	unitTests := goldens.UnitTests_const()
	all := unitTests.ToSlice()

	// Verify sequential test numbers.
	if len(all) > 0 {
		first := all[0].TestNumber()
		for i, ut := range all {
			expected := first + int32(i)
			if ut.TestNumber() != expected {
				t.Fatalf("test numbers are not sequential at index %d: found %d, expected %d",
					i, ut.TestNumber(), expected)
			}
		}
	}

	for _, ut := range all {
		ut := ut // capture for closure
		t.Run(fmt.Sprintf("test #%d", ut.TestNumber()), func(t *testing.T) {
			verifyAssertion(t, ut.Assertion())
		})
	}
}
