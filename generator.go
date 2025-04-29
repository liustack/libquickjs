// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	cc "modernc.org/cc/v4"
	ccgo "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
)

const (
	qjs       = "quickjs-" + qjsCommit
	qjsCommit = "19abf1888db5884a5758036ff6e7fa2b340acedc"
	qjsDiff   = "quickjs.diff"
	qjsZip    = qjsCommit + ".zip"
	tmp       = "tmp"
)

var (
	args = []string{
		"--prefix-enumerator=E",
		"--prefix-external=x_",
		"--prefix-field=F",
		"--prefix-macro=M",
		"--prefix-static-internal=_",
		"--prefix-static-none=_",
		"--prefix-tagged-enum=_",
		"--prefix-tagged-struct=T",
		"--prefix-tagged-union=T",
		"--prefix-typename=T",
		"--prefix-undefined=_",
		"-DNDEBUG",
		"-UHAVE_CLOSEFROM",
		"-U__SIZEOF_INT128__",
		"-eval-all-macros",
		"-extended-errors",
		"-ignore-unsupported-alignment",
	}
	dev    = os.Getenv("GO_GENERATE_DEV") != ""
	gmake  = "make"
	goarch = runtime.GOARCH
	goos   = runtime.GOOS
	gsed   = "sed"
	j      = strconv.Itoa(runtime.GOMAXPROCS(-1))
	target = fmt.Sprintf("%s/%s", goos, goarch)
)

func fail(rc int, msg string, args ...any) {
	fmt.Fprintln(os.Stderr, strings.TrimSpace(fmt.Sprintf("FAIL: "+msg, args...)))
	panic(rc)
}

func main() {
	if ccgo.IsExecEnv() {
		if err := ccgo.NewTask(goos, goarch, os.Args, os.Stdout, os.Stderr, nil).Main(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		fail(1, "%v", err)
	}

	if wd, err = filepath.Abs(wd); err != nil {
		fail(1, "%v", err)
	}

	if s := cc.LongDouble64Flag(goos, goarch); s != "" {
		args = append(args, s)
	}
	if dev {
		args = append(args,
			"-absolute-paths",
			"-keep-object-files",
			"-positions",
		)
	}

	removeAll(tmp)
	mkdirAll(tmp)
	unzip(qjsZip, tmp)
	quickjs := filepath.Join(tmp, qjs)
	util.MustCopyFile(true, filepath.Join(quickjs, qjsDiff), filepath.Join(wd, "internal", "patch", qjsDiff), nil)
	git("-C", quickjs, "init")
	git("-C", quickjs, "apply", qjsDiff)
	ccExec(append(args, "-exec", "make", "-C", quickjs, "libquickjs.a", "qjs", "run-test262"))
	result := fmt.Sprintf("ccgo_%s_%s.go", goos, goarch)
	ccMain(append(args,
		"--package-name", "libquickjs",
		"-ignore-link-errors",
		"-o", result,
		filepath.Join(quickjs, "libquickjs.a"),
	))
	util.MustShell(true, nil, gsed, "-i", `s/\<T__\([a-zA-Z0-9][a-zA-Z0-9_]\+\)/t__\1/g`, result)
	util.MustShell(true, nil, gsed, "-i", `s/\<x_\([a-zA-Z0-9_][a-zA-Z0-9_]\+\)/X\1/g`, result)
	util.MustShell(true, nil, gsed, "-i", `s/\<X__JS_\([a-zA-Z0-9][a-zA-Z0-9_]\+\)/x__JS_\1/g`, result)
	util.MustCopyFile(true, filepath.Join("gqjs", result), filepath.Join(quickjs, "qjs.go"), nil)
	util.MustCopyFile(true, filepath.Join("internal", "run-test262", result), filepath.Join(quickjs, "run-test262.go"), nil)
	util.MustCopyDir(true, filepath.Join("testdata", "tests"), filepath.Join(quickjs, "tests"), nil)
}

func git(args ...string) {
	if out, err := util.Shell(nil, "git", args...); err != nil {
		fail(1, "err=%v out=%s", err, out)
	}
}

func ccMain(args []string) {
	if err := ccgo.NewTask(goos, goarch, append([]string{"ccgo"}, args...), os.Stdout, os.Stderr, nil).Main(); err != nil {
		fail(1, "%v", err)
	}
}

func ccExec(args []string) {
	if err := ccgo.NewTask(goos, goarch, append([]string{"ccgo"}, args...), os.Stdout, os.Stderr, nil).Exec(); err != nil {
		fail(1, "%v", err)
	}
}

func mkdirAll(path string) {
	if err := os.MkdirAll(path, 0770); err != nil {
		fail(1, "%v", err)
	}
}

func unzip(src, dest string) {
	if out, err := util.Shell(nil, "unzip", "-q", src, "-d", dest); err != nil {
		fail(1, "err=%v out=%s", err, out)
	}
}

func removeAll(path string) {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		fail(1, "%v", err)
	}
}
