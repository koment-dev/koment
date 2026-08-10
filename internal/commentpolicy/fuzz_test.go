package commentpolicy

import (
	"testing"

	"github.com/koment-dev/koment/internal/policy"
)

var fuzzSeeds = []string{
	"# rationale\nretries: 5\n",
	"retries: 5  # trailing\n",
	"url: \"https://example.com/a//b\"\n",
	"name: 'a # not a comment'\n",
	"/* block\n   spanning */\nfn main() {}\n",
	"-- lua\nlocal x = 1\n",
	"#!/usr/bin/env bash\n# type: ignore\n# real rationale\n",
	"--[[ unterminated block\n",
	"/*",
	"#",
	"\"\"\"\n",
	"a\\\"# still in a string\n",
	"\r\n# crlf\r\nkey: value\r\n",
	"",
}

func fuzzSyntaxes() map[string]commentSyntax {
	return map[string]commentSyntax{
		"app.yaml": syntaxByExtension[".yaml"],
		"main.rs":  syntaxByExtension[".rs"],
		"mod.lua":  syntaxByExtension[".lua"],
		"unknown":  fallbackSyntax,
	}
}

// FuzzScanSpansStaysInsideTheContent is the invariant every caller depends on:
// a span is used to slice the file, so an offset outside it panics in
// production rather than in a test.
func FuzzScanSpansStaysInsideTheContent(f *testing.F) {
	for _, seed := range fuzzSeeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, body string) {
		content := []byte(body)
		for name, syntax := range fuzzSyntaxes() {
			previousEnd := -1
			for _, found := range scanSpans(content, syntax) {
				if found.start < 0 || found.end < found.start || found.end > len(content) {
					t.Fatalf("%s: span [%d,%d) is outside content of %d bytes",
						name, found.start, found.end, len(content))
				}
				if found.line < 1 {
					t.Fatalf("%s: span reports line %d", name, found.line)
				}
				if found.start < previousEnd {
					t.Fatalf("%s: spans overlap; [%d,%d) starts before the previous ended at %d",
						name, found.start, found.end, previousEnd)
				}
				previousEnd = found.end
				_ = string(content[found.start:found.end])
			}
		}
	})
}

func FuzzScanNeverPanicsOnAnyFile(f *testing.F) {
	for _, seed := range fuzzSeeds {
		f.Add("app.yaml", seed)
	}
	f.Add("README.md", "# heading\n")
	f.Add("weird.name.zzz", "// what\n")
	f.Fuzz(func(t *testing.T, name, body string) {
		comments, intrinsic, err := markerDetector{}.Scan(name, []byte(body), policy.Default())
		if err != nil {
			return
		}
		if len(comments) != len(intrinsic) {
			t.Fatalf("%d comments but %d intrinsic flags, so the caller reads past the end",
				len(comments), len(intrinsic))
		}
		for _, comment := range comments {
			if comment.Start < 0 || comment.End > len(body) || comment.End < comment.Start {
				t.Fatalf("comment offsets [%d,%d) escape a body of %d bytes",
					comment.Start, comment.End, len(body))
			}
			if comment.Raw != body[comment.Start:comment.End] {
				t.Fatalf("Raw does not match the bytes it claims to span")
			}
		}
	})
}
