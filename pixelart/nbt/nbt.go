// Package nbt implements a minimal big-endian NBT (Named Binary Tag) reader
// and writer, enough to build and inspect Minecraft Java Edition save files
// (level.dat and Anvil region chunks).
package nbt

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

// Type is an NBT tag type identifier.
type Type byte

const (
	TagEnd Type = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
	TagIntArray
	TagLongArray
)

// Value is any NBT payload value (Byte, Short, Int, Long, Float32, Float64,
// ByteArray, String, *List, *Compound, IntArray or LongArray).
type Value interface {
	Type() Type
}

type Byte int8

func (Byte) Type() Type { return TagByte }

type Short int16

func (Short) Type() Type { return TagShort }

type Int int32

func (Int) Type() Type { return TagInt }

type Long int64

func (Long) Type() Type { return TagLong }

type Float32 float32

func (Float32) Type() Type { return TagFloat }

type Float64 float64

func (Float64) Type() Type { return TagDouble }

type ByteArray []int8

func (ByteArray) Type() Type { return TagByteArray }

type String string

func (String) Type() Type { return TagString }

// List is an ordered, homogeneously-typed NBT list.
type List struct {
	Elem  Type
	Items []Value
}

func (*List) Type() Type { return TagList }

// Compound is an ordered set of named NBT values.
type Compound struct {
	names  []string
	values []Value
}

func (*Compound) Type() Type { return TagCompound }

// NewCompound creates an empty, ordered NBT compound.
func NewCompound() *Compound {
	return &Compound{}
}

// Put appends or overwrites a named value, preserving insertion order.
func (c *Compound) Put(name string, v Value) *Compound {
	for i, n := range c.names {
		if n == name {
			c.values[i] = v
			return c
		}
	}
	c.names = append(c.names, name)
	c.values = append(c.values, v)
	return c
}

// Len reports the number of entries in the compound.
func (c *Compound) Len() int { return len(c.names) }

// At returns the name/value pair at index i.
func (c *Compound) At(i int) (string, Value) { return c.names[i], c.values[i] }

// Get looks up a value by name.
func (c *Compound) Get(name string) (Value, bool) {
	for i, n := range c.names {
		if n == name {
			return c.values[i], true
		}
	}
	return nil, false
}

type IntArray []int32

func (IntArray) Type() Type { return TagIntArray }

type LongArray []int64

func (LongArray) Type() Type { return TagLongArray }

// Encode writes a complete NBT document (named root compound) to w.
func Encode(w io.Writer, rootName string, root *Compound) error {
	bw := &byteWriter{w: w}
	bw.writeByte(byte(TagCompound))
	bw.writeString(rootName)
	writePayload(bw, root)
	return bw.err
}

type byteWriter struct {
	w   io.Writer
	err error
}

func (bw *byteWriter) write(p []byte) {
	if bw.err != nil {
		return
	}
	_, bw.err = bw.w.Write(p)
}

func (bw *byteWriter) writeByte(b byte) { bw.write([]byte{b}) }

func (bw *byteWriter) writeUint16(v uint16) {
	bw.write([]byte{byte(v >> 8), byte(v)})
}

func (bw *byteWriter) writeInt32(v int32) {
	u := uint32(v)
	bw.write([]byte{byte(u >> 24), byte(u >> 16), byte(u >> 8), byte(u)})
}

func (bw *byteWriter) writeInt64(v int64) {
	u := uint64(v)
	bw.write([]byte{
		byte(u >> 56), byte(u >> 48), byte(u >> 40), byte(u >> 32),
		byte(u >> 24), byte(u >> 16), byte(u >> 8), byte(u),
	})
}

func (bw *byteWriter) writeString(s string) {
	bw.writeUint16(uint16(len(s)))
	bw.write([]byte(s))
}

func writeNamed(bw *byteWriter, name string, v Value) {
	bw.writeByte(byte(v.Type()))
	bw.writeString(name)
	writePayload(bw, v)
}

