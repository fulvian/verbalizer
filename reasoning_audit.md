# Reasoning Audit - Verbalizer

## Architectural Decisions

### 1. Three-Tier IPC Model (Extension -> NM Host -> Daemon)
- **Decision**: Use a Go Native Messaging (NM) Host as a bridge between the Manifest V3 extension and a persistent background daemon.
- **Reasoning**: Manifest V3 background scripts are ephemeral (service workers). A direct connection would be killed when the worker idles. The NM host is also spawned per message or for short durations. The **persistent daemon** is necessary to maintain long-running audio capture sessions and background transcription tasks that can outlive the browser tab or session.

### 2. Native Messaging vs. WebSocket/HTTP
- **Decision**: Native Messaging for Extension -> Host.
- **Reasoning**: Provides a secure, built-in mechanism for Chrome to talk to native code without opening network ports or managing local server authentication. Host -> Daemon uses Unix Domain Sockets for high-performance, local-only IPC.

### 3. ScreenCaptureKit (macOS) & PipeWire (Linux)
- **Decision**: Avoid loopback devices (BlackHole/VB-Audio) on macOS by using native APIs. Use PipeWire on Linux for per-app routing.
- **Reasoning**: User friction is the enemy. Requiring users to install and configure complex audio routing (loopback) makes the tool hard to adopt. ScreenCaptureKit (available since macOS 12.3) allows programmatic capture without virtual drivers, even if it captures system-wide audio by default.

### 4. whisper.cpp (C++) via CLI wrapper
- **Decision**: Call `whisper.cpp` as a subprocess instead of CGo bindings.
- **Reasoning**: While CGo is possible, it complicates cross-compilation and build stability. The CLI wrapper is robust, allows for easy memory isolation (if the transcriber crashes, the daemon survives), and simplifies the integration of different model versions.

## Trade-offs and Mitigations

| Trade-off | Mitigation |
|-----------|------------|
| macOS System Audio Capture (privacy) | Clearly documented in `ARCHITECTURE.md` and `README.md`. Local processing ensures audio never leaves the machine. |
| Manifest V3 Service Worker Lifecycle | State is managed entirely in the Daemon; Extension only sends events. |
| Model Size (ggml-small ~500MB) | Defaulted to 'small' for best accuracy/speed balance; provided script to download smaller/larger models. |

## Verification Strategy
- **TDD**: Extension was built strictly test-first, ensuring all detector edge cases (DOM changes) were handled.
- **Mocks**: Native Host and Daemon components were tested using mock IPC interfaces to simulate full data flow without requiring a live browser during CI.
- **CGo Safety**: macOS capture logic isolated in `capture_macos.go` with conditional build tags to ensure Linux builds don't fail.
