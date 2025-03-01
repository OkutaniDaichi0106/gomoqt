# Frame Struct

## New And Release

### FrameDirect
```go
type FrameDirect struct {
    bytes []byte
}
```
- value
```
BenchmarkNewFrame_FrameDirect_BytesPool_Append-16                       30381361                40.54 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_BytesPool_Copy-16                         25918933                43.61 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_PtrBytesPool_Append-16                    29114412                43.26 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_PtrBytesPool_Copy-16                      27403515                42.24 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_FrameDirectPool_Append-16                 28088900                41.91 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_FrameDirectPool_Copy-16                   26562306                44.16 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_PtrFrameDirectPool_Append-16              27826538                43.32 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_PtrFrameDirectPool_Copy-16                27060547                43.04 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_FrameDirectAndBytesPool_Copy-16           14396066                83.34 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FrameDirect_FrameDirectAndBytesPool_Append-16         14364943                84.10 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FrameDirect_FrameDirectAndPtrBytesPool_Copy-16        14782274                80.89 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FrameDirect_FrameDirectAndPtrBytesPool_Append-16      14770120                80.96 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FrameDirect_PtrFrameDirectAndBytes_Copy-16            13386656                82.44 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FrameDirect_PtrFrameDirectAndBytes_Append-16          14788560                81.74 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FrameDirect_PtrFrameDirectAndPtrBytes_Copy-16         21294945                57.16 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FrameDirect_PtrFrameDirectAndPtrBytes_Append-16       22230866                54.46 ns/op           24 B/op          1 allocs/op
```
- pointer
```
BenchmarkNewFrame_PtrFrameDirect_BytesPool_Append-16                            29013048                40.84 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFrameDirect_BytesPool_Copy-16                              17303083                68.41 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrBytesPool_Append-16                         27496447                43.28 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrBytesPool_Copy-16                           29015784                43.07 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFrameDirect_FrameDirectPool_Append-16                      16999718                69.72 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_PtrFrameDirect_FrameDirectPool_Copy-16                        17254864                70.41 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectPool_Append-16                   81654315                15.21 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectPool_Copy-16                     82448160                14.98 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectAndBytesPool_Copy-16             21201338                56.24 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectAndBytesPool_Append-16           21389649                56.23 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectAndPtrBytesPool_Copy-16          44824790                26.68 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectAndPtrBytesPool_Append-16        46811730                25.83 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFrameDirect_FrameDirectAndBytesPool_Copy-16                10875879               108.2 ns/op            72 B/op          3 allocs/op
BenchmarkNewFrame_PtrFrameDirect_FrameDirectAndBytesPool_Append-16              11028884               110.1 ns/op            72 B/op          3 allocs/op
BenchmarkNewFrame_PtrFrameDirect_FrameDirectAndPtrBytesPool_Copy-16             14592184                81.36 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_PtrFrameDirect_FrameDirectAndPtrBytesPool_Append-16           14231178                82.40 ns/op           48 B/op          2 allocs/op
```
### FramePointer
```go
type FramePointer struct {
    bytes *[]byte
}
```
- value
```
BenchmarkNewFrame_FramePointer_BytesPool_Append-16                              16678108                69.64 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FramePointer_BytesPool_Copy-16                                17429750                69.62 ns/op           48 B/op          2 allocs/op
BenchmarkNewFrame_FramePointer_PtrBytesPool_Append-16                           70854146                16.00 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_PtrBytesPool_Copy-16                             77979802                15.91 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerPool_Append-16                       78890276                15.71 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerPool_Copy-16                         76477448                16.78 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_PtrFramePointerPool_Append-16                    29642949                40.74 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_FramePointer_PtrFramePointerPool_Copy-16                      29585943                40.16 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerAndBytesPool_Copy-16                 20500624                57.67 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerAndBytesPool_Append-16               20226673                58.23 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerAndPtrBytesPool_Copy-16              41028306                28.78 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerAndPtrBytesPool_Append-16            43587228                28.31 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_PtrFramePointerAndBytesPool_Copy-16              15093655                81.15 ns/op           32 B/op          2 allocs/op
BenchmarkNewFrame_FramePointer_PtrFramePointerAndBytesPool_Append-16            14475410                81.48 ns/op           32 B/op          2 allocs/op
BenchmarkNewFrame_FramePointer_PtrFramePointerAndPtrBytesPool_Copy-16           22152154                53.83 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_FramePointer_PtrFramePointerAndPtrBytesPool_Append-16         21757182                52.76 ns/op            8 B/op          1 allocs/op
```
- pointer
```
BenchmarkNewFrame_PtrFramePointer_BytesPool_Append-16                           12522422                94.73 ns/op           56 B/op          3 allocs/op
BenchmarkNewFrame_PtrFramePointer_BytesPool_Copy-16                             12607783                96.14 ns/op           56 B/op          3 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrBytesPool_Append-16                        29675498                40.34 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrBytesPool_Copy-16                          28888305                41.29 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_FramePointerPool_Append-16                    29514050                40.80 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_FramePointerPool_Copy-16                      29703410                39.97 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerPool_Append-16                 71867044                16.84 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerPool_Copy-16                   73897084                16.95 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerAndBytesPool_Copy-16           21045430                57.29 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerAndBytesPool_Append-16         20929221                57.08 ns/op           24 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerAndPtrBytesPool_Copy-16        40366798                29.27 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerAndPtrBytesPool_Append-16      39724707                29.54 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFramePointer_FramePointerAndBytesPool_Copy-16              14561180                82.61 ns/op           32 B/op          2 allocs/op
BenchmarkNewFrame_PtrFramePointer_FramePointerAndBytesPool_Append-16            15067993                81.04 ns/op           32 B/op          2 allocs/op
BenchmarkNewFrame_PtrFramePointer_FramePointerAndPtrBytesPool_Copy-16           21212018                55.22 ns/op            8 B/op          1 allocs/op
BenchmarkNewFrame_PtrFramePointer_FramePointerAndPtrBytesPool_Append-16         21826118                55.23 ns/op            8 B/op          1 allocs/op
```
### Summary

