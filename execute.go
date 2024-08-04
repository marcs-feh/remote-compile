package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"time"
)

type ExecStatus struct {
	Success bool
	Stdout []byte
	Stderr []byte
	Elapsed time.Duration
}

const CommandBufferSize = 512; // Size of stdout and stderr buffers, in KiB

func Execute(binName string, args ...string) ExecStatus {
	cmd := exec.Command(binName, args...)

	size := CommandBufferSize * 2 * 1024
	buf := make([]byte, size)

	outBuf := bytes.NewBuffer(buf[:size])
	errBuf := bytes.NewBuffer(buf[size:])

	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	begin := time.Now()
	cmdErr := cmd.Run()
	elapsed := time.Since(begin)

	status := ExecStatus {
		Success: cmdErr == nil,
		Stdout: outBuf.Bytes(),
		Stderr: errBuf.Bytes(),
		Elapsed: elapsed,
	}

	return status
}

func buildSource(filename string, sourceData string, builder LanguageBuilder) ExecStatus {
	filename = filename + "." + builder.Ext()

	file, err := os.Create(filename)
	if !errors.Is(err, os.ErrExist) {
		panic(err.Error())
	}
	if err := file.Chmod(0o700); err != nil {
		panic(err.Error())
	}

	file.Write([]byte(sourceData))

	args := builder.Build(filename)
	status := Execute(args[0], args[1:]...)
	return status
}

