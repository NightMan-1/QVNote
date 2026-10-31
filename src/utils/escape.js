/**
 * Escapes a value for interpolation into an HTML string.
 *
 * Use it before putting untrusted data (notebook/tag names, URLs coming from
 * imported libraries) into the markup produced for tingle modals, which is
 * assigned through innerHTML.
 */
export function escapeHtml (value) {
    if (value === null || value === undefined) {
        return ''
    }
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
}
