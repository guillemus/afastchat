/**
 * @param {string} key
 * @param {unknown} value
 */
function setSessionVal(key, value) {
    try {
        sessionStorage.setItem(key, JSON.stringify({ value }))
        console.debug('set sessionStorage', key, value)
    } catch (err) {
        console.error(err)
    }
}

window.setSessionVal = setSessionVal

/**
 * @param {string} key
 * @returns {unknown}
 */
function getSessionVal(key) {
    const raw = sessionStorage.getItem(key)

    if (!raw) return null

    try {
        const item = /** @type {{ value: unknown }} */ (JSON.parse(raw))
        console.debug('get sessionStorage', key, item.value)
        return item.value
    } catch (err) {
        console.error(err)
        return null
    }
}

window.getSessionVal = getSessionVal

/** @returns {void} */
function updateVisibleMinimapMessages() {
    const messages = document.getElementById('messages')
    const minimap = document.getElementById('conversation-minimap')
    if (!messages || !minimap) return

    const viewport = messages.getBoundingClientRect()
    const visibleIds = new Set()
    for (const message of messages.querySelectorAll('.message[id^="msg-"]')) {
        const rect = message.getBoundingClientRect()
        if (rect.bottom > viewport.top && rect.top < viewport.bottom) {
            visibleIds.add(message.id.slice(4))
        }
    }

    for (const card of minimap.querySelectorAll('.minimap-card[data-message-id]')) {
        const element = /** @type {HTMLElement} */ (card)
        element.classList.toggle('is-visible', visibleIds.has(element.dataset.messageId))
    }
}

window.updateVisibleMinimapMessages = updateVisibleMinimapMessages
