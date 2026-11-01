<template>
    <div class="mdx">
        <div ref="host" class="mdx__ace"></div>
        <div v-if="!modelValue" class="mdx__hint">{{ placeholder }}</div>
    </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { ACE_THEME, loadAce } from '../utils/aceLoader'

/**
 * Markdown editor for the mobile shell: Ace (already shipped in
 * /static/ace) with a hidden gutter and mobile-friendly options. Ace brings
 * syntax highlighting, undo and IME support without a 400 KB framework.
 */
const props = defineProps({
    modelValue: { type: String, default: '' },
    placeholder: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue'])

const host = ref(null)
let editor = null
let observer = null
let disposed = false

function onChange () {
    if (editor) {
        emit('update:modelValue', editor.getValue())
    }
}

onMounted(async () => {
    let ace
    try {
        ace = await loadAce()
    } catch (e) {
        console.error(e)
        return
    }
    if (disposed) {
        return
    }

    editor = ace.edit(host.value)
    editor.setTheme(ACE_THEME)
    editor.session.setMode('ace/mode/markdown')
    editor.setOptions({
        fontSize: '16px',
        showPrintMargin: false,
        showGutter: false,
        highlightActiveLine: false,
        showFoldWidgets: false,
        displayIndentGuides: false,
        wrap: true,
        useWorker: false,
        tabSize: 2,
        useSoftTabs: true,
        scrollPastEnd: 0.35
    })
    editor.renderer.setPadding(14)
    editor.renderer.setScrollMargin(8, 24)
    editor.setValue(props.modelValue || '', -1)
    editor.session.on('change', onChange)

    if (window.ResizeObserver) {
        observer = new ResizeObserver(() => {
            if (editor) {
                editor.resize()
            }
        })
        observer.observe(host.value)
    }
})

watch(() => props.modelValue, value => {
    if (editor && editor.getValue() !== (value || '')) {
        editor.setValue(value || '', -1)
    }
})

onBeforeUnmount(() => {
    disposed = true
    if (observer) {
        observer.disconnect()
        observer = null
    }
    if (editor) {
        editor.session.off('change', onChange)
        editor.destroy()
        editor = null
    }
})
</script>

<style scoped>
.mdx {
    position: relative;
    height: 100%;
}

.mdx__ace {
    position: absolute;
    inset: 0;
}

.mdx__hint {
    position: absolute;
    top: 14px;
    left: 14px;
    right: 14px;
    color: var(--m-text-muted);
    font-size: 1rem;
    pointer-events: none;
}
</style>
