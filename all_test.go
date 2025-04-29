// Copyright 2025 The libquickjs-go Authors. All rights reserved.
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
	goos = runtime.GOOS
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	tmp := t.TempDir()
	if !filepath.IsAbs(tmp) {
		t.Fatal()
	}

	bin := filepath.Join(tmp, "gqjs")
	shell(t, "go", "build", "-o", bin, "./gqjs")
	for _, v := range []string{
		"test_closure.js",
		"test_language.js",
		"--std test_builtin.js",
		"test_loop.js",
		"test_bigint.js",
		"test_cyclic_import.js",
		"test_worker.js",
		//TODO "test_std.js",
	} {
		t.Run(v, func(t *testing.T) {
			args := strings.Fields(v)
			x := len(args) - 1
			args[x] = filepath.Join("testdata", "tests", args[x])
			if out, err := util.Shell(nil, bin, args...); err != nil && t != nil {
				t.Errorf("FAIL: err=%v out=%s", err, out)
			}
		})
	}
}

func shell(t *testing.T, cmd string, args ...string) (out []byte, err error) {
	if out, err = util.Shell(nil, cmd, args...); err != nil && t != nil {
		t.Fatalf("FAIL: err=%v out=%s", err, out)
	}

	return out, err
}
