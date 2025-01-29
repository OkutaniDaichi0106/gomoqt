package benchmarks_test

import (
	"bytes"
	"testing"
)

// NewGroupDirect_BytesPool_Copy creates a new Group with a direct []byte payload using copy
func NewGroupDirect_BytesPool_Copy(payload []byte) GroupDirect {
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = (buf)[:len(payload)]
	}
	copy(buf, payload)
	return GroupDirect{
		payload: buf,
	}
}

// NewGroupDirect_BytesPool_Append creates a new Group with a direct []byte payload using append
func NewGroupDirect_BytesPool_Append(payload []byte) GroupDirect {
	buf := bytesPool.Get().([]byte)
	buf = (buf)[:0]
	buf = append(buf, payload...)
	return GroupDirect{
		payload: buf,
	}
}

// NewGroupDirect_PtrBytesPool_Copy creates a new Group with a direct []byte payload using copy
func NewGroupDirect_PtrBytesPool_Copy(payload []byte) GroupDirect {
	buf := ptrBytesPool.Get().(*[]byte)
	if cap(*buf) < len(payload) {
		*buf = make([]byte, len(payload))
	} else {
		*buf = (*buf)[:len(payload)]
	}
	copy(*buf, payload)
	return GroupDirect{
		payload: *buf,
	}
}

// NewGroupDirect_PtrBytesPool_Append creates a new Group with a direct []byte payload using append
func NewGroupDirect_PtrBytesPool_Append(payload []byte) GroupDirect {
	buf := ptrBytesPool.Get().(*[]byte)
	*buf = (*buf)[:0]
	*buf = append(*buf, payload...)
	return GroupDirect{
		payload: *buf,
	}
}

// NewGroupDirect_GroupDirectPool_Copy creates a new GroupDirect from GroupDirect pool using copy
func NewGroupDirect_GroupDirectPool_Copy(payload []byte) GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	if cap(group.payload) < len(payload) {
		group.payload = make([]byte, len(payload))
	} else {
		group.payload = group.payload[:len(payload)]
	}
	copy(group.payload, payload)
	return group
}

// NewGroupDirect_GroupDirectPool_Append creates a new GroupDirect from GroupDirect pool using append
func NewGroupDirect_GroupDirectPool_Append(payload []byte) GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	group.payload = group.payload[:0]
	group.payload = append(group.payload, payload...)
	return group
}

// NewGroupDirect_GroupDirectAndBytesPool_Copy creates a new GroupDirect using two pools
func NewGroupDirect_GroupDirectAndBytesPool_Copy(payload []byte) GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	group.payload = buf
	return group
}

// NewGroupDirect_GroupDirectAndBytesPool_Append creates a new GroupDirect using two pools
func NewGroupDirect_GroupDirectAndBytesPool_Append(payload []byte) GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	group.payload = buf
	return group
}

// NewGroupDirect_GroupDirectAndPtrBytesPool_Copy creates a new GroupDirect using two pools
func NewGroupDirect_GroupDirectAndPtrBytesPool_Copy(payload []byte) GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	group.payload = *ptr
	return group
}

// NewGroupDirect_GroupDirectAndPtrBytesPool_Append creates a new GroupDirect using two pools
func NewGroupDirect_GroupDirectAndPtrBytesPool_Append(payload []byte) GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	group.payload = *ptr
	return group
}

// NewGroupDirect_PtrGroupDirectPool_Append creates a new GroupDirect from PtrGroupDirect pool
func NewGroupDirect_PtrGroupDirectPool_Append(payload []byte) GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	group.payload = group.payload[:0]
	group.payload = append(group.payload, payload...)
	return *group
}

// NewGroupDirect_PtrGroupDirectPool_Copy creates a new GroupDirect from PtrGroupDirect pool
func NewGroupDirect_PtrGroupDirectPool_Copy(payload []byte) GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	if cap(group.payload) < len(payload) {
		group.payload = make([]byte, len(payload))
	} else {
		group.payload = group.payload[:len(payload)]
	}
	copy(group.payload, payload)
	return *group
}

// NewPtrGroupDirect_BytesPool_Append creates a new PtrGroupDirect from Bytes pool using append
func NewPtrGroupDirect_BytesPool_Append(payload []byte) *GroupDirect {
	buf := bytesPool.Get().([]byte)
	buf = (buf)[:0]
	buf = append(buf, payload...)
	return &GroupDirect{
		payload: buf,
	}
}

