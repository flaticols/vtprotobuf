# Investigation: JSON Support for vtprotobuf

## Executive Summary

This document investigates the possibilities for adding JSON marshaling/unmarshaling support to vtprotobuf, including considerations for streaming JSON. The investigation covers technical approaches, implementation complexity, performance implications, and compatibility with the existing vtprotobuf architecture.

## Background

**vtprotobuf** is a protoc plugin that generates optimized marshal/unmarshal code for Protocol Buffers. It provides significant performance improvements over the standard `proto.Marshal`/`proto.Unmarshal` by generating unrolled, reflection-free code.

### Current Features
- **marshal**: Binary protobuf marshaling (MarshalVT, MarshalToVT, etc.)
- **unmarshal**: Binary protobuf unmarshaling (UnmarshalVT)
- **size**: Efficient size calculation (SizeVT)
- **equal**: Optimized equality checks (EqualVT)
- **clone**: Fast cloning without reflection (CloneVT)
- **pool**: Memory pooling for reduced allocations
- **grpc**: Optimized gRPC service generation

### Issue Requirements
1. **JSON Marshaling**: Ability to marshal proto messages to JSON
2. **Streaming Support**: Support for streaming JSON operations

## Technical Analysis

### 1. JSON Marshaling in Protocol Buffers

#### Standard Library Support
Go's `google.golang.org/protobuf/encoding/protojson` package provides comprehensive JSON support:

```go
import "google.golang.org/protobuf/encoding/protojson"

// Marshal to JSON
data, err := protojson.Marshal(msg)

// Unmarshal from JSON
err := protojson.Unmarshal(data, msg)
```

**Key Features:**
- **MarshalOptions**: Configurable marshaling behavior
  - `Multiline`: Pretty-printed JSON output
  - `Indent`: Custom indentation
  - `UseProtoNames`: Use proto field names instead of camelCase
  - `UseEnumNumbers`: Emit enum values as numbers
  - `EmitUnpopulated`: Include zero-value fields
  - `EmitDefaultValues`: Include default values
  - `AllowPartial`: Allow messages with missing required fields

- **UnmarshalOptions**: Configurable unmarshaling behavior
  - `AllowPartial`: Accept messages with missing required fields
  - `DiscardUnknown`: Ignore unknown fields
  - `RecursionLimit`: Prevent stack overflow
  - `Resolver`: Custom type resolution for Any/extensions

