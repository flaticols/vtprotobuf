package proto3opt

import (
	"testing"

	"github.com/planetscale/vtprotobuf/protohelpers"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestEmptyBytesMarshalling(t *testing.T) {
	a := &OptionalFieldInProto3{
		OptionalBytes: nil,
	}
	b := &OptionalFieldInProto3{
		OptionalBytes: []byte{},
	}

	type Message interface {
		proto.Message
		MarshalVT() ([]byte, error)
	}

	for _, msg := range []Message{a, b} {
		vt, err := msg.MarshalVT()
		require.NoError(t, err)
		goog, err := proto.Marshal(msg)
		require.NoError(t, err)
		require.Equal(t, goog, vt)
	}
}

func TestInvalidUTF8Rejected(t *testing.T) {
	// Create a message with valid string
	validMsg := &OptionalFieldInProto3{
		OptionalString: proto.String("valid string"),
	}
	validBytes, err := validMsg.MarshalVT()
	require.NoError(t, err)

	// Verify valid message can be unmarshaled
	unmarshaledValid := &OptionalFieldInProto3{}
	err = unmarshaledValid.UnmarshalVT(validBytes)
	require.NoError(t, err)
	require.Equal(t, "valid string", *unmarshaledValid.OptionalString)

	// Create malformed bytes with invalid UTF-8 in the string field
	// Field 14 (optional_string) with wire type 2 (length-delimited)
	// Tag: (14 << 3) | 2 = 114
	// Invalid UTF-8 bytes: 0xff, 0xfe are not valid UTF-8
	invalidUTF8Bytes := []byte{
		114,       // tag for field 14, wire type 2
		4,         // length = 4
		0xff, 0xfe, 0xfd, 0xfc, // invalid UTF-8 sequence
	}

	// Verify invalid UTF-8 is rejected
	invalidMsg := &OptionalFieldInProto3{}
	err = invalidMsg.UnmarshalVT(invalidUTF8Bytes)
	require.Error(t, err)
	require.ErrorIs(t, err, protohelpers.ErrInvalidUTF8)
}
