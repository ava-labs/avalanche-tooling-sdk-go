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

	"github.com/ava-labs/avalanchego/utils/crypto/bls"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

// SSHConfig contains the configuration for connecting to a node over SSH
type SSHConfig struct {
	// Username to use when connecting to the node
	User string

	// Path to the private key to use when connecting to the node
	// If this is empty, the SSH agent will be used
	PrivateKeyPath string

	// Parameters to pass to the ssh command.
	// See man ssh_config(5) for more information
	// By defalult it's StrictHostKeyChecking=no
	Params map[string]string // additional parameters to pass to the ssh command
}

// Node is an output of CreateNodes
type Node struct {
	// NodeID is Avalanche Node ID of the node
	NodeID string

	// IP address of the node
	IP string

	// SSH configuration for the node
	SSHConfig SSHConfig

	// Cloud is the cloud service that the node is on
	// Full list of cloud service:
	// - AWS
	// - GCP
	// - Docker
	Cloud SupportedCloud

	// CloudConfig is the cloud specific configuration for the node
	CloudConfig CloudParams

	// connection to the node
	connection *goph.Client

	// Roles of the node
	// Full list of node roles:
	// - Validator
	// - API
	// - AWM Relayer
	// - Load Test
	// - Monitoring
	Roles []SupportedRole

	// Logger for node
	Logger avalanche.LeveledLogger

	// BLS provides a way to aggregate signatures off chain into a single signature that can be efficiently verified on chain.
	// For more information about how BLS is used on the P-Chain, please head to https://docs.avax.network/cross-chain/avalanche-warp-messaging/deep-dive#bls-multi-signatures-with-public-key-aggregation
	BlsSecretKey *bls.SecretKey
}

// NewNodeConnection creates a new SSH connection to the node
func NewNodeConnection(h *Node, port uint) (*goph.Client, error) {
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
		Callback: ssh.InsecureIgnoreHostKey(), // we don't verify node key ( similar to ansible)
	})
	if err != nil {
		return nil, err
	}
	return cl, nil
}

// GetConnection returns the SSH connection client for the Node.
// Returns a pointer to a goph.Client.
func (h *Node) GetConnection() *goph.Client {
	return h.connection
}

// GetSSHClient returns the SSH client for the Node.
// Returns a pointer to an ssh.Client.
func (h *Node) GetSSHClient() *ssh.Client {
	return h.connection.Client
}

// GetCloudID returns the cloudID for the node if it is a cloud node
func (h *Node) GetCloudID() string {
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
func (h *Node) Connect(port uint) error {
	if port == 0 {
		port = constants.SSHTCPPort
	}
	if h.connection != nil {
		return nil
	}
	var err error
	for i := 0; h.connection == nil && i < sshConnectionRetries; i++ {
		h.connection, err = NewNodeConnection(h, port)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %w", h.IP, err)
	}
	return nil
}

func (h *Node) Connected() bool {
	return h.connection != nil
}

func (h *Node) Disconnect() error {
	if h.connection == nil {
		return nil
	}
	err := h.connection.Close()
	return err
}

// Upload uploads a local file to a remote file on the node.
func (h *Node) Upload(localFile string, remoteFile string, timeout time.Duration) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	_, err := utils.CallWithTimeout(
		"upload",
		func() (interface{}, error) {
			return nil, h.connection.Upload(localFile, remoteFile)
		},
		timeout,
	)
	if err != nil {
		err = fmt.Errorf("%w for node %s", err, h.IP)
	}
	return err
}

// UploadBytes uploads a byte array to a remote file on the host.
func (h *Node) UploadBytes(data []byte, remoteFile string, timeout time.Duration) error {
	tmpFile, err := os.CreateTemp("", "NodeUploadBytes-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}
	return h.Upload(tmpFile.Name(), remoteFile, timeout)
}

// Download downloads a file from the remote server to the local machine.
func (h *Node) Download(remoteFile string, localFile string, timeout time.Duration) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(localFile), os.ModePerm); err != nil {
		return err
	}
	_, err := utils.CallWithTimeout(
		"download",
		func() (interface{}, error) {
			return nil, h.connection.Download(remoteFile, localFile)
		},
		timeout,
	)
	if err != nil {
		err = fmt.Errorf("%w for node %s", err, h.IP)
	}
	return err
}

