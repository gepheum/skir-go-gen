package skir_client

import (
	"testing"
)

// helpers shared by serializer tests
func mustFromJson[T any](t *testing.T, s Serializer[T], code string) T {
	t.Helper()
	v, err := s.FromJson(code)
	if err != nil {
		t.Fatalf("FromJson(%q): %v", code, err)
	}
	return v
}

func mustFromBytes[T any](t *testing.T, s Serializer[T], b []byte) T {
	t.Helper()
	v, err := s.FromBytes(b)
	if err != nil {
		t.Fatalf("FromBytes: %v", err)
	}
	return v
}

// ─────────────────────────────────────────────────────────────────────────────
// BoolSerializer – ToJson / FromJson
// ─────────────────────────────────────────────────────────────────────────────

func TestBoolSerializer_ToJson(t *testing.T) {
	s := BoolSerializer()
	if got := s.ToJson(true); got != "1" {
		t.Errorf("ToJson(true) = %q, want %q", got, "1")
	}
	if got := s.ToJson(false); got != "0" {
		t.Errorf("ToJson(false) = %q, want %q", got, "0")
	}
}

func TestBoolSerializer_FromJson_trueFalse(t *testing.T) {
	s := BoolSerializer()
	if got := mustFromJson(t, s, "true"); got != true {
		t.Error("FromJson('true') = false, want true")
	}
	if got := mustFromJson(t, s, "false"); got != false {
		t.Error("FromJson('false') = true, want false")
	}
}

func TestBoolSerializer_FromJson_numericOneZero(t *testing.T) {
	s := BoolSerializer()
	if got := mustFromJson(t, s, "1"); !got {
		t.Error("FromJson('1') = false, want true")
	}
	if got := mustFromJson(t, s, "0"); got {
		t.Error("FromJson('0') = true, want false")
	}
}

func TestBoolSerializer_FromJson_nonZeroNumber(t *testing.T) {
	s := BoolSerializer()
	if got := mustFromJson(t, s, "42"); !got {
		t.Error("FromJson('42') = false, want true")
	}
	if got := mustFromJson(t, s, "-1"); !got {
		t.Error("FromJson('-1') = false, want true")
	}
}

