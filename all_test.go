// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package libquickjs // import "modernc.org/libquickjs"

import (
	"os"
	"runtime"
	"testing"

	_ "modernc.org/ccgo/v4/lib"
	_ "modernc.org/fileutil/ccgo"
	_ "modernc.org/libc"
)

var (
	goos   = runtime.GOOS
	goarch = runtime.GOARCH
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	t.Log("TODO")
}
