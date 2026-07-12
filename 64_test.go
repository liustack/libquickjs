// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(386 || arm)

package libquickjs // import "modernc.org/libquickjs"

import (
	"unsafe"
)

func tag(v TJSValue) int32 {
	return int32(v.Ftag)
}

func jsvToFloat64(v TJSValue) (r float64) {
	return *(*float64)(unsafe.Pointer(&v))
}

// jsvToInt32 extracts a JS_TAG_INT payload. On 64-bit the value union is backed
// by a uint64 word (TJSValueUnion.Fuint641); the int is stored in its low 32
// bits, so read that field directly rather than reinterpreting the struct's
// first 4 bytes, which would pick the wrong half on big-endian (e.g. s390x).
func jsvToInt32(v TJSValue) int32 {
	return int32(v.Fu.Fuint641)
}
