// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package libquickjs is a CGo-free, ccgo/v4 version of libquickjs.a, a library
// implementing an embeddable Javascript engine.
//
// It supports the ES2023 specification including modules, asynchronous
// generators, proxies and BigInt.  It optionally supports mathematical
// extensions such as big decimal floating point numbers (BigDecimal), big
// binary floating point numbers (BigFloat) and operator overloading.
//
// # Supported platforms and architectures
//
// These combinations of GOOS and GOARCH are currently supported
//
//	OS      Arch
//	---------------
//	linux	amd64
//	linux	loong64
//
// # Builders
//
// Builder results available at:
//
// https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2flibquickjs
//
// If a builder reports PASS then this package passed all the [test262] suite
// test cases the original C version passed on the respective target.
//
// [test262]: https://github.com/tc39/test262
package libquickjs // import "modernc.org/libquickjs"

import (
	"runtime"
	"sync"

	"modernc.org/libc"
	"modernc.org/libquickjs/internal/snprintf"
)

var (
	pinnersMu sync.Mutex
	pinners   = map[*Library]*runtime.Pinner{}
)

func _snprintf_with_rounding_mode(tls *libc.TLS, rounding_mode int32, s uintptr, n Tsize_t, nfmt, width int32, d float64) (r int32) {
	return snprintf.SnprintfWithRoundingMode(tls, rounding_mode, s, n, nfmt, width, d)
}

// ArrayOf returns a new Javascript array from a vector if TJSValues.
func (qjs *Library) ArrayOf(tls *libc.TLS, ctx uintptr, argc int32, argv uintptr) (r TJSValue) {
	return g_js_array_of(tls, qjs, ctx, TJSValue{Ftag: EJS_TAG_UNDEFINED}, argc, argv)
}

// NewLibrary returns a newly created Library instance.
func NewLibrary() (r *Library, err error) {
	r = &Library{}
	p := &runtime.Pinner{}
	p.Pin(r)
	pinnersMu.Lock()
	pinners[r] = p
	pinnersMu.Unlock()
	r.initInstance()
	return r, nil
}

// Close releases the resouces held by 'qjs'.
func (qjs *Library) Close() error {
	pinnersMu.Lock()
	p := pinners[qjs]
	delete(pinners, qjs)
	pinnersMu.Unlock()
	p.Unpin()
	*qjs = Library{}
	return nil
}
