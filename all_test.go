// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libquickjs // import "modernc.org/libquickjs"

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	_ "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
	_ "modernc.org/libc"
)

const (
	hash262 = "cba890eef02121251b7f11eb890b0202edf56d6f"
	zip262  = hash262 + ".zip"
)

var (
	goos  = runtime.GOOS
	keep  bool
	re    *regexp.Regexp
	xtags string
)

func TestMain(m *testing.M) {
	oRe := flag.String("re", "", "")
	flag.BoolVar(&keep, "keep", false, "")
	flag.StringVar(&xtags, "xtags", "", "")
	flag.Parse()
	if s := *oRe; s != "" {
		re = regexp.MustCompile(s)
	}
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	tmp := t.TempDir()
	if !filepath.IsAbs(tmp) {
		t.Fatal()
	}

	bin := filepath.Join(tmp, "gqjs")
	shell(t, "go", "build", "-o", bin, fmt.Sprintf("-tags=%s", xtags), "./gqjs")
	for _, v := range []string{
		"test_closure.js",
		"test_language.js",
		"--std test_builtin.js",
		"test_loop.js",
		"test_bigint.js",
		"test_cyclic_import.js",
		"test_worker.js",
		"test_std.js",
	} {
		if re != nil && !re.MatchString(v) {
			continue
		}

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

func git(t *testing.T, args ...string) {
	shell(t, "git", args...)
}

func Test262(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const dir = "test262"
	if _, err := os.Stat(zip262); err != nil {
		t.Fatalf("%s: to download the test262 archive issue '$ make download'", err)
	}

	os.RemoveAll(dir)

	if !keep {
		defer func() {
			os.RemoveAll(dir)
		}()
	}

	shell(t, "unzip", "-q", zip262)
	if err := os.Rename("test262-"+hash262, dir); err != nil {
		t.Fatal(err)
	}

	if _, err := util.CopyFile(filepath.Join(dir, "test262.diff"), filepath.Join("internal", "patch", "test262.diff"), nil); err != nil {
		t.Fatal(err)
	}

	git(t, "-C", dir, "init")
	git(t, "-C", dir, "apply", "-v", "test262.diff")
	bin := filepath.Join(dir, "run-test262")
	shell(t, "go", "build", "-o", bin, fmt.Sprintf("-tags=%s", xtags), "./internal/run-test262")
	shell(t, bin, "-t", "-m", "-c", "test262.conf", "-a", "-v")
}
