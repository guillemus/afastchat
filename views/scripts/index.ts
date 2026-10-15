import '/public/datastar-rocket-v1.0.4.js'
import './web_components'

declare global {
    interface Window {
        hiddenMessagesBelow(el: HTMLElement): number
    }
}

window.hiddenMessagesBelow = (el) => {
    const content = el.firstElementChild
    if (!content) {
        return 0
    }
    let bottom = content.getBoundingClientRect().bottom
    const pinned = content.querySelector(':scope > [data-scroll-pin]')
    if (pinned) {
        bottom = pinned.getBoundingClientRect().top
        for (const child of pinned.children) {
            bottom = Math.max(bottom, child.getBoundingClientRect().bottom)
        }
    }
    return Math.max(0, bottom - el.getBoundingClientRect().bottom)
}
