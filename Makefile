# Copyright 2025 The libquickjs-go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all clean edit editor

all: editor
	golint 2>&1
	staticcheck 2>&1

clean:
	rm -f log-* cpu.test mem.test *.out go.work*
	go clean

edit:
	@touch log
	@if [ -f "Session.vim" ]; then gvim -S & else gvim -p Makefile go.mod builder.json all_test.go generator.go libquickjs.go & fi

editor:
	gofmt -l -s -w .
	go test -c -o /dev/null
	go build -o /dev/null generator*.go
