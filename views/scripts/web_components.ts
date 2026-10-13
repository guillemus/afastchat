import { rocket } from '/public/datastar-rocket-v1.0.4.js'

rocket('sensitive-text', {
    mode: 'light',
    setup: ({ $$ }) => {
        $$('revealed', false)
    },
})

document.cookie = `tz=${Intl.DateTimeFormat().resolvedOptions().timeZone}; path=/; max-age=31536000; samesite=lax; secure`

abstract class ShadowText extends HTMLElement {
    static observedAttributes = ['datetime', 'since']
    private timer: number | undefined

    protected abstract text(): string
    protected interval(): number {
        return 0
    }

    connectedCallback() {
        this.render()
        const every = this.interval()
        if (every > 0) {
            this.timer = window.setInterval(() => this.render(), every)
        }
    }

    disconnectedCallback() {
        window.clearInterval(this.timer)
    }

    attributeChangedCallback() {
        this.render()
    }

    private render() {
        const root = this.shadowRoot ?? this.attachShadow({ mode: 'open' })
        root.textContent = this.text()
    }
}

customElements.define(
    'local-time',
    class extends ShadowText {
        protected text() {
            const date = new Date(this.getAttribute('datetime') ?? '')
            if (Number.isNaN(date.getTime())) return ''
            const pad = (n: number) => String(n).padStart(2, '0')
            return `${pad(date.getHours())}:${pad(date.getMinutes())}`
        }
    },
)

customElements.define(
    'elapsed-time',
    class extends ShadowText {
        protected interval() {
            return 250
        }
        protected text() {
            const since = Number(this.getAttribute('since'))
            return `${Math.max(1, Math.round((Date.now() - since) / 1000))}s`
        }
    },
)

customElements.define(
    'time-ago',
    class extends ShadowText {
        protected interval() {
            return 30_000
        }
        protected text() {
            const date = new Date(this.getAttribute('datetime') ?? '')
            if (Number.isNaN(date.getTime())) return ''
            const minutes = Math.floor((Date.now() - date.getTime()) / 60_000)
            if (minutes < 1) return 'now'
            if (minutes < 60) return `${minutes}m`
            const hours = Math.floor(minutes / 60)
            if (hours < 24) return `${hours}h`
            const days = Math.floor(hours / 24)
            if (days < 7) return `${days}d`
            const weeks = Math.floor(days / 7)
            if (weeks < 52) return `${weeks}w`
            return `${Math.floor(days / 365)}y`
        }
    },
)
