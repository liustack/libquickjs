// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build 386 || arm

package libquickjs // import "modernc.org/libquickjs"

import "modernc.org/libc"

// ArrayOf returns a new Javascript array from a vector if TJSValues.
func ArrayOf(tls *libc.TLS, ctx uintptr, argc int32, argv uintptr) (r TJSValue) {
	return _js_array_of(tls, ctx, libc.Uint64FromInt32(EJS_TAG_UNDEFINED)<<32, argc, argv)
}
