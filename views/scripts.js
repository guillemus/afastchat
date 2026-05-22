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
