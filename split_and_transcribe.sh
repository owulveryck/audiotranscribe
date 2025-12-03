#!/bin/bash

# Script to split audio files into 25-minute chunks and transcribe them
# Usage: ./split_and_transcribe.sh input_audio_file

set -e
# This snippet will resolve all symlinks to find the script's true directory
SOURCE=${BASH_SOURCE[0]}
while [ -h "$SOURCE" ]; do # Resolve $SOURCE until the file is no longer a symlink
  DIR=$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)
  SOURCE=$(readlink "$SOURCE")
  # If $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
  [[ $SOURCE != /* ]] && SOURCE=$DIR/$SOURCE
done
SCRIPT_DIR=$(cd -P "$(dirname "$SOURCE")" >/dev/null 2>&1 && pwd)

# Now, change into that directory
cd "$SCRIPT_DIR"
source .envrc

if [ $# -ne 1 ]; then
  echo "Usage: $0 <input_audio_file>"
  exit 1
fi

INPUT_FILE="$1"
INPUT_DIR=$(dirname "$INPUT_FILE")
BASENAME=$(basename "$INPUT_FILE" | sed 's/\.[^.]*$//')
OUTPUT_FILE="${INPUT_DIR}/${BASENAME}.md"
CHUNK_DURATION="25:00"

# Check if input file exists
if [ ! -f "$INPUT_FILE" ]; then
  echo "Error: Input file '$INPUT_FILE' not found"
  exit 1
fi

# Check if ffmpeg is available
if ! command -v ffmpeg &>/dev/null; then
  echo "Error: ffmpeg is required but not installed"
  exit 1
fi

# Check if audiotranscribe binary exists
if [ ! -f "./audiotranscribe" ]; then
  echo "Error: audiotranscribe binary not found. Please build it first with 'go build -o audiotranscribe .'"
  exit 1
fi

# Create temporary directory
TMPDIR=$(mktemp -d)
echo "Using temporary directory: $TMPDIR"

# Function to cleanup on exit
cleanup() {
  echo "Cleaning up temporary files..."
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

# Split audio file into 25-minute chunks
echo "Splitting audio file into 25-minute chunks..."
ffmpeg -i "$INPUT_FILE" -f segment -segment_time 1500 -c copy "$TMPDIR/chunk_%03d.${INPUT_FILE##*.}" -y

# Find all chunk files
CHUNK_FILES=($(ls "$TMPDIR"/chunk_*.* | sort))

if [ ${#CHUNK_FILES[@]} -eq 0 ]; then
  echo "Error: No chunks were created"
  exit 1
fi

echo "Created ${#CHUNK_FILES[@]} chunks"

# Transcribe all chunks
echo "Transcribing chunks..."
./audiotranscribe "${CHUNK_FILES[@]}" | tee "$OUTPUT_FILE"

echo "Transcription complete. Output saved to: $OUTPUT_FILE"
