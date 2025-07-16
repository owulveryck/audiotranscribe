# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based audio transcription tool that uses Google Cloud Platform services (Vertex AI Gemini) to transcribe audio files and generate summaries. The application processes audio files directly from local storage, transcribes them using Gemini models, and provides structured output with timecoded speaker identification. Supports processing multiple audio files at once with combined summary generation.

## Core Architecture

The application consists of four main components:

- `main.go` - CLI interface, configuration handling, and workflow orchestration
- `gemini_processing.go` - Vertex AI Gemini integration for transcription and post-processing
- `prompts.go` - Contains the AI prompts for transcription and summarization
- `gcp_plumbing.go` - Google Cloud Storage operations (currently unused but provides upload/delete functionality)

Additional utility:
- `split_and_transcribe.sh` - Shell script for processing large audio files by splitting them into 25-minute chunks

The workflow: read local audio file(s) → transcribe using Gemini → combine transcripts → generate unified summary.

## Environment Configuration

Required environment variables:
- `GCP_PROJECT` - Google Cloud project ID

Optional environment variables:
- `GEMINI_MODEL` - Gemini model to use (default: "gemini-2.0-flash")
- `GCP_REGION` - GCP region (default: "europe-west9")

## Build and Run Commands

Build the application:
```bash
go build -o audiotranscribe .
```

Run the application:
```bash
# Single file
./audiotranscribe [-o output.md] audio1.m4a

# Multiple files
./audiotranscribe [-o output.md] audio1.m4a audio2.m4a audio3.m4a
```

Key flags:
- `-o` - Output file path (optional, defaults to stdout)
- `-h` - Show help

Arguments:
- Audio file paths (required) - one or more audio files to transcribe

## Processing Large Files

For audio files longer than Gemini's limits, use the provided shell script:
```bash
./split_and_transcribe.sh large_audio.m4a
```

This script:
- Splits audio into 25-minute chunks using ffmpeg
- Processes all chunks with the main application
- Combines results into a single output file

Requirements for large file processing:
- ffmpeg must be installed
- Built audiotranscribe binary must exist in current directory

## Dependencies

Uses Go modules with key dependencies:
- `cloud.google.com/go/vertexai` - Gemini AI integration for transcription and summarization
- `cloud.google.com/go/storage` - Google Cloud Storage operations (for potential future use)
- `github.com/kelseyhightower/envconfig` - Environment variable configuration

Go version: 1.23.2+

## Authentication

Requires Google Cloud authentication via Application Default Credentials (ADC). Ensure `gcloud auth application-default login` is configured or service account credentials are available.

## AI Prompts

The application uses two main prompts defined in `prompts.go`:
- `TranscriptionPrompt` - Guides Gemini to produce clean speaker-identified transcripts without timestamps
- `SummaryPrompt` - Instructs Gemini to create structured summaries with key takeaways and markdown formatting

Both prompts are tuned for interview-style content and multi-language support.