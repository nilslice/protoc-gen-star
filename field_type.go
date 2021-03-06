package pgs

// FieldType describes the type of a Field.
type FieldType interface {
	// Field returns the parent Field of this type. While two FieldTypes might be
	// equivalent, each instance of a FieldType is tied to its Field.
	Field() Field

	// Name returns the TypeName for this Field, which represents the type of the
	// field as it would exist in Go source code.
	Name() TypeName

	// IsRepeated returns true if and only if the field is marked as "repeated".
	// While map fields may be labeled as repeated, this method will not return
	// true for them.
	IsRepeated() bool

	// IsMap returns true if the field is a map type.
	IsMap() bool

	// IsEnum returns true if the field is a singular enum value. Maps or
	// repeated fields containing enums will still return false.
	IsEnum() bool

	// IsEmbed returns true if the field is a singular message value. Maps or
	// repeated fields containing embeds will still return false.
	IsEmbed() bool

	// IsOptional returns true if the message's syntax is not Proto2 or
	// the field is prefixed as optional.
	IsOptional() bool

	// IsRequired returns true if and only if the field is prefixed as required.
	IsRequired() bool

	// IsSlice returns true if the field is represented in Go as a slice. This
	// method returns true only for repeated and bytes-type fields.
	IsSlice() bool

	// ProtoType returns the ProtoType value for this field.
	ProtoType() ProtoType

	// ProtoLabel returns the ProtoLabel value for this field.
	ProtoLabel() ProtoLabel

	// Imports includes all external packages required by this field.
	Imports() []Package

	// Enum returns the Enum associated with this FieldType. If IsEnum returns
	// false, this value will be nil.
	Enum() Enum

	// Embed returns the embedded Message associated with this FieldType. If
	// IsEmbed returns false, this value will be nil.
	Embed() Message

	// Element returns the FieldTypeElem representing the element component of
	// the type. Nil will be returned if IsRepeated and IsMap return false.
	Element() FieldTypeElem

	// Key returns the FieldTypeElem representing the key component of the type.
	// Nil will be return sif IsMap returns false.
	Key() FieldTypeElem

	setField(f Field)
	toElem() FieldTypeElem
}

type scalarT struct {
	fld  Field
	name TypeName
}

func (s *scalarT) Field() Field           { return s.fld }
func (s *scalarT) IsRepeated() bool       { return false }
func (s *scalarT) IsMap() bool            { return false }
func (s *scalarT) IsEnum() bool           { return false }
func (s *scalarT) IsEmbed() bool          { return false }
func (s *scalarT) Name() TypeName         { return s.name }
func (s *scalarT) IsSlice() bool          { return s.ProtoType().IsSlice() }
func (s *scalarT) ProtoType() ProtoType   { return ProtoType(s.fld.Descriptor().GetType()) }
func (s *scalarT) ProtoLabel() ProtoLabel { return ProtoLabel(s.fld.Descriptor().GetLabel()) }
func (s *scalarT) Imports() []Package     { return nil }
func (s *scalarT) setField(f Field)       { s.fld = f }
func (s *scalarT) Enum() Enum             { return nil }
func (s *scalarT) Embed() Message         { return nil }
func (s *scalarT) Element() FieldTypeElem { return nil }
func (s *scalarT) Key() FieldTypeElem     { return nil }

func (s *scalarT) IsOptional() bool {
	return !s.fld.Syntax().SupportsRequiredPrefix() || s.ProtoLabel() == Optional
}

func (s *scalarT) IsRequired() bool {
	return s.fld.Syntax().SupportsRequiredPrefix() && s.ProtoLabel() == Required
}

func (s *scalarT) toElem() FieldTypeElem {
	return &scalarE{
		typ:   s,
		ptype: s.ProtoType(),
		name:  s.name,
	}
}

type enumT struct {
	*scalarT
	enum Enum
}

func (e *enumT) Enum() Enum   { return e.enum }
func (e *enumT) IsEnum() bool { return true }

func (e *enumT) Imports() []Package {
	if pkg := e.enum.Package(); pkg.GoName() != e.fld.Package().GoName() {
		return []Package{pkg}
	}
	return nil
}

func (e *enumT) toElem() FieldTypeElem {
	return &enumE{
		scalarE: e.scalarT.toElem().(*scalarE),
		enum:    e.enum,
	}
}

type embedT struct {
	*scalarT
	msg Message
}

func (e *embedT) Embed() Message { return e.msg }
func (e *embedT) IsEmbed() bool  { return true }

func (e *embedT) Imports() []Package {
	if pkg := e.msg.Package(); pkg.GoName() != e.fld.Package().GoName() {
		return []Package{pkg}
	}
	return nil
}

func (e *embedT) toElem() FieldTypeElem {
	return &embedE{
		scalarE: e.scalarT.toElem().(*scalarE),
		msg:     e.msg,
	}
}

type repT struct {
	*scalarT
	el FieldTypeElem
}

func (r *repT) IsRepeated() bool       { return true }
func (r *repT) Element() FieldTypeElem { return r.el }
func (r *repT) IsSlice() bool          { return true }

func (r *repT) Imports() []Package { return r.el.Imports() }

func (r *repT) toElem() FieldTypeElem { panic("cannot convert repeated FieldType to FieldTypeElem") }

type mapT struct {
	*repT
	key FieldTypeElem
}

func (m *mapT) IsRepeated() bool   { return false }
func (m *mapT) IsMap() bool        { return true }
func (m *mapT) IsSlice() bool      { return false }
func (m *mapT) Key() FieldTypeElem { return m.key }

var (
	_ FieldType = (*scalarT)(nil)
	_ FieldType = (*enumT)(nil)
	_ FieldType = (*embedT)(nil)
	_ FieldType = (*repT)(nil)
	_ FieldType = (*mapT)(nil)
)
