# fluent-logger-go

## Benchmark

```
$ go test -bench . -benchmem
testing: warning: no tests to run
BenchmarkStructFluentLoggerGo-4           200000              7927 ns/op            1503 B/op          8 allocs/op
BenchmarkStructFluentLoggerGolang-4        50000             29026 ns/op            4725 B/op         34 allocs/op
BenchmarkMapFluentLoggerGo-4              200000              8340 ns/op            1141 B/op          6 allocs/op
BenchmarkMapFluentLoggerGolang-4           50000             33745 ns/op            5820 B/op         60 allocs/op
PASS
ok      github.com/daichirata/fluent-logger-go/benchmark        7.242s
```
