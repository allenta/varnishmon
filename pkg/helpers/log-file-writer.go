package helpers

/*
	See:
		- https://github.com/client9/reopen.
		- https://stackoverflow.com/questions/34772012/capturing-panic-in-golang.
*/

import (
	"fmt"
	"os"
	"sync"
	"syscall"
)

type LogFileWriter struct {
	name     string
	mode     os.FileMode
	redirect bool
	mutex    sync.Mutex
	file     *os.File
	stdout   int
	stderr   int
}

func NewLogFileWriter(name string, mode os.FileMode, redirect bool) (*LogFileWriter, error) {
	// Keep references to stdout and stderr.
	var stdout, stderr int
	if redirect {
		var err error
		if stdout, err = syscall.Dup(int(os.Stdout.Fd())); err != nil {
			return nil, fmt.Errorf(
				"failed to duplicate stdout's file descriptor: %w", err)
		}
		if stderr, err = syscall.Dup(int(os.Stderr.Fd())); err != nil {
			return nil, fmt.Errorf(
				"failed to duplicate stderr's file descriptor: %w", err)
		}
	} else {
		stdout, stderr = -1, -1
	}

	writer := LogFileWriter{
		name:     name,
		mode:     mode,
		redirect: redirect,
		file:     nil,
		stdout:   stdout,
		stderr:   stderr,
	}
	if err := writer.unsafeReopen(); err != nil {
		return nil, fmt.Errorf("failed to reopen underlying file: %w", err)
	}
	return &writer, nil
}

func (fw *LogFileWriter) Write(p []byte) (int, error) {
	fw.mutex.Lock()
	n, err := fw.file.Write(p)
	fw.mutex.Unlock()
	if err != nil {
		return n, fmt.Errorf("failed to write underlying file: %w", err)
	}
	return n, nil
}

func (fw *LogFileWriter) Reopen() error {
	fw.mutex.Lock()
	err := fw.unsafeReopen()
	fw.mutex.Unlock()
	if err != nil {
		return fmt.Errorf("failed to reopen underlying file: %w", err)
	}
	return nil
}

func (fw *LogFileWriter) Close() error {
	// Undo stdout and stderr redirections.
	if fw.redirect {
		if err := PortableDup2(fw.stdout, int(os.Stdout.Fd())); err != nil {
			return fmt.Errorf("failed to undo stdout redirection: %w", err)
		}
		if err := PortableDup2(fw.stderr, int(os.Stderr.Fd())); err != nil {
			return fmt.Errorf("failed to undo stderr redirection: %w", err)
		}
	}

	// Close underlying file.
	fw.mutex.Lock()
	err := fw.file.Close()
	fw.mutex.Unlock()
	if err != nil {
		return fmt.Errorf("failed to close underlying file: %w", err)
	}
	return nil
}

func (fw *LogFileWriter) unsafeReopen() error {
	// Close previous file?
	if fw.file != nil {
		fw.file.Close()
		fw.file = nil
	}

	// Open new file.
	newFile, err := os.OpenFile(fw.name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, fw.mode) //nolint:nosnakecase
	if err != nil {
		return fmt.Errorf("failed to open underlying file: %w", err)
	}

	// Redirect stdout & stderr to the new file.
	if fw.redirect {
		files := []*os.File{os.Stdout, os.Stderr}
		for _, file := range files {
			if err := PortableDup2(int(newFile.Fd()), int(file.Fd())); err != nil {
				return fmt.Errorf("failed to redirect '%s' to underlying file: %w", file.Name(), err)
			}
		}
	}

	// Update underlying file.
	fw.file = newFile

	// Done!
	return nil
}
