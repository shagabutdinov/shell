package shell

import (
	"io"
	"strings"

	"golang.org/x/crypto/ssh"
)

type reader struct {
	reader io.Reader
}

func (reader reader) Read(bytes []byte) (int, error) {
	return reader.reader.Read(bytes)
}

func (reader reader) Close() error {
	return nil
}

type RemoteConfig struct {
	Address   string
	Auth      []ssh.AuthMethod
	LineLimit int
}

func NewRemote(config RemoteConfig) (*Remote, error) {
	address := config.Address
	user := "root"
	if strings.Contains(address, "@") {
		parts := strings.Split("@", config.Address)
		user = parts[0]
		address = parts[1]
	}

	clientConfig := &ssh.ClientConfig{User: user, Auth: config.Auth}
	if !strings.Contains(address, ":") {
		address += ":22"
	}

	client, err := ssh.Dial("tcp", address, clientConfig)
	if err != nil {
		return nil, err
	}

	shell := &Remote{client: client}

	shell.limit = config.LineLimit
	shell.messages = make(chan message, 4096)

	shell.session, err = client.NewSession()
	if err != nil {
		return shell, err
	}
	shell.stdin, err = shell.session.StdinPipe()
	if err != nil {
		return shell, err
	}

	stdout, err := shell.session.StdoutPipe()
	shell.stdout = reader{stdout}
	if err != nil {
		return shell, err
	}

	stderr, err := shell.session.StderrPipe()
	shell.stderr = reader{stderr}
	if err != nil {
		return shell, err
	}

	err = shell.session.Start("/bin/sh")
	if err != nil {
		return shell, err
	}

	shell.start()

	return shell, nil
}

type Remote struct {
	shell
	client  *ssh.Client
	session *ssh.Session
}

func (shell *Remote) Close() error {
	closeErr := shell.close()
	sessionCloseErr := shell.session.Close()
	clientCloseErr := shell.client.Close()

	if closeErr != nil {
		return closeErr
	}

	if sessionCloseErr != nil {
		return sessionCloseErr
	}

	if clientCloseErr != nil {
		return clientCloseErr
	}

	return nil
}
