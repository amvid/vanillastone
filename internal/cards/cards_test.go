package cards

import "testing"

// TestDefaultDeckIsLegal: the auto-generated fallback deck must pass the deck
// rules — otherwise queuing without a saved deck would build an illegal game.
func TestDefaultDeckIsLegal(t *testing.T) {
	if err := ValidateDeck(DefaultDeck(), ClassMage); err != nil {
		t.Fatalf("DefaultDeck must be legal: %v", err)
	}
}

// TestValidateDeck enforces size, copy cap, and pool membership.
func TestValidateDeck(t *testing.T) {
	legal := DefaultDeck()

	if err := ValidateDeck(legal[:DeckSize-1], ClassMage); err == nil {
		t.Fatal("a deck under the size limit should be rejected")
	}
	// Too many copies: fill with a single card.
	tooMany := make([]string, DeckSize)
	for i := range tooMany {
		tooMany[i] = "mote"
	}
	if err := ValidateDeck(tooMany, ClassMage); err == nil {
		t.Fatal("exceeding the copy cap should be rejected")
	}
	// Unknown / non-pool card (a token).
	withToken := append([]string{"broken_golem"}, legal[1:]...)
	if err := ValidateDeck(withToken, ClassMage); err == nil {
		t.Fatal("a token must not be allowed in a deck")
	}
	// The hero power is not a deck card either.
	withPower := append([]string{"fire_dart"}, legal[1:]...)
	if err := ValidateDeck(withPower, ClassMage); err == nil {
		t.Fatal("a hero power must not be allowed in a deck")
	}
}

// TestValidateDeckLegendaryCap: a legendary may appear at most once in a deck
// (the HS rule), while non-legendaries allow MaxCopies. This encodes the
// collection constraint players rely on when building, not just a count.
func TestValidateDeckLegendaryCap(t *testing.T) {
	var legID string
	for _, id := range DeckPoolIDs() {
		if c, ok := set[id]; ok && c.Rarity == RarityLegendary {
			legID = id
			break
		}
	}
	if legID == "" {
		t.Fatal("expected a legendary in the pool")
	}

	// DefaultDeck is legal and already caps legendaries at one copy.
	deck := DefaultDeck()
	if err := ValidateDeck(deck, ClassMage); err != nil {
		t.Fatalf("default deck (one copy of each legendary) should be legal: %v", err)
	}

	// Forcing a second copy of the legendary makes the deck illegal.
	dup := append([]string{}, deck...)
	dup[0], dup[1] = legID, legID
	if err := ValidateDeck(dup, ClassMage); err == nil {
		t.Fatal("two copies of a legendary should be rejected")
	}
}

// TestValidateDeckClass: a deck binds to one playable class. A non-playable
// class is rejected outright, and a class card from another class may not sit in
// a deck of a different class — this is what makes "select a class first, then
// only that class + neutral" enforceable server-side, not just a client filter.
func TestValidateDeckClass(t *testing.T) {
	legal := DefaultDeck()

	// A class that is not in PlayableClasses (no hero / no cards) is rejected.
	if err := ValidateDeck(legal, Class("warlock")); err == nil {
		t.Fatal("a non-playable deck class should be rejected")
	}
	if err := ValidateDeck(legal, ClassNeutral); err == nil {
		t.Fatal("neutral is not a deck class")
	}

	// Confirm a Mage card is legal in a Mage deck. Build a deterministic, clearly
	// legal 30-card deck — two copies of a non-legendary Mage card + distinct
	// neutrals (two each) — so we isolate the class check, not the copy cap.
	// (Iterating DeckPoolIDs, not the map, keeps this order-independent.)
	var mageID string
	for _, id := range DeckPoolIDs() {
		if c := set[id]; c.Class == ClassMage && c.Rarity != RarityLegendary {
			mageID = id
			break
		}
	}
	if mageID == "" {
		t.Fatal("expected at least one non-legendary collectible Mage card")
	}
	withMage := []string{mageID, mageID}
	for _, id := range DeckPoolIDs() {
		if len(withMage) >= DeckSize {
			break
		}
		if c := set[id]; c.Class == ClassNeutral && c.Rarity != RarityLegendary {
			withMage = append(withMage, id, id)
		}
	}
	withMage = withMage[:DeckSize]
	if err := ValidateDeck(withMage, ClassMage); err != nil {
		t.Fatalf("a Mage card must be legal in a Mage deck: %v", err)
	}
	// Cross-class rejection: a Mage card may not sit in a Hunter deck. Swap the two
	// Mage copies into an otherwise-legal Hunter deck.
	crossClass := append([]string{mageID, mageID}, DefaultDeckFor(ClassHunter)[2:]...)
	if err := ValidateDeck(crossClass, ClassHunter); err == nil {
		t.Fatal("a Mage card must be rejected in a Hunter deck")
	}
}

