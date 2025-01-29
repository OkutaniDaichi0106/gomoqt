package benchmarks_test

import "testing"

// CopyPayload for GroupPointer copies the payload data
func (f *GroupPointer) CopyPayload() []byte {
	copyData := make([]byte, len(*f.payload))
	copy(copyData, *f.payload)
	return copyData
}

// Size for GroupPointer returns the size of the payload
func (f *GroupPointer) Size() int {
	return len(*f.payload)
}

// ReleaseBytes releases the payload back to the Bytes pool
func (f *GroupPointer) ReleaseBytes() {
	arr := *f.payload
	arr = arr[0:0]
	bytesPool.Put(arr)
}

// ReleasePtrBytes releases the payload back to the PtrBytes pool
func (f *GroupPointer) ReleasePtrBytes() {
	ptr := f.payload
	arr := *ptr
	arr = arr[0:0]
	*ptr = arr
	ptrBytesPool.Put(ptr)
}

// ReleaseGroupPointer releases the GroupPointer back to the GroupPointer pool
func (f *GroupPointer) ReleaseGroupPointer() {
	ptr := f.payload
	arr := *ptr
	arr = arr[0:0]
	*ptr = arr
	groupPointerPool.Put(*f)
}

// ReleasePtrGroupPointer releases the GroupPointer back to the PtrGroupPointer pool
func (f *GroupPointer) ReleasePtrGroupPointer() {
	ptr := f.payload
	arr := *ptr
	arr = arr[0:0]
	*ptr = arr
	ptrGroupPointerPool.Put(f)
}

// NewGroupPointer_BytesPool_Copy creates a new Group with a *[]byte payload using copy
func NewGroupPointer_BytesPool_Copy(payload []byte) GroupPointer {
	arr := bytesPool.Get().([]byte)
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	return GroupPointer{
		payload: &arr,
	}
}

// NewGroupPointer_BytesPool_Append creates a new Group with a *[]byte payload using append
func NewGroupPointer_BytesPool_Append(payload []byte) GroupPointer {
	arr := bytesPool.Get().([]byte)
	arr = arr[0:0]
	arr = append(arr, payload...)
	return GroupPointer{
		payload: &arr,
	}
}

// NewGroupPointer_PtrBytesPool_Copy creates a new Group with a *[]byte payload using copy
func NewGroupPointer_PtrBytesPool_Copy(payload []byte) GroupPointer {
	ptr := ptrBytesPool.Get().(*[]byte)
	arr := *ptr
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	*ptr = arr
	return GroupPointer{
		payload: ptr,
	}
}

// NewGroupPointer_PtrBytesPool_Append creates a new Group with a *[]byte payload using append
func NewGroupPointer_PtrBytesPool_Append(payload []byte) GroupPointer {
	ptr := ptrBytesPool.Get().(*[]byte)
	arr := *ptr
	arr = arr[0:0]
	arr = append(arr, payload...)
	*ptr = arr
	return GroupPointer{
		payload: ptr,
	}
}

// NewGroupPointer_GroupPointerPool_Copy creates a new GroupPointer from GroupPointer pool using copy
func NewGroupPointer_GroupPointerPool_Copy(payload []byte) GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	arr := *group.payload
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	*group.payload = arr
	return group
}

// NewGroupPointer_GroupPointerPool_Append creates a new GroupPointer from GroupPointer pool using append
func NewGroupPointer_GroupPointerPool_Append(payload []byte) GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	arr := *group.payload
	arr = arr[0:0]
	arr = append(arr, payload...)
	*group.payload = arr
	return group
}

// NewPtrGroupPointer_BytesPool_Append creates a new PtrGroupPointer from Bytes pool using append
func NewPtrGroupPointer_BytesPool_Append(payload []byte) *GroupPointer {
	arr := bytesPool.Get().([]byte)
	arr = arr[0:0]
	arr = append(arr, payload...)
	return &GroupPointer{
		payload: &arr,
	}
}

// NewPtrGroupPointer_BytesPool_Copy creates a new PtrGroupPointer from Bytes pool using copy
func NewPtrGroupPointer_BytesPool_Copy(payload []byte) *GroupPointer {
	arr := bytesPool.Get().([]byte)
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	return &GroupPointer{
		payload: &arr,
	}
}

// NewPtrGroupPointer_PtrBytesPool_Append creates a new PtrGroupPointer from PtrBytes pool using append
func NewPtrGroupPointer_PtrBytesPool_Append(payload []byte) *GroupPointer {
	ptr := ptrBytesPool.Get().(*[]byte)
	arr := *ptr
	arr = arr[0:0]
	arr = append(arr, payload...)
	*ptr = arr
	return &GroupPointer{
		payload: ptr,
	}
}

// NewPtrGroupPointer_PtrBytesPool_Copy creates a new PtrGroupPointer from PtrBytes pool using copy
func NewPtrGroupPointer_PtrBytesPool_Copy(payload []byte) *GroupPointer {
	ptr := ptrBytesPool.Get().(*[]byte)
	arr := *ptr
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	*ptr = arr
	return &GroupPointer{
		payload: ptr,
	}
}

// NewPtrGroupPointer_GroupPointerPool_Append creates a new PtrGroupPointer from GroupPointer pool using append
func NewPtrGroupPointer_GroupPointerPool_Append(payload []byte) *GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	arr := *group.payload
	arr = arr[0:0]
	arr = append(arr, payload...)
	*group.payload = arr
	return &group
}

// NewPtrGroupPointer_GroupPointerPool_Copy creates a new PtrGroupPointer from GroupPointer pool using copy
func NewPtrGroupPointer_GroupPointerPool_Copy(payload []byte) *GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	arr := *group.payload
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	*group.payload = arr
	return &group
}

