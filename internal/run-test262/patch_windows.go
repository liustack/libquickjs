// Copyright 2025 The libquickjs-go Authors. All rights reserved.
// Use of the source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"modernc.org/libc"
	"modernc.org/libc/sys/types"
)

func __putenv(*libc.TLS, uintptr) {
	panic(todo(""))
}

func ___p__environ(*libc.TLS) uintptr {
	panic(todo(""))
}

func _tmpfile(*libc.TLS) uintptr {
	panic(todo(""))
}

func _feof(*libc.TLS, uintptr) int32 {
	panic(todo(""))
}

func _clearerr(*libc.TLS, uintptr) {
	panic(todo(""))
}

func __get_osfhandle(*libc.TLS, int32) uintptr {
	panic(todo(""))
}

func _signal(*libc.TLS, int32, uintptr) {
	panic(todo(""))
}

// func _WaitForMultipleObjects(*libc.TLS, uint32, uintptr, int32, uint32) TDWORD {
// 	panic(todo(""))
// }

func __utime64(*libc.TLS, uintptr, uintptr) int32 {
	panic(todo(""))
}

func __fullpath(*libc.TLS, uintptr, uintptr, uint64) int32 {
	panic(todo(""))
}

func _pthread_attr_setdetachstate(*libc.TLS, uintptr, int32) {
	panic(todo(""))
}

func _mkdir0(*libc.TLS, uintptr, uint16) int32 {
	panic(todo(""))
}

func _clock(*libc.TLS) int32 {
	panic(todo(""))
}

func _getcwd0(tls *libc.TLS, p uintptr, n types.Size_t) uintptr {
	return libc.Xgetcwd(tls, p, n)
}

// int ftw(const char *dirpath,
//
//	int (*fn)(const char *fpath, const struct stat *sb,
//	          int typeflag),
//	int nopenfd);
func _ftw(tls *libc.TLS, dir, fn uintptr, n int32) int32 {
	panic(todo(""))
}
