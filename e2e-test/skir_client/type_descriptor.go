// Package skir_client provides the reflection/introspection types for skir schemas.
//
// A TypeDescriptor describes a skir type (primitive, optional, array, struct or enum)
// and can serialize itself to a self-contained JSON representation, as well as be
// reconstructed from such a representation via ParseTypeDescriptorFromJson.
package skir_client

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// ─────────────────────────────────────────────────────────────────────────────
// PrimitiveType
// ─────────────────────────────────────────────────────────────────────────────

// PrimitiveType enumerates all primitive types supported by skir.
type PrimitiveType int

const (
	PrimitiveTypeBool PrimitiveType = iota
	PrimitiveTypeInt32
	PrimitiveTypeInt64
	PrimitiveTypeHash64
	PrimitiveTypeFloat32
	PrimitiveTypeFloat64
	PrimitiveTypeTimestamp
	PrimitiveTypeString
	PrimitiveTypeBytes
)

func (p PrimitiveType) String() string {
	switch p {
	case PrimitiveTypeBool:
		return "bool"
	case PrimitiveTypeInt32:
		return "int32"
	case PrimitiveTypeInt64:
		return "int64"
	case PrimitiveTypeHash64:
		return "hash64"
	case PrimitiveTypeFloat32:
		return "float32"
	case PrimitiveTypeFloat64:
		return "float64"
	case PrimitiveTypeTimestamp:
		return "timestamp"
	case PrimitiveTypeString:
		return "string"
	case PrimitiveTypeBytes:
		return "bytes"
	default:
		return fmt.Sprintf("PrimitiveType(%d)", int(p))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// TypeDescriptor – sealed interface
// ─────────────────────────────────────────────────────────────────────────────

// TypeDescriptor describes a skir type.
//
// The concrete types that implement TypeDescriptor are:
//   - *PrimitiveDescriptor
//   - *OptionalDescriptor
//   - *ArrayDescriptor
//   - *StructDescriptor
//   - *EnumDescriptor
//
// The unexported method sealTypeDescriptor prevents implementations outside
// this package.
type TypeDescriptor interface {
	sealTypeDescriptor()

	// AsJson returns the complete, self-describing JSON representation of this
	// type descriptor as a compact JSON string.
	AsJson() string
}

// ─────────────────────────────────────────────────────────────────────────────
// PrimitiveDescriptor
// ─────────────────────────────────────────────────────────────────────────────

// PrimitiveDescriptor describes a primitive skir type.
type PrimitiveDescriptor struct {
	PrimitiveType PrimitiveType
}

// Singleton instances – one per primitive type.
var (
	BoolDescriptor      = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeBool}
	Int32Descriptor     = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeInt32}
	Int64Descriptor     = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeInt64}
	Hash64Descriptor    = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeHash64}
	Float32Descriptor   = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeFloat32}
	Float64Descriptor   = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeFloat64}
	TimestampDescriptor = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeTimestamp}
	StringDescriptor    = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeString}
	BytesDescriptor     = &PrimitiveDescriptor{PrimitiveType: PrimitiveTypeBytes}
)

// primitiveDescriptorByType returns the singleton PrimitiveDescriptor for pt.
func primitiveDescriptorByType(pt PrimitiveType) *PrimitiveDescriptor {
	switch pt {
	case PrimitiveTypeBool:
		return BoolDescriptor
	case PrimitiveTypeInt32:
		return Int32Descriptor
	case PrimitiveTypeInt64:
		return Int64Descriptor
	case PrimitiveTypeHash64:
		return Hash64Descriptor
	case PrimitiveTypeFloat32:
		return Float32Descriptor
	case PrimitiveTypeFloat64:
		return Float64Descriptor
	case PrimitiveTypeTimestamp:
		return TimestampDescriptor
	case PrimitiveTypeString:
		return StringDescriptor
	case PrimitiveTypeBytes:
		return BytesDescriptor
	default:
		panic(fmt.Sprintf("unknown PrimitiveType %d", int(pt)))
	}
}

func (d *PrimitiveDescriptor) sealTypeDescriptor() {}
func (d *PrimitiveDescriptor) AsJson() string      { return typeDescriptorToJson(d) }

