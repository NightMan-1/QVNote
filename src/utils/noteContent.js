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