func writePayload(bw *byteWriter, v Value) {
	switch t := v.(type) {
	case Byte:
		bw.writeByte(byte(t))
	case Short:
		bw.writeUint16(uint16(t))
	case Int:
		bw.writeInt32(int32(t))
	case Long:
		bw.writeInt64(int64(t))
	case Float32:
		bw.writeInt32(int32(math.Float32bits(float32(t))))
	case Float64:
		bw.writeInt64(int64(math.Float64bits(float64(t))))
	case ByteArray:
		bw.writeInt32(int32(len(t)))
		for _, b := range t {
			bw.writeByte(byte(b))
		}
	case String:
		bw.writeString(string(t))
	case *List:
		elem := t.Elem
		if elem == TagEnd && len(t.Items) > 0 {
			elem = t.Items[0].Type()
		}
		bw.writeByte(byte(elem))
		bw.writeInt32(int32(len(t.Items)))
		for _, item := range t.Items {
			writePayload(bw, item)
		}
	case *Compound:
		for i := 0; i < t.Len(); i++ {
			name, val := t.At(i)
			writeNamed(bw, name, val)
		}
		bw.writeByte(byte(TagEnd))
	case IntArray:
		bw.writeInt32(int32(len(t)))
		for _, i := range t {
			bw.writeInt32(i)
		}
	case LongArray:
		bw.writeInt32(int32(len(t)))
		for _, l := range t {
			bw.writeInt64(l)
		}
	default:
		panic(fmt.Sprintf("nbt: unsupported value type %T", v))
	}
}

// Decode reads a complete NBT document, returning the root name and compound.
func Decode(r io.Reader) (string, *Compound, error) {
	br := &byteReader{r: r}
	tagType := br.readByte()
	if Type(tagType) != TagCompound {
		return "", nil, fmt.Errorf("nbt: expected root compound, got type %d", tagType)
	}
	name := br.readString()
	c := readCompoundPayload(br)
	if br.err != nil {
		return "", nil, br.err
	}
	return name, c, nil
}

type byteReader struct {
	r   io.Reader
	err error
}

func (br *byteReader) read(n int) []byte {
	if br.err != nil {
		return make([]byte, n)
	}
	buf := make([]byte, n)
	_, br.err = io.ReadFull(br.r, buf)
	return buf
}

func (br *byteReader) readByte() byte { return br.read(1)[0] }

func (br *byteReader) readUint16() uint16 {
	b := br.read(2)
	return uint16(b[0])<<8 | uint16(b[1])
}

func (br *byteReader) readInt32() int32 {
	b := br.read(4)
	return int32(uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]))
}

func (br *byteReader) readInt64() int64 {
	b := br.read(8)
	var u uint64
	for _, x := range b {
		u = u<<8 | uint64(x)
	}
	return int64(u)
}

func (br *byteReader) readString() string {
	n := br.readUint16()
	return string(br.read(int(n)))
}

func readPayload(br *byteReader, t Type) Value {
	switch t {
	case TagByte:
		return Byte(int8(br.readByte()))
	case TagShort:
		return Short(int16(br.readUint16()))
	case TagInt:
		return Int(br.readInt32())
	case TagLong:
		return Long(br.readInt64())
	case TagFloat:
		return Float32(math.Float32frombits(uint32(br.readInt32())))
	case TagDouble:
		return Float64(math.Float64frombits(uint64(br.readInt64())))
	case TagByteArray:
		n := br.readInt32()
		out := make(ByteArray, n)
		for i := range out {
			out[i] = int8(br.readByte())
		}
		return out
	case TagString:
		return String(br.readString())
	case TagList:
		elem := Type(br.readByte())
		n := br.readInt32()
		l := &List{Elem: elem, Items: make([]Value, 0, n)}
		for i := int32(0); i < n; i++ {
			l.Items = append(l.Items, readPayload(br, elem))
		}
		return l
	case TagCompound:
		return readCompoundPayload(br)
	case TagIntArray:
		n := br.readInt32()
		out := make(IntArray, n)
		for i := range out {
			out[i] = br.readInt32()
		}
		return out
	case TagLongArray:
		n := br.readInt32()
		out := make(LongArray, n)
		for i := range out {
			out[i] = br.readInt64()
		}
		return out
	default:
		br.err = fmt.Errorf("nbt: unsupported tag type %d", t)
		return nil
	}
}

func readCompoundPayload(br *byteReader) *Compound {
	c := NewCompound()
	for {
		if br.err != nil {
			return c
		}
		t := Type(br.readByte())
		if t == TagEnd {
			return c
		}
		name := br.readString()
		v := readPayload(br, t)
		c.Put(name, v)
	}
}

// EncodeBytes encodes a document to a plain byte slice.
func EncodeBytes(rootName string, root *Compound) ([]byte, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, rootName, root); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
