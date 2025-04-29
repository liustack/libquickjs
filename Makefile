# Copyright 2025 The libquickjs-go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all clean edit editor generate download dev test nuke

QJS_ZIP = 19abf1888db5884a5758036ff6e7fa2b340acedc.zip
QJS_URL = https://github.com/bellard/quickjs/archive/$(QJS_ZIP)

all: editor
	golint 2>&1
	staticcheck 2>&1

clean:
	rm -f log-* cpu.test mem.test *.out go.work* tmp_file.txt
	go clean

dev:
	GO_GENERATE_DEV=1 make generate

download:
	@if [ ! -f $(QJS_ZIP) ]; then wget $(QJS_URL) ; fi

edit:
	@touch log
	@if [ -f "Session.vim" ]; then gvim -S & else gvim -p Makefile go.mod builder.json all_test.go generator.go libquickjs.go & fi

editor:
	@gofmt -l -s -w $(shell find -name \*.go | grep -v o.go)
	go test -c -o /dev/null
	go build -o /dev/null generator*.go

generate: download
	go generate
	gofmt -l -s -w .
	go build -v -o /dev/null .
	go build -v -o /dev/null ./gqjs
	#TODO go build -v -o /dev/null ./internal/run-test262
	rm -rf ./tmp/
	git status

nuke: clean
	rm -rf ./tmp/

test:
	go test -v -timeout 24h
