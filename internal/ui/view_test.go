package ui

import (
	"strings"
	"testing"
)

func paragraphsOf(lengths ...int) []string {
	paragraphs := make([]string, 0, len(lengths))
	for _, length := range lengths {
		paragraphs = append(paragraphs, strings.Repeat("x", length))
	}
	return paragraphs
}

func TestAShortBodyIsNeverFolded(t *testing.T) {
	body := paragraphsOf(200, 150, 100)
	shown, folded := splitBody(body)
	if folded != nil {
		t.Fatalf("a %d character body was folded; the budget is %d", lengthOf(body), visibleBodyBudget)
	}
	if len(shown) != len(body) {
		t.Fatalf("shown %d paragraphs, want all %d", len(shown), len(body))
	}
}

func TestALongBodyFoldsAtAParagraphBoundary(t *testing.T) {
	shown, folded := splitBody(paragraphsOf(400, 400, 400))
	if len(shown) != 1 || len(folded) != 2 {
		t.Fatalf("split 400/400/400 into %d shown and %d folded, want 1 and 2", len(shown), len(folded))
	}
	if lengthOf(shown) > visibleBodyBudget {
		t.Errorf("the visible lead is %d characters, over the %d budget", lengthOf(shown), visibleBodyBudget)
	}
}

func TestASingleParagraphIsNeverCutInHalf(t *testing.T) {
	body := paragraphsOf(1500)
	shown, folded := splitBody(body)
	if folded != nil {
		t.Fatal("a single paragraph was split, which would cut a sentence mid-word")
	}
	if len(shown) != 1 || len(shown[0]) != 1500 {
		t.Fatalf("the paragraph was altered: got %d characters, want 1500", len(shown[0]))
	}
}

func TestATailTooShortToBeWorthHidingStaysVisible(t *testing.T) {
	shown, folded := splitBody(paragraphsOf(700, 40))
	if folded != nil {
		t.Fatalf("hid a %d character tail; below %d the disclosure costs more than it saves",
			40, shortestWorthHiding)
	}
	if len(shown) != 2 {
		t.Fatalf("shown %d paragraphs, want both", len(shown))
	}
}

func TestFoldingLosesNoText(t *testing.T) {
	for _, body := range [][]string{
		paragraphsOf(400, 400, 400),
		paragraphsOf(200, 150, 100),
		paragraphsOf(1500),
		paragraphsOf(700, 40),
		paragraphsOf(100, 100, 100, 100, 100, 100, 100, 100),
	} {
		shown, folded := splitBody(body)
		if got, want := lengthOf(shown)+lengthOf(folded), lengthOf(body); got != want {
			t.Errorf("split a %d character body into %d characters", want, got)
		}
		if len(shown)+len(folded) != len(body) {
			t.Errorf("split %d paragraphs into %d", len(body), len(shown)+len(folded))
		}
		if len(shown) == 0 {
			t.Error("folded every paragraph, leaving a note with no visible body")
		}
	}
}

func TestTheBudgetMatchesTheRepositoryItWasMeasuredAgainst(t *testing.T) {
	median := paragraphsOf(190, 195)
	if _, folded := splitBody(median); folded != nil {
		t.Error("a median-length annotation folds; the budget was measured to leave it whole")
	}
}

func TestTheBudgetCountsCharactersRatherThanUTF8Bytes(t *testing.T) {
	shown, folded := splitBody([]string{
		strings.Repeat("界", 300),
		strings.Repeat("界", 300),
		strings.Repeat("界", 200),
	})
	if len(shown) != 2 || len(folded) != 1 {
		t.Fatalf("split a 300/300/200 character body into %d shown and %d folded paragraphs", len(shown), len(folded))
	}
}