// ─────────────────────────────────────────────────────────────────────────────
// OptionalDescriptor
// ─────────────────────────────────────────────────────────────────────────────

// OptionalDescriptor describes an optional type that can hold either a value of
// the wrapped type or a null/zero value.
type OptionalDescriptor struct {
	// OtherType is the type descriptor for the wrapped non-null type.
	OtherType TypeDescriptor
}

func (d *OptionalDescriptor) sealTypeDescriptor() {}
func (d *OptionalDescriptor) AsJson() string      { return typeDescriptorToJson(d) }

// ─────────────────────────────────────────────────────────────────────────────
// ArrayDescriptor
// ─────────────────────────────────────────────────────────────────────────────

// ArrayDescriptor describes an ordered collection of elements of a single type.
type ArrayDescriptor struct {
	// ItemType describes the type of each element.
	ItemType TypeDescriptor

	// KeyExtractor, when non-empty, is the key chain specified in the '.skir'
	// file after the pipe character. It identifies a property of ItemType that
	// can be used for fast keyed lookup.
	KeyExtractor string
}

func (d *ArrayDescriptor) sealTypeDescriptor() {}
func (d *ArrayDescriptor) AsJson() string      { return typeDescriptorToJson(d) }

// ─────────────────────────────────────────────────────────────────────────────
// FieldOrVariant
// ─────────────────────────────────────────────────────────────────────────────

// FieldOrVariant is the common interface for struct fields and enum variants.
type FieldOrVariant interface {
	// Name of the field/variant as specified in the '.skir' file.
	GetName() string

	// Number used for binary serialization.
	GetNumber() int

	// Documentation extracted from doc comments in the '.skir' file.
	GetDoc() string
}

// ─────────────────────────────────────────────────────────────────────────────
// StructField
// ─────────────────────────────────────────────────────────────────────────────

// StructField describes a single field of a skir struct.
type StructField struct {
	Name   string
	Number int
	Type   TypeDescriptor
	Doc    string
}

func (f *StructField) GetName() string { return f.Name }
func (f *StructField) GetNumber() int  { return f.Number }
func (f *StructField) GetDoc() string  { return f.Doc }

// ─────────────────────────────────────────────────────────────────────────────
// EnumVariant – sealed interface
// ─────────────────────────────────────────────────────────────────────────────

// EnumVariant is the common interface for skir enum variants.
// Concrete types: *EnumConstantVariant, *EnumWrapperVariant.
type EnumVariant interface {
	FieldOrVariant
	sealEnumVariant()
}

// ─────────────────────────────────────────────────────────────────────────────
// EnumConstantVariant
// ─────────────────────────────────────────────────────────────────────────────

// EnumConstantVariant describes a constant (non-wrapping) enum variant.
type EnumConstantVariant struct {
	Name   string
	Number int
	Doc    string
}

func (v *EnumConstantVariant) GetName() string  { return v.Name }
func (v *EnumConstantVariant) GetNumber() int   { return v.Number }
func (v *EnumConstantVariant) GetDoc() string   { return v.Doc }
func (v *EnumConstantVariant) sealEnumVariant() {}

// ─────────────────────────────────────────────────────────────────────────────
// EnumWrapperVariant
// ─────────────────────────────────────────────────────────────────────────────

// EnumWrapperVariant describes an enum variant that wraps a value of another type.
type EnumWrapperVariant struct {
	Name   string
	Number int
	Type   TypeDescriptor
	Doc    string
}

func (v *EnumWrapperVariant) GetName() string  { return v.Name }
func (v *EnumWrapperVariant) GetNumber() int   { return v.Number }
func (v *EnumWrapperVariant) GetDoc() string   { return v.Doc }
func (v *EnumWrapperVariant) sealEnumVariant() {}

// ─────────────────────────────────────────────────────────────────────────────
// RecordDescriptorBase
// ─────────────────────────────────────────────────────────────────────────────

