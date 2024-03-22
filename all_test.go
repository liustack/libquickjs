// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libquickjs // import "modernc.org/libquickjs"

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	_ "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
	_ "modernc.org/libc"
)

var (
	goos   = runtime.GOOS
	goarch = runtime.GOARCH
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	tmpdir := t.TempDir()
	bin := "./" + filepath.Join(tmpdir, "qjs")
	if out, err := util.Shell(nil, "go", "build", "-o", bin, "./"+filepath.Join("internal", "qjs")); err != nil {
		t.Fatalf("FAIL err=%v out=%s", err, out)
	}

	root := filepath.Join("internal", "tests")
	for _, test := range [][]string{
		// test: qjs
		// 	./qjs tests/test_closure.js
		// 	./qjs tests/test_language.js
		// 	./qjs --std tests/test_builtin.js
		// 	./qjs tests/test_loop.js
		// 	./qjs tests/test_bignum.js
		// 	./qjs tests/test_std.js
		// 	./qjs tests/test_worker.js
		// ifdef CONFIG_SHARED_LIBS
		// ifdef CONFIG_BIGNUM
		// 	./qjs --bignum tests/test_bjson.js
		// else
		// 	./qjs tests/test_bjson.js
		// endif
		// 	./qjs examples/test_point.js
		// endif
		// ifdef CONFIG_BIGNUM
		// 	./qjs --bignum tests/test_op_overloading.js
		// 	./qjs --bignum tests/test_bigfloat.js
		// 	./qjs --qjscalc tests/test_qjscalc.js
		// endif
		{filepath.Join(root, "test_closure.js")},
		{filepath.Join(root, "test_language.js")},
		{"--std", filepath.Join(root, "test_builtin.js")}, //TODO
		{filepath.Join(root, "test_loop.js")},
		{filepath.Join(root, "test_bignum.js")},
		//TODO {filepath.Join(root, "test_std.js")},
		//TODO {filepath.Join(root, "test_worker.js")},
		//TODO {"--bignum", filepath.Join(root, "test_bjson.js")}, // could not load module filename 'internal/tests/bjson.so' as shared library
		//TODO {filepath.Join(root, "test_bjson.js")}, // // could not load module filename 'internal/tests/bjson.so' as shared library
		//TODO examples
		{"--bignum", filepath.Join(root, "test_op_overloading.js")},
		//TODO {"--bignum", filepath.Join(root, "test_bigfloat.js")},
		//TODO {"--qjscalc", filepath.Join(root, "test_qjscalc.js")},
	} {
		out, err := util.Shell(nil, bin, test...)
		if err != nil {
			s := string(out)
			s = strings.TrimSpace(s)
			a := strings.Split(s, "\n")
			s = ""
			if len(a) != 0 {
				s = a[0]
			}
			t.Errorf("%v: err=%v out=%s", test, err, s)
		}
	}
}
