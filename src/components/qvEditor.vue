<template>
    <div id="qv-editor">
        <div id="qv-editor-header">
            <h4 class="mt-1 mb-0 float-start" v-if="(articleCurrentEditable.uuid === '')">{{$t('editor.titleNew')}}</h4>
            <h4 class="mt-1 mb-0 float-start" v-if="(articleCurrentEditable.uuid !== '')">{{$t('editor.titleExist')}}</h4>
            <div class="float-end">
                <button class="btn btn-outline-primary ms-auto"
                        @click="$router.push('/notes/' + articleCurrentEditable.NoteBookUUID + '/' + articleCurrentEditable.uuid + '/')"
                        v-if="(articleCurrentEditable.uuid !== '' && articleCurrentEditable.NoteBookUUID !== '')"><i class="bi bi-eye-fill"></i></button>
                <button class="btn btn-outline-success ms-2" @click="saveData"
                        :class="{'ms-auto':articleCurrentEditable.uuid === '' && articleCurrentEditable.NoteBookUUID === ''}"><i class="bi bi-floppy-fill"></i>
                </button>
                <!--<button class="btn btn-outline-secondary ms-2"><i class="bi bi-eraser text-muted"></i></button>-->
                <div class="btn-group ms-4" role="group" aria-label="Button group">
                    <button class="btn btn-outline-secondary" :class="{'active':articleCurrentEditable.type === 'text'}"
                            @click="articleCurrentEditable.type = 'text'"><i class="bi bi-pencil-fill"></i></button>
                    <!--<button class="btn btn-outline-secondary" :class="{'active':editorType === 'markdown'}" @click="editorType = 'markdown'"><i class="bi bi-layout-split"></i></button>-->
                    <button class="btn btn-outline-secondary" :class="{'active':articleCurrentEditable.type === 'code'}"
                            @click="articleCurrentEditable.type = 'code'"><i class="bi bi-code-slash"></i></button>
                </div>
            </div>
        </div>

        <div id="qv-editor-main">
                <input type="text" class="form-control mb-2 mt-3 text-dark font-size-normal"
                        :placeholder="$t('editor.inputTitlePlaceholder')" v-model="articleCurrentEditable.title" ref='editorTitle'/>
                <div class="row">
                    <div class="col-6">
                        <div class="mb-3">
                            <label><b>{{$t('editor.titleURL')}}</b></label>
                            <div class="input-group">
                                <span class="input-group-text"><i class="bi bi-box-arrow-up-right"></i></span>
                                <input type="text" class="form-control text-dark font-size-normal"
                                        :placeholder="$t('editor.inputURLPlaceholder')" v-model="articleCurrentEditable.url_src"/>
                            </div>
                        </div>
                    </div>
                    <div class="col-6">
                        <div class="mb-3">
                            <label><b>{{$t('editor.titleTags')}}</b></label>
                            <multiselect
                                ref='editorTags'
                                v-model="articleCurrentEditable.tags"
                                :placeholder="$t('editor.inputTagsPlaceholder')"
                                :options="tagsListFormatted"
                                :multiple="true"
                                :taggable="true"
                                @tag="addTag"
                                :selectLabel="multiselectLang.selectLabel"
                                :deselectLabel="multiselectLang.deselectLabel"
                                :selectedLabel="multiselectLang.selectedLabel"></multiselect>
                        </div>

                    </div>
                </div>
                 <div class="editor mt-2" v-if="articleCurrentEditable.type === 'text' && editorReady">
                    <Editor v-model="articleCurrentEditable.content" :init="editorConfig" />
                </div>
                <div class="editor prism mt-2" v-if="articleCurrentEditable.type === 'code'">
                    <prism-editor v-model="articleCurrentEditable.content" language="html" :line-numbers="true" :highlight="highlighter"></prism-editor>
                </div>
                <div class="clearfix"></div>
        </div>
    </div>
</template>

<script>
import { useNoteStore } from '../store'
import Prism from 'prismjs'
import 'prismjs/components/prism-markup'
import { PrismEditor } from 'vue-prism-editor'
import 'vue-prism-editor/dist/prismeditor.min.css'

