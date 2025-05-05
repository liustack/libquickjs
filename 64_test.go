// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build !(386 || arm)

package libquickjs // import "modernc.org/libquickjs"

func tag(v TJSValue) int32 {
	return int32(v.Ftag)
}
