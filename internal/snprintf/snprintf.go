// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package snprintf

import (
	"fmt"
	"math"
	"unsafe"

	"modernc.org/libc"
)

func SnprintfWithRoundingMode(tls *libc.TLS, rounding_mode int32, s uintptr, n libc.Tsize_t, nfmt, width int32, d float64) (r int32) {
	var gs string
	switch nfmt {
	case 0:
		switch rounding_mode {
		case libc.FE_DOWNWARD:
			d -= .5 * math.Pow10(int(width)-2)
		case libc.FE_UPWARD:
			d += .5 * math.Pow10(int(width)-2)
		default:
			panic(rounding_mode)
		}
		gs = fmt.Sprintf("%+.*e\x00", width, d)
	case 1:
		switch rounding_mode {
		case libc.FE_DOWNWARD:
			d -= .5 * math.Pow10(-int(width))
		case libc.FE_UPWARD:
			d += .5 * math.Pow10(-int(width))
		default:
			panic(rounding_mode)
		}
		gs = fmt.Sprintf("%.*f\x00", width, d)
	default:
		panic(nfmt)
	}
	r = int32(len(gs)) - 1
	if libc.Tsize_t(len(gs)) > n {
		gs = gs[:n]
	}
	for i := 0; i < len(gs); i++ {
		*(*byte)(unsafe.Pointer(s)) = gs[i]
		s++
	}
	return r
}