// RecordDescriptorBase provides common metadata for struct and enum descriptors.
type RecordDescriptorBase interface {
	TypeDescriptor

	// Name of the record as specified in the '.skir' file.
	GetName() string

	// QualifiedName contains all names in the hierarchical sequence, e.g.
	// "Foo.Bar" if Bar is nested within Foo, or simply "Bar" for top-level.
	GetQualifiedName() string

	// ModulePath is the path to the '.skir' file relative to the skir source root.
	GetModulePath() string

	// Doc is the documentation extracted from doc comments in the '.skir' file.
	GetDoc() string

	// RemovedNumbers returns the set of field/variant numbers that have been
	// marked as removed (reserved).
	GetRemovedNumbers() map[int]struct{}

	// recordId returns "modulePath:qualifiedName".
	recordId() string
}

// ─────────────────────────────────────────────────────────────────────────────
// StructDescriptor
// ─────────────────────────────────────────────────────────────────────────────

// StructDescriptor describes a skir struct type.
type StructDescriptor struct {
	name           string
	qualifiedName  string
	modulePath     string
	Doc            string
	RemovedNumbers map[int]struct{}
	Fields         []*StructField

	// Lazy lookup maps, initialised on first use.
	once          sync.Once
	nameToField   map[string]*StructField
	numberToField map[int]*StructField
}

func newStructDescriptor(modulePath, qualifiedName, doc string, removedNumbers map[int]struct{}, fields []*StructField) *StructDescriptor {
	name := qualifiedName
	if i := lastDotIndex(qualifiedName); i >= 0 {
		name = qualifiedName[i+1:]
	}
	return &StructDescriptor{
		name:           name,
		qualifiedName:  qualifiedName,
		modulePath:     modulePath,
		Doc:            doc,
		RemovedNumbers: removedNumbers,
		Fields:         fields,
	}
}

func (d *StructDescriptor) sealTypeDescriptor()                 {}
func (d *StructDescriptor) GetName() string                     { return d.name }
func (d *StructDescriptor) GetQualifiedName() string            { return d.qualifiedName }
func (d *StructDescriptor) GetModulePath() string               { return d.modulePath }
func (d *StructDescriptor) GetDoc() string                      { return d.Doc }
func (d *StructDescriptor) GetRemovedNumbers() map[int]struct{} { return d.RemovedNumbers }
func (d *StructDescriptor) recordId() string                    { return d.modulePath + ":" + d.qualifiedName }
func (d *StructDescriptor) AsJson() string                      { return typeDescriptorToJson(d) }

func (d *StructDescriptor) initLookups() {
	d.once.Do(func() {
		nm := make(map[string]*StructField, len(d.Fields))
		num := make(map[int]*StructField, len(d.Fields))
		for _, f := range d.Fields {
			nm[f.Name] = f
			num[f.Number] = f
		}
		d.nameToField = nm
		d.numberToField = num
	})
}

// GetFieldByName returns the field with the given name, or nil if not found.
func (d *StructDescriptor) GetFieldByName(name string) *StructField {
	d.initLookups()
	return d.nameToField[name]
}

// GetFieldByNumber returns the field with the given number, or nil if not found.
func (d *StructDescriptor) GetFieldByNumber(number int) *StructField {
	d.initLookups()
	return d.numberToField[number]
}

// ─────────────────────────────────────────────────────────────────────────────
// EnumDescriptor
// ─────────────────────────────────────────────────────────────────────────────

// EnumDescriptor describes a skir enum type.
type EnumDescriptor struct {
	name           string
	qualifiedName  string
	modulePath     string
	Doc            string
	RemovedNumbers map[int]struct{}
	Variants       []EnumVariant

	// Lazy lookup maps, initialised on first use.
	once            sync.Once
	nameToVariant   map[string]EnumVariant
	numberToVariant map[int]EnumVariant
}

func newEnumDescriptor(modulePath, qualifiedName, doc string, removedNumbers map[int]struct{}, variants []EnumVariant) *EnumDescriptor {
	name := qualifiedName
	if i := lastDotIndex(qualifiedName); i >= 0 {
		name = qualifiedName[i+1:]
	}
	return &EnumDescriptor{
		name:           name,
		qualifiedName:  qualifiedName,
		modulePath:     modulePath,
		Doc:            doc,
		RemovedNumbers: removedNumbers,
		Variants:       variants,
	}
}

