package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

const (
	testRemotePrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAr3sYWZ/qmmDXEr/hLqMVt1Kge8o33IVAqvDwW6o/26s8BVY3
SsgpAJo+jovJh7o+KCRedul4c6vT6j9mReo5OTyEt1tyN+Lq4YVPIGFqM5VyI+yj
v8WffjvGkMdfYb70qwXK8WhY2Ov9S+CWz7UK20Psc3ViledyVzRmx6pjsuWSGh7d
sN57XZQKEdouatleVfwrslaGeU4t4BPjcAo5TgOH7SIZLA14JLVDjUAVUOQ0PeBG
Doc1BHsPg5bLIFllXMZCSMqP3uh0mApWs1urFARGz13HCmeRKpmUAOvnphWvtQBi
piqkdXEYME5qxhaOLDRu1u+1nKpzqR4NIO5eWwIDAQABAoIBACMU6uNQEyjV32mC
LtSSCg9iV28oGE7f3PPPw12wBaA29YLjn541sezK6WK6E4os86w2ySPgvRHy8iTM
k/e6QcJtlOLLR2Rg2zBG5HDGyOKTKASClKIMMjycWrArC6iQ8n0WZWIpyElltHfs
6HmR6h+3zpeuig0J/lPsx/d22wOyiIgGUZQaBS/wm5+Nvhc4h+vJa+CqHgv9DyV5
pPJFRyb2Zm4HBmcU60OEn3Bfz7N2o+LsojxDmGONZotA0xFJZipi2a0FnaOo5gC1
HiJ4huLgHO5PCFSCq4ASQLCMOaWsLXhOKyRB5K6BZlkb0Jxqa0EsjBU/ZiMeIUTl
T5zB0zECgYEA6QyX4i3t0UXaQx1mZvYDchqWHDDp73BUbBvSOEkoxI5K825HHJZG
/RWoPAaBduI8GBFnXeZL37siBmIgjcBIAM8ZQsOYeMDBCZ0Rl8B/PbFgD3oYZnWx
LBfmCDyMBn+FFEgK+I82gHv23+0S6lwNVJWJT5yqKczQGcOlbUObWX0CgYEAwMMl
LY/r/Sxc5bTEe25MJoe7GdIKzYv7iEW+3cJyIlQl5KLbXnku9/GMmbKa0muQFvwf
ZURzC64iUxe0aBXZhZ4blnEpmPL2V677bQKYt7kuAkvGmIKtkB9eEUfuqq2LMSnA
ZXha27Yk+VZMn44Lu33lmOGnIsyar38JUgUT3rcCgYEA4CjBysc8sYdL2jRkM/K7
+7vA2RrfWPSl71qPSJ5t/SmUxDM1aTTO0U5MhHgJfQ08mjXbRAFox6lan+MwxsfF
Oin5dTNn6/YTdyjCLEH24aND738UCzF3Gh9enluGoBV2Lrzg+JXf6ZPTMS36fLvQ
AqTW72GUdyJOpAmAi+l0eo0CgYEAh0HnOns8mh+g58XaAMo6QOcI0GtkDSrmSo62
CSFBuPjWxgHIcBgUbwJdKBP9Mx0qtBYhaYuINm/vweJmQoriYTG6Jewtxhbl6Ezu
O2WRB4m20QGezrETXIfGF9YjkDvJ6atkR08WYgkDFLoLdgHUFbGXg5kDEBYFjJNV
VOofMS0CgYAeYUWFQYb4VjRprO2sKhQdMM/sJTrDISWuEn+gGGWsVCEn5JL+YOpS
4lMo9gaqRSEtxJNsDDRCW8fKf3CUn21ZBJaT4jNe7ZOVNXGJBnQ3zxqXEL1HPCIp
bBcn9sw/WgAgQQy+8UYH69V0SsY0LwWhVbi3nX9g+H/TAc9iZYffEQ==
-----END RSA PRIVATE KEY-----
`
)

var (
	testRemoteShell *Remote = nil
)

func newTestRemoteShell() *Remote {
	remote, err := NewRemote(getTestRemoteConfig())

	if err != nil {
		panic(err)
	}

	return &remote
}

func getTestRemoteConfig() RemoteConfig {
	key, err := ssh.ParsePrivateKey([]byte(testRemotePrivateKey))
	if err != nil {
		panic(err)
	}

	auth := []ssh.AuthMethod{ssh.PublicKeys(key)}

	return RemoteConfig{
		Address: "ssh.shell.local",
		// Address: "172.17.0.2", // use plain ip address for local testing
		Auth: auth,
	}
}

type testRemoteState struct {
	args  []string
	shell Remote
}

func (state *testRemoteState) handler(kind MessageType, result string) error {
	if kind == StdOut {
		result = "OUT: " + result
	} else if kind == StdErr {
		result = "ERR: " + result
	}

	state.args = append(state.args, result)
	return nil
}

func newTestRemoteState() testRemoteState {
	if testRemoteShell == nil {
		testRemoteShell = newTestRemoteShell()
	}

	state := testRemoteState{shell: *testRemoteShell}
	return state
}

func TestRemoteHostNameIsValid(test *testing.T) {
	state := newTestRemoteState()

	status, err := state.shell.Run("cd /etc", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 0, status)

	status, err = state.shell.Run("cat hostname", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 0, status)
	assert.Equal(test, "OUT: ssh.shell.local", state.args[0])
}

func TestRemoteReturnsError(test *testing.T) {
	state := newTestRemoteState()

	status, err := state.shell.Run("echo ERROR 1>&2 && false", state.handler)
	assert.NoError(test, err)
	assert.Equal(test, 1, status)
	assert.Equal(test, "ERR: ERROR", state.args[0])
}

func TestRemoteExitsWithoutError(test *testing.T) {
	shell, err := NewRemote(getTestRemoteConfig())
	assert.NoError(test, err)
	err = shell.Close()
	assert.NoError(test, err)
}

func TestRemoteExitsWithoutErrorWhileExecutingCommand(test *testing.T) {
	shell, err := NewRemote(getTestRemoteConfig())
	assert.NoError(test, err)
	go func() { shell.Run("sleep 100", nil) }()
	err = shell.Close()
	assert.NoError(test, err)
}
