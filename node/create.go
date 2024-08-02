// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// NodeParams is an input for CreateNodes
type NodeParams struct {
	// CloudParams contains the specs of the node being created on the Cloud Service (AWS / GCP)
	CloudParams *CloudParams

	// Count is how many Avalanche Nodes to be created during CreateNodes
	Count int

	// Roles pertain to whether the created node is going to be a Validator / API / Monitoring
	// node. See CheckRoles to see which combination of roles for a node is supported.
	Roles []SupportedRole

	// Network is whether the Validator / API node is meant to track AvalancheGo Primary Network
	// in Fuji / Mainnet / Devnet
	Network avalanche.Network

	// SubnetIDs is the list of subnet IDs that the created nodes will be tracking
	// For primary network, it should be empty
	SubnetIDs []string

	// SSHPrivateKeyPath is the file path to the private key of the SSH key pair that is used
	// to gain access to the created nodes
	SSHPrivateKeyPath string

	// AvalancheGoVersion is the version of Avalanche Go to install in the created node
	AvalancheGoVersion string

	// UseStaticIP is whether the created node should have static IP attached to it. Note that
	// assigning Static IP to a node may incur additional charges on AWS / GCP. There could also be
	// a limit to how many Static IPs you can have in a region in AWS & GCP.
	UseStaticIP bool
}

// CreateNodes launches the specified number of nodes on the selected cloud platform.
// The role of the node (Avalanche Validator / API / monitoring node) is
// specified through CloudParams in the input NodeParams
//
// Prior to calling CreateNodes, the credentials for AWS / GCP will first need to be set up:
// - To set up AWS credentials, more info can be found at https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html#file-format-creds
// Location on where to store AWS credentials file can be found at https://docs.aws.amazon.com/sdkref/latest/guide/file-location.html
//
// - To set up GCP credentials, more info can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-gcp#prerequisites
//
// When CreateNodes is used to create Avalanche Validator / API Nodes, it will:
//   - Launch the specified number of validator nodes on AWS / GCP
//   - Install Docker and have required dependencies (Avalanche Go, gcc, go) installed as Docker
//     images.
//   - Begin bootstrapping process on the nodes
//
// When CreateNodes is used to create Monitoring Nodes, it will:
//   - Launch the specified number of validator nodes on AWS / GCP
//   - Install Docker and have required dependencies (gcc, go, Prometheus, Grafana, Loki,
//     node_exporter) installed as Docker images.
//   - For more
//
// NOTE:
// Monitoring node enables you to have a centralized Grafana Dashboard where you can view
// metrics relevant to any Validator & API nodes that the monitoring node is linked to as well
// as a centralized logs for the X/P/C Chain and Subnet logs for the Validator & API nodes.
//
// An example on how the dashboard and logs look like can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-aws
//
// To enable centralized Grafana Dashboard and Logs, monitoring nodes will have to be linked to the
// Validator / API nodes by calling MonitorNodes function.
func CreateNodes(
	ctx context.Context,
	nodeParams *NodeParams,
) ([]Node, error) {
	nodes, err := createCloudInstances(ctx, *nodeParams.CloudParams, nodeParams.Count, nodeParams.UseStaticIP, nodeParams.SSHPrivateKeyPath)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wgResults := NodeResults{}
	// wait for all hosts to be ready and provision based on the role list
	for i, node := range nodes {
		wg.Add(1)
		go func(nodeResults *NodeResults, node Node) {
			defer wg.Done()
			if err := node.WaitForSSHShell(constants.SSHScriptTimeout); err != nil {
				nodeResults.AddResult(node.NodeID, nil, err)
				return
			}
			if err := provisionHost(node, nodeParams); err != nil {
				nodeResults.AddResult(node.NodeID, nil, err)
				return
			}
		}(&wgResults, node)
		nodes[i].Roles = nodeParams.Roles
	}
	wg.Wait()
	return nodes, wgResults.Error()
}

// preCreateCheck checks if the cloud parameters are valid.
func preCreateCheck(cp CloudParams, count int, sshPrivateKeyPath string) error {
	if count < 1 {
		return fmt.Errorf("count must be at least 1")
	}
	if err := cp.Validate(); err != nil {
		return err
	}
	if sshPrivateKeyPath != "" && !utils.FileExists(sshPrivateKeyPath) {
		return fmt.Errorf("ssh private key path %s does not exist", sshPrivateKeyPath)
	}
	return nil
}

