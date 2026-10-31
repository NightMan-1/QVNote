<template>
    <div class="mshell">
        <header class="mshell__bar">
            <button v-if="focusMode" class="mshell__btn" :aria-label="$t('articleList.backToList')" @click="backFromFocus">
                <i class="bi bi-arrow-left"></i>
            </button>

            <button v-else-if="activeTab === 'notes'" class="mshell__library" @click="libraryOpen = true">
                <span class="mshell__library-name">{{ pageTitle }}</span>
                <i class="bi bi-chevron-down"></i>
            </button>

            <QvBrand v-else-if="activeTab === 'overview'" />

            <div v-else class="mshell__title">{{ pageTitle }}</div>

            <div class="mshell__grow"></div>

            <template v-if="focusMode && isNoteRoute && article.uuid">
                <a
                    v-if="article.url_src"
                    class="mshell__btn"
                    :href="article.url_src"
                    target="_blank"
                    rel="noopener"
                ><i class="bi bi-box-arrow-up-right"></i></a>
                <button class="mshell__btn" :aria-label="$t('editor.titleExist')" @click="editNote">
                    <i class="bi bi-pencil"></i>
                </button>
                <button class="mshell__btn" :aria-label="$t('general.actions')" @click="actionSheet = 'actions'">
                    <i class="bi bi-three-dots-vertical"></i>
                </button>
            </template>
            <button v-else-if="activeTab === 'search'" class="mshell__btn" @click="closeSearch">
                <i class="bi bi-x-lg"></i>
            </button>
        </header>

        <main class="mshell__content">
            <qv-editor v-if="pageType === 'editor'" />

            <NoteArticle v-else-if="isNoteRoute && article.uuid" :article="article" />

            <div v-else-if="activeTab === 'search'" class="mshell__search">
                <input
                    ref="searchField"
                    v-model="searchText"
                    type="search"
                    class="form-control"
                    :placeholder="$t('articleList.searchPlaceholder')"
                >
                <div class="mshell__list">
                    <div v-if="searchText.trim().length < 3" class="mshell__hint">{{ $t('articleList.searchHint') }}</div>
                    <div v-else-if="searchResults.length === 0" class="mshell__hint">{{ $t('articleList.searchNothing') }}</div>
                    <div v-else class="mshell__hint">{{ $t('articleList.searchFound', { count: searchResults.length }) }}</div>
                    <button
                        v-for="note in searchResults"
                        :key="note.uuid"
                        class="mshell__item"
                        @click="openNote(note.uuid, note.NoteBookUUID)"
                    >{{ note.title }}</button>
                </div>
            </div>

            <div v-else-if="activeTab === 'more'" class="mshell__list">
                <button class="mshell__item mshell__item--action" @click="selectTab('overview')">
                    <i class="bi bi-bar-chart me-2"></i>{{ $t('dashboard.infoTitle') }}
                </button>
                <button class="mshell__item mshell__item--action" @click="openSettings">
                    <i class="bi bi-gear me-2"></i>{{ $t('general.buttonSettings') }}
                </button>
                <button v-if="canInstall" class="mshell__item mshell__item--action" @click="installApp">
                    <i class="bi bi-download me-2"></i>{{ $t('general.installApp') }}
                </button>
            </div>

            <qv-dashboard v-else-if="activeTab === 'overview'" />

            <div v-else class="mshell__list">
                <div v-if="notes.length === 0" class="mshell__hint">{{ $t('articleList.emptyNotebook') }}</div>
                <button
                    v-for="note in notes"
                    :key="note.uuid"
                    class="mshell__item"
                    @click="openNote(note.uuid, note.NoteBookUUID)"
                >{{ note.title }}</button>
            </div>
        </main>

        <nav v-if="!focusMode" class="mshell__tabs">
            <button
                v-for="tab in tabs"
                :key="tab.id"
                class="mshell__tab"
                :class="{ 'mshell__tab--active': activeTab === tab.id }"
                @click="selectTab(tab.id)"
            >
                <i class="bi" :class="tab.icon"></i>
                <span>{{ tab.label }}</span>
            </button>
        </nav>

        <div v-if="actionSheet" class="mshell__backdrop" @click="actionSheet = null"></div>
        <div v-if="actionSheet" class="mshell__sheet mshell__sheet--bottom">
            <header class="mshell__sheet-bar">
                <div class="mshell__title">{{ actionSheetTitle }}</div>
                <div class="mshell__grow"></div>
                <button class="mshell__btn" @click="actionSheet = null"><i class="bi bi-x-lg"></i></button>
            </header>
            <div class="mshell__sheet-list">
                <template v-if="actionSheet === 'actions'">
                    <button class="mshell__item mshell__item--action" @click="toggleFavorite">
                        <i class="bi me-2" :class="article.favorites ? 'bi-star-fill text-warning' : 'bi-star'"></i>
                        {{ article.favorites ? $t('articleList.btnRemoveFavorites') : $t('articleList.btnFavorites') }}
                    </button>
                    <button class="mshell__item mshell__item--action" @click="actionSheet = 'move'">
                        <i class="bi bi-arrow-left-right me-2"></i>{{ $t('articleList.btnMove') }}
                    </button>
                    <button class="mshell__item mshell__item--action text-danger" @click="actionSheet = 'delete'">
                        <i class="bi bi-trash me-2"></i>{{ $t('articleList.btnDelete') }}
                    </button>
                </template>

                <template v-else-if="actionSheet === 'move'">
                    <div class="mshell__group">{{ $t('articleList.modalMoveNotebook') }}</div>
                    <button
                        v-for="notebook in notebooks"
                        :key="notebook.uuid"
                        class="mshell__item"
                        :class="{ 'fw-semibold': notebook.uuid === currentNotebookID }"
                        @click="moveTo(notebook.uuid)"
                    >{{ notebook.name }}</button>
                </template>

                <template v-else-if="actionSheet === 'delete'">
                    <div class="mshell__group">{{ $t('articleList.modalDeleteTitle') }}</div>
                    <button class="mshell__item mshell__item--action text-danger" @click="deleteNote">
                        <i class="bi bi-trash me-2"></i>{{ $t('general.yesBig') }}
                    </button>
                    <button class="mshell__item mshell__item--action" @click="actionSheet = 'actions'">
                        {{ $t('general.noBig') }}
                    </button>
                </template>
            </div>
        </div>

        <div v-if="libraryOpen" class="mshell__sheet">
            <header class="mshell__sheet-bar">
                <div class="mshell__title">{{ $t('general.libraryTitle') }}</div>
                <div class="mshell__grow"></div>
                <button class="mshell__btn" @click="libraryOpen = false"><i class="bi bi-x-lg"></i></button>
            </header>

            <div class="mshell__sheet-switch">
                <button
                    class="btn"
                    :class="libraryType === 'notebooks' ? 'btn-primary' : 'btn-outline-primary'"
                    @click="libraryType = 'notebooks'"
                >{{ $t('general.sidebarSwitchNotebooks') }}</button>
                <button
                    class="btn"
                    :class="libraryType === 'tags' ? 'btn-primary' : 'btn-outline-primary'"
                    @click="libraryType = 'tags'"
                >{{ $t('general.sidebarSwitchTags') }}</button>
            </div>

            <div class="mshell__sheet-list">
                <template v-if="libraryType === 'notebooks'">
                    <button class="mshell__item mshell__item--row" @click="openNotebook('Allnotes')">
                        <span><i class="bi bi-collection me-2"></i>{{ $t('general.sidebarAllNotes') }}</span>
                        <span class="badge bg-primary">{{ notesCountTotal }}</span>
                    </button>
                    <button class="mshell__item mshell__item--row" @click="openNotebook('Inbox')">
                        <span><i class="bi bi-inbox me-2"></i>{{ $t('general.sidebarInbox') }}</span>
                        <span class="badge bg-primary">{{ notesCountInbox }}</span>
                    </button>
                    <button class="mshell__item mshell__item--row" @click="openNotebook('Favorites')">
                        <span><i class="bi bi-star me-2"></i>{{ $t('general.sidebarFavorites') }}</span>
                        <span class="badge bg-primary">{{ notesCountFavorites }}</span>
                    </button>
                    <button class="mshell__item mshell__item--row" @click="openNotebook('Trash')">
                        <span><i class="bi bi-trash me-2"></i>{{ $t('general.sidebarTrash') }}</span>
                        <span class="badge bg-primary">{{ notesCountTrash }}</span>
                    </button>
                    <div class="mshell__group">{{ $t('general.sidebarNotebooks') }}</div>
                    <button
                        v-for="notebook in regularNotebooks"
                        :key="notebook.uuid"
                        class="mshell__item mshell__item--row"
                        @click="openNotebook(notebook.uuid)"
                    >
                        <span>{{ notebook.name }}</span>
                        <span class="badge bg-primary">{{ notebook.notesCount }}</span>
                    </button>
                </template>
                <template v-else>
                    <button
                        v-for="tag in tags"
                        :key="tag.url"
                        class="mshell__item mshell__item--row"
                        @click="openTag(tag.url)"
                    >
                        <span>#{{ tag.name }}</span>
                        <span class="badge bg-primary">{{ tag.count }}</span>
                    </button>
                </template>
            </div>
        </div>
    </div>