**Bests**
```
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectPool_Append-16                   81654315                15.21 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectPool_Copy-16                     82448160                14.98 ns/op            0 B/op          0 allocs/op

BenchmarkNewFrame_FramePointer_PtrBytesPool_Append-16                           70854146                16.00 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_PtrBytesPool_Copy-16                             77979802                15.91 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerPool_Append-16                       78890276                15.71 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerPool_Copy-16                         76477448                16.78 ns/op            0 B/op          0 allocs/op

BenchmarkNewFrame_PtrFramePointer_PtrFramePointerPool_Append-16                 71867044                16.84 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerPool_Copy-16                   73897084                16.95 ns/op            0 B/op          0 allocs/op
```

**Betters**
```
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectAndPtrBytesPool_Copy-16          44824790                26.68 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFrameDirect_PtrFrameDirectAndPtrBytesPool_Append-16        46811730                25.83 ns/op            0 B/op          0 allocs/op

BenchmarkNewFrame_FramePointer_FramePointerAndPtrBytesPool_Copy-16              41028306                28.78 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_FramePointer_FramePointerAndPtrBytesPool_Append-16            43587228                28.31 ns/op            0 B/op          0 allocs/op

BenchmarkNewFrame_PtrFramePointer_PtrFramePointerAndPtrBytesPool_Copy-16        40366798                29.27 ns/op            0 B/op          0 allocs/op
BenchmarkNewFrame_PtrFramePointer_PtrFramePointerAndPtrBytesPool_Append-16      39724707                29.54 ns/op            0 B/op          0 allocs/op
```

**Implementation in gomoqt**
```go
type Frame struct {
    bytes []byte
}

var defaultCap int

ptrFramePool := sync.Pool{
    New: func() {
        return &Frame{
            bytes: make([]byte, 0, defaultCap)
        }
    }
}

func NewFrameBuffer() *Frame {
    // Get
    Frame := ptrFramePool.Get().(*Frame)

    // Initialize the bytes
    *Frame.bytes = *Frame.bytes[:0]

    return Frame
}
```