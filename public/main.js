/**
 * @param {string} key
 * @param {string} value
 * @param {number} ms
 */
function setSessionVal(key, value, ms) {
    sessionStorage.setItem(key, JSON.stringify({ value, expiresAt: Date.now() + ms }))
}

window.setSessionVal = setSessionVal

/**
 * @param {string} key
 * @returns {string | null}
 */
function getSessionVal(key) {
    const raw = sessionStorage.getItem(key)

    if (!raw) return null

    const item = JSON.parse(raw)

    if (item.expiresAt <= Date.now()) {
        sessionStorage.removeItem(key)
        return null
    }

    return item.value
}

window.getSessionVal = getSessionVal
