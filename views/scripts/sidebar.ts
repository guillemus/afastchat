declare global {
    interface Window {
        sidebarViewport(scroller: HTMLElement): { y: number; h: number }
        sidebarObserve(scroller: HTMLElement): void
        sidebarShouldFetch(el: HTMLElement, y: number, h: number, searchRev: number): boolean
    }
}

const designRemPixels = 16
const refetchAfterMs = 3000
const fetchedAt = new WeakMap<HTMLElement, number>()

window.sidebarViewport = (scroller) => {
    const virtual = scroller.querySelector('#thread-virtual') ?? scroller
    const remPixels =
        Number.parseFloat(getComputedStyle(document.documentElement).fontSize) || designRemPixels
    const scale = designRemPixels / remPixels
    return {
        y: Math.round(
            (scroller.getBoundingClientRect().top - virtual.getBoundingClientRect().top) * scale,
        ),
        h: Math.round(scroller.clientHeight * scale),
    }
}

window.sidebarObserve = (scroller) => {
    const observer = new ResizeObserver(() => scroller.dispatchEvent(new Event('scroll')))
    observer.observe(scroller)
    const items = scroller.querySelector('#thread-items')
    if (items) {
        observer.observe(items)
    }
}

window.sidebarShouldFetch = (el, y, h, searchRev) => {
    const now = performance.now()
    const last = fetchedAt.get(el)
    if (last !== undefined && now - last < refetchAfterMs) {
        return false
    }

    const bounds = el.dataset
    let stale = Number(bounds.searchRev) !== searchRev
    if (!stale) {
        const margin = Number(bounds.margin)
        const height = Math.min(Math.max(h, 0), Number(bounds.maxHeight))
        const lo = Math.max(0, y - margin)
        const hi = Math.min(Number(bounds.total), y + height + margin)
        stale = lo < hi && (lo < Number(bounds.top) || hi > Number(bounds.bottom))
    }
    if (stale) {
        fetchedAt.set(el, now)
    }
    return stale
}

export {}
