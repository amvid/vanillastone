package cards

import "testing"

// TestAIDecksAreLegal guards every prebuilt AI deck: each must be a legal 30-card
// deck for its class (size, copy caps, legal-class membership). This catches a
// typo'd or wrong-class card id at build time rather than at match start.
func TestAIDecksAreLegal(t *testing.T) {
	for _, class := range PlayableClasses() {
		decks := AIDecks(class)
		if len(decks) == 0 {
			t.Fatalf("playable class %q has no AI decks", class)
		}
		for i, ids := range decks {
			if err := ValidateDeck(ids, class); err != nil {
				t.Errorf("%s AI deck %d is illegal: %v", class, i, err)
			}
		}
	}
}