</template>

<script>
import { defineAsyncComponent } from 'vue'
import { useNoteStore } from '../store'
import NoteArticle from './NoteArticle.vue'
import QvBrand from './QvBrand.vue'

const qvDashboard = defineAsyncComponent(() => import('./qvDashboard.vue'))
const qvEditor = defineAsyncComponent(() => import('./qvEditor.vue'))

export default {
    name: 'MobileShell',
    components: { NoteArticle, qvDashboard, qvEditor, QvBrand },
    data () {
        return {
            actionSheet: null,
            installPrompt: null,
            libraryOpen: false,
            libraryType: 'notebooks',
            searchText: '',
            searchResults: [],
            searchTimer: null
        }
    },
    setup () {
        return { noteStore: useNoteStore() }
    },
    computed: {
        activeTab: {
            get () { return this.noteStore.mobileTab },
            set (value) { this.noteStore.setMobileTab(value) }
        },
        canInstall () {
            return !!this.installPrompt && !window.matchMedia('(display-mode: standalone)').matches
        },
        tabs () {
            return [
                { id: 'overview', icon: 'bi-bar-chart', label: this.$t('general.tabOverview') },
                { id: 'notes', icon: 'bi-journal-text', label: this.$t('general.tabNotes') },
                { id: 'search', icon: 'bi-search', label: this.$t('general.tabSearch') },
                { id: 'more', icon: 'bi-three-dots', label: this.$t('general.tabMore') }
            ]
        },
        pageType () { return this.noteStore.pageType },
        article () { return this.noteStore.currentArticle || {} },
        notes () {
            const list = this.noteStore.notesList
            return Array.isArray(list) ? list : []
        },
        notebooks () {
            const list = this.noteStore.notebooksList
            return Array.isArray(list) ? list : []
        },
        regularNotebooks () {
            return this.notebooks.filter(n => n.name !== 'Inbox' && n.name !== 'Trash')
        },
        tags () {
            const list = this.noteStore.tagsList
            return Array.isArray(list) ? list : []
        },
        notesCountTotal () { return this.noteStore.notesCountTotal },
        notesCountInbox () { return this.noteStore.notesCountInbox },
        notesCountFavorites () { return this.noteStore.notesCountFavorites },
        notesCountTrash () { return this.noteStore.notesCountTrash },
        currentNotebookID () { return this.noteStore.currentNotebookID },
        currentTagURL () { return this.noteStore.currentTagURL },
        isNoteRoute () { return this.$route.name === 'qvNote' || this.$route.name === 'qvTag' },
        actionSheetTitle () {
            if (this.actionSheet === 'move') {
                return this.$t('articleList.btnMove')
            }
            if (this.actionSheet === 'delete') {
                return this.$t('articleList.btnDelete')
            }
            return this.$t('general.actions')
        },
        focusMode () { return this.isNoteRoute || this.pageType === 'editor' },
        pageTitle () {
            if (this.activeTab === 'overview') {
                return this.$t('dashboard.infoTitle')
            }
            if (this.activeTab === 'search') {
                return this.$t('general.tabSearch')
            }
            if (this.activeTab === 'more') {
                return this.$t('general.tabMore')
            }
            if (this.isNoteRoute && this.$route.name === 'qvTag') {
                return this.currentTagURL || this.$t('general.sidebarTags')
            }
            const id = this.currentNotebookID
            const special = {
                Inbox: 'general.sidebarInbox',
                Favorites: 'general.sidebarFavorites',
                Trash: 'general.sidebarTrash',
                Allnotes: 'general.sidebarAllNotes'
            }
            if (special[id]) {
                return this.$t(special[id])
            }
            const current = this.notebooks.find(n => n.uuid === id)
            return current ? current.name : this.$t('general.sidebarAllNotes')
        }
    },
    watch: {
        searchText () {
            this.runSearch()
        },
        '$route' () {
            this.libraryOpen = false
            this.actionSheet = null
            this.syncTabWithRoute()
        },
        pageType () {
            this.syncTabWithRoute()
        }
    },
    mounted () {
        // deep links to a note/tag should open the notes tab; on the root
        // route keep the last tab (so Settings returns to "More")
        if (this.$route.name !== 'qvApp') {
            this.syncTabWithRoute()
        }
        window.addEventListener('beforeinstallprompt', this.onBeforeInstallPrompt)
    },
    beforeUnmount () {
        window.removeEventListener('beforeinstallprompt', this.onBeforeInstallPrompt)
    },
    methods: {
        onBeforeInstallPrompt (event) {
            event.preventDefault()
            this.installPrompt = event
        },
        installApp () {
            const prompt = this.installPrompt
            if (!prompt) {
                return
            }
            this.installPrompt = null
            prompt.prompt()
        },
        syncTabWithRoute () {
            if (this.$route.name === 'qvApp') {
                this.activeTab = 'overview'
            } else if (this.$route.name === 'qvNote' || this.$route.name === 'qvNotebooks' ||
                this.$route.name === 'qvTag' || this.$route.name === 'qvTagsList' ||
                this.$route.name === 'qvEditor') {
                this.activeTab = 'notes'
            }
        },
        selectTab (id) {
            this.activeTab = id
            if (id === 'overview') {
                if (this.$route.name !== 'qvApp') {
                    this.$router.push('/')
                }
                return
            }
            if (id === 'notes') {
                if (this.$route.name !== 'qvNote' && this.$route.name !== 'qvNotebooks' &&
                    this.$route.name !== 'qvTag' && this.$route.name !== 'qvTagsList') {
                    this.$router.push('/notes/' + (this.currentNotebookID || 'Allnotes') + '/')
                }
                return
            }
            if (id === 'search') {
                this.$nextTick(() => this.$refs.searchField && this.$refs.searchField.focus())
            }
        },
        openNotebook (uuid) {
            this.libraryOpen = false
            this.activeTab = 'notes'
            this.$router.push('/notes/' + uuid + '/')
        },
        openTag (url) {
            this.libraryOpen = false
            this.activeTab = 'notes'
            this.$router.push('/tags/' + url + '/')
        },
        openNote (uuid, notebookUUID) {
            const nb = notebookUUID || this.currentNotebookID
            this.searchText = ''
            this.searchResults = []
            this.activeTab = 'notes'
            this.$router.push('/notes/' + nb + '/' + uuid + '/')
        },
        backFromFocus () {
            if (this.pageType === 'editor' && this.article.uuid) {
                this.$router.push('/notes/' + this.article.NoteBookUUID + '/' + this.article.uuid + '/')
                return
            }
            if (this.isNoteRoute && this.$route.name === 'qvTag') {
                this.$router.push('/tags/' + this.$route.params.nbUUID + '/')
                return
            }
            this.$router.push('/notes/' + (this.currentNotebookID || 'Allnotes') + '/')
        },
        editNote () {
            this.$router.push({ name: 'qvEditor' })
        },
        toggleFavorite () {
            const uuid = this.article.uuid
            const add = !this.article.favorites
            fetch(this.noteStore.apiFolder + '/favorites.json', {
                method: 'POST',
                body: JSON.stringify({ action: add ? 'add' : 'remove', UUID: uuid })
            })
                .then(() => {
                    this.noteStore.setFavoritesStatus(add)
                    this.noteStore.getFavoritesCount()
                })
                .catch(error => console.error('Favorites error:', error))
            this.actionSheet = null
        },
        moveTo (target) {
            const uuid = this.article.uuid
            fetch(this.noteStore.apiFolder + '/note_move.json', {
                method: 'POST',
                body: JSON.stringify({ action: 'move', uuid, target })
            })
                .then(() => {
                    this.noteStore.getAllData()
                    this.actionSheet = null
                    this.$router.push('/notes/' + target + '/' + uuid + '/')
                })
                .catch(error => console.error('Move error:', error))
        },
        deleteNote () {
            const uuid = this.article.uuid
            const route = this.$route
            // the article object can be stale after a move, so prefer the
            // notebook from the current route
            const notebook = route.params.nbUUID || this.currentNotebookID
            fetch(this.noteStore.apiFolder + '/note_move.json', {
                method: 'POST',
                body: JSON.stringify({ action: 'delete', uuid })
            })
                .then(() => {
                    this.noteStore.getAllData()
                    this.actionSheet = null
                    if (route.name === 'qvTag') {
                        this.$router.push('/tags/' + route.params.nbUUID + '/')
                    } else {
                        this.$router.push('/notes/' + notebook + '/')
                    }
                })
                .catch(error => console.error('Delete error:', error))
        },
        openSettings () {
            this.$router.push('/settings/')
        },
        closeSearch () {
            this.activeTab = 'notes'
            this.searchText = ''
            this.searchResults = []
        },
        runSearch () {
            if (this.searchTimer) {
                clearTimeout(this.searchTimer)
            }
            const text = this.searchText.trim()
            if (text.length < 3) {
                this.searchResults = []
                return
            }
            this.searchTimer = setTimeout(() => {
                fetch(this.noteStore.apiFolder + '/search.json', {
                    method: 'POST',
                    body: JSON.stringify({ text })
                })
                    .then(response => response.json())
                    .then(data => {
                        if (text === this.searchText.trim()) {
                            this.searchResults = Array.isArray(data) ? data : []
                        }
                    })
                    .catch(error => {
                        console.error('Search error:', error)
                        this.searchResults = []
                    })
            }, 250)
        }
    }
}
</script>

