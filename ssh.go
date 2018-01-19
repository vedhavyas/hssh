package hssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os/user"
	"path/filepath"

	"github.com/mikkeloscar/sshconfig"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

type SSHHostDetails struct {
	Host         string
	User         string
	Port         int
	IdentityFile string
}

// SSHClient holds the details of a client
type SSHClient struct {
	Config  *ssh.ClientConfig
	Details *SSHHostDetails
	W       io.Writer
}

// readSSHConfig reads and returns host details if present
func readSSHConfig(host, path string) (*SSHHostDetails, error) {
	hosts, err := sshconfig.ParseSSHConfig(path)
	if err != nil {
		return nil, err
	}

	h := filterHost(hosts, host)
	if h == nil {
		return nil, fmt.Errorf("failed to find host in ssh config")
	}

	return newSSHHostDetails(h.HostName, h.User, h.IdentityFile, h.Port)
}

// getHostDetails will return the private file for the given host
// we will try to read from ~/.ssh/config and /etc/ssh/ssh_config
func getHostDetails(host string) (sshHost *SSHHostDetails, err error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	hdir, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	paths := []string{filepath.Join(hdir, ".ssh", "config"), filepath.Join("/", "etc", "ssh", "ssh_config")}
	for _, path := range paths {
		h, err := readSSHConfig(host, path)
		if err != nil {
			continue
		}

		return h, nil
	}

	return newSSHHostDetails(host, u.Name, filepath.Join(u.HomeDir, ".ssh", "id_rsa"), 22)
}

// newSSHHostDetails returns the new SSHHostDetails. Expands a relative identity file if required
func newSSHHostDetails(host, user, identityFile string, port int) (*SSHHostDetails, error) {
	extPath, err := homedir.Expand(identityFile)
	if err != nil {
		return nil, err
	}
	return &SSHHostDetails{
		Host:         host,
		User:         user,
		Port:         port,
		IdentityFile: extPath}, nil
}

// filterHosts will return the host matching "host" from hosts and open host.
// returns error if not found
func filterHost(hosts []*sshconfig.SSHHost, host string) (h *sshconfig.SSHHost) {
	for _, h := range hosts {
		if h.HostName == host {
			return h
		}
	}

	return nil
}

// getSSHConfig will return the Client ssh configuration
func getSSHConfig(filePath, user string) (*ssh.ClientConfig, error) {
	buffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the private key: %v", err)
	}

	return &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}, nil
}

func createSession(c *SSHClient) (*ssh.Session, error) {
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Details.Host, c.Details.Port), c.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to remote: %v", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to request xterm: %v", err)
	}

	return session, nil
}

// ExecuteCommand execute the command
func (c *SSHClient) ExecuteCommand(cmd string) error {
	session, err := createSession(c)
	if err != nil {
		return fmt.Errorf("failed to create new session on remote: %v", err)
	}

	defer session.Close()
	session.Stdout = c.W
	session.Stderr = c.W
	return session.Run(cmd)
}

// NewSSHClient returns a new SSHClient
func NewSSHClient(host string, w io.Writer) (*SSHClient, error) {
	h, err := getHostDetails(host)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the host details: %v", err)
	}

	config, err := getSSHConfig(h.IdentityFile, h.User)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &SSHClient{
		Config:  config,
		Details: h,
		W:       w,
	}, nil
}
