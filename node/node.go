// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// SSHConfig contains the configuration for connecting to a host over SSH
type SSHConfig struct {
	// Username to use when connecting to the host
	User string

	// Path to the private key to use when connecting to the host
	// If this is empty, the SSH agent will be used
	PrivateKeyPath string

	// Parameters to pass to the ssh command.
	// See man ssh_config(5) for more information
	// By defalult it's StrictHostKeyChecking=no
	Params map[string]string // additional parameters to pass to the ssh command
}

// Host represents a cloud host that can be connected to over SSH
type Host struct {
	// ID of the host
	NodeID string

	// IP address of the host
	IP string

	// SSH configuration for the host
	SSHConfig SSHConfig

	// Cloud configuration for the host
	Cloud SupportedCloud

	// CloudConfig is the cloud specific configuration for the host
	CloudConfig CloudParams

	// Connection to the host
	Connection *goph.Client

	// Roles of the host
	Roles []SupportedRole

	// Logger for host
	Logger avalanche.LeveledLogger
}

// NewHostConnection creates a new SSH connection to the host
func NewHostConnection(h *Host, port uint) (*goph.Client, error) {
	if port == 0 {
		port = constants.SSHTCPPort
	}
	var (
		auth goph.Auth
		err  error
	)

	if h.SSHConfig.PrivateKeyPath == "" {
		auth, err = goph.UseAgent()
	} else {
		auth, err = goph.Key(h.SSHConfig.PrivateKeyPath, "")
	}
	if err != nil {
		return nil, err
	}
	cl, err := goph.NewConn(&goph.Config{
		User:    h.SSHConfig.User,
		Addr:    h.IP,
		Port:    port,
		Auth:    auth,
		Timeout: sshConnectionTimeout,
		// #nosec G106
		Callback: ssh.InsecureIgnoreHostKey(), // we don't verify host key ( similar to ansible)
	})
	if err != nil {
		return nil, err
	}
	return cl, nil
}

// GetCloudID returns the cloudID for the host if it is a cloud node
func (h *Host) GetCloudID() string {
	switch {
	case strings.HasPrefix(h.NodeID, constants.AWSNodeIDPrefix+"_"):
		return strings.TrimPrefix(h.NodeID, constants.AWSNodeIDPrefix+"_")
	case strings.HasPrefix(h.NodeID, constants.GCPNodeIDPrefix+"_"):
		return strings.TrimPrefix(h.NodeID, constants.GCPNodeIDPrefix+"_")
	default:
		return h.NodeID
	}
}

// Connect starts a new SSH connection with the provided private key.
func (h *Host) Connect(port uint) error {
	if port == 0 {
		port = constants.SSHTCPPort
	}
	if h.Connection != nil {
		return nil
	}
	var err error
	for i := 0; h.Connection == nil && i < sshConnectionRetries; i++ {
		h.Connection, err = NewHostConnection(h, port)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to host %s: %w", h.IP, err)
	}
	return nil
}

func (h *Host) Connected() bool {
	return h.Connection != nil
}

func (h *Host) Disconnect() error {
	if h.Connection == nil {
		return nil
	}
	err := h.Connection.Close()
	return err
}

// Upload uploads a local file to a remote file on the host.
func (h *Host) Upload(localFile string, remoteFile string, timeout time.Duration) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	_, err := utils.TimedFunction(
		func() (interface{}, error) {
			return nil, h.Connection.Upload(localFile, remoteFile)
		},
		"upload",
		timeout,
	)
	if err != nil {
		err = fmt.Errorf("%w for host %s", err, h.IP)
	}
	return err
}

// Download downloads a file from the remote server to the local machine.
func (h *Host) Download(remoteFile string, localFile string, timeout time.Duration) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(localFile), os.ModePerm); err != nil {
		return err
	}
	_, err := utils.TimedFunction(
		func() (interface{}, error) {
			return nil, h.Connection.Download(remoteFile, localFile)
		},
		"download",
		timeout,
	)
	if err != nil {
		err = fmt.Errorf("%w for host %s", err, h.IP)
	}
	return err
}

