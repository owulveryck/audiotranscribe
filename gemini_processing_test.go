package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestBufferedWriterFlushing tests if buffered writer properly flushes between writes
func TestBufferedWriterFlushing(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	// Simulate writing three transcripts like we do in the real code
	transcripts := []string{
		"Transcript 1: This is the first audio chunk",
		"Transcript 2: This is the second audio chunk",
		"Transcript 3: This is the third audio chunk",
	}

	for i, transcript := range transcripts {
		// Write the transcript header and content
		io.WriteString(bufWriter, "Generated transcript for chunk_"+string(rune('0'+i))+".m4a:\n")
		io.WriteString(bufWriter, transcript+"\n\n")

		// Try to flush like we do in gemini_processing.go:98-101
		if bw, ok := interface{}(bufWriter).(*bufio.Writer); ok {
			if err := bw.Flush(); err != nil {
				t.Fatalf("failed to flush: %v", err)
			}
		}

		// Check what's in the buffer so far
		currentContent := buf.String()
		if !strings.Contains(currentContent, transcript) {
			t.Errorf("After writing transcript %d, content not found in buffer.\nExpected to contain: %s\nGot: %s",
				i, transcript, currentContent)
		}
	}

	// Final check
	finalContent := buf.String()
	for i, transcript := range transcripts {
		if !strings.Contains(finalContent, transcript) {
			t.Errorf("Transcript %d not found in final output: %s", i, transcript)
		}
	}
}

// TestWriterInterfaceFlush tests if type assertion on io.Writer interface works
func TestWriterInterfaceFlush(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	// Pass as io.Writer interface (like we do in transcribeAudio)
	var writer io.Writer = bufWriter

	// Write some content
	io.WriteString(writer, "Test content\n")

	// Try the type assertion that's used in gemini_processing.go:98
	if bw, ok := writer.(*bufio.Writer); ok {
		t.Log("Type assertion succeeded")
		if err := bw.Flush(); err != nil {
			t.Fatalf("failed to flush: %v", err)
		}
	} else {
		t.Error("Type assertion failed - this is the bug!")
	}

	content := buf.String()
	if !strings.Contains(content, "Test content") {
		t.Errorf("Content not found after flush. Got: %s", content)
	}
}

// TestMultipleFlushes tests multiple flush operations in sequence
func TestMultipleFlushes(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	writes := []string{"First\n", "Second\n", "Third\n"}

	for i, content := range writes {
		if _, err := bufWriter.WriteString(content); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}

		// Flush after each write
		if err := bufWriter.Flush(); err != nil {
			t.Fatalf("flush %d failed: %v", i, err)
		}

		// Verify content is in buffer
		current := buf.String()
		if !strings.Contains(current, content) {
			t.Errorf("After write %d, content not found. Expected: %s, Got: %s", i, content, current)
		}
	}
}

// TestEmptyWriteBeforeFlush tests what happens with empty writes
func TestEmptyWriteBeforeFlush(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	// Write empty string
	if _, err := bufWriter.WriteString(""); err != nil {
		t.Fatalf("empty write failed: %v", err)
	}

	// Flush
	if err := bufWriter.Flush(); err != nil {
		t.Fatalf("flush after empty write failed: %v", err)
	}

	// Write actual content
	if _, err := bufWriter.WriteString("Content\n"); err != nil {
		t.Fatalf("content write failed: %v", err)
	}

	// Flush again
	if err := bufWriter.Flush(); err != nil {
		t.Fatalf("second flush failed: %v", err)
	}

	content := buf.String()
	if content != "Content\n" {
		t.Errorf("Expected 'Content\\n', got: %q", content)
	}
}

