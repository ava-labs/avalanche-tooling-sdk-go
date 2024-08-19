// Copyright (C) 2022, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.
package install

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ava-labs/avalanche-tooling-sdk-go/constants"
)

type ArchiveKind int64

const (
	UndefinedArchive ArchiveKind = iota
	Zip
	TarGz
)

const maxCopy = 2147483648 // 2 GB

// Sanitize archive file pathing from "G305: Zip Slip vulnerability"
func sanitizeArchivePath(d, t string) (v string, err error) {
	v = filepath.Join(d, t)
	if strings.HasPrefix(v, filepath.Clean(d)) {
		return v, nil
	}

	return "", fmt.Errorf("%s: %s", "content filepath is tainted", t)
}

// ExtractArchive extracts the archive given as bytes slice [archive], into [outputDir]
func ExtractArchive(kind ArchiveKind, archive []byte, outputDir string) error {
	if err := os.MkdirAll(outputDir, constants.DefaultPerms755); err != nil {
		return err
	}
	switch kind {
	case Zip:
		return extractZip(archive, outputDir)
	case TarGz:
		return extractTarGz(archive, outputDir)
	}
	return fmt.Errorf("unsupported archive kind: %d", kind)
}

// extractZip expects a byte stream of a zip file
func extractZip(zipfile []byte, outputDir string) error {
	bytesReader := bytes.NewReader(zipfile)
	zipReader, err := zip.NewReader(bytesReader, int64(len(zipfile)))
	if err != nil {
		return fmt.Errorf("failed creating zip reader from binary stream: %w", err)
	}

	// Closure to address file descriptors issue, uses Close to to not leave open descriptors
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed opening zip file: %w", err)
		}

		// check for zip slip
		path, err := sanitizeArchivePath(outputDir, f.Name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return fmt.Errorf("failed creating directory from zip entry: %w", err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), f.Mode()); err != nil {
				return fmt.Errorf("failed creating file from zip entry: %w", err)
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed opening file from zip entry: %w", err)
			}

			_, err = io.CopyN(f, rc, maxCopy)
			if err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("failed writing zip file entry to disk: %w", err)
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
		if err := rc.Close(); err != nil {
			return err
		}
		return nil
	}

	for _, f := range zipReader.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractTarGz expects a byte array in targz format
func extractTarGz(targz []byte, outputDir string) error {
	byteReader := bytes.NewReader(targz)
	uncompressedStream, err := gzip.NewReader(byteReader)
	if err != nil {
		return fmt.Errorf("failed creating gzip reader from avalanchego binary stream: %w", err)
	}

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		switch {
		// if no more files are found return
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return fmt.Errorf("failed reading next tar entry: %w", err)
		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		// check for zip slip
		target, err := sanitizeArchivePath(outputDir, header.Name)
		if err != nil {
			return err
		}

		// check the file type
		switch header.Typeflag {
		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, constants.DefaultPerms755); err != nil {
					return fmt.Errorf("failed creating directory from tar entry %w", err)
				}
			}
		// if it's a file create it
		case tar.TypeReg:
			// if the containing directory doesn't exist yet, create it
			containingDir := filepath.Dir(target)
			if err := os.MkdirAll(containingDir, constants.DefaultPerms755); err != nil {
				return fmt.Errorf("failed creating directory from tar entry %w", err)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed opening new file from tar entry %w", err)
			}
			// copy over contents
			if _, err := io.CopyN(f, tarReader, maxCopy); err != nil && !errors.Is(err, io.EOF) {
				return fmt.Errorf("failed writing tar entry contents to disk: %w", err)
			}
			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			if err := f.Close(); err != nil {
				return err
			}
		}
	}
}
