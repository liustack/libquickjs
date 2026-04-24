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
	hash262 = "a5e69a1534de88d1eb29b76657d84c8541b72df7" // 2025-09-10
	zip262  = hash262 + ".zip"
)

var (
	builder    = os.Getenv("MODERNC_BUILDER") != ""
	goarch     = runtime.GOARCH
	goos       = runtime.GOOS
	keep       bool
	re         *regexp.Regexp
	singleFile string
	target     = fmt.Sprintf("%s/%s", goos, goarch)
	xtags      string
)

func TestMain(m *testing.M) {
	oRe := flag.String("re", "", "")
	flag.BoolVar(&keep, "keep", false, "")
	flag.StringVar(&xtags, "xtags", os.Getenv("XTAGS"), "")
	flag.StringVar(&singleFile, "f", "", "")
	flag.Parse()
	if s := *oRe; s != "" {
		re = regexp.MustCompile(s)
	}
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	for _, v := range []string{
		"test_bigint.js",
		"test_closure.js",
		"test_language.js",
		"test_loop.js",
		"test_std.js",
	} {
		if re != nil && !re.MatchString(v) {
			continue
		}
		t.Run(v, func(t *testing.T) {
			fn := filepath.Join("testdata", "tests", v)
			evalFile(t, fn)
		})
	}
}

func Test2(t *testing.T) {
	tests := []struct{ in, out string }{
		{"12+34", "46"},
		{"new Date(1899, 0).getYear()", "-1"},
		{"new Date(1899, 11, 31, 23, 59, 59, 999).getYear()", "-1"},
		{"new Date(1900, 0).getYear()", "0"},
		{"new Date(1900, 11, 31, 23, 59, 59, 999).getYear()", "0"},
		{"new Date(1970, 0).getYear()", "70"},
		{"new Date(1970, 11, 31, 23, 59, 59, 999).getYear()", "70"},
		{"new Date(2000, 0).getYear()", "100"},
		{"new Date(2000, 11, 31, 23, 59, 59, 999).getYear()", "100"},
	}
	for i, v := range tests {
		fn := fmt.Sprint(i)
		if re != nil && !re.MatchString(fn) {
			continue
		}

		t.Run(fn, func(t *testing.T) {
			switch g, e := eval(t, fn, v.in), v.out; {
			case g != e:
				t.Errorf("%v: src=%s got=%q expected=%q", fn, v.in, g, e)
			default:
				t.Logf("%3d: %q\n%[2]s", i, g)
			}
		})
	}
	tls := libc.NewTLS()

	defer tls.Close()

	tz, err := libc.CString("TZ")
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	defer libc.Xfree(tls, tz)

	const la = "America/Los_Angeles"
	loc, err := libc.CString(la)
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	defer libc.Xfree(tls, loc)

	libc.Xsetenv(tls, tz, loc, 1)
	if g, e := libc.GoString(libc.Xgetenv(tls, tz)), la; g != e {
		t.Fatalf("%q", g)
	}

	for i, v := range tests {
		fn := fmt.Sprint(i)
		if re != nil && !re.MatchString(fn) {
			continue
		}

		t.Run(fn, func(t *testing.T) {
			switch g, e := eval(t, fn, v.in), v.out; {
			case g != e:
				t.Errorf("%v: src=%s got=%q expected=%q", fn, v.in, g, e)
			default:
				t.Logf("%3d: %q\n%[2]s", i, g)
			}
		})
	}
}

var evalFN = [...]byte{'<', 'e', 'v', 'a', 'l', '>', 0}

func evalFile(t testing.TB, fn string) (r string) {
	b, err := os.ReadFile(filepath.FromSlash(fn))
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	return eval(t, fn, string(b))
}

func eval0(t testing.TB, tls *libc.TLS, ctx uintptr, js string) (r TJSValue) {
	cs, err := libc.CString(js)
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	defer libc.Xfree(tls, cs)

	return XJS_Eval(tls, ctx, cs, Size_t(len(js)), uintptr(unsafe.Pointer(&evalFN)), MJS_EVAL_TYPE_GLOBAL)
}

func eval1(t testing.TB, tls *libc.TLS, ctx uintptr, js string) (r string) {
	return toString(t, tls, ctx, eval0(t, tls, ctx, js))
}

func toString(t testing.TB, tls *libc.TLS, ctx uintptr, val TJSValue) (r string) {
	tv := tag(val)
	if tv < 0 {
		defer XFreeValue(tls, ctx, val)
	}

	switch tv {
	case EJS_TAG_UNDEFINED:
		return "<undefined>"
	case EJS_TAG_EXCEPTION:
		e := XJS_GetException(tls, ctx)

		defer XFreeValue(tls, ctx, e)

		p := XToCString(tls, ctx, e)

		defer XJS_FreeCString(tls, ctx, p)

		return fmt.Sprintf("<exception=%s>", libc.GoString(p))
	case EJS_TAG_INT:
		return fmt.Sprint(*(*int32)(unsafe.Pointer(&val)))
	case EJS_TAG_FLOAT64:
		return fmt.Sprint(jsvToFloat64(val))
	case
		EJS_TAG_STRING_ROPE, // -6,
		EJS_TAG_STRING:      // -7,

		p := XToCString(tls, ctx, val)

		defer XJS_FreeCString(tls, ctx, p)

		return libc.GoString(p)
	}

	t.Fatalf("FAIL: tag=%v", tv)
	panic(todo(""))
}

