package moqtrelay_test

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/OkutaniDaichi0106/gomoqt/moqt"
	moqtrelay "github.com/OkutaniDaichi0106/gomoqt/moqt_relay"
	"github.com/quic-go/quic-go/quicvarint"
	"github.com/stretchr/testify/assert"
)

func TestGroupBuffer_WriteAndRead(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	testData := []byte("hello world")

	// Test WriteFrame
	wroteN, err := gb.WriteFrame(testData)
	assert.NoError(t, err)

	// Test ReadFrame
	data, readN, err := gb.ReadFrame()
	assert.NoError(t, err)
	assert.Equal(t, testData, data)

	// Test number of bytes written and read
	assert.Equal(t, readN, wroteN)

	// Test WriteFrame with empty data
	emptyData := []byte{}
	wroteN, err = gb.WriteFrame(emptyData)
	assert.NoError(t, err)
	assert.Equal(t, quicvarint.Len(0), wroteN)
}

func TestGroupBuffer_SingleReader(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	reader := moqtrelay.NewGroupReader(gb)

	testData := []byte("hello world")

	// Write data
	_, err := gb.WriteFrame(testData)
	assert.NoError(t, err)

	// Read data
	data, _, err := reader.ReadFrame()
	assert.NoError(t, err)

	// Verify data
	assert.Equal(t, testData, data)
}

func TestGroupBuffer_MultipleReaders(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	const numReaders = 3
	const numMessages = 5

	var errChan = make(chan error, numReaders)
	var wg sync.WaitGroup
	wg.Add(numReaders)

	// Initialize readers
	readers := make([]moqt.GroupReader, numReaders)
	for i := 0; i < numReaders; i++ {
		readers[i] = moqtrelay.NewGroupReader(gb)
	}

	// Start reader goroutines
	for i := 0; i < numReaders; i++ {
		go func(reader moqt.GroupReader, id int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				data, _, err := reader.ReadFrame()
				if err != nil {
					errChan <- fmt.Errorf("reader %d failed: %v", id, err)
					return
				}

				expectedData := []byte(fmt.Sprintf("message-%d", j))
				if !bytes.Equal(data, expectedData) {
					errChan <- fmt.Errorf("reader %d: got %s, want %s", id, data, expectedData)
					return
				}
			}
		}(readers[i], i)
	}

	// Write messages to buffer
	for i := 0; i < numMessages; i++ {
		msg := []byte(fmt.Sprintf("message-%d", i))
		_, err := gb.WriteFrame(msg)
		if err != nil {
			t.Fatalf("Failed to write message %d: %v", i, err)
		}
		time.Sleep(time.Millisecond) // Small delay between writes
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test completed successfully
	case err := <-errChan:
		t.Fatal(err)
	case <-time.After(5 * time.Second):
		t.Fatal("test timed out")
	}
}

func TestGroupBuffer_Concurrent(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	const numWriters = 1
	const numReaders = 3
	const messagesPerWriter = 10

	// Create channels for synchronization
	writersDone := make(chan struct{})
	var errChan = make(chan error, numWriters+numReaders)
	var wg sync.WaitGroup
	wg.Add(numWriters + numReaders)

	// Multiple writer goroutines
	go func() {
		defer close(writersDone)
		for i := 0; i < messagesPerWriter; i++ {
			msg := []byte(fmt.Sprintf("msg-%d", i))
			if _, err := gb.WriteFrame(msg); err != nil {
				errChan <- fmt.Errorf("writer failed: %v", err)
				return
			}
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}()

	// Multiple reader goroutines
	receivedCount := &sync.Map{}
	for i := 0; i < numReaders; i++ {
		reader := moqtrelay.NewGroupReader(gb)
		go func(readerID int) {
			defer wg.Done()
			count := 0
			for {
				_, _, err := reader.ReadFrame()
				if err == moqt.ErrGroupClosed {
					receivedCount.Store(readerID, count)
					return
				}
				if err != nil {
					errChan <- fmt.Errorf("reader %d failed: %v", readerID, err)
					return
				}
				count++
			}
		}(i)
	}

	// Wait for writers to complete before closing
	<-writersDone
	gb.Close()

	// Wait with reasonable timeout
	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-doneChan:
		// Success case
	case err := <-errChan:
		t.Fatal(err)
	case <-time.After(2 * time.Second):
		t.Fatal("test timed out")
	}

	// Verify results
	expectedMessages := messagesPerWriter
	receivedCount.Range(func(key, value interface{}) bool {
		count := value.(int)
		if count != expectedMessages {
			t.Errorf("Reader %v received %d messages, expected %d",
				key, count, expectedMessages)
		}
		return true
	})
}

func TestGroupBuffer_LargeMessages(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	reader := moqtrelay.NewGroupReader(gb)

	largeData := make([]byte, 1<<20) // 1MB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	_, err := gb.WriteFrame(largeData)
	assert.NoError(t, err)

	received, _, err := reader.ReadFrame()
	assert.NoError(t, err)
	assert.Equal(t, largeData, received)
}

func TestGroupBuffer_CloseOperation(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	testData := []byte("test data")

	_, err := gb.WriteFrame(testData)
	assert.NoError(t, err)

	assert.NoError(t, gb.Close())

	// Verify write after close fails
	_, err = gb.WriteFrame(testData)
	assert.Equal(t, moqt.ErrGroupClosed, err)

	// Verify existing data can still be read
	data, _, err := gb.ReadFrame()
	assert.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestGroupReader_CloseAndReopen(t *testing.T) {
	gb := moqtrelay.NewGroupBuffer(0)
	reader := moqtrelay.NewGroupReader(gb)

	_, _, err := reader.ReadFrame()
	assert.Equal(t, moqt.ErrGroupClosed, err)

	reader2 := moqtrelay.NewGroupReader(gb)
	testData := []byte("new data")

	_, err = gb.WriteFrame(testData)
	assert.NoError(t, err)

	data, _, err := reader2.ReadFrame()
	assert.NoError(t, err)
	assert.Equal(t, testData, data)
}
