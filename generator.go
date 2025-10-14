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
	qjsCommit     = "fa628f8c523ecac8ce560c081411e91fcaba2d20" // release 2025-09-13
	qjsDiff       = "quickjs.diff"
	qjsUnzip      = "quickjs-" + qjsCommit
	qjsZip        = qjsCommit + ".zip"
	test262Commit = "a5e69a1534de88d1eb29b76657d84c8541b72df7" // 2025-09-10
	test262Diff   = "test262.diff"
	test262Dir    = "test262"
	test262Unzip  = "test262-" + test262Commit
	test262Zip    = test262Commit + ".zip"
	tmp           = "tmp"
)

var (
	ccArgs = []string{
		"--hot-switch=^JS_CallInternal$",
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
		"-hide", "js_check_stack_overflow",
		"-ignore-link-errors",
		"-ignore-unsupported-alignment",
	}
	dev    = os.Getenv("GO_GENERATE_DEV") != ""
	gmake  = "make"
	goarch = env("TARGET_GOARCH", env("GOARCH", runtime.GOARCH))
	goos   = env("TARGET_GOOS", env("GOOS", runtime.GOOS))
	gsed   = "sed"
	j      = strconv.Itoa(runtime.GOMAXPROCS(-1))
	target = fmt.Sprintf("%s/%s", goos, goarch)
	w64    = os.Getenv("GO_GENERATE_W64") != ""
)

