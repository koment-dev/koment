package commentpolicy

import (
	"path/filepath"
	"sort"
	"strings"
)

// BlockDelimiter is one pair of markers that open and close a block comment.
type BlockDelimiter struct {
	Open  string
	Close string
}

// FiletypeSupport describes how the marker scan finds comments in one
// filetype. It is the shape the generated language reference prints, so a
// filetype koment gains cannot appear in the code without appearing in the
// manual.
type FiletypeSupport struct {
	Extension  string
	Line       []string
	Block      []BlockDelimiter
	Directives []string
}

// DetectedFiletypes reports every extension the syntax table names, sorted by
// extension.
func DetectedFiletypes() []FiletypeSupport {
	supported := make([]FiletypeSupport, 0, len(syntaxByExtension))
	for extension, syntax := range syntaxByExtension {
		supported = append(supported, describe(extension, syntax))
	}
	sort.Slice(supported, func(earlier, later int) bool {
		return supported[earlier].Extension < supported[later].Extension
	})
	return supported
}

func describe(extension string, syntax commentSyntax) FiletypeSupport {
	described := FiletypeSupport{
		Extension:  extension,
		Line:       append([]string(nil), syntax.line...),
		Directives: append([]string(nil), syntax.directives...),
	}
	for _, delimiter := range syntax.block {
		described.Block = append(described.Block, BlockDelimiter{Open: delimiter.open, Close: delimiter.close})
	}
	return described
}

// FallbackMarkers reports the line markers koment assumes for an extension the
// syntax table does not name.
func FallbackMarkers() []string {
	return append([]string(nil), fallbackSyntax.line...)
}

// UndetectedExtensions reports the prose and data formats koment never scans,
// sorted.
func UndetectedExtensions() []string {
	extensions := make([]string, 0, len(uncommentableExtensions))
	for extension := range uncommentableExtensions {
		extensions = append(extensions, extension)
	}
	sort.Strings(extensions)
	return extensions
}

// ScriptFilenames reports the extensionless filenames koment reads as shell,
// sorted.
func ScriptFilenames() []string {
	names := append([]string(nil), scriptFilenames...)
	sort.Strings(names)
	return names
}

// DetectorName reports which detector claims a path, or an empty string when
// no detector does.
func DetectorName(file string) string {
	detector := detectorFor(file)
	if detector == nil {
		return ""
	}
	return detector.Name()
}

const goExtension = ".go"

func hasScriptFilename(file string) bool {
	base := strings.ToLower(filepath.Base(file))
	for _, known := range scriptFilenames {
		if base == known || strings.HasPrefix(base, known+".") {
			return true
		}
	}
	return false
}
