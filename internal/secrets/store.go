// internal/secrets/store.go
package secrets

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/dub-cli/internal/config"
)

type Store interface {
	Keys() ([]string, error)
	Set(name string, creds Credentials) error
	Get(name string) (Credentials, error)
	Delete(name string) error
	List() ([]Credentials, error)
}

type KeyringStore struct {
	ring keyring.Keyring
}

type Credentials struct {
	Name      string    `json:"name"`
	APIKey    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type storedCredentials struct {
	APIKey    string    `json:"api_key"`
	CreatedAt time.Time `json:"created_at"`
}

func OpenDefault() (Store, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: config.AppName,
	})
	if err != nil {
		return nil, err
	}
	return &KeyringStore{ring: ring}, nil
}

func (s *KeyringStore) Keys() ([]string, error) {
	return s.ring.Keys()
}

func (s *KeyringStore) Set(name string, creds Credentials) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing workspace name")
	}
	if creds.APIKey == "" {
		return fmt.Errorf("missing API key")
	}
	if creds.CreatedAt.IsZero() {
		creds.CreatedAt = time.Now().UTC()
	}

	payload, err := json.Marshal(storedCredentials{
		APIKey:    creds.APIKey,
		CreatedAt: creds.CreatedAt,
	})
	if err != nil {
		return err
	}

	return s.ring.Set(keyring.Item{
		Key:  credentialKey(name),
		Data: payload,
	})
}

func (s *KeyringStore) Get(name string) (Credentials, error) {
	name = normalize(name)
	if name == "" {
		return Credentials{}, fmt.Errorf("missing workspace name")
	}
	item, err := s.ring.Get(credentialKey(name))
	if err != nil {
		return Credentials{}, err
	}
	var stored storedCredentials
	if err := json.Unmarshal(item.Data, &stored); err != nil {
		return Credentials{}, err
	}

	return Credentials{
		Name:      name,
		APIKey:    stored.APIKey,
		CreatedAt: stored.CreatedAt,
	}, nil
}

func (s *KeyringStore) Delete(name string) error {
	name = normalize(name)
	if name == "" {
		return fmt.Errorf("missing workspace name")
	}
	return s.ring.Remove(credentialKey(name))
}

func (s *KeyringStore) List() ([]Credentials, error) {
	keys, err := s.Keys()
	if err != nil {
		return nil, err
	}
	var out []Credentials
	for _, k := range keys {
		name, ok := ParseCredentialKey(k)
		if !ok {
			continue
		}
		creds, err := s.Get(name)
		if err != nil {
			return nil, err
		}
		out = append(out, creds)
	}
	return out, nil
}

func ParseCredentialKey(k string) (name string, ok bool) {
	const prefix = "workspace:"
	if !strings.HasPrefix(k, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(k, prefix)
	if strings.TrimSpace(rest) == "" {
		return "", false
	}
	return rest, true
}

func credentialKey(name string) string {
	return fmt.Sprintf("workspace:%s", name)
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
