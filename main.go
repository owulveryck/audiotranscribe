package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type configuration struct {
	GCPProject  string `envconfig:"GCP_PROJECT" required:"true"`
	GeminiModel string `envconfig:"GEMINI_MODEL" default:"gemini-2.0-flash"`
	GCPRegion   string `envconfig:"GCP_REGION" default:"europe-west9"`
}

var logger *slog.Logger

func main() {
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

	var config configuration
	err := envconfig.Process("", &config)
	if err != nil {
		logger.Error("failed to process environment variables", "error", err)
		envconfig.Usage("", &config)
		os.Exit(1)
	}

	var (
		outputFile = flag.String("o", "", "Path to the output file. If empty, stdout will be used.")
		help       = flag.Bool("h", false, "Help")
	)
	flag.Parse()

	if *help {
		envconfig.Usage("", &config)
		flag.Usage()
		os.Exit(1)
	}

	// Get audio files from positional arguments
	filePaths := flag.Args()
	if len(filePaths) == 0 {
		logger.Error("at least one audio file required as argument")
		fmt.Fprintf(os.Stderr, "Usage: %s [-o output.md] audio1.m4a [audio2.m4a ...]\n", os.Args[0])
		flag.Usage()
		os.Exit(1)
	}

	// Determine the output writer.
	var outputWriter io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			logger.Error("failed to create output file", "error", err)
			os.Exit(1)
		}
		defer f.Close()
		outputWriter = f
	}

	// Transcribe all audio files using Vertex AI.
	logger.Info("transcribing audio files", "count", len(filePaths))
	var allTranscripts []string
	
	for i, audioFilePath := range filePaths {
		logger.Info("transcribing audio file", "file", audioFilePath, "progress", fmt.Sprintf("%d/%d", i+1, len(filePaths)))
		
		transcript, err := transcribeAudio(outputWriter, config.GCPProject, config.GCPRegion, config.GeminiModel, audioFilePath)
		if err != nil {
			logger.Error("failed to transcribe audio file", "file", audioFilePath, "error", err)
			os.Exit(1)
		}
		
		allTranscripts = append(allTranscripts, transcript)
		logger.Info("audio file transcribed successfully", "file", audioFilePath)
	}

	// Combine all transcripts
	combinedTranscript := strings.Join(allTranscripts, "\n\n---\n\n")
	
	err = postProcess(combinedTranscript, outputWriter, config.GCPProject, config.GCPRegion, config.GeminiModel)
	if err != nil {
		logger.Error("failed to do the post-processing", "error", err)
		os.Exit(1)
	}
	logger.Info("post processing completed successfully")
}