<style scoped>
.mshell {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: #fff;
}

.mshell__bar {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: .35rem;
    min-height: 3.2rem;
    padding: .25rem .75rem;
    background: #fff;
    border-bottom: 1px solid var(--nord4);
}

.mshell__grow {
    flex: 1 1 auto;
}

.mshell__title {
    font-weight: 600;
    font-size: 1.05rem;
    color: var(--nord0);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.mshell__library {
    display: flex;
    align-items: center;
    gap: .3rem;
    max-width: 75%;
    padding: .35rem .5rem;
    border: 0;
    border-radius: .35rem;
    background: none;
    color: var(--nord0);
    font-size: 1.05rem;
    font-weight: 600;
}

.mshell__library-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.mshell__btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.5rem;
    height: 2.5rem;
    padding: 0;
    border: 1px solid var(--nord4);
    border-radius: .35rem;
    background: #fff;
    color: var(--nord3);
    font-size: 1.1rem;
    text-decoration: none;
}

.mshell__btn:active {
    background: var(--nord5);
}

.mshell__content {
    flex: 1 1 auto;
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
}

.mshell__list {
    display: flex;
    flex-direction: column;
}

.mshell__item {
    display: block;
    width: 100%;
    padding: .8rem 1rem;
    border: 0;
    border-bottom: 1px solid var(--nord5);
    background: none;
    color: var(--grey-700);
    font-size: 1rem;
    line-height: 1.35;
    text-align: left;
}

