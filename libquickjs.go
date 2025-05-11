// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run generator.go

// Package libquickjs is a pure Go embeddable Javascript engine. It supports
// the ECMA script 14 (ES2023) specification including modules, asynchronous
// generators, proxies and BigInt.
//
// This package is the ccgo transpilation of [QuickJS]. The idiomatic Go
// wrapper is available at [modernc.org/quickjs].
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
//	darwin  amd64
//	darwin  arm64
//	freebsd amd64
//	freebsd arm64
//	linux   386
//	linux   amd64
//	linux   arm
//	linux   arm64
//	linux   loong64
//	linux   ppc64le
//	linux   riscv64
//	linux   s390x
//	windows amd64
//	windows arm64
//
// # Builders
//
// Builder results are available [here]:
//
// [QuickJS]: https://bellard.org/quickjs/
// [changelog]: https://bellard.org/quickjs/Changelog
// [here]: https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2flibquickjs
// [modernc.org/quickjs]: https://pkg.go.dev/modernc.org/quickjs
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
	// 1, 2, 4, 8, 16: linux/amd64: nothing works
	// 32: linux/amd64: Some simple tests work
	// 64, 128: linux/amd64: Simple tests pass, some test262 tests work
	// 512: linux/amd64: All tests pass

	// 1<<0, 1, 2, 3: darwin/arm64: Simple tests pass(???), Test2 fails
	// 1<<4: darwin/arm64: Test3 fails, 0: got=<exception=SyntaxError: stack overflow>: result=<exception=ReferenceError: 'result' is not defined>
	// 1<<5: darwin/arm64: Test3 fails, 0: got=<exception=ReferenceError: 'Test262Error' is not defined>: result=FAIL: function at() {\n[native code]\n}
	// 1<<6: darwin/arm64: Test262 completes:  Result: 105/78178 errors, 1610 excluded, 7236 skipped, 36 new, 1 fixed
	// 1<<7: darwin/arm64: Test262 completes:  Result:  77/78178 errors, 1610 excluded, 7236 skipped,  8 new, 1 fixed
	// 1<<8: darwin/arm64: Test262 completes:  Result:  72/78178 errors, 1610 excluded, 7236 skipped,  3 new, 1 fixed
	// 1<<9: darwin/arm64: Test262 completes:  Result:  70/78178 errors, 1610 excluded, 7236 skipped,  1 new, 1 fixed
	// 1<<10: darwin/arm64: Test262 completes: Result:  70/78178 errors, 1610 excluded, 7236 skipped,  1 new, 1 fixed
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