// TestRealisticMainFlow simulates the exact flow from main.go
func TestRealisticMainFlow(t *testing.T) {
	// Create a temporary file to simulate output file
	tmpFile, err := os.CreateTemp("", "test-output-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Create buffered writer exactly like main.go:65-66
	bufWriter := bufio.NewWriter(tmpFile)
	var outputWriter io.Writer = bufWriter

	// Simulate transcribing multiple files
	transcripts := []string{
		"Speaker A: This is chunk zero content that should appear first.",
		"Speaker B: This is chunk one content that should appear second.",
		"Speaker A: This is chunk two content that should appear third.",
	}

	var allTranscripts []string

	for i, transcript := range transcripts {
		audioFilePath := fmt.Sprintf("/tmp/chunk_%03d.m4a", i)

		// Simulate transcribeAudio function writing to outputWriter
		if _, err := fmt.Fprintf(outputWriter, "Generated transcript for %s:\n%s\n\n", audioFilePath, transcript); err != nil {
			t.Fatalf("failed to write transcript %d: %v", i, err)
		}

		// Flush after each transcript like main.go:84-89
		if err := bufWriter.Flush(); err != nil {
			t.Fatalf("failed to flush after transcript %d: %v", i, err)
		}

		allTranscripts = append(allTranscripts, transcript)

		// Read file content so far
		currentContent, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("failed to read temp file after transcript %d: %v", i, err)
		}

		// Verify all previous transcripts are present
		for j := 0; j <= i; j++ {
			if !strings.Contains(string(currentContent), transcripts[j]) {
				t.Errorf("After writing transcript %d, transcript %d not found in file.\nExpected: %s\nFile content:\n%s",
					i, j, transcripts[j], string(currentContent))
			}
		}
	}

	// Final flush before reading (like main.go:67)
	if err := bufWriter.Flush(); err != nil {
		t.Fatalf("final flush failed: %v", err)
	}

	// Read final content
	finalContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read final content: %v", err)
	}

	// Verify all transcripts are in final file
	for i, transcript := range transcripts {
		if !strings.Contains(string(finalContent), transcript) {
			t.Errorf("Transcript %d missing from final output: %s\nFinal content:\n%s",
				i, transcript, string(finalContent))
		}
	}
}

// TestConcurrentWriteAndRead tests race conditions with concurrent writes and reads
func TestConcurrentWriteAndRead(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)
	var mu sync.Mutex

	var wg sync.WaitGroup
	numWrites := 5

	// Writer goroutines
	for i := 0; i < numWrites; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			content := fmt.Sprintf("Content from writer %d\n", id)

			mu.Lock()
			bufWriter.WriteString(content)
			bufWriter.Flush()
			mu.Unlock()

			time.Sleep(10 * time.Millisecond)
		}(i)
	}

	// Reader goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numWrites; i++ {
			time.Sleep(15 * time.Millisecond)
			mu.Lock()
			_ = buf.String()
			mu.Unlock()
		}
	}()

	wg.Wait()

	// Verify all content is present
	finalContent := buf.String()
	for i := 0; i < numWrites; i++ {
		expected := fmt.Sprintf("Content from writer %d", i)
		if !strings.Contains(finalContent, expected) {
			t.Errorf("Missing content from writer %d in final output", i)
		}
	}
}

// TestSequentialWriteWithIntermediateReads simulates writing and reading file during processing
func TestSequentialWriteWithIntermediateReads(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-sequential-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	bufWriter := bufio.NewWriter(tmpFile)

	transcripts := []string{
		"First transcript with important content",
		"Second transcript with more content",
		"Third transcript with final content",
	}

	for i, transcript := range transcripts {
		// Write
		_, err := fmt.Fprintf(bufWriter, "Transcript %d:\n%s\n\n", i, transcript)
		if err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}

		// Flush
		if err := bufWriter.Flush(); err != nil {
			t.Fatalf("flush %d failed: %v", i, err)
		}

		// Simulate external process reading file (like shell script might do)
		content, err := os.ReadFile(tmpFile.Name())
		if err != nil {
			t.Fatalf("read after write %d failed: %v", i, err)
		}

		// Check that current and all previous transcripts are present
		for j := 0; j <= i; j++ {
			if !strings.Contains(string(content), transcripts[j]) {
				t.Errorf("After write %d, transcript %d missing from file\nExpected: %s\nGot:\n%s",
					i, j, transcripts[j], string(content))
			}
		}
	}
}

