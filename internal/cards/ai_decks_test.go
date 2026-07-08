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

// TestStarterDecksNamedAndLegal guards the player-facing prefill decks: every
// playable class must expose starters, each with a non-empty flavor name and a
// legal card list. Catches a new class added to AIDecks without a starterNames
// entry (would surface as an unnamed prefill button).
func TestStarterDecksNamedAndLegal(t *testing.T) {
	for _, class := range PlayableClasses() {
		starters := StarterDecks(class)
		if len(starters) == 0 {
			t.Fatalf("playable class %q has no starter decks", class)
		}
		for i, s := range starters {
			if s.Name == "" {
				t.Errorf("%s starter %d has no name", class, i)
			}
			if err := ValidateDeck(s.Cards, class); err != nil {
				t.Errorf("%s starter %q is illegal: %v", class, s.Name, err)
			}
		}
	}
}
