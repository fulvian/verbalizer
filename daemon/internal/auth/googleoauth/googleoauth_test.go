package googleoauth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGeneratePKCEVerifier(t *testing.T) {
	verifier, err := generatePKCEVerifier()
	if err != nil {
		t.Fatalf("generatePKCEVerifier() error = %v", err)
	}

	// PKCE verifier should be URL-safe base64 encoded, 32 bytes
	// RawURLEncoding produces 43 characters from 32 bytes
	if len(verifier) != 43 {
		t.Errorf("generatePKCEVerifier() length = %d, want 43", len(verifier))
	}

	// Should be valid URL-safe base64 (no padding)
	for _, c := range verifier {
		if !isValidBase64URLChar(byte(c)) {
			t.Errorf("generatePKCEVerifier() contains invalid character: %c", c)
		}
	}
}

func TestGeneratePKCEVerifier_Uniqueness(t *testing.T) {
	verifiers := make(map[string]bool)
	for i := 0; i < 100; i++ {
		verifier, err := generatePKCEVerifier()
		if err != nil {
			t.Fatalf("generatePKCEVerifier() error = %v", err)
		}
		if verifiers[verifier] {
			t.Errorf("generatePKCEVerifier() produced duplicate verifier after %d iterations", i)
		}
		verifiers[verifier] = true
	}
}

func TestGeneratePKCEChallenge(t *testing.T) {
	// Use a known verifier to test challenge generation
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := generatePKCEChallenge(verifier)

	// Expected: S256 hash of verifier, URL-safe base64 encoded
	h := sha256.New()
	h.Write([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if challenge != expected {
		t.Errorf("generatePKCEChallenge(%q) = %q, want %q", verifier, challenge, expected)
	}
}

func TestGeneratePKCEChallenge_S256(t *testing.T) {
	// Test that S256 is actually used by verifying the challenge matches
	verifier := "test-verifier-123"
	challenge := generatePKCEChallenge(verifier)

	// Manually compute S256
	h := sha256.New()
	h.Write([]byte(verifier))
	expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	if challenge != expected {
		t.Errorf("generatePKCEChallenge() should use S256, got %q, expected %q", challenge, expected)
	}
}

func TestGenerateState(t *testing.T) {
	state, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	// State should be 16 bytes, URL-safe base64 encoded = 22 characters
	if len(state) != 22 {
		t.Errorf("generateState() length = %d, want 22", len(state))
	}
}

func TestGenerateState_Uniqueness(t *testing.T) {
	states := make(map[string]bool)
	for i := 0; i < 100; i++ {
		state, err := generateState()
		if err != nil {
			t.Fatalf("generateState() error = %v", err)
		}
		if states[state] {
			t.Errorf("generateState() produced duplicate state after %d iterations", i)
		}
		states[state] = true
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test@Example.COM", "test@example.com"},
		{"  user@gmail.com  ", "user@gmail.com"},
		{"Mixed.Case@Domain.COM", "mixed.case@domain.com"},
		{"already@lowercase.com", "already@lowercase.com"},
		{"  spaces@around.com  ", "spaces@around.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Helper function to check if a character is valid in URL-safe base64
func isValidBase64URLChar(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '_'
}

// TestPKCEVerifierFormat tests that generated verifiers meet RFC 7636 requirements
func TestPKCEVerifierFormat_RFC7636(t *testing.T) {
	verifier, err := generatePKCEVerifier()
	if err != nil {
		t.Fatalf("generatePKCEVerifier() error = %v", err)
	}

	// RFC 7636: verifier MUST be between 43 and 128 characters
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("PKCE verifier length %d not in valid range [43, 128]", len(verifier))
	}

	// RFC 7636: verifier MUST contain only unreserved characters
	// unreserved = ALPHA / DIGIT / "-" / "." / "_" / "~"
	for _, c := range verifier {
		if !isRFC7636Valid(byte(c)) {
			t.Errorf("PKCE verifier contains invalid character: %c (0x%02x)", c, c)
		}
	}
}

func isRFC7636Valid(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' ||
		c == '.' ||
		c == '_' ||
		c == '~'
}

// TestPKCEChallengeConsistency verifies that same verifier always produces same challenge
func TestPKCEChallengeConsistency(t *testing.T) {
	verifier := "this-is-a-test-verifier-for-consistency"

	challenge1 := generatePKCEChallenge(verifier)
	challenge2 := generatePKCEChallenge(verifier)

	if challenge1 != challenge2 {
		t.Errorf("generatePKCEChallenge() not consistent: %q vs %q", challenge1, challenge2)
	}
}

// TestDifferentVerifiersProduceDifferentChallenges verifies variety
func TestDifferentVerifiersProduceDifferentChallenges(t *testing.T) {
	challenges := make(map[string]bool)

	for i := 0; i < 50; i++ {
		verifier, _ := generatePKCEVerifier()
		challenge := generatePKCEChallenge(verifier)

		if challenges[challenge] {
			t.Errorf("Challenge collision after %d iterations", i)
		}
		challenges[challenge] = true
	}
}

// TestStateFormat verifies state meets OAuth 2.0 security requirements
func TestStateFormat(t *testing.T) {
	state, err := generateState()
	if err != nil {
		t.Fatalf("generateState() error = %v", err)
	}

	// State should be cryptographically random enough for CSRF protection
	// At least 16 bytes = 128 bits of entropy
	if len(state) < 20 { // Account for base64 encoding overhead
		t.Errorf("State may not have enough entropy, length = %d", len(state))
	}
}
