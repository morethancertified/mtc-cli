package auth

import (
	"github.com/zalando/go-keyring"
)

const (
	service = "sprintctl"
	account = "default"
)

// StoreToken stores the access token securely in the system keychain
func StoreToken(token string) error {
	return keyring.Set(service, account, token)
}

// GetToken retrieves the access token from the system keychain
func GetToken() (string, error) {
	return keyring.Get(service, account)
}

// DeleteToken removes the access token from the system keychain
func DeleteToken() error {
	return keyring.Delete(service, account)
}

// HasToken checks if a token exists in the keychain
func HasToken() bool {
	_, err := GetToken()
	return err == nil
}
