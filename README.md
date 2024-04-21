# libquickjs

Package libquickjs is a CGo-free, ccgo/v4 version of libquickjs.a, a library
implementing an embeddable Javascript engine.

It supports the ES2023 specification including modules, asynchronous
generators, proxies and BigInt.  It optionally supports mathematical
extensions such as big decimal floating point numbers (BigDecimal), big
binary floating point numbers (BigFloat) and operator overloading.

# Supported platforms and architectures

These combinations of GOOS and GOARCH are currently supported

   OS      Arch
   ---------------
   linux	amd64
   linux	loong64

# Builders

Builder results available at:

https://modern-c.appspot.com/-/builder/?importpath=modernc.org%2flibquickjs

If a builder reports PASS then this package passed all the [test262] suite
test cases the original C version passed on the respective target.

[test262]: https://github.com/tc39/test262
