package main

const (
	// TranscriptionPrompt is the prompt used for audio transcription
	TranscriptionPrompt = `Transcribe this audio interview accurately. Follow these guidelines:

1. Format: Speaker: [spoken content]
2. Use "Speaker A", "Speaker B", etc. to identify different speakers
3. DO NOT include timestamps or timecodes
4. Focus on complete, meaningful sentences - avoid fragmentary repetitions
5. If you hear repetitive words (like "yes yes yes"), transcribe it only once unless the repetition is clearly intentional and meaningful
6. Capture the essence of what is being said, not every single utterance
7. If there are unclear sections, use [unclear] rather than guessing or repeating
8. Maintain natural conversation flow and avoid artificial line breaks

Provide a clean, readable transcript without timestamps.`

	// SummaryPrompt is the prompt used for post-processing and summarization
	SummaryPrompt = `The context is about designing a new eCommerce platform.

	Create a comprehensive summary of these interview transcripts. Follow these guidelines:

1. Extract the main ideas, key insights, and solutions discussed
2. Structure the output using markdown formatting
3. Generate the summary in the same language as the original content
4. If multiple audio files are provided, identify common themes and insights across all files
5. Include a "Key Takeaways" section with the most important points
6. Add a "Potential Issues/Pitfalls" section at the end if any concerns were identified
7. Organize content logically with clear headings and bullet points
8. Focus on actionable insights and concrete information

Provide a well-structured, comprehensive analysis of the content.`
)