// TestBufferSizeIssues tests if buffer size affects content visibility
func TestBufferSizeIssues(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
		content    string
	}{
		{"Small buffer small content", 16, "Short"},
		{"Small buffer large content", 16, strings.Repeat("A", 100)},
		{"Default buffer small content", 4096, "Short"},
		{"Default buffer large content", 4096, strings.Repeat("B", 10000)},
		{"Large buffer large content", 65536, strings.Repeat("C", 100000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			bufWriter := bufio.NewWriterSize(&buf, tt.bufferSize)

			// Write content
			if _, err := bufWriter.WriteString(tt.content); err != nil {
				t.Fatalf("write failed: %v", err)
			}

			// Check before flush
			beforeFlush := buf.String()
			beforeLen := len(beforeFlush)

			// Flush
			if err := bufWriter.Flush(); err != nil {
				t.Fatalf("flush failed: %v", err)
			}

			// Check after flush
			afterFlush := buf.String()
			afterLen := len(afterFlush)

			if afterLen != len(tt.content) {
				t.Errorf("Content length mismatch. Expected %d, got %d (before flush: %d)",
					len(tt.content), afterLen, beforeLen)
			}

			if afterFlush != tt.content {
				t.Errorf("Content mismatch after flush")
			}
		})
	}
}

// TestFileWriteOrderPreservation tests that write order is preserved across flushes
func TestFileWriteOrderPreservation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-order-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	bufWriter := bufio.NewWriter(tmpFile)

	// Write in specific order
	for i := 0; i < 10; i++ {
		line := fmt.Sprintf("Line %d\n", i)
		if _, err := bufWriter.WriteString(line); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
		if err := bufWriter.Flush(); err != nil {
			t.Fatalf("flush %d failed: %v", i, err)
		}
	}

	// Read and verify order
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for i := 0; i < 10; i++ {
		expected := fmt.Sprintf("Line %d", i)
		if i >= len(lines) || lines[i] != expected {
			t.Errorf("Line %d mismatch. Expected %q, got %q", i, expected, lines[i])
		}
	}
}

// TestWriterWrapping tests various ways writers can be wrapped
func TestWriterWrapping(t *testing.T) {
	var buf bytes.Buffer

	// Create chain: bytes.Buffer -> bufio.Writer -> io.Writer interface
	bufWriter := bufio.NewWriter(&buf)
	var writer io.Writer = bufWriter

	content := "Test content that should survive wrapping"

	// Write through interface
	if _, err := io.WriteString(writer, content); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Flush through type assertion
	if bw, ok := writer.(*bufio.Writer); ok {
		if err := bw.Flush(); err != nil {
			t.Fatalf("flush failed: %v", err)
		}
	} else {
		t.Fatal("type assertion failed")
	}

	// Verify
	result := buf.String()
	if result != content {
		t.Errorf("Content mismatch. Expected %q, got %q", content, result)
	}
}

