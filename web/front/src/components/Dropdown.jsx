import { useEffect, useLayoutEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

/**
 * Custom dropdown that replaces <select> for proper styling.
 *
 *  • The menu is rendered in a React portal (escapes any parent stacking
 *    context, e.g. backdrop-blur cards below it).
 *  • Position is recomputed on open and on scroll/resize.
 *  • Keyboard navigation: ↑/↓/Enter/Esc/Home/End.
 *  • Closes on outside click / Esc / selection.
 *  • In RTL, the chevron sits on the LEFT (end of inline flow) which is
 *    the natural place for an indicator in a right-aligned text field.
 *  • Accessible: aria-haspopup, aria-expanded, role=listbox.
 *
 * Props:
 *   value, onChange           — controlled value
 *   options                   — [{ value, label }]  OR  [string]
 *   placeholder               — shown when value is empty
 *   label                     — optional floating label
 *   icon                      — optional React node shown next to the label
 *   menuAlign                 — 'start' (default, aligned to trigger's edge
 *                                that faces the start side) | 'end'
 */
export default function Dropdown({
  value,
  onChange,
  options,
  placeholder = 'انتخاب کنید…',
  label,
  icon,
  className = '',
  menuAlign = 'start',
}) {
  const [open, setOpen] = useState(false)
  const [activeIdx, setActiveIdx] = useState(-1)
  const [menuPos, setMenuPos] = useState(null)
  const wrapRef = useRef(null)
  const buttonRef = useRef(null)
  const menuRef = useRef(null)

  const normalized = options.map((o) =>
    typeof o === 'string' ? { value: o, label: o } : o,
  )
  const selected = normalized.find((o) => o.value === value)
  const displayLabel = selected?.label ?? placeholder
  const isPlaceholder = !selected

  // Compute menu position relative to viewport (for fixed positioning).
  useLayoutEffect(() => {
    if (!open || !buttonRef.current) return
    function place() {
      const r = buttonRef.current.getBoundingClientRect()
      // In RTL, "start" means the right edge of the trigger.
      // For 'start' alignment, menu's right edge aligns with trigger's right edge.
      // For 'end'   alignment, menu's left  edge aligns with trigger's left  edge.
      const pos = {
        top: r.bottom + 6,
        width: r.width,
      }
      if (menuAlign === 'end') {
        pos.left = r.left
      } else {
        pos.right = window.innerWidth - r.right
      }
      setMenuPos(pos)
    }
    place()
    window.addEventListener('resize', place)
    window.addEventListener('scroll', place, true)
    return () => {
      window.removeEventListener('resize', place)
      window.removeEventListener('scroll', place, true)
    }
  }, [open, menuAlign])

  // Reset activeIdx and scroll the selected item into view when menu opens.
  useEffect(() => {
    if (open) {
      const idx = normalized.findIndex((o) => o.value === value)
      setActiveIdx(idx >= 0 ? idx : 0)
      requestAnimationFrame(() => {
        const item = menuRef.current?.querySelector('[data-active="true"]')
        item?.scrollIntoView?.({ block: 'nearest' })
      })
    }
  }, [open]) // eslint-disable-line react-hooks/exhaustive-deps

  // Close on outside click + Esc.
  useEffect(() => {
    if (!open) return
    function onDocMouseDown(e) {
      if (
        wrapRef.current?.contains(e.target) ||
        menuRef.current?.contains(e.target)
      ) {
        return
      }
      setOpen(false)
    }
    function onKey(e) {
      if (e.key === 'Escape') {
        setOpen(false)
        buttonRef.current?.focus()
      }
    }
    document.addEventListener('mousedown', onDocMouseDown)
    document.addEventListener('keydown', onKey)
    return () => {
      document.removeEventListener('mousedown', onDocMouseDown)
      document.removeEventListener('keydown', onKey)
    }
  }, [open])

  function handleKeyDown(e) {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (!open) {
        setOpen(true)
      } else {
        setActiveIdx((i) => Math.min(normalized.length - 1, i + 1))
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (!open) {
        setOpen(true)
      } else {
        setActiveIdx((i) => Math.max(0, i - 1))
      }
    } else if (e.key === 'Enter' || e.key === ' ') {
      if (!open) {
        e.preventDefault()
        setOpen(true)
      } else if (activeIdx >= 0) {
        e.preventDefault()
        onChange(normalized[activeIdx].value)
        setOpen(false)
        buttonRef.current?.focus()
      }
    } else if (e.key === 'Home') {
      if (open) {
        e.preventDefault()
        setActiveIdx(0)
      }
    } else if (e.key === 'End') {
      if (open) {
        e.preventDefault()
        setActiveIdx(normalized.length - 1)
      }
    }
  }

  const menu =
    open && menuPos
      ? createPortal(
          <div
            ref={menuRef}
            role="listbox"
            tabIndex={-1}
            style={{
              position: 'fixed',
              top: menuPos.top,
              left: menuPos.left,
              right: menuPos.right,
              width: menuPos.width,
              zIndex: 9999,
            }}
            className="max-h-72 overflow-auto rounded-xl border border-white/10 shadow-2xl shadow-black/60 animate-slide_up focus:outline-none"
            onKeyDown={handleKeyDown}
          >
            <div
              className="py-1.5 backdrop-blur-xl"
              style={{
                background:
                  'linear-gradient(180deg, rgba(19,26,43,0.97), rgba(12,18,32,0.97))',
              }}
            >
              <ul>
                {normalized.map((opt, i) => {
                  const isSelected = opt.value === value
                  const isActive = i === activeIdx
                  return (
                    <li
                      key={opt.value}
                      role="option"
                      aria-selected={isSelected}
                      data-active={isActive}
                      onMouseEnter={() => setActiveIdx(i)}
                      onClick={() => {
                        onChange(opt.value)
                        setOpen(false)
                        buttonRef.current?.focus()
                      }}
                      className={`relative flex items-center gap-2 px-3 py-2 mx-1 rounded-lg cursor-pointer text-sm select-none transition
                        ${isSelected ? 'text-cyan-100' : 'text-slate-200'}
                        ${
                          isActive
                            ? 'bg-gradient-to-l from-cyan-500/25 to-violet-500/25 ring-1 ring-cyan-400/40'
                            : 'hover:bg-white/5'
                        }
                      `}
                    >
                      {opt.icon && <span className="shrink-0">{opt.icon}</span>}
                      <span className="flex-1 truncate">{opt.label}</span>
                      {isSelected && (
                        <svg
                          width="14"
                          height="14"
                          viewBox="0 0 24 24"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="2.5"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          className="text-cyan-400 shrink-0"
                        >
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      )}
                    </li>
                  )
                })}
              </ul>
            </div>
          </div>,
          document.body,
        )
      : null

  return (
    <div ref={wrapRef} className={`relative ${className}`}>
      {label && <label className="label">{label}</label>}

      {/*
        Trigger layout (RTL-friendly):
          [icon] [label ………………] [chevron]
        In RTL flex (default row) the order goes start → end, i.e. right → left,
        so the icon sits at the right edge, label truncates in the middle, and
        the chevron sits at the LEFT edge — exactly where an indicator belongs.
      */}
      <button
        ref={buttonRef}
        type="button"
        onClick={() => setOpen((o) => !o)}
        aria-haspopup="listbox"
        aria-expanded={open}
        className={`w-full flex items-center gap-2 rounded-xl bg-white/5 border ps-3.5 pe-9 py-2 text-sm text-start transition
          focus:outline-none focus:ring-2 focus:ring-cyan-400/40 focus:border-cyan-400/40
          ${
            open
              ? 'border-cyan-400/40 ring-2 ring-cyan-400/30'
              : 'border-white/10 hover:bg-white/[0.07] hover:border-white/15'
          }
          ${isPlaceholder ? 'text-slate-500' : 'text-slate-100'}
        `}
      >
        {icon && <span className="shrink-0 text-slate-400">{icon}</span>}
        <span className="flex-1 min-w-0 truncate text-start">{displayLabel}</span>
        <span
          className={`shrink-0 text-slate-400 transition-transform ${
            open ? '-rotate-90 text-cyan-400' : 'rotate-90'
          }`}
          aria-hidden="true"
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <polyline points="9 18 15 12 9 6" />
          </svg>
        </span>
      </button>

      {menu}
    </div>
  )
}
