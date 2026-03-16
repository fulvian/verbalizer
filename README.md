# Verbalizer

Automatic audio recording, transcription, and documentation for Google Meet and Microsoft Teams calls.

## Features

- **Automatic Detection**: Detects when you join a Google Meet or MS Teams call
- **Background Recording**: Records audio automatically without manual intervention
- **Local Transcription**: Uses whisper.cpp for privacy-first, local speech-to-text
- **Markdown Output**: Generates clean, timestamped transcripts in Markdown format
- **Cross-Platform**: Works on macOS and Linux

## Requirements

- macOS 12.3+ (Monterey) or Linux with PipeWire
- Chrome or Chromium browser
- Go 1.21+ (for building)
- Node.js 18+ (for extension build)
- FFmpeg (usually pre-installed on macOS/Linux)

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/verbalizer.git
cd verbalizer

# Build all components
make build

# Install (requires permissions setup)
make install
```

## Usage

Once installed, Verbalizer runs automatically in the background:

1. Open Chrome and join a Google Meet or MS Teams call
2. Verbalizer automatically detects the call and starts recording
3. When you leave the call, recording stops and transcription begins
4. Find your transcript in `~/verbalizer/transcripts/`

## Output Structure

```
~/verbalizer/
├── recordings/          # MP3 audio files
│   └── 2026-03-16_09-30-00_google-meet.mp3
├── transcripts/         # Markdown transcripts
│   └── 2026-03-16_09-30-00_google-meet.md
└── metadata.db          # SQLite index
```

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture documentation.

## Development

```bash
# Build extension only
make extension

# Build native host only
make native-host

# Build daemon only
make daemon

# Run tests
make test

# Install development version
make install-dev
```

## License

MIT License - see [LICENSE](LICENSE) for details.
