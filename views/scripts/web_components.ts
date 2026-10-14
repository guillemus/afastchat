import { rocket } from '/public/datastar-rocket-v1.0.4.js'

rocket('sensitive-text', {
    mode: 'light',
    setup: ({ $$ }) => {
        $$('revealed', false)
    },
})

document.cookie = `tz=${Intl.DateTimeFormat().resolvedOptions().timeZone}; path=/; max-age=31536000; samesite=lax; secure`

const tick =
    (every: number) =>
    ({ render, cleanup }: { render: (ctx: object) => void; cleanup: (fn: () => void) => void }) => {
        const timer = window.setInterval(() => render({}), every)
        cleanup(() => window.clearInterval(timer))
    }

rocket('local-time', {
    props: ({ date }) => ({ datetime: date }),
    render: ({ props }) => {
        const pad = (n: number) => String(n).padStart(2, '0')
        return `${pad(props.datetime.getHours())}:${pad(props.datetime.getMinutes())}`
    },
})

rocket('elapsed-time', {
    props: ({ number }) => ({ since: number }),
    setup: tick(250),
    render: ({ props }) => `${Math.max(1, Math.round((Date.now() - props.since) / 1000))}s`,
})

rocket('time-ago', {
    props: ({ date }) => ({ datetime: date }),
    setup: tick(30_000),
    render: ({ props }) => {
        const minutes = Math.floor((Date.now() - props.datetime.getTime()) / 60_000)
        if (minutes < 1) return 'now'
        if (minutes < 60) return `${minutes}m`
        const hours = Math.floor(minutes / 60)
        if (hours < 24) return `${hours}h`
        const days = Math.floor(hours / 24)
        if (days < 7) return `${days}d`
        const weeks = Math.floor(days / 7)
        if (weeks < 52) return `${weeks}w`
        return `${Math.floor(days / 365)}y`
    },
})

rocket('flip-list', {
    mode: 'light',
    setup: ({ host, defineHostProp }) => {
        const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)')

        defineHostProp('capture', {
            value: () => {
                if (reducedMotion.matches) return

                const items = ([...host.children] as HTMLElement[]).filter(
                    (item) => item.offsetParent !== null,
                )
                const before = new Map(
                    items.map((item) => [item, item.getBoundingClientRect().top]),
                )

                requestAnimationFrame(() => {
                    for (const item of items) {
                        const from = before.get(item)
                        if (from === undefined || item.offsetParent === null) continue

                        const dy = from - item.getBoundingClientRect().top
                        if (dy === 0) continue

                        item.animate(
                            [{ transform: `translateY(${dy}px)` }, { transform: 'translateY(0)' }],
                            {
                                duration: 220,
                                easing: 'cubic-bezier(0.2, 0, 0, 1)',
                            },
                        )
                    }
                })
            },
        })
    },
})
