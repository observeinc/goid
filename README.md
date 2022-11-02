# goid
[![Test](https://github.com/observeinc/goid/workflows/Test/badge.svg)](https://github.com/observeinc/goid/actions?query=workflow%3ATest)
[![Lint](https://github.com/observeinc/goid/workflows/Lint/badge.svg)](https://github.com/observeinc/goid/actions?query=workflow%3ALint)
[![codecov](https://codecov.io/gh/observeinc/goid/branch/master/graph/badge.svg)](https://codecov.io/gh/observeinc/goid)
[![Go Report Card](https://goreportcard.com/badge/github.com/observeinc/goid)](https://goreportcard.com/report/github.com/observeinc/goid)

An inelegant but efficient way to get the goroutine id.

Forked from [rpccloud/goid](https://github.com/rpccloud/goid). All credit to [tslearn](https://github.com/tslearn) for figuring out how to call `getg()` from user code.

## Usage
```go
package main

import (
  "fmt"
  "github.com/observeinc/goid"
)

func main() {
  fmt.Println("Current Goroutine ID:", goid.GetGoID())
}
```

## Benchmark
```bash
$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/observeinc/goid
cpu: 11th Gen Intel(R) Core(TM) i9-11950H @ 2.60GHz
BenchmarkSlowGid-16       472669              2565 ns/op              64 B/op          2 allocs/op
BenchmarkFastGid-16     950177301                1.352 ns/op           0 B/op          0 allocs/op
BenchmarkGetGoID-16     563974591                2.009 ns/op           0 B/op          0 allocs/op
PASS
ok      github.com/observeinc/goid      4.026s
```
