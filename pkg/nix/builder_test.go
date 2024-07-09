// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nix_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stv0g/flux-nix-controller/pkg/nix"
)

func TestBuilder(t *testing.T) {
	require := require.New(t)

	slog.SetLogLoggerLevel(slog.LevelDebug)

	var logEntries []nix.LogEntry

	cb := func(entry nix.LogEntry) {
		logEntries = append(logEntries, entry)
	}

	results, err := nix.Build(cb,
		[]string{
			// "--rebuild",
			"--no-use-registries",
		},
		[]string{
			"github:NixOS/nixpkgs/nixos-24.05#hello",
			"github:NixOS/nixpkgs/nixos-24.05#cowsay",
		})
	require.NoError(err)
	require.Len(results, 2)

	t.Logf("Logs: %+#v", logEntries)
	t.Logf("Results: %+#v", results)
}
