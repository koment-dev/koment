package store

import "testing"

// FuzzDecodeAnnotationRejectsRatherThanPanics guards the one place koment
// parses bytes it did not write. A record arrives from git, from a pull
// request, or from a remote repository over the API; the decoder must refuse
// anything malformed rather than crash the command reading it.
func FuzzDecodeAnnotationRejectsRatherThanPanics(f *testing.F) {
	f.Add(legacyRecordYAML)
	f.Add("apiVersion: " + APIVersion + "\nkind: Annotation\n")
	f.Add("apiVersion: koment.dev/v9\nkind: Annotation\n")
	f.Add("version: 1\n")
	f.Add("")
	f.Add("---\n---\n")
	f.Add("apiVersion: " + APIVersion + "\nkind: Annotation\nmetadata:\n  id: " + firstID + "\n")

	f.Fuzz(func(t *testing.T, body string) {
		annotation, err := DecodeAnnotation(firstID, []byte(body))
		if err != nil {
			return
		}
		if annotation == nil {
			t.Fatal("decode returned no error and no annotation, so a caller dereferences nil")
		}
		if annotation.Metadata.ID != firstID {
			t.Fatalf("accepted a record whose id is %q, not the filename's %q",
				annotation.Metadata.ID, firstID)
		}
		if err := annotation.Validate(); err != nil {
			t.Fatalf("accepted a record that does not validate: %v", err)
		}
	})
}
