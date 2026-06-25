# Card art

Drop one image per card here, named by its **card id**:

```
web/public/art/<cardId>.png
```

e.g. `emberwing_matron.png`, `glacial_splinter.png`, `moonsilver_guardian.png`.

Card ids are the engine ids (`internal/cards/*.go` / the gitignored
`.notes/classic-mapping.md`). The client requests `/art/<cardId>.png`; a missing
file falls back to a placeholder type glyph, so art can be added incrementally —
no code change per card. Vite copies `web/public/` into `web/static/` (embedded in
the prod binary) on build, so committed art ships in prod.

⚠️ **Never put real Blizzard/HS card or character names** in prompts you commit,
filenames, or here. Use our ids + generic mechanic/subject descriptions. Real names
live ONLY in `.notes/classic-mapping.md` (gitignored).

---

## Workflow (proven, updated 2026-06-24)

1. **A prompt session** writes a per-card prompt: the fixed STYLE PREFIX (below,
   unchanged for the whole set) + a one-line SUBJECT for that card (derived from the
   card's id / concept). Keeps the set visually consistent.
2. **Generate** the image in an AI generator (ChatGPT/DALL-E worked) and drop the raw
   file in this folder as `<cardId>.png` (any size — it gets downscaled).
3. **Placement pass** (assistant): generate/place the file, resize to ~512px wide,
   compress to <150KB, leave it at `web/public/art/<cardId>.png`, update
   `.notes/art-prompts.md` + `HANDOFF.md`, then stop. To save tokens, **do not
   visually review generated art unless the user asks**; the user reviews and
   requests changes.

### Fixed STYLE PREFIX (reuse verbatim for EVERY card)
Lead with STYLE — the model drifts to semi-realistic concept art unless style is
front-loaded with explicit negatives. (Full canonical copy lives in `.notes/art-prompts.md`.)
```
STYLE (most important — obey before anything else):
Clean cel-shaded cartoon fantasy illustration, classic Saturday-morning cartoon / anime style.
Bold dark outlines. Flat cel shading, only 2-3 tones per color. Bright saturated colors.
Simple clean shapes, smooth flat surfaces, large readable silhouette.
NO photorealism. NO concept art. NO digital painting. NO realistic scale/skin texture.
NO gritty rendering. NO painterly brushwork. NO cinematic realism. NO fine texture detail.

CHARACTER DESIGN (any creature): original animated-film character — exaggerated cartoon
proportions, oversized head, large expressive eyes, broad readable features, simplified
scales/fur, chunky limbs, strong silhouette and personality. NOT a realistic monster.

COMPOSITION & SIZE: square 1:1, render 1024x1024. Whole subject INSIDE the frame —
entire wings and tail visible, DO NOT crop wings, DO NOT crop tail. Camera pulled back
enough to show the complete silhouette. No text, borders, card frame, UI, logos, watermark.

ON-CARD FRAMING (critical — the card slot crops this square top & bottom, anchored BOTTOM):
- The ENTIRE TOP 35% = ONLY empty smoky sky / background. NOTHING of the subject there —
  no head, horns, crown, flames, ears, wings, weapon tips or any detail. A title bar + cost
  badge cover it AND it gets cropped off, so anything in that band is lost.
- The subject's HIGHEST point (top of head / horns / flames / weapon) must sit BELOW the
  35% line. Pin the HEAD/face to the vertical MIDDLE (~50-60% height), never higher.
- The BOTTOM edge MUST be FILLED by the subject's body / feet / tail / ground — it stays on
  screen. Put ALL slack / breathing room in the TOP sky, never at the bottom.

BACKGROUND: vary per card — do NOT default to volcano + lava river. Only go volcanic if the
card itself is volcanic.
```
Tip: attach a prior good cartoon card art as a STYLE reference image in the generator —
the text prompt alone tends to drift toward detailed/painterly.
Then append: `Subject: <one line>.`, an optional `Action:` (flight etc.), and a `Background:` line.

Example that produced a good result (`emberwing_matron`):
> Subject: a colossal matriarch fire-dragon perched on a volcanic crag, vast
> molten-orange wings half-spread, embers and heat-haze rising, glowing cracks along
> obsidian scales, fierce maternal posture, lava glow underlighting, smoke-filled dark
> sky behind.

### Aspect ratio
- The art slot is **full-bleed** (reaches the top + side edges; the cost gem + title plate
  overlay its top) and renders **~square 1:1** (art region = flex 10 vs text flex 4). Painted
  as a `background-size: cover` layer.
- Generate **square 1:1 at 1024×1024**; obey the FRAMING rule (top third = sky only, head at
  ~mid-height) so the overlaid title/cost don't cover the creature.
- The slot is **wider than the square**, so `cover` crops vertically and is anchored **bottom**
  (`background-position: center bottom`) — the crop eats the empty-sky TOP, keeping feet/tail/
  ground. So keep the bottom edge clean (subject sits ON it) and waste all slack in the top sky.

---

## Placement pipeline (per file)

Target: **512px wide, PNG, <150KB** (the whole set is embedded in the binary).

```sh
sips -Z 512 web/public/art/<id>.png            # downscale longest side to 512
python3 -c "from PIL import Image; im=Image.open('web/public/art/<id>.png').convert('RGB'); im.quantize(256).save('web/public/art/<id>.png',optimize=True)"
```

- `sips` ships with macOS. The PIL line palette-quantizes (lossy) to a photo PNG.
- **If still >150KB, drop the palette**: re-run the PIL line with fewer colors —
  `im.quantize(160)` got the cartoon dragon from 173KB → 147KB. Step down (200→160→128)
  until under 150KB; fewer colors = smaller. `pngquant --quality=65-85 --force --ext .png <id>.png`
  is better if installed.
- Palettes <~160 can band on smooth gradients — use the highest color count that still fits.

## Serving / dev reload
- **Dev art is served by VITE (`:5173`), NOT the Go server.** Vite serves `web/public/art`
  directly; the Go server (`:8080`, wgo) only embeds `web/static` and is irrelevant to dev art.
  Restarting the Go server / pointing wgo at the art dir does nothing for what you see — don't.
- A vite plugin (`reloadOnArtChange` in `web/vite.config.ts`) watches `public/art/**` and pushes
  a full browser reload on add/change, so **dropping/overwriting `<id>.png` auto-reloads the
  page — no manual refresh or container restart.** (If you ever edit `vite.config.ts` itself,
  that needs one `docker compose restart web` to take effect.)
- Prod (`:8080`) needs `make build-web` to copy `public/` → embedded `static/`.
