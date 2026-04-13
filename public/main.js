// scrollMessages makes the thread messages appear scrolled to the bottom, which is normally the
// place that you are mostly interested with when starting to read a thread. The
// requestAnimationFrame does avoid the manual scrollTop glitchiness.
function scrollMessages() {
    const el = document.querySelector('#messages')
    if (!el) return

    requestAnimationFrame(() => {
        el.scrollTop = el.scrollHeight
        requestAnimationFrame(() => {
            el.classList.remove('hidden')
        })
    })
}

window.addEventListener('pageshow', scrollMessages)
