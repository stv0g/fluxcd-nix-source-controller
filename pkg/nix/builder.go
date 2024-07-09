// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nix

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
)

func Build(logCallback func(e LogEntry), extraFlags []string, installables []string) ([]BuildResult, error) {
	stdout := &bytes.Buffer{}

	args := []string{
		"--experimental-features", "nix-command flakes",
		"build",
		"--log-format", "internal-json",
		"--no-link",
		"--json",
	}
	args = append(args, extraFlags...)
	args = append(args, installables...)

	cmd := exec.Command("nix", args...)
	cmd.Stdout = stdout

	var done chan any

	if logCallback != nil {
		done = make(chan any)

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create pipe: %w", err)
		}

		go func() {
			logPrefix := []byte("@nix ")

			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				line := scanner.Bytes()

				if !bytes.HasPrefix(line, logPrefix) {
					continue
				}

				var entry LogEntry
				if err := json.Unmarshal(line[len(logPrefix):], &entry); err != nil {
					slog.Warn("Failed to parse log line", slog.String("line", string(line)))
				}

				logCallback(entry)
			}

			done <- struct{}{}
		}()
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start nix: %w", err)
	}

	if logCallback != nil {
		<-done
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("failed to wait for nix: %w", err)
	}

	var results []BuildResult
	if err := json.Unmarshal(stdout.Bytes(), &results); err != nil {
		return nil, fmt.Errorf("failed to parse result: %w", err)
	}

	return results, nil
}
