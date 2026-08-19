package commentpolicy

import (
	"slices"
	"testing"
)

func TestEveryTableExtensionReachesTheReference(t *testing.T) {
	reported := map[string]bool{}
	for _, filetype := range DetectedFiletypes() {
		reported[filetype.Extension] = true
	}

	for extension := range syntaxByExtension {
		if !reported[extension] {
			t.Errorf("%s is in the syntax table but never reaches the generated reference", extension)
		}
	}
	if len(reported) != len(syntaxByExtension) {
		t.Errorf("reference lists %d extensions, the table declares %d", len(reported), len(syntaxByExtension))
	}
}

func TestDetectedFiletypesAreSortedSoTheReferenceDiffIsStable(t *testing.T) {
	filetypes := DetectedFiletypes()
	if !slices.IsSortedFunc(filetypes, func(earlier, later FiletypeSupport) int {
		switch {
		case earlier.Extension < later.Extension:
			return -1
		case earlier.Extension > later.Extension:
			return 1
		default:
			return 0
		}
	}) {
		t.Fatal("DetectedFiletypes is unsorted, so regenerating the reference reorders its rows")
	}
}

func TestGoFilesGoToTheParserAndEverythingElseToTheMarkerScan(t *testing.T) {
	for file, want := range map[string]string{
		"internal/commentpolicy/syntax.go": "go/ast",
		"kubernetes/app.yaml":              "marker scan",
		"scripts/deploy.sh":                "marker scan",
		"Makefile":                         "marker scan",
		".gitignore":                       "marker scan",
		"generics.go2":                     "marker scan",
		"README.md":                        "",
		"package-lock.json":                "",
	} {
		if got := DetectorName(file); got != want {
			t.Errorf("DetectorName(%q) = %q, want %q", file, got, want)
		}
	}
}

func TestEveryUndetectedExtensionReallyHasNoDetector(t *testing.T) {
	for _, extension := range UndetectedExtensions() {
		if name := DetectorName("prose" + extension); name != "" {
			t.Errorf("%s is listed as never scanned but %s claims it", extension, name)
		}
	}
}

func TestEveryScriptFilenameIsRecognisedAsAScript(t *testing.T) {
	for _, name := range ScriptFilenames() {
		if !looksLikeScript(name) {
			t.Errorf("%s is published as a script filename but looksLikeScript rejects it", name)
		}
	}
}

func TestAccessorsCopySoACallerCannotEditThePolicy(t *testing.T) {
	FallbackMarkers()[0] = "mutated"
	if fallbackSyntax.line[0] == "mutated" {
		t.Fatal("FallbackMarkers handed out the live slice")
	}

	first := DetectedFiletypes()[0]
	if len(first.Line) > 0 {
		first.Line[0] = "mutated"
	}
	if syntaxByExtension[first.Extension].line[0] == "mutated" {
		t.Fatal("DetectedFiletypes handed out the live slice")
	}
}
