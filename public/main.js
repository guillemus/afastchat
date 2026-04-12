function scrollElement() {
    const el = document.querySelector('#messages')
    if (!el) return

    requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight
        requestAnimationFrame(() => {
            el.classList.remove('hidden')
        })
    })
}

window.addEventListener('pageshow', scrollElement)
