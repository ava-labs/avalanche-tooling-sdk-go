// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package constants

import "time"

const (
	// clouds
	UbuntuVersionLTS       = "20.04"
	CloudOperationTimeout  = 2 * time.Minute
	CloudServerStorageSize = 1000

	AWSCloudServerRunningState = "running"
	AWSDefaultInstanceType     = "c5.2xlarge"

	GCPDefaultImageProvider = "avalabs-experimental"
	GCPDefaultInstanceType  = "e2-standard-8"
	GCPImageFilter          = "family=avalanchecli-ubuntu-2204 AND architecture=x86_64"
	GCPEnvVar               = "GOOGLE_APPLICATION_CREDENTIALS"
	GCPDefaultAuthKeyPath   = "~/.config/gcloud/application_default_credentials.json"
	GCPStaticIPPrefix       = "static-ip"
	GCPErrReleasingStaticIP = "failed to release gcp static ip"

	// ports
	SSHTCPPort                    = 22
	AvalanchegoAPIPort            = 9650
	AvalanchegoP2PPort            = 9651
	AvalanchegoGrafanaPort        = 3000
	AvalanchegoLokiPort           = 23101
	AvalanchegoMonitoringPort     = 9090
	AvalanchegoMachineMetricsPort = 9100

	// ssh
	SSHSleepBetweenChecks       = 1 * time.Second
	SSHLongRunningScriptTimeout = 10 * time.Minute
	SSHFileOpsTimeout           = 100 * time.Second
	SSHPOSTTimeout              = 10 * time.Second
	AnsibleSSHUser              = "ubuntu"

	// misc
	IPAddressSuffix = "/32"

	// avago
	LocalAPIEndpoint = "http://127.0.0.1:9650"
)
