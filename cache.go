package lavish

import (
	"encoding/binary"
	"fmt"
	gohash "hash"
	"hash/fnv"

	"github.com/dop251/goja"
)

type ProgramCache interface {
	AddProgram(k uint64, v *goja.Program)
	GetProgram(k uint64) (cached *goja.Program)
}

func hash(src string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(src)) // fnv.sum64a.Write cannot error
	return binary.BigEndian.Uint64(h.Sum(nil))
}

// Hashable allows data objects to provide a cache key for render caching.
type Hashable interface {
	Hash64() uint64
}

type RenderCache interface {
	AddRender(program uint64, data uint64, v string)
	GetRender(program uint64, data uint64) (cached string, ok bool)
}

// HashBuilder provides a fluent API for building hash values.
// Use this to simplify implementing the Hashable interface.
//
// Example:
//
//	func (d MyData) Hash64() uint64 {
//	    return lavish.NewHashBuilder().
//	        String(d.Name).
//	        Int(d.Count).
//	        Strings(d.Items).
//	        Sum()
//	}
type HashBuilder struct {
	h gohash.Hash64
}

// NewHashBuilder creates a new HashBuilder using FNV-64a.
func NewHashBuilder() *HashBuilder {
	return &HashBuilder{h: fnv.New64a()}
}

// String adds a string value to the hash.
func (hb *HashBuilder) String(s string) *HashBuilder {
	_, _ = hb.h.Write([]byte(s))
	return hb
}

// Strings adds multiple string values to the hash.
func (hb *HashBuilder) Strings(ss []string) *HashBuilder {
	for _, s := range ss {
		_, _ = hb.h.Write([]byte(s))
	}
	return hb
}

// Int adds an integer value to the hash.
func (hb *HashBuilder) Int(i int) *HashBuilder {
	_ = binary.Write(hb.h, binary.BigEndian, int64(i))
	return hb
}

// Int64 adds an int64 value to the hash.
func (hb *HashBuilder) Int64(i int64) *HashBuilder {
	_ = binary.Write(hb.h, binary.BigEndian, i)
	return hb
}

// Bool adds a boolean value to the hash.
func (hb *HashBuilder) Bool(b bool) *HashBuilder {
	if b {
		_, _ = hb.h.Write([]byte{1})
	} else {
		_, _ = hb.h.Write([]byte{0})
	}
	return hb
}

// Bytes adds raw bytes to the hash.
func (hb *HashBuilder) Bytes(b []byte) *HashBuilder {
	_, _ = hb.h.Write(b)
	return hb
}

// Any adds any value to the hash using fmt.Sprint.
// For better performance, prefer the typed methods.
func (hb *HashBuilder) Any(v any) *HashBuilder {
	_, _ = hb.h.Write([]byte(fmt.Sprint(v)))
	return hb
}

// Sum returns the final hash value.
func (hb *HashBuilder) Sum() uint64 {
	return binary.BigEndian.Uint64(hb.h.Sum(nil))
}
