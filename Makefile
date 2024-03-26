# Copyright 2024 The libquickjs-go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all clean dev download edit editor generate work test diff

DIR = /tmp/libquickjs
TAR0 = 6a89d7c27099be84e5312a7ec73205d6a7abe1b4
TAR = $(TAR0).zip
URL = https://github.com/bellard/quickjs/archive/$(TAR)
TMP = /tmp/libquickjs/quickjs-$(TAR0)/

all: editor
	golint 2>&1
	staticcheck 2>&1

build_all_targets:
	./build_all_targets.sh
	echo done

clean:
	rm -f log-* cpu.test mem.test *.out go.work*
	go clean

edit:
	@touch log
	@if [ -f "Session.vim" ]; then novim -S & else novim -p Makefile go.mod builder.json all_test.go libquickjs.go generator.go & fi

editor:
	gofmt -l -s -w . 2>&1 | tee log-editor
	go test -c -o /dev/null 2>&1 | tee -a log-editor
	go install -v  ./... 2>&1 | tee -a log-editor
	go build -o /dev/null generator*.go

download:
	@if [ ! -f $(TAR) ]; then wget $(URL) ; fi


generate: download
	mkdir -p $(DIR) || true
	rm -rf $(DIR)/*
	echo -n > log-generate
	echo -n > log-generate-errors
	GO_GENERATE_DIR=$(DIR) go run generator*.go 2> log-generate-errors | tee log-generate
	cat log-generate-errors
	go build -v ./...
	# go install github.com/mdempsky/unconvert@latest
	go build -v ./...  | tee -a log-generate
	go test -v -count=1 ./... 2>&1 | tee -a log-generate
	git status
	grep 'PASS\|TRC\|TODO\|ERRORF\|FAIL' log-generate || true
	grep 'PASS\|TRC\|TODO\|ERRORF\|FAIL' log-generate-errors || true

dev: download
	mkdir -p $(DIR) || true
	rm -rf $(DIR)/*
	echo -n > /tmp/ccgo.log
	echo -n > log-generate
	echo -n > log-generate-errors
	GO_GENERATE_DIR=$(DIR) GO_GENERATE_DEV=1 go run -tags=ccgo.dmesg,ccgo.assert generator*.go 2>&1 | tee log-generate
	go build -v ./...  | tee -a log-generate
	go test -v -count=1 ./... 2>&1 | tee -a log-generate
	git status
	grep 'PASS\|TRC\|TODO\|ERRORF\|FAIL' log-generate || true
	grep 'PASS\|TRC\|TODO\|ERRORF\|FAIL' log-generate-errors || true
	grep 'PASS\|TRC\|TODO\|ERRORF\|FAIL' /tmp/ccgo.log || true

test:
	go test -v -timeout 24h -count=1 2>&1 | tee log-test
	grep -a 'TRC\|TODO\|ERRORF\|FAIL' log-test || true 2>&1 | tee -a log-test

work:
	rm -f go.work*
	go work init
	go work use .
	go work use ../ccgo/v4
	go work use ../libc

diff0:
	make -C $(TMP) clean
	find $(TMP) -name \*.ago -delete
	find $(TMP) -name \*.go -delete
	find $(TMP) -name go.\* -delete
	git diff quickjs-$(TAR0) $(TMP) | sed -e 's|/quickjs-$(TAR0)||g' -e 's|/tmp/libquickjs||g'

diff:
	@make diff0 | tail -n +16 | head -n -1 > internal/patch/quickjs.diff
