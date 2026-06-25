import { tribeLabel } from './format'

// CardTooltip is the on-hover info box shown over a card or minion. cost is
// omitted for minions already in play (the view carries no cost there). A minion
// with a tribe shows the tribe in place of "Minion".
export function CardTooltip(props: {
  name: string
  kind: 'minion' | 'spell' | 'secret' | 'weapon' | 'heroPower'
  cost?: number
  stats?: string
  text?: string
  tribe?: string
}) {
  const { name, kind, cost, stats, text, tribe } = props
  const labels: Record<string, string> = {
    minion: 'Minion',
    spell: 'Spell',
    secret: 'Secret',
    weapon: 'Weapon',
    heroPower: 'Hero Power',
  }
  const label = (kind === 'minion' && tribeLabel(tribe)) || labels[kind] || 'Minion'
  return (
    <div className="tooltip">
      <div className="tt-name">{name}</div>
      <div className="tt-meta">
        {label}
        {cost !== undefined ? ` ${cost} mana` : ''}
        {stats ? ` ${stats}` : ''}
      </div>
      {text && <div className="tt-text">{text}</div>}
    </div>
  )
}
