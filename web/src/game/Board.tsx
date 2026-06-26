import { useState } from 'react'
import type { CardView, MinionView } from '../protocol'
import type { CharKind } from './types'
import { cardArtIcon, cardColorClass } from './format'
import { CardFace } from './CardFace'

// MinionArt paints the board minion's portrait from /art/<cardId>.png (same source
// as the hand card's CardArt), falling back to the placeholder type glyph until art
// exists. The square art reserves its top ~35% for empty sky (the framing rule); the
// portrait box is only slightly tall, so a plain cover would show that whole sky band.
// Zoom in (height 160% of the box) anchored to the BOTTOM so the crop eats the top sky
// and frames the creature's face/body instead.
function MinionArt({ cardId }: { cardId: string }) {
  const [failed, setFailed] = useState(false)
  if (failed) return <div className="m-portrait">{cardArtIcon('minion')}</div>
  const src = `/art/${cardId}.png`
  return (
    <div
      className="m-portrait"
      style={{ backgroundImage: `url(${src})`, backgroundSize: 'auto 160%', backgroundPosition: 'center bottom' }}
    >
      <img src={src} alt="" style={{ display: 'none' }} onError={() => setFailed(true)} />
    </div>
  )
}

// previewCard builds a CardView from a board minion so the hover preview shows the
// exact card (cost/art/text/type), like a hand card. Attack/health use the current
// in-play values; baseCost = cost so the cost gem doesn't get the cheaper/pricier
// recolour (that's a hand-only signal).
function previewCard(m: MinionView): CardView {
  return {
    cardId: m.cardId,
    name: m.name,
    cardType: 'minion',
    class: m.class,
    rarity: m.rarity,
    cost: m.cost,
    baseCost: m.cost,
    attack: m.attack,
    health: m.health,
    tribe: m.tribe,
    text: m.text,
  }
}

// statClass colors a stat relative to its printed base: green when buffed above
// base, red when below. '' (default) when unchanged.
function statClass(cur: number, base: number): string {
  if (cur > base) return ' stat-up'
  if (cur < base) return ' stat-down'
  return ''
}