func (d *EnumDescriptor) sealTypeDescriptor()                 {}
func (d *EnumDescriptor) GetName() string                     { return d.name }
func (d *EnumDescriptor) GetQualifiedName() string            { return d.qualifiedName }
func (d *EnumDescriptor) GetModulePath() string               { return d.modulePath }
func (d *EnumDescriptor) GetDoc() string                      { return d.Doc }
func (d *EnumDescriptor) GetRemovedNumbers() map[int]struct{} { return d.RemovedNumbers }
func (d *EnumDescriptor) recordId() string                    { return d.modulePath + ":" + d.qualifiedName }
func (d *EnumDescriptor) AsJson() string                      { return typeDescriptorToJson(d) }

func (d *EnumDescriptor) initLookups() {
	d.once.Do(func() {
		nm := make(map[string]EnumVariant, len(d.Variants))
		num := make(map[int]EnumVariant, len(d.Variants))
		for _, v := range d.Variants {
			nm[v.GetName()] = v
			num[v.GetNumber()] = v
		}
		d.nameToVariant = nm
		d.numberToVariant = num
	})
}

// GetVariantByName returns the variant with the given name, or nil if not found.
func (d *EnumDescriptor) GetVariantByName(name string) EnumVariant {
	d.initLookups()
	return d.nameToVariant[name]
}

// GetVariantByNumber returns the variant with the given number, or nil if not found.
func (d *EnumDescriptor) GetVariantByNumber(number int) EnumVariant {
	d.initLookups()
	return d.numberToVariant[number]
}

// ─────────────────────────────────────────────────────────────────────────────
// JSON serialization
// ─────────────────────────────────────────────────────────────────────────────

// typeDescriptorToJson serialises td to a compact JSON string:
//
//	{"type":<type-signature>,"records":[<record-definition>,...] }
func typeDescriptorToJson(td TypeDescriptor) string {
	var out strings.Builder

	var order []string
	recordIdToJson := map[string]string{}

	var addRecordDefinitions func(t TypeDescriptor)
	addRecordDefinitions = func(t TypeDescriptor) {
		switch v := t.(type) {
		case *PrimitiveDescriptor:
			// no record definitions
		case *OptionalDescriptor:
			addRecordDefinitions(v.OtherType)
		case *ArrayDescriptor:
			addRecordDefinitions(v.ItemType)
		case *StructDescriptor:
			rid := v.recordId()
			if _, seen := recordIdToJson[rid]; seen {
				return
			}
			recordIdToJson[rid] = "" // mark as in-progress to break cycles
			var sb strings.Builder
			sb.WriteString(`{"kind":"struct","id":`)
			writeJsonEscapedString(rid, &sb)
			if v.Doc != "" {
				sb.WriteString(`,"doc":`)
				writeJsonEscapedString(v.Doc, &sb)
			}
			sb.WriteString(`,"fields":[`)
			for i, f := range v.Fields {
				if i > 0 {
					sb.WriteByte(',')
				}
				sb.WriteString(`{"name":`)
				writeJsonEscapedString(f.Name, &sb)
				sb.WriteString(`,"number":`)
				sb.WriteString(strconv.Itoa(f.Number))
				sb.WriteString(`,"type":`)
				typeSignatureToJson(f.Type, &sb)
				if f.Doc != "" {
					sb.WriteString(`,"doc":`)
					writeJsonEscapedString(f.Doc, &sb)
				}
				sb.WriteByte('}')
			}
			sb.WriteByte(']')
			removedSlice := removedNumbersToSortedSlice(v.RemovedNumbers)
			if len(removedSlice) > 0 {
				sb.WriteString(`,"removed_numbers":[`)
				for i, n := range removedSlice {
					if i > 0 {
						sb.WriteByte(',')
					}
					sb.WriteString(strconv.Itoa(n))
				}
				sb.WriteByte(']')
			}
			sb.WriteByte('}')
			recordIdToJson[rid] = sb.String()
			order = append(order, rid)
			for _, f := range v.Fields {
				addRecordDefinitions(f.Type)
			}
		case *EnumDescriptor:
			rid := v.recordId()
			if _, seen := recordIdToJson[rid]; seen {
				return
			}
			recordIdToJson[rid] = "" // mark as in-progress to break cycles
			var sb strings.Builder
			sb.WriteString(`{"kind":"enum","id":`)
			writeJsonEscapedString(rid, &sb)
			if v.Doc != "" {
				sb.WriteString(`,"doc":`)
				writeJsonEscapedString(v.Doc, &sb)
			}
			sb.WriteString(`,"variants":[`)
			for i, variant := range v.Variants {
				if i > 0 {
					sb.WriteByte(',')
				}
				sb.WriteString(`{"name":`)
				writeJsonEscapedString(variant.GetName(), &sb)
				sb.WriteString(`,"number":`)
				sb.WriteString(strconv.Itoa(variant.GetNumber()))
				if w, ok := variant.(*EnumWrapperVariant); ok {
					sb.WriteString(`,"type":`)
					typeSignatureToJson(w.Type, &sb)
				}
				if variant.GetDoc() != "" {
					sb.WriteString(`,"doc":`)
					writeJsonEscapedString(variant.GetDoc(), &sb)
				}
				sb.WriteByte('}')
			}
			sb.WriteByte(']')
			removedSlice := removedNumbersToSortedSlice(v.RemovedNumbers)
			if len(removedSlice) > 0 {
				sb.WriteString(`,"removed_numbers":[`)
				for i, n := range removedSlice {
					if i > 0 {
						sb.WriteByte(',')
					}
					sb.WriteString(strconv.Itoa(n))
				}
				sb.WriteByte(']')
			}
			sb.WriteByte('}')
			recordIdToJson[rid] = sb.String()
			order = append(order, rid)
			for _, variant := range v.Variants {
				if w, ok := variant.(*EnumWrapperVariant); ok {
					addRecordDefinitions(w.Type)
				}
			}
		}
	}

	addRecordDefinitions(td)

	out.WriteString(`{"type":`)
	typeSignatureToJson(td, &out)
	out.WriteString(`,"records":[`)
	for i, id := range order {
		if i > 0 {
			out.WriteByte(',')
		}
		out.WriteString(recordIdToJson[id])
	}
	out.WriteString(`]}`)
	return out.String()
}

