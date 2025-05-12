// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package libquickjs // import "modernc.org/libquickjs"

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"modernc.org/libc"
)

// 2025-05-12 C						   Go
//
//                                                         goos: linux
//                                                         goarch: amd64
//                                                         pkg: modernc.org/libquickjs
//                                                         cpu: Intel(R) Xeon(R) CPU E5-1650 v2 @ 3.50GHz
//                                                         BenchmarkMicro
//	                  TEST          N TIME (ns)                          TEST          N TIME (ns)
//	            empty_loop     200000     19.77                    empty_loop      50000     82.43
//	       empty_down_loop     100000     25.22               empty_down_loop      20000    112.82
//	      empty_down_loop2     100000     27.51              empty_down_loop2      20000    141.99
//	         empty_do_loop     100000     24.64                 empty_do_loop      20000    108.44
//	              date_now       5000    902.15                      date_now       2000   1069.28
//	            date_parse        500   1220.14                    date_parse        100   4520.99
//	             prop_read      50000     22.82                     prop_read       5000    111.92
//	            prop_write      50000     19.94                    prop_write       5000    103.19
//	           prop_update      20000     29.57                   prop_update       5000    184.86
//	           prop_create       2000     86.18                   prop_create        500    270.46
//	            prop_clone       1000    108.88                    prop_clone        500    395.47
//	           prop_delete       1000    170.97                   prop_delete        200    617.05
//	            array_read      10000     23.02                    array_read       5000     79.65
//	           array_write      10000     21.08                   array_write       5000     60.57
//	     array_prop_create        100     27.36             array_prop_create         20    124.39
//	           array_slice        100     28.36                   array_slice         20    183.53
//	     array_length_decr         50     68.87             array_length_decr          5    417.92
//	array_hole_length_decr         50     94.98        array_hole_length_decr          5    464.69
//	            array_push         50    104.32                    array_push         10    595.57
//	             array_pop         50    115.34                     array_pop         10    723.05
//	      typed_array_read      10000     21.57              typed_array_read       5000     77.96
//	     typed_array_write      10000     22.23             typed_array_write       5000     61.28
//	           global_read      50000     20.11                   global_read       5000    115.40
//	          global_write      50000     24.10                  global_write       5000    122.20
//	   global_write_strict      50000     24.29           global_write_strict       5000    122.00
//	        local_destruct       1000    294.73                local_destruct        200   1194.36
//	       global_destruct       2000    141.37               global_destruct        500    813.64
//	global_destruct_strict       2000    141.24        global_destruct_strict        500    813.83
//	      global_func_call      20000     45.74              global_func_call       2000    285.54
//	             func_call      20000     38.37                     func_call       5000    222.60
//	     func_closure_call      20000     38.94             func_closure_call       5000    219.62
//	             int_arith        200     18.73                     int_arith         50     89.66
//	           float_arith        100     27.89                   float_arith         20    107.60
//	        map_set_string         10    393.73                map_set_string          2   1890.97
//	           map_set_int         10    203.35                   map_set_int          5    892.55
//	        map_set_bigint         10    261.84                map_set_bigint          2   1502.86
//	            map_delete          5    414.65                    map_delete          1   2024.65
//	          weak_map_set         20    162.75                  weak_map_set          1   2042.25
//	       weak_map_delete          5    448.91               weak_map_delete          1   2193.88
//	             array_for       1000     23.52                     array_for        200    107.32
//	          array_for_in        200    104.38                  array_for_in         50    487.78
//	          array_for_of       1000     37.41                  array_for_of        100    237.09
//	              math_min         50     60.36                      math_min         10    353.19
//	          regexp_ascii          2   1042.39                  regexp_ascii          1   3550.83
//	          regexp_utf16          2   1072.32                  regexp_utf16          1   3606.71
//	         string_build1        100     34.19                 string_build1         20    128.84
//	        string_build1x        100     34.16                string_build1x         20    128.76
//	        string_build2c         50     41.94                string_build2c         20    127.95
//	         string_build2         50     71.66                 string_build2         10    317.67
//	         string_build3         50     86.03                 string_build3         10    370.35
//	         string_build4         50     81.81                 string_build4         10    368.26
//	   string_build_large1          1    112.64           string_build_large1          1    390.19
//	   string_build_large2          1    109.75           string_build_large2          1    434.89
//	         int_to_string      10000     85.69                 int_to_string       5000    320.44
//	          int_toString      10000    100.82                  int_toString       2000    476.21
//	       float_to_string       2000    477.14               float_to_string       1000    987.52
//	        float_toString       2000    493.05                float_toString       1000   1234.06
//	         float_toFixed       5000    235.89                 float_toFixed       1000    874.72
//	     float_toPrecision       5000    251.85             float_toPrecision       1000    946.73
//	   float_toExponential       5000    249.45           float_toExponential       1000    950.92
//	         string_to_int      20000    153.47                 string_to_int       5000    429.61
//	       string_to_float      10000    206.96               string_to_float       5000    516.89
//	        bigint32_arith        100     39.07                bigint32_arith         10    287.02
//	        bigint64_arith         50     65.93                bigint64_arith          5    445.02
//	       bigint256_arith         10    202.26               bigint256_arith          5    712.35
//	            sort_bench          1     24.21                    sort_bench          1     81.02
//	                 total             11384.01                         total             44033.46

func BenchmarkMicro(b *testing.B) {
	buf, err := os.ReadFile(filepath.Join("testdata", "microbench.js"))
	if err != nil {
		b.Fatal(err)
	}

	tls := libc.NewTLS()

	defer tls.Close()

	rt := XJS_NewRuntime(tls)

	defer XJS_FreeRuntime(tls, rt)

	ctx := XJS_NewContext(tls, rt)

	defer XJS_FreeContext(tls, ctx)

	XJS_SetStripInfo(tls, rt, 0)

	Xjs_std_add_helpers(tls, ctx, -1, 0)
	js := string(buf)
	cs, err := libc.CString(js)
	if err != nil {
		b.Fatalf("FAIL: err=%v", err)
	}

	defer libc.Xfree(tls, cs)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := XJS_Eval(tls, ctx, cs, Size_t(len(js)), uintptr(unsafe.Pointer(&evalFN)), MJS_EVAL_TYPE_GLOBAL)
		tg := tag(v)
		XFreeValue(tls, ctx, v)
		if tg == EJS_TAG_EXCEPTION {
			b.Fatal(exception(tls, ctx))
		}
	}
}

func exception(tls *libc.TLS, ctx uintptr) (err error) {
	e := XJS_GetException(tls, ctx)

	defer XFreeValue(tls, ctx, e)

	p := XToCString(tls, ctx, e)

	defer XJS_FreeCString(tls, ctx, p)

	return fmt.Errorf("%s", libc.GoString(p))
}