#### JSON Mapping Rules
The protobuf JSON specification (https://protobuf.dev/programming-guides/proto3#json) defines:
- Field names: lowerCamelCase by default (or proto names with option)
- Enums: String names or numbers
- bytes: Base64 encoding
- Timestamps: RFC 3339 format
- Duration: String with "s" suffix
- Well-known types: Special JSON representations

### 2. Implementation Approaches

#### Option A: Generate Optimized JSON Methods (Full Implementation)

**Description**: Generate `MarshalJSONVT()` and `UnmarshalJSONVT()` methods similar to existing binary marshal methods.

**Pros:**
- Maximum performance potential through unrolled code generation
- Consistency with vtprotobuf's philosophy
- Full control over serialization behavior
- Potential for zero-allocation implementations
- Could skip reflection entirely

**Cons:**
- **Very high implementation complexity**
  - JSON spec has complex edge cases
  - Must handle all well-known types correctly
  - Need to implement JSON escaping, encoding
  - Must support all MarshalOptions/UnmarshalOptions
- **Large code generation overhead**
  - Would significantly increase generated file sizes
  - More complexity = more potential bugs
- **Maintenance burden**
  - Must stay in sync with protobuf JSON spec
  - Any spec changes require generator updates
- **Testing complexity**
  - Need comprehensive test coverage for all edge cases
  - Must validate against standard protojson behavior

**Estimated Complexity**: Very High (likely 2-3x the complexity of binary marshal)

#### Option B: Thin Wrapper Around protojson (Hybrid Approach)

**Description**: Generate helper methods that use the standard `protojson` package but optimize specific operations.

**Example Implementation:**
```go
func (m *Message) MarshalJSONVT() ([]byte, error) {
    return protojson.Marshal(m)
}

func (m *Message) MarshalJSONVTWithOptions(opts protojson.MarshalOptions) ([]byte, error) {
    return opts.Marshal(m)
}

func (m *Message) UnmarshalJSONVT(data []byte) error {
    return protojson.Unmarshal(data, m)
}
```

**Pros:**
- **Low implementation complexity**
- Maintains compatibility with protobuf JSON spec automatically
- Updates to protojson are automatically inherited
- Smaller generated code size
- Easier to test and maintain
- Can add convenience methods (e.g., predefined MarshalOptions)

**Cons:**
- Limited performance improvement over direct protojson usage
- Still uses reflection internally (in protojson)
- Doesn't align with vtprotobuf's zero-reflection philosophy

**Estimated Complexity**: Low to Medium

#### Option C: Partial Optimization with protojson Fallback

**Description**: Generate optimized JSON code for simple types, fall back to protojson for complex types.

**Strategy:**
- Optimize simple scalar fields (int, string, bool)
- Use protojson for:
  - Well-known types (Timestamp, Duration, Any, etc.)
  - Nested messages
  - Maps with complex values
  - Oneof fields

**Pros:**
- Balance between performance and complexity
- Automatic spec compliance for complex cases
- Smaller generated code than full implementation
- Performance gains for common cases

**Cons:**
- Still moderate implementation complexity
- Code generation for JSON is non-trivial
- Must handle JSON escaping correctly
- Testing complexity for partial implementation
- Inconsistent performance characteristics

**Estimated Complexity**: High

### 3. Streaming JSON Support

Streaming JSON is more complex because JSON isn't inherently a streaming format (unlike Protocol Buffers).

#### Streaming Approaches:

##### A. Newline-Delimited JSON (NDJSON)
```
{"id":1,"name":"Alice"}
{"id":2,"name":"Bob"}
{"id":3,"name":"Charlie"}
```

**Pros:**
- Simple to implement
- Each line is a complete JSON object
- Easy to parse line-by-line
- Widely used (log files, data pipelines)

**Cons:**
- Not part of JSON spec
- No nested structure
- Each message must be self-contained

##### B. JSON Array Streaming
```json
[
  {"id":1,"name":"Alice"},
  {"id":2,"name":"Bob"},
  {"id":3,"name":"Charlie"}
]
```

**Challenges:**
- Need to handle opening `[` and closing `]`
- Comma placement between objects
- Harder to parse incrementally
- Not true streaming (need to buffer)

##### C. JSON Lines Format (Similar to NDJSON)
Uses line-delimited JSON with each line being a valid JSON value.

##### D. Custom Streaming Protocol
Could define a custom protocol like:
- Length-prefixed JSON
- Delimiter-separated JSON
- Protocol Buffers wrapper around JSON (defeats the purpose)

#### Recommended Streaming Approach

**Newline-Delimited JSON (NDJSON)** is the most practical:

```go
// Streaming writer
type JSONStreamWriter struct {
    w io.Writer
}

func (w *JSONStreamWriter) WriteMessage(msg proto.Message) error {
    data, err := protojson.Marshal(msg)
    if err != nil {
        return err
    }
    _, err = w.w.Write(data)
    if err != nil {
        return err
    }
    _, err = w.w.Write([]byte("\n"))
    return err
}

// Streaming reader
type JSONStreamReader struct {
    scanner *bufio.Scanner
}

func (r *JSONStreamReader) ReadMessage(msg proto.Message) error {
    if !r.scanner.Scan() {
        return io.EOF
    }
    return protojson.Unmarshal(r.scanner.Bytes(), msg)
}
```

### 4. Performance Considerations

#### Binary Protobuf vs JSON
- **Binary**: Compact, fast, type-safe, harder to debug
- **JSON**: Human-readable, larger size, slower parsing, easier debugging

#### vtprotobuf Binary Performance Gains
Current vtprotobuf achieves 2-5x performance improvements for binary operations through:
- Eliminated reflection
- Unrolled loops
- Inline code generation
- Optimized buffer management
- Memory pooling

#### Expected JSON Performance Gains

**Option A (Full Generation):**
- Potential: 2-4x improvement over protojson
- Reality: Likely less due to:
  - JSON parsing/generation inherent complexity
  - String escaping overhead
  - Base64 encoding for bytes
  - Floating-point formatting

**Option B (Wrapper):**
- Minimal or no performance improvement
- Main benefit: API consistency

**Option C (Partial Optimization):**
- Potential: 1.5-2x improvement for simple messages
- Highly dependent on message structure

### 5. Implementation Recommendations

#### Primary Recommendation: Option B (Wrapper Approach)

**Rationale:**
1. **Complexity vs. Benefit**: JSON serialization is fundamentally different from binary protobuf. The complexity of a full implementation (Option A) likely doesn't justify the performance gains, especially given:
   - JSON is already human-readable and debugging-friendly
   - Network/IO overhead often dominates JSON use cases
   - Most users choosing JSON prioritize interoperability over performance

2. **Maintenance**: Using the standard `protojson` package ensures:
   - Automatic compliance with protobuf JSON spec
   - Bug fixes and improvements from the protobuf team
   - No need to track spec changes

3. **API Consistency**: Adding wrapper methods still provides value:
   - Consistent API across vtprotobuf features
   - Convenient access to JSON marshaling
   - Can add helper methods for common configurations

4. **Future Flexibility**: Can always optimize later if needed:
   - Start with wrappers
   - Profile to identify bottlenecks
   - Optimize specific operations if proven necessary

#### Implementation Plan (If Proceeding with Option B)

**Phase 1: Basic JSON Feature**
1. Create `features/json` package
2. Register "json" feature in the generator
3. Generate methods:
   ```go
   func (m *Message) MarshalJSON() ([]byte, error)
   func (m *Message) UnmarshalJSON(data []byte) error
   func (m *Message) MarshalJSONVT() ([]byte, error)
   func (m *Message) UnmarshalJSONVT(data []byte) error
   ```

4. Implement with protojson wrappers
5. Add tests comparing output with protojson

**Phase 2: Configuration Options**
1. Add common MarshalOptions presets:
   ```go
   func (m *Message) MarshalJSONVTIndent() ([]byte, error)
   func (m *Message) MarshalJSONVTProtoNames() ([]byte, error)
   ```

2. Add full options support:
   ```go
   func (m *Message) MarshalJSONVTWithOptions(opts protojson.MarshalOptions) ([]byte, error)
   ```

**Phase 3: Streaming Support**
1. Create streaming utilities in a separate package
2. Implement NDJSON streaming:
   ```go
   package vtjsonstream
   
   type Writer struct { /* ... */ }
   func (w *Writer) WriteMessage(msg proto.Message) error
   
   type Reader struct { /* ... */ }
   func (r *Reader) ReadMessage(msg proto.Message) error
   ```

3. Add examples and documentation

#### Alternative: Option C for Future Optimization

If performance profiling later shows JSON is a bottleneck, consider Option C:
1. Start with Option B implementation
2. Profile real-world usage
3. Identify hot paths (likely: string fields, simple scalars)
4. Generate optimized code for those specific cases
5. Fall back to protojson for complex types

### 6. Integration with Existing Architecture

#### Feature Registration
Follow existing pattern in `features/json/json.go`:
```go
func init() {
    generator.RegisterFeature("json", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
        return &jsonFeature{GeneratedFile: gen}
    })
}
```

#### Command-Line Usage
```bash
protoc \
  --go_out=. \
  --go-vtproto_out=. \
  --go-vtproto_opt=features=marshal+unmarshal+size+json \
  message.proto
```

#### Generated Code Structure
```go
// message_vtproto.pb.go

import (
    "google.golang.org/protobuf/encoding/protojson"
)

// Existing binary methods
func (m *Message) MarshalVT() ([]byte, error) { /* ... */ }
func (m *Message) UnmarshalVT(data []byte) error { /* ... */ }

// New JSON methods
func (m *Message) MarshalJSON() ([]byte, error) {
    return protojson.Marshal(m)
}

func (m *Message) UnmarshalJSON(data []byte) error {
    return protojson.Unmarshal(data, m)
}

// Additional convenience methods
func (m *Message) MarshalJSONIndent() ([]byte, error) {
    return protojson.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(m)
}
```

### 7. Comparison with Other Ecosystems

#### gogoproto/jsonpb (deprecated)
- Had custom JSON marshaling
- Now deprecated in favor of protojson
- Performance was similar to current protojson

#### Other Languages
- **C++**: protobuf provides JsonUtil::MessageToJson
- **Java**: JsonFormat in protobuf-java
- **Python**: json_format in protobuf
- All use library-provided JSON implementations

**Pattern**: Most ecosystems don't generate JSON serialization code; they use library implementations.

### 8. Testing Strategy

#### Unit Tests
1. Test basic marshaling/unmarshaling
2. Test with all field types (scalars, repeated, maps, nested)
3. Test well-known types
4. Test edge cases (empty messages, nil fields)
5. Compare output with protojson

#### Integration Tests
1. Test streaming scenarios
2. Test with real-world message structures
3. Test interoperability with other protobuf implementations

#### Benchmarks
1. Compare performance with direct protojson usage
2. Benchmark streaming scenarios
3. Measure generated code size impact

### 9. Documentation Requirements

#### README Updates
1. Add "json" to feature list
2. Explain JSON marshaling support
3. Provide usage examples
4. Document streaming utilities

#### Code Examples
```go
// Basic usage
msg := &MyMessage{Id: 1, Name: "test"}

// JSON marshaling
jsonData, err := msg.MarshalJSON()

// JSON unmarshaling
newMsg := &MyMessage{}
err = newMsg.UnmarshalJSON(jsonData)

// Streaming
writer := vtjsonstream.NewWriter(conn)
for _, msg := range messages {
    writer.WriteMessage(msg)
}
```

### 10. Open Questions and Considerations

#### Questions to Address:
1. **Should MarshalJSON/UnmarshalJSON be generated?**
   - These are standard Go json.Marshaler/json.Unmarshaler interfaces
   - Generating them enables `json.Marshal(msg)` to work automatically
   - Consider: This might interfere with users who want custom JSON behavior

2. **Should streaming be part of the feature or a separate package?**
   - Recommendation: Separate package for flexibility
   - Avoids code generation for streaming utilities

3. **What about JSON Schema generation?**
   - Out of scope for initial implementation
   - Could be future enhancement

4. **Performance SLAs?**
   - Should we commit to any performance guarantees?
   - Recommendation: Start with wrapper, measure, optimize if needed

#### Risks:
1. **User Expectations**: Users might expect vtprotobuf-level performance gains
   - Mitigation: Clear documentation about performance characteristics
   
2. **JSON Spec Compliance**: Must ensure full compliance with protobuf JSON spec
   - Mitigation: Use protojson library, comprehensive testing

3. **Code Size**: Even wrapper methods increase generated code
   - Mitigation: Make feature optional (already the case)

### 11. Resource Estimates

#### Option B Implementation (Wrapper Approach)

**Development Time Estimate:**
- Feature implementation: 1-2 days
- Streaming utilities: 1 day  
- Testing: 2-3 days
- Documentation: 1 day
- **Total: ~5-7 days**

**Maintenance Overhead:**
- Low: Primarily depends on protojson updates
- Test updates as protobuf evolves

#### Option A Implementation (Full Generation)

**Development Time Estimate:**
- Core JSON generation: 1-2 weeks
- Well-known types: 3-5 days
- Edge cases and spec compliance: 1 week
- Testing: 1-2 weeks
- Documentation: 2-3 days
- **Total: ~4-6 weeks**

**Maintenance Overhead:**
- High: Must track protobuf JSON spec changes
- Significant testing requirements

### 12. Conclusion and Recommendation

#### Recommendation: Implement Option B (Wrapper Approach)

**Rationale:**
1. **Best cost/benefit ratio**: Low complexity, good user experience
2. **Maintainable**: Leverages well-tested protojson library
3. **Practical**: JSON users typically prioritize simplicity over maximum performance
4. **Extensible**: Can optimize later if profiling shows bottlenecks
5. **Consistent**: Follows patterns from other language implementations

#### Feature Specification

**Generated Methods:**
```go
// Standard json.Marshaler interface
func (m *Message) MarshalJSON() ([]byte, error)

// Standard json.Unmarshaler interface  
func (m *Message) UnmarshalJSON(data []byte) error

// With options support
func (m *Message) MarshalJSONVT(opts ...protojson.MarshalOptions) ([]byte, error)
func (m *Message) UnmarshalJSONVT(data []byte, opts ...protojson.UnmarshalOptions) error
```

**Streaming Package:**
```go
package vtjsonstream

// NDJSON streaming support
type Writer struct { /* io.Writer wrapper */ }
func NewWriter(w io.Writer) *Writer
func (w *Writer) WriteMessage(msg proto.Message) error

type Reader struct { /* bufio.Scanner wrapper */ }
func NewReader(r io.Reader) *Reader  
func (r *Reader) ReadMessage(msg proto.Message) error
func (r *Reader) HasNext() bool
```

#### Next Steps (If Approved)

1. Create GitHub issue for tracking
2. Implement `features/json` package following Option B
3. Add streaming utilities in separate package
4. Add comprehensive tests
5. Update documentation
6. Create benchmarks comparing with direct protojson usage
7. Solicit community feedback

#### Long-term Optimization Path

If performance becomes critical:
1. Profile real-world usage patterns
2. Identify specific bottlenecks
3. Consider Option C: selective optimization
4. Generate optimized code for proven hot paths
5. Maintain protojson fallback for correctness

---

## Appendix: Code Structure

### Proposed File Structure
```
features/
  json/
    json.go          # Feature implementation
    json_test.go     # Unit tests
    
vtjsonstream/         # Separate package for streaming
  stream.go           # NDJSON streaming implementation
  stream_test.go      # Streaming tests
  examples_test.go    # Usage examples
  
testproto/
  json/               # Test proto files for JSON feature
    test.proto
    test_vtproto.pb.go
    json_test.go
```

### Feature Implementation Skeleton

```go
// features/json/json.go
package json

import (
    "github.com/planetscale/vtprotobuf/generator"
    "google.golang.org/protobuf/compiler/protogen"
    "google.golang.org/protobuf/encoding/protojson"
)

func init() {
    generator.RegisterFeature("json", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
        return &jsonFeature{GeneratedFile: gen}
    })
}

type jsonFeature struct {
    *generator.GeneratedFile
    once bool
}

func (f *jsonFeature) GenerateFile(file *protogen.File) bool {
    for _, message := range file.Messages {
        f.generateMessage(message)
    }
    return f.once
}

func (f *jsonFeature) generateMessage(message *protogen.Message) {
    // Skip well-known types, opaque API messages, etc.
    if f.IsWellKnownType(message) || f.IsOpaqueMessage(message) {
        return
    }
    
    f.once = true
    
    // Generate MarshalJSON
    f.P(`func (m *`, message.GoIdent.GoName, `) MarshalJSON() ([]byte, error) {`)
    f.P(`return `, f.Ident("protojson", "Marshal"), `(m)`)
    f.P(`}`)
    f.P()
    
    // Generate UnmarshalJSON
    f.P(`func (m *`, message.GoIdent.GoName, `) UnmarshalJSON(data []byte) error {`)
    f.P(`return `, f.Ident("protojson", "Unmarshal"), `(data, m)`)
    f.P(`}`)
    f.P()
}
```

---

## References

1. Protocol Buffers JSON Mapping: https://protobuf.dev/programming-guides/proto3#json
2. Go protojson package: https://pkg.go.dev/google.golang.org/protobuf/encoding/protojson
3. NDJSON specification: http://ndjson.org/
4. JSON Lines format: https://jsonlines.org/
5. vtprotobuf architecture: Current codebase analysis

---

**Document Version**: 1.0  
**Date**: January 7, 2026  
**Author**: Investigation for vtprotobuf JSON support feature request
