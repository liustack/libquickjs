// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libquickjs // import "modernc.org/libquickjs"

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	_ "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
	"modernc.org/libc"
)

var (
	goos   = runtime.GOOS
	goarch = runtime.GOARCH
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var str [128]byte

func TestSnprintfWithRoundingMode(t *testing.T) {
	// $ cat main.c
	// #include <stdio.h>
	// #include <fenv.h>
	//
	// char str[128];
	//
	// int main() {
	// 	char *names[] = {"FE_DOWNWARD", "FE_UPWARD"};
	// 	int modes[] = {FE_DOWNWARD, FE_UPWARD};
	// 	double ds[] = {1234.5678, 25};
	// 	double signs[] = {-1, 1};
	// 	for (int i = 0; i < 2; i++) {
	// 		int mode = modes[i];
	//              fesetround(mode);
	// 		char *name = names[i];
	// 		for (int j = 0; j < 2; j++) {
	// 			double d0 = ds[j];
	// 			for (int k = 0; k < 2; k++) {
	// 				double d = d0*signs[k];
	// 				for (int prec = 0; prec < 8; prec++) {
	// 					snprintf(str, 128, "%+.*e", prec, d);
	// 					printf("mode=%s d=%f prec=%d s=`%s`\n", name, d, prec, str);
	// 				}
	// 			}
	// 		}
	// 	}
	// }
	// $ gcc main.c && ./a.out
	// mode=FE_DOWNWARD d=-1234.567801 prec=0 s=`-2e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=1 s=`-1.3e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=2 s=`-1.24e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=3 s=`-1.235e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=4 s=`-1.2346e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=5 s=`-1.23457e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=6 s=`-1.234568e+03`
	// mode=FE_DOWNWARD d=-1234.567801 prec=7 s=`-1.2345679e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=0 s=`+1e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=1 s=`+1.2e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=2 s=`+1.23e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=3 s=`+1.234e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=4 s=`+1.2345e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=5 s=`+1.23456e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=6 s=`+1.234567e+03`
	// mode=FE_DOWNWARD d=1234.567800 prec=7 s=`+1.2345678e+03`
	// mode=FE_DOWNWARD d=-25.000000 prec=0 s=`-3e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=1 s=`-2.5e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=2 s=`-2.50e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=3 s=`-2.500e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=4 s=`-2.5000e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=5 s=`-2.50000e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=6 s=`-2.500000e+01`
	// mode=FE_DOWNWARD d=-25.000000 prec=7 s=`-2.5000000e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=0 s=`+2e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=1 s=`+2.5e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=2 s=`+2.50e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=3 s=`+2.500e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=4 s=`+2.5000e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=5 s=`+2.50000e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=6 s=`+2.500000e+01`
	// mode=FE_DOWNWARD d=25.000000 prec=7 s=`+2.5000000e+01`
	// mode=FE_UPWARD d=-1234.567800 prec=0 s=`-1e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=1 s=`-1.2e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=2 s=`-1.23e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=3 s=`-1.234e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=4 s=`-1.2345e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=5 s=`-1.23456e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=6 s=`-1.234567e+03`
	// mode=FE_UPWARD d=-1234.567800 prec=7 s=`-1.2345678e+03`
	// mode=FE_UPWARD d=1234.567801 prec=0 s=`+2e+03`
	// mode=FE_UPWARD d=1234.567801 prec=1 s=`+1.3e+03`
	// mode=FE_UPWARD d=1234.567801 prec=2 s=`+1.24e+03`
	// mode=FE_UPWARD d=1234.567801 prec=3 s=`+1.235e+03`
	// mode=FE_UPWARD d=1234.567801 prec=4 s=`+1.2346e+03`
	// mode=FE_UPWARD d=1234.567801 prec=5 s=`+1.23457e+03`
	// mode=FE_UPWARD d=1234.567801 prec=6 s=`+1.234568e+03`
	// mode=FE_UPWARD d=1234.567801 prec=7 s=`+1.2345679e+03`
	// mode=FE_UPWARD d=-25.000000 prec=0 s=`-2e+01`
	// mode=FE_UPWARD d=-25.000000 prec=1 s=`-2.5e+01`
	// mode=FE_UPWARD d=-25.000000 prec=2 s=`-2.50e+01`
	// mode=FE_UPWARD d=-25.000000 prec=3 s=`-2.500e+01`
	// mode=FE_UPWARD d=-25.000000 prec=4 s=`-2.5000e+01`
	// mode=FE_UPWARD d=-25.000000 prec=5 s=`-2.50000e+01`
	// mode=FE_UPWARD d=-25.000000 prec=6 s=`-2.500000e+01`
	// mode=FE_UPWARD d=-25.000000 prec=7 s=`-2.5000000e+01`
	// mode=FE_UPWARD d=25.000000 prec=0 s=`+3e+01`
	// mode=FE_UPWARD d=25.000000 prec=1 s=`+2.5e+01`
	// mode=FE_UPWARD d=25.000000 prec=2 s=`+2.50e+01`
	// mode=FE_UPWARD d=25.000000 prec=3 s=`+2.500e+01`
	// mode=FE_UPWARD d=25.000000 prec=4 s=`+2.5000e+01`
	// mode=FE_UPWARD d=25.000000 prec=5 s=`+2.50000e+01`
	// mode=FE_UPWARD d=25.000000 prec=6 s=`+2.500000e+01`
	// mode=FE_UPWARD d=25.000000 prec=7 s=`+2.5000000e+01`
	// $
	tls := libc.NewTLS()

	defer tls.Close()

	sp := uintptr(unsafe.Pointer(&str))
	var fails []string
	ok := 0
	for itest, test := range []struct {
		mode int32
		d    float64
		prec int32
		s    string
	}{
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 0, s: `-2e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 1, s: `-1.3e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 2, s: `-1.24e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 3, s: `-1.235e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 4, s: `-1.2346e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 5, s: `-1.23457e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 6, s: `-1.234568e+03`},
		// {mode: libc.FE_DOWNWARD, d: -1234.567801, prec: 7, s: `-1.2345679e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 0, s: `+1e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 1, s: `+1.2e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 2, s: `+1.23e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 3, s: `+1.234e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 4, s: `+1.2345e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 5, s: `+1.23456e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 6, s: `+1.234567e+03`},
		// {mode: libc.FE_DOWNWARD, d: 1234.567800, prec: 7, s: `+1.2345678e+03`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 0, s: `-3e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 1, s: `-2.5e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 2, s: `-2.50e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 3, s: `-2.500e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 4, s: `-2.5000e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 5, s: `-2.50000e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 6, s: `-2.500000e+01`},
		// {mode: libc.FE_DOWNWARD, d: -25.000000, prec: 7, s: `-2.5000000e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 0, s: `+2e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 1, s: `+2.5e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 2, s: `+2.50e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 3, s: `+2.500e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 4, s: `+2.5000e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 5, s: `+2.50000e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 6, s: `+2.500000e+01`},
		// {mode: libc.FE_DOWNWARD, d: 25.000000, prec: 7, s: `+2.5000000e+01`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 0, s: `-1e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 1, s: `-1.2e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 2, s: `-1.23e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 3, s: `-1.234e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 4, s: `-1.2345e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 5, s: `-1.23456e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 6, s: `-1.234567e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567800, prec: 7, s: `-1.2345678e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 0, s: `+2e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 1, s: `+1.3e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 2, s: `+1.24e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 3, s: `+1.235e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 4, s: `+1.2346e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 5, s: `+1.23457e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 6, s: `+1.234568e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567801, prec: 7, s: `+1.2345679e+03`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 0, s: `-2e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 1, s: `-2.5e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 2, s: `-2.50e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 3, s: `-2.500e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 4, s: `-2.5000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 5, s: `-2.50000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 6, s: `-2.500000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000, prec: 7, s: `-2.5000000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 0, s: `+3e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 1, s: `+2.5e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 2, s: `+2.50e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 3, s: `+2.500e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 4, s: `+2.5000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 5, s: `+2.50000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 6, s: `+2.500000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000, prec: 7, s: `+2.5000000e+01`},
	} {
		_snprintf_with_rounding_mode(tls, test.mode, sp, 128, 0, test.prec, test.d)
		s := libc.GoString(sp)
		t.Logf("%2d: test.mode=%v test.prec=%v test.d=%v test.s=%q s=%q", itest, test.mode, test.prec, test.d, test.s, s)
		switch {
		case s != test.s:
			t.Logf("FAIL")
			fails = append(fails, fmt.Sprintf("FAIL %2d: test.mode=%v test.prec=%v test.d=%v test.s=%q s=%q", itest, test.mode, test.prec, test.d, test.s, s))
		default:
			ok++
		}
	}
	if len(fails) != 0 {
		for _, v := range fails {
			t.Error(v)
		}
	}
	t.Logf("tests=%v ok=%v fails=%v", ok+len(fails), ok, len(fails))
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
