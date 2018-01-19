package hssh

import (
	"strings"
	"testing"

	"github.com/mikkeloscar/sshconfig"
)

func Test_readSSHConfig(t *testing.T) {
	tests := []struct {
		path   string
		host   string
		idFile string
		err    bool
	}{
		{
			path:   "./testdata/ssh_config_1",
			host:   "52.12.34.92",
			idFile: "/.ssh/id_rsa_test",
		},

		{
			path:   "./testdata/ssh_config_1",
			host:   "10.10.10.100",
			idFile: "/.ssh/id_rsa_test2",
		},

		{
			path: "./testdata/ssh_config_1",
			host: "10.10.10.102",
			err:  true,
		},
	}

	for _, c := range tests {
		h, err := readSSHConfig(c.host, c.path)
		if err != nil {
			if c.err {
				continue
			}

			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(h.IdentityFile, c.idFile) {
			t.Fatalf("expected %s but got %s", c.idFile, h.IdentityFile)
		}
	}
}

func Test_filterHosts(t *testing.T) {
	type hosts struct {
		host      string
		hostNames []string
	}

	tests := []struct {
		hosts    []hosts
		host     string
		expected string
	}{
		{
			hosts: []hosts{
				{host: "10.10.10.1", hostNames: []string{"test"}},
				{host: "10.10.10.2", hostNames: []string{"test2"}},
				{host: "10.10.10.3", hostNames: []string{"*"}},
			},
			host:     "10.10.10.2",
			expected: "10.10.10.2",
		},

		{
			hosts: []hosts{
				{host: "10.10.10.1", hostNames: []string{"test"}},
				{host: "10.10.10.2", hostNames: []string{"test2"}},
				{host: "10.10.10.3", hostNames: []string{"*"}},
			},
			host:     "10.10.10.3",
			expected: "10.10.10.3",
		},

		{
			hosts: []hosts{
				{host: "10.10.10.1", hostNames: []string{"test"}},
				{host: "10.10.10.2", hostNames: []string{"test2"}},
				{host: "10.10.10.3", hostNames: []string{"*"}},
			},
			host: "10.10.10.5",
		},
	}

	for _, c := range tests {
		var sshHosts []*sshconfig.SSHHost
		for _, h := range c.hosts {
			sshHosts = append(sshHosts, &sshconfig.SSHHost{Host: h.hostNames, HostName: h.host})
		}

		fh := filterHost(sshHosts, c.host)
		if fh == nil {
			if c.expected == "" {
				continue
			}

			t.Fatalf("expected %s but got nothing", c.expected)
		}

		if c.expected != fh.HostName {
			t.Fatalf("expected %s but got %s", c.expected, fh.HostName)
		}
	}
}
