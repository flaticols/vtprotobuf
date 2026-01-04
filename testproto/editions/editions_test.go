package editions

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestImplicitFieldPresenceUnmarshal tests that Edition 2023 fields with IMPLICIT presence
// are correctly unmarshaled as direct (non-pointer) assignments
func TestImplicitFieldPresenceUnmarshal(t *testing.T) {
	// Create a message with all fields set
	original := &ImplicitFieldPresence{
		CurrencyCode: "USD",
		Units:        12345,
		Scale:        100,
		IsActive:     true,
		Rate:         1.5,
		Amount:       999.99,
	}

	// Marshal using VT
	data, err := original.MarshalVT()
	require.NoError(t, err)

	// Unmarshal into a new message using VT
	decoded := &ImplicitFieldPresence{}
	err = decoded.UnmarshalVT(data)
	require.NoError(t, err)

	// Verify all fields were decoded correctly (direct assignment, not pointer)
	require.Equal(t, "USD", decoded.CurrencyCode)
	require.Equal(t, int64(12345), decoded.Units)
	require.Equal(t, int32(100), decoded.Scale)
	require.Equal(t, true, decoded.IsActive)
	require.Equal(t, float32(1.5), decoded.Rate)
	require.Equal(t, 999.99, decoded.Amount)

	// Verify compatibility with standard protobuf
	standardData, err := proto.Marshal(original)
	require.NoError(t, err)
	
	standardDecoded := &ImplicitFieldPresence{}
	err = proto.Unmarshal(standardData, standardDecoded)
	require.NoError(t, err)
	
	require.Equal(t, decoded.CurrencyCode, standardDecoded.CurrencyCode)
	require.Equal(t, decoded.Units, standardDecoded.Units)
	require.Equal(t, decoded.Scale, standardDecoded.Scale)
}

// TestImplicitFieldPresenceZeroValues tests that zero values are handled correctly
func TestImplicitFieldPresenceZeroValues(t *testing.T) {
	// Create a message with zero values
	original := &ImplicitFieldPresence{
		CurrencyCode: "",
		Units:        0,
		Scale:        0,
		IsActive:     false,
		Rate:         0.0,
		Amount:       0.0,
	}

	// Marshal using VT
	data, err := original.MarshalVT()
	require.NoError(t, err)

	// Unmarshal into a new message using VT
	decoded := &ImplicitFieldPresence{}
	err = decoded.UnmarshalVT(data)
	require.NoError(t, err)

	// Zero values should be preserved (this is IMPLICIT presence behavior)
	require.Equal(t, "", decoded.CurrencyCode)
	require.Equal(t, int64(0), decoded.Units)
	require.Equal(t, int32(0), decoded.Scale)
	require.Equal(t, false, decoded.IsActive)
	require.Equal(t, float32(0.0), decoded.Rate)
	require.Equal(t, 0.0, decoded.Amount)
}

// TestExplicitFieldPresenceUnmarshal tests that Edition 2023 fields with EXPLICIT presence
// are correctly unmarshaled as pointer assignments
func TestExplicitFieldPresenceUnmarshal(t *testing.T) {
	currencyCode := "EUR"
	units := int64(67890)
	scale := int32(200)
	isActive := true
	rate := float32(2.5)
	amount := 1234.56

	// Create a message with all fields set
	original := &ExplicitFieldPresence{
		CurrencyCode: &currencyCode,
		Units:        &units,
		Scale:        &scale,
		IsActive:     &isActive,
		Rate:         &rate,
		Amount:       &amount,
	}

	// Marshal using VT
	data, err := original.MarshalVT()
	require.NoError(t, err)

	// Unmarshal into a new message using VT
	decoded := &ExplicitFieldPresence{}
	err = decoded.UnmarshalVT(data)
	require.NoError(t, err)

	// Verify all fields were decoded correctly (pointer assignment)
	require.NotNil(t, decoded.CurrencyCode)
	require.Equal(t, "EUR", *decoded.CurrencyCode)
	require.NotNil(t, decoded.Units)
	require.Equal(t, int64(67890), *decoded.Units)
	require.NotNil(t, decoded.Scale)
	require.Equal(t, int32(200), *decoded.Scale)
	require.NotNil(t, decoded.IsActive)
	require.Equal(t, true, *decoded.IsActive)
	require.NotNil(t, decoded.Rate)
	require.Equal(t, float32(2.5), *decoded.Rate)
	require.NotNil(t, decoded.Amount)
	require.Equal(t, 1234.56, *decoded.Amount)
}

// TestExplicitFieldPresenceNil tests that nil pointer values work correctly with EXPLICIT presence
func TestExplicitFieldPresenceNil(t *testing.T) {
	// Create a message with no fields set (all nil)
	original := &ExplicitFieldPresence{}

	// Marshal using VT
	data, err := original.MarshalVT()
	require.NoError(t, err)

	// Unmarshal into a new message using VT
	decoded := &ExplicitFieldPresence{}
	err = decoded.UnmarshalVT(data)
	require.NoError(t, err)

	// All fields should be nil (this is EXPLICIT presence behavior)
	require.Nil(t, decoded.CurrencyCode)
	require.Nil(t, decoded.Units)
	require.Nil(t, decoded.Scale)
	require.Nil(t, decoded.IsActive)
	require.Nil(t, decoded.Rate)
	require.Nil(t, decoded.Amount)
}

