package application

import (
	"context"
	"errors"
	"testing"
)

type fakeSigningKeyStore struct {
	pair         SigningKeyPair
	loadErr      error
	createErr    error
	createCalls  int
	onCreate     func(SigningKeyPair)
}

func (f *fakeSigningKeyStore) Load(context.Context) (SigningKeyPair, error) {
	if f.loadErr != nil {
		return SigningKeyPair{}, f.loadErr
	}
	return f.pair, nil
}

func (f *fakeSigningKeyStore) CreateIfAbsent(_ context.Context, pair SigningKeyPair) error {
	f.createCalls++
	if f.createErr != nil {
		return f.createErr
	}
	if f.onCreate != nil {
		f.onCreate(pair)
	} else if f.pair.PrivateKeyPEM == "" {
		f.pair = pair
	}
	f.loadErr = nil
	return nil
}

func generatedPair(t *testing.T) SigningKeyPair {
	t.Helper()
	key, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	return SigningKeyPair{
		PrivateKeyPEM: MarshalPrivateKeyPEM(key),
		PublicKeyPEM:  MarshalPublicKeyPEM(&key.PublicKey),
	}
}

func TestLoadOrCreateSigningKeyPairKeepsExistingKey(t *testing.T) {
	existing := generatedPair(t)
	store := &fakeSigningKeyStore{pair: existing}

	got, err := LoadOrCreateSigningKeyPair(context.Background(), store)
	if err != nil {
		t.Fatalf("LoadOrCreateSigningKeyPair: %v", err)
	}
	if got.PrivateKeyPEM != existing.PrivateKeyPEM || got.PublicKeyPEM != existing.PublicKeyPEM {
		t.Fatal("existing signing key was not preserved")
	}
	if store.createCalls != 0 {
		t.Fatalf("CreateIfAbsent called %d times, want 0", store.createCalls)
	}
}

func TestLoadOrCreateSigningKeyPairCreatesAndReloadsKey(t *testing.T) {
	store := &fakeSigningKeyStore{loadErr: ErrSigningKeyNotFound}

	got, err := LoadOrCreateSigningKeyPair(context.Background(), store)
	if err != nil {
		t.Fatalf("LoadOrCreateSigningKeyPair: %v", err)
	}
	if got.PrivateKeyPEM == "" || got.PublicKeyPEM == "" {
		t.Fatal("generated signing key is empty")
	}
	if store.createCalls != 1 {
		t.Fatalf("CreateIfAbsent called %d times, want 1", store.createCalls)
	}
}

func TestLoadOrCreateSigningKeyPairUsesConcurrentWinner(t *testing.T) {
	winner := generatedPair(t)
	store := &fakeSigningKeyStore{loadErr: ErrSigningKeyNotFound}
	store.onCreate = func(SigningKeyPair) {
		// Simulate another instance winning INSERT ... ON CONFLICT DO NOTHING.
		store.pair = winner
	}

	got, err := LoadOrCreateSigningKeyPair(context.Background(), store)
	if err != nil {
		t.Fatalf("LoadOrCreateSigningKeyPair: %v", err)
	}
	if got.PrivateKeyPEM != winner.PrivateKeyPEM || got.PublicKeyPEM != winner.PublicKeyPEM {
		t.Fatal("service did not reload the concurrently persisted winner")
	}
}

func TestLoadOrCreateSigningKeyPairRejectsMismatchedPair(t *testing.T) {
	private := generatedPair(t)
	other := generatedPair(t)
	store := &fakeSigningKeyStore{pair: SigningKeyPair{
		PrivateKeyPEM: private.PrivateKeyPEM,
		PublicKeyPEM:  other.PublicKeyPEM,
	}}

	_, err := LoadOrCreateSigningKeyPair(context.Background(), store)
	if err == nil {
		t.Fatal("expected mismatched signing key pair to fail")
	}
}

func TestLoadOrCreateSigningKeyPairPropagatesStoreFailure(t *testing.T) {
	boom := errors.New("database unavailable")
	store := &fakeSigningKeyStore{loadErr: boom}

	_, err := LoadOrCreateSigningKeyPair(context.Background(), store)
	if !errors.Is(err, boom) {
		t.Fatalf("error = %v, want wrapped %v", err, boom)
	}
}
