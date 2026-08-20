import tingle from 'tingle.js'

const defaultOptions = {
    footer: true,
    stickyFooter: false,
    closeMethods: ['overlay', 'button', 'escape']
}

export function useModal () {
    return {
        createModal (closeLabel, options = {}) {
            const modal = new tingle.modal({
                ...defaultOptions,
                closeLabel,
                onClose: () => { modal.destroy() },
                ...options
            })
            // tingle crashes when destroy() runs twice (e.g. a footer button
            // destroys the modal and the close handler fires destroy again)
            let destroyed = false
            const origDestroy = modal.destroy.bind(modal)
            modal.destroy = () => {
                if (!destroyed) {
                    destroyed = true
                    origDestroy()
                }
            }
            return modal
        }
    }
}
