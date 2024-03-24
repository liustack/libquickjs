// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package snprintf

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"modernc.org/libc"
)

func SnprintfWithRoundingMode(tls *libc.TLS, rounding_mode int32, cs uintptr, n libc.Tsize_t, nfmt, prec int32, d float64) (r int32) {
	// fmt.Printf("rounding_mode=%v nfmt=%v prec=%v d=%.15f\n", rounding_mode, nfmt, prec, d)
	var s string
	switch nfmt {
	case 0:
		switch {
		case prec < 0:
			prec = 0
		case prec > 17:
			prec = 17
		}
		s = fmt.Sprintf("%+.17e", d)
		// fmt.Printf("s=%q\n", s)
		sgn := s[:1]
		s = s[1:]
		dot := strings.Index(s, ".")
		before := s[:dot]
		s = s[dot+1:]
		expX := strings.Index(s, "e")
		after := s[:expX]
		exp := s[expX:]
		// fmt.Printf("sgn=%q before=%q after=%q exp=%q\n", sgn, before, after, exp)
		switch rounding_mode {
		case libc.FE_DOWNWARD:
			switch {
			case d < 0:
				switch {
				case prec == 0:
					i, _ := strconv.ParseUint(before, 10, 64)
					if strings.TrimRight(after, "0") != "" {
						i++
					}
					s = fmt.Sprintf("%s%d%s%s\x00", sgn, i, after[:prec], exp)
				default:
					i, _ := strconv.ParseUint(after[:prec], 10, 64)
					if strings.TrimRight(after[prec:], "0") != "" {
						i++
					}
					s = fmt.Sprintf("%s%s.%d%s\x00", sgn, before, i, exp)
				}
			default:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%s%s%s%s\x00", sgn, before, after[:prec], exp)
				default:
					s = fmt.Sprintf("%s%s.%s%s\x00", sgn, before, after[:prec], exp)
				}
			}
		case libc.FE_UPWARD:
			switch {
			case d < 0:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%s%s%s%s\x00", sgn, before, after[:prec], exp)
				default:
					s = fmt.Sprintf("%s%s.%s%s\x00", sgn, before, after[:prec], exp)
				}
			default:
				switch {
				case prec == 0:
					i, _ := strconv.ParseUint(before, 10, 64)
					if strings.TrimRight(after, "0") != "" {
						i++
					}
					s = fmt.Sprintf("%s%d%s%s\x00", sgn, i, after[:prec], exp)
				default:
					i, _ := strconv.ParseUint(after[:prec], 10, 64)
					if strings.TrimRight(after[prec:], "0") != "" {
						i++
					}
					s = fmt.Sprintf("%s%s.%d%s\x00", sgn, before, i, exp)
				}
			}
		default:
			panic(rounding_mode)
		}
	case 1:
		switch {
		case prec < 0:
			prec = 0
		case prec > 17:
			prec = 17
		}
		s = fmt.Sprintf("%.17f", d)
		// fmt.Printf("s=%q\n", s)
		var sgn string
		if s[0] == '-' {
			sgn = s[:1]
			s = s[1:]
		}
		dot := strings.Index(s, ".")
		before := s[:dot]
		s = s[dot+1:]
		after := s
		// fmt.Printf("sgn=%q before=%q after=%q\n", sgn, before, after)
		switch rounding_mode {
		case libc.FE_DOWNWARD:
			switch {
			case d < 0:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%.*f\x00", prec, d)
				default:
					i, _ := strconv.ParseUint(after[:prec], 10, 64)
					switch {
					case strings.TrimRight(after[prec:], "0") == "":
						s = fmt.Sprintf("%s%s.%s\x00", sgn, before, after[:prec])
					default:
						i++
						s = fmt.Sprintf("%s%s.%d\x00", sgn, before, i)
					}
				}
			default:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%s%s\x00", sgn, before)
				default:
					s = fmt.Sprintf("%s%s.%s\x00", sgn, before, after[:prec])
				}
			}
		case libc.FE_UPWARD:
			switch {
			case d < 0:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%s%s\x00", sgn, before)
				default:
					s = fmt.Sprintf("%s%s.%s\x00", sgn, before, after[:prec])
				}
			default:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%.*f\x00", prec, d)
				default:
					// all_test.go:298: FAIL
					// all_test.go:306: FAIL 51: test.mode=2048 test.prec=1 test.d=1234.5 678901234567 test.s="1234.6" s="151: 1234.6"
					// all_test.go:306: FAIL 52: test.mode=2048 test.prec=2 test.d=1234.56 78901234567 test.s="1234.57" s="151: 1234.57"
					// all_test.go:306: FAIL 53: test.mode=2048 test.prec=3 test.d=1234.567 8901234567 test.s="1234.568" s="151: 1234.568"
					// all_test.go:306: FAIL 54: test.mode=2048 test.prec=4 test.d=1234.5678 901234567 test.s="1234.5679" s="151: 1234.5679"
					// all_test.go:306: FAIL 55: test.mode=2048 test.prec=5 test.d=1234.56789 01234567 test.s="1234.56790" s="151: 1234.56789"
					// all_test.go:306: FAIL 56: test.mode=2048 test.prec=6 test.d=1234.567890 1234567 test.s="1234.567891" s="151: 1234.567890"
					// all_test.go:306: FAIL 57: test.mode=2048 test.prec=7 test.d=1234.5678901 234567 test.s="1234.5678902" s="151: 1234.5678901"
					// all_test.go:306: FAIL 58: test.mode=2048 test.prec=8 test.d=1234.56789012 34567 test.s="1234.56789013" s="151: 1234.56789012"
					// all_test.go:306: FAIL 59: test.mode=2048 test.prec=9 test.d=1234.567890123 4567 test.s="1234.567890124" s="151: 1234.567890123"
					// all_test.go:306: FAIL 71: test.mode=2048 test.prec=1 test.d=25 test.s="25.0" s="151: 25.0"
					// all_test.go:306: FAIL 72: test.mode=2048 test.prec=2 test.d=25 test.s="25.00" s="151: 25.00"
					// all_test.go:306: FAIL 73: test.mode=2048 test.prec=3 test.d=25 test.s="25.000" s="151: 25.000"
					// all_test.go:306: FAIL 74: test.mode=2048 test.prec=4 test.d=25 test.s="25.0000" s="151: 25.0000"
					// all_test.go:306: FAIL 75: test.mode=2048 test.prec=5 test.d=25 test.s="25.00000" s="151: 25.00000"
					// all_test.go:306: FAIL 76: test.mode=2048 test.prec=6 test.d=25 test.s="25.000000" s="151: 25.000000"
					// all_test.go:306: FAIL 77: test.mode=2048 test.prec=7 test.d=25 test.s="25.0000000" s="151: 25.0000000"
					// all_test.go:306: FAIL 78: test.mode=2048 test.prec=8 test.d=25 test.s="25.00000000" s="151: 25.00000000"
					// all_test.go:306: FAIL 79: test.mode=2048 test.prec=9 test.d=25 test.s="25.000000000" s="151: 25.000000000"
					i, _ := strconv.ParseUint(after[:prec], 10, 64)
					switch {
					case strings.TrimRight(after[prec:], "0") == "":
						s = fmt.Sprintf("%s%s.%s\x00", sgn, before, after[:prec])
					default:
						i++
						s = fmt.Sprintf("%s%s.%d\x00", sgn, before, i)
					}
				}
			}
		default:
			panic(rounding_mode)
		}
	default:
		panic(nfmt)
	}
	r = int32(len(s)) - 1
	if libc.Tsize_t(len(s)) > n {
		s = s[:n]
	}
	for i := 0; i < len(s); i++ {
		*(*byte)(unsafe.Pointer(cs)) = s[i]
		cs++
	}
	return r
}