// typeSignatureToJson writes the compact "type signature" JSON for td to out.
// This is the value stored under the "type" key of fields and of the root document.
func typeSignatureToJson(td TypeDescriptor, out *strings.Builder) {
	switch v := td.(type) {
	case *PrimitiveDescriptor:
		out.WriteString(`{"kind":"primitive","value":"`)
		out.WriteString(v.PrimitiveType.String())
		out.WriteString(`"}`)
	case *OptionalDescriptor:
		out.WriteString(`{"kind":"optional","value":`)
		typeSignatureToJson(v.OtherType, out)
		out.WriteByte('}')
	case *ArrayDescriptor:
		out.WriteString(`{"kind":"array","value":{"item":`)
		typeSignatureToJson(v.ItemType, out)
		if v.KeyExtractor != "" {
			out.WriteString(`,"key_extractor":`)
			writeJsonEscapedString(v.KeyExtractor, out)
		}
		out.WriteString(`}}`)
	case *StructDescriptor:
		out.WriteString(`{"kind":"record","value":`)
		writeJsonEscapedString(v.recordId(), out)
		out.WriteByte('}')
	case *EnumDescriptor:
		out.WriteString(`{"kind":"record","value":`)
		writeJsonEscapedString(v.recordId(), out)
		out.WriteByte('}')
	default:
		panic(fmt.Sprintf("skir_client: typeSignatureToJson: unknown TypeDescriptor %T", td))
	}
}

// removedNumbersToSortedSlice converts the set to a sorted int slice.
func removedNumbersToSortedSlice(m map[int]struct{}) []int {
	if len(m) == 0 {
		return nil
	}
	s := make([]int, 0, len(m))
	for n := range m {
		s = append(s, n)
	}
	sort.Ints(s)
	return s
}

// lastDotIndex returns the index of the last '.' in s, or -1 if absent.
func lastDotIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}
