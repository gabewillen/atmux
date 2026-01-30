package remote

import "testing"

func TestCredentialStoreLifecycle(t *testing.T) {
	store, err := NewCredentialStore(t.TempDir())
	if err != nil {
		t.Fatalf("new credential store: %v", err)
	}
	if _, err := store.HubAuth(); err != nil {
		t.Fatalf("hub auth: %v", err)
	}
	if store.CredentialPath("alpha") == "" {
		t.Fatalf("expected credential path")
	}
	if _, err := store.GetOrCreate("", "amux", "bucket"); err == nil {
		t.Fatalf("expected invalid host error")
	}
	if _, err := store.DirectorCredential(); err != nil {
		t.Fatalf("director credential: %v", err)
	}
}

func TestCredentialStoreNil(t *testing.T) {
	var store *CredentialStore
	if _, err := store.HubAuth(); err == nil {
		t.Fatalf("expected hub auth error")
	}
	if store.CredentialPath("alpha") != "" {
		t.Fatalf("expected empty credential path")
	}
	if _, err := store.GetOrCreate("host", "amux", "bucket"); err == nil {
		t.Fatalf("expected get or create error")
	}
	if _, err := store.DirectorCredential(); err == nil {
		t.Fatalf("expected director credential error")
	}
}
