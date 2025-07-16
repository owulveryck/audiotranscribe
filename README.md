# Audio Transcribe

A Go-based audio transcription tool that uses Google Cloud Platform's Vertex AI Gemini to transcribe audio files and generate summaries with timecoded speaker identification.

## Quickstart

### Prerequisites

1. **Google Cloud Setup**
   ```bash
   # Install gcloud CLI and authenticate
   gcloud auth application-default login
   
   # Set your project ID
   export GCP_PROJECT="your-project-id"
   ```

2. **Install Dependencies**
   - Go 1.19+ 
   - ffmpeg (for audio splitting)

### Installation

1. Clone and build:
   ```bash
   git clone https://github.com/owulveryck/audiotranscribe.git
   cd audiotranscribe
   go build -o audiotranscribe .
   ```

### Usage

**Single audio file:**
```bash
./audiotranscribe audio.m4a
```

**Multiple audio files with output file:**
```bash
./audiotranscribe -o transcript.md audio1.m4a audio2.m4a
```

**Large files (auto-split into 25min chunks):**
```bash
./split_and_transcribe.sh large_audio.m4a
```

### Environment Variables

- `GCP_PROJECT` (required) - Your Google Cloud project ID
- `GEMINI_MODEL` (optional) - Gemini model to use (default: "gemini-2.0-flash")
- `GCP_REGION` (optional) - GCP region (default: "europe-west9")

### Output

The tool generates markdown files with:
- Timestamped transcripts with speaker identification
- Combined summaries for multiple files
- Structured format for easy reading

Example output placed in same directory as input files.
