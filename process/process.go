// Copyright (C) 2022, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.
package process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

type RunFile struct {
	Pid int `json:"pid"`
}

func saveRunFile(pid int, runFilePath string) error {
	rf := RunFile{
		Pid: pid,
	}
	bs, err := json.Marshal(&rf)
	if err != nil {
		return err
	}
	if err := os.WriteFile(runFilePath, bs, constants.WriteReadReadPerms); err != nil {
		return fmt.Errorf("could not write awm relayer run file to %s: %w", runFilePath, err)
	}
	return nil
}

func loadRunFile(runFilePath string) (int, error) {
	var pid int
	if runFilePath != "" {
		if !utils.FileExists(runFilePath) {
			return 0, fmt.Errorf("run file %s does not exist", runFilePath)
		}
		bs, err := os.ReadFile(runFilePath)
		if err != nil {
			return 0, err
		}
		rf := RunFile{}
		if err := json.Unmarshal(bs, &rf); err != nil {
			return 0, err
		}
		pid = rf.Pid
	}
	return pid, nil
}

func removeRunFile(runFilePath string) error {
	if runFilePath != "" {
		err := os.Remove(runFilePath)
		if err != nil {
			err = fmt.Errorf("failed removing relayer run file %s: %w", runFilePath, err)
		}
		return err
	}
	return nil
}

func Execute(
	binPath string,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	runFilePath string,
	setupTime time.Duration,
) (int, error) {
	cmd := exec.Command(binPath, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return 0, err
	}
	if runFilePath != "" {
		if err := saveRunFile(cmd.Process.Pid, runFilePath); err != nil {
			return 0, err
		}
	}
	if setupTime > 0 {
		ch := make(chan struct{})
		go func() {
			_ = cmd.Wait()
			ch <- struct{}{}
		}()
		time.Sleep(setupTime)
		select {
		case <-ch:
			return 0, fmt.Errorf("process stopped during setup")
		default:
		}
	}
	return cmd.Process.Pid, nil
}

func GetProcess(pid int) (*os.Process, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}
	if err := proc.Signal(syscall.Signal(0)); err != nil {
		// sometimes FindProcess returns without error, but Signal 0 will surely fail if the process doesn't exist
		return nil, err
	}
	return proc, nil
}

func IsRunning(pid int, runFilePath string) (bool, int, *os.Process, error) {
	if pid != 0 && runFilePath != "" {
		return false, 0, nil, fmt.Errorf("either provide a pid or a runFile to check a process")
	}
	if runFilePath != "" {
		var err error
		if pid, err = loadRunFile(runFilePath); err != nil {
			return false, 0, nil, err
		}
	}
	proc, err := GetProcess(pid)
	if err != nil {
		return false, 0, nil, removeRunFile(runFilePath)
	}
	return true, pid, proc, nil
}

func Cleanup(
	pid int,
	runFilePath string,
	tmpDir string,
	poolTime time.Duration,
	timeout time.Duration,
) error {
	if tmpDir != "" {
		if err := os.RemoveAll(tmpDir); err != nil {
			return err
		}
	}
	isRunning, pid, proc, err := IsRunning(pid, runFilePath)
	if err != nil {
		return err
	}
	if isRunning {
		waitedCh := make(chan struct{})
		go func() {
			for {
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					if errors.Is(err, os.ErrProcessDone) {
						close(waitedCh)
						return
					} else {
						fmt.Printf("failure checking to process pid %d aliveness due to: %s\n", proc.Pid, err)
					}
				}
				time.Sleep(poolTime)
			}
		}()
		if err := proc.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed sending interrupt signal to relayer process with pid %d: %w", pid, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		select {
		case <-ctx.Done():
			if err := proc.Signal(os.Kill); err != nil {
				return fmt.Errorf("failed killing relayer process with pid %d: %w", pid, err)
			}
		case <-waitedCh:
		}
		return removeRunFile(runFilePath)
	}
	return nil
}
