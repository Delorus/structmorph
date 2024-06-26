package allsupportedtypes

type Type struct {
	StringField  string
	IntField     int
	Int8Field    int8
	Int16Field   int16
	Int32Field   int32
	Int64Field   int64
	UintField    uint
	Uint8Field   uint8
	Uint16Field  uint16
	Uint32Field  uint32
	Uint64Field  uint64
	FloatField   float64
	F32Field     float32
	BoolField    bool
	ByteField    byte
	RuneField    rune
	SliceField   []string
	ArrayField   [3]string
	MapField     map[string]string
	PointerField *string
	StructField  struct {
		InnerStringField string
	}
	InterfaceField interface{}
	AnotherStruct  AnotherStruct
	CompositeType
}

type AnotherStruct struct {
	AnotherField string
}

type CompositeType struct {
	CompositeField string
}

//go:generate go run ../../../cmd/structmorph/structmorph.go --from=allsupportedtypes.Type --to=allsupportedtypes.TypeDTO
type TypeDTO struct {
	StringField  string
	IntField     int
	Int8Field    int8
	Int16Field   int16
	Int32Field   int32
	Int64Field   int64
	UintField    uint
	Uint8Field   uint8
	Uint16Field  uint16
	Uint32Field  uint32
	Uint64Field  uint64
	FloatField   float64
	F32Field     float32
	BoolField    bool
	ByteField    byte
	RuneField    rune
	SliceField   []string
	ArrayField   [3]string
	MapField     map[string]string
	PointerField *string
	StructField  struct {
		InnerStringField string
	}
	InterfaceField interface{}
	AnotherStruct  AnotherStruct
	CompositeType
}
