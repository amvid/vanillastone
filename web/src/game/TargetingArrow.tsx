// TargetingArrow draws an HS-style aiming line from the source character/card
// (located by its data-cid) to the cursor. Read-only DOM lookup; the server is
// still authoritative for whether the eventual target is legal.
export function TargetingArrow({
  sourceId,
  pointer,
}: {
  sourceId: string
  pointer: { x: number; y: number } | null
}) {
  if (!pointer) return null
  const el = document.querySelector<HTMLElement>(`[data-cid="${CSS.escape(sourceId)}"]`)
  if (!el) return null
  const r = el.getBoundingClientRect()
  const sx = r.left + r.width / 2
  const sy = r.top + r.height / 2
  return (
    <svg className="targeting-arrow">
      <defs>
        <marker id="arrowhead" markerWidth="8" markerHeight="8" refX="5" refY="4" orient="auto">
          <path d="M0,0 L8,4 L0,8 Z" fill="#e05555" />
        </marker>
      </defs>
      <line
        x1={sx}
        y1={sy}
        x2={pointer.x}
        y2={pointer.y}
        stroke="#e05555"
        strokeWidth="4"
        strokeLinecap="round"
        markerEnd="url(#arrowhead)"
      />
      <circle cx={sx} cy={sy} r="6" fill="#e05555" />
    </svg>
  )
}
