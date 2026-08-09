package commentpolicy

import (
	"bytes"
	"strings"

	"github.com/koment-dev/koment/internal/policy"
)

const binarySniffLimit = 8 << 10

type markerDetector struct{}

func (markerDetector) Handles(file string) bool {
	_, commentable := syntaxFor(file)
	return commentable
}

func (markerDetector) Scan(file string, content []byte, configured policy.Policy) ([]SourceComment, []bool, error) {
	syntax, commentable := syntaxFor(file)
	if !commentable || isBinary(content) {
		return nil, nil, nil
	}

	spans := scanSpans(content, syntax)
	comments := make([]SourceComment, 0, len(spans))
	intrinsic := make([]bool, 0, len(spans))
	for _, span := range spans {
		raw := string(content[span.start:span.end])
		body := markerBody(raw, syntax)
		comments = append(comments, SourceComment{
			File: file, Raw: raw, Body: body, Line: span.line,
			Start: span.start, End: span.end,
		})
		intrinsic = append(intrinsic, isMarkerIntrinsic(raw, body, span, syntax, configured))
	}
	return comments, intrinsic, nil
}

func isBinary(content []byte) bool {
	head := content
	if len(head) > binarySniffLimit {
		head = head[:binarySniffLimit]
	}
	return bytes.IndexByte(head, 0) >= 0
}

type span struct {
	start    int
	end      int
	line     int
	ownsLine bool
	shebang  bool
}

func scanSpans(content []byte, syntax commentSyntax) []span {
	var spans []span
	starts := lineStarts(content)

	for index := 0; index < len(starts); index++ {
		lineFrom := starts[index]
		lineTo := lineEnd(content, starts, index)
		text := string(content[lineFrom:lineTo])

		at, block := commentStart(text, syntax)
		if at < 0 {
			continue
		}

		found := span{
			start:    lineFrom + at,
			line:     index + 1,
			ownsLine: strings.TrimSpace(text[:at]) == "",
		}
		found.shebang = index == 0 && strings.HasPrefix(text, "#!")

		if block.open != "" {
			end, lastLine := blockEnd(content, starts, index, found.start+len(block.open), block.close)
			found.end = end
			spans = append(spans, found)
			index = lastLine
			continue
		}

		found.end = lineTo
		if found.ownsLine && !found.shebang && !isDirectiveLine(text, syntax) {
			index = extendLineGroup(content, starts, index, syntax, &found)
		}
		spans = append(spans, found)
	}
	return spans
}

func extendLineGroup(content []byte, starts []int, index int, syntax commentSyntax, group *span) int {
	last := index
	for next := index + 1; next < len(starts); next++ {
		lineFrom := starts[next]
		lineTo := lineEnd(content, starts, next)
		text := string(content[lineFrom:lineTo])
		at, block := commentStart(text, syntax)
		if at < 0 || block.open != "" || strings.TrimSpace(text[:at]) != "" || isDirectiveLine(text, syntax) {
			break
		}
		group.end = lineTo
		last = next
	}
	return last
}

func isDirectiveLine(text string, syntax commentSyntax) bool {
	return matchesAnyDirective(markerBody(strings.TrimSpace(text), syntax), syntax.directives)
}

func blockEnd(content []byte, starts []int, index, from int, closing string) (int, int) {
	relative := bytes.Index(content[from:], []byte(closing))
	if relative < 0 {
		return len(content), len(starts) - 1
	}
	end := from + relative + len(closing)
	line := index
	for line+1 < len(starts) && starts[line+1] <= end {
		line++
	}
	return end, line
}

func commentStart(text string, syntax commentSyntax) (int, blockDelimiter) {
	best, bestBlock := -1, blockDelimiter{}
	consider := func(at int, block blockDelimiter) {
		if at < 0 || (best >= 0 && at >= best) {
			return
		}
		if !markerIsReal(text, at) {
			return
		}
		best, bestBlock = at, block
	}
	for _, block := range syntax.block {
		consider(strings.Index(text, block.open), block)
	}
	for _, marker := range syntax.line {
		consider(strings.Index(text, marker), blockDelimiter{})
	}
	return best, bestBlock
}

func markerIsReal(text string, at int) bool {
	before := text[:at]
	if strings.TrimSpace(before) != "" && !strings.HasSuffix(before, " ") && !strings.HasSuffix(before, "\t") {
		return false
	}
	return !insideQuotes(before)
}

func insideQuotes(before string) bool {
	single, double, escaped := 0, 0, false
	for _, character := range before {
		switch {
		case escaped:
			escaped = false
		case character == '\\':
			escaped = true
		case character == '\'' && double%2 == 0:
			single++
		case character == '"' && single%2 == 0:
			double++
		}
	}
	return single%2 == 1 || double%2 == 1
}

func markerBody(raw string, syntax commentSyntax) string {
	lines := strings.Split(raw, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		for _, block := range syntax.block {
			line = strings.TrimSpace(strings.TrimPrefix(line, block.open))
			line = strings.TrimSpace(strings.TrimSuffix(line, block.close))
		}
		for _, marker := range syntax.line {
			line = strings.TrimSpace(strings.TrimPrefix(line, marker))
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "*"))
		cleaned = append(cleaned, line)
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func isMarkerIntrinsic(raw, body string, found span, syntax commentSyntax, configured policy.Policy) bool {
	switch {
	case isManagedRegionMarker(body):
		return true
	case found.shebang && configured.Allows(policy.IntrinsicToolchain):
		return true
	case configured.Allows(policy.IntrinsicDeprecated) && strings.Contains(body, "Deprecated:"):
		return true
	case configured.Allows(policy.IntrinsicUpstreamLink) &&
		(strings.Contains(raw, "https://") || strings.Contains(raw, "http://")):
		return true
	case configured.Allows(policy.IntrinsicGeneratedMarker) && isGeneratedMarker(body):
		return true
	case configured.Allows(policy.IntrinsicToolchain) && hasConfiguredDirective(body, syntax):
		return true
	case configured.MatchesAllowedAnnotation(body):
		return true
	default:
		return false
	}
}

func isManagedRegionMarker(body string) bool {
	return strings.HasPrefix(strings.TrimSpace(body), "koment:managed-")
}

func isGeneratedMarker(body string) bool {
	upper := strings.ToUpper(body)
	return strings.Contains(upper, "DO NOT EDIT") ||
		strings.Contains(upper, "@GENERATED") ||
		(strings.Contains(body, "Code generated") && strings.Contains(upper, "DO NOT EDIT"))
}

func hasConfiguredDirective(body string, syntax commentSyntax) bool {
	if body == "" {
		return false
	}
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if !matchesAnyDirective(line, syntax.directives) {
			return false
		}
	}
	return true
}

func matchesAnyDirective(line string, directives []string) bool {
	for _, prefix := range directives {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
