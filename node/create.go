// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"sync"

	awsAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/aws"
	gcpAPI "github.com/ava-labs/avalanche-tooling-sdk-go/cloud/gcp"
	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// preCreateCheck checks if the cloud parameters are valid.
func preCreateCheck(cp CloudParams, count int) error {
	if count < 1 {
		return fmt.Errorf("count must be at least 1")
	}
	if err := cp.Validate(); err != nil {
		return err
	}
	return nil
}

// Create creates a new node.
// If wait is true, this function will block until the node is ready.
func createInstanceList(ctx context.Context, cp CloudParams, count int) ([]Host, error) {
	if err := preCreateCheck(cp, count); err != nil {
		return nil, err
	}
	hosts := make([]Host, 0, count)
	switch cp.Cloud() {
	case AWSCloud:
		ec2Svc, err := awsAPI.NewAwsCloud(
			ctx,
			cp.AWSProfile,
			cp.Region,
		)
		if err != nil {
			return nil, err
		}
		instanceIds, err := ec2Svc.CreateEC2Instances(
			cp.Name,
			count,
			cp.Image,
			cp.InstanceType,
			cp.AWSKeyPair,
			cp.AWSSecurityGroupID,
			cp.AWSVolumeIOPS,
			cp.AWSVolumeThroughput,
			cp.AWSVolumeType,
			cp.AWSVolumeSize,
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
		instanceEIPMap, err := ec2Svc.GetInstancePublicIPs(instanceIds)
		if err != nil {
			return nil, err
		}
		for _, instanceID := range instanceIds {
			hosts = append(hosts, Host{
				NodeID:      instanceID,
				IP:          instanceEIPMap[instanceID],
				Cloud:       cp.Cloud(),
				CloudConfig: cp,
				SSHConfig: SSHConfig{
					User: constants.AnsibleSSHUser,
				},
				Roles: nil,
			})
		}
		return hosts, nil
	case GCPCloud:
		gcpSvc, err := gcpAPI.NewGcpCloud(
			ctx,
			cp.GCPProject,
			cp.GCPCredentials,
		)
		if err != nil {
			return nil, err
		}
		computeInstances, err := gcpSvc.SetupInstances(
			cp.Name,
			cp.GCPZone,
			cp.GCPNetwork,
			cp.GCPSSHKey,
			cp.Image,
			cp.Name,
			cp.InstanceType,
			[]string{cp.StaticIP},
			1,
			cp.GCPVolumeSize,
		)
		if err != nil {
			return nil, err
		}
		if len(computeInstances) != count {
			return nil, fmt.Errorf("failed to create all instances. Expected %d, got %d", count, len(computeInstances))
		}
		for _, computeInstance := range computeInstances {
			hosts = append(hosts, Host{
				NodeID:      computeInstance.Name,
				IP:          computeInstance.NetworkInterfaces[0].NetworkIP,
				Cloud:       cp.Cloud(),
				CloudConfig: cp,
				SSHConfig: SSHConfig{
					User: constants.AnsibleSSHUser,
				},
				Roles: nil,
			})
		}
	default:
		return nil, fmt.Errorf("unsupported cloud")
	}
	return hosts, nil
}

// CreateInstanceList creates a list of nodes.
func CreateInstanceList(
	ctx context.Context,
	cp CloudParams,
	count int,
	roles []SupportedRole,
	networkID string,
	avalancheGoVersion string,
	avalancheCliVersion string,
	withMonitoring bool,
) ([]Host, error) {
	hosts, err := createInstanceList(ctx, cp, count)
	if err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	wgResults := NodeResults{}
	// wait for all hosts to be ready and provision based on the role list
	for _, host := range hosts {
		wg.Add(1)
		go func(NodeResults *NodeResults, host Host) {
			defer wg.Done()
			if err := host.WaitForSSHShell(constants.SSHScriptTimeout); err != nil {
				NodeResults.AddResult(host.NodeID, nil, err)
				return
			}
			if err := provisionHost(host, roles, networkID, avalancheGoVersion, avalancheCliVersion, withMonitoring); err != nil {
				NodeResults.AddResult(host.NodeID, nil, err)
				return
			}
		}(&wgResults, host)
		host.Roles = roles
	}
	wg.Wait()
	if wgResults.HasErrors() {
		// if there are errors, collect and return them with nodeIds
		hostErrorMap := wgResults.GetErrorHostMap()
		errStr := ""
		for nodeID, err := range hostErrorMap {
			errStr += fmt.Sprintf("NodeID: %s, Error: %s\n", nodeID, err)
		}
		return nil, fmt.Errorf("failed to provision all hosts: %s", errStr)

	}

	return hosts, nil
}

// provisionHost provisions a host with the given roles.
func provisionHost(host Host, roles []SupportedRole, networkID string, avalancheGoVersion string, avalancheCliVersion string, withMonitoring bool) error {
	if err := CheckRoles(roles); err != nil {
		return err
	}
	if err := host.Connect(constants.SSHTCPPort); err != nil {
		return err
	}
	for _, role := range roles {
		switch role {
		case Validator:
			if err := provisionAvagoHost(host, networkID, avalancheGoVersion, avalancheCliVersion, withMonitoring); err != nil {
				return err
			}
		case API:
			if err := provisionAvagoHost(host, networkID, avalancheGoVersion, avalancheCliVersion, withMonitoring); err != nil {
				return err
			}
		case Loadtest:
			if err := provisionLoadTestHost(host); err != nil {
				return err
			}
		case Monitor:
			if err := provisionMonitoringHost(host); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported role %s", role)
		}
		return nil
	}
	return nil
}

func provisionAvagoHost(host Host, networkID string, avalancheGoVersion string, avalancheCliVersion string, withMonitoring bool) error {
	if err := host.RunSSHSetupNode(avalancheCliVersion); err != nil {
		return err
	}
	if err := host.RunSSHSetupDockerService(); err != nil {
		return err
	}
	if err := host.ComposeSSHSetupNode(networkID, avalancheGoVersion, withMonitoring); err != nil {
		return err
	}
	if err := host.StartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}
	return nil
}

func provisionLoadTestHost(host Host) error { //stub
	if err := host.ComposeSSHSetupLoadTest(); err != nil {
		return err
	}
	if err := host.RestartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}
	return nil
}

func provisionMonitoringHost(host Host) error { //stub
	if err := host.ComposeSSHSetupMonitoring(); err != nil {
		return err
	}
	if err := host.RestartDockerCompose(constants.SSHScriptTimeout); err != nil {
		return err
	}
	return nil
}

func provisionAWMRelayerHost(host Host) error { //stub
	if err := host.ComposeSSHSetupAWMRelayer(); err != nil {
		return err
	}
	return host.StartDockerComposeService(utils.GetRemoteComposeFile(), "awm-relayer", constants.SSHLongRunningScriptTimeout)
}