// NewPtrGroupPointer_PtrGroupPointerPool_Append creates a new PtrGroupPointer from PtrGroupPointer pool using append
func NewPtrGroupPointer_PtrGroupPointerPool_Append(payload []byte) *GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	arr := *group.payload
	arr = arr[0:0]
	arr = append(arr, payload...)
	*group.payload = arr
	return group
}

// NewPtrGroupPointer_PtrGroupPointerPool_Copy creates a new PtrGroupPointer from PtrGroupPointer pool using copy
func NewPtrGroupPointer_PtrGroupPointerPool_Copy(payload []byte) *GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	arr := *group.payload
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	*group.payload = arr
	return group
}

func NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Copy(payload []byte) *GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	*group.payload = buf
	return group
}

func NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Append(payload []byte) *GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	*group.payload = buf
	return group
}

func NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(payload []byte) *GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	*group.payload = *ptr
	return group
}

func NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Append(payload []byte) *GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	*group.payload = *ptr
	return group
}

// NewGroupPointer_PtrGroupPointerPool_Copy creates a new GroupPointer from PtrGroupPointer pool using copy
func NewGroupPointer_PtrGroupPointerPool_Copy(payload []byte) GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	arr := *group.payload
	if cap(arr) < len(payload) {
		arr = make([]byte, len(payload))
	} else {
		arr = arr[:len(payload)]
	}
	copy(arr, payload)
	*group.payload = arr
	return *group
}

// NewGroupPointer_PtrGroupPointerPool_Append creates a new GroupPointer from PtrGroupPointer pool using append
func NewGroupPointer_PtrGroupPointerPool_Append(payload []byte) GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	arr := *group.payload
	arr = arr[0:0]
	arr = append(arr, payload...)
	*group.payload = arr
	return *group
}

// New constructor functions
func NewGroupPointer_GroupPointerAndBytesPool_Copy(payload []byte) GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	*group.payload = buf
	return group
}

func NewGroupPointer_GroupPointerAndBytesPool_Append(payload []byte) GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	*group.payload = buf
	return group
}

func NewGroupPointer_GroupPointerAndPtrBytesPool_Copy(payload []byte) GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	*group.payload = *ptr
	return group
}

func NewGroupPointer_GroupPointerAndPtrBytesPool_Append(payload []byte) GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	*group.payload = *ptr
	return group
}

// New constructor functions for GroupPointer with PtrGroupPointer combined pools
func NewGroupPointer_PtrGroupPointerAndBytesPool_Copy(payload []byte) GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	*group.payload = buf
	return *group
}

func NewGroupPointer_PtrGroupPointerAndBytesPool_Append(payload []byte) GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	*group.payload = buf
	return *group
}

func NewGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(payload []byte) GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	*group.payload = *ptr
	return *group
}

func NewGroupPointer_PtrGroupPointerAndPtrBytesPool_Append(payload []byte) GroupPointer {
	group := ptrGroupPointerPool.Get().(*GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	*group.payload = *ptr
	return *group
}

// New constructor functions for combined pools
func NewPtrGroupPointer_GroupPointerAndBytesPool_Copy(payload []byte) *GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	*group.payload = buf
	return &group
}

func NewPtrGroupPointer_GroupPointerAndBytesPool_Append(payload []byte) *GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	*group.payload = buf
	return &group
}

func NewPtrGroupPointer_GroupPointerAndPtrBytesPool_Copy(payload []byte) *GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	*group.payload = *ptr
	return &group
}

func NewPtrGroupPointer_GroupPointerAndPtrBytesPool_Append(payload []byte) *GroupPointer {
	group := groupPointerPool.Get().(GroupPointer)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	*group.payload = *ptr
	return &group
}

// Fifth group - GroupPointer basic operations
func BenchmarkNewGroup_GroupPointer_BytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_BytesPool_Append(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_GroupPointer_BytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_BytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_GroupPointer_GroupPointerPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_GroupPointerPool_Append(defaultPayload)
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_GroupPointerPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_GroupPointerPool_Copy(defaultPayload)
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrGroupPointerPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrGroupPointerPool_Append(defaultPayload)
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrGroupPointerPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrGroupPointerPool_Copy(defaultPayload)
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_GroupPointerAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_GroupPointerAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_GroupPointerAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_GroupPointerAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_GroupPointerAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_GroupPointerAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_GroupPointerAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_GroupPointerAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrGroupPointerAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrGroupPointerAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrGroupPointerAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrGroupPointerAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_GroupPointer_PtrGroupPointerAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupPointer_PtrGroupPointerAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupPointer()
	}
}

// Sixth group - PtrGroupPointer all operations
func BenchmarkNewGroup_PtrGroupPointer_BytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_BytesPool_Append(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_BytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_BytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_GroupPointerPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_GroupPointerPool_Append(defaultPayload)
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_GroupPointerPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_GroupPointerPool_Copy(defaultPayload)
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrGroupPointerPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrGroupPointerPool_Append(defaultPayload)
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrGroupPointerPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrGroupPointerPool_Copy(defaultPayload)
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrGroupPointerAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrGroupPointerAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupPointer()
	}
}

// Benchmark functions
func BenchmarkNewGroup_PtrGroupPointer_GroupPointerAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_GroupPointerAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_GroupPointerAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_GroupPointerAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_GroupPointerAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_GroupPointerAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupPointer()
	}
}

func BenchmarkNewGroup_PtrGroupPointer_GroupPointerAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupPointer_GroupPointerAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupPointer()
	}
}
