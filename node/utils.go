// Copyright (C) 2024, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package node

// getDefaultProjectNameFromGCPCredentials returns the default GCP project name
func getDefaultProjectNameFromGCPCredentials(credentialsFilePath string) (string, error) {
	type GCPConfig struct {
		ClientID       string `json:"client_id"`
		ClientSecret   string `json:"client_secret"`
		QuotaProjectID string `json:"quota_project_id"`
		RefreshToken   string `json:"refresh_token"`
		Type           string `json:"type"`
	}
	file, err := os.Open(utils.ExpandHome(credentialsFilePath))
	if err != nil {
		return "", err
	}
	defer file.Close()
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	var gcpConfig GCPConfig
	if err := json.Unmarshal(bytes, &gcpConfig); err != nil {
		return "", err
	}
	return GCPConfig.QuotaProjectID, nil
}
