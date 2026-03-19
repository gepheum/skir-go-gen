package skir_client

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

// mustParseTypeDescriptor fails the test if parsing returns an error.
func mustParseTypeDescriptor(t *testing.T, jsonCode string) TypeDescriptor {
	t.Helper()
	td, err := ParseTypeDescriptorFromJson(jsonCode)
	if err != nil {
		t.Fatalf("ParseTypeDescriptorFromJson(%q): %v", jsonCode, err)
	}
	return td
}

// normalizeJSON round-trips through encoding/json so that key order and
// whitespace don't affect equality assertions.
func normalizeJSON(t *testing.T, s string) string {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("normalizeJSON: invalid JSON %q: %v", s, err)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("normalizeJSON: re-marshal failed: %v", err)
	}
	return string(out)
}

// assertJsonEqual fails unless got and want are equivalent JSON documents.
func assertJsonEqual(t *testing.T, got, want string) {
	t.Helper()
	if normalizeJSON(t, got) != normalizeJSON(t, want) {
		t.Errorf("JSON mismatch:\n got  %s\n want %s", got, want)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Primitive descriptors
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_Primitive_AsJson(t *testing.T) {
	cases := []struct {
		desc     *PrimitiveDescriptor
		wantType string
	}{
		{BoolDescriptor, "bool"},
		{Int32Descriptor, "int32"},
		{Int64Descriptor, "int64"},
		{Hash64Descriptor, "hash64"},
		{Float32Descriptor, "float32"},
		{Float64Descriptor, "float64"},
		{TimestampDescriptor, "timestamp"},
		{StringDescriptor, "string"},
		{BytesDescriptor, "bytes"},
	}
	for _, tc := range cases {
		t.Run(tc.wantType, func(t *testing.T) {
			got := tc.desc.AsJson()
			want := `{"type":{"kind":"primitive","value":"` + tc.wantType + `"},"records":[]}`
			assertJsonEqual(t, got, want)
		})
	}
}

func TestTypeDescriptor_Primitive_RoundTrip(t *testing.T) {
	for _, desc := range []TypeDescriptor{
		BoolDescriptor, Int32Descriptor, Int64Descriptor, Hash64Descriptor,
		Float32Descriptor, Float64Descriptor, TimestampDescriptor,
		StringDescriptor, BytesDescriptor,
	} {
		j := desc.AsJson()
		got := mustParseTypeDescriptor(t, j)
		if got != desc {
			t.Errorf("%T round-trip: got %v, want %v", desc, got, desc)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Optional descriptor
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_Optional_AsJson(t *testing.T) {
	desc := &OptionalDescriptor{otherType: Int32Descriptor}
	want := `{"type":{"kind":"optional","value":{"kind":"primitive","value":"int32"}},"records":[]}`
	assertJsonEqual(t, desc.AsJson(), want)
}

func TestTypeDescriptor_Optional_RoundTrip(t *testing.T) {
	desc := &OptionalDescriptor{otherType: StringDescriptor}
	got := mustParseTypeDescriptor(t, desc.AsJson())
	opt, ok := got.(*OptionalDescriptor)
	if !ok {
		t.Fatalf("got %T, want *OptionalDescriptor", got)
	}
	if opt.GetOtherType() != StringDescriptor {
		t.Errorf("OtherType = %v, want StringDescriptor", opt.GetOtherType())
	}
}

func TestTypeDescriptor_NestedOptional_RoundTrip(t *testing.T) {
	// optional(optional(bool))
	desc := &OptionalDescriptor{otherType: &OptionalDescriptor{otherType: BoolDescriptor}}
	got := mustParseTypeDescriptor(t, desc.AsJson())
	outer, ok := got.(*OptionalDescriptor)
	if !ok {
		t.Fatalf("outer: got %T, want *OptionalDescriptor", got)
	}
	inner, ok := outer.GetOtherType().(*OptionalDescriptor)
	if !ok {
		t.Fatalf("inner: got %T, want *OptionalDescriptor", outer.GetOtherType())
	}
	if inner.GetOtherType() != BoolDescriptor {
		t.Errorf("innermost type = %v, want BoolDescriptor", inner.GetOtherType())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Array descriptor
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_Array_AsJson(t *testing.T) {
	desc := &ArrayDescriptor{itemType: StringDescriptor}
	want := `{"type":{"kind":"array","value":{"item":{"kind":"primitive","value":"string"}}},"records":[]}`
	assertJsonEqual(t, desc.AsJson(), want)
}

func TestTypeDescriptor_Array_WithKeyExtractor_AsJson(t *testing.T) {
	desc := &ArrayDescriptor{itemType: StringDescriptor, keyExtractor: "id"}
	got := desc.AsJson()
	// key_extractor must appear in the output
	var v map[string]any
	if err := json.Unmarshal([]byte(got), &v); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	typeObj := v["type"].(map[string]any)
	value := typeObj["value"].(map[string]any)
	if value["key_extractor"] != "id" {
		t.Errorf("key_extractor = %v, want %q", value["key_extractor"], "id")
	}
}

func TestTypeDescriptor_Array_RoundTrip(t *testing.T) {
	desc := &ArrayDescriptor{itemType: Int64Descriptor}
	got := mustParseTypeDescriptor(t, desc.AsJson())
	arr, ok := got.(*ArrayDescriptor)
	if !ok {
		t.Fatalf("got %T, want *ArrayDescriptor", got)
	}
	if arr.GetItemType() != Int64Descriptor {
		t.Errorf("ItemType = %v, want Int64Descriptor", arr.GetItemType())
	}
	if arr.GetKeyExtractor() != "" {
		t.Errorf("KeyExtractor = %q, want empty", arr.GetKeyExtractor())
	}
}

func TestTypeDescriptor_Array_KeyExtractor_RoundTrip(t *testing.T) {
	desc := &ArrayDescriptor{itemType: StringDescriptor, keyExtractor: "some.key"}
	got := mustParseTypeDescriptor(t, desc.AsJson())
	arr, ok := got.(*ArrayDescriptor)
	if !ok {
		t.Fatalf("got %T, want *ArrayDescriptor", got)
	}
	if arr.GetKeyExtractor() != "some.key" {
		t.Errorf("KeyExtractor = %q, want %q", arr.GetKeyExtractor(), "some.key")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Struct descriptor
// ─────────────────────────────────────────────────────────────────────────────

func makeSimpleStruct() *StructDescriptor {
	return newStructDescriptor(
		"my/module",
		"MyStruct",
		"doc for MyStruct",
		map[int]struct{}{5: {}},
		[]*StructField{
			{name: "name", number: 1, fieldType: StringDescriptor, doc: "the name"},
			{name: "age", number: 2, fieldType: Int32Descriptor},
		},
	)
}

func TestTypeDescriptor_Struct_AsJson_ContainsFields(t *testing.T) {
	desc := makeSimpleStruct()
	got := desc.AsJson()

	var root map[string]any
	if err := json.Unmarshal([]byte(got), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	records := root["records"].([]any)
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	rec := records[0].(map[string]any)
	if rec["kind"] != "struct" {
		t.Errorf("kind = %v, want struct", rec["kind"])
	}
	if rec["id"] != "my/module:MyStruct" {
		t.Errorf("id = %v, want my/module:MyStruct", rec["id"])
	}
	if rec["doc"] != "doc for MyStruct" {
		t.Errorf("doc = %v, want 'doc for MyStruct'", rec["doc"])
	}
	fields := rec["fields"].([]any)
	if len(fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(fields))
	}
	f0 := fields[0].(map[string]any)
	if f0["name"] != "name" || f0["number"] != float64(1) || f0["doc"] != "the name" {
		t.Errorf("field 0 = %v", f0)
	}
	f1 := fields[1].(map[string]any)
	if f1["name"] != "age" || f1["number"] != float64(2) {
		t.Errorf("field 1 = %v", f1)
	}
	// doc omitted when empty
	if _, hasDoc := f1["doc"]; hasDoc {
		t.Error("field 1 should not have 'doc' key")
	}

	// removed_numbers
	rn := rec["removed_numbers"].([]any)
	if len(rn) != 1 || rn[0] != float64(5) {
		t.Errorf("removed_numbers = %v, want [5]", rn)
	}
}

func TestTypeDescriptor_Struct_RoundTrip(t *testing.T) {
	orig := makeSimpleStruct()
	got := mustParseTypeDescriptor(t, orig.AsJson())
	sd, ok := got.(*StructDescriptor)
	if !ok {
		t.Fatalf("got %T, want *StructDescriptor", got)
	}
	if sd.GetQualifiedName() != "MyStruct" {
		t.Errorf("QualifiedName = %q, want MyStruct", sd.GetQualifiedName())
	}
	if sd.GetModulePath() != "my/module" {
		t.Errorf("ModulePath = %q, want my/module", sd.GetModulePath())
	}
	if sd.GetDoc() != "doc for MyStruct" {
		t.Errorf("Doc = %q, want 'doc for MyStruct'", sd.GetDoc())
	}
	if len(sd.GetFields()) != 2 {
		t.Fatalf("len(Fields) = %d, want 2", len(sd.GetFields()))
	}
	if sd.GetFields()[0].GetName() != "name" || sd.GetFields()[0].GetNumber() != 1 || sd.GetFields()[0].GetType() != StringDescriptor {
		t.Errorf("Field 0 = %+v", sd.GetFields()[0])
	}
	if sd.GetFields()[0].GetDoc() != "the name" {
		t.Errorf("Field 0 Doc = %q, want 'the name'", sd.GetFields()[0].GetDoc())
	}
	if sd.GetFields()[1].GetName() != "age" || sd.GetFields()[1].GetType() != Int32Descriptor {
		t.Errorf("Field 1 = %+v", sd.GetFields()[1])
	}
	if _, ok := sd.GetRemovedNumbers()[5]; !ok {
		t.Error("RemovedNumbers should contain 5")
	}
}

func TestTypeDescriptor_Struct_RoundTrip_AsJsonIsStable(t *testing.T) {
	orig := makeSimpleStruct()
	j1 := orig.AsJson()
	reparsed := mustParseTypeDescriptor(t, j1)
	j2 := reparsed.AsJson()
	assertJsonEqual(t, j1, j2)
}

// ─────────────────────────────────────────────────────────────────────────────
// Enum descriptor
// ─────────────────────────────────────────────────────────────────────────────

func makeSimpleEnum() *EnumDescriptor {
	return newEnumDescriptor(
		"my/module",
		"Color",
		"a color enum",
		map[int]struct{}{4: {}},
		[]EnumVariant{
			&EnumConstantVariant{name: "RED", number: 1, doc: "red color"},
			&EnumConstantVariant{name: "GREEN", number: 2},
			&EnumConstantVariant{name: "BLUE", number: 3},
		},
	)
}

func TestTypeDescriptor_Enum_AsJson_ContainsVariants(t *testing.T) {
	desc := makeSimpleEnum()
	got := desc.AsJson()

	var root map[string]any
	if err := json.Unmarshal([]byte(got), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	records := root["records"].([]any)
	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	rec := records[0].(map[string]any)
	if rec["kind"] != "enum" {
		t.Errorf("kind = %v, want enum", rec["kind"])
	}
	variants := rec["variants"].([]any)
	if len(variants) != 3 {
		t.Fatalf("variants len = %d, want 3", len(variants))
	}
	v0 := variants[0].(map[string]any)
	if v0["name"] != "RED" || v0["number"] != float64(1) || v0["doc"] != "red color" {
		t.Errorf("variant 0 = %v", v0)
	}
	// doc omitted when empty
	v1 := variants[1].(map[string]any)
	if _, hasDoc := v1["doc"]; hasDoc {
		t.Error("variant 1 should not have 'doc' key")
	}
	rn := rec["removed_numbers"].([]any)
	if len(rn) != 1 || rn[0] != float64(4) {
		t.Errorf("removed_numbers = %v, want [4]", rn)
	}
}

func TestTypeDescriptor_Enum_RoundTrip(t *testing.T) {
	orig := makeSimpleEnum()
	got := mustParseTypeDescriptor(t, orig.AsJson())
	ed, ok := got.(*EnumDescriptor)
	if !ok {
		t.Fatalf("got %T, want *EnumDescriptor", got)
	}
	if ed.GetQualifiedName() != "Color" {
		t.Errorf("QualifiedName = %q, want Color", ed.GetQualifiedName())
	}
	if ed.GetDoc() != "a color enum" {
		t.Errorf("Doc = %q, want 'a color enum'", ed.GetDoc())
	}
	if len(ed.GetVariants()) != 3 {
		t.Fatalf("len(Variants) = %d, want 3", len(ed.GetVariants()))
	}
	if ed.GetVariants()[0].GetName() != "RED" || ed.GetVariants()[0].GetNumber() != 1 {
		t.Errorf("Variant 0 = %+v", ed.GetVariants()[0])
	}
	if ed.GetVariants()[0].GetDoc() != "red color" {
		t.Errorf("Variant 0 doc = %q, want 'red color'", ed.GetVariants()[0].GetDoc())
	}
	if _, ok := ed.GetRemovedNumbers()[4]; !ok {
		t.Error("RemovedNumbers should contain 4")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Enum with wrapper variant
// ─────────────────────────────────────────────────────────────────────────────

func makeWrapperEnum() *EnumDescriptor {
	return newEnumDescriptor(
		"pkg",
		"Result",
		"",
		map[int]struct{}{},
		[]EnumVariant{
			&EnumConstantVariant{name: "EMPTY", number: 1},
			&EnumWrapperVariant{name: "VALUE", number: 2, variantType: StringDescriptor},
			&EnumWrapperVariant{name: "ERROR", number: 3, variantType: Int32Descriptor, doc: "error code"},
		},
	)
}

func TestTypeDescriptor_Enum_WrapperVariant_AsJson(t *testing.T) {
	desc := makeWrapperEnum()
	got := desc.AsJson()

	var root map[string]any
	if err := json.Unmarshal([]byte(got), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	records := root["records"].([]any)
	rec := records[0].(map[string]any)
	variants := rec["variants"].([]any)

	// constant variant has no "type" key
	v0 := variants[0].(map[string]any)
	if _, hasType := v0["type"]; hasType {
		t.Error("constant variant should not have 'type' key")
	}
	// wrapper variants have a "type" key
	v1 := variants[1].(map[string]any)
	if _, hasType := v1["type"]; !hasType {
		t.Error("wrapper variant should have 'type' key")
	}
}

func TestTypeDescriptor_Enum_WrapperVariant_RoundTrip(t *testing.T) {
	orig := makeWrapperEnum()
	got := mustParseTypeDescriptor(t, orig.AsJson())
	ed, ok := got.(*EnumDescriptor)
	if !ok {
		t.Fatalf("got %T, want *EnumDescriptor", got)
	}
	if len(ed.GetVariants()) != 3 {
		t.Fatalf("len(Variants) = %d, want 3", len(ed.GetVariants()))
	}
	if _, ok := ed.GetVariants()[0].(*EnumConstantVariant); !ok {
		t.Errorf("Variant 0: got %T, want *EnumConstantVariant", ed.GetVariants()[0])
	}
	wv, ok := ed.GetVariants()[1].(*EnumWrapperVariant)
	if !ok {
		t.Fatalf("Variant 1: got %T, want *EnumWrapperVariant", ed.GetVariants()[1])
	}
	if wv.GetType() != StringDescriptor {
		t.Errorf("Variant 1 type = %v, want StringDescriptor", wv.GetType())
	}
	wv2, ok := ed.GetVariants()[2].(*EnumWrapperVariant)
	if !ok {
		t.Fatalf("Variant 2: got %T, want *EnumWrapperVariant", ed.GetVariants()[2])
	}
	if wv2.GetDoc() != "error code" {
		t.Errorf("Variant 2 doc = %q, want 'error code'", wv2.GetDoc())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Struct with record-typed field (struct references another struct)
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_Struct_RecordField_RoundTrip(t *testing.T) {
	inner := newStructDescriptor("pkg", "Inner", "", map[int]struct{}{}, []*StructField{
		{name: "x", number: 1, fieldType: Int32Descriptor},
	})
	outer := newStructDescriptor("pkg", "Outer", "", map[int]struct{}{}, []*StructField{
		{name: "inner", number: 1, fieldType: inner},
	})

	got := mustParseTypeDescriptor(t, outer.AsJson())
	sd, ok := got.(*StructDescriptor)
	if !ok {
		t.Fatalf("got %T, want *StructDescriptor", got)
	}
	if sd.GetQualifiedName() != "Outer" {
		t.Errorf("QualifiedName = %q, want Outer", sd.GetQualifiedName())
	}
	innerField := sd.GetFields()[0]
	innerSD, ok := innerField.GetType().(*StructDescriptor)
	if !ok {
		t.Fatalf("inner field type: got %T, want *StructDescriptor", innerField.GetType())
	}
	if innerSD.GetQualifiedName() != "Inner" {
		t.Errorf("inner QualifiedName = %q, want Inner", innerSD.GetQualifiedName())
	}
	if len(innerSD.GetFields()) != 1 || innerSD.GetFields()[0].GetName() != "x" {
		t.Errorf("inner fields = %+v", innerSD.GetFields())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Array of structs
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_ArrayOfStruct_RoundTrip(t *testing.T) {
	item := newStructDescriptor("mod", "Item", "", map[int]struct{}{}, []*StructField{
		{name: "id", number: 1, fieldType: Hash64Descriptor},
	})
	desc := &ArrayDescriptor{itemType: item, keyExtractor: "id"}

	got := mustParseTypeDescriptor(t, desc.AsJson())
	arr, ok := got.(*ArrayDescriptor)
	if !ok {
		t.Fatalf("got %T, want *ArrayDescriptor", got)
	}
	if arr.GetKeyExtractor() != "id" {
		t.Errorf("KeyExtractor = %q, want id", arr.GetKeyExtractor())
	}
	itemSD, ok := arr.GetItemType().(*StructDescriptor)
	if !ok {
		t.Fatalf("ItemType: got %T, want *StructDescriptor", arr.GetItemType())
	}
	if itemSD.GetQualifiedName() != "Item" {
		t.Errorf("item QualifiedName = %q, want Item", itemSD.GetQualifiedName())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Optional of struct
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_OptionalOfStruct_RoundTrip(t *testing.T) {
	sd := newStructDescriptor("mod", "Foo", "", map[int]struct{}{}, nil)
	desc := &OptionalDescriptor{otherType: sd}

	got := mustParseTypeDescriptor(t, desc.AsJson())
	opt, ok := got.(*OptionalDescriptor)
	if !ok {
		t.Fatalf("got %T, want *OptionalDescriptor", got)
	}
	inner, ok := opt.GetOtherType().(*StructDescriptor)
	if !ok {
		t.Fatalf("OtherType: got %T, want *StructDescriptor", opt.GetOtherType())
	}
	if inner.GetQualifiedName() != "Foo" {
		t.Errorf("QualifiedName = %q, want Foo", inner.GetQualifiedName())
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Records appear only once even when referenced multiple times
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_Struct_SharedRecord_AppearsOnce(t *testing.T) {
	shared := newStructDescriptor("mod", "Shared", "", map[int]struct{}{}, []*StructField{
		{name: "v", number: 1, fieldType: Int32Descriptor},
	})
	outer := newStructDescriptor("mod", "Outer", "", map[int]struct{}{}, []*StructField{
		{name: "a", number: 1, fieldType: shared},
		{name: "b", number: 2, fieldType: shared},
	})

	got := outer.AsJson()
	var root map[string]any
	if err := json.Unmarshal([]byte(got), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	records := root["records"].([]any)
	// Outer + Shared = 2 records; Shared must not appear twice.
	if len(records) != 2 {
		t.Errorf("records len = %d, want 2 (Outer + Shared)", len(records))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Error cases
// ─────────────────────────────────────────────────────────────────────────────

func TestParseTypeDescriptorFromJson_InvalidJson(t *testing.T) {
	_, err := ParseTypeDescriptorFromJson("not json")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseTypeDescriptorFromJson_MissingTypeKey(t *testing.T) {
	_, err := ParseTypeDescriptorFromJson(`{"records":[]}`)
	if err == nil {
		t.Error("expected error when 'type' key is missing, got nil")
	}
}

func TestParseTypeDescriptorFromJson_UnknownPrimitive(t *testing.T) {
	_, err := ParseTypeDescriptorFromJson(`{"type":{"kind":"primitive","value":"complex128"},"records":[]}`)
	if err == nil {
		t.Error("expected error for unknown primitive type, got nil")
	}
}

func TestParseTypeDescriptorFromJson_UnknownRecordId(t *testing.T) {
	_, err := ParseTypeDescriptorFromJson(`{"type":{"kind":"record","value":"mod:Missing"},"records":[]}`)
	if err == nil {
		t.Error("expected error for unknown record id, got nil")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// RemovedNumbers ordering
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_RemovedNumbers_SortedInOutput(t *testing.T) {
	desc := newStructDescriptor("m", "S", "", map[int]struct{}{10: {}, 2: {}, 7: {}}, nil)
	got := desc.AsJson()
	var root map[string]any
	if err := json.Unmarshal([]byte(got), &root); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	rn := root["records"].([]any)[0].(map[string]any)["removed_numbers"].([]any)
	want := []any{float64(2), float64(7), float64(10)}
	if !reflect.DeepEqual(rn, want) {
		t.Errorf("removed_numbers = %v, want %v", rn, want)
	}
}

func TestTypeDescriptor_RemovedNumbers_RoundTrip(t *testing.T) {
	orig := newStructDescriptor("m", "S", "", map[int]struct{}{3: {}, 7: {}}, nil)
	got := mustParseTypeDescriptor(t, orig.AsJson())
	sd := got.(*StructDescriptor)
	if _, ok := sd.GetRemovedNumbers()[3]; !ok {
		t.Error("RemovedNumbers should contain 3")
	}
	if _, ok := sd.GetRemovedNumbers()[7]; !ok {
		t.Error("RemovedNumbers should contain 7")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// JSON string escaping in names/docs
// ─────────────────────────────────────────────────────────────────────────────

func TestTypeDescriptor_SpecialCharsInDoc_RoundTrip(t *testing.T) {
	doc := "line1\nline2\t\"quoted\"\\"
	desc := newStructDescriptor("m", "S", doc, map[int]struct{}{}, nil)
	got := mustParseTypeDescriptor(t, desc.AsJson())
	sd := got.(*StructDescriptor)
	if sd.GetDoc() != doc {
		t.Errorf("Doc = %q, want %q", sd.GetDoc(), doc)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Recursive type descriptors
// ─────────────────────────────────────────────────────────────────────────────

// TestTypeDescriptor_RecursiveStruct_AsJson reproduces the nil-pointer panic
// that occurs when a struct contains an optional field of its own type and the
// caller invokes AsJson() on the resulting serializer's type descriptor.
//
// The bug: OptionalSerializer eagerly calls inner.TypeDescriptor() at
// construction time, before the struct adapter's Finalize() has run and set
// its own descriptor.  This stores a typed-nil *StructDescriptor inside the
// OptionalDescriptor, which later panics in typeSignatureToJson.
func TestTypeDescriptor_RecursiveStruct_AsJson(t *testing.T) {
	type recStruct struct {
		self *recStruct
	}
	type recBuilder struct {
		s recStruct
	}

	adapter := Internal__NewStructAdapter(
		&recStruct{},
		func() *recBuilder { return &recBuilder{} },
		func(b *recBuilder) recStruct { return b.s },
		"test", "RecStruct", "",
		func(s *recStruct) *Internal__UnrecognizedFields { return nil },
		func(b *recBuilder, u *Internal__UnrecognizedFields) {},
	)
	// The self-referential field: Optional<RecStruct>.
	// Before the fix, OptionalSerializer calls adapter.Serializer().TypeDescriptor()
	// right here, which returns a nil *StructDescriptor (Finalize not called yet).
	Internal__AddField(
		adapter, "self", 0,
		OptionalSerializer(adapter.Serializer()),
		"",
		func(s *recStruct) **recStruct { return &s.self },
		func(b *recBuilder, v *recStruct) { b.s.self = v },
	)
	adapter.Finalize()

	// Must not panic.
	got := adapter.Serializer().TypeDescriptor().AsJson()
	// Sanity-check: the JSON must reference the record id "test:RecStruct".
	if !strings.Contains(got, "test:RecStruct") {
		t.Errorf("AsJson() = %q, expected it to contain %q", got, "test:RecStruct")
	}
}

// TestTypeDescriptor_RecursiveEnum_AsJson reproduces the same nil-pointer panic
// for a self-referential enum  (e.g.  enum RecEnum { e: RecEnum; }).
//
// The bug: Internal__AddWrapperVariant eagerly calls ser.adapter.typeDescriptor()
// before Finalize, storing a typed-nil *EnumDescriptor in the variant, which
// later panics in typeSignatureToJson.
func TestTypeDescriptor_RecursiveEnum_AsJson(t *testing.T) {
	type recEnum struct {
		kind  int
		value any
	}

	const (
		kindUnknown = 0
		kindE       = 1
	)

	unknown := recEnum{}

	adapter := NewEnumAdapter[recEnum](
		"test", "RecEnum", "",
		func(e recEnum) int { return e.kind },
		2, // kindCount: Unknown + E
		unknown,
		func(u *Internal__UnrecognizedVariant) recEnum { return recEnum{kind: kindUnknown, value: u} },
		func(e recEnum) *Internal__UnrecognizedVariant {
			if v, ok := e.value.(*Internal__UnrecognizedVariant); ok {
				return v
			}
			return nil
		},
	)
	// The self-referential wrapper variant.
	// Before the fix, Internal__AddWrapperVariant eagerly calls
	// adapter.Serializer().TypeDescriptor(), which returns a nil *EnumDescriptor.
	Internal__AddWrapperVariant(
		adapter, 1, "e", kindE,
		adapter.Serializer(),
		"",
		func(v recEnum) recEnum { return recEnum{kind: kindE, value: v} },
		func(e recEnum) *recEnum {
			if v, ok := e.value.(recEnum); ok {
				return &v
			}
			return nil
		},
	)
	adapter.Finalize()

	// Must not panic.
	got := adapter.Serializer().TypeDescriptor().AsJson()
	// Sanity-check: the JSON must reference the record id "test:RecEnum".
	if !strings.Contains(got, "test:RecEnum") {
		t.Errorf("AsJson() = %q, expected it to contain %q", got, "test:RecEnum")
	}
}
