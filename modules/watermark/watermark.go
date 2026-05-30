// Copyright 2026 The 42w.shop Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package watermark

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"time"
)

// WatermarkData represents the data embedded in archive watermarks
type WatermarkData struct {
	Server    string `json:"server"`
	CloneUser string `json:"clone_user,omitempty"`
	CloneTime string `json:"clone_time"`
	CloneIP   string `json:"clone_ip,omitempty"`
	Repo      string `json:"repo"`
	Ref       string `json:"ref"`
}

// WatermarkFilename is the filename of the watermark entry in archives
const WatermarkFilename = ".gitea-watermark"

// GenerateWatermarkData creates a WatermarkData struct
func GenerateWatermarkData(server, userName, ip, repoFullName, ref string) WatermarkData {
	return WatermarkData{
		Server:    server,
		CloneUser: userName,
		CloneTime: time.Now().UTC().Format(time.RFC3339),
		CloneIP:   ip,
		Repo:      repoFullName,
		Ref:       ref,
	}
}

// ToJSON marshals the watermark data to JSON
func (d WatermarkData) ToJSON() ([]byte, error) {
	return json.MarshalIndent(d, "", "  ")
}

// InjectIntoZip injects a .gitea-watermark file into an existing zip archive.
// It reads the original zip from `r`, adds the watermark entry, and writes the result to `w`.
func InjectIntoZip(r io.Reader, w io.Writer, data WatermarkData) error {
	watermarkBytes, err := data.ToJSON()
	if err != nil {
		return err
	}

	// Read the entire original zip
	originalBytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	// Create a new zip that includes all original entries plus the watermark
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// Read the original zip and copy all entries
	reader, err := zip.NewReader(bytes.NewReader(originalBytes), int64(len(originalBytes)))
	if err != nil {
		return err
	}

	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}

		wf, err := zipWriter.Create(f.Name)
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(wf, rc); err != nil {
			rc.Close()
			return err
		}
		rc.Close()
	}

	// Add the watermark file
	wf, err := zipWriter.Create(WatermarkFilename)
	if err != nil {
		return err
	}
	if _, err := wf.Write(watermarkBytes); err != nil {
		return err
	}

	return nil
}
