import type { CardView, PlayerView } from '../protocol'
import { cardColorClass } from './format'
import { CardFace } from './CardFace'

// Hero is a player's portrait panel: secret gems, health/armor/weapon orbs, the
// nameplate, and the draw pile. Mana lives in the corner mana bar; the hand
// counts are shown visually (own hand / opponent's card backs) elsewhere.
export function Hero(props: {
  side: 'self' | 'opp'
  p: PlayerView
  targetable: boolean
  selected: boolean
  ready: boolean
  // Displayed hero health override, so a damage number lags until its hit lands.
  hpOverride?: number
  onClick: () => void
}) {
  const { side, p, targetable, selected, ready, hpOverride, onClick } = props
  const cls = [
    'heropanel',
    side,
    targetable ? 'targetable' : '',
    selected ? 'selected' : '',
    ready ? 'ready' : '',
    p.frozen ? 'frozen' : '',
    p.immune ? 'immune' : '',
  ]
    .filter(Boolean)
    .join(' ')

  // Hero portrait art + glyph are class-driven, inferred from the hero power's
  // class (the snapshot has no separate hero-class field). Defaults to Mage. A
  // hero-replacement (heroArt set, e.g. Overlord Xathul) overrides the portrait art.
  const heroClass = p.heroPower?.class ?? 'mage'
  const heroIcon =
    heroClass === 'hunter' ? '🏹' : heroClass === 'warrior' ? '⚔️' : heroClass === 'warlock' ? '🔮' : '🧙'
  const heroArt = p.heroArt ? `/art/${p.heroArt}.png` : `/art/${heroClass}_hero.png`

  // Secret gems shown over the portrait, HS-style. Own secrets reveal their name
  // on hover; the opponent's are anonymous "?" tokens.
  const gems: { key: string | number; card?: CardView }[] =
    side === 'self'
      ? (p.secrets ?? []).map((s) => ({ key: s.cardId, card: s }))
      : Array.from({ length: p.secretCount ?? 0 }, (_, i) => ({ key: i }))

  return (
    <div className={cls}>
      <div className="portrait" data-cid={side === 'self' ? 'selfHero' : 'oppHero'} onClick={onClick}>
        {gems.length > 0 && (
          <div className="secret-gems">
            {gems.map((g) => (
              <span key={g.key} className="secret-gem">
                ?
                {g.card && (
                  <span className={'card minion-preview secret-preview' + cardColorClass(g.card)} aria-hidden="true">
                    <CardFace card={g.card} />
                  </span>
                )}
              </span>
            ))}
          </div>
        )}
        <span className="hero-portrait-art" style={{ backgroundImage: `url('${heroArt}')` }} aria-hidden="true" />
        <span className="hero-portrait-icon">{heroIcon}</span>
        {p.frozen && (
          <div className="hero-frost-badge" title="Frozen — can't attack this turn" aria-hidden="true">
            ❄
          </div>
        )}
        {p.immune && (
          <div className="immune-badge" title="Immune — ignores all damage this turn">
            🛡️
          </div>
        )}
        <div className="hp-orb">{hpOverride ?? p.heroHP}</div>
        {!!p.armor && <div className="armor-orb">{p.armor}</div>}
        {p.weapon && (
          <div className="weapon-orb" title={`${p.weapon.name} — ${p.weapon.text ?? ''}`}>
            <span className="weapon-icon">🗡️</span>
            <span className="weapon-stats">
              {p.weapon.attack}/{p.weapon.durability}
            </span>
          </div>
        )}
      </div>
    </div>
  )
}