func env(name, deflt string) (r string) {
	r = deflt
	if s := os.Getenv(name); s != "" {
		r = s
	}
	return r
}

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

	switch {
	case w64:
		goos = "windows"
		goarch = "amd64"
		os.Setenv("TARGET_GOOS", goos)
		os.Setenv("TARGET_GOARCH", goarch)
		xgcc := "/home/jnml/bin/llvm/bin/x86_64-w64-mingw32-gcc"
		os.Setenv("CC", xgcc)
		os.Setenv("CCGO_CC", xgcc)
		os.Setenv("CCGO_GCC", xgcc)
		os.Setenv("CCGO_CLANG", xgcc)
		os.Setenv("CCGO_CPP", xgcc)
		ccArgs = append(ccArgs,
			"--cpp", xgcc,
			"--goarch", goarch,
			"--goos", goos,
			"-D__builtin_bit_cast(T, V)=__builtin_bit_cast()",
			"-D_FILE_OFFSET_BITS",
			"-D_FILE_OFFSET_BITS=64",
			"-build-lines", "//go:build windows && (amd64 || arm64)\n",
			"-map", "gcc=/home/jnml/bin/llvm/bin/x86_64-w64-mingw32-gcc,ar=/home/jnml/bin/llvm/bin/x86_64-w64-mingw32-ar",
		)
	}
	if s := cc.LongDouble64Flag(goos, goarch); s != "" {
		ccArgs = append(ccArgs, s)
	}
	if dev {
		ccArgs = append(ccArgs,
			"-absolute-paths",
			"-keep-object-files",
			"-positions",
		)
	}

	switch target {
	case "darwin/arm64", "darwin/amd64":
		gsed = "gsed"
	case "freebsd/arm64", "freebsd/amd64":
		gmake = "gmake"
		gsed = "gsed"
	}

	removeAll(tmp)
	mkdirAll(tmp)
	unzip(qjsZip, tmp)
	quickjs := filepath.Join(tmp, qjsUnzip)
	unzip(test262Zip, quickjs)
	if err := os.Rename(filepath.Join(quickjs, test262Unzip), filepath.Join(quickjs, test262Dir)); err != nil {
		fail(1, "%v", err)
	}

	test262 := filepath.Join(quickjs, test262Dir)
	copyFile(filepath.Join("internal", "patch", test262Diff), filepath.Join(quickjs, "tests", "test262.patch"))
	copyFile(filepath.Join(test262, "test262.patch"), filepath.Join(quickjs, "tests", "test262.patch"))
	git("-C", test262, "init")
	git("-C", test262, "apply", "-v", "test262.patch")
	copyFile("LICENSE-LIBQUICKJS", filepath.Join(quickjs, "LICENSE"))
	copyFile(filepath.Join("testdata", "test262.conf"), filepath.Join(quickjs, "test262.conf"))
	copyFile(filepath.Join(quickjs, qjsDiff), filepath.Join(wd, "internal", "patch", qjsDiff))
	git("-C", quickjs, "init")
	git("-C", quickjs, "apply", "-v", qjsDiff)
	if w64 {
		shell(gsed, "-i", `s/__SIZEOF_INT128__/__nothing__/`, filepath.Join(quickjs, "quickjs.h"))
	}
	switch target {
	case "freebsd/amd64", "freebsd/arm64":
		ccExec(append(ccArgs,
			"-std=c99",
			"-D__sync_fetch_and_add(p, v, ...)=__atomic_fetch_add((p), (v), 0)",
			"-D__sync_fetch_and_and(p, v, ...)=__atomic_fetch_and((p), (v), 0)",
			"-D__sync_fetch_and_or(p, v, ...)=__atomic_fetch_or((p), (v), 0)",
			"-D__sync_fetch_and_sub(p, v, ...)=__atomic_fetch_sub((p), (v), 0)",
			"-D__sync_fetch_and_xor(p, v, ...)=__atomic_fetch_xor((p), (v), 0)",
			"-D__sync_lock_test_and_set(p, v, ...)=__c11_atomic_exchange((p), (v), 0)",
			"-exec", gmake, "-C", quickjs, "libquickjs.a", "run-test262"))
	case "linux/arm":
		ccExec(append(ccArgs, "-exec", gmake, "LIBS=-lm -latomic", "-C", quickjs, "libquickjs.a", "run-test262"))
	default:
		switch {
		case w64:
			ccExec(append(ccArgs, "-exec", gmake, "CONFIG_WIN32=y", "-C", quickjs, "libquickjs.a", "run-test262.exe"))
		default:
			ccExec(append(ccArgs, "-exec", gmake, "-C", quickjs, "libquickjs.a", "run-test262"))
		}
	}
	copyFile("test262_errors.txt", filepath.Join(quickjs, "test262_errors.txt"))
	result := fmt.Sprintf("ccgo_%s_%s.go", goos, goarch)
	switch {
	case w64:
		result = "ccgo_windows.go"
	}
	ccMain(append(ccArgs,
		"--package-name", "libquickjs",
		"-o", result,
		filepath.Join(quickjs, "libquickjs.a"),
	))
	shell(gsed, "-i", `s/\<T__\([a-zA-Z0-9][a-zA-Z0-9_]\+\)/t__\1/g`, result)
	shell(gsed, "-i", `s/\<x_\([a-zA-Z0-9_][a-zA-Z0-9_]\+\)/X\1/g`, result)
	shell(gsed, "-i", `s/\<X__JS_\([a-zA-Z0-9][a-zA-Z0-9_]\+\)/x__JS_\1/g`, result)
	switch {
	case w64:
		copyFile(filepath.Join("internal", "run-test262", result), filepath.Join(quickjs, "run-test262.exe.go"))
	default:
		copyFile(filepath.Join("internal", "run-test262", result), filepath.Join(quickjs, "run-test262.go"))
	}
	util.MustCopyDir(true, filepath.Join("testdata", "tests"), filepath.Join(quickjs, "tests"), nil)
}

func shell(cmd string, args ...string) {
	if out, err := util.Shell(nil, cmd, args...); err != nil {
		fail(1, "err=%v out=%s", err, out)
	}
}

func copyFile(dst, src string) {
	if _, err := util.CopyFile(dst, src, nil); err != nil {
		fail(1, "err=%v", err)
	}
}

func git(args ...string) {
	shell("git", args...)
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
	shell("unzip", "-q", src, "-d", dest)
}

func removeAll(path string) {
	if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
		fail(1, "%v", err)
	}
}
