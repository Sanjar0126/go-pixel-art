package nbt

import (
	"bytes"
	"reflect"
	"testing"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	root := NewCompound()
	root.Put("DataVersion", Int(2586))
	root.Put("Name", String("hello"))
	root.Put("Pi", Float64(3.5))
	root.Put("Flag", Byte(1))
	root.Put("Big", Long(-123456789))
	root.Put("Ints", IntArray{1, 2, 3})
	root.Put("Longs", LongArray{10, 20, 30})
	list := &List{Elem: TagString, Items: []Value{String("a"), String("b")}}
	root.Put("List", list)
	nested := NewCompound()
	nested.Put("Inner", Short(7))
	root.Put("Nested", nested)

	var buf bytes.Buffer
	if err := Encode(&buf, "", root); err != nil {
		t.Fatalf("encode: %v", err)
	}

	name, got, err := Decode(&buf)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if name != "" {
		t.Fatalf("expected empty root name, got %q", name)
	}

	check := func(key string, want Value) {
		v, ok := got.Get(key)
		if !ok {
			t.Fatalf("missing key %q", key)
		}
		if !reflect.DeepEqual(v, want) {
			t.Fatalf("key %q: got %#v, want %#v", key, v, want)
		}
	}
	check("DataVersion", Int(2586))
	check("Name", String("hello"))
	check("Pi", Float64(3.5))
	check("Flag", Byte(1))
	check("Big", Long(-123456789))
	check("Ints", IntArray{1, 2, 3})
	check("Longs", LongArray{10, 20, 30})

	gotList, ok := got.Get("List")
	if !ok {
		t.Fatal("missing List")
	}
	l, ok := gotList.(*List)
	if !ok || len(l.Items) != 2 || l.Items[0] != String("a") || l.Items[1] != String("b") {
		t.Fatalf("unexpected list: %#v", gotList)
	}

	gotNested, ok := got.Get("Nested")
	if !ok {
		t.Fatal("missing Nested")
	}
	nc, ok := gotNested.(*Compound)
	if !ok {
		t.Fatalf("Nested is not a compound: %#v", gotNested)
	}
	inner, ok := nc.Get("Inner")
	if !ok || inner != Short(7) {
		t.Fatalf("unexpected Inner: %#v", inner)
	}
}

func TestEmptyList(t *testing.T) {
	root := NewCompound()
	root.Put("Empty", &List{Elem: TagEnd})

	data, err := EncodeBytes("", root)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	_, got, err := Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	v, ok := got.Get("Empty")
	if !ok {
		t.Fatal("missing Empty")
	}
	l := v.(*List)
	if len(l.Items) != 0 {
		t.Fatalf("expected empty list, got %d items", len(l.Items))
	}
}
