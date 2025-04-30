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
	"strconv"
	"strings"
	"testing"

	_ "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
	_ "modernc.org/libc"
)

const (
	hash262 = "cba890eef02121251b7f11eb890b0202edf56d6f" // 2025-04-24
	zip262  = hash262 + ".zip"
)

var (
	goarch = runtime.GOARCH
	goos   = runtime.GOOS
	keep   bool
	re     *regexp.Regexp
	xtags  string
)

func TestMain(m *testing.M) {
	oRe := flag.String("re", "", "")
	flag.BoolVar(&keep, "keep", false, "")
	flag.StringVar(&xtags, "xtags", os.Getenv("XTAGS"), "")
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

func copyFile(t *testing.T, dst, src string) {
	if _, err := util.CopyFile(dst, src, nil); err != nil {
		t.Fatal(err)
	}
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

	copyFile(t, filepath.Join(dir, "test262.diff"), filepath.Join("internal", "patch", "test262.diff"))
	copyFile(t, "test262_errors.txt", filepath.Join("testdata", goos, goarch, "test262_errors_updated.txt"))
	git(t, "-C", dir, "init")
	git(t, "-C", dir, "apply", "-v", "test262.diff")
	bin := filepath.Join(dir, "run-test262")
	shell(t, "go", "build", "-o", bin, fmt.Sprintf("-tags=%s", xtags), "./internal/run-test262")
	out, _ := shell(nil, bin, "-t", "-m", "-c", "test262.conf", "-a")
	s := strings.TrimSpace(string(out))
	x := strings.Index(s, "\nResult: ")
	if x < 0 {
		t.Fatal("cannot find test results")
	}

	s = s[x+1:]
	if x = strings.IndexByte(s, '\n'); x > 0 {
		s = s[:x]
	}
	resultLine := strings.TrimSpace(s)
	t.Log(resultLine)
	if _, _, _, _, _, new := parseResultLine(resultLine); new != 0 {
		t.Fatalf("FAIL: new errors=%v", new)
	}

}

// Example: "Result: 36/76685 errors, 1519 excluded, 8703 skipped, 2 fixed\n"
func parseResultLine(s string) (errors, count, excluded, skipped, fixed, new int64) {
	s = strings.TrimSpace(s)
	a := strings.Fields(s)
	// ["Result:" "36/76685" "errors," "1519" "excluded," "8703" "skipped," "2" "fixed"]
	if len(a) < 7 {
		return
	}

	b := strings.Split(a[1], "/")
	if len(b) != 2 {
		return
	}

	errors, _ = strconv.ParseInt(b[0], 0, 64)
	count, _ = strconv.ParseInt(b[1], 0, 64)
	excluded, _ = strconv.ParseInt(a[3], 0, 64)
	skipped, _ = strconv.ParseInt(a[5], 0, 64)
	if len(a) == 9 {
		switch a[8] {
		case "fixed":
			fixed, _ = strconv.ParseInt(a[7], 0, 64)
		case "new":
			new, _ = strconv.ParseInt(a[7], 0, 64)
		}
	}
	return
}
