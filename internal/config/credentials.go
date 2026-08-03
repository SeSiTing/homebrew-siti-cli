package config

import (
	"errors"
	"fmt"
	"strings"

	keyring "github.com/zalando/go-keyring"
)

const credentialService = "siti-cli-ai"

// SetCredential stores a provider token in the operating system credential store.
func SetCredential(provider, token string) error {
	provider = normalizeCredentialProvider(provider)
	if provider == "" || strings.TrimSpace(token) == "" {
		return fmt.Errorf("provider 和 token 不能为空")
	}
	return keyring.Set(credentialService, provider, token)
}

// GetCredential reads a provider token from the operating system credential store.
func GetCredential(provider string) (string, error) {
	provider = normalizeCredentialProvider(provider)
	if provider == "" {
		return "", fmt.Errorf("provider 不能为空")
	}
	return keyring.Get(credentialService, provider)
}

// HasCredential reports whether a provider token exists.
func HasCredential(provider string) (bool, error) {
	_, err := GetCredential(provider)
	if errors.Is(err, keyring.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

// DeleteCredential removes a provider token.
func DeleteCredential(provider string) error {
	err := keyring.Delete(credentialService, normalizeCredentialProvider(provider))
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func normalizeCredentialProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}
