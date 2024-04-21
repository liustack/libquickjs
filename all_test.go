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

const (
	zip262 = "837def66ac1700ce398df4a3b4eaeea915aa1df5.zip"
)

var (
	goos   = runtime.GOOS
	goarch = runtime.GOARCH
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var str [128]byte

func TestSnprintfWithRoundingMode0(t *testing.T) {
	// #include <stdio.h>
	// #include <fenv.h>
	//
	// char str[128];
	//
	// int main() {
	// 	char *names[] = {"FE_DOWNWARD", "FE_UPWARD"};
	// 	int modes[] = {FE_DOWNWARD, FE_UPWARD};
	// 	double ds[] = {1234.5678901234567, 25};
	// 	double signs[] = {-1, 1};
	// 	for (int i = 0; i < 2; i++) {
	// 		int mode = modes[i];
	//              fesetround(mode);
	// 		char *name = names[i];
	// 		for (int j = 0; j < 2; j++) {
	// 			double d0 = ds[j];
	// 			for (int k = 0; k < 2; k++) {
	// 				double d = d0*signs[k];
	// 				for (int prec = 0; prec < 10; prec++) {
	// 					snprintf(str, 128, "%+.*e", prec, d);
	// 					printf("mode=%s d=%.15f prec=%d s=`%s`\n", name, d, prec, str);
	// 				}
	// 			}
	// 		}
	// 	}
	// }
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
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 0, s: `-2e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 1, s: `-1.3e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 2, s: `-1.24e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 3, s: `-1.235e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 4, s: `-1.2346e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 5, s: `-1.23457e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 6, s: `-1.234568e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 7, s: `-1.2345679e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 8, s: `-1.23456790e+03`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 9, s: `-1.234567891e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 0, s: `+1e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 1, s: `+1.2e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 2, s: `+1.23e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 3, s: `+1.234e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 4, s: `+1.2345e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 5, s: `+1.23456e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 6, s: `+1.234567e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 7, s: `+1.2345678e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 8, s: `+1.23456789e+03`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 9, s: `+1.234567890e+03`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 0, s: `-3e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 1, s: `-2.5e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 2, s: `-2.50e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 3, s: `-2.500e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 4, s: `-2.5000e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 5, s: `-2.50000e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 6, s: `-2.500000e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 7, s: `-2.5000000e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 8, s: `-2.50000000e+01`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 9, s: `-2.500000000e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 0, s: `+2e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 1, s: `+2.5e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 2, s: `+2.50e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 3, s: `+2.500e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 4, s: `+2.5000e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 5, s: `+2.50000e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 6, s: `+2.500000e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 7, s: `+2.5000000e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 8, s: `+2.50000000e+01`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 9, s: `+2.500000000e+01`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 0, s: `-1e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 1, s: `-1.2e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 2, s: `-1.23e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 3, s: `-1.234e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 4, s: `-1.2345e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 5, s: `-1.23456e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 6, s: `-1.234567e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 7, s: `-1.2345678e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 8, s: `-1.23456789e+03`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 9, s: `-1.234567890e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 0, s: `+2e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 1, s: `+1.3e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 2, s: `+1.24e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 3, s: `+1.235e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 4, s: `+1.2346e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 5, s: `+1.23457e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 6, s: `+1.234568e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 7, s: `+1.2345679e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 8, s: `+1.23456790e+03`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 9, s: `+1.234567891e+03`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 0, s: `-2e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 1, s: `-2.5e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 2, s: `-2.50e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 3, s: `-2.500e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 4, s: `-2.5000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 5, s: `-2.50000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 6, s: `-2.500000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 7, s: `-2.5000000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 8, s: `-2.50000000e+01`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 9, s: `-2.500000000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 0, s: `+3e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 1, s: `+2.5e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 2, s: `+2.50e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 3, s: `+2.500e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 4, s: `+2.5000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 5, s: `+2.50000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 6, s: `+2.500000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 7, s: `+2.5000000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 8, s: `+2.50000000e+01`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 9, s: `+2.500000000e+01`},
	} {
		_snprintf_with_rounding_mode(tls, test.mode, sp, 128, 0, test.prec, test.d)
		s := libc.GoString(sp)
		// t.Logf("%2d: test.mode=%v test.prec=%v test.d=%.15f test.s=%q s=%q", itest, test.mode, test.prec, test.d, test.s, s)
		switch {
		case s != test.s:
			t.Logf("FAIL")
			fails = append(fails, fmt.Sprintf("FAIL %2d: test.mode=%v test.prec=%v test.d=%.15f test.s=%q s=%q", itest, test.mode, test.prec, test.d, test.s, s))
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

func TestSnprintfWithRoundingMode1(t *testing.T) {
	// #include <stdio.h>
	// #include <fenv.h>
	//
	// char str[128];
	//
	// int main() {
	// 	char *names[] = {"FE_DOWNWARD", "FE_UPWARD"};
	// 	int modes[] = {FE_DOWNWARD, FE_UPWARD};
	// 	double ds[] = {1234.5678901234567, 25};
	// 	double signs[] = {-1, 1};
	// 	for (int i = 0; i < 2; i++) {
	// 		int mode = modes[i];
	// 		fesetround(mode);
	// 		char *name = names[i];
	// 		for (int j = 0; j < 2; j++) {
	// 			double d0 = ds[j];
	// 			for (int k = 0; k < 2; k++) {
	// 				double d = d0*signs[k];
	// 				for (int prec = 0; prec < 10; prec++) {
	// 					snprintf(str, 128, "%.*f", prec, d);
	// 					printf("mode=%s d=%f prec=%d s=`%s`\n", name, d, prec, str);
	// 				}
	// 			}
	// 		}
	// 	}
	// }
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
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 0, s: `-1235`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 1, s: `-1234.6`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 2, s: `-1234.57`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 3, s: `-1234.568`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 4, s: `-1234.5679`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 5, s: `-1234.56790`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 6, s: `-1234.567891`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 7, s: `-1234.5678902`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 8, s: `-1234.56789013`},
		{mode: libc.FE_DOWNWARD, d: -1234.567890123456664, prec: 9, s: `-1234.567890124`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 0, s: `1234`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 1, s: `1234.5`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 2, s: `1234.56`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 3, s: `1234.567`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 4, s: `1234.5678`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 5, s: `1234.56789`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 6, s: `1234.567890`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 7, s: `1234.5678901`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 8, s: `1234.56789012`},
		{mode: libc.FE_DOWNWARD, d: 1234.567890123456663, prec: 9, s: `1234.567890123`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 0, s: `-25`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 1, s: `-25.0`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 2, s: `-25.00`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 3, s: `-25.000`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 4, s: `-25.0000`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 5, s: `-25.00000`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 6, s: `-25.000000`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 7, s: `-25.0000000`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 8, s: `-25.00000000`},
		{mode: libc.FE_DOWNWARD, d: -25.000000000000000, prec: 9, s: `-25.000000000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 0, s: `25`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 1, s: `25.0`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 2, s: `25.00`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 3, s: `25.000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 4, s: `25.0000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 5, s: `25.00000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 6, s: `25.000000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 7, s: `25.0000000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 8, s: `25.00000000`},
		{mode: libc.FE_DOWNWARD, d: 25.000000000000000, prec: 9, s: `25.000000000`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 0, s: `-1234`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 1, s: `-1234.5`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 2, s: `-1234.56`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 3, s: `-1234.567`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 4, s: `-1234.5678`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 5, s: `-1234.56789`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 6, s: `-1234.567890`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 7, s: `-1234.5678901`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 8, s: `-1234.56789012`},
		{mode: libc.FE_UPWARD, d: -1234.567890123456663, prec: 9, s: `-1234.567890123`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 0, s: `1235`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 1, s: `1234.6`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 2, s: `1234.57`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 3, s: `1234.568`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 4, s: `1234.5679`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 5, s: `1234.56790`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 6, s: `1234.567891`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 7, s: `1234.5678902`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 8, s: `1234.56789013`},
		{mode: libc.FE_UPWARD, d: 1234.567890123456664, prec: 9, s: `1234.567890124`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 0, s: `-25`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 1, s: `-25.0`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 2, s: `-25.00`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 3, s: `-25.000`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 4, s: `-25.0000`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 5, s: `-25.00000`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 6, s: `-25.000000`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 7, s: `-25.0000000`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 8, s: `-25.00000000`},
		{mode: libc.FE_UPWARD, d: -25.000000000000000, prec: 9, s: `-25.000000000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 0, s: `25`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 1, s: `25.0`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 2, s: `25.00`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 3, s: `25.000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 4, s: `25.0000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 5, s: `25.00000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 6, s: `25.000000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 7, s: `25.0000000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 8, s: `25.00000000`},
		{mode: libc.FE_UPWARD, d: 25.000000000000000, prec: 9, s: `25.000000000`},
	} {
		_snprintf_with_rounding_mode(tls, test.mode, sp, 128, 1, test.prec, test.d)
		s := libc.GoString(sp)
		// t.Logf("%2d: test.mode=%v test.prec=%v test.d=%v test.s=%q s=%q", itest, test.mode, test.prec, test.d, test.s, s)
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
	bin := filepath.Join(tmpdir, "qjs")
	if !filepath.IsAbs(bin) {
		bin = "./" + bin
	}
	shell(t, "go", "build", "-o", bin, "./"+filepath.Join("internal", "qjs"))
	tests := filepath.Join("internal", "tests")
	for _, test := range [][]string{
		{filepath.Join(tests, "test_closure.js")},
		{filepath.Join(tests, "test_language.js")},
		{"--std", filepath.Join(tests, "test_builtin.js")},
		{filepath.Join(tests, "test_loop.js")},
		{filepath.Join(tests, "test_bignum.js")},
		{filepath.Join(tests, "test_std.js")}, //TODO popen, os.exec
		{filepath.Join(tests, "test_worker.js")},
		//TODO {"--bignum", filepath.Join(tests, "test_bjson.js")}, // could not load module filename 'internal/tests/bjson.so' as shared library
		//TODO {filepath.Join(tests, "test_bjson.js")}, // // could not load module filename 'internal/tests/bjson.so' as shared library
		{"--bignum", filepath.Join(tests, "test_op_overloading.js")},
		{"--bignum", filepath.Join(tests, "test_bigfloat.js")},
		{"--qjscalc", filepath.Join(tests, "test_qjscalc.js")},
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

func shell(t *testing.T, cmd string, args ...string) (out []byte, err error) {
	if out, err = util.Shell(nil, cmd, args...); err != nil && t != nil {
		t.Fatalf("FAIL err=%v\n%s", err, out)
	}

	return out, err
}

func copyFile(t *testing.T, dest, src string) {
	if _, err := util.CopyFile(dest, src, nil); err != nil {
		t.Fatalf("FAIL err=%v", err)
	}
}

func inDir(t *testing.T, dir string, f func() error) {
	if err := util.InDir(dir, f); err != nil {
		t.Fatalf("FAIL err=%v", err)
	}
}

func Test262(t *testing.T) {
	if _, err := os.Stat(zip262); err != nil {
		t.Fatalf("%s: to download the archive issue '$ make download'", err)
	}

	tempdir := os.Getenv("TEST_DIR")
	if tempdir == "" {
		tempdir = t.TempDir()
	}
	t.Logf("tempdir=%s", tempdir)
	shell(t, "unzip", "-q", "-d", tempdir, zip262)
	testdir := filepath.Join(tempdir, "test262")
	bin := filepath.Join(tempdir, "run-test262")
	if !filepath.IsAbs(bin) {
		bin = "./" + bin
	}
	shell(t, "go", "build", "-o", bin, "./"+filepath.Join("internal", "run-test262"))
	copyFile(t, filepath.Join(tempdir, "test262.conf"), filepath.Join("internal", "tests", "test262.conf"))
	copyFile(t, filepath.Join(tempdir, "test262_errors.txt"), filepath.Join("internal", "tests", "test262_errors.txt"))
	shell(t, "mv", filepath.Join(tempdir, "test262-"+zip262[:len(zip262)-len(".zip")]), testdir)
	b, err := os.ReadFile(filepath.Join("internal", "run-test262", "exclude"))
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if v = strings.TrimSpace(v); v != "" && !strings.HasPrefix(v, "#") {
			shell(t, "sh", "-c", fmt.Sprintf("rm -rfv %s", filepath.Join(tempdir, v)))
		}
	}
	var resultLine string
	inDir(t, tempdir, func() error {
		out, err := shell(nil, bin, "-c", "test262.conf", "-a")
		if err != nil {
			t.Logf("run-test262 FAIL err=%v", err)
		}

		s := strings.TrimSpace(string(out))
		x := strings.Index(s, "\nResult: ")
		if x < 0 {
			t.Fatal("cannot find test results")
		}

		s = s[x+1:]
		if x = strings.IndexByte(s, '\n'); x > 0 {
			s = s[:x]
		}
		resultLine = strings.TrimSpace(s) + "\n"
		return nil
	})
	if b, err = os.ReadFile(filepath.Join("internal", "run-test262", "results")); err != nil {
		t.Fatal(err)
	}

	if g, e := resultLine, string(b); g != e {
		b, _ = os.ReadFile(filepath.Join(testdir, "test262_report.txt"))
		t.Logf("\n%s", b)
		t.Fatalf("got %q expected %q", g, e)
	} else {
		t.Logf("OK, got %q expected %q", g, e)
	}
}
