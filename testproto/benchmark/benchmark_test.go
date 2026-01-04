package benchmark

import (
	"testing"

	"github.com/planetscale/vtprotobuf/testproto/pool"
	"google.golang.org/protobuf/proto"
)

// Run with: go test -bench=. -benchmem ./testproto/benchmark/...

// Using MemoryPoolExtension from testproto/pool for benchmarks

func BenchmarkMarshal_Proto(b *testing.B) {
	msg := createTestMessage()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = proto.Marshal(msg)
	}
}

func BenchmarkMarshal_VT(b *testing.B) {
	msg := createTestMessage()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = msg.MarshalVT()
	}
}

func BenchmarkUnmarshal_Proto(b *testing.B) {
	msg := createTestMessage()
	data, _ := proto.Marshal(msg)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := &pool.MemoryPoolExtension{}
		_ = proto.Unmarshal(data, m)
	}
}

func BenchmarkUnmarshal_VT(b *testing.B) {
	msg := createTestMessage()
	data, _ := msg.MarshalVT()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		m := &pool.MemoryPoolExtension{}
		_ = m.UnmarshalVT(data)
	}
}

func BenchmarkSize_Proto(b *testing.B) {
	msg := createTestMessage()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = proto.Size(msg)
	}
}

func BenchmarkSize_VT(b *testing.B) {
	msg := createTestMessage()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = msg.SizeVT()
	}
}

func BenchmarkClone_Proto(b *testing.B) {
	msg := createTestMessage()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = proto.Clone(msg)
	}
}

func BenchmarkClone_VT(b *testing.B) {
	msg := createTestMessage()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = msg.CloneVT()
	}
}

func createTestMessage() *pool.MemoryPoolExtension {
	return &pool.MemoryPoolExtension{
		Foo1: "benchmark test message with some content for testing vtprotobuf performance",
		Foo2: 12345678901234,
		Foo3: &pool.OptionalMessage{},
	}
}
