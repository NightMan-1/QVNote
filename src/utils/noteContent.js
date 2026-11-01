import Prism from './prism'

/**
 * Prepares stored note HTML for display.
 *
 * Local .png links are rendered through the WebP converter (the /resources
 * handler serves a cached WebP for *.webp URLs); the editor keeps working with
 * the original .png links.
 */
export function renderNoteContent (content) {
    const html = String(content || '')
    return html.replace(/(\/resources\/[^"'\s]+?)\.png/gi, '$1.webp')
}

/** Applies Prism highlighting to code blocks inside the given element. */
export function highlightCodeBlocks (root) {
    if (root) {
        Prism.highlightAllUnder(root)
    }
}

/**
 * Handles clicks on the "refresh" button of a CodePen embed. The iframe is
 * proxied by the server, so refreshing means asking the server to re-fetch the
 * pen and reloading the frame with a cache-busting parameter.
 */
export function handleCodePenRefresh (event) {
    const btn = event.target.closest('.codepen-refresh')
    if (!btn) {
        return
    }
    event.preventDefault()
    const src = btn.getAttribute('data-codepen-refresh')
    if (!src) {
        return
    }
    const figure = btn.closest('.codepen-figure')
    const iframe = figure ? figure.querySelector('iframe') : null
    if (!iframe) {
        return
    }
    const separator = src.indexOf('?') >= 0 ? '&' : '?'
    btn.classList.add('codepen-refresh--loading')
    fetch(src + separator + 'refresh=1', { method: 'POST' })
        .catch(() => {})
        .then(() => {
            btn.classList.remove('codepen-refresh--loading')
            iframe.src = src + separator + 't=' + Date.now()
        })
}