// TestExactShellScriptFlow simulates EXACTLY what split_and_transcribe.sh does
func TestExactShellScriptFlow(t *testing.T) {
	// Create temp output file EXACTLY like the shell script does
	tmpDir, err := os.MkdirTemp("", "test-shell-flow-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outputFile := tmpDir + "/output.md"

	// Simulate what happens in main.go when -o flag is provided
	f, err := os.Create(outputFile)
	if err != nil {
		t.Fatalf("failed to create output file: %v", err)
	}

	bufWriter := bufio.NewWriter(f)
	var outputWriter io.Writer = bufWriter

	// Simulate transcribing multiple chunk files
	// This is EXACTLY what happens when shell passes all chunks to one invocation
	transcripts := []struct {
		filename string
		content  string
	}{
		{"/tmp/chunk_000.m4a", ""}, // EMPTY - this is the issue!
		{"/tmp/chunk_001.m4a", ""}, // EMPTY
		{"/tmp/chunk_002.m4a", ""}, // EMPTY
		{"/tmp/chunk_003.m4a", "Speaker A: This should appear"},
	}

	var allTranscripts []string

	for i, tc := range transcripts {
		// Simulate transcribeAudio writing
		transcriptText := tc.content
		if _, err := fmt.Fprintf(outputWriter, "Generated transcript for %s:\n%s\n\n", tc.filename, transcriptText); err != nil {
			t.Fatalf("failed to write transcript %d: %v", i, err)
		}

		// Flush inside transcribeAudio (gemini_processing.go:98)
		if bw, ok := outputWriter.(*bufio.Writer); ok {
			if err := bw.Flush(); err != nil {
				t.Fatalf("flush inside transcribeAudio failed: %v", err)
			}
		}

		allTranscripts = append(allTranscripts, transcriptText)

		// Flush in main.go:84-88
		if err := bufWriter.Flush(); err != nil {
			t.Fatalf("flush in main failed: %v", err)
		}
	}

	// Final defer flush (main.go:67)
	if err := bufWriter.Flush(); err != nil {
		t.Fatalf("final defer flush failed: %v", err)
	}

	// Close file (main.go:64 defer)
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close file: %v", err)
	}

	// NOW read the file and check what we got
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	t.Logf("File content:\n%s", string(content))

	// Verify the pattern matches what user sees
	if !strings.Contains(string(content), "Generated transcript for /tmp/chunk_000.m4a:") {
		t.Error("Missing header for chunk_000")
	}
	if !strings.Contains(string(content), "Generated transcript for /tmp/chunk_003.m4a:") {
		t.Error("Missing header for chunk_003")
	}
	if !strings.Contains(string(content), "Speaker A: This should appear") {
		t.Error("Missing content from chunk_003")
	}

	// The key test: verify that empty transcripts produce the pattern user sees
	// Between "chunk_000" and "chunk_001", there should only be empty lines
	chunk000Idx := strings.Index(string(content), "chunk_000.m4a:")
	chunk001Idx := strings.Index(string(content), "chunk_001.m4a:")

	if chunk000Idx == -1 || chunk001Idx == -1 {
		t.Fatal("Could not find chunk markers")
	}

	betweenChunks := string(content)[chunk000Idx+len("chunk_000.m4a:") : chunk001Idx]
	betweenChunks = strings.TrimSpace(betweenChunks)

	if betweenChunks != "" && betweenChunks != "Generated transcript for /tmp/" {
		t.Errorf("Expected empty content between chunks, got: %q", betweenChunks)
	}
}

// TestEmptyVsNonEmptyContent tests the specific pattern user is seeing
func TestEmptyVsNonEmptyContent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-empty-content-*.md")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	bufWriter := bufio.NewWriter(tmpFile)

	// Write pattern exactly as user sees it
	items := []struct {
		name    string
		content string
	}{
		{"chunk_000.m4a", ""},
		{"chunk_001.m4a", ""},
		{"chunk_002.m4a", ""},
		{"chunk_003.m4a", "Speaker A: avait écrit en gros"},
	}

	for _, item := range items {
		fmt.Fprintf(bufWriter, "Generated transcript for %s:\n%s\n\n", item.name, item.content)
		bufWriter.Flush()
	}

	bufWriter.Flush()
	tmpFile.Sync() // Force OS to write to disk

	// Read back
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	t.Logf("Output:\n%s", string(content))

	// This should produce EXACTLY what the user sees
	expected := `Generated transcript for chunk_000.m4a:


Generated transcript for chunk_001.m4a:


Generated transcript for chunk_002.m4a:


Generated transcript for chunk_003.m4a:
Speaker A: avait écrit en gros

`
	if string(content) != expected {
		t.Errorf("Output doesn't match expected pattern.\nExpected:\n%s\nGot:\n%s", expected, string(content))
	}
}
