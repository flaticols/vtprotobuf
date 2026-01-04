package grpc

import (
	"testing"

	"github.com/planetscale/vtprotobuf/testproto/pool"
)

func TestCodecUnmarshalResetsMessage(t *testing.T) {
	codec := Codec{}

	// Create and marshal a message with values
	msg1 := &pool.MemoryPoolExtension{
		Foo1: "hello",
		Foo2: 42,
	}
	data, err := codec.Marshal(msg1)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Create another message with different values and marshal it
	msg2 := &pool.MemoryPoolExtension{
		Foo1: "world",
		Foo2: 100,
	}
	data2, err := codec.Marshal(msg2)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Now unmarshal msg2 data into msg1 - this simulates reusing a message in a stream
	// The old behavior would merge, the new behavior should replace
	if err := codec.Unmarshal(data2, msg1); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// msg1 should now have msg2's values, not a merge
	if msg1.Foo1 != "world" {
		t.Errorf("Foo1 = %q, want %q", msg1.Foo1, "world")
	}
	if msg1.Foo2 != 100 {
		t.Errorf("Foo2 = %d, want %d", msg1.Foo2, 100)
	}

	// Unmarshal empty message - all fields should be reset
	empty := &pool.MemoryPoolExtension{}
	emptyData, err := codec.Marshal(empty)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if err := codec.Unmarshal(emptyData, msg1); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// msg1 should now be empty, not retain old values
	if msg1.Foo1 != "" {
		t.Errorf("Foo1 = %q, want empty string", msg1.Foo1)
	}
	if msg1.Foo2 != 0 {
		t.Errorf("Foo2 = %d, want 0", msg1.Foo2)
	}

	// Verify normal unmarshal still works
	target := &pool.MemoryPoolExtension{}
	if err := codec.Unmarshal(data, target); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if target.Foo1 != "hello" {
		t.Errorf("Foo1 = %q, want %q", target.Foo1, "hello")
	}
	if target.Foo2 != 42 {
		t.Errorf("Foo2 = %d, want %d", target.Foo2, 42)
	}
}
