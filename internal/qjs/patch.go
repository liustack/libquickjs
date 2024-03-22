// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"sync/atomic"
	"unsafe"

	"modernc.org/libc"
	"modernc.org/libquickjs/internal/snprintf"
)

// type __atomic_fetch_add (type *ptr, type val, int memorder)
func ___atomic_fetch_add(tls *libc.TLS, p uintptr, n, _ int32) int32 {
	return atomic.AddInt32((*int32)(unsafe.Pointer(p)), n)
}

func _snprintf_with_rounding_mode(tls *libc.TLS, rounding_mode int32, s uintptr, n Tsize_t, nfmt, width int32, d float64) (r int32) {
	return snprintf.SnprintfWithRoundingMode(tls, rounding_mode, s, n, nfmt, width, d)
}