func TestBoolSerializer_FromJson_invalidJson(t *testing.T) {
	_, err := BoolSerializer().FromJson("not-json")
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BoolSerializer – ToJson / FromJson round-trip
// ─────────────────────────────────────────────────────────────────────────────

func TestBoolSerializer_JsonRoundTrip(t *testing.T) {
	s := BoolSerializer()
	for _, v := range []bool{true, false} {
		code := s.ToJson(v)
		got := mustFromJson(t, s, code)
		if got != v {
			t.Errorf("JSON round-trip(%v): got %v", v, got)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BoolSerializer – ToBytes / FromBytes
// ─────────────────────────────────────────────────────────────────────────────

func TestBoolSerializer_ToBytes_hasSkírPrefix(t *testing.T) {
	s := BoolSerializer()
	b := s.ToBytes(true)
	if len(b) < 4 || string(b[:4]) != "skir" {
		t.Errorf("ToBytes: missing 'skir' prefix, got %v", b)
	}
}

func TestBoolSerializer_ToBytes_trueEncoding(t *testing.T) {
	s := BoolSerializer()
	b := s.ToBytes(true)
	// "skir" (4 bytes) + one byte payload: 0x01
	if len(b) != 5 {
		t.Fatalf("ToBytes(true) len = %d, want 5", len(b))
	}
	if b[4] != 0x01 {
		t.Errorf("ToBytes(true) payload = %#x, want 0x01", b[4])
	}
}

func TestBoolSerializer_ToBytes_falseEncoding(t *testing.T) {
	s := BoolSerializer()
	b := s.ToBytes(false)
	if len(b) != 5 {
		t.Fatalf("ToBytes(false) len = %d, want 5", len(b))
	}
	if b[4] != 0x00 {
		t.Errorf("ToBytes(false) payload = %#x, want 0x00", b[4])
	}
}

func TestBoolSerializer_BinaryRoundTrip(t *testing.T) {
	s := BoolSerializer()
	for _, v := range []bool{true, false} {
		b := s.ToBytes(v)
		got := mustFromBytes(t, s, b)
		if got != v {
			t.Errorf("binary round-trip(%v): got %v", v, got)
		}
	}
}

func TestBoolSerializer_FromBytes_fallsBackToJson(t *testing.T) {
	// FromBytes treats payloads without the "skir" prefix as JSON.
	s := BoolSerializer()
	if got := mustFromBytes(t, s, []byte("true")); !got {
		t.Error("FromBytes('true' as JSON) = false, want true")
	}
	if got := mustFromBytes(t, s, []byte("false")); got {
		t.Error("FromBytes('false' as JSON) = true, want false")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BoolSerializer – TypeDescriptor
// ─────────────────────────────────────────────────────────────────────────────

func TestBoolSerializer_TypeDescriptor(t *testing.T) {
	d := BoolSerializer().TypeDescriptor()
	pd, ok := d.(*PrimitiveDescriptor)
	if !ok {
		t.Fatalf("TypeDescriptor type = %T, want *PrimitiveDescriptor", d)
	}
	if pd.PrimitiveType != PrimitiveTypeBool {
		t.Errorf("PrimitiveType = %v, want PrimitiveTypeBool", pd.PrimitiveType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BoolSerializer – Readable{} flavor
// ─────────────────────────────────────────────────────────────────────────────

// boolAdapter ignores the readable flag, so ToJson with Readable{} still
// produces "1"/"0" (just routed through json.Indent, which is a no-op for
// single-token values).
func TestBoolSerializer_ToJson_readableFlavor(t *testing.T) {
	s := BoolSerializer()
	if got := s.ToJson(true, Readable{}); got != "1" {
		t.Errorf("ToJson(true, Readable{}) = %q, want %q", got, "1")
	}
	if got := s.ToJson(false, Readable{}); got != "0" {
		t.Errorf("ToJson(false, Readable{}) = %q, want %q", got, "0")
	}
}

func TestBoolSerializer_ToJson_readableFlavorMatchesDense(t *testing.T) {
	s := BoolSerializer()
	for _, v := range []bool{true, false} {
		dense := s.ToJson(v)
		readable := s.ToJson(v, Readable{})
		if dense != readable {
			t.Errorf("bool %v: dense=%q readable=%q, expected equal", v, dense, readable)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Int32Serializer
// ─────────────────────────────────────────────────────────────────────────────

func TestInt32Serializer_ToJson(t *testing.T) {
	s := Int32Serializer()
	cases := []struct {
		v    int32
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{2147483647, "2147483647"},
		{-2147483648, "-2147483648"},
	}
	for _, tc := range cases {
		if got := s.ToJson(tc.v); got != tc.want {
			t.Errorf("ToJson(%d) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestInt32Serializer_FromJson(t *testing.T) {
	s := Int32Serializer()
	cases := []struct {
		code string
		want int32
	}{
		{"0", 0},
		{"42", 42},
		{"-1", -1},
		{"2147483647", 2147483647},
		{"-2147483648", -2147483648},
		// string input
		{`"100"`, 100},
		{`"-5"`, -5},
		// unknown JSON type → 0
		{"null", 0},
	}
	for _, tc := range cases {
		got := mustFromJson(t, s, tc.code)
		if got != tc.want {
			t.Errorf("FromJson(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
}

func TestInt32Serializer_JsonRoundTrip(t *testing.T) {
	s := Int32Serializer()
	for _, v := range []int32{0, 1, -1, 231, 232, 255, 256, 65535, 65536, 2147483647, -256, -257, -65536, -65537, -2147483648} {
		code := s.ToJson(v)
		got := mustFromJson(t, s, code)
		if got != v {
			t.Errorf("JSON round-trip(%d): got %d", v, got)
		}
	}
}

func TestInt32Serializer_BinaryRoundTrip(t *testing.T) {
	s := Int32Serializer()
	// covers all wire-format branches
	for _, v := range []int32{0, 1, 231, 232, 65535, 65536, 2147483647, -1, -255, -256, -257, -65535, -65536, -65537, -2147483648} {
		b := s.ToBytes(v)
		got := mustFromBytes(t, s, b)
		if got != v {
			t.Errorf("binary round-trip(%d): got %d", v, got)
		}
	}
}

func TestInt32Serializer_TypeDescriptor(t *testing.T) {
	d := Int32Serializer().TypeDescriptor()
	pd, ok := d.(*PrimitiveDescriptor)
	if !ok {
		t.Fatalf("TypeDescriptor type = %T, want *PrimitiveDescriptor", d)
	}
	if pd.PrimitiveType != PrimitiveTypeInt32 {
		t.Errorf("PrimitiveType = %v, want PrimitiveTypeInt32", pd.PrimitiveType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Int64Serializer
// ─────────────────────────────────────────────────────────────────────────────

func TestInt64Serializer_ToJson_safeIntegers(t *testing.T) {
	s := Int64Serializer()
	// Only assert exact JSON text for values fastjson emits as plain integers.
	cases := []struct {
		v    int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
	}
	for _, tc := range cases {
		if got := s.ToJson(tc.v); got != tc.want {
			t.Errorf("ToJson(%d) = %q, want %q", tc.v, got, tc.want)
		}
	}
	// Values within the safe range must NOT be wrapped in quotes.
	for _, v := range []int64{9007199254740991, -9007199254740991} {
		code := s.ToJson(v)
		if len(code) > 0 && code[0] == '"' {
			t.Errorf("ToJson(%d) = %q: safe integer should not be quoted", v, code)
		}
	}
}

func TestInt64Serializer_ToJson_unsafeIntegers(t *testing.T) {
	s := Int64Serializer()
	// Values outside safe range → JSON string
	cases := []struct {
		v    int64
		want string
	}{
		{9007199254740991, "9007199254740991"},
		{-9007199254740991, "-9007199254740991"},
		{9007199254740992, `"9007199254740992"`},
		{-9007199254740992, `"-9007199254740992"`},
		{9223372036854775807, `"9223372036854775807"`},   // math.MaxInt64
		{-9223372036854775808, `"-9223372036854775808"`}, // math.MinInt64
	}
	for _, tc := range cases {
		if got := s.ToJson(tc.v); got != tc.want {
			t.Errorf("ToJson(%d) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestInt64Serializer_FromJson(t *testing.T) {
	s := Int64Serializer()
	cases := []struct {
		code string
		want int64
	}{
		{"0", 0},
		{"42", 42},
		{"-1", -1},
		// large value encoded as string
		{`"9007199254740992"`, 9007199254740992},
		{`"-9007199254740992"`, -9007199254740992},
		{`"9223372036854775807"`, 9223372036854775807},
		// unknown type → 0
		{"null", 0},
	}
	for _, tc := range cases {
		got := mustFromJson(t, s, tc.code)
		if got != tc.want {
			t.Errorf("FromJson(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
}

func TestInt64Serializer_JsonRoundTrip(t *testing.T) {
	s := Int64Serializer()
	for _, v := range []int64{0, 1, -1, 9007199254740991, -9007199254740991, 9007199254740992, -9007199254740992, 9223372036854775807, -9223372036854775808} {
		code := s.ToJson(v)
		got := mustFromJson(t, s, code)
		if got != v {
			t.Errorf("JSON round-trip(%d): got %d", v, got)
		}
	}
}

func TestInt64Serializer_BinaryRoundTrip(t *testing.T) {
	s := Int64Serializer()
	// covers: zero, int32 range reuse, wire-238 large values
	for _, v := range []int64{0, 1, -1, 231, 232, 2147483647, -2147483648, 2147483648, -2147483649, 9223372036854775807, -9223372036854775808} {
		b := s.ToBytes(v)
		got := mustFromBytes(t, s, b)
		if got != v {
			t.Errorf("binary round-trip(%d): got %d", v, got)
		}
	}
}

func TestInt64Serializer_TypeDescriptor(t *testing.T) {
	d := Int64Serializer().TypeDescriptor()
	pd, ok := d.(*PrimitiveDescriptor)
	if !ok {
		t.Fatalf("TypeDescriptor type = %T, want *PrimitiveDescriptor", d)
	}
	if pd.PrimitiveType != PrimitiveTypeInt64 {
		t.Errorf("PrimitiveType = %v, want PrimitiveTypeInt64", pd.PrimitiveType)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Hash64Serializer
// ─────────────────────────────────────────────────────────────────────────────

func TestHash64Serializer_ToJson_safeIntegers(t *testing.T) {
	s := Hash64Serializer()
	// Only assert exact JSON text for values fastjson emits as plain integers.
	cases := []struct {
		v    uint64
		want string
	}{
		{0, "0"},
		{1, "1"},
	}
	for _, tc := range cases {
		if got := s.ToJson(tc.v); got != tc.want {
			t.Errorf("ToJson(%d) = %q, want %q", tc.v, got, tc.want)
		}
	}
	// MAX_SAFE_INTEGER must NOT be wrapped in quotes.
	code := s.ToJson(9007199254740991)
	if len(code) > 0 && code[0] == '"' {
		t.Errorf("ToJson(MAX_SAFE_INTEGER) = %q: should not be quoted", code)
	}
}

func TestHash64Serializer_ToJson_unsafeIntegers(t *testing.T) {
	s := Hash64Serializer()
	cases := []struct {
		v    uint64
		want string
	}{
		{9007199254740992, `"9007199254740992"`},
		{18446744073709551615, `"18446744073709551615"`}, // math.MaxUint64
	}
	for _, tc := range cases {
		if got := s.ToJson(tc.v); got != tc.want {
			t.Errorf("ToJson(%d) = %q, want %q", tc.v, got, tc.want)
		}
	}
}

func TestHash64Serializer_FromJson(t *testing.T) {
	s := Hash64Serializer()
	cases := []struct {
		code string
		want uint64
	}{
		{"0", 0},
		{"255", 255},
		// negative number → 0
		{"-1", 0},
		// large value as string
		{`"9007199254740992"`, 9007199254740992},
		{`"18446744073709551615"`, 18446744073709551615},
		// unknown type → 0
		{"null", 0},
	}
	for _, tc := range cases {
		got := mustFromJson(t, s, tc.code)
		if got != tc.want {
			t.Errorf("FromJson(%q) = %d, want %d", tc.code, got, tc.want)
		}
	}
}

func TestHash64Serializer_JsonRoundTrip(t *testing.T) {
	s := Hash64Serializer()
	for _, v := range []uint64{0, 1, 231, 9007199254740991, 9007199254740992, 18446744073709551615} {
		code := s.ToJson(v)
		got := mustFromJson(t, s, code)
		if got != v {
			t.Errorf("JSON round-trip(%d): got %d", v, got)
		}
	}
}

func TestHash64Serializer_BinaryRoundTrip(t *testing.T) {
	s := Hash64Serializer()
	// covers all wire branches: single-byte, uint16, uint32, uint64
	for _, v := range []uint64{0, 1, 231, 232, 65535, 65536, 4294967295, 4294967296, 18446744073709551615} {
		b := s.ToBytes(v)
		got := mustFromBytes(t, s, b)
		if got != v {
			t.Errorf("binary round-trip(%d): got %d", v, got)
		}
	}
}

func TestHash64Serializer_TypeDescriptor(t *testing.T) {
	d := Hash64Serializer().TypeDescriptor()
	pd, ok := d.(*PrimitiveDescriptor)
	if !ok {
		t.Fatalf("TypeDescriptor type = %T, want *PrimitiveDescriptor", d)
	}
	if pd.PrimitiveType != PrimitiveTypeHash64 {
		t.Errorf("PrimitiveType = %v, want PrimitiveTypeHash64", pd.PrimitiveType)
	}
}
