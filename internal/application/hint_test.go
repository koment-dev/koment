package application

import (
	"strings"
	"testing"
)

func TestNearMissHintNamesTheInvisibleDifference(t *testing.T) {
	source := []byte("silences:\n  - name: flux-oci\n    matchers: []\n")

	hint := nearMissHint(source, "- name:  flux-oci")
	if !strings.Contains(hint, "whitespace is ignored") {
		t.Errorf("a whitespace-only difference was not explained: %q", hint)
	}
	if strings.Contains(hint, "CRLF") {
		t.Errorf("a file with Unix endings should not be blamed on CRLF: %q", hint)
	}
}

func TestNearMissHintBlamesCRLFWhenTheFileHasIt(t *testing.T) {
	source := []byte("silences:\r\n  - name: flux-oci\r\n")

	hint := nearMissHint(source, "- name:  flux-oci")
	if !strings.Contains(hint, "CRLF") {
		t.Errorf("CRLF endings were not named: %q", hint)
	}
}

func TestNearMissHintStaysSilentWhenTheTextIsAbsent(t *testing.T) {
	source := []byte("silences:\n  - name: flux-oci\n")

	if hint := nearMissHint(source, "nothing like this exists"); hint != "" {
		t.Errorf("a genuinely missing excerpt was given a misleading hint: %q", hint)
	}
}
