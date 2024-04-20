// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package libquickjs is a ccgo/v4 version of libquickjs.a, a library
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
package libquickjs // import "modernc.org/libquickjs"

import (
	"sync"

	"modernc.org/libc"
	"modernc.org/libquickjs/internal/snprintf"
)

var initOnce sync.Once

// Init initializes this package. Init can be safely called multiple times.
func Init() {
	initOnce.Do(func() {
		tls := libc.NewTLS()

		defer tls.Close()

		Xinit_quickjs(tls)
	})
}

func _snprintf_with_rounding_mode(tls *libc.TLS, rounding_mode int32, s uintptr, n Tsize_t, nfmt, width int32, d float64) (r int32) {
	return snprintf.SnprintfWithRoundingMode(tls, rounding_mode, s, n, nfmt, width, d)
}
