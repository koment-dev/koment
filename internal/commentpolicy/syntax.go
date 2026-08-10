package commentpolicy

import (
	"path/filepath"
	"strings"
)

type blockDelimiter struct {
	open  string
	close string
}

type commentSyntax struct {
	line       []string
	block      []blockDelimiter
	directives []string
}

var cFamily = commentSyntax{
	line:  []string{"//"},
	block: []blockDelimiter{{"/*", "*/"}},
}

var hashOnly = commentSyntax{line: []string{"#"}}

func withDirectives(base commentSyntax, directives ...string) commentSyntax {
	base.directives = directives
	return base
}

var syntaxByExtension = map[string]commentSyntax{
	".yaml": withDirectives(hashOnly, "yaml-language-server:", "yamllint", "renovate:", "noqa",
		"zizmor:", "x-release-please", "checkov:", "kubeconform", "helm-docs"),
	".yml": withDirectives(hashOnly, "yaml-language-server:", "yamllint", "renovate:", "noqa",
		"zizmor:", "x-release-please", "checkov:", "kubeconform", "helm-docs"),
	".toml": withDirectives(hashOnly, "schema:", "renovate:"),
	".ini":  hashOnly,
	".cfg":  hashOnly,
	".conf": hashOnly,
	".env":  hashOnly,

	".py":  withDirectives(hashOnly, "type:", "noqa", "pylint:", "mypy:", "ruff:", "fmt:", "pragma:", "-*-"),
	".pyi": withDirectives(hashOnly, "type:", "noqa", "pylint:"),

	".sh":   withDirectives(hashOnly, "shellcheck", "!"),
	".bash": withDirectives(hashOnly, "shellcheck", "!"),
	".zsh":  withDirectives(hashOnly, "shellcheck", "!"),
	".fish": hashOnly,

	".rb":  withDirectives(hashOnly, "frozen_string_literal:", "rubocop:", "encoding:"),
	".pl":  hashOnly,
	".r":   hashOnly,
	".tf":  hashOnly,
	".hcl": hashOnly,

	".js":  withDirectives(cFamily, "eslint", "@ts-", "prettier-", "istanbul ", "jshint", "global ", "jsx"),
	".jsx": withDirectives(cFamily, "eslint", "@ts-", "prettier-"),
	".mjs": withDirectives(cFamily, "eslint", "@ts-", "prettier-"),
	".cjs": withDirectives(cFamily, "eslint", "@ts-", "prettier-"),
	".ts": withDirectives(cFamily,
		"eslint", "@ts-", "prettier-", "tslint:", "/ <reference", "istanbul "),
	".tsx":   withDirectives(cFamily, "eslint", "@ts-", "prettier-", "tslint:"),
	".jsonc": cFamily,

	".rs": withDirectives(cFamily, "!", "/", "rustfmt::", "clippy::", "allow(", "cfg("),

	".c":   withDirectives(cFamily, "NOLINT", "clang-format", "cppcheck-suppress"),
	".h":   withDirectives(cFamily, "NOLINT", "clang-format", "cppcheck-suppress"),
	".cc":  withDirectives(cFamily, "NOLINT", "clang-format", "cppcheck-suppress"),
	".cpp": withDirectives(cFamily, "NOLINT", "clang-format", "cppcheck-suppress"),
	".cxx": withDirectives(cFamily, "NOLINT", "clang-format"),
	".hpp": withDirectives(cFamily, "NOLINT", "clang-format"),
	".hh":  withDirectives(cFamily, "NOLINT", "clang-format"),

	".java":  withDirectives(cFamily, "CHECKSTYLE", "NOPMD", "SUPPRESS"),
	".kt":    withDirectives(cFamily, "ktlint-", "noinspection"),
	".kts":   cFamily,
	".scala": cFamily,
	".swift": withDirectives(cFamily, "swiftlint:", "swift-format-"),
	".cs":    cFamily,
	".php":   withDirectives(cFamily, "phpcs:", "@phpstan-"),
	".dart":  withDirectives(cFamily, "ignore:", "ignore_for_file:"),
	".go2":   cFamily,
	".zig":   cFamily,
	".proto": cFamily,
	".css":   {block: []blockDelimiter{{"/*", "*/"}}},
	".scss":  cFamily,
	".less":  cFamily,

	".lua": withDirectives(commentSyntax{
		line:  []string{"--"},
		block: []blockDelimiter{{"--[[", "]]"}},
	}, "luacheck:", "@"),
	".sql": withDirectives(commentSyntax{
		line:  []string{"--"},
		block: []blockDelimiter{{"/*", "*/"}},
	}, "noqa"),
	".hs":  {line: []string{"--"}, block: []blockDelimiter{{"{-", "-}"}}},
	".ex":  hashOnly,
	".mod": withDirectives(cFamily, "indirect"),
	".vim": {line: []string{`"`}},
}

var fallbackSyntax = commentSyntax{
	line: []string{"#", "//"},
}

var uncommentableExtensions = map[string]bool{
	".md": true, ".markdown": true, ".mdx": true, ".mdc": true,
	".rst": true, ".txt": true, ".adoc": true,
	".json": true, ".lock": true, ".sum": true,
	".csv": true, ".tsv": true,
	".svg": true, ".html": true, ".htm": true, ".xml": true,
}

func syntaxFor(file string) (commentSyntax, bool) {
	extension := strings.ToLower(filepath.Ext(file))
	if uncommentableExtensions[extension] {
		return commentSyntax{}, false
	}
	if known, found := syntaxByExtension[extension]; found {
		return known, true
	}
	if extension == "" && looksLikeScript(file) {
		return withDirectives(hashOnly, "shellcheck", "!", "syntax=", "escape="), true
	}
	return fallbackSyntax, true
}

func looksLikeScript(file string) bool {
	base := strings.ToLower(filepath.Base(file))
	for _, known := range []string{"makefile", "dockerfile", "justfile", "rakefile", "gemfile", "brewfile", "vagrantfile"} {
		if base == known || strings.HasPrefix(base, known+".") {
			return true
		}
	}
	return strings.HasPrefix(base, ".") && !strings.Contains(base[1:], ".")
}
