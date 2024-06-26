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

type NodeParams struct {
	CloudParams         *CloudParams
	Count               int
	Roles               []SupportedRole
	Network             avalanche.Network
	SSHPrivateKey       string
	AvalancheGoVersion  string
	AvalancheCliVersion string
	UseStaticIP         bool
}

// CreateNodes creates a list of nodes.
func CreateNodes(
	ctx context.Context,
	nodeParams *NodeParams,
) ([]Node, error) {
	nodes, err := createCloudInstances(ctx, *nodeParams.CloudParams, nodeParams.Count, nodeParams.UseStaticIP, nodeParams.SSHPrivateKey)
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
	if err := node.RunSSHSetupNode(nodeParams.AvalancheCliVersion); err != nil {
		return err
	}
	if err := node.RunSSHSetupDockerService(); err != nil {
		return err
	}
	// provide dummy config for promtail
	if err := node.RunSSHSetupPromtailConfig("127.0.0.1", constants.AvalanchegoLokiPort, node.NodeID, "", ""); err != nil {
		return err
	}
	if err := node.ComposeSSHSetupNode(nodeParams.Network.HRP(), nodeParams.AvalancheGoVersion, withMonitoring); err != nil {
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