// NewPtrGroupDirect_BytesPool_Copy creates a new PtrGroupDirect from Bytes pool using copy
func NewPtrGroupDirect_BytesPool_Copy(payload []byte) *GroupDirect {
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = (buf)[:len(payload)]
	}
	copy(buf, payload)
	return &GroupDirect{
		payload: buf,
	}
}

// NewPtrGroupDirect_PtrBytesPool_Append creates a new PtrGroupDirect from PtrBytes pool using append
func NewPtrGroupDirect_PtrBytesPool_Append(payload []byte) *GroupDirect {
	buf := ptrBytesPool.Get().(*[]byte)
	*buf = (*buf)[:0]
	*buf = append(*buf, payload...)
	return &GroupDirect{
		payload: *buf,
	}
}

// NewPtrGroupDirect_PtrBytesPool_Copy creates a new PtrGroupDirect from PtrBytes pool using copy
func NewPtrGroupDirect_PtrBytesPool_Copy(payload []byte) *GroupDirect {
	buf := ptrBytesPool.Get().(*[]byte)
	if cap(*buf) < len(payload) {
		*buf = make([]byte, len(payload))
	} else {
		*buf = (*buf)[:len(payload)]
	}
	copy(*buf, payload)
	return &GroupDirect{
		payload: *buf,
	}
}

// NewPtrGroupDirect_GroupDirectPool_Append creates a new PtrGroupDirect from GroupDirect pool using append
func NewPtrGroupDirect_GroupDirectPool_Append(payload []byte) *GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	group.payload = (group.payload)[:0]
	group.payload = append(group.payload, payload...)
	return &group
}

// NewPtrGroupDirect_GroupDirectPool_Copy creates a new PtrGroupDirect from GroupDirect pool using copy
func NewPtrGroupDirect_GroupDirectPool_Copy(payload []byte) *GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	if cap(group.payload) < len(payload) {
		group.payload = make([]byte, len(payload))
	} else {
		group.payload = group.payload[:len(payload)]
	}
	copy(group.payload, payload)
	return &group
}

// NewPtrGroupDirect_PtrGroupDirectPool_Append creates a new PtrGroupDirect from PtrGroupDirect pool using append
func NewPtrGroupDirect_PtrGroupDirectPool_Append(payload []byte) *GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	group.payload = group.payload[:0]
	group.payload = append(group.payload, payload...)
	return group
}

// NewPtrGroupDirect_PtrGroupDirectPool_Copy creates a new PtrGroupDirect from PtrGroupDirect pool using copy
func NewPtrGroupDirect_PtrGroupDirectPool_Copy(payload []byte) *GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	if cap(group.payload) < len(payload) {
		group.payload = make([]byte, len(payload))
	} else {
		group.payload = group.payload[:len(payload)]
	}
	copy(group.payload, payload)
	return group
}
func NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Copy(payload []byte) *GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	group.payload = buf
	return group
}

func NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Append(payload []byte) *GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	group.payload = buf
	return group
}

func NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Copy(payload []byte) *GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	group.payload = *ptr
	return group
}

func NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Append(payload []byte) *GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	group.payload = *ptr
	return group
}

// New constructor functions for PtrGroupDirect with combined pools
func NewPtrGroupDirect_GroupDirectAndBytesPool_Copy(payload []byte) *GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	group.payload = buf
	return &group
}

func NewPtrGroupDirect_GroupDirectAndBytesPool_Append(payload []byte) *GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	group.payload = buf
	return &group
}

func NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Copy(payload []byte) *GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	group.payload = *ptr
	return &group
}

func NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Append(payload []byte) *GroupDirect {
	group := groupDirectPool.Get().(GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	group.payload = *ptr
	return &group
}

// New constructor functions for combined GroupDirect and PtrGroupDirect pools
func NewGroupDirect_PtrGroupDirectAndBytes_Copy(payload []byte) GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	buf := bytesPool.Get().([]byte)
	if cap(buf) < len(payload) {
		buf = make([]byte, len(payload))
	} else {
		buf = buf[:len(payload)]
	}
	copy(buf, payload)
	group.payload = buf
	return *group
}

func NewGroupDirect_PtrGroupDirectAndBytes_Append(payload []byte) GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	buf := bytesPool.Get().([]byte)
	buf = buf[:0]
	buf = append(buf, payload...)
	group.payload = buf
	return *group
}

func NewGroupDirect_PtrGroupDirectAndPtrBytes_Copy(payload []byte) GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	if cap(*ptr) < len(payload) {
		*ptr = make([]byte, len(payload))
	} else {
		*ptr = (*ptr)[:len(payload)]
	}
	copy(*ptr, payload)
	group.payload = *ptr
	return *group
}

