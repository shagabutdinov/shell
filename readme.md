Go shell library
================

Library that allows execute shell commands both locally and remotelly


Features
--------

  * Local shell commands execution

  * Remote shell commands execution (over ssh)

  * State preserving (previous commands result are saved: cd, variables and
    etc.)

  * Commands are executed as its run from normal shell

  * Note: running incomplete commands (e.g. `echo "TEST` - no final quote) will
    result execution stucking, so all commands should be verified carefully
    before execution

  * Note: output from stdout and stderr can come in different order from it was
    really sent


Installation
------------

```
go get github.com/shagabutdinov/shell
```


Usage
-----

Define message processor handler:

```
verify := func(error err) {
    if(err != nil) {
        panic(err)
    }
}

handler := func(outputType int, message string) {
    if(outputType == shell.Stdout) {
        log.Println("stdout: ", message)
    } else if(outputType == shell.Stdout) {
        log.Println("stderr: ", message)
    }
}
```

Run commands locally:

```
shell, err = shell.NewLocal(shell.LocalConfig{})
verify(err)

_, err := shell.Run("echo TEST 1>2", handler)
verify(err)

_, err := shell.Run("cd /var/lib", handler)
verify(err)

_, err := shell.Run(`echo "current path is: $(pwd)"`, handler)
verify(err)

status, err := shell.Run("/bin/false", handler)
verify(err)
log.Println("execution status is ", status) // execution status is 1
```

Run commands remotelly:

```
key, err := ssh.ParsePrivateKey([]byte(YOUR_PRIVATE_KEY))
verify(err)

shell, err := shell.NewRemote(shell.RemoteConfig{
    Host: "root@example.com:22",
    Auth: []ssh.AuthMethod{ssh.PublicKeys(key)},
})

verify(err)

_, err := shell.Run("cat /etc/hostname", handler)
verify(err)

_, err := shell.Run("cd /var/lib", handler)
verify(err)

_, err := shell.Run(`echo "current path is: $(pwd)"`, handler)
verify(err)

status, err := shell.Run("/bin/false", handler)
verify(err)
log.Println("execution status is ", status) // execution status is 1
```


Similar projects
----------------

* [runcmd](https://github.com/theairkit/runcmd)


License
-------

The MIT License (MIT)


Authors
-------

[Leonid Shagabutdinov](http://github.com/shagabutdinov)