package hssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os/user"

	"github.com/mikkeloscar/sshconfig"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
)

// SSHClient holds the details of a client
type SSHClient struct {
	Config *ssh.ClientConfig
	Host   string
	Port   int
	W      io.Writer
}

// getHostDetails will return the private file for the given host
// we will try to read from ~/.ssh/config
func getHostDetails(host string) (sshHost *sshconfig.SSHHost, err error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	cpath := u.HomeDir + "/.ssh/config"
	hosts, err := sshconfig.ParseSSHConfig(cpath)
	if err != nil {
		return nil, fmt.Errorf("failed to read sshConfig: %v", err)
	}

	h, bh := filterHost(hosts, host)
	if h != nil {
		return h, nil
	}

	if bh != nil {
		return bh, nil
	}

	return &sshconfig.SSHHost{
		HostName:     host,
		User:         u.Name,
		Port:         22,
		IdentityFile: u.HomeDir + "/.ssh/id_rsa"}, nil
}

// filterHosts will return the host matching "host" from hosts and open host.
// returns error if not found
func filterHost(hosts []*sshconfig.SSHHost, host string) (mhost, ghost *sshconfig.SSHHost) {
	for _, h := range hosts {
		if h.HostName == host {
			mhost = h
			continue
		}

		if h.HostName == "" {
			ghost = h
		}
	}

	return mhost, ghost
}

// getSSHConfig will return the Client ssh configuration
func getSSHConfig(h *sshconfig.SSHHost) (*ssh.ClientConfig, error) {
	expath, err := homedir.Expand(h.IdentityFile)
	if err != nil {
		return nil, err
	}

	buffer, err := ioutil.ReadFile(expath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %v", err)
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the private key: %v", err)
	}

	return &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}, nil
}

func createSession(c *SSHClient) (*ssh.Session, error) {
	connection, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port), c.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial to remote: %v", err)
	}

	session, err := connection.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %v", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
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

	config, err := getSSHConfig(h)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	return &SSHClient{
		Config: config,
		Host:   h.HostName,
		Port:   h.Port,
		W:      w,
	}, nil
}
