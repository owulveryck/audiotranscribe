# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based audio transcription tool that uses Google Cloud Platform services (Vertex AI Gemini) to transcribe audio files and generate summaries. The application processes audio files directly from local storage, transcribes them using Gemini models, and provides structured output with timecoded speaker identification. Supports processing multiple audio files at once with combined summary generation.

## Core Architecture

The application consists of three main components:

- `main.go` - CLI interface, configuration handling, and workflow orchestration
- `gcp_plumbing.go` - Google Cloud Storage operations (currently unused)
- `gemini_processing.go` - Vertex AI Gemini integration for transcription and post-processing

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

## Dependencies

Uses Go modules with key dependencies:
- `cloud.google.com/go/vertexai` - Gemini AI integration for transcription and summarization
- `github.com/kelseyhightower/envconfig` - Environment variable configuration

## Authentication

Requires Google Cloud authentication via Application Default Credentials (ADC). Ensure `gcloud auth application-default login` is configured or service account credentials are available.