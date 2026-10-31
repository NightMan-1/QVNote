<template>
    <div ref="host" class="ace-host"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import { loadAce, ACE_THEME } from '../utils/aceLoader'

/**
 * Thin wrapper around Ace (ace-builds), loaded on demand from /static/ace/.
 *
 * Ace is not ESM-friendly, so instead of bundling it into the chunk it is
 * loaded once per page from the static assets copied by the `postinstall`
 * script (see package.json) through the shared utils/aceLoader module.
 *
 * Mode/theme/ext files are fetched by Ace itself from the configured
 * basePath; only the files listed in postinstall are shipped.
 */
const props = defineProps({
    modelValue: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue'])

const host = ref(null)
let editor = null
let resizeObserver = null
let disposed = false

function onEditorChange () {
    if (editor) {
        emit('update:modelValue', editor.getValue())
    }
}

function applyValue (value) {
    if (editor && editor.getValue() !== (value || '')) {
        editor.setValue(value || '', -1)
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
    editor.session.setMode('ace/mode/html')
    editor.setOptions({
        fontSize: '14px',
        showPrintMargin: false,
        wrap: true,
        useWorker: false,
        tabSize: 4,
        useSoftTabs: true,
        displayIndentGuides: true,
        highlightActiveLine: true,
        showFoldWidgets: true
    })
    editor.setValue(props.modelValue || '', -1)
    editor.session.on('change', onEditorChange)

    if (window.ResizeObserver) {
        resizeObserver = new ResizeObserver(() => {
            if (editor) {
                editor.resize()
            }
        })
        resizeObserver.observe(host.value)
    }
})

watch(() => props.modelValue, applyValue)

onBeforeUnmount(() => {
    disposed = true
    if (resizeObserver) {
        resizeObserver.disconnect()
        resizeObserver = null
    }
    if (editor) {
        editor.session.off('change', onEditorChange)
        editor.destroy()
        editor = null
    }
})
</script>

<style scoped>
.ace-host {
    height: 100%;
    min-height: 400px;
    border: 1px solid var(--nord4);
    border-radius: .25rem;
    overflow: hidden;
}
</style>
