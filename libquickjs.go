// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package libquickjs is a ccgo/v4 version of libquickjs.a, a library
// implementing an embeddable Javascript engine.
package libquickjs // import "modernc.org/libquickjs"

import (
	"sync/atomic"
	"unsafe"

	"modernc.org/libc"
)

// type __atomic_fetch_add (type *ptr, type val, int memorder)
func ___atomic_fetch_add(tls *libc.TLS, p uintptr, n, _ int32) int32 {
	return atomic.AddInt32((*int32)(unsafe.Pointer(p)), n)
}