import Editor from '@hugerte/hugerte-vue'
import 'ace-builds/src-min-noconflict/ace'
import 'ace-builds/src-min-noconflict/theme-crimson_editor'
import 'ace-builds/src-min-noconflict/mode-html'
import 'ace-builds/src-min-noconflict/ext-language_tools'
import { html as htmlBeautify } from 'js-beautify'

if (window.ace && window.ace.config) {
    window.ace.config.set('basePath', '/static/ace/')
}
window.html_beautify = htmlBeautify
// Self-hosted: core JS bundled by Vite, static assets (skins/icons) served from public/static/hugerte/
import 'hugerte'
import 'hugerte/themes/silver'
import 'hugerte/icons/default'
import 'hugerte/models/dom'
// Import plugins so HugeRTE finds them already loaded (no dynamic loading)
import 'hugerte/plugins/lists'
import 'hugerte/plugins/link'
import 'hugerte/plugins/image'
import 'hugerte/plugins/fullscreen'
import 'hugerte/plugins/codesample'

import Multiselect from 'vue-multiselect'
import { registerBootstrapIcons } from '../hugerte-icons'
// import { html as BeautifyHtml } from 'js-beautify'

export default {
    name: 'qvEditor',
    data () {
        return {
            multiselectLang: {
                selectLabel: this.$t('editor.multiselectLang.selectLabel'),
                deselectLabel: this.$t('editor.multiselectLang.deselectLabel'),
                selectedLabel: this.$t('editor.multiselectLang.selectedLabel')
            },
            articleCurrentEditable: { title: '', uuid: '', NoteBookUUID: '', status: '', tags: [], CreatedDate: '', UpdatedDate: '', cells: {}, content: '', type: 'text', url_src: '' },
            tagsListFormatted: [],
            editorReady: false,
            editorConfig: {
                skin_url: '/static/hugerte/skins/ui/oxide',
                content_css: ['/static/hugerte/skins/content/default/content.css', '/static/prism/prism-atom-one-light.css'],
                theme_url: '/static/hugerte/themes/silver/theme.min.js',
                toolbar: 'undo redo | blocks | bold italic underline | blockquote inlinecode codeblock removeformat | bullist numlist | forecolor backcolor | alignleft aligncenter alignright | link image | eraser | supercode | fullscreen',
                plugins: 'lists link image codesample fullscreen',
                external_plugins: {
                    codesample: '/static/hugerte/plugins/codesample/plugin.min.js',
                    supercode: '/static/supercode/plugin.min.js'
                },
                supercode: {
                    theme: 'crimson_editor',
                    language: 'html',
                    fontSize: 14,
                    wrap: true,
                    fallbackModal: false,
                    shortcut: true
                },
                codesample_languages: [
                    { text: 'Plain text', value: 'plaintext' },
                    { text: 'HTML/XML', value: 'markup' },
                    { text: 'JavaScript', value: 'javascript' },
                    { text: 'CSS', value: 'css' },
                    { text: 'PHP', value: 'php' },
                    { text: 'Ruby', value: 'ruby' },
                    { text: 'Python', value: 'python' },
                    { text: 'Java', value: 'java' },
                    { text: 'C', value: 'c' },
                    { text: 'C#', value: 'csharp' },
                    { text: 'C++', value: 'cpp' },
                    { text: 'Bash', value: 'bash' },
                    { text: 'SQL', value: 'sql' },
                    { text: 'JSON', value: 'json' }
                ],
                menubar: true,
                branding: false,
                // Word paste cleanup — handled by paste_from_word plugin
                paste_data_images: true,
                // Disable automatic upload — images are embedded as base64 via paste_data_images
                automatic_uploads: false,
                content_style: 'body { font-family: Montserrat, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; font-size: 16px; line-height: 1.7; margin: 0; padding: 0 1rem 0 1rem; } h1 { font-size: 2.2rem; font-weight: 500; } h2 { font-size: 1.8rem; font-weight: 500; } h3 { font-size: 1.5rem; font-weight: 500; } h4 { font-size: 1.25rem; } h5 { font-size: 1.1rem; } h6 { font-size: 1rem; color: #6c757d; } p { margin-bottom: 0.75rem; } img { max-width: 100%; height: auto; } pre:not([class*="language-"]) { background-color: #f0f3f5; color: #363636; padding: 0.5rem; } pre:not([class*="language-"]) code { background-color: transparent; color: inherit; padding: 0; }  .mce-content-body pre [data-mce-selected="inline-boundary"] { background-color: transparent; } :not(pre) > code[class*="language-"], pre[class*="language-"] {color: #383942;} pre[class*="language-"] {padding: 0.5rem; margin: 1rem 0;}',
                setup: (editor) => {
                    registerBootstrapIcons(editor)
                    // Syntax highlighting for code blocks using PrismJS
                    editor.on('init', function () {
                        var doc = editor.getDoc()
                        if (!doc) return
                        var highlightCodeBlocks = function () {
                            // PrismJS is loaded by codesample plugin
                            if (window.Prism && window.Prism.highlightElement) {
                                var pres = doc.querySelectorAll('pre[class*="language-"]')
                                for (var i = 0; i < pres.length; i++) {
                                    // Check if already highlighted
                                    if (!pres[i].querySelector('.token')) {
                                        try { window.Prism.highlightElement(pres[i]) } catch(e) {}
                                    }
                                }
                            }
                        }
                        editor.on('SetContent', highlightCodeBlocks)
                        editor.on('NodeChange', highlightCodeBlocks)
                        // Also highlight after a short delay to catch async plugin loading
                        setTimeout(highlightCodeBlocks, 500)
                    })
                    editor.ui.registry.addButton('fullscreen', {
                        icon: 'fullscreen',
                        tooltip: 'Fullscreen',
                        onAction: () => {
                            if (!document.fullscreenElement) {
                                document.querySelector('.editor').requestFullscreen()
                            } else {
                                document.exitFullscreen()
                            }
                        }
                    })
                    editor.ui.registry.addButton('eraser', {
                        icon: 'eraser',
                        tooltip: 'Clean HTML',
                        onAction: () => {
                            fetch(this.noteStore.apiFolder + '/cleanup_html.json', {
                                method: 'POST',
                                body: JSON.stringify({ content: editor.getContent() })
                            }).then(response => response.json())
                                .then(jsonData => {
                                    editor.setContent(jsonData.content)
                                })
                                .catch(error => {
                                    console.error('Error cleanup html:', error)
                                })
                        }
                    })
                    // Inline code — wraps selection in <code>, or creates a code block for multi-block selections
                    editor.ui.registry.addButton('inlinecode', {
                        icon: 'inlinecode',
                        tooltip: 'Inline code',
                        onAction: () => {
                            var rng = editor.selection.getRng()
                            if (!rng.collapsed) {
                                var startBlock = editor.dom.getParent(rng.startContainer, editor.dom.isBlock)
                                var endBlock = editor.dom.getParent(rng.endContainer, editor.dom.isBlock)
                                if (startBlock && endBlock && startBlock !== endBlock) {
                                    var text = editor.selection.getContent({ format: 'text' })
                                    if (!text || !text.trim()) {
                                        text = ' '
                                    }
                                    var escaped = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
                                    var html = '<pre class="language-plaintext"><code>' + escaped + '</code></pre><p><br></p>'
                                    editor.execCommand('mceInsertContent', false, html)
                                    return
                                }
                            }
                            editor.execCommand('mceToggleFormat', false, 'code')
                        }
                    })
                    // Code block — opens the HugeRTE Insert/Edit Code Sample dialog
                    editor.ui.registry.addButton('codeblock', {
                        icon: 'codeblock',
                        tooltip: 'Code block',
                        onAction: () => {
                            var node = editor.selection.getNode()
                            var pre = editor.dom.getParent(node, 'pre')
                            if (pre && /language-/.test(pre.className)) {
                                // Editing an existing code sample — open the dialog
                                editor.execCommand('codesample')
                                return
                            }
                            var rng = editor.selection.getRng()
                            if (!rng.collapsed) {
                                var text = editor.selection.getContent({ format: 'text' })
                                if (text && text.trim()) {
                                    var escaped = text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
                                    editor.execCommand('mceInsertContent', false, '<pre class="language-plaintext"><code>' + escaped + '</code></pre><p><br></p>')
                                    return
                                }
                            }
                            // No selection — open the dialog to insert a new code sample
                            editor.execCommand('codesample')
                        }
                    })
                    // Paste cleanup: tag unclassified <pre><code> as plaintext code samples
                    editor.on('PastePreProcess', function (e) {
                        if (!e.content || typeof e.content !== 'string') return
                        var parser = new DOMParser()
                        var doc = parser.parseFromString(e.content, 'text/html')
                        var pres = doc.querySelectorAll('pre')
                        var changed = false
                        for (var i = 0; i < pres.length; i++) {
                            var pre = pres[i]
                            if (pre.getAttribute('class')) continue
                            var hasDirectCode = false
                            for (var c = pre.firstChild; c; c = c.nextSibling) {
                                if (c.nodeType === 1 && c.tagName === 'CODE') {
                                    hasDirectCode = true
                                    break
                                }
                            }
                            if (hasDirectCode) {
                                pre.setAttribute('class', 'language-plaintext')
                                changed = true
                            }
                        }
                        if (changed) {
                            e.content = doc.body.innerHTML
                        }
                    })

                    // Paste from Word cleanup
                    var pasteFromWordLib = window['tinymce-paste-from-word-lib']
                    if (pasteFromWordLib && typeof pasteFromWordLib.default === 'function') {
                        editor.on('PastePreProcess', function (e) {
                            pasteFromWordLib.default(editor, e)
                        })
                    }
                }
            }
        }
    },
    components: {
        Editor,
        Multiselect,
        PrismEditor
    },
    setup () {
        return { noteStore: useNoteStore() }
    },
    computed: {
        articleCurrent () { return this.noteStore.currentArticle },
        tagsList () { return this.noteStore.tagsList }
    },
    mounted () {
        // Load paste_from_word library before rendering the editor.
        // The library is a TinyMCE UMD module that expects window.tinymce.
        if (typeof window.tinymce === 'undefined' && typeof window.hugerte !== 'undefined') {
            window.tinymce = window.hugerte
        }
        var self = this
        var script = document.createElement('script')
        script.src = '/static/hugerte/plugins/paste_from_word/plugin.min.js'
        script.onload = function () {
            self.editorReady = true
        }
        script.onerror = function () {
            // Still render the editor even if the library failed to load
            self.editorReady = true
        }
        document.head.appendChild(script)

        this.$refs.editorTitle.focus()
        // this.articleCurrentEditable = Object.assign({}, this.articleCurrent)
        this.articleCurrentEditable = JSON.parse(JSON.stringify(this.articleCurrent))
        if (this.articleCurrentEditable.type === '') {
            this.articleCurrentEditable.type = 'text'
        }

        for (const tag in this.tagsList) {
            if (this.tagsList[tag].name !== '') {
                this.tagsListFormatted.push(this.tagsList[tag].name)
            }
        }
    },
    methods: {
        highlighter (code) {
            if (!code) return ''
            return Prism.highlight(code, Prism.languages.markup, 'markup')
        },
        saveData () {
            fetch(this.noteStore.apiFolder + '/note_edit.json',
                { method: 'POST',
                    body: JSON.stringify({
                        'title': this.articleCurrentEditable.title,
                        'url': this.articleCurrentEditable.url_src,
                        'uuid': this.articleCurrentEditable.uuid,
                        'type': this.articleCurrentEditable.type,
                        'tags': this.articleCurrentEditable.tags,
                        'content': this.articleCurrentEditable.content
                    }) }).then(response => { return response.json() })
                .then(jsonData => {
                    this.articleCurrentEditable.uuid = jsonData.uuid
                    this.articleCurrentEditable.NoteBookUUID = jsonData.NoteBookUUID
                    // this.articleCurrentEditable.content = jsonData.html // slow
                    this.noteStore.getAllData()
                })
                .catch(error => {
                    console.error('Error save note data:', error)
                    this.noteStore.setStatus({ errorType: 2, errorText: this.$t('editor.errorSave') })
                })
        },
        addTag (newTag) {
            this.articleCurrentEditable.tags.push(newTag)
            this.$refs.editorTags.$el.focus()
        }
    },
    watch: {
        'articleCurrentEditable.type' () {
            if (this.articleCurrentEditable.type === 'code') {
                // this.articleCurrentEditable.content = BeautifyHtml(this.articleCurrentEditable.content)
            }
        }
    }

}
</script>