// ReadFileBytes downloads a file from the remote server to a byte array
func (h *Node) ReadFileBytes(remoteFile string, timeout time.Duration) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "NodeDownloadBytes-*.tmp")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	if err := h.Download(remoteFile, tmpFile.Name(), timeout); err != nil {
		return nil, err
	}
	return os.ReadFile(tmpFile.Name())
}

// ExpandHome expands the ~ symbol to the home directory.
func (h *Node) ExpandHome(path string) string {
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
func (h *Node) MkdirAll(remoteDir string, timeout time.Duration) error {
	remoteDir = h.ExpandHome(remoteDir)
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	_, err := utils.CallWithTimeout(
		"mkdir",
		func() (interface{}, error) {
			return nil, h.UntimedMkdirAll(remoteDir)
		},
		timeout,
	)
	if err != nil {
		err = fmt.Errorf("%w for node %s", err, h.IP)
	}
	return err
}

// UntimedMkdirAll creates a folder on the remote server.
// Does not support timeouts on the operation.
func (h *Node) UntimedMkdirAll(remoteDir string) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	sftp, err := h.connection.NewSftp()
	if err != nil {
		return err
	}
	defer sftp.Close()
	return sftp.MkdirAll(remoteDir)
}

// Cmd returns a new command to be executed on the remote node.
func (h *Node) Cmd(ctx context.Context, name string, script string) (*goph.Cmd, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	return h.connection.CommandContext(ctx, name, script)
}

// Command executes a shell command on a remote node.
func (h *Node) Command(env []string, timeout time.Duration, script string) ([]byte, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd, err := h.connection.CommandContext(ctx, "", script)
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
func (h *Node) Commandf(env []string, timeout time.Duration, format string, args ...interface{}) ([]byte, error) {
	return h.Command(env, timeout, fmt.Sprintf(format, args...))
}

// Forward forwards the TCP connection to a remote address.
func (h *Node) Forward(httpRequest string, timeout time.Duration) ([]byte, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return nil, err
		}
	}
	return utils.Retry(
		utils.WrapContext(
			func() ([]byte, error) {
				return h.UntimedForward(httpRequest)
			},
		),
		timeout,
		3,
		fmt.Sprintf("failure on node %s post over ssh", h.IP),
	)
}

// UntimedForward forwards the TCP connection to a remote address.
// Does not support timeouts on the operation.
func (h *Node) UntimedForward(httpRequest string) ([]byte, error) {
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
		proxy, err = h.connection.DialTCP("tcp", nil, avalancheGoAddr)
		if err != nil {
			return nil, fmt.Errorf("unable to port forward to %s via %s", h.connection.RemoteAddr(), "ssh")
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
func (h *Node) FileExists(path string) (bool, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return false, err
		}
	}

	sftp, err := h.connection.NewSftp()
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
func (h *Node) CreateTempFile() (string, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return "", err
		}
	}
	sftp, err := h.connection.NewSftp()
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
func (h *Node) CreateTempDir() (string, error) {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return "", err
		}
	}
	sftp, err := h.connection.NewSftp()
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
func (h *Node) Remove(path string, recursive bool) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}
	sftp, err := h.connection.NewSftp()
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

// WaitForSSHShell waits for the SSH shell to be available on the node within the specified timeout.
func (h *Node) WaitForSSHShell(timeout time.Duration) error {
	if h.IP == "" {
		return fmt.Errorf("node IP is empty")
	}
	start := time.Now()
	if err := h.WaitForPort(constants.SSHTCPPort, timeout); err != nil {
		return err
	}

	deadline := start.Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout: SSH shell on node %s is not available after %ds", h.IP, int(timeout.Seconds()))
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

// StreamSSHCommand streams the execution of an SSH command on the node.
func (h *Node) StreamSSHCommand(env []string, timeout time.Duration, command string) error {
	if !h.Connected() {
		if err := h.Connect(0); err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	session, err := h.connection.NewSession()
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

// HasSystemDAvailable checks if systemd is available on a remote host.
func (h *Node) HasSystemDAvailable() bool {
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
