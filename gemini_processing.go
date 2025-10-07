package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"cloud.google.com/go/vertexai/genai"
)

func postProcess(input string, w io.Writer, projectID, location, modelName string) error {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)

	// Optional: set an explicit temperature
	model.SetTemperature(0.4)
	res, err := model.GenerateContent(ctx, genai.Text(SummaryPrompt), genai.Text(input))
	if err != nil {
		return fmt.Errorf("unable to generate contents: %w", err)
	}

	if len(res.Candidates) == 0 ||
		len(res.Candidates[0].Content.Parts) == 0 {
		return errors.New("empty response from model")
	}
	logger.Info("Usage Metadata", "Prompt Token", res.UsageMetadata.PromptTokenCount, "Candidates Token", res.UsageMetadata.CandidatesTokenCount, "Total Token", res.UsageMetadata.TotalTokenCount)
	logger.Info("Finish", "Finished Reason", res.Candidates[0].FinishReason, "Finish Message", res.Candidates[0].FinishMessage)

	fmt.Fprintf(w, "\n\nSynthesis:\n%s\n", res.Candidates[0].Content.Parts[0])
	return nil
}

// transcribeAudio transcribes an audio file and returns the transcript text
func transcribeAudio(w io.Writer, projectID, location, modelName, audioFilePath string) (string, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return "", fmt.Errorf("unable to create client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(modelName)

	// Optional: set an explicit temperature
	model.SetTemperature(0.4)

	// Read audio file into memory
	audioData, err := os.ReadFile(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	// Create Blob with audio data
	audio := genai.Blob{
		MIMEType: mime.TypeByExtension(filepath.Ext(audioFilePath)),
		Data:     audioData,
	}
	logger.Info("Audio info", "mimetype", audio.MIMEType, "size", len(audioData), "file", audioFilePath)

	res, err := model.GenerateContent(ctx, audio, genai.Text(TranscriptionPrompt))
	if err != nil {
		return "", fmt.Errorf("unable to generate contents: %w", err)
	}

	if len(res.Candidates) == 0 ||
		len(res.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("empty response from model")
	}
	logger.Info("Usage Metadata", "Prompt Token", res.UsageMetadata.PromptTokenCount, "Candidates Token", res.UsageMetadata.CandidatesTokenCount, "Total Token", res.UsageMetadata.TotalTokenCount)
	logger.Info("Finish", "Finished Reason", res.Candidates[0].FinishReason, "Finish Message", res.Candidates[0].FinishMessage)
	logger.Info("Response parts", "num_parts", len(res.Candidates[0].Content.Parts), "part_type", fmt.Sprintf("%T", res.Candidates[0].Content.Parts[0]))

	transcriptText := string(res.Candidates[0].Content.Parts[0].(genai.Text))
	logger.Info("Transcript length", "length", len(transcriptText), "file", audioFilePath)

	// Check if transcript is empty
	if len(transcriptText) == 0 {
		logger.Warn("received empty transcript from Gemini", "file", audioFilePath)
	}

	if _, err := fmt.Fprintf(w, "Generated transcript for %s:\n%s\n\n", audioFilePath, transcriptText); err != nil {
		return "", fmt.Errorf("failed to write transcript: %w", err)
	}

	// Flush the buffer if the writer is a buffered writer
	if bw, ok := w.(*bufio.Writer); ok {
		if err := bw.Flush(); err != nil {
			return "", fmt.Errorf("failed to flush output: %w", err)
		}
	}

	return transcriptText, nil
}
