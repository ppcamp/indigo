package indigo

import (
	"fmt"
	"strings"
)

// Schema defines the keys (variable names) and their data types used in a
// rule expression. The same keys and types must be supplied in the data map
// when rules are evaluated.
type Schema struct {
	// Identifier for the schema. Useful for the hosting application; not used by Indigo internally.
	ID string `json:"id,omitempty"`
	// User-friendly name for the schema
	Name string `json:"name,omitempty"`
	// A user-friendly description of the schema
	Description string `json:"description,omitempty"`
	// User-defined value
	Meta interface{} `json:"-"`
	// List of data elements supported by this schema
	Elements []DataElement `json:"elements,omitempty"`
}

func (s *Schema) String() string {
	x := strings.Builder{}
	x.WriteString(s.ID)
	if s.Name != "" {
		x.WriteString("  '" + s.Name + "'")
	}
	x.WriteString("\n")
	for _, e := range s.Elements {
		x.WriteString(e.String())
		x.WriteString("\n")
	}

	return x.String()
}

func (e *DataElement) String() string {
	return fmt.Sprintf("  %s (%s)", e.Name, e.Type)
}

// DataElement defines a named variable in a schema
type DataElement struct {
	// Short, user-friendly name of the variable. This is the name
	// that will be used in rules to refer to data passed in.
	//
	// RESERVED NAMES:
	//   selfKey (see const)
	Name string `json:"name"`

	// One of the Type interface defined.
	Type Type `json:"type"`

	// Optional description of the type.
	Description string `json:"description"`
}

// Type defines a type in the Indigo type system.
// These types are used to define schemas, define required
// evaluation results, and to interpret evaluation results.
// Not all implementations of Evaluator support all types.
type Type interface {
	String() string
}

// String defines an Indigo string type.
type String struct{}

// Int defines an Indigo int type. The exact "Int" implementation and size
// depends on the evaluator used.
type Int struct{}

// Float defines an Indigo float type. The implementation of the float (size, precision)
// depends on the evaluator used.
type Float struct{}

// Any defines an Indigo type for an "undefined" or unspecified type.
type Any struct{}

// Bool defines an Indigo type for true/false.
type Bool struct{}

// Duration defines an Indigo type for the time.Duration type.
type Duration struct{}

// Timestamp defines an Indigo type for the time.Time type.
type Timestamp struct{}

// Proto defines an Indigo type for a protobuf type.
type Proto struct {
	Protoname string      // fully qualified name of the protobuf type
	Message   interface{} // an empty protobuf instance of the type
}

// List defines an Indigo type representing a slice of values
type List struct {
	ValueType Type // the type of element stored in the list
}

// Map defines an Indigo type representing a map of keys and values.
type Map struct {
	KeyType   Type // the type of the map key
	ValueType Type // the type of the value stored in the map
}

func (t Int) String() string       { return "int" }
func (t Bool) String() string      { return "bool" }
func (t String) String() string    { return "string" }
func (t List) String() string      { return "[]" + t.ValueType.String() }
func (t Map) String() string       { return "map[" + t.KeyType.String() + "]" + t.ValueType.String() }
func (t Any) String() string       { return "any" }
func (t Duration) String() string  { return "duration" }
func (t Timestamp) String() string { return "timestamp" }
func (t Float) String() string     { return "float" }
func (t Proto) String() string     { return "proto(" + t.Protoname + ")" }

// Value is the result of evaluation returned in the Result.
// Inspect the Type to determine what it is.
type Value struct {
	Val  interface{} // the value stored
	Type Type        // the Indigo type stored
}

// ParseType parses a string that represents an Indigo type and returns the type.
// The primitive types are their lower-case names (string, int, duration, etc.)
// Maps and lists look like Go maps and slices: map[string]float and []string.
// Proto types look like this: proto(protoname)
func ParseType(t string) (Type, error) {

	if strings.Contains(t, "map") {
		return parseMap(t)
	}

	if strings.Contains(t, "[]") {
		return parseList(t)
	}

	if strings.Contains(t, "proto(") {
		return parseProto(t)
	}

	switch t {
	case "string":
		return String{}, nil
	case "int":
		return Int{}, nil
	case "float":
		return Float{}, nil
	case "bool":
		return Bool{}, nil
	case "duration":
		return Duration{}, nil
	case "timestamp":
		return Timestamp{}, nil
	case "any":
		return Any{}, nil
	default:
		return Any{}, fmt.Errorf("unrecognized type: %s", t)
	}
}

// parseMap parses a string and returns an Indigo map type.
// The string must in the format map[<keytype]<valuetype>.
// Example: map[string]int
func parseMap(t string) (Type, error) {
	startBracket := strings.Index(t, "[")
	endBracket := strings.Index(t, "]")
	if startBracket == -1 || endBracket == -1 || startBracket > endBracket || endBracket > len(t) {
		return nil, fmt.Errorf("bad map specification")
	}
	keyTypeName := t[startBracket+1 : endBracket]
	valueTypeName := t[endBracket+1:]
	keyType, err := ParseType(keyTypeName)
	if err != nil {
		return Any{}, err
	}
	valueType, err := ParseType(valueTypeName)
	if err != nil {
		return Any{}, err
	}

	return Map{
		KeyType:   keyType,
		ValueType: valueType,
	}, nil

}

// parseList parses a string and returns an Indigo list type.
// The string must be in the format []<valuetype>
// Example: []string
func parseList(t string) (Type, error) {
	typeName := t[2:]
	valueType, err := ParseType(typeName)
	if err != nil {
		return Any{}, err
	}

	return List{
		ValueType: valueType,
	}, nil

}

// parseProto parses a string and returns a partial Indigo proto type.
// The "Message" field of the proto struct must be suppplied later.
// The string must be in the form proto(<protoname>).
// Example: proto("school.Student")
func parseProto(t string) (Type, error) {

	startParen := strings.Index(t, "(")
	endParen := strings.Index(t, ")")

	if startParen == -1 || endParen == -1 || startParen > endParen || endParen > len(t) || endParen-startParen == 1 {
		return nil, fmt.Errorf("bad proto specification")
	}

	name := t[startParen+1 : endParen]
	return Proto{Protoname: name}, nil
}