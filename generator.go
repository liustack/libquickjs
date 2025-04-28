// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	ccgo "modernc.org/ccgo/v4/lib"
)

func fail(rc int, msg string, args ...any) {
	fmt.Fprintln(os.Stderr, strings.TrimSpace(fmt.Sprintf(msg, args...)))
	panic(rc)
}

func main() {
	if ccgo.IsExecEnv() {
		if err := ccgo.NewTask(goos, goarch, os.Args, os.Stdout, os.Stderr, nil).Main(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
}