// Board renders one player's row of minions. The enemy side is the opponent's;
// the friendly side shows attack-ready/selected state.
export function Board(props: {
  minions: MinionView[]
  enemy?: boolean
  myTurn?: boolean
  attacker: string | null
  // When dragging a minion onto this board, the slot index to open up (a gap the
  // existing minions shift around to make room for). null = no drag in progress.
  dropIndex?: number | null
  // Per-minion displayed health override (by instanceId), so a damage number can
  // lag until its hit animation connects instead of dropping the instant the
  // snapshot arrives. Falls back to the live snapshot health.
  held?: Record<string, number>
  targetable: (kind: CharKind, m?: MinionView) => boolean
  onChar: (targetId: string, kind: CharKind, m?: MinionView) => void
}) {
  const { minions, enemy, myTurn, attacker, dropIndex, held, targetable, onChar } = props
  const kind: CharKind = enemy ? 'enemyMinion' : 'friendlyMinion'
  const nodes = minions.map((m) => {
        const hp = held?.[m.instanceId] ?? m.health
        const cls = ['minion']
        if (enemy) cls.push('enemy') // top row → hover tooltip opens downward
        if (m.rarity) cls.push(`rarity-${m.rarity}`) // rarity-coloured frame + glow
        if (m.taunt) cls.push('taunt')
        if (m.frozen) cls.push('frozen')
        if (m.aegis) cls.push('shielded')
        if (m.immune) cls.push('immune')
        if (m.stealth) cls.push('stealthed')
        if (m.silenced) cls.push('silenced')
        if (targetable(kind, m)) cls.push('targetable')
        if (!enemy && myTurn && m.canAttack && m.attack > 0) cls.push('ready')
        if (!enemy && attacker === m.instanceId) cls.push('selected')
        // Taunt, Aegis and Frozen have dedicated on-object visuals (a stone
        // frame, a golden aura, an icy sheen) rather than a small badge — so they're
        // left out of the badge row below to keep it uncluttered.
        const badges: { icon: string; label: string; desc: string }[] = []
        if (m.twinstrike)
          badges.push({ icon: '🌀', label: 'Twinstrike', desc: 'Can attack twice each turn.' })
        if (m.stealth)
          badges.push({ icon: '🌫️', label: 'Stealth', desc: "Can't be targeted by the enemy until it attacks." })
        if (m.poisonous)
          badges.push({ icon: '🧪', label: 'Poisonous', desc: 'Any minion it damages is destroyed.' })
        if (m.lifesteal)
          badges.push({ icon: '🩸', label: 'Lifesteal', desc: 'Damage it deals also heals your hero.' })
        if (m.finalGasp)
          badges.push({ icon: '💀', label: 'Final Gasp', desc: 'Triggers an effect when this minion dies.' })
        if (m.spellDamage)
          badges.push({
            icon: `🔮+${m.spellDamage}`,
            label: `Spell Damage +${m.spellDamage}`,
            desc: `Your spells deal ${m.spellDamage} extra damage.`,
          })
        if (m.hasEnrage)
          badges.push({
            icon: '💢',
            label: m.enraged ? 'Enraged' : 'Enrage',
            desc: 'Has bonus Attack while damaged.',
          })
        if (m.elusive)
          badges.push({ icon: '🌀', label: 'Elusive', desc: "Can't be targeted by spells or Hero Powers." })
        if (m.cantAttack)
          badges.push({ icon: '🚫', label: "Can't Attack", desc: 'This minion cannot attack.' })
        if (m.silenced)
          badges.push({ icon: '🔇', label: 'Silenced', desc: 'Has lost all abilities and enchantments.' })
        return (
          <div
            key={m.instanceId}
            data-cid={m.instanceId}
            className={cls.join(' ')}
            onClick={() => onChar(m.instanceId, kind, m)}
          >
            {/* Dominant art from /art/<cardId>.png (placeholder glyph until it lands). */}
            <MinionArt key={m.cardId} cardId={m.cardId} />
            {/* Taunt: a stone guard frame. Aegis: a translucent golden aura
               on top. Both are overlays so the art stays visible. */}
            {m.taunt && <div className="taunt-frame" aria-hidden="true" />}
            {m.aegis && <div className="ds-aura" aria-hidden="true" />}
            {m.frozen && (
              <div className="frost-badge" title="Frozen — can't attack this turn" aria-hidden="true">
                ❄
              </div>
            )}
            {m.rarity && <div className={`rarity-gem rarity-${m.rarity}`} />}
            <div className={'m-nametag' + (m.rarity ? ` rarity-${m.rarity}` : '')}>{m.name}</div>
            {badges.length > 0 && (
              <div className="m-badges">
                {badges.map((b, i) => (
                  <span key={i} className="kw">
                    {b.icon}
                    <span className="kw-tip">
                      <strong>{b.label}</strong> — {b.desc}
                    </span>
                  </span>
                ))}
              </div>
            )}
            <div className={'m-atk' + statClass(m.attack, m.baseAttack)}>{m.attack}</div>
            <div className={'m-hp' + (hp < m.maxHealth ? ' stat-down' : statClass(m.maxHealth, m.baseHealth))}>
              {hp}
            </div>
            {/* Full card on hover (the board object itself is art-only). */}
            <div
              className={'card minion-preview' + cardColorClass(previewCard(m))}
              aria-hidden="true"
            >
              <CardFace card={previewCard(m)} />
            </div>
          </div>
        )
      })
  if (dropIndex != null) {
    const at = Math.max(0, Math.min(dropIndex, nodes.length))
    nodes.splice(at, 0, <div key="drop-slot" className="drop-slot" />)
  }
  return <div className="board">{nodes}</div>
}
