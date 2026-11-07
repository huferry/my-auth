package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Account struct {
	Name   string `json:"name"`
	Secret string `json:"secret"` // base64-encoded raw TOTP secret
}

func configDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(dir, "my-auth")
	if err := os.MkdirAll(d, 0700); err != nil {
		return "", err
	}
	return d, nil
}

func storePath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "accounts.enc"), nil
}

func keyPath() (string, error) {
	d, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "store.key"), nil
}

func loadOrCreateKey() ([]byte, error) {
	kp, err := keyPath()
	if err != nil {
		return nil, err
	}

	key, err := os.ReadFile(kp)
	if err == nil && len(key) == 32 {
		return key, nil
	}

	key, err = generateKey()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(kp, key, 0600); err != nil {
		return nil, err
	}
	return key, nil
}

func isHexString(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

type csvRecord struct {
	name, encryptedSeed, salt, iv string
}

func parseTokensCSV(filename string) ([]csvRecord, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var records []string
	var current string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\"")

		if line == "name,encrypted_seed,salt,iv" {
			continue
		}

		if current != "" {
			current += line
		} else {
			current = line
		}

		parts := strings.Split(current, ",")
		if len(parts) >= 4 {
			lastPart := parts[len(parts)-1]
			if isHexString(lastPart) && len(lastPart) == 32 {
				records = append(records, current)
				current = ""
			}
		}
	}

	var result []csvRecord
	for _, record := range records {
		parts := strings.Split(record, ",")
		if len(parts) < 4 {
			continue
		}
		result = append(result, csvRecord{
			iv:            parts[len(parts)-1],
			salt:          parts[len(parts)-2],
			encryptedSeed: parts[len(parts)-3],
			name:          strings.Join(parts[:len(parts)-3], ","),
		})
	}

	return result, nil
}

func importAccounts(records []csvRecord, password string) ([]Account, error) {
	var accounts []Account
	for _, r := range records {
		secret, err := decryptSeed(r.encryptedSeed, r.salt, r.iv, password)
		if err != nil {
			return nil, fmt.Errorf("decrypting %q: %w", r.name, err)
		}
		accounts = append(accounts, Account{
			Name:   r.name,
			Secret: base64.StdEncoding.EncodeToString(secret),
		})
	}
	return accounts, nil
}

func saveAccounts(accounts []Account) error {
	key, err := loadOrCreateKey()
	if err != nil {
		return err
	}

	data, err := json.Marshal(accounts)
	if err != nil {
		return err
	}

	encrypted, err := encryptWithKey(data, key)
	if err != nil {
		return err
	}

	path, err := storePath()
	if err != nil {
		return err
	}

	return os.WriteFile(path, encrypted, 0600)
}

func loadAccounts() ([]Account, error) {
	key, err := loadOrCreateKey()
	if err != nil {
		return nil, err
	}

	path, err := storePath()
	if err != nil {
		return nil, err
	}

	encrypted, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no accounts imported yet; use -i to import")
		}
		return nil, err
	}

	data, err := decryptWithKey(encrypted, key)
	if err != nil {
		return nil, fmt.Errorf("corrupted data or key mismatch")
	}

	var accounts []Account
	if err := json.Unmarshal(data, &accounts); err != nil {
		return nil, fmt.Errorf("corrupted data")
	}

	return accounts, nil
}