<style src="vue-multiselect/dist/vue-multiselect.css"></style>

<style>
    /* HugeRTE editor area margin */
    .tox .tox-edit-area { margin: 0; }
    .tox .tox-edit-area::before { display: none; }
    /* Dynamic editor height */
    .tox.tox-hugerte {
        height: calc(100vh - 270px) !important;
        min-height: 40rem;
    }
    .tox.tox-hugerte .tox-edit-area {
        height: 100%;
    }
    /* Constrain image resize handles */
    .tox .tox-image-tools__image-bg,
    .tox .tox-image-tools__image {
        max-width: 100% !important;
    }
    .tox .tox-image-tools__resize-handle {
        max-width: 100%;
    }

    /* Prism editor: fix invisible cursor and text selection */
    .prism-editor-wrapper .prism-editor__textarea {
        caret-color: #363636;
    }
    .prism-editor-wrapper .prism-editor__textarea::selection {
        background-color: rgba(32, 168, 216, 0.3);
    }
    .prism-editor-wrapper .prism-editor__textarea::-moz-selection {
        background-color: rgba(32, 168, 216, 0.3);
    }
    .prism-editor-wrapper :is(.prism-editor__editor, .prism-editor__textarea) {
        font-variant-numeric: tabular-nums;
    }
    .tox-shadowhost.tox-fullscreen, .tox.tox-hugerte.tox-fullscreen {
        height: 100% !important;
    }
