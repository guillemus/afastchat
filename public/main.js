function setSessionVal(key, value) {
    try {
        sessionStorage.setItem(key, JSON.stringify({ value }))
        console.debug('set sessionStorage', key, value)
    } catch (err) {
        console.error(err)
    }
}

window.setSessionVal = setSessionVal

function getSessionVal(key) {
    const raw = sessionStorage.getItem(key)

    if (!raw) return null

    try {
        const item = JSON.parse(raw)
        console.debug('get sessionStorage', key, item.value)
        return item.value
    } catch (err) {
        console.error(err)
        return null
    }
}

window.getSessionVal = getSessionVal
