// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package libquickjs is a CGo-free, ccgo/v4 version of libquickjs.a, a library
// implementing [QuickJS] - an embeddable Javascript engine.
//
// It supports the ES2023 specification including modules, asynchronous
// generators, proxies and BigInt.
//
// # Supported platforms and architectures
//
// These combinations of GOOS and GOARCH are currently supported
//
//	OS      Arch
//	---------------
//	linux	amd64
//
// # Builders
//
// Builder results available at:
//
// https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2flibquickjs
//
// [QuickJS]: https://bellard.org/quickjs/
package libquickjs // import "modernc.org/libquickjs"

func init() {
	Xinit(nil)
}
