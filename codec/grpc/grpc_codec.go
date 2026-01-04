package grpc

import "fmt"

// Name is the name registered for the proto compressor.
const Name = "proto"

type Codec struct{}

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

func (Codec) Marshal(v interface{}) ([]byte, error) {
	vt, ok := v.(vtprotoMessage)
	if !ok {
		return nil, fmt.Errorf("failed to marshal, message is %T (missing vtprotobuf helpers)", v)
	}
	return vt.MarshalVT()
}

func (Codec) Unmarshal(data []byte, v interface{}) error {
	vt, ok := v.(vtprotoMessage)
	if !ok {
		return fmt.Errorf("failed to unmarshal, message is %T (missing vtprotobuf helpers)", v)
	}
	// Reset the message before unmarshaling to match the semantics of the
	// default protobuf codec, which replaces rather than merges messages.
	if r, ok := v.(vtprotoResetter); ok {
		r.ResetVT()
	} else if r, ok := v.(protoResetter); ok {
		r.Reset()
	}
	return vt.UnmarshalVT(data)
}

func (Codec) Name() string {
	return Name
}
