# Frame Pool Benchmark

## New

```
BenchmarkPoolNew_Bytes-16                5244406               224.0 ns/op          1048 B/op          2 allocs/op
BenchmarkPoolNew_PtrBytes-16             5957877               229.6 ns/op          1048 B/op          2 allocs/op
BenchmarkPoolNew_FrameDirect-16          5007092               226.7 ns/op          1048 B/op          2 allocs/op
BenchmarkPoolNew_PtrFrameDirect-16       4277594               240.2 ns/op          1048 B/op          2 allocs/op
BenchmarkPoolNew_FramePointer-16         4929920               230.2 ns/op          1048 B/op          2 allocs/op
BenchmarkPoolNew_PtrFramePointer-16      4519972               254.5 ns/op          1056 B/op          3 allocs/op
```

### Get and Put

```
BenchmarkPoolGetPut_Bytes-16                      279939              3784 ns/op            2401 B/op        100 allocs/op
BenchmarkPoolGetPut_PtrBytes-16                  1210275               985.1 ns/op             0 B/op          0 allocs/op
BenchmarkPoolGetPut_FrameDirect-16                314594              3763 ns/op            2401 B/op        100 allocs/op
BenchmarkPoolGetPut_PtrFrameDirect-16            1000000              1039 ns/op               0 B/op          0 allocs/op
BenchmarkPoolGetPut_FramePointer-16              1203433              1020 ns/op               0 B/op          0 allocs/op
BenchmarkPoolGetPut_PtrFramePointer-16           1208767              1015 ns/op               0 B/op          0 allocs/op
```