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
	"unsafe"

	_ "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
	"modernc.org/libc"
)

const (
	hash262 = "56e77d6325067a545ea7e8ff5be5d9284334e33c" // 2025-03-25
	zip262  = hash262 + ".zip"
)

var (
	goarch = runtime.GOARCH
	goos   = runtime.GOOS
	keep   bool
	re     *regexp.Regexp
	target = fmt.Sprintf("%s/%s", goos, goarch)
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
	for _, v := range []string{
		"test_closure.js",
		"test_language.js",
		"test_loop.js",
		"test_bigint.js",
	} {
		if re != nil && !re.MatchString(v) {
			continue
		}
		t.Run(v, func(t *testing.T) {
			fn := filepath.Join("testdata", "tests", v)
			eval(t, fn)
		})
	}
}

var evalFN = [...]byte{'<', 'e', 'v', 'a', 'l', '>', 0}

func eval(t *testing.T, fn string) {
	b, err := os.ReadFile(fn)
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	tls := libc.NewTLS()

	defer tls.Close()

	rt := XJS_NewRuntime(tls)

	defer XJS_FreeRuntime(tls, rt)

	ctx := XJS_NewContext(tls, rt)

	defer XJS_FreeContext(tls, ctx)

	cs, err := libc.CString(string(b))
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	defer libc.Xfree(tls, cs)

	val := XJS_Eval(tls, ctx, cs, Size_t(len(b)), uintptr(unsafe.Pointer(&evalFN)), MJS_EVAL_TYPE_GLOBAL)
	tv := tag(val)
	if tv < 0 {
		defer XFreeValue(tls, ctx, val)
	}

	switch tv {
	case EJS_TAG_UNDEFINED:
		// ok
	case EJS_TAG_EXCEPTION:
		e := XJS_GetException(tls, ctx)

		defer XFreeValue(tls, ctx, e)

		p := XToCString(tls, ctx, e)

		defer XJS_FreeCString(tls, ctx, p)

		t.Fatalf("FAIL: fn=%s err=%s", fn, libc.GoString(p))
	default:
		t.Fatalf("FAIL: fn=%s tag=%v", fn, tv)
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
	_, err := os.Stat(zip262)
	if err != nil {
		t.Fatalf("%s: to download the test262 archive issue '$ make download'", err)
	}

	os.RemoveAll(dir)

	if !keep {
		defer func() {
			os.RemoveAll(dir)
		}()
	}

	shell(t, "unzip", "-q", zip262)
	if err = os.Rename("test262-"+hash262, dir); err != nil {
		t.Fatal(err)
	}

	copyFile(t, filepath.Join(dir, "test262.diff"), filepath.Join("internal", "patch", "test262.diff"))
	git(t, "-C", dir, "init")
	git(t, "-C", dir, "apply", "-v", "test262.diff")
	bin := filepath.Join(dir, "run-test262")
	if goos == "windows" {
		bin += ".exe"
	}
	shell(t, "go", "build", "-o", bin, fmt.Sprintf("-tags=%s", xtags), "./internal/run-test262")
	var out []byte
	switch goos {
	case "windows":
		out, err = shell(nil, bin, "-v", "-t", "-m", "-c", "test262.conf", "-a")
	default:
		out, err = shell(nil, bin, "-v", "-t", "-m", "-c", "test262.conf", "-a")
	}
	if err != nil {
		t.Logf("err=%v", err)
	}
	s := strings.TrimSpace(string(out))
	x := strings.Index(s, "Result: ")
	if x < 0 {
		t.Fatal("cannot find test results")
	}

	s = s[x:]
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