func NewGroupDirect_PtrGroupDirectAndPtrBytes_Append(payload []byte) GroupDirect {
	group := ptrGroupDirectPool.Get().(*GroupDirect)
	ptr := ptrBytesPool.Get().(*[]byte)
	*ptr = (*ptr)[:0]
	*ptr = append(*ptr, payload...)
	group.payload = *ptr
	return *group
}

// CopyPayload for GroupDirect copies the payload data
func (f *GroupDirect) CopyPayload() []byte {
	copyData := make([]byte, len(f.payload))
	copy(copyData, f.payload)
	return copyData
}

// Size for GroupDirect returns the size of the payload
func (f *GroupDirect) Size() int {
	return len(f.payload)
}

// ReleaseBytes releases the payload back to the Bytes pool
func (f *GroupDirect) ReleaseBytes() {
	f.payload = f.payload[:0]
	bytesPool.Put(f.payload)
}

// ReleaseGroupDirect releases the GroupDirect back to the GroupDirect pool
func (f *GroupDirect) ReleaseGroupDirect() {
	f.payload = f.payload[:0]
	groupDirectPool.Put(*f)
}

// ReleasePtrGroupDirect releases the GroupDirect back to the PtrGroupDirect pool
func (f *GroupDirect) ReleasePtrGroupDirect() {
	f.payload = f.payload[:0]
	ptrGroupDirectPool.Put(f)
}

func (f *GroupDirect) ReleasePtrBytes() {
	f.payload = f.payload[:0]
	ptrBytesPool.Put(&f.payload)
}

// First group - GroupDirect basic operations
func BenchmarkNewGroup_GroupDirect_BytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_BytesPool_Append(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_GroupDirect_BytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_BytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_GroupDirect_GroupDirectPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_GroupDirectPool_Append(defaultPayload)
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_GroupDirectPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_GroupDirectPool_Copy(defaultPayload)
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrGroupDirectPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrGroupDirectPool_Append(defaultPayload)
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrGroupDirectPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrGroupDirectPool_Copy(defaultPayload)
		group.ReleasePtrGroupDirect()
	}
}

// Second group - GroupDirect combined pools
func BenchmarkNewGroup_GroupDirect_GroupDirectAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_GroupDirectAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_GroupDirectAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_GroupDirectAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_GroupDirectAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_GroupDirectAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_GroupDirectAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_GroupDirectAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupDirect()
	}
}

// Third group - PtrGroupDirect basic operations
func BenchmarkNewGroup_PtrGroupDirect_BytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_BytesPool_Append(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_BytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_BytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_GroupDirectPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_GroupDirectPool_Append(defaultPayload)
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_GroupDirectPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_GroupDirectPool_Copy(defaultPayload)
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrGroupDirectPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrGroupDirectPool_Append(defaultPayload)
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrGroupDirectPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrGroupDirectPool_Copy(defaultPayload)
		group.ReleasePtrGroupDirect()
	}
}

// Fourth group - PtrGroupDirect combined pools
func BenchmarkNewGroup_PtrGroupDirect_PtrGroupDirectAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrGroupDirectAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_GroupDirectAndBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_GroupDirectAndBytesPool_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_GroupDirectAndBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_GroupDirectAndBytesPool_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_GroupDirectAndPtrBytesPool_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_PtrGroupDirect_GroupDirectAndPtrBytesPool_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleaseGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrGroupDirectAndBytes_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrGroupDirectAndBytes_Copy(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrGroupDirectAndBytes_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrGroupDirectAndBytes_Append(defaultPayload)
		group.ReleaseBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrGroupDirectAndPtrBytes_Copy(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrGroupDirectAndPtrBytes_Copy(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupDirect()
	}
}

func BenchmarkNewGroup_GroupDirect_PtrGroupDirectAndPtrBytes_Append(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		group := NewGroupDirect_PtrGroupDirectAndPtrBytes_Append(defaultPayload)
		group.ReleasePtrBytes()
		group.ReleasePtrGroupDirect()
	}
}

