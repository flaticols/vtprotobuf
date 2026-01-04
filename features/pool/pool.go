package pool

import (
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/planetscale/vtprotobuf/generator"
)

func init() {
	generator.RegisterFeature("pool", func(gen *generator.GeneratedFile) generator.FeatureGenerator {
		return &pool{GeneratedFile: gen}
	})
}

type pool struct {
	*generator.GeneratedFile
	once bool
}

var _ generator.FeatureGenerator = (*pool)(nil)

func (p *pool) GenerateFile(file *protogen.File) bool {
	for _, message := range file.Messages {
		p.message(message)
	}
	return p.once
}

func (p *pool) message(message *protogen.Message) {
	for _, nested := range message.Messages {
		p.message(nested)
	}

	if message.Desc.IsMapEntry() || !p.ShouldPool(message) {
		return
	}

	// Skip opaque API messages - fields are private and cannot be accessed directly.
	if p.IsOpaque(message) {
		return
	}

	p.once = true
	ccTypeName := message.GoIdent

	p.P(`var vtprotoPool_`, ccTypeName, ` = `, p.Ident("sync", "Pool"), `{`)
	p.P(`New: func() interface{} {`)
	p.P(`return &`, ccTypeName, `{}`)
	p.P(`},`)
	p.P(`}`)

	p.P(`func (m *`, ccTypeName, `) ResetVT() {`)
	p.P(`if m != nil {`)
	var saved []*protogen.Field
	var oneofBytes []*protogen.Field // Track oneof bytes fields

	for _, field := range message.Fields {
		fieldName := field.GoName

		if field.Desc.IsList() {
			switch field.Desc.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				p.P(`for _, mm := range m.`, fieldName, `{`)
				if p.ShouldPool(field.Message) {
					p.P(`mm.ResetVT()`)
				} else {
					p.P(`mm.Reset()`)
				}
				p.P(`}`)
			case protoreflect.BytesKind, protoreflect.StringKind:
				p.P(`clear(m.`, fieldName, `)`)
			}
			p.P(fmt.Sprintf("f%d", len(saved)), ` := m.`, fieldName, `[:0]`)
			saved = append(saved, field)
		} else if field.Oneof != nil && !field.Oneof.Desc.IsSynthetic() {
			switch field.Desc.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				if p.ShouldPool(field.Message) {
					p.P(`if oneof, ok := m.`, field.Oneof.GoName, `.(*`, field.GoIdent, `); ok {`)
					p.P(`oneof.`, fieldName, `.ReturnToVTPool()`)
					p.P(`}`)
				}
			case protoreflect.BytesKind:
				oneofBytes = append(oneofBytes, field)
			}
		} else {
			switch field.Desc.Kind() {
			case protoreflect.MessageKind, protoreflect.GroupKind:
				if !field.Desc.IsMap() && p.ShouldPool(field.Message) {
					p.P(`m.`, fieldName, `.ReturnToVTPool()`)
				}
			case protoreflect.BytesKind:
				p.P(fmt.Sprintf("f%d", len(saved)), ` := m.`, fieldName, `[:0]`)
				saved = append(saved, field)
			}
		}
	}

	// Handle oneof bytes fields - save the oneof wrapper with preserved capacity
	if len(oneofBytes) > 0 {
		// Group by oneof
		oneofGroups := make(map[string][]*protogen.Field)
		for _, field := range oneofBytes {
			oneofGroups[field.Oneof.GoName] = append(oneofGroups[field.Oneof.GoName], field)
		}

		for oneofName, fields := range oneofGroups {
			p.P(`var saved`, oneofName, ` is`, ccTypeName, `_`, oneofName)
			p.P(`switch c := m.`, oneofName, `.(type) {`)
			for _, field := range fields {
				p.P(`case *`, field.GoIdent, `:`)
				p.P(`c.`, field.GoName, ` = c.`, field.GoName, `[:0]`)
				p.P(`saved`, oneofName, ` = c`)
			}
			p.P(`}`)
		}
	}

	p.P(`m.Reset()`)
	for i, field := range saved {
		p.P(`m.`, field.GoName, ` = `, fmt.Sprintf("f%d", i))
	}

	// Restore oneof bytes fields
	if len(oneofBytes) > 0 {
		oneofGroups := make(map[string][]*protogen.Field)
		for _, field := range oneofBytes {
			oneofGroups[field.Oneof.GoName] = append(oneofGroups[field.Oneof.GoName], field)
		}
		for oneofName := range oneofGroups {
			p.P(`m.`, oneofName, ` = saved`, oneofName)
		}
	}
	p.P(`}`)
	p.P(`}`)

	p.P(`func (m *`, ccTypeName, `) ReturnToVTPool() {`)
	p.P(`if m != nil {`)
	p.P(`m.ResetVT()`)
	p.P(`vtprotoPool_`, ccTypeName, `.Put(m)`)
	p.P(`}`)
	p.P(`}`)

	p.P(`func `, ccTypeName, `FromVTPool() *`, ccTypeName, `{`)
	p.P(`return vtprotoPool_`, ccTypeName, `.Get().(*`, ccTypeName, `)`)
	p.P(`}`)
}
