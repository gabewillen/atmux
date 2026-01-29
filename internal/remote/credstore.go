package remote

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/agentflare-ai/amux/pkg/api"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
)

// HubAuth contains JWT material for hub server configuration.
type HubAuth struct {
	OperatorPublicKey string
	SystemAccountKey  string
	SystemAccountJWT  string
	AccountPublicKey  string
	AccountJWT        string
}

// CredentialStore persists host credentials on disk.
type CredentialStore struct {
	baseDir     string
	operatorKP  nkeys.KeyPair
	systemKP    nkeys.KeyPair
	accountKP   nkeys.KeyPair
	operatorJWT string
	systemJWT   string
	accountJWT  string
	mu          sync.Mutex
}

// NewCredentialStore constructs a credential store rooted at baseDir.
func NewCredentialStore(baseDir string) (*CredentialStore, error) {
	if baseDir == "" {
		return nil, fmt.Errorf("credential store: base dir is empty")
	}
	authDir := filepath.Join(baseDir, "auth")
	if err := os.MkdirAll(authDir, 0o700); err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	operatorKP, err := loadOrCreateKeyPair(filepath.Join(authDir, "operator.nk"), nkeys.PrefixByteOperator)
	if err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	systemKP, err := loadOrCreateKeyPair(filepath.Join(authDir, "system.nk"), nkeys.PrefixByteAccount)
	if err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	accountKP, err := loadOrCreateKeyPair(filepath.Join(authDir, "account.nk"), nkeys.PrefixByteAccount)
	if err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	store := &CredentialStore{
		baseDir:    baseDir,
		operatorKP: operatorKP,
		systemKP:   systemKP,
		accountKP:  accountKP,
	}
	if err := store.refreshJWTs(); err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(baseDir, "creds"), 0o700); err != nil {
		return nil, fmt.Errorf("credential store: %w", err)
	}
	return store, nil
}

// HubAuth returns operator and account JWT material for hub configuration.
func (c *CredentialStore) HubAuth() (HubAuth, error) {
	if c == nil {
		return HubAuth{}, fmt.Errorf("credential store: nil")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	operatorPub, err := c.operatorKP.PublicKey()
	if err != nil {
		return HubAuth{}, fmt.Errorf("credential store: %w", err)
	}
	accountPub, err := c.accountKP.PublicKey()
	if err != nil {
		return HubAuth{}, fmt.Errorf("credential store: %w", err)
	}
	systemPub, err := c.systemKP.PublicKey()
	if err != nil {
		return HubAuth{}, fmt.Errorf("credential store: %w", err)
	}
	if c.operatorJWT == "" || c.accountJWT == "" {
		if err := c.refreshJWTs(); err != nil {
			return HubAuth{}, err
		}
	}
	return HubAuth{
		OperatorPublicKey: operatorPub,
		SystemAccountKey:  systemPub,
		SystemAccountJWT:  c.systemJWT,
		AccountPublicKey:  accountPub,
		AccountJWT:        c.accountJWT,
	}, nil
}

// CredentialPath returns the on-disk path for a named credential.
func (c *CredentialStore) CredentialPath(name string) string {
	if c == nil {
		return ""
	}
	return filepath.Join(c.baseDir, "creds", name+".creds")
}

// GetOrCreate returns a credential for the host, creating one if missing.
func (c *CredentialStore) GetOrCreate(hostID string, subjectPrefix string, kvBucket string) (Credential, error) {
	if c == nil {
		return Credential{}, fmt.Errorf("credential store: nil")
	}
	parsed, err := api.ParseHostID(hostID)
	if err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	perms := HostPermissions(subjectPrefix, parsed, kvBucket)
	return c.getOrCreateCredential(hostID, perms)
}

// DirectorCredential returns a credential with full subject access.
func (c *CredentialStore) DirectorCredential() (Credential, error) {
	if c == nil {
		return Credential{}, fmt.Errorf("credential store: nil")
	}
	perms := jwt.Permissions{
		Pub: jwt.Permission{Allow: []string{">"}},
		Sub: jwt.Permission{Allow: []string{">"}},
	}
	return c.getOrCreateCredential("director", perms)
}

func (c *CredentialStore) getOrCreateCredential(name string, perms jwt.Permissions) (Credential, error) {
	path := filepath.Join(c.baseDir, "creds", name+".creds")
	data, err := os.ReadFile(path)
	if err == nil {
		cred, err := ParseCredential(data)
		if err != nil {
			return Credential{}, fmt.Errorf("credential store: %w", err)
		}
		return cred, nil
	}
	cred, err := c.newCredential(name, perms)
	if err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	encoded, err := cred.Marshal()
	if err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return Credential{}, fmt.Errorf("credential store: %w", err)
	}
	return cred, nil
}

