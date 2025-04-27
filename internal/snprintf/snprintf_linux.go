// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package snprintf

import (
	"modernc.org/libc"
)

const (
	feDownward = libc.FE_DOWNWARD
	feUpward   = libc.FE_UPWARD
)

type sizeT = libc.Tsize_t
