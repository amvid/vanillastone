package match

import (
	"testing"

	"github.com/amvid/vanillastone/internal/cards"
)

// spellDamageText must bump exactly the number the engine boosts (the EffectDamage
// Amount, or FrozenDamage when Amount is 0) and ONLY for damage spells — because
// the green marker is the player's promise of what the spell will actually deal.
// If it ever marks a non-boosted number, the card lies about its output.
func TestSpellDamageText(t *testing.T) {
	cases := []struct {
		name string
		card cards.Card
		sp   int
		want string
	}{
		{
			name: "deal N damage spell bumps the amount",
			card: cards.Card{Type: cards.TypeSpell, Text: "Deal 3 damage.",
				Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 3}},
			sp:   2,
			want: "Deal {sd:5} damage.",
		},
		{
			name: "frozenDamage spell (Amount 0) bumps the frozen figure",
			card: cards.Card{Type: cards.TypeSpell, Text: "Freeze a character. If it was already Frozen, deal 4 damage instead.",
				Effect: &cards.Effect{Kind: cards.EffectDamage, FrozenDamage: 4}},
			sp:   1,
			want: "Freeze a character. If it was already Frozen, deal {sd:5} damage instead.",
		},
		{
			name: "only the FIRST occurrence of the base number is the damage figure",
			card: cards.Card{Type: cards.TypeSpell, Text: "Deal 2 damage. Draw 2 cards.",
				Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 2}},
			sp:   1,
			want: "Deal {sd:3} damage. Draw 2 cards.",
		},
		{
			name: "non-damage spell is left untouched (Spell Damage does not boost it)",
			card: cards.Card{Type: cards.TypeSpell, Text: "Draw 2 cards.",
				Effect: &cards.Effect{Kind: cards.EffectDraw, Amount: 2}},
			sp:   3,
			want: "Draw 2 cards.",
		},
		{
			name: "pure-Freeze damage spell (Amount 0, no FrozenDamage) is untouched",
			card: cards.Card{Type: cards.TypeSpell, Text: "Freeze a minion.",
				Effect: &cards.Effect{Kind: cards.EffectDamage, Freeze: true}},
			sp:   2,
			want: "Freeze a minion.",
		},
		{
			name: "word boundary: base 2 does not match inside 12",
			card: cards.Card{Type: cards.TypeSpell, Text: "Deal 12 damage to a 2-cost minion.",
				Effect: &cards.Effect{Kind: cards.EffectDamage, Amount: 12}},
			sp:   1,
			want: "Deal {sd:13} damage to a 2-cost minion.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := spellDamageText(tc.card, tc.sp); got != tc.want {
				t.Errorf("spellDamageText() = %q, want %q", got, tc.want)
			}
		})
	}
}
