package moqt

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"
)

func TestGroupBuffer(t *testing.T) {
	t.Run("basic operations", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024)

		// Test WriteFrame
		testData := []byte("test frame")
		err := gb.WriteFrame(testData)
		if err != nil {
			t.Fatalf("WriteFrame failed: %v", err)
		}

		// Test ReadFrame
		frame, err := gb.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame failed: %v", err)
		}
		if !bytes.Equal(frame, testData) {
			t.Errorf("ReadFrame returned wrong data. got: %v, want: %v", frame, testData)
		}
	})

	t.Run("multiple frames", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024)
		frames := [][]byte{
			[]byte("frame1"),
			[]byte("frame2"),
			[]byte("frame3"),
		}

		// Write multiple frames
		for _, frame := range frames {
			err := gb.WriteFrame(frame)
			if err != nil {
				t.Fatalf("WriteFrame failed: %v", err)
			}
		}

		// Read and verify all frames
		for _, expected := range frames {
			frame, err := gb.ReadFrame()
			if err != nil {
				t.Fatalf("ReadFrame failed: %v", err)
			}
			if !bytes.Equal(frame, expected) {
				t.Errorf("ReadFrame returned wrong data. got: %v, want: %v", frame, expected)
			}
		}
	})

	t.Run("close behavior", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024)

		err := gb.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Write after close should fail
		err = gb.WriteFrame([]byte("test"))
		if err != ErrGroupClosed {
			t.Errorf("Expected ErrGroupClosed, got: %v", err)
		}

		// Read after close should return EOF
		_, err = gb.ReadFrame()
		if err != io.EOF {
			t.Errorf("Expected EOF, got: %v", err)
		}
	})

	t.Run("reset behavior", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024)

		// Write initial data
		err := gb.WriteFrame([]byte("initial"))
		if err != nil {
			t.Fatalf("WriteFrame failed: %v", err)
		}

		// Reset with new sequence
		newSeq := GroupSequence(2)
		gb.Reset(newSeq)

		if gb.GroupSequence() != newSeq {
			t.Errorf("Wrong sequence after reset. got: %v, want: %v", gb.GroupSequence(), newSeq)
		}

		// Write new data after reset
		err = gb.WriteFrame([]byte("after reset"))
		if err != nil {
			t.Errorf("WriteFrame after reset failed: %v", err)
		}
	})
}

func TestRelay(t *testing.T) {
	t.Run("buffer to buffer relay", func(t *testing.T) {
		src := NewGroupBuffer(GroupSequence(1), 1024)
		dst := NewGroupBuffer(GroupSequence(1), 1024)

		testData := []byte("test data")
		err := src.WriteFrame(testData)
		if err != nil {
			t.Fatalf("WriteFrame to source failed: %v", err)
		}

		// Close source after writing
		src.Close()

		// Start relay and wait for completion
		err = RelayGroup(src, dst)
		if err != nil {
			t.Fatalf("Relay failed: %v", err)
		}

		// Verify data was relayed correctly
		frame, err := dst.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame from destination failed: %v", err)
		}
		if !bytes.Equal(frame, testData) {
			t.Errorf("Relayed data doesn't match. got: %v, want: %v", frame, testData)
		}
	})

	t.Run("large data relay", func(t *testing.T) {
		src := NewGroupBuffer(GroupSequence(1), 1024*1024)
		dst := NewGroupBuffer(GroupSequence(1), 1024*1024)

		// Create large test data
		largeData := make([]byte, 1024*512) // 512KB
		for i := range largeData {
			largeData[i] = byte(i % 256)
		}

		err := src.WriteFrame(largeData)
		if err != nil {
			t.Fatalf("WriteFrame large data failed: %v", err)
		}

		done := make(chan error)
		go func() {
			done <- RelayGroup(src, dst)
		}()

		src.Close()
		if err := <-done; err != nil {
			t.Fatalf("Large data relay failed: %v", err)
		}

		frame, err := dst.ReadFrame()
		if err != nil {
			t.Fatalf("ReadFrame large data failed: %v", err)
		}
		if !bytes.Equal(frame, largeData) {
			t.Error("Large data relay corrupted")
		}
	})

	t.Run("invalid relay", func(t *testing.T) {
		err := RelayGroup(nil, nil)
		if err == nil {
			t.Error("Expected error for nil reader/writer, got nil")
		}

		src := NewGroupBuffer(GroupSequence(1), 1024)
		dst := NewGroupBuffer(GroupSequence(2), 1024)
		err = RelayGroup(src, dst)
		if err == nil {
			t.Error("Expected error for different sequences, got nil")
		}
	})
}

func TestGroupBufferConcurrency(t *testing.T) {
	t.Run("concurrent read/write", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024*1024)

		const numGoroutines = 10
		const numFrames = 100

		done := make(chan bool)
		var receivedMu sync.Mutex
		received := make(map[string]bool)

		// Start writer goroutines
		for i := 0; i < numGoroutines; i++ {
			go func(n int) {
				defer func() { done <- true }()

				for j := 0; j < numFrames; j++ {
					data := []byte{byte(n), byte(j)}
					err := gb.WriteFrame(data)
					if err != nil {
						t.Errorf("WriteFrame failed: %v", err)
						return
					}
				}
			}(i)
		}

		// Start reader goroutines
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer func() { done <- true }()

				for j := 0; j < numFrames; j++ {
					frame, err := gb.ReadFrame()
					if err != nil {
						t.Errorf("ReadFrame failed: %v", err)
						return
					}

					receivedMu.Lock()
					received[string(frame)] = true
					receivedMu.Unlock()
				}
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines*2; i++ {
			<-done
		}

		if len(received) != numGoroutines*numFrames {
			t.Errorf("Expected %d unique frames, got %d", numGoroutines*numFrames, len(received))
		}
	})

	t.Run("concurrent close", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024)
		const numGoroutines = 5

		// Start multiple goroutines trying to close
		var wg sync.WaitGroup
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := gb.Close()
				if err != nil && err != ErrGroupClosed {
					t.Errorf("Unexpected error on close: %v", err)
				}
			}()
		}

		wg.Wait()
	})

	t.Run("read timeout", func(t *testing.T) {
		gb := NewGroupBuffer(GroupSequence(1), 1024)

		// Try to read with timeout
		done := make(chan struct{})
		go func() {
			_, err := gb.ReadFrame()
			if err != io.EOF {
				t.Errorf("Expected EOF, got: %v", err)
			}
			close(done)
		}()

		// Close after short delay
		time.Sleep(100 * time.Millisecond)
		gb.Close()

		select {
		case <-done:
			// Success
		case <-time.After(time.Second):
			t.Error("Read timeout test failed")
		}
	})
}
