// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ava-labs/avalanche-tooling-sdk-go/avalanche"
	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

func TestCreateNodes(t *testing.T) {
	ctx := context.Background()
	// Get the default cloud parameters for AWS
	cp, err := GetDefaultCloudParams(ctx, AWSCloud)
	if err != nil {
		panic(err)
	}

	fmt.Println("Creating security group...")
	//securityGroupName := "SECURITY_GROUP_NAME"
	securityGroupName := "rs_security_group_sdk"
	sgID := "sg-045199f593930a378"
	//sgID, err := awsAPI.CreateSecurityGroup(ctx, securityGroupName, cp.AWSConfig.AWSProfile, cp.Region)
	//if err != nil {
	//	panic(err)
	//}
	fmt.Println("Security group successfully created")
	// Set the security group we are using when creating our Avalanche Nodes
	cp.AWSConfig.AWSSecurityGroupID = sgID
	cp.AWSConfig.AWSSecurityGroupName = securityGroupName

	fmt.Println("Creating key pair...")
	//keyPairName := "KEY_PAIR_NAME"
	keyPairName := "rs_key_pair_sdk"
	//sshPrivateKeyPath := utils.ExpandHome("PRIVATE_KEY_FILEPATH")
	sshPrivateKeyPath := utils.ExpandHome("/Users/raymondsukanto/.ssh/rs_key_pair_sdk.pem")
	//if err := awsAPI.CreateSSHKeyPair(ctx, cp.AWSConfig.AWSProfile, cp.Region, keyPairName, sshPrivateKeyPath); err != nil {
	//	panic(err)
	//}
	fmt.Println("Key pair successfully created and saved into sshPrivateKeyPath")
	// Set the key pair we are using when creating our Avalanche Nodes
	cp.AWSConfig.AWSKeyPair = keyPairName

	// Avalanche-CLI is installed in nodes to enable them to join subnets as validators
	// Avalanche-CLI dependency by Avalanche nodes will be deprecated in the next release
	// of Avalanche Tooling SDK
	const (
		avalancheGoVersion = "v1.11.8"
	)

	// Create two new Avalanche Validator nodes on Fuji Network on AWS without Elastic IPs
	// attached. Once CreateNodes is completed, the validators will begin bootstrapping process
	// to Primary Network in Fuji Network. Nodes need to finish bootstrapping process
	// before they can validate Avalanche Primary Network / Subnet.
	//
	// SDK function for nodes to start validating Primary Network / Subnet will be available
	// in the next Avalanche Tooling SDK release.
	hosts, err := CreateNodes(ctx,
		&NodeParams{
			CloudParams:        cp,
			Count:              2,
			Roles:              []SupportedRole{Validator},
			Network:            avalanche.FujiNetwork(),
			AvalancheGoVersion: avalancheGoVersion,
			UseStaticIP:        false,
			SSHPrivateKeyPath:  sshPrivateKeyPath,
		})
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully created Avalanche Validators")

	const (
		sshTimeout        = 120 * time.Second
		sshCommandTimeout = 10 * time.Second
	)

	// Examples showing how to run ssh commands on the created nodes
	for _, h := range hosts {
		// Wait for the host to be ready (only needs to be done once for newly created nodes)
		fmt.Println("Waiting for SSH shell")
		if err := h.WaitForSSHShell(sshTimeout); err != nil {
			panic(err)
		}
		fmt.Println("SSH shell ready to execute commands")
		// Run a command on the host
		if output, err := h.Commandf(nil, sshCommandTimeout, "echo 'Hello, %s!'", "World"); err != nil {
			panic(err)
		} else {
			fmt.Println(string(output))
		}
		// sleep for 10 seconds allowing AvalancheGo container to start
		time.Sleep(10 * time.Second)
		// check if avalanchego is running
		if output, err := h.Commandf(nil, sshCommandTimeout, "docker ps"); err != nil {
			panic(err)
		} else {
			fmt.Println(string(output))
		}
	}

	fmt.Println("Creating monitoring node...")

	// Create a monitoring node.
	// Monitoring node enables you to have a centralized Grafana Dashboard where you can view
	// metrics relevant to any Validator & API nodes that the monitoring node is linked to as well
	// as a centralized logs for the X/P/C Chain and Subnet logs for the Validator & API nodes.
	// An example on how the dashboard and logs look like can be found at https://docs.avax.network/tooling/cli-create-nodes/create-a-validator-aws
	monitoringHosts, err := CreateNodes(ctx,
		&NodeParams{
			CloudParams:       cp,
			Count:             1,
			Roles:             []SupportedRole{Monitor},
			UseStaticIP:       false,
			SSHPrivateKeyPath: sshPrivateKeyPath,
		})
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully created monitoring node")
	fmt.Println("Linking monitoring node with Avalanche Validator nodes ...")
	// Link the 2 validator nodes previously created with the monitoring host so that
	// the monitoring host can start tracking the validator nodes metrics and collecting their logs
	if err := monitoringHosts[0].MonitorNodes(ctx, hosts, ""); err != nil {
		panic(err)
	}
	fmt.Println("Successfully linked monitoring node with Avalanche Validator nodes")
	wg := sync.WaitGroup{}
	wgResults := NodeResults{}
	for _, host := range hosts {
		wg.Add(1)
		keyPath := fmt.Sprintf("/Users/raymondsukanto/.avalanche-cli/nodes/%s", host.NodeID)
		go func(nodeResults *NodeResults, node Node) {
			defer wg.Done()
			if err := node.ProvideStakingCertAndKey(keyPath); err != nil {
				nodeResults.AddResult(host.NodeID, nil, err)
				return
			}
		}(&wgResults, host)
	}
	wg.Wait()

	fmt.Println("Terminating all created nodes ...")
	//// Destroy all created nodes
	//for _, h := range hosts {
	//	err = h.Destroy(ctx)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//err = monitoringHosts[0].Destroy(ctx)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("All nodes terminated")

	//*****
	//subnetParams := subnet.GetDefaultSubnetEVMGenesis()
	//require := require.New(t)
	//newSubnet, err := subnet.New(&subnetParams)
	//require.NoError(err)
	//network := avalanche.FujiNetwork()
	//keychain, err := keychain.NewKeychain(network, "KEY_PATH")
	//require.NoError(err)
	//controlKeys := keychain.Addresses().List()
	//subnetAuthKeys := keychain.Addresses().List()
	//threshold := 1
	//newSubnet.SetSubnetCreateParams(controlKeys, uint32(threshold))
	//wallet, err := wallet.New(
	//	context.Background(),
	//	&primary.WalletConfig{
	//		URI:              network.Endpoint,
	//		AVAXKeychain:     keychain.Keychain,
	//		EthKeychain:      secp256k1fx.NewKeychain(),
	//		PChainTxsToFetch: nil,
	//	},
	//)
	//require.NoError(err)
	//deploySubnetTx, err := newSubnet.CreateSubnetTx(wallet)
	//require.NoError(err)
	//subnetID, err := newSubnet.Commit(*deploySubnetTx, wallet, true)
	//require.NoError(err)
	//fmt.Printf("subnetID %s \n", subnetID.String())
	//time.Sleep(2 * time.Second)
	//newSubnet.SetBlockchainCreateParams(subnetAuthKeys)
	//deployChainTx, err := newSubnet.CreateBlockchainTx(wallet)
	//require.NoError(err)
	//blockchainID, err := newSubnet.Commit(*deployChainTx, wallet, true)
	//require.NoError(err)
	//fmt.Printf("blockchainID %s \n", blockchainID.String())
	//newSubnet, err := subnet.New(&subnetParams)
	//require.NoError(err)
	//
	//subnetParams := subnet.SubnetParams{
	//	GenesisFilePath: "/Users/raymondsukanto/.avalanche-cli/subnets/sdkSubnetNew/genesis.json",
	//}
	//newSubnet, _ := subnet.New(&subnetParams)
	//network := avalanche.FujiNetwork()
	//keychain, err := keychain.NewKeychain(network, "/Users/raymondsukanto/.avalanche-cli/key/newTestKeyNew.pk")
	//subnetID, err := ids.FromString("2VsqBt64W9qayKttmGTiAmtsQVnp9e9U4gSHF1yuLKHuquck5j")
	//chainID, err := ids.FromString("2RjcKE3zjxpmEtA3uydmfzkZ9i4c1hVLcZizNVsNbwNKVFfAQU")
	//wallet, err := wallet.New(
	//	context.Background(),
	//	&primary.WalletConfig{
	//		URI:              network.Endpoint,
	//		AVAXKeychain:     keychain.Keychain,
	//		EthKeychain:      secp256k1fx.NewKeychain(),
	//		PChainTxsToFetch: set.Of(subnetID, chainID),
	//	},
	//)
	//validatorList := []subnet.ValidatorParams{}
	//for _, h := range hosts {
	//	nodeID, _ := ids.NodeIDFromString(h.NodeID)
	//	validator := subnet.ValidatorParams{
	//		NodeID:   nodeID,
	//		Duration: 48 * time.Hour,
	//		Weight:   20,
	//	}
	//	validatorList = append(validatorList, validator)
	//}
	//for _, validator := range validatorList {
	//	newSubnet.AddValidator(wallet, validator)
	//}
	///////
	//
	//subnetIDsToValidate := []string{"7kf38G6yt4ZZzKUUU1mvtkrmma52kQhRWxC6H2TL2YBaifFAS"}
	//for _, h := range hosts {
	//	fmt.Println("Reconfiguring node %s to track subnet %s", h.NodeID, subnetIDsToValidate)
	//	if err := h.ValidateSubnets(subnetIDsToValidate); err != nil {
	//		panic(err)
	//	}
	//}
}

//
//func getDefaultSubnetEVMGenesis() subnet.SubnetParams {
//	allocation := core.GenesisAlloc{}
//	defaultAmount, _ := new(big.Int).SetString(vm.DefaultEvmAirdropAmount, 10)
//	allocation[common.HexToAddress("INITIAL_ALLOCATION_ADDRESS")] = core.GenesisAccount{
//		Balance: defaultAmount,
//	}
//	return subnet.SubnetParams{
//		SubnetEVM: &subnet.SubnetEVMParams{
//			ChainID:     big.NewInt(123456),
//			FeeConfig:   vm.StarterFeeConfig,
//			Allocation:  allocation,
//			Precompiles: params.Precompiles{},
//		},
//		Name: "TestSubnet",
//	}
//}
