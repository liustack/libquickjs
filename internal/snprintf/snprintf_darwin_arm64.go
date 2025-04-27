// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package snprintf

import (
	"modernc.org/libc/sys/types"
)

const (
	feDownward = 0x00800000
	feUpward   = 0x00400000
)

type sizeT = types.Size_t
