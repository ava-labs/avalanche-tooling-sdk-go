// Copyright (C) 2022, Ava Labs, Inc. All rights reserved
// See the file LICENSE for licensing terms.
package install

import (
	"fmt"
	"path/filepath"

	"github.com/ava-labs/avalanche-tooling-sdk-go/utils"
)

type ReleaseKind int64

const (
	UndefinedRelease ReleaseKind = iota
	LatestRelease
	LatestPreRelease
	CustomRelease
)

// installs archive from given [url] into [outputDir], unless
// [relativeBinPath] exists in [outputDir] and is executable.
// verifies [relativeBinPath] exists in [outputDir] after installation,
// and returns full path
func InstallBinary(
	url string,
	archiveKind ArchiveKind,
	outputDir string,
	relativeBinPath string,
) (string, error) {
	expectedBinPath := filepath.Join(outputDir, relativeBinPath)
	if utils.IsExecutable(expectedBinPath) {
		return expectedBinPath, nil
	}
	bs, err := utils.HTTPGet(url, "")
	if err != nil {
		return "", err
	}
	if err := ExtractArchive(archiveKind, bs, outputDir); err != nil {
		return "", err
	}
	if !utils.FileExists(expectedBinPath) {
		return "", fmt.Errorf("%s does not exist after installing release", expectedBinPath)
	}
	if !utils.IsExecutable(expectedBinPath) {
		return "", fmt.Errorf("release asset %s is not an executable", expectedBinPath)
	}
	return expectedBinPath, nil
}

// installs archive into a [version] subdir of [baseDir]
// see InstallBinary from installation checks
func InstallBinaryVersion(
	url string,
	archiveKind ArchiveKind,
	baseDir string,
	relativeBinPath string,
	version string,
) (string, error) {
	return InstallBinary(
		url,
		archiveKind,
		filepath.Join(baseDir, version),
		relativeBinPath,
	)
}

func InstallGithubRelease(
	org string,
	repo string,
	authToken string,
	releaseKind ReleaseKind,
	customVersion string,
	getAssetName func(string) (string, error),
	archiveKind ArchiveKind,
	baseDir string,
	relativeBinPath string,
) (string, error) {
	var (
		version string
		err     error
	)
	switch releaseKind {
	case LatestRelease:
		version, err = utils.GetLatestGithubReleaseVersion(org, repo, authToken)
		if err != nil {
			return "", err
		}
	case LatestPreRelease:
		version, err = utils.GetLatestGithubPreReleaseVersion(org, repo, authToken)
		if err != nil {
			return "", err
		}
	case CustomRelease:
		version = customVersion
	default:
		return "", fmt.Errorf("unsupported release kind %d", releaseKind)
	}
	asset, err := getAssetName(version)
	if err != nil {
		return "", err
	}
	url := utils.GetGithubReleaseAssetURL(
		org,
		repo,
		version,
		asset,
	)
	return InstallBinaryVersion(
		url,
		archiveKind,
		baseDir,
		relativeBinPath,
		version,
	)
}