func eval(t testing.TB, fn, js string) (r string) {
	tls := libc.NewTLS()

	defer tls.Close()

	rt := XJS_NewRuntime(tls)

	defer XJS_FreeRuntime(tls, rt)

	ctx := XJS_NewContext(tls, rt)

	defer XJS_FreeContext(tls, ctx)

	cs, err := libc.CString(js)
	if err != nil {
		t.Fatalf("FAIL: err=%v", err)
	}

	defer libc.Xfree(tls, cs)

	if r = eval1(t, tls, ctx, js); strings.HasPrefix(r, "<") {
		r = fmt.Sprintf("%s: result=%s", r, eval1(t, tls, ctx, "result"))
	}
	return r
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

func fileExists(fn string) bool {
	switch fi, err := os.Stat(fn); {
	case err == nil && !fi.IsDir():
		return true
	default:
		return false
	}
}

func Test262(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	const dir = "test262"
	if !fileExists(zip262) {
		if builder {
			t.Fatal("missing test262 archive")
		}

		t.Skip("to download the test262 archive issue '$ make download'")
	}

	os.RemoveAll(dir)

	if !keep {
		defer func() {
			os.RemoveAll(dir)
		}()
	}

	switch goos {
	case "windows":
		shell(t, "git-bash", "-c", fmt.Sprintf("unzip -q %s", zip262))
	default:
		shell(t, "unzip", "-q", zip262)
	}
	err := os.Rename("test262-"+hash262, dir)
	if err != nil {
		t.Fatal(err)
	}

	cnf := filepath.Join("testdata", fmt.Sprintf("test262_%s_%s.conf", goos, goarch))
	if !fileExists(cnf) {
		cnf = filepath.Join("testdata", "test262.conf")
	}
	t.Logf("using %s", cnf)
	copyFile(t, "test262.conf", cnf)
	copyFile(t, filepath.Join(dir, "test262.diff"), filepath.Join("internal", "patch", "test262.diff"))
	git(t, "-C", dir, "init")
	git(t, "-C", dir, "apply", "-v", "test262.diff")
	bin := filepath.Join(dir, "run-test262")
	if goos == "windows" {
		bin += ".exe"
	}
	shell(t, "go", "build", "-o", bin, fmt.Sprintf("-tags=%s", xtags), "./internal/run-test262")
	var out []byte
	args := []string{"-v", "-t", "-m", "-c", "test262.conf"}
	switch {
	case singleFile != "":
		args = append(args, "-f", singleFile)
	default:
		args = append(args, "-a")
	}
	if fn := filepath.Join("testdata", fmt.Sprintf("test262_exclude_%s_%s", goos, goarch)); fileExists(fn) {
		args = append(args, "-x", fn)
	}
	out, err = shell(nil, bin, args...)
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
	switch newErrors, err := parseResultLine(resultLine); {
	case err != nil:
		t.Fatalf("FAIL: err=%v", err)
	case newErrors != 0:
		t.Fatalf("FAIL: new errors=%v", newErrors)
	}
	t.Log(resultLine)
}

// Example: "Result: 36/76685 errors, 1519 excluded, 8703 skipped, 2 fixed\n"
func parseResultLine(s string) (newErrors int64, err error) {
	s = strings.TrimSpace(s)
	a := strings.Fields(s)
	// [0: "Result:" 1: "36/76685" 2: "errors," 3: "1519" 4: "excluded," 5: "8703" 6: "skipped," 7: "2" 8: "fixed"]
	if len(a) < 7 {
		return 0, fmt.Errorf("malformed result line: `%s`", s)
	}

	if len(a) > 7 {
		switch a[8] {
		case "new":
			newErrors, err = strconv.ParseInt(a[7], 0, 64)
		}
	}
	return newErrors, err
}

// https://gitlab.com/cznic/quickjs/-/issues/6
//
// Turns out it's not bug in this package, see https://gitlab.com/cznic/quickjs/-/issues/6#note_2830309150
//
// func TestIssue6(t *testing.T) {
// 	tests := []struct{ in, out string }{
// 		{`100000000000000000000000.000000000000001`, "1.0000000000000001e+23"},              // https://gitlab.com/cznic/quickjs/-/issues/6
// 		{`+"100000000000000000000000.000000000000001"`, "1.0000000000000001e+23"},           // https://gitlab.com/cznic/quickjs/-/issues/6
// 		{`(100000000000000000000000.000000000000001).toString()`, "1.0000000000000001e+23"}, // https://gitlab.com/cznic/quickjs/-/issues/6
// 	}
// 	for i, v := range tests {
// 		fn := fmt.Sprint(i)
// 		if re != nil && !re.MatchString(fn) {
// 			continue
// 		}
//
// 		t.Run(fn, func(t *testing.T) {
// 			switch g, e := eval(t, fn, v.in), v.out; {
// 			case g != e:
// 				t.Errorf("%v: src=%s got=%q expected=%q", fn, v.in, g, e)
// 			default:
// 				t.Logf("%3d: %q\n%[2]s", i, g)
// 			}
// 		})
// 	}
// }
