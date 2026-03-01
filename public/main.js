window.datastarChat = window.datastarChat || {}

/**
 * Streaming markdown renderer. Pairs visible message elements with hidden buffer siblings.
 *
 * Server contract:
 * - containerSelector: element whose direct children include message divs (server inserts into this).
 * - messageSelector: CSS selector matching visible assistant message elements. Server must create
 *   these with a matching class and a stable `id`.
 * - bufferSuffix: for each assistant message the server must create a hidden sibling
 *   `<div id="{msgId}{bufferSuffix}" hidden>` and append raw HTML-escaped tokens to it.
 *   This function reads `textContent` from the buffer, parses with `marked.parse()`,
 *   and writes `innerHTML` to the paired visible element.
 *
 * @param {string} containerSelector - CSS selector for the messages container
 * @param {string} messageSelector  - CSS selector for visible assistant message elements
 * @param {string} bufferSuffix     - suffix appended to msgId to locate the hidden buffer div
 */
window.datastarChat.initMarkdown = function (containerSelector, messageSelector, bufferSuffix) {
    const container = document.querySelector(containerSelector)
    if (!container) return

    const pending = new Set()

    function getBuffer(el) {
        return document.getElementById(el.id + bufferSuffix)
    }

    function renderMarkdown(el) {
        const buf = getBuffer(el)
        if (!buf) return
        const raw = buf.textContent
        if (!raw) return
        el.innerHTML = marked.parse(raw)
    }

    function scheduleRender(el) {
        if (pending.has(el)) return
        pending.add(el)
        requestAnimationFrame(function () {
            pending.delete(el)
            renderMarkdown(el)
        })
    }

    function watchBuffer(msgEl) {
        const buf = getBuffer(msgEl)
        if (!buf) return
        bufferObserver.observe(buf, { childList: true, characterData: true, subtree: true })
        scheduleRender(msgEl)
    }

    // Watches hidden buffer divs for appended tokens
    const bufferObserver = new MutationObserver(function (mutations) {
        for (const m of mutations) {
            // Walk up to the hidden buffer, then find its paired message element
            const buf = m.target.closest ? m.target.closest('[hidden]') : m.target.parentElement
            if (!buf || !buf.id) continue
            const msgId = buf.id.replace(bufferSuffix, '')
            const msgEl = document.getElementById(msgId)
            if (msgEl) {
                scheduleRender(msgEl)
            }
        }
    })

    // Watches container for new message elements
    const containerObserver = new MutationObserver(function (mutations) {
        for (const m of mutations) {
            for (const node of m.addedNodes) {
                if (node.nodeType === 1 && node.matches(messageSelector)) {
                    watchBuffer(node)
                }
            }
        }
    })

    containerObserver.observe(container, { childList: true })
    container.querySelectorAll(messageSelector).forEach(watchBuffer)
}
