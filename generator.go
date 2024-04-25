// Copyright 2024 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"modernc.org/cc/v4"
	ccgo "modernc.org/ccgo/v4/lib"
	util "modernc.org/fileutil/ccgo"
	"modernc.org/gc/v3"
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
			// methods: disabled
			// "--prefix-static-internal=g_",
			// "--prefix-static-none=g_",
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

		// methods: disabled
		//
		// if err := toMethods(result); err != nil {
		// 	return err
		// }
		//
		// util.MustShell(true, nil, "sed", "-i", `s/(\*(\*func(\*libc\.TLS/(*(*func(*libc.TLS, *Library/g`, result)

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

	results := fmt.Sprintf("results_%s_%s", goos, goarch)
	if err := os.WriteFile(filepath.Join("internal", "run-test262", results), []byte(resultLine), 0660); err != nil {
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

func toMethods(fn string) (err error) {
	b, err := os.ReadFile(fn)
	if err != nil {
		return err
	}

	ast, err := gc.ParseFile(fn, b)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	w := func(s string, args ...interface{}) {
		fmt.Fprintf(buf, s, args...)
	}

	src := ast.SourceFile
	w("%s\n", src.PackageClause.Source(true))
	if src.ImportDeclList != nil {
		w("%s\n", src.ImportDeclList.Source(true))
	}

	vars := map[*gc.VarDeclNode]string{}
	xvars := map[string]*gc.VarDeclNode{}
	funcs := map[string]struct{}{}
	initFuncs := map[string]*gc.FunctionDeclNode{}
	initNum := 0
	for l := src.TopLevelDeclList; l != nil; l = l.List {
		switch x := l.TopLevelDecl.(type) {
		case *gc.FunctionDeclNode:
			switch nm := x.FunctionName.IDENT.Src(); {
			case nm == "__ccgo_fp":
				// nop
			case nm == "main":
				// nop
			case nm == "init":
				// nop
			default:
				funcs[nm] = struct{}{}
			}
		}
	}
	var a []string
	// for k := range funcs {
	// 	a = append(a, k)
	// }
	// slices.Sort(a)
	// for _, k := range a {
	// 	trc("FUNC %s", k)
	// }
	for l := src.TopLevelDeclList; l != nil; l = l.List {
		switch x := l.TopLevelDecl.(type) {
		case *gc.VarDeclNode:
			if x.LPAREN.IsValid() {
				w("%s\n\n", x.Source(true))
				break
			}

			switch y := x.VarSpec.(type) {
			case *gc.VarSpecNode:
				nm := y.IDENT.Src()
				if !strings.HasPrefix(nm, "x_") && !strings.HasPrefix(nm, "g_") {
					w("%s\n\n", x.Source(true))
					break
				}

				vars[x] = nm
				xvars[nm] = x
			default:
				w("%s\n\n", x.Source(true))
			}
		case *gc.ConstDeclNode:
			w("%s\n\n", x.Source(true))
		case *gc.TypeDeclNode:
			w("%s\n\n", x.Source(true))
		case *gc.FunctionDeclNode:
			switch nm := x.FunctionName.IDENT.Src(); {
			case skip[nm]:
				w("%s\n\n", x.Source(true))
				continue
			case nm == "main":
				// nop
			case nm == "init":
				s := x.Source(true)
				initNum++
				nm := fmt.Sprintf("%s%v", nm, initNum)
				initFuncs[nm] = x
				ix := strings.Index(s, "func init")
				// func init
				head := s[:ix]
				tail := s[ix+9:]
				tail = rename(tail, funcs)
				w("%sfunc (qjs *Library) init%v%s\n\n", head, initNum, tail)
			default:
				s := x.Source(true)
				ix := strings.Index(s, "{\n")
				// func foo(tls *libc.TLS... {
				head := s[:ix]
				head = strings.Replace(head, "(tls *libc.TLS", "(tls *libc.TLS, qjs *Library", 1)
				tail := s[ix:]
				tail = rename(tail, funcs)
				w("%s%s\n\n", head, tail)

			}
		default:
			panic(todo("%T", x))
		}
	}

	w("\n// Library represents an instance of the quickjs library")
	w("\n\ntype Library struct{\n")
	a = a[:0]
	for k := range xvars {
		a = append(a, k)
	}
	slices.Sort(a)
	initializers := map[string]gc.Node{}
	for _, nm := range a {
		vd := xvars[nm]
		vs := vd.VarSpec.(*gc.VarSpecNode)
		switch {
		case vs.TypeNode != nil:
			w("%s %s\n", nm, vs.TypeNode.Source(true))
			if vs.ExpressionList != nil {
				panic(todo(""))
			}
		default:
			switch x := vs.ExpressionList.Expression.(type) {
			case *gc.CompositeLitNode:
				initializers[nm] = x
				w("%s %s\n", nm, x.LiteralType.Source(true))
			case *gc.PrimaryExprNode:
				initializers[nm] = x
				w("%s %s\n", nm, primaryExprType(x))
			case *gc.OperandNode:
				if x.OperandName != nil {
					initializers[nm] = x
					w("%s %s\n", nm, x.OperandName.Source(true))
					break
				}

				if x.TypeArgs != nil {
					panic(todo("%T\n%s", x, x.Source(true)))
				}

				if x.LiteralValue != nil {
					panic(todo("%T\n%s", x, x.Source(true)))
				}

				panic(todo(""))
			default:
				panic(todo("%T\n%s", x, x.Source(true)))
			}

		}
	}
	w("}\n\n")
	w("func (qjs *Library) initInstance() {\n")
	a = a[:0]
	for k := range initializers {
		a = append(a, k)
	}
	slices.Sort(a)
	for _, nm := range a {
		w("qjs.%s = %s\n", nm, rename(initializers[nm].Source(true), funcs))
	}
	a = a[:0]
	for k := range initFuncs {
		a = append(a, k)
	}
	slices.Sort(a)
	for _, nm := range a {
		w("qjs.%s()\n", nm)
	}
	w("tls := libc.NewTLS()\n")
	w("defer tls.Close()\n")
	w("Xinit_quickjs(tls, qjs)\n")
	w("}\n")

	if err := os.WriteFile(fn, buf.Bytes(), 0660); err != nil {
		return err
	}

	return nil
}

var (
	skip = map[string]bool{
		"__ccgo_fp": true,
	}
	re  = regexp.MustCompile(`\b[gx]_[a-zA-Z0-9_]+\b`)
	re2 = regexp.MustCompile(`\b[gx]_[a-zA-Z0-9_]+\b\(tls`)
	re3 = regexp.MustCompile(`}\)\)\)\(tls\b`)
)

func rename(s string, funcs map[string]struct{}) string {
	s = re.ReplaceAllStringFunc(s, func(s string) string {
		if _, ok := funcs[s]; !ok {
			return "qjs." + s
		}

		return s
	})
	s = re2.ReplaceAllStringFunc(s, func(s string) string {
		return s + ", qjs"
	})
	return re3.ReplaceAllStringFunc(s, func(s string) string {
		return s + ", qjs"
	})
}

func origin(skip int) string {
	pc, fn, fl, _ := runtime.Caller(skip)
	f := runtime.FuncForPC(pc)
	var fns string
	if f != nil {
		fns = f.Name()
		if x := strings.LastIndex(fns, "."); x > 0 {
			fns = fns[x+1:]
		}
		if strings.HasPrefix(fns, "func") {
			num := true
			for _, c := range fns[len("func"):] {
				if c < '0' || c > '9' {
					num = false
					break
				}
			}
			if num {
				return origin(skip + 2)
			}
		}
	}
	return fmt.Sprintf("%s:%d:%s", filepath.Base(fn), fl, fns)
}

func todo(s string, args ...interface{}) string {
	switch {
	case s == "":
		s = fmt.Sprintf(strings.Repeat("%v ", len(args)), args...)
	default:
		s = fmt.Sprintf(s, args...)
	}
	r := fmt.Sprintf("%s\n\tTODO %s", origin(2), s)
	// fmt.Fprintf(os.Stderr, "%s\n", r)
	// os.Stdout.Sync()
	return r
}

func primaryExprType(n *gc.PrimaryExprNode) string {
	if n.Postfix != nil {
		return n.PrimaryExpr.Source(true)
	}

	switch x := n.PrimaryExpr.(type) {
	default:
		panic(todo("%T", x))
	}
}
