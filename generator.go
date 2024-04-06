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
	archivePath = "3b45d155c77bbdfe9177b1e03db830d2aff0b2a8.zip"
	test262Path = "837def66ac1700ce398df4a3b4eaeea915aa1df5.zip"
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
	_, extractedTest262Path := filepath.Split(test262Path)
	extractedTest262Path = "test262-" + extractedTest262Path[:len(extractedTest262Path)-len(".zip")]
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
	fmt.Fprintf(os.Stderr, "archivePath %s\n", archivePath)
	fmt.Fprintf(os.Stderr, "extractedArchivePath %s\n", extractedArchivePath)
	fmt.Fprintf(os.Stderr, "extractedTest262Path %s\n", extractedTest262Path)
	fmt.Fprintf(os.Stderr, "tempDir %s\n", tempDir)
	fmt.Fprintf(os.Stderr, "libRoot %s\n", libRoot)
	fmt.Fprintf(os.Stderr, "makeRoot %s\n", makeRoot)
	util.MustShell(true, nil, "unzip", archivePath, "-d", tempDir)
	util.MustShell(true, nil, "unzip", test262Path, "-d", makeRoot)
	util.MustShell(true, nil, "mv", filepath.Join(makeRoot, extractedTest262Path), filepath.Join(makeRoot, "test262"))
	b, err := os.ReadFile(filepath.Join("internal", "run-test262", "exclude"))
	if err != nil {
		fail(1, "%s\n", err)
	}

	for _, v := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if v = strings.TrimSpace(v); v != "" && !strings.HasPrefix(v, "#") {
			util.Shell(nil, "sh", "-c", fmt.Sprintf("rm -rfv %s", filepath.Join(makeRoot, v)))
		}
	}
	util.MustCopyFile(true, "LICENSE-LIBQUICKJS", filepath.Join(makeRoot, "LICENSE"), nil)
	result := filepath.Join("ccgo.go")
	patch := filepath.Join(util.MustAbsCwd(true), "internal", "patch", "quickjs.diff")
	var resultLine string
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
			"-U__SIZEOF_INT128__",
			"-eval-all-macros",
			"-extended-errors",
		)
		if err := ccgo.NewTask(goos, goarch, append(args, "-exec", "make", "-j", j, "libquickjs.a", "qjs", "run-test262"), os.Stdout, os.Stderr, nil).Exec(); err != nil {
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

		out, err := util.Shell(nil, "./run-test262", "-u", "-c", "test262.conf", "-a")
		if err != nil {
			fmt.Fprintf(os.Stderr, "run-test262 FAIL err=%v", err)
		}
		s := strings.TrimSpace(string(out))
		x := strings.Index(s, "\nResult: ")
		if x < 0 {
			return fmt.Errorf("cannot find test results")
		}
		s = s[x+1:]
		if x = strings.IndexByte(s, '\n'); x > 0 {
			s = s[:x]
		}
		resultLine = strings.TrimSpace(s) + "\n"
		return nil
	})

	if err := os.WriteFile(filepath.Join("internal", "run-test262", "results"), []byte(resultLine), 0660); err != nil {
		fail(1, "cannot save results: %v", err)
	}

	fn := fmt.Sprintf("ccgo_%s_%s.go", goos, goarch)
	util.MustCopyDir(true, filepath.Join("internal", "tests"), filepath.Join(makeRoot, "tests"), nil)
	util.MustCopyFile(true, filepath.Join("internal", "qjs", fn), filepath.Join(makeRoot, "qjs.go"), nil)
	util.MustCopyFile(true, filepath.Join("internal", "run-test262", fn), filepath.Join(makeRoot, "run-test262.go"), nil)
	util.MustCopyFile(true, filepath.Join("internal", "tests", "test262.conf"), filepath.Join(makeRoot, "test262.conf"), nil)
	util.MustCopyFile(true, filepath.Join("internal", "tests", "test262_errors.txt"), filepath.Join(makeRoot, "test262_errors.txt"), nil)
	util.MustCopyFile(true, fn, filepath.Join(makeRoot, result), nil)
	util.Shell(nil, "sh", "-c", "./unconvert.sh")
	util.MustShell(true, nil, "gofmt", "-l", "-s", "-w", ".")
	util.MustShell(true, nil, "go", "test", "-run", "@")
	util.Shell(nil, "git", "status")
}