</style>

<style scoped>
    /* purgecss start ignore */
    #qv-editor{
        padding: .5rem 1.5rem 2rem;
    }
    #qv-editor-header{
        position: fixed;
        width: calc(100% - var(--sidebar-width));
        top: 0;
        left:var(--sidebar-width);
        z-index: 1000;
        background-color: var(--nord6);
        border-bottom: 1px solid var(--nord4);
        padding: .5rem .75rem;
    }
    #qv-editor-main{
        overflow: hidden;
    }

    .font-size-normal {
        font-size: 1em;
    }

    .editor {
      min-height: 100% !important;
      display: flex;
      flex-direction: column;
    }

    .editor.prism {
      background-color: rgb(240, 243, 245);
      height: 70vh;
      overflow: auto;
      font-size: .85rem;
    }

    .prism-editor__editor {
        background: none !important;
        padding: 0 !important;
        margin: 0 !important;
    }
    .prism-editor-wrapper .prism-editor__container {
        background: white;
        padding-left: 5px;
    }

    .multiselect {
        font-size: 1rem;
        font-weight: 400;
        min-height: auto;
    }

    .multiselect__tags {
        border-radius: 0.25rem;
        font-size: 1rem;
        font-weight: 400;
        line-height: 1.2;
        padding: .5em 0 0 .7em;
        min-height: 1rem;
        border-color: #dcdfe2;
        margin-bottom: 4px;
    }

    .multiselect__tags:focus {
        border-color: #8ad4ee;
        box-shadow: 0 0 0 .2rem rgba(32, 168, 216, .25);
    }

    .multiselect__tag {
        border-radius: 0.25rem;
        font-size: 0.9rem;
        /*background: var(--nord14);*/
    }

    .multiselect__input, .multiselect__single {
        padding-left: 0;
        margin-bottom: 0.425em;
        font-size: 1rem;
        font-weight: 400;
    }

    .multiselect__single {
        color: rgba(0, 0, 0, 0.5);
    }

    .multiselect__select {
        width: 2.85rem;
        height: 2.3rem;
    }
    .multiselect__placeholder{
        margin-bottom: 0.53rem;
    }
    .prism-editor__code, pre[class*="language-"] { overflow: inherit; margin: 0;}
    .prism-editor__code code { padding: 0; overflow: inherit;}
    .prism-editor__line-numbers {
        width: 2rem;
        float: left;
        margin-top: 0 !important;
        text-align: right;
        margin-right: 0.5rem;
    }

    /* purgecss end ignore */
</style>