// ExpandHome expands the ~ symbol to the home directory.
func (h *Host) ExpandHome(path string) string {
	userHome := filepath.Join("/home", h.SSHConfig.User)
	if path == "" {
		return userHome
	}
	if len(path) > 0 && path[0] == '~' {
		path = filepath.Join(userHome, path[1:])
	}
	return path
}

// MkdirAll creates a folder on the remote server.
func (h *Host) MkdirAll(remoteDir string, timeout time.Duration) error {
	remoteDir = h.ExpandHome(remoteDir)
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	_, err := utils.TimedFunction(
		func() (interface{}, error) {
			return nil, h.UntimedMkdirAll(remoteDir)
		},
		"mkdir",
		timeout,
	)
	if err != nil {
		err = fmt.Errorf("%w for host %s", err, h.IP)
	}
	return err
}

// UntimedMkdirAll creates a folder on the remote server.
// Does not support timeouts on the operation.
func (h *Host) UntimedMkdirAll(remoteDir string) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	sftp, err := h.Connection.NewSftp()
	if err != nil {
		return err
	}
	defer sftp.Close()
	return sftp.MkdirAll(remoteDir)
}

// Cmd returns a new command to be executed on the remote host.
func (h *Host) Cmd(ctx context.Context, name string, script string) (*goph.Cmd, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	return h.Connection.CommandContext(ctx, name, script)
}

// Command executes a shell command on a remote host.
func (h *Host) Command(env []string, timeout time.Duration, script string) ([]byte, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd, err := h.Connection.CommandContext(ctx, "", script)
	if err != nil {
		return nil, err
	}
	if env != nil {
		cmd.Env = env
	}
	output, err := cmd.CombinedOutput()
	return output, err
}

// Commandf is a shorthand for Command with a formatted script.
func (h *Host) Commandf(env []string, timeout time.Duration, format string, args ...interface{}) ([]byte, error) {
	return h.Command(env, timeout, fmt.Sprintf(format, args...))
}

// Forward forwards the TCP connection to a remote address.
func (h *Host) Forward(httpRequest string, timeout time.Duration) ([]byte, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	retI, err := utils.TimedFunctionWithRetry(
		func() (interface{}, error) {
			return h.UntimedForward(httpRequest)
		},
		"post over ssh",
		timeout,
		3,
		2*time.Second,
	)
	if err != nil {
		err = fmt.Errorf("%w for host %s", err, h.IP)
	}
	ret := []byte(nil)
	if retI != nil {
		ret = retI.([]byte)
	}
	return ret, err
}

// UntimedForward forwards the TCP connection to a remote address.
// Does not support timeouts on the operation.
func (h *Host) UntimedForward(httpRequest string) ([]byte, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	avalancheGoEndpoint := strings.TrimPrefix(constants.LocalAPIEndpoint, "http://")
	avalancheGoAddr, err := net.ResolveTCPAddr("tcp", avalancheGoEndpoint)
	if err != nil {
		return nil, err
	}
	var proxy net.Conn
	if utils.IsE2E() {
		avalancheGoEndpoint = fmt.Sprintf("%s:%d", utils.E2EConvertIP(h.IP), constants.AvalanchegoAPIPort)
		proxy, err = net.Dial("tcp", avalancheGoEndpoint)
		if err != nil {
			return nil, fmt.Errorf("unable to port forward E2E to %s", avalancheGoEndpoint)
		}
	} else {
		proxy, err = h.Connection.DialTCP("tcp", nil, avalancheGoAddr)
		if err != nil {
			return nil, fmt.Errorf("unable to port forward to %s via %s", h.Connection.RemoteAddr(), "ssh")
		}
	}

	defer proxy.Close()
	// send request to server
	if _, err = proxy.Write([]byte(httpRequest)); err != nil {
		return nil, err
	}
	// Read and print the server's response
	response := make([]byte, maxResponseSize)
	responseLength, err := proxy.Read(response)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewReader(response[:responseLength]))
	parsedResponse, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil, err
	}
	buffer := new(bytes.Buffer)
	if _, err = buffer.ReadFrom(parsedResponse.Body); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// FileExists checks if a file exists on the remote server.
