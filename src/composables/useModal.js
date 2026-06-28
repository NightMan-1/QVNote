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
            return modal
        }
    }
}
