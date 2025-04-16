# Copyright 2024 The libquickjs-go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

.PHONY:	all clean dev download edit editor generate work test diff

DIR = /tmp/libquickjs

TAR0 = 993660621a0a834c93eaad68713b2dc25508b326
TAR = $(TAR0).zip
URL = https://github.com/bellard/quickjs/archive/$(TAR)
ZIP = ce7e72d2107f99d165f4259571f10aa75753d997.zip
URL2 = https://github.com/tc39/test262/archive/$(ZIP)
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
	@if [ -f "Session.vim" ]; then novim -S & else novim -p Makefile go.mod builder.json all_test.go libquickjs.go libquickjs32.go libquickjs64.go generator.go & fi

editor:
	gofmt -l -s -w .
	go test -c -o /dev/null
	go install -v  ./...
	go build -o /dev/null generator*.go

download:
	@if [ ! -f $(TAR) ]; then wget $(URL) ; fi
	@if [ ! -f $(ZIP) ]; then wget $(URL2) ; fi


generate: download
	mkdir -p $(DIR) || true
	rm -rf $(DIR)/*
	GO_GENERATE_DIR=$(DIR) go run generator*.go
	go build -v ./...
	git status

dev: download
	mkdir -p $(DIR) || true
	rm -rf $(DIR)/*
	echo -n > /tmp/ccgo.log
	GO_GENERATE_DIR=$(DIR) GO_GENERATE_DEV=1 go run -tags=ccgo.dmesg,ccgo.assert generator*.go
	go build -v ./...
	git status
	grep 'PASS\|TRC\|TODO\|ERRORF\|FAIL' /tmp/ccgo.log || true

test:
	go test -v -timeout 24h -count=1

work:
	rm -f go.work*
	go work init
	go work use .
	go work use ../cc/v4
	go work use ../ccgo/v4
	go work use ../libc

diff0:
	make -C $(TMP) clean
	find $(TMP) -name \*.ago -delete
	find $(TMP) -name \*.go -delete
	find $(TMP) -name go.\* -delete
	git diff quickjs-$(TAR0) $(TMP) | sed -e 's|/quickjs-$(TAR0)||g' -e 's|/tmp/libquickjs||g'

diff:
	rm -rf /tmp/libquickjs/quickjs-$(TAR0)/test262
	rm -rf /tmp/libquickjs/quickjs-$(TAR0)/test262_report.txt
	cp quickjs-$(TAR0)/test262_errors.txt /tmp/libquickjs/quickjs-$(TAR0)/test262_errors.txt
	@make diff0 | tail -n +16 | head -n -1 > internal/patch/quickjs.diff
