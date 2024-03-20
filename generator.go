// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

//TODO float rounding mode
//TODO pthreads

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"modernc.org/cc/v4"
	ccgo "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
)

const (
	archivePath = "6a89d7c27099be84e5312a7ec73205d6a7abe1b4.zip"
)

var (
	goos   = runtime.GOOS
	goarch = runtime.GOARCH
	j      = strconv.Itoa(runtime.GOMAXPROCS(-1))
)

func fail(rc int, msg string, args ...any) {
	fmt.Fprintln(os.Stderr, strings.TrimSpace(fmt.Sprintf(msg, args...)))
	os.Exit(rc)
}

func main() {
	if ccgo.IsExecEnv() {
		if err := ccgo.NewTask(goos, goarch, os.Args, os.Stdout, os.Stderr, nil).Main(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	f, err := os.Open(archivePath)
	if err != nil {
		fail(1, "cannot open tar file: %v\n", err)
	}

	f.Close()

	_, extractedArchivePath := filepath.Split(archivePath)
	extractedArchivePath = "quickjs-" + extractedArchivePath[:len(extractedArchivePath)-len(".zip")]
	tempDir := os.Getenv("GO_GENERATE_DIR")
	dev := os.Getenv("GO_GENERATE_DEV") != ""
	switch {
	case tempDir != "":
		util.MustShell(true, nil, "sh", "-c", fmt.Sprintf("rm -rf %s", filepath.Join(tempDir, extractedArchivePath)))
	default:
		var err error
		if tempDir, err = os.MkdirTemp("", "libquickjs-generate-"); err != nil {
			fail(1, "creating temp dir: %v\n", err)
		}

		defer func() {
			switch os.Getenv("GO_GENERATE_KEEP") {
			case "":
				os.RemoveAll(tempDir)
			default:
				fmt.Printf("%s: temporary directory kept\n", tempDir)
			}
		}()
	}
	libRoot := filepath.Join(tempDir, extractedArchivePath)
	makeRoot := libRoot
	tests := filepath.Join("internal", goos, goarch)
	os.MkdirAll(tests, 0770)
	if tests, err = filepath.Abs(tests); err != nil {
		fail(1, "%s\n", err)
	}

	if err := os.RemoveAll(tests); err != nil {
		fail(1, "%s\n", err)
	}

	fmt.Fprintf(os.Stderr, "archivePath %s\n", archivePath)
	fmt.Fprintf(os.Stderr, "extractedArchivePath %s\n", extractedArchivePath)
	fmt.Fprintf(os.Stderr, "tempDir %s\n", tempDir)
	fmt.Fprintf(os.Stderr, "libRoot %s\n", libRoot)
	fmt.Fprintf(os.Stderr, "makeRoot %s\n", makeRoot)
	fmt.Fprintf(os.Stderr, "tests %s\n", tests)
	util.MustShell(true, nil, "unzip", archivePath, "-d", tempDir)
	util.MustCopyFile(true, "LICENSE-LIBQUICKJS", filepath.Join(makeRoot, "LICENSE"), nil)
	result := filepath.Join("ccgo.go")
	ilibc := filepath.Join(util.MustAbsCwd(true), "..", "libc", "include", goos, goarch)
	patch := filepath.Join(util.MustAbsCwd(true), "internal", "patch", "quickjs.diff")
	util.MustInDir(true, makeRoot, func() (err error) {
		util.MustShell(true, nil, "git", "apply", patch)
		util.MustShell(true, nil, "sh", "-c", "go mod init example.com/libquickjs ; go get modernc.org/libc@latest")
		if dev {
			util.MustShell(true, nil, "sh", "-c", "go work init ; go work use $GOPATH/src/modernc.org/libc")
		}
		args := []string{os.Args[0]}
		if s := cc.LongDouble64Flag(goos, goarch); s != "" {
			args = append(args, s)
		}
		if dev {
			args = append(
				args,
				"-absolute-paths",
				"-keep-object-files",
				"-positions",
			)
		}
		args = append(args,
			"--prefix-enumerator=_",
			"--prefix-external=x_",
			"--prefix-field=F",
			"--prefix-macro=m_",
			"--prefix-static-internal=_",
			"--prefix-static-none=_",
			"--prefix-tagged-enum=_",
			"--prefix-tagged-struct=T",
			"--prefix-tagged-union=T",
			"--prefix-typename=T",
			"--prefix-undefined=_",
			"-DNDEBUG",
			"-I", ilibc,
			"-U__SIZEOF_INT128__",
			"-extended-errors",
		)
		if err := ccgo.NewTask(goos, goarch, append(args, "-exec", "make", "-j", j, "libquickjs.a"), os.Stdout, os.Stderr, nil).Exec(); err != nil {
			return err
		}

		if err := ccgo.NewTask(goos, goarch, append(args,
			"--package-name", "libquickjs",
			"-ignore-link-errors",
			"-o", result,
			filepath.Join("libquickjs.a"),
		), os.Stdout, os.Stderr, nil).Main(); err != nil {
			return err
		}

		util.MustShell(true, nil, "sed", "-i", `s/\<T__\([a-zA-Z0-9][a-zA-Z0-9_]\+\)/t__\1/g`, result)
		util.MustShell(true, nil, "sed", "-i", `s/\<x_\([a-zA-Z0-9_][a-zA-Z0-9_]\+\)/X\1/g`, result)
		return nil
	})

	fn := fmt.Sprintf("ccgo_%s_%s.go", goos, goarch)
	util.MustCopyFile(false, fn, filepath.Join(makeRoot, result), nil)
	util.Shell(nil, "sh", "-c", "./unconvert.sh")
	util.MustShell(true, nil, "gofmt", "-l", "-s", "-w", ".")
	util.MustShell(true, nil, "go", "test", "-run", "@")
	util.Shell(nil, "git", "status")
}