func TestNewFunctions(t *testing.T) {
	payload := []byte("test payload")

	tests := map[string]struct {
		fn func([]byte) interface{}
	}{
		"NewGroupDirect_BytesPool_Copy":                            {func(p []byte) interface{} { return NewGroupDirect_BytesPool_Copy(p) }},
		"NewGroupDirect_BytesPool_Append":                          {func(p []byte) interface{} { return NewGroupDirect_BytesPool_Append(p) }},
		"NewGroupDirect_PtrBytesPool_Copy":                         {func(p []byte) interface{} { return NewGroupDirect_PtrBytesPool_Copy(p) }},
		"NewGroupDirect_PtrBytesPool_Append":                       {func(p []byte) interface{} { return NewGroupDirect_PtrBytesPool_Append(p) }},
		"NewGroupDirect_GroupDirectPool_Copy":                      {func(p []byte) interface{} { return NewGroupDirect_GroupDirectPool_Copy(p) }},
		"NewGroupDirect_GroupDirectPool_Append":                    {func(p []byte) interface{} { return NewGroupDirect_GroupDirectPool_Append(p) }},
		"NewGroupPointer_BytesPool_Copy":                           {func(p []byte) interface{} { return NewGroupPointer_BytesPool_Copy(p) }},
		"NewGroupPointer_BytesPool_Append":                         {func(p []byte) interface{} { return NewGroupPointer_BytesPool_Append(p) }},
		"NewGroupPointer_PtrBytesPool_Copy":                        {func(p []byte) interface{} { return NewGroupPointer_PtrBytesPool_Copy(p) }},
		"NewGroupPointer_PtrBytesPool_Append":                      {func(p []byte) interface{} { return NewGroupPointer_PtrBytesPool_Append(p) }},
		"NewGroupPointer_GroupPointerPool_Copy":                    {func(p []byte) interface{} { return NewGroupPointer_GroupPointerPool_Copy(p) }},
		"NewGroupPointer_GroupPointerPool_Append":                  {func(p []byte) interface{} { return NewGroupPointer_GroupPointerPool_Append(p) }},
		"NewPtrGroupDirect_BytesPool_Append":                       {func(p []byte) interface{} { return NewPtrGroupDirect_BytesPool_Append(p) }},
		"NewPtrGroupDirect_BytesPool_Copy":                         {func(p []byte) interface{} { return NewPtrGroupDirect_BytesPool_Copy(p) }},
		"NewPtrGroupDirect_PtrBytesPool_Append":                    {func(p []byte) interface{} { return NewPtrGroupDirect_PtrBytesPool_Append(p) }},
		"NewPtrGroupDirect_PtrBytesPool_Copy":                      {func(p []byte) interface{} { return NewPtrGroupDirect_PtrBytesPool_Copy(p) }},
		"NewPtrGroupDirect_GroupDirectPool_Append":                 {func(p []byte) interface{} { return NewPtrGroupDirect_GroupDirectPool_Append(p) }},
		"NewPtrGroupDirect_GroupDirectPool_Copy":                   {func(p []byte) interface{} { return NewPtrGroupDirect_GroupDirectPool_Copy(p) }},
		"NewPtrGroupDirect_PtrGroupDirectPool_Append":              {func(p []byte) interface{} { return NewPtrGroupDirect_PtrGroupDirectPool_Append(p) }},
		"NewPtrGroupDirect_PtrGroupDirectPool_Copy":                {func(p []byte) interface{} { return NewPtrGroupDirect_PtrGroupDirectPool_Copy(p) }},
		"NewPtrGroupPointer_BytesPool_Append":                      {func(p []byte) interface{} { return NewPtrGroupPointer_BytesPool_Append(p) }},
		"NewPtrGroupPointer_BytesPool_Copy":                        {func(p []byte) interface{} { return NewPtrGroupPointer_BytesPool_Copy(p) }},
		"NewPtrGroupPointer_PtrBytesPool_Append":                   {func(p []byte) interface{} { return NewPtrGroupPointer_PtrBytesPool_Append(p) }},
		"NewPtrGroupPointer_PtrBytesPool_Copy":                     {func(p []byte) interface{} { return NewPtrGroupPointer_PtrBytesPool_Copy(p) }},
		"NewPtrGroupPointer_GroupPointerPool_Append":               {func(p []byte) interface{} { return NewPtrGroupPointer_GroupPointerPool_Append(p) }},
		"NewPtrGroupPointer_GroupPointerPool_Copy":                 {func(p []byte) interface{} { return NewPtrGroupPointer_GroupPointerPool_Copy(p) }},
		"NewPtrGroupPointer_PtrGroupPointerPool_Append":            {func(p []byte) interface{} { return NewPtrGroupPointer_PtrGroupPointerPool_Append(p) }},
		"NewPtrGroupPointer_PtrGroupPointerPool_Copy":              {func(p []byte) interface{} { return NewPtrGroupPointer_PtrGroupPointerPool_Copy(p) }},
		"NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Copy":      {func(p []byte) interface{} { return NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Copy(p) }},
		"NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Append":    {func(p []byte) interface{} { return NewPtrGroupPointer_PtrGroupPointerAndBytesPool_Append(p) }},
		"NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy":   {func(p []byte) interface{} { return NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Copy(p) }},
		"NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Append": {func(p []byte) interface{} { return NewPtrGroupPointer_PtrGroupPointerAndPtrBytesPool_Append(p) }},
		"NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Copy":        {func(p []byte) interface{} { return NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Copy(p) }},
		"NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Append":      {func(p []byte) interface{} { return NewPtrGroupDirect_PtrGroupDirectAndBytesPool_Append(p) }},
		"NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Copy":     {func(p []byte) interface{} { return NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Copy(p) }},
		"NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Append":   {func(p []byte) interface{} { return NewPtrGroupDirect_PtrGroupDirectAndPtrBytesPool_Append(p) }},
		"NewGroupDirect_GroupDirectAndBytesPool_Copy":              {func(p []byte) interface{} { return NewGroupDirect_GroupDirectAndBytesPool_Copy(p) }},
		"NewGroupDirect_GroupDirectAndBytesPool_Append":            {func(p []byte) interface{} { return NewGroupDirect_GroupDirectAndBytesPool_Append(p) }},
		"NewGroupDirect_GroupDirectAndPtrBytesPool_Copy":           {func(p []byte) interface{} { return NewGroupDirect_GroupDirectAndPtrBytesPool_Copy(p) }},
		"NewGroupDirect_GroupDirectAndPtrBytesPool_Append":         {func(p []byte) interface{} { return NewGroupDirect_GroupDirectAndPtrBytesPool_Append(p) }},
		"NewGroupDirect_PtrGroupDirectPool_Append":                 {func(p []byte) interface{} { return NewGroupDirect_PtrGroupDirectPool_Append(p) }},
		"NewGroupDirect_PtrGroupDirectPool_Copy":                   {func(p []byte) interface{} { return NewGroupDirect_PtrGroupDirectPool_Copy(p) }},
		"NewGroupPointer_PtrGroupPointerPool_Copy":                 {func(p []byte) interface{} { return NewGroupPointer_PtrGroupPointerPool_Copy(p) }},
		"NewGroupPointer_PtrGroupPointerPool_Append":               {func(p []byte) interface{} { return NewGroupPointer_PtrGroupPointerPool_Append(p) }},
		"NewPtrGroupDirect_GroupDirectAndBytesPool_Copy":           {func(p []byte) interface{} { return NewPtrGroupDirect_GroupDirectAndBytesPool_Copy(p) }},
		"NewPtrGroupDirect_GroupDirectAndBytesPool_Append":         {func(p []byte) interface{} { return NewPtrGroupDirect_GroupDirectAndBytesPool_Append(p) }},
		"NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Copy":        {func(p []byte) interface{} { return NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Copy(p) }},
		"NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Append":      {func(p []byte) interface{} { return NewPtrGroupDirect_GroupDirectAndPtrBytesPool_Append(p) }},
		"NewGroupDirect_PtrGroupDirectAndBytes_Copy":               {func(p []byte) interface{} { return NewGroupDirect_PtrGroupDirectAndBytes_Copy(p) }},
		"NewGroupDirect_PtrGroupDirectAndBytes_Append":             {func(p []byte) interface{} { return NewGroupDirect_PtrGroupDirectAndBytes_Append(p) }},
		"NewGroupDirect_PtrGroupDirectAndPtrBytes_Copy":            {func(p []byte) interface{} { return NewGroupDirect_PtrGroupDirectAndPtrBytes_Copy(p) }},
		"NewGroupDirect_PtrGroupDirectAndPtrBytes_Append":          {func(p []byte) interface{} { return NewGroupDirect_PtrGroupDirectAndPtrBytes_Append(p) }},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			group := tt.fn(payload)
			var groupPayload []byte

			switch f := group.(type) {
			case GroupDirect:
				groupPayload = f.payload
			case *GroupDirect:
				groupPayload = f.payload
			case GroupPointer:
				groupPayload = *f.payload
			case *GroupPointer:
				groupPayload = *f.payload
			default:
				t.Fatalf("unexpected group type: %T", group)
			}

			if !bytes.Equal(groupPayload, payload) {
				t.Errorf("payload mismatch: got %v, want %v", groupPayload, payload)
			}
		})
	}
}