// createCloudInstances launches the specified number of instances on the selected cloud platform.
func createCloudInstances(ctx context.Context, cp CloudParams, count int, useStaticIP bool, sshPrivateKeyPath string) ([]Node, error) {
	if err := preCreateCheck(cp, count, sshPrivateKeyPath); err != nil {
		return nil, err
	}
	nodes := make([]Node, 0, count)
	switch cp.Cloud() {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			cp.AWSConfig.AWSProfile,
			cp.Region,
		)
		if err != nil {
			return nil, err
		}
		instanceIds, err := ec2Svc.CreateEC2Instances(
			count,
			cp.ImageID,
			cp.InstanceType,
			cp.AWSConfig.AWSKeyPair,
			cp.AWSConfig.AWSSecurityGroupID,
			cp.AWSConfig.AWSVolumeIOPS,
			cp.AWSConfig.AWSVolumeThroughput,
			cp.AWSConfig.AWSVolumeType,
			cp.AWSConfig.AWSVolumeSize,
		)
		if err != nil {
			return nil, err
		}
		if len(instanceIds) != count {
			return nil, fmt.Errorf("failed to create all instances. Expected %d, got %d", count, len(instanceIds))
		}
		if err := ec2Svc.WaitForEC2Instances(instanceIds, types.InstanceStateNameRunning); err != nil {
			return nil, err
		}
		// elastic IP
		instanceEIPMap := make(map[string]string)
		if useStaticIP {
			for _, instanceID := range instanceIds {
				allocationID, publicIP, err := ec2Svc.CreateEIP(cp.Region)
				if err != nil {
					return nil, err
				}
				err = ec2Svc.AssociateEIP(instanceID, allocationID)
				if err != nil {
					return nil, err
				}
				instanceEIPMap[instanceID] = publicIP
			}
		} else {
			instanceEIPMap, err = ec2Svc.GetInstancePublicIPs(instanceIds)
			if err != nil {
				return nil, err
			}
		}
		for _, instanceID := range instanceIds {
			nodes = append(nodes, Node{
				NodeID:      instanceID,
				IP:          instanceEIPMap[instanceID],
				Cloud:       cp.Cloud(),
				CloudConfig: cp,
				SSHConfig: SSHConfig{
					User:           constants.AnsibleSSHUser,
					PrivateKeyPath: sshPrivateKeyPath,
				},
				Roles: nil,
			})
		}
		return nodes, nil
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			cp.GCPConfig.GCPProject,
			cp.GCPConfig.GCPCredentials,
		)
		if err != nil {
			return nil, err
		}
		staticIPs := []string{}
		if useStaticIP {
			staticIPs, err = gcpSvc.SetPublicIP(cp.GCPConfig.GCPZone, "", count)
			if err != nil {
				return nil, err
			}
		}
		computeInstances, err := gcpSvc.SetupInstances(
			cp.GCPConfig.GCPZone,
			cp.GCPConfig.GCPNetwork,
			cp.GCPConfig.GCPSSHKey,
			cp.ImageID,
			cp.InstanceType,
			staticIPs,
			1,
			cp.GCPConfig.GCPVolumeSize,
		)
		if err != nil {
			return nil, err
		}
		if len(computeInstances) != count {
			return nil, fmt.Errorf("failed to create all instances. Expected %d, got %d", count, len(computeInstances))
		}
		for _, computeInstance := range computeInstances {
			nodes = append(nodes, Node{
				NodeID:      computeInstance.Name,
				IP:          computeInstance.NetworkInterfaces[0].NetworkIP,
				Cloud:       cp.Cloud(),
				CloudConfig: cp,
				SSHConfig: SSHConfig{
					User:           constants.AnsibleSSHUser,
					PrivateKeyPath: sshPrivateKeyPath,
				},
				Roles: nil,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported cloud")
	}
	return nodes, nil
}

// provisionHost provisions a host with the given roles.
func provisionHost(node Node, nodeParams *NodeParams) error {
	if err := CheckRoles(nodeParams.Roles); err != nil {
		return err
	}
	if err := node.Connect(constants.SSHTCPPort); err != nil {
		return err
	}
	for _, role := range nodeParams.Roles {
		switch role {
		case Validator:
			if err := provisionAvagoHost(node, nodeParams); err != nil {
				return err
			}
		case API:
			if err := provisionAvagoHost(node, nodeParams); err != nil {
				return err
			}
		case Loadtest:
			if err := provisionLoadTestHost(node); err != nil {
				return err
			}
		case Monitor:
			if err := provisionMonitoringHost(node); err != nil {
				return err
			}
		case AWMRelayer:
			if err := provisionAWMRelayerHost(node); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported role %v", role)
		}
	}
	return nil
}

func provisionAvagoHost(node Node, nodeParams *NodeParams) error {
	const withMonitoring = true
	if err := node.RunSSHSetupNode(); err != nil {
		return err
	}
	if err := node.RunSSHSetupDockerService(); err != nil {
		return err
	}
	// provide dummy config for promtail
	if err := node.RunSSHSetupPromtailConfig("127.0.0.1", constants.AvalanchegoLokiPort, node.NodeID, "", ""); err != nil {
		return err
	}
	if err := node.ComposeSSHSetupNode(nodeParams.Network.HRP(), nodeParams.SubnetIDs, nodeParams.AvalancheGoVersion, withMonitoring); err != nil {
		return err
	}
	if err := node.StartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}
	return nil
}

func provisionLoadTestHost(node Node) error { // stub
	if err := node.ComposeSSHSetupLoadTest(); err != nil {
		return err
	}
	if err := node.RestartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}
	return nil
}

func provisionMonitoringHost(node Node) error {
	if err := node.RunSSHSetupDockerService(); err != nil {
		return err
	}
	if err := node.RunSSHSetupMonitoringFolders(); err != nil {
		return err
	}
	if err := node.ComposeSSHSetupMonitoring(); err != nil {
		return err
	}
	if err := node.RestartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}
	return nil
}

func provisionAWMRelayerHost(node Node) error { // stub
	if err := node.ComposeSSHSetupAWMRelayer(); err != nil {
		return err
	}
	return node.StartDockerComposeService(utils.GetRemoteComposeFile(), constants.ServiceAWMRelayer, constants.SSHLongRunningScriptTimeout)
}
