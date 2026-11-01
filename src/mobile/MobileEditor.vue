<template>
    <div class="med">
        <header class="med__bar">
            <button class="med__bar-btn" @click="cancel">{{ $t('general.cancel') }}</button>
            <div class="med__bar-title">{{ isNew ? $t('editor.titleNew') : $t('editor.titleExist') }}</div>
            <button
                v-if="mode === 'text'"
                class="med__bar-btn med__bar-btn--icon"
                :class="{ 'med__bar-btn--on': preview }"
                :aria-label="$t('editor.preview')"
                @click="preview = !preview"
            >
                <i class="bi" :class="preview ? 'bi-pencil' : 'bi-eye'"></i>
            </button>
            <button class="med__bar-btn med__bar-btn--save" :disabled="!canSave || saving" @click="save">
                <span v-if="saving" class="med__spinner"></span>
                <i v-else class="bi bi-check-lg"></i>
            </button>
        </header>

        <div class="med__meta">
            <input v-model="title" class="med__title-input" :placeholder="$t('editor.inputTitlePlaceholder')">

            <div class="med__tags">
                <span v-for="tag in tags" :key="tag" class="med__tag">
                    #{{ tag }}
                    <button class="med__tag-x" :aria-label="tag" @click="removeTag(tag)">
                        <i class="bi bi-x"></i>
                    </button>
                </span>
                <input
                    v-model="tagInput"
                    class="med__tag-input"
                    :placeholder="tags.length ? '' : $t('editor.inputTagsPlaceholder')"
                    @keydown.enter.prevent="addTagFromInput"
                >
            </div>
            <div v-if="tagSuggestions.length" class="med__suggest">
                <button v-for="name in tagSuggestions" :key="name" @click="addTag(name)">#{{ name }}</button>
            </div>

            <div v-if="isNew" class="med__modes">
                <button :class="{ 'med__mode--active': mode === 'text' }" @click="mode = 'text'">
                    {{ $t('editor.modeText') }}
                </button>
                <button :class="{ 'med__mode--active': mode === 'link' }" @click="mode = 'link'">
                    {{ $t('editor.modeLink') }}
                </button>
            </div>

            <div v-if="draftRestored" class="med__draft">
                <span>{{ $t('editor.draftRestored') }}</span>
                <button @click="clearDraft">{{ $t('editor.draftClear') }}</button>
            </div>
        </div>

        <div v-if="mode === 'text'" class="med__body">
            <MarkdownEditor v-if="!preview" v-model="markdown" :placeholder="$t('editor.markdownPlaceholder')" />
            <div v-else class="med__preview m-note__content" v-html="previewHtml"></div>
        </div>

        <div v-else class="med__body med__body--link">
            <div class="med__link-row">
                <input
                    v-model="linkUrl"
                    class="med__link-input"
                    type="url"
                    inputmode="url"
                    placeholder="https://..."
                    @keydown.enter.prevent="loadLink"
                >
                <button class="med__load" :disabled="!linkUrl.trim() || importing" @click="loadLink">
                    <span v-if="importing" class="med__spinner"></span>
                    <i v-else class="bi bi-cloud-arrow-down"></i>
                </button>
            </div>
            <p class="med__hint">{{ $t('editor.defuddleHint') }}</p>
            <div v-if="linkHtml" class="med__preview m-note__content" v-html="linkHtml"></div>
        </div>

        <div v-if="loading" class="med__loading">
            <div class="med__spinner med__spinner--big"></div>
        </div>
    </div>
</template>

<script>
import MarkdownIt from 'markdown-it'
import { push } from 'notivue'
import { useNoteStore } from '../store'
import { fetchDefuddleArticle } from '../utils/defuddle'
import MarkdownEditor from './MarkdownEditor.vue'

const md = new MarkdownIt({ html: true, linkify: true })

const DRAFT_KEY = 'qvnote-draft-new'

/**
 * Mobile editor. Two scenarios only:
 *  - "Text": a small markdown note with syntax highlighting (stored as HTML,
 *    converted by the server, so the Quiver format does not change);
 *  - "Link": the same defuddle import the desktop editor uses.
 * Imported notes are not round-tripped through markdown: they get a
 * "prepend a note at the top" action in the note sheet instead.
 */
