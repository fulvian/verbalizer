//go:build linux

package audio

import (
	"fmt"
	"os/exec"
	"strings"
)

// AudioSource represents an available audio source on Linux
type AudioSource struct {
	Name        string // Source name (e.g., "Built-in Audio Analog Stereo")
	ID          string // Source index or name
	IsMonitor   bool   // True if this is a monitor source (system audio)
	Description string // Description from PulseAudio/PipeWire
}

// SourceDiscovery discovers available audio sources on Linux
type SourceDiscovery struct{}

// NewSourceDiscovery creates a new source discovery instance
func NewSourceDiscovery() *SourceDiscovery {
	return &SourceDiscovery{}
}

// DiscoverSources returns list of available audio sources
func (sd *SourceDiscovery) DiscoverSources() ([]AudioSource, error) {
	// Try PulseAudio first, then PipeWire
	sources, err := sd.discoverPulseSources()
	if err != nil {
		// Try alternative method
		sources, err = sd.discoverPipeWireSources()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to discover audio sources: %w", err)
	}
	return sources, nil
}

// discoverPulseSources uses pactl to list sources
func (sd *SourceDiscovery) discoverPulseSources() ([]AudioSource, error) {
	cmd := exec.Command("pactl", "list", "sources", "short")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pactl failed: %w", err)
	}

	sources := parsePactlOutput(string(output))
	return sources, nil
}

// discoverPipeWireSources uses pw-cli or wpctl to list sources
func (sd *SourceDiscovery) discoverPipeWireSources() ([]AudioSource, error) {
	// Try wpctl (WirePlumber CLI)
	cmd := exec.Command("wpctl", "audio-sources")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try pw-cli
		cmd = exec.Command("pw-cli", "list-objects", "source")
		output, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("no audio source discovery tool available: %w", err)
		}
		return parsePwCliOutput(string(output)), nil
	}
	return parseWpctlOutput(string(output)), nil
}

// parsePactlOutput parses pactl list sources short output
func parsePactlOutput(output string) []AudioSource {
	var sources []AudioSource
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Format: index\tname\tproperty\t...
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		id := fields[0]
		name := fields[1]
		isMonitor := strings.HasSuffix(name, ".monitor")

		source := AudioSource{
			ID:        id,
			Name:      name,
			IsMonitor: isMonitor,
		}

		// Get description from extended output
		if desc := getSourceDescription(name); desc != "" {
			source.Description = desc
		}

		sources = append(sources, source)
	}

	return sources
}

// parseWpctlOutput parses wpctl audio-sources output
func parseWpctlOutput(output string) []AudioSource {
	var sources []AudioSource
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Objects") {
			continue
		}

		// Format: * ID Name ... (asterisk for default)
		name := strings.TrimPrefix(line, "* ")
		fields := strings.Fields(name)
		if len(fields) < 2 {
			continue
		}

		id := fields[0]
		sourceName := strings.Join(fields[1:], " ")
		isMonitor := strings.HasSuffix(sourceName, ".monitor")

		sources = append(sources, AudioSource{
			ID:          id,
			Name:        sourceName,
			IsMonitor:   isMonitor,
			Description: sourceName,
		})
	}

	return sources
}

// parsePwCliOutput parses pw-cli list-objects source output
func parsePwCliOutput(output string) []AudioSource {
	var sources []AudioSource
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if !strings.Contains(line, "id:") {
			continue
		}

		// Extract ID and name
		var id, name string
		fields := strings.Fields(line)
		for i, f := range fields {
			if f == "id:" && i+1 < len(fields) {
				id = fields[i+1]
			}
			if f == "node.name=" {
				name = strings.Trim(fields[i+1], "\"")
			}
		}

		if id != "" && name != "" {
			sources = append(sources, AudioSource{
				ID:        id,
				Name:      name,
				IsMonitor: strings.HasSuffix(name, ".monitor"),
			})
		}
	}

	return sources
}

// getSourceDescription gets extended description for a source
func getSourceDescription(sourceName string) string {
	cmd := exec.Command("pactl", "list", "sources")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	var inSource bool
	var description string

	for _, line := range lines {
		if strings.Contains(line, "Name: "+sourceName) {
			inSource = true
			continue
		}

		if inSource {
			if strings.HasPrefix(strings.TrimSpace(line), "Description:") {
				description = strings.TrimSpace(strings.TrimPrefix(line, "Description:"))
				break
			}
			if strings.HasPrefix(line, "Name:") && !strings.Contains(line, sourceName) {
				break
			}
		}
	}

	return description
}

// FindMonitorSource finds the preferred monitor source (system audio)
// Returns the source name for ffmpeg -i parameter
func (sd *SourceDiscovery) FindMonitorSource() (string, error) {
	sources, err := sd.DiscoverSources()
	if err != nil {
		return "", fmt.Errorf("DiscoverSources failed: %w", err)
	}

	fmt.Printf("[DEBUG] FindMonitorSource: discovered %d sources\n", len(sources))
	for i, s := range sources {
		fmt.Printf("[DEBUG]   Source[%d]: ID=%s, Name=%s, IsMonitor=%v, Desc=%s\n", i, s.ID, s.Name, s.IsMonitor, s.Description)
	}

	// Look for monitor sources
	var monitors []AudioSource
	for _, s := range sources {
		if s.IsMonitor {
			monitors = append(monitors, s)
		}
	}

	fmt.Printf("[DEBUG] FindMonitorSource: found %d monitor sources\n", len(monitors))

	if len(monitors) == 0 {
		return "", fmt.Errorf("no monitor sources found (checked %d total sources)", len(sources))
	}

	// Prefer default monitor (usually has .monitor suffix and is default)
	for _, m := range monitors {
		if m.Name == "auto_null.monitor" {
			// Skip null sink monitor
			continue
		}
		// Return the name for ffmpeg
		fmt.Printf("[DEBUG] FindMonitorSource: selected monitor=%s\n", m.Name)
		return m.Name, nil
	}

	// Fallback to first monitor
	fmt.Printf("[DEBUG] FindMonitorSource: falling back to first monitor=%s\n", monitors[0].Name)
	return monitors[0].Name, nil
}

// GetSourceForFFmpeg returns the ffmpeg input string for the given source name
func GetSourceForFFmpeg(sourceName string) string {
	// PulseAudio uses source name directly
	// PipeWire can use name or index
	if sourceName == "" {
		return "default"
	}
	return sourceName
}

// ValidateSource checks if the specified source exists and is available
func (sd *SourceDiscovery) ValidateSource(sourceName string) (bool, error) {
	sources, err := sd.DiscoverSources()
	if err != nil {
		return false, err
	}

	for _, s := range sources {
		if s.Name == sourceName || s.ID == sourceName {
			return true, nil
		}
	}

	return false, fmt.Errorf("source %q not found", sourceName)
}

// GetDefaultSource returns the default source name
func (sd *SourceDiscovery) GetDefaultSource() (string, error) {
	cmd := exec.Command("pactl", "get-default-source")
	output, err := cmd.Output()
	if err != nil {
		// Try wpctl
		cmd = exec.Command("wpctl", "get-default", "@default.audio.source")
		output, err = cmd.Output()
		if err != nil {
			return "default", nil // Fallback to default
		}
		return strings.TrimSpace(string(output)), nil
	}

	return strings.TrimSpace(string(output)), nil
}