func (h *Host) FileExists(path string) (bool, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return false, err
		}
	}

	sftp, err := h.Connection.NewSftp()
	if err != nil {
		return false, nil
	}
	defer sftp.Close()
	_, err = sftp.Stat(path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// CreateTemp creates a temporary file on the remote server.
func (h *Host) CreateTempFile() (string, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return "", err
		}
	}
	sftp, err := h.Connection.NewSftp()
	if err != nil {
		return "", err
	}
	defer sftp.Close()
	tmpFileName := filepath.Join("/tmp", utils.RandomString(10))
	_, err = sftp.Create(tmpFileName)
	if err != nil {
		return "", err
	}
	return tmpFileName, nil
}

// CreateTempDir creates a temporary directory on the remote server.
func (h *Host) CreateTempDir() (string, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return "", err
		}
	}
	sftp, err := h.Connection.NewSftp()
	if err != nil {
		return "", err
	}
	defer sftp.Close()
	tmpDirName := filepath.Join("/tmp", utils.RandomString(10))
	err = sftp.Mkdir(tmpDirName)
	if err != nil {
		return "", err
	}
	return tmpDirName, nil
}

// Remove removes a file on the remote server.
func (h *Host) Remove(path string, recursive bool) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	sftp, err := h.Connection.NewSftp()
	if err != nil {
		return err
	}
	defer sftp.Close()
	if recursive {
		// return sftp.RemoveAll(path) is very slow
		_, err := h.Commandf(nil, constants.SSHLongRunningScriptTimeout, "rm -rf %s", path)
		return err
	} else {
		return sftp.Remove(path)
	}
}

// WaitForSSHShell waits for the SSH shell to be available on the host within the specified timeout.
func (h *Host) WaitForSSHShell(timeout time.Duration) error {
	if h.IP == "" {
		return fmt.Errorf("host IP is empty")
	}
	start := time.Now()
	if err := h.WaitForPort(constants.SSHTCPPort, timeout); err != nil {
		return err
	}

	deadline := start.Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: SSH shell on host %s is not available after %ds", h.IP, int(timeout.Seconds()))
		}
		if err := h.Connect(0); err != nil {
			time.Sleep(constants.SSHSleepBetweenChecks)
			continue
		}
		if h.Connected() {
			output, err := h.Command(nil, timeout, "echo")
			if err == nil || len(output) > 0 {
				return nil
			}
		}
		time.Sleep(constants.SSHSleepBetweenChecks)
	}
}

// StreamSSHCommand streams the execution of an SSH command on the host.
func (h *Host) StreamSSHCommand(command string, env []string, timeout time.Duration) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	session, err := h.Connection.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	for _, item := range env {
		envPair := strings.SplitN(item, "=", 2)
		if len(envPair) != 2 {
			return fmt.Errorf("invalid env variable %s", item)
		}
		if err := session.Setenv(envPair[0], envPair[1]); err != nil {
			return err
		}
	}
	// Use a WaitGroup to synchronize goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := consumeOutput(ctx, stdout); err != nil {
			fmt.Printf("Error reading stdout: %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := consumeOutput(ctx, stderr); err != nil {
			fmt.Printf("Error reading stderr: %v\n", err)
		}
	}()

	if err := session.Run(command); err != nil {
		return fmt.Errorf("failed to run command %s: %w", command, err)
	}
	wg.Wait()
	return nil
}

func consumeOutput(ctx context.Context, output io.Reader) error {
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		// Check if the context is done
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
	return scanner.Err()
}

// HasSystemDAvaliable checks if systemd is available on a remote host.
func (h *Host) IsSystemD() bool {
	// check for the folder
	if _, err := h.FileExists("/run/systemd/system"); err != nil {
		return false
	}
	tmpFile, err := os.CreateTemp("", "avalanchecli-proc-systemd-*.txt")
	if err != nil {
		return false
	}
	defer os.Remove(tmpFile.Name())
	// check for the service
	if err := h.Download("/proc/1/comm", tmpFile.Name(), constants.SSHFileOpsTimeout); err != nil {
		return false
	}
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == "systemd"
}
