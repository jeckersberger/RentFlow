package application

import (
	"context"
	"errors"
	"fmt"
)

var ErrSigningKeyNotFound = errors.New("signing key not found")

// SigningKeyPair stores the PEM-encoded active RS256 key pair.
type SigningKeyPair struct {
	PrivateKeyPEM string
	PublicKeyPEM  string
}

// SigningKeyStore persists the active signing key pair.
type SigningKeyStore interface {
	Load(ctx context.Context) (SigningKeyPair, error)
	CreateIfAbsent(ctx context.Context, pair SigningKeyPair) error
}

// LoadOrCreateSigningKeyPair returns the persisted signing key pair. If no pair
// exists yet it generates one, stores it atomically, and then reloads the
// persisted value so concurrent service instances converge on the same key.
func LoadOrCreateSigningKeyPair(ctx context.Context, store SigningKeyStore) (SigningKeyPair, error) {
	pair, err := store.Load(ctx)
	if err == nil {
		if err := validateSigningKeyPair(pair); err != nil {
			return SigningKeyPair{}, err
		}
		return pair, nil
	}
	if !errors.Is(err, ErrSigningKeyNotFound) {
		return SigningKeyPair{}, fmt.Errorf("load signing key: %w", err)
	}

	key, err := GenerateKeyPair()
	if err != nil {
		return SigningKeyPair{}, fmt.Errorf("generate signing key: %w", err)
	}
	candidate := SigningKeyPair{
		PrivateKeyPEM: MarshalPrivateKeyPEM(key),
		PublicKeyPEM:  MarshalPublicKeyPEM(&key.PublicKey),
	}

	if err := store.CreateIfAbsent(ctx, candidate); err != nil {
		return SigningKeyPair{}, fmt.Errorf("persist signing key: %w", err)
	}

	pair, err = store.Load(ctx)
	if err != nil {
		return SigningKeyPair{}, fmt.Errorf("reload signing key: %w", err)
	}
	if err := validateSigningKeyPair(pair); err != nil {
		return SigningKeyPair{}, err
	}
	return pair, nil
}

func validateSigningKeyPair(pair SigningKeyPair) error {
	manager, err := NewTokenManager(pair.PrivateKeyPEM, pair.PublicKeyPEM)
	if err != nil {
		return fmt.Errorf("invalid persisted signing key: %w", err)
	}
	if manager.privateKey == nil || manager.publicKey == nil {
		return fmt.Errorf("invalid persisted signing key: incomplete key pair")
	}
	if manager.privateKey.PublicKey.N.Cmp(manager.publicKey.N) != 0 || manager.privateKey.PublicKey.E != manager.publicKey.E {
		return fmt.Errorf("invalid persisted signing key: private/public key mismatch")
	}
	return nil
}