.mshell__item:active {
    background: var(--nord5);
}

.mshell__item--row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: .75rem;
}

.mshell__item--action {
    display: flex;
    align-items: center;
}

.mshell__hint {
    padding: 1.25rem 1rem;
    color: var(--nord3);
}

.mshell__group {
    padding: .9rem 1rem .35rem;
    color: var(--nord3);
    font-size: .8rem;
    text-transform: uppercase;
    letter-spacing: .03em;
}

.mshell__search {
    padding: .75rem .75rem 0;
}

.mshell__tabs {
    flex: 0 0 auto;
    display: flex;
    border-top: 1px solid var(--nord4);
    background: #fff;
    padding-bottom: env(safe-area-inset-bottom);
}

.mshell__tab {
    flex: 1 1 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: .1rem;
    padding: .45rem .25rem .35rem;
    border: 0;
    background: none;
    color: var(--nord3);
    font-size: .72rem;
}

.mshell__tab i {
    font-size: 1.25rem;
}

.mshell__tab--active {
    color: var(--nord10);
}

/* two classes so it wins over the base .mshell__sheet rule below */
.mshell__sheet.mshell__sheet--bottom {
    inset: auto 0 0 0;
    height: auto;
    max-height: 70vh;
    border-radius: 1rem 1rem 0 0;
    box-shadow: 0 -8px 24px rgba(0, 0, 0, .18);
}

.mshell__sheet {
    position: fixed;
    inset: 0;
    z-index: 1200;
    display: flex;
    flex-direction: column;
    background: #fff;
}

.mshell__sheet-bar {
    display: flex;
    align-items: center;
    gap: .5rem;
    padding: .4rem .5rem;
    border-bottom: 1px solid var(--nord4);
}

.mshell__sheet-switch {
    display: flex;
    gap: .4rem;
    padding: .6rem .75rem;
    background: var(--nord6);
}

.mshell__sheet-switch .btn {
    flex: 1 1 0;
}

.mshell__sheet-list {
    flex: 1 1 auto;
    overflow-y: auto;
}
</style>
