// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package snprintf

import (
	"fmt"
	"math"
	"unsafe"
	"strconv"
	"strings"

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
		s = fmt.Sprintf("%+.17e\x00", d)
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
					ibefore, _ := strconv.ParseUint(before, 10, 64)
					if strings.TrimRight(after, "0") != "" {
						ibefore++
					}
					s = fmt.Sprintf("%s%d%s%s", sgn, ibefore, after[:prec], exp)
				default:
					iafter, _ := strconv.ParseUint(after[:prec], 10, 64)
					if strings.TrimRight(after[prec:], "0") != "" {
						iafter++
					}
					s = fmt.Sprintf("%s%s.%d%s", sgn, before, iafter, exp)
				}
			default:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%s%s%s%s", sgn, before, after[:prec], exp)
				default:
					s = fmt.Sprintf("%s%s.%s%s", sgn, before, after[:prec], exp)
				}
			}
		case libc.FE_UPWARD:
			switch {
			case d < 0:
				switch {
				case prec == 0:
					s = fmt.Sprintf("%s%s%s%s", sgn, before, after[:prec], exp)
				default:
					s = fmt.Sprintf("%s%s.%s%s", sgn, before, after[:prec], exp)
				}
			default:
				switch {
				case prec == 0:
					ibefore, _ := strconv.ParseUint(before, 10, 64)
					if strings.TrimRight(after, "0") != "" {
						ibefore++
					}
					s = fmt.Sprintf("%s%d%s%s", sgn, ibefore, after[:prec], exp)
				default:
					iafter, _ := strconv.ParseUint(after[:prec], 10, 64)
					if strings.TrimRight(after[prec:], "0") != "" {
						iafter++
					}
					s = fmt.Sprintf("%s%s.%d%s", sgn, before, iafter, exp)
				}
			}
		default:
			panic(rounding_mode)
		}
	case 1:
		switch rounding_mode {
		case libc.FE_DOWNWARD:
			d -= .5 * math.Pow10(-int(prec))
		case libc.FE_UPWARD:
			d += .5 * math.Pow10(-int(prec))
		default:
			panic(rounding_mode)
		}
		s = fmt.Sprintf("%.*f\x00", prec, d)
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
