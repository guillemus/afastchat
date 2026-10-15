import { rocket } from '/public/datastar-rocket-v1.0.4.js'

rocket('sensitive-text', {
    mode: 'light',
    setup: ({ $$ }) => {
        $$('revealed', false)
    },
})

document.cookie = `tz=${Intl.DateTimeFormat().resolvedOptions().timeZone}; path=/; max-age=31536000; samesite=lax; secure`

type ClockContext = {
    host: HTMLElement
    cleanup: (fn: () => void) => void
    observeProps: (fn: () => void) => () => void
}

const syncText = (
    { host, cleanup, observeProps }: ClockContext,
    format: () => string,
    every?: number,
) => {
    const update = () => {
        const text = format()
        if (host.textContent !== text) host.textContent = text
    }
    update()
    cleanup(observeProps(update))
    if (every === undefined) return
    const timer = window.setInterval(update, every)
    cleanup(() => window.clearInterval(timer))
}

rocket('local-time', {
    mode: 'light',
    props: ({ date }) => ({ datetime: date }),
    setup: (ctx) =>
        syncText(ctx, () => {
            const pad = (n: number) => String(n).padStart(2, '0')
            return `${pad(ctx.props.datetime.getHours())}:${pad(ctx.props.datetime.getMinutes())}`
        }),
})

rocket('elapsed-time', {
    mode: 'light',
    props: ({ number }) => ({ since: number }),
    setup: (ctx) =>
        syncText(
            ctx,
            () => `${Math.max(1, Math.round((Date.now() - ctx.props.since) / 1000))}s`,
            250,
        ),
})

rocket('time-ago', {
    mode: 'light',
    props: ({ date }) => ({ datetime: date }),
    setup: (ctx) =>
        syncText(
            ctx,
            () => {
                const minutes = Math.floor((Date.now() - ctx.props.datetime.getTime()) / 60_000)
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
            30_000,
        ),
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

const htmlPreviewMaxHeight = 640

const htmlPreviewShim = `<style>:where(html) { color-scheme: dark; }</style><script>(() => {
    const post = () => {
        const body = document.body
        if (!body) return
        const style = getComputedStyle(body)
        const height = body.scrollHeight + parseFloat(style.marginTop) + parseFloat(style.marginBottom)
        parent.postMessage({ htmlPreviewHeight: height }, '*')
    }
    new ResizeObserver(post).observe(document.documentElement)
    addEventListener('load', post)
    post()
})()<\/script>`

rocket('html-preview', {
    mode: 'light',
    setup: ({ host, cleanup }) => {
        const frame = host.querySelector('.html-preview-frame')
        const code = host.querySelector('pre')?.textContent
        if (!frame || code === undefined) return

        const iframe = document.createElement('iframe')
        iframe.title = 'HTML preview'
        iframe.referrerPolicy = 'no-referrer'
        iframe.sandbox.add('allow-scripts')
        frame.replaceChildren(iframe)

        const onMessage = (evt: MessageEvent) => {
            if (evt.source !== iframe.contentWindow) return
            const height = evt.data?.htmlPreviewHeight
            if (typeof height !== 'number' || !Number.isFinite(height)) return
            iframe.style.height = `${Math.min(Math.ceil(height), htmlPreviewMaxHeight)}px`
        }
        window.addEventListener('message', onMessage)
        cleanup(() => window.removeEventListener('message', onMessage))

        const observer = new IntersectionObserver(
            (entries) => {
                if (!entries.some((entry) => entry.isIntersecting)) return
                observer.disconnect()
                iframe.srcdoc = code + htmlPreviewShim
            },
            { rootMargin: '200px' },
        )
        observer.observe(frame)
        cleanup(() => observer.disconnect())
    },
})
