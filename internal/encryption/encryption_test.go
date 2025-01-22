package encryption_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vatsal3003/viswals/internal/encryption"
)

func TestEncryptDecrypt(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		errString string
	}{
		{
			name:      "Normal string",
			input:     "Hello, World!",
			wantErr:   false,
			errString: "",
		},
		{
			name:      "Empty string",
			input:     "",
			wantErr:   false,
			errString: "",
		},
		{
			name:      "Long string",
			input:     strings.Repeat("Long text ", 100),
			wantErr:   false,
			errString: "",
		},
		{
			name:      "Special characters",
			input:     "!@#$%^&*()_+-=[]{}|;:,.<>?`~",
			wantErr:   false,
			errString: "",
		},
		{
			name:      "Unicode characters",
			input:     "Hello, 世界!",
			wantErr:   false,
			errString: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := encryption.Encrypt(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, encrypted)
			assert.NotEqual(t, tt.input, encrypted)

			// Test decryption
			decrypted, err := encryption.Decrypt(encrypted)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, decrypted)
		})
	}
}