func (c *CredentialStore) newCredential(name string, perms jwt.Permissions) (Credential, error) {
	userKP, err := nkeys.CreatePair(nkeys.PrefixByteUser)
	if err != nil {
		return Credential{}, fmt.Errorf("create user key: %w", err)
	}
	userPub, err := userKP.PublicKey()
	if err != nil {
		return Credential{}, fmt.Errorf("create user key: %w", err)
	}
	claims := jwt.NewUserClaims(userPub)
	claims.Name = name
	claims.Permissions = perms
	userJWT, err := claims.Encode(c.accountKP)
	if err != nil {
		return Credential{}, fmt.Errorf("encode user jwt: %w", err)
	}
	seed, err := userKP.Seed()
	if err != nil {
		return Credential{}, fmt.Errorf("encode user jwt: %w", err)
	}
	creds, err := jwt.FormatUserConfig(userJWT, seed)
	if err != nil {
		return Credential{}, fmt.Errorf("encode user jwt: %w", err)
	}
	return ParseCredential(creds)
}

func (c *CredentialStore) refreshJWTs() error {
	if c.operatorKP == nil || c.accountKP == nil || c.systemKP == nil {
		return fmt.Errorf("credential store: missing keys")
	}
	operatorPub, err := c.operatorKP.PublicKey()
	if err != nil {
		return fmt.Errorf("credential store: %w", err)
	}
	accountPub, err := c.accountKP.PublicKey()
	if err != nil {
		return fmt.Errorf("credential store: %w", err)
	}
	systemPub, err := c.systemKP.PublicKey()
	if err != nil {
		return fmt.Errorf("credential store: %w", err)
	}
	operatorClaims := jwt.NewOperatorClaims(operatorPub)
	operatorJWT, err := operatorClaims.Encode(c.operatorKP)
	if err != nil {
		return fmt.Errorf("credential store: %w", err)
	}
	systemClaims := jwt.NewAccountClaims(systemPub)
	systemJWT, err := systemClaims.Encode(c.operatorKP)
	if err != nil {
		return fmt.Errorf("credential store: %w", err)
	}
	accountClaims := jwt.NewAccountClaims(accountPub)
	accountClaims.Limits.JetStreamLimits = jwt.JetStreamLimits{
		MemoryStorage:        jwt.NoLimit,
		DiskStorage:          jwt.NoLimit,
		Streams:              jwt.NoLimit,
		Consumer:             jwt.NoLimit,
		MaxAckPending:        jwt.NoLimit,
		MemoryMaxStreamBytes: jwt.NoLimit,
		DiskMaxStreamBytes:   jwt.NoLimit,
	}
	accountJWT, err := accountClaims.Encode(c.operatorKP)
	if err != nil {
		return fmt.Errorf("credential store: %w", err)
	}
	c.operatorJWT = operatorJWT
	c.systemJWT = systemJWT
	c.accountJWT = accountJWT
	return nil
}

func loadOrCreateKeyPair(path string, prefix nkeys.PrefixByte) (nkeys.KeyPair, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		kp, err := nkeys.FromSeed(data)
		if err != nil {
			return nil, err
		}
		return kp, nil
	}
	if !os.IsNotExist(err) {
		return nil, err
	}
	kp, err := nkeys.CreatePair(prefix)
	if err != nil {
		return nil, err
	}
	seed, err := kp.Seed()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, seed, 0o600); err != nil {
		return nil, err
	}
	return kp, nil
}
