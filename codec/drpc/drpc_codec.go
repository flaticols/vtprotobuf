package drpc

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type vtprotoMessage interface {
	MarshalVT() ([]byte, error)
	UnmarshalVT([]byte) error
}

type vtprotoResetter interface {
	ResetVT()
}

type protoResetter interface {
	Reset()
}

func Marshal(msg interface{}) ([]byte, error) {
	return msg.(vtprotoMessage).MarshalVT()
}

func Unmarshal(buf []byte, msg interface{}) error {
	// Reset the message before unmarshaling to match the semantics of the
	// default protobuf codec, which replaces rather than merges messages.
	if r, ok := msg.(vtprotoResetter); ok {
		r.ResetVT()
	} else if r, ok := msg.(protoResetter); ok {
		r.Reset()
	}
	return msg.(vtprotoMessage).UnmarshalVT(buf)
}

func JSONMarshal(msg interface{}) ([]byte, error) {
	return protojson.Marshal(msg.(proto.Message))
}

func JSONUnmarshal(buf []byte, msg interface{}) error {
	// Reset the message before unmarshaling to match the semantics of the
	// default protobuf codec, which replaces rather than merges messages.
	if r, ok := msg.(protoResetter); ok {
		r.Reset()
	}
	return protojson.Unmarshal(buf, msg.(proto.Message))
}
