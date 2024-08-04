package main

import (
	"bytes"
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