// TestScalarTypesUnmarshal tests unmarshaling of all scalar types with default (IMPLICIT) presence
func TestScalarTypesUnmarshal(t *testing.T) {
	original := &ScalarTypes{
		DoubleField:   3.14159,
		FloatField:    2.71828,
		Int32Field:    -12345,
		Int64Field:    -9876543210,
		Uint32Field:   54321,
		Uint64Field:   9876543210,
		Sint32Field:   -99999,
		Sint64Field:   -8888888888,
		Fixed32Field:  11111,
		Fixed64Field:  22222222222,
		Sfixed32Field: -33333,
		Sfixed64Field: -4444444444,
		BoolField:     true,
		StringField:   "test string",
		BytesField:    []byte{0x01, 0x02, 0x03, 0x04},
	}

	// Marshal using VT
	data, err := original.MarshalVT()
	require.NoError(t, err)

	// Unmarshal into a new message using VT
	decoded := &ScalarTypes{}
	err = decoded.UnmarshalVT(data)
	require.NoError(t, err)

	// Verify all scalar types were decoded correctly
	require.Equal(t, original.DoubleField, decoded.DoubleField)
	require.Equal(t, original.FloatField, decoded.FloatField)
	require.Equal(t, original.Int32Field, decoded.Int32Field)
	require.Equal(t, original.Int64Field, decoded.Int64Field)
	require.Equal(t, original.Uint32Field, decoded.Uint32Field)
	require.Equal(t, original.Uint64Field, decoded.Uint64Field)
	require.Equal(t, original.Sint32Field, decoded.Sint32Field)
	require.Equal(t, original.Sint64Field, decoded.Sint64Field)
	require.Equal(t, original.Fixed32Field, decoded.Fixed32Field)
	require.Equal(t, original.Fixed64Field, decoded.Fixed64Field)
	require.Equal(t, original.Sfixed32Field, decoded.Sfixed32Field)
	require.Equal(t, original.Sfixed64Field, decoded.Sfixed64Field)
	require.Equal(t, original.BoolField, decoded.BoolField)
	require.Equal(t, original.StringField, decoded.StringField)
	require.Equal(t, original.BytesField, decoded.BytesField)
}

// TestRegularMessageUnmarshal tests unmarshaling of a regular message with nested messages
func TestRegularMessageUnmarshal(t *testing.T) {
	original := &RegularMessage{
		Id:   12345,
		Name: "test message",
		Values: []int64{1, 2, 3, 4, 5},
		Nested: &NestedMessage{
			Id:   999,
			Name: "nested",
			Data: []byte{0xAA, 0xBB, 0xCC},
		},
		Metadata: map[string]int32{
			"key1": 100,
			"key2": 200,
			"key3": 300,
		},
	}

	// Marshal using VT
	data, err := original.MarshalVT()
	require.NoError(t, err)

	// Unmarshal into a new message using VT
	decoded := &RegularMessage{}
	err = decoded.UnmarshalVT(data)
	require.NoError(t, err)

	// Verify all fields were decoded correctly
	require.Equal(t, original.Id, decoded.Id)
	require.Equal(t, original.Name, decoded.Name)
	require.Equal(t, original.Values, decoded.Values)
	require.NotNil(t, decoded.Nested)
	require.Equal(t, original.Nested.Id, decoded.Nested.Id)
	require.Equal(t, original.Nested.Name, decoded.Nested.Name)
	require.Equal(t, original.Nested.Data, decoded.Nested.Data)
	require.Equal(t, original.Metadata, decoded.Metadata)
}

// TestEditionsCompatibilityWithStandardProto ensures vtprotobuf marshaling
// is compatible with standard protobuf unmarshaling and vice versa
func TestEditionsCompatibilityWithStandardProto(t *testing.T) {
	original := &ImplicitFieldPresence{
		CurrencyCode: "GBP",
		Units:        55555,
		Scale:        50,
		IsActive:     true,
		Rate:         3.75,
		Amount:       777.77,
	}

	// Marshal with VT, unmarshal with standard proto
	vtData, err := original.MarshalVT()
	require.NoError(t, err)

	standardDecoded := &ImplicitFieldPresence{}
	err = proto.Unmarshal(vtData, standardDecoded)
	require.NoError(t, err)

	require.Equal(t, original.CurrencyCode, standardDecoded.CurrencyCode)
	require.Equal(t, original.Units, standardDecoded.Units)
	require.Equal(t, original.Scale, standardDecoded.Scale)

	// Marshal with standard proto, unmarshal with VT
	standardData, err := proto.Marshal(original)
	require.NoError(t, err)

	vtDecoded := &ImplicitFieldPresence{}
	err = vtDecoded.UnmarshalVT(standardData)
	require.NoError(t, err)

	require.Equal(t, original.CurrencyCode, vtDecoded.CurrencyCode)
	require.Equal(t, original.Units, vtDecoded.Units)
	require.Equal(t, original.Scale, vtDecoded.Scale)
}