export default {
    name: 'MobileEditor',
    components: { MarkdownEditor },
    data () {
        return {
            isNew: true,
            uuid: '',
            type: 'text',
            title: '',
            tags: [],
            markdown: '',
            url: '',
            mode: 'text',
            preview: false,
            tagInput: '',
            linkUrl: '',
            linkHtml: '',
            importing: false,
            loading: false,
            saving: false,
            draftRestored: false,
            draftTimer: null
        }
    },
    setup () {
        return { noteStore: useNoteStore() }
    },
    computed: {
        apiFolder () { return this.noteStore.apiFolder },
        tagsList () { return Array.isArray(this.noteStore.tagsList) ? this.noteStore.tagsList : [] },
        previewHtml () { return md.render(this.markdown || '') },
        canSave () {
            if (this.mode === 'link') {
                return !!this.linkHtml
            }
            return !!(this.title.trim() || this.markdown.trim())
        },
        tagSuggestions () {
            const query = this.tagInput.trim().toLowerCase()
            if (!query) {
                return []
            }
            const added = new Set(this.tags)
            return this.tagsList
                .filter(tag => tag.name.toLowerCase().includes(query) && !added.has(tag.name))
                .slice(0, 8)
                .map(tag => tag.name)
        }
    },
    watch: {
        title () { this.queueDraft() },
        markdown () { this.queueDraft() },
        tags () { this.queueDraft() }
    },
    created () {
        const article = this.noteStore.currentArticle || {}
        if (article.uuid) {
            this.isNew = false
            this.uuid = article.uuid
            this.type = article.type || 'text'
            this.title = article.title || ''
            this.tags = Array.isArray(article.tags) ? [...article.tags] : []
            this.url = article.url_src || ''
            this.loadMarkdown()
        } else {
            this.restoreDraft()
        }
    },
    beforeUnmount () {
        clearTimeout(this.draftTimer)
    },
    methods: {
        loadMarkdown () {
            this.loading = true
            fetch(this.apiFolder + '/note.json', {
                method: 'POST',
                body: JSON.stringify({ NoteID: this.uuid, raw: true, format: 'markdown' })
            })
                .then(response => response.json())
                .then(data => {
                    if (!data || data.content === undefined) {
                        push.error({ message: this.$t('editor.loadOriginalError') })
                        return
                    }
                    this.markdown = data.content
                    if (data.title) {
                        this.title = data.title
                    }
                    if (Array.isArray(data.tags)) {
                        this.tags = [...data.tags]
                    }
                    this.url = data.url_src || ''
                })
                .catch(error => {
                    console.error('Error loading note as markdown:', error)
                    push.error({ message: this.$t('editor.loadOriginalError') })
                })
                .finally(() => { this.loading = false })
        },
        suggestedTitle () {
            const line = (this.markdown || '').split('\n').find(item => item.trim())
            if (!line) {
                return ''
            }
            return line.replace(/^[#>*\-\s]+/, '').trim().slice(0, 80)
        },
        addTag (name) {
            const value = String(name || '').trim()
            if (!value || this.tags.includes(value)) {
                return
            }
            this.tags.push(value)
            this.tagInput = ''
        },
        addTagFromInput () {
            this.addTag(this.tagInput)
        },
        removeTag (tag) {
            this.tags = this.tags.filter(item => item !== tag)
        },
        async loadLink () {
            const target = this.linkUrl.trim()
            if (!target || this.importing) {
                return
            }
            this.importing = true
            try {
                const article = await fetchDefuddleArticle(target)
                this.linkHtml = article.html
                if (!this.title.trim() && article.title) {
                    this.title = article.title
                }
                if (!this.url) {
                    this.url = target
                }
            } catch (e) {
                console.error('Defuddle import error:', e)
                push.error({ message: this.$t('editor.defuddleError') + ': ' + e.message })
            } finally {
                this.importing = false
            }
        },
        save () {
            if (!this.canSave || this.saving) {
                return
            }
            const isLink = this.mode === 'link'
            const title = this.title.trim() || this.suggestedTitle()
            const payload = {
                title,
                url: isLink ? this.linkUrl.trim() : this.url,
                uuid: this.uuid,
                type: this.type || 'text',
                tags: this.tags,
                content: isLink ? this.linkHtml : this.markdown,
                format: isLink ? 'html' : 'markdown'
            }

            this.saving = true
            const toast = push.info({ message: this.$t('editor.saving'), duration: Infinity })
            fetch(this.apiFolder + '/note_edit.json', { method: 'POST', body: JSON.stringify(payload) })
                .then(response => response.json())
                .then(data => {
                    toast.clear()
                    this.discardDraft()
                    this.noteStore.getAllData()
                    if (this.isNew) {
                        this.$router.replace('/notes/' + data.NoteBookUUID + '/' + data.uuid + '/')
                    } else {
                        this.$router.back()
                    }
                })
                .catch(error => {
                    toast.clear()
                    console.error('Error saving note:', error)
                    push.error({ message: this.$t('editor.errorSave') })
                })
                .finally(() => { this.saving = false })
        },
        cancel () {
            if (this.isNew) {
                if (this.title.trim() || this.markdown.trim()) {
                    this.saveDraft()
                    push.info({ message: this.$t('editor.draftSaved') })
                } else {
                    this.discardDraft()
                }
                this.$router.push('/notes/' + (this.noteStore.currentNotebookID || 'Allnotes') + '/')
                return
            }
            this.$router.back()
        },
        queueDraft () {
            if (!this.isNew) {
                return
            }
            clearTimeout(this.draftTimer)
            this.draftTimer = setTimeout(() => this.saveDraft(), 800)
        },
        saveDraft () {
            if (!this.isNew) {
                return
            }
            if (!this.title.trim() && !this.markdown.trim()) {
                localStorage.removeItem(DRAFT_KEY)
                return
            }
            localStorage.setItem(DRAFT_KEY, JSON.stringify({
                title: this.title,
                markdown: this.markdown,
                tags: this.tags,
                url: this.url
            }))
        },
        restoreDraft () {
            let raw = null
            try {
                raw = localStorage.getItem(DRAFT_KEY)
            } catch (e) {
                return
            }
            if (!raw) {
                return
            }
            try {
                const draft = JSON.parse(raw)
                this.title = draft.title || ''
                this.markdown = draft.markdown || ''
                this.tags = Array.isArray(draft.tags) ? draft.tags : []
                this.url = draft.url || ''
                this.draftRestored = true
            } catch (e) {
                localStorage.removeItem(DRAFT_KEY)
            }
        },
        clearDraft () {
            this.title = ''
            this.markdown = ''
            this.tags = []
            this.url = ''
            this.tagInput = ''
            this.discardDraft()
        },
        discardDraft () {
            clearTimeout(this.draftTimer)
            try {
                localStorage.removeItem(DRAFT_KEY)
            } catch (e) {
                // private mode: nothing to do
            }
            this.draftRestored = false
        }
    }
}
</script>

<style scoped>
.med {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--m-bg);
}

.med__bar {
    display: flex;
    flex: 0 0 auto;
    align-items: center;
    gap: .35rem;
    padding: .35rem .5rem;
    border-bottom: 1px solid var(--m-border);
    background: var(--m-bg);
}

.med__bar-title {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    color: var(--m-text);
    font-size: .92rem;
    font-weight: 600;
    text-align: center;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.med__bar-btn {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    min-width: var(--m-tap);
    height: var(--m-tap);
    padding: 0 .35rem;
    border: 0;
    border-radius: var(--m-radius-sm);
    background: none;
    color: var(--m-accent);
    font-size: .92rem;
}

.med__bar-btn--icon {
    font-size: 1.1rem;
}

.med__bar-btn--on {
    background: var(--m-accent-soft);
}

.med__bar-btn--save {
    color: #fff;
    background: var(--m-accent);
    font-size: 1.15rem;
}

.med__bar-btn--save:disabled {
    opacity: .45;
}

.med__meta {
    flex: 0 0 auto;
    padding: .55rem .75rem .25rem;
    border-bottom: 1px solid var(--m-surface-2);
}

.med__title-input {
    width: 100%;
    margin-bottom: .45rem;
    padding: .35rem 0;
    border: 0;
    border-bottom: 1px solid var(--m-border);
    outline: none;
    background: none;
    color: var(--m-text);
    font-size: 1.15rem;
    font-weight: 600;
}

.med__tags {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: .3rem;
    min-height: 2rem;
}

.med__tag {
    display: inline-flex;
    align-items: center;
    gap: .1rem;
    padding: .15rem .2rem .15rem .45rem;
    border-radius: var(--m-radius-sm);
    background: var(--m-chip-bg);
    color: var(--m-chip-text);
    font-size: .78rem;
}

.med__tag-x {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 1.15rem;
    height: 1.15rem;
    border: 0;
    background: none;
    color: inherit;
    font-size: .8rem;
}

.med__tag-input {
    flex: 1 1 6rem;
    min-width: 5rem;
    padding: .2rem 0;
    border: 0;
    outline: none;
    background: none;
    color: var(--m-text);
    font-size: .85rem;
}

.med__suggest {
    display: flex;
    gap: .35rem;
    padding: .35rem 0 .25rem;
    overflow-x: auto;
    scrollbar-width: none;
}

.med__suggest::-webkit-scrollbar {
    display: none;
}

.med__suggest button {
    flex: 0 0 auto;
    padding: .3rem .55rem;
    border: 1px solid var(--m-border);
    border-radius: 999px;
    background: var(--m-bg);
    color: var(--m-text-muted);
    font-size: .8rem;
    white-space: nowrap;
}

.med__modes {
    display: inline-flex;
    gap: .2rem;
    margin: .45rem 0 .35rem;
    padding: .18rem;
    border-radius: var(--m-radius);
    background: var(--m-surface);
}

.med__modes button {
    padding: .35rem .9rem;
    border: 0;
    border-radius: var(--m-radius-sm);
    background: none;
    color: var(--m-text-muted);
    font-size: .85rem;
    font-weight: 600;
}

.med__modes .med__mode--active {
    background: var(--m-bg);
    color: var(--m-text);
    box-shadow: 0 1px 2px rgba(46, 52, 64, .12);
}

.med__draft {
    display: flex;
    align-items: center;
    gap: .5rem;
    padding: .3rem 0 .15rem;
    color: var(--m-text-muted);
    font-size: .78rem;
}

.med__draft button {
    padding: 0;
    border: 0;
    background: none;
    color: var(--m-danger);
    font-size: .78rem;
    text-decoration: underline;
}

.med__body {
    position: relative;
    flex: 1 1 auto;
    min-height: 0;
}

.med__body--link {
    padding: .75rem;
    overflow-y: auto;
}

.med__preview {
    height: 100%;
    padding: .75rem 1rem 2rem;
    overflow-y: auto;
}

.med__link-row {
    display: flex;
    gap: .5rem;
}

.med__link-input {
    flex: 1 1 auto;
    min-width: 0;
    padding: .55rem .65rem;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius-sm);
    outline: none;
    background: var(--m-bg);
    color: var(--m-text);
    font-size: .9rem;
}

.med__link-input:focus {
    border-color: var(--m-accent);
}

.med__load {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    width: var(--m-tap);
    height: var(--m-tap);
    border: 0;
    border-radius: var(--m-radius-sm);
    background: var(--m-accent);
    color: #fff;
    font-size: 1.05rem;
}

.med__load:disabled {
    opacity: .45;
}

.med__hint {
    margin: .5rem 0 .25rem;
    color: var(--m-text-muted);
    font-size: .8rem;
    line-height: 1.4;
}

.med__loading {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, .7);
}

.med__spinner {
    display: inline-block;
    width: 1rem;
    height: 1rem;
    border: 2px solid currentColor;
    border-top-color: transparent;
    border-radius: 50%;
    animation: med-spin .7s linear infinite;
}

.med__spinner--big {
    width: 1.75rem;
    height: 1.75rem;
    color: var(--m-accent);
}

@keyframes med-spin {
    to { transform: rotate(360deg); }
}
</style>
