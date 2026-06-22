import { useEffect, useState } from 'react'

/**
 * Simple fullscreen-PWA installer.
 *
 * Listens for `beforeinstallprompt`, captures it, and on click fires it so
 * the browser shows its native install prompt. Falls back to a Persian hint
 * when no prompt is available (e.g. iOS Safari, browsers without PWA support).
 */
export default function InstallButton() {
  const [deferredPrompt, setDeferredPrompt] = useState(null)
  const [installed, setInstalled] = useState(false)

  useEffect(() => {
    function onBeforeInstall(e) {
      e.preventDefault()
      setDeferredPrompt(e)
    }
    function onInstalled() {
      setInstalled(true)
      setDeferredPrompt(null)
    }
    window.addEventListener('beforeinstallprompt', onBeforeInstall)
    window.addEventListener('appinstalled', onInstalled)
    return () => {
      window.removeEventListener('beforeinstallprompt', onBeforeInstall)
      window.removeEventListener('appinstalled', onInstalled)
    }
  }, [])

  async function handleClick() {
    if (!deferredPrompt) return
    await deferredPrompt.prompt()
    const { outcome } = await deferredPrompt.userChoice
    if (outcome === 'accepted') setInstalled(true)
    setDeferredPrompt(null)
  }

  if (installed) {
    return (
      <section className="card p-4 md:p-5">
        <div className="flex items-center gap-2 text-emerald-300">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <polyline points="20 6 9 17 4 12" />
          </svg>
          <span className="text-sm font-semibold">اپلیکیشن نصب شد</span>
        </div>
        <p className="text-xs text-slate-400 mt-1.5 leading-relaxed">
          می‌توانید از منوی برنامه‌های دستگاه خود «قطعی برق» را اجرا کنید.
        </p>
      </section>
    )
  }

  if (!deferredPrompt) return null

  return (
    <section className="card p-4 md:p-5">
      <p className="text-xs text-slate-400 mb-3 leading-relaxed">
        اپلیکیشن تحت وب را روی دستگاه خود نصب کنید تا در حالت تمام‌صفحه و
        بدون نوار آدرس در دسترس باشد.
      </p>
      <button
        onClick={handleClick}
        type="button"
        className="btn-primary w-full"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
          <polyline points="7 10 12 15 17 10" />
          <line x1="12" y1="15" x2="12" y2="3" />
        </svg>
        نصب اپلیکیشن
      </button>
    </section>
  )
}
