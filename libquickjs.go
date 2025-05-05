// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package libquickjs is a CGo-free, ccgo/v4 version of libquickjs.a, a library
// implementing [C QuickJS] - an embeddable Javascript engine.
//
// It supports the ES2023 specification including modules, asynchronous
// generators, proxies and BigInt.
//
// # QuickJS Version
//
// Release 2025-04-26, [changelog].
//
// # Supported platforms and architectures
//
// These combinations of GOOS and GOARCH are currently supported
//
//	OS      Arch
//	---------------
//	linux   386
//	linux   amd64
//	linux   arm
//	linux   arm64
//	linux   loong64
//	linux   ppc64le
//	linux   riscv64
//	linux   s390x
//
// # Builders
//
// Builder results available [here]:
//
// # Known bugs or limitations
//
//   - The quickjs std package does not implement popen(), exec().
//
// [C QuickJS]: https://bellard.org/quickjs/
// [changelog]: https://bellard.org/quickjs/Changelog
// [here]: https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2flibquickjs
package libquickjs // import "modernc.org/libquickjs"

import (
	"modernc.org/libc"
)

func init() {
	Xinit(nil)
}

const (
	// MaxStackSlots defines the maximum allowed number of libc.TLS stack slots
	// before a stack overflow is detected.
	MaxStackSlots = 1000
	// 1, 2, 4, 8, 16: nothing works
	// 32: Some simple tests work
	// 64, 128: Simple tests pass, some test262 tests work
	// 512: All tests pass
)

func _js_check_stack_overflow(tls *libc.TLS, rt uintptr, sz Size_t) int32 {
	if tls.StackSlots() < MaxStackSlots {
		return 0
	}

	return 1
}

func _getcwd0(tls *libc.TLS, p uintptr, n Size_t) uintptr {
	return libc.Xgetcwd(tls, p, n)
}
