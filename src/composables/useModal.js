import tingle from '../vendor/tingle/tingle.js'

const defaultOptions = {
    footer: true,
    stickyFooter: false,
    closeMethods: ['overlay', 'button', 'escape']
}

/**
 * Thin wrapper around the vendored tingle.js (see src/vendor/tingle).
 *
 * The vendored copy already handles idempotent destroy(), per-instance busy
 * state, focus management and the scroll lock, so no local patches are needed.
 *
 * Modal content is set as an HTML string (setContent → innerHTML): escape any
 * interpolated data with `escapeHtml` from utils/escape.js.
 */
export function useModal () {
    return {
        createModal (closeLabel, options = {}) {
            const modal = new tingle.modal({
                ...defaultOptions,
                closeLabel,
                onClose: () => { modal.destroy() },
                ...options
            })
            return modal
        }
    }
}
