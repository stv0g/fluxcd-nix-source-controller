// SPDX-FileCopyrightText: 2024 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package nix

type LogPosition struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	File   string `json:"file"`
}

//nolint:tagliatelle
type LogTrace struct {
	LogPosition

	RawMessage string `json:"raw_msg"`
}

//nolint:tagliatelle
type LogEntry struct {
	LogPosition

	ID     int `json:"id"`
	Parent int `json:"parent"`

	Action     string     `json:"action"`
	Level      int        `json:"lvl"`
	Message    int        `json:"msg"`
	RawMessage int        `json:"raw_msg"`
	Type       int        `json:"type"`
	Text       string     `json:"text"`
	Fields     []any      `json:"fields"`
	Trace      []LogTrace `json:"trace"`
}

//nolint:tagliatelle
type BuildResult struct {
	DerivationPath string            `json:"drvPath"`
	Outputs        map[string]string `json:"outputs"`
}