// TestDefaultHunterDeckIsLegal: the curated Hunter fallback deck must be a legal
// Hunter deck (so vs-AI / fallback queuing with Hunter never builds an illegal
// game).
func TestDefaultHunterDeckIsLegal(t *testing.T) {
	if err := ValidateDeck(DefaultDeckFor(ClassHunter), ClassHunter); err != nil {
		t.Fatalf("default Hunter deck must be legal: %v", err)
	}
}

// TestHeroPowerForClass: each playable class resolves to its own hero power card;
// an unknown class falls back to the Mage hero power (never empty).
func TestHeroPowerForClass(t *testing.T) {
	if hp := HeroPowerForClass(ClassMage); hp.ID != "fire_dart" {
		t.Fatalf("Mage hero power want fire_dart, got %q", hp.ID)
	}
	if hp := HeroPowerForClass(ClassHunter); hp.ID != "quick_shot" {
		t.Fatalf("Hunter hero power want quick_shot, got %q", hp.ID)
	}
	if hp := HeroPowerForClass(ClassNeutral); hp.ID != "fire_dart" {
		t.Fatalf("unknown class must fall back to Mage hero power, got %q", hp.ID)
	}
}

// TestDeckClass: a deck's class is inferred from its first non-neutral card; an
// all-neutral deck falls back to the first playable class.
func TestDeckClass(t *testing.T) {
	if got := DeckClass(Deck(defaultHunterDeck)); got != ClassHunter {
		t.Fatalf("Hunter deck should infer Hunter, got %q", got)
	}
	if got := DeckClass(Deck(defaultMageDeck)); got != ClassMage {
		t.Fatalf("Mage deck should infer Mage, got %q", got)
	}
	if got := DeckClass(Deck([]string{"mirefang_raptor", "river_snapper"})); got != PlayableClasses()[0] {
		t.Fatalf("all-neutral deck should fall back to %q, got %q", PlayableClasses()[0], got)
	}
}

// TestClassicMechanicsHaveCards: every CLASSIC-era mechanic the engine supports
// must have at least one collectible card. The pool is scoped to HS-Classic, so
// post-Classic keywords (Rush, Lifesteal, Seek) are deliberately card-less —
// the engine still supports them; cards arrive when the set scope expands. If you
// add a Classic mechanic, add a card + extend this test.
func TestClassicMechanicsHaveCards(t *testing.T) {
	hasKeyword := func(k Keyword) bool {
		for _, c := range set {
			if inPool(c) && c.Has(k) {
				return true
			}
		}
		return false
	}
	for _, k := range []Keyword{
		KeywordTaunt, KeywordCharge, KeywordAegis,
		KeywordTwinstrike, KeywordStealth, KeywordPoisonous,
	} {
		if !hasKeyword(k) {
			t.Errorf("no collectible card has Classic keyword %q", k)
		}
	}

	// Spell Damage is the one non-keyword Classic mechanic with a card at this build
	// stage. Aura (tribe-auras), weapons, destroy, transform, seek etc. are
	// engine features whose Classic cards are still gated / out-of-class, so they
	// are intentionally card-less for now (see TASKS.md) and not asserted here.
	hasSpellDmg := false
	hasEnrage := false
	hasCostMod := false
	for _, c := range set {
		if !inPool(c) {
			continue
		}
		if c.SpellDamage > 0 {
			hasSpellDmg = true
		}
		if c.Enrage != nil {
			hasEnrage = true
		}
		if c.CostAura != nil || c.CostRule != nil {
			hasCostMod = true
		}
	}
	if !hasSpellDmg {
		t.Errorf("no collectible card exercises Classic mechanic %q", "spell-damage")
	}
	if !hasEnrage {
		t.Errorf("no collectible card exercises Classic mechanic %q", "enrage")
	}
	if !hasCostMod {
		t.Errorf("no collectible card exercises Classic mechanic %q", "cost-modification")
	}
}
