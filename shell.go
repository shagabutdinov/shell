package shell

import (
	"io"
	"regexp"
	"strconv"
	"strings"
)

const (
	Stdout = iota
	stdoutComplete
	Stderr
	stderrComplete
	fatal
)

type message struct {
	kind    int
	message string
	err     error
}

type shell struct {
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	limit  int

	messages chan message
}

type Shell interface {
	Run(command string, handler func(int, string) error) (int, error)
	Close() error
}

func (shell shell) Run(
	command string,
	handler func(int, string) error,
) (int, error) {
	query := strings.TrimRight(command, "\n") + "\n" +
		"echo -n __SHELL_EXIT_STATUS_$?__ | tee /dev/stderr\n"

	if _, err := shell.stdin.Write([]byte(query)); err != nil {
		return -1, err
	}

	result, err := shell.wait(handler)
	return result, err
}

func (shell shell) start() {
	go func() {
		shell.read(shell.stdout, Stdout, stdoutComplete)
	}()

	go func() {
		shell.read(shell.stderr, Stderr, stderrComplete)
	}()
}

var (
	exitStatusRegexp = regexp.MustCompile(`__SHELL_EXIT_STATUS_(\w*)__`)
)

func (shell shell) read(reader io.Reader, kind int, comlete int) {
	buffer := ""

	for {
		line := make([]byte, 1024)
		count, err := reader.Read(line)

		if err != nil {
			shell.messages <- message{fatal, "", err}
			break
		}

		buffer += string(line[:count])

		matches := exitStatusRegexp.FindStringSubmatch(buffer)
		if len(matches) > 0 {
			parts := strings.SplitN(buffer, matches[0], 2)

			lines := strings.Split(strings.TrimRight(parts[0], "\n"), "\n")
			for _, line := range lines {
				if len(line) > 0 {
					shell.messages <- message{kind, line, nil}
				}
			}

			shell.messages <- message{comlete, matches[1], nil}
			buffer = parts[1]
		} else if strings.Contains(buffer, "\n") {
			lines := strings.Split(buffer, "\n")
			for _, line := range lines[:len(lines)-1] {
				shell.messages <- message{kind, line, nil}
			}

			buffer = lines[len(lines)-1]
		}

		if shell.limit != 0 && len(buffer) > shell.limit {
			buffer = buffer[len(buffer)-shell.limit:]
		}
	}
}

func (shell shell) wait(handler func(int, string) error) (int, error) {
	result := "-1"
	var handlerErr error

	stdoutCompleted := false
	stderrCompleted := false

	for {
		message := <-shell.messages
		if message.kind == fatal || message.err != nil {
			return -1, message.err
		}

		if message.kind == stdoutComplete {
			stdoutCompleted = true
			result = message.message
			if stderrCompleted {
				break
			}

			continue
		}

		if message.kind == stderrComplete {
			stderrCompleted = true
			result = message.message
			if stdoutCompleted {
				break
			}

			continue
		}

		if message.kind == Stdout && stdoutCompleted {
			continue
		}

		if message.kind == Stderr && stderrCompleted {
			continue
		}

		if handler != nil && handlerErr == nil {
			err := handler(message.kind, message.message)
			if err != nil {
				handlerErr = err
			}
		}
	}

	status, err := strconv.Atoi(result)
	if err != nil {
		return -1, err
	}

	return status, handlerErr
}

func (shell *shell) close() error {
	_, err := shell.stdin.Write([]byte("exit"))
	if err != nil {
		return err
	}

	stdinErr := shell.stdin.Close()
	stderrErr := shell.stdout.Close()
	stdoutErr := shell.stderr.Close()

	if stdinErr != nil {
		return stdinErr
	}

	if stderrErr != nil {
		return stderrErr
	}

	if stdoutErr != nil {
		return stdoutErr
	}

	return nil
}
