package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amvid/vanillastone/internal/cards"
)

func main() {
	for _, id := range cards.DeckPoolIDs() {
		if _, err := os.Stat(filepath.Join("web/public/art", id+".png")); err == nil {
			continue
		}
		c, _ := cards.Get(id)
		fmt.Printf("%-22s %-8s cost=%d atk=%d hp=%d rarity=%s name=%s text=%s\n",
			id, c.Type, c.Cost, c.Attack, c.Health, c.Rarity, c.Name, strings.ReplaceAll(c.Text, "\n", " "))
	}
}
