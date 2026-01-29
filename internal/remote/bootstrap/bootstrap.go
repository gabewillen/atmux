package bootstrap

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client wraps an SSH client for bootstrapping.
type Client struct {
	client *ssh.Client
}

// Dial connects to the remote host via SSH.
// keyInfo is path to private key file. If empty, tries agent.
func Dial(host, user, keyPath string) (*Client, error) {
	var authMethods []ssh.AuthMethod

	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("read private key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else {
		// Try agent? or Fail?
		// For now fail if no key.
		// TODO: Implement SSH agent auth support.
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: StrictHostKeyChecking
		Timeout:         10 * time.Second,
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("ssh dial: %w", err)
	}

	return &Client{client: client}, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Upload copies a local file to the remote path with permissions.
func (c *Client) Upload(localPath, remotePath string, mode os.FileMode) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()

	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fmt.Fprintf(w, "C%04o %d %s\n", mode.Perm(), size, remotePath) // SCP protocol basic
		io.Copy(w, f)
		fmt.Fprint(w, "\x00")
	}()

	// But "scp" needs to be running on remote?
	// Using "scp -t" on remote.
	// Standard scp usage:
	// This is complex to implement raw scp.
	// simpler: cat via ssh exec.

	return c.writeViaCat(f, remotePath, mode)
}

func (c *Client) writeViaCat(r io.Reader, remotePath string, mode os.FileMode) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdin = r
	// chmod after write?
	cmd := fmt.Sprintf("cat > %s && chmod %04o %s", remotePath, mode.Perm(), remotePath)
	return session.Run(cmd)
}

// Exec runs a command and returns output.
func (c *Client) Exec(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), fmt.Errorf("exec %q: %w\nOutput: %s", cmd, err, string(out))
	}
	return string(out), nil
}
