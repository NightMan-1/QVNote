<template>
    <div class="mshell" :style="{ '--mshell-tabs': tabsHeight + 'px' }">
        <header class="mshell__bar">
            <div v-if="focusMode" class="mshell__lead">
                <button class="mshell__btn" :aria-label="$t('articleList.backToList')" @click="backFromFocus">
                    <i class="bi bi-arrow-left"></i>
                </button>
                <div v-if="isNoteRoute" class="mshell__crumb">
                    <i class="bi" :class="sectionIcon"></i>
                    <span class="mshell__ellipsis">{{ pageTitle }}</span>
                </div>
            </div>

            <button v-else-if="activeTab === 'notes'" class="mshell__library" @click="showLibrary">
                <i class="bi" :class="sectionIcon"></i>
                <span class="mshell__ellipsis">{{ pageTitle }}</span>
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

            <div v-else-if="activeTab === 'search'" class="mshell__search">
                <input
                    ref="searchField"
                    v-model="searchText"
                    type="search"
                    class="mshell__search-field"
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

            <MobileSettings v-else-if="activeTab === 'settings'" />

            <div v-else-if="activeTab === 'more'" class="mshell__list">
                <button class="mshell__item mshell__item--action" @click="openEditor">
                    <i class="bi bi-pencil"></i>{{ $t('general.addNewNote') }}
                </button>
                <button class="mshell__item mshell__item--action" @click="openSettings">
                    <i class="bi bi-gear"></i>{{ $t('general.buttonSettings') }}
                </button>
                <button v-if="canInstall" class="mshell__item mshell__item--action" @click="installApp">
                    <i class="bi bi-download"></i>{{ $t('general.installApp') }}
                </button>
            </div>

            <MobileDashboard v-else-if="activeTab === 'overview'" />

            <MobileNote v-else-if="isNoteRoute && article.uuid" :article="article" @tag-click="openTagByName" />

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

        <nav v-if="pageType !== 'editor'" ref="tabs" class="mshell__tabs">
            <button
                v-for="tab in tabs"
                :key="tab.id"
                class="mshell__tab"
                :class="{ 'mshell__tab--active': tabActive(tab.id) }"
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
                        <i class="bi" :class="article.favorites ? 'bi-star-fill text-warning' : 'bi-star'"></i>
                        {{ article.favorites ? $t('articleList.btnRemoveFavorites') : $t('articleList.btnFavorites') }}
                    </button>
                    <button class="mshell__item mshell__item--action" @click="actionSheet = 'move'">
                        <i class="bi bi-arrow-left-right"></i>{{ $t('articleList.btnMove') }}
                    </button>
                    <button class="mshell__item mshell__item--action mshell__danger" @click="actionSheet = 'delete'">
                        <i class="bi bi-trash"></i>{{ $t('articleList.btnDelete') }}
                    </button>
                </template>

                <template v-else-if="actionSheet === 'move'">
                    <div class="mshell__group">{{ $t('articleList.modalMoveNotebook') }}</div>
                    <button
                        v-for="notebook in notebooks"
                        :key="notebook.uuid"
                        class="mshell__item"
                        :class="{ 'mshell__item--selected': notebook.uuid === currentNotebookID }"
                        @click="moveTo(notebook.uuid)"
                    >{{ notebook.name }}</button>
                </template>

                <template v-else-if="actionSheet === 'delete'">
                    <div class="mshell__group">{{ $t('articleList.modalDeleteTitle') }}</div>
                    <button class="mshell__item mshell__item--action mshell__danger" @click="deleteNote">
                        <i class="bi bi-trash"></i>{{ $t('general.yesBig') }}
                    </button>
                    <button class="mshell__item mshell__item--action" @click="actionSheet = 'actions'">
                        {{ $t('general.noBig') }}
                    </button>
                </template>
            </div>
        </div>

        <div
            v-if="libraryOpen"
            class="mshell__backdrop mshell__backdrop--above-tabs"
            @click="libraryOpen = false"
        ></div>
        <div v-if="libraryOpen" class="mshell__sheet mshell__sheet--library">
            <div class="mshell__grip"></div>

            <header class="mshell__sheet-bar">
                <div class="mshell__sheet-title">{{ $t('general.libraryTitle') }}</div>
                <div class="mshell__grow"></div>
                <button class="mshell__btn" @click="libraryOpen = false"><i class="bi bi-x-lg"></i></button>
            </header>

            <div class="mshell__sheet-switch">
                <button
                    class="mshell__seg"
                    :class="{ 'mshell__seg--active': libraryType === 'notebooks' }"
                    @click="libraryType = 'notebooks'"
                >{{ $t('general.sidebarSwitchNotebooks') }}</button>
                <button
                    class="mshell__seg"
                    :class="{ 'mshell__seg--active': libraryType === 'tags' }"
                    @click="libraryType = 'tags'"
                >{{ $t('general.sidebarSwitchTags') }}</button>
            </div>

            <div ref="libraryList" class="mshell__sheet-list">
                <template v-if="libraryType === 'notebooks'">
                    <button
                        v-for="notebook in regularNotebooks"
                        :key="notebook.uuid"
                        class="mshell__item mshell__item--row mshell__item--library"
                        :class="{ 'mshell__item--current': notebook.uuid === currentNotebookID }"
                        @click="openNotebook(notebook.uuid)"
                    >
                        <span class="mshell__entry">
                            <i class="bi bi-journal-bookmark"></i>
                            <span class="mshell__ellipsis">{{ notebook.name }}</span>
                        </span>
                        <span class="mshell__item-count">{{ notebook.notesCount }}</span>
                    </button>
                    <div v-if="regularNotebooks.length === 0" class="mshell__hint">{{ $t('general.libraryEmpty') }}</div>
                </template>
                <template v-else>
                    <button
                        v-for="tag in tags"
                        :key="tag.url"
                        class="mshell__item mshell__item--row mshell__item--library"
                        :class="{ 'mshell__item--current': tagIsCurrent(tag) }"
                        @click="openTag(tag.name)"
                    >
                        <span class="mshell__entry">
                            <i class="bi bi-hash"></i>
                            <span class="mshell__ellipsis">{{ tag.name }}</span>
                        </span>
                        <span class="mshell__item-count">{{ tag.count }}</span>
                    </button>
                </template>
            </div>

            <footer v-if="libraryType === 'notebooks'" class="mshell__shortcuts">
                <button
                    class="mshell__shortcut"
                    :class="{ 'mshell__shortcut--active': currentNotebookID === 'Allnotes' }"
                    @click="openNotebook('Allnotes')"
                >
                    <i class="bi bi-collection"></i>
                    <span class="mshell__ellipsis">{{ $t('general.sidebarAllNotes') }}</span>
                    <span class="mshell__shortcut-count">{{ notesCountTotal }}</span>
                </button>
                <button
                    class="mshell__shortcut"
                    :class="{ 'mshell__shortcut--active': currentNotebookID === 'Inbox' }"
                    @click="openNotebook('Inbox')"
                >
                    <i class="bi bi-inbox"></i>
                    <span class="mshell__ellipsis">{{ $t('general.sidebarInbox') }}</span>
                    <span class="mshell__shortcut-count">{{ notesCountInbox }}</span>
                </button>
                <button
                    class="mshell__shortcut"
                    :class="{ 'mshell__shortcut--active': currentNotebookID === 'Favorites' }"
                    @click="openNotebook('Favorites')"
                >
                    <i class="bi bi-star"></i>
                    <span class="mshell__ellipsis">{{ $t('general.sidebarFavorites') }}</span>
                    <span class="mshell__shortcut-count">{{ notesCountFavorites }}</span>
                </button>
                <button
                    class="mshell__shortcut"
                    :class="{ 'mshell__shortcut--active': currentNotebookID === 'Trash' }"
                    @click="openNotebook('Trash')"
                >
                    <i class="bi bi-trash"></i>
                    <span class="mshell__ellipsis">{{ $t('general.sidebarTrash') }}</span>
                    <span class="mshell__shortcut-count">{{ notesCountTrash }}</span>
                </button>
            </footer>
        </div>
    </div>
</template>

<script>
import { defineAsyncComponent } from 'vue'
import { useNoteStore } from '../store'
import MobileNote from '../mobile/MobileNote.vue'
import QvBrand from './QvBrand.vue'

const MobileDashboard = defineAsyncComponent(() => import('../mobile/MobileDashboard.vue'))
const MobileSettings = defineAsyncComponent(() => import('../mobile/MobileSettings.vue'))
const qvEditor = defineAsyncComponent(() => import('./qvEditor.vue'))

export default {
    name: 'MobileShell',
    components: { MobileNote, MobileDashboard, MobileSettings, qvEditor, QvBrand },
    data () {
        return {
            actionSheet: null,
            installPrompt: null,
            libraryOpen: false,
            libraryType: 'notebooks',
            tabsHeight: 0,
            lastNotesTabTap: 0,
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
        isTagContext () {
            return this.$route.name === 'qvTag' || this.$route.name === 'qvTagsList'
        },
        // the icon in front of the current notebook/tag name, so a notebook
        // and a tag with the same name can be told apart
        sectionIcon () {
            if (this.isTagContext) {
                return 'bi-hash'
            }
            const icons = {
                Allnotes: 'bi-collection',
                Inbox: 'bi-inbox',
                Favorites: 'bi-star',
                Trash: 'bi-trash'
            }
            return icons[this.currentNotebookID] || 'bi-journal-bookmark'
        },
        focusMode () {
            // reading a note (or editing it) gets the compact header with the
            // note actions; other tabs show their own header instead
            return this.activeTab === 'notes' && (this.isNoteRoute || this.pageType === 'editor')
        },
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
            if (this.activeTab === 'settings') {
                return this.$t('general.buttonSettings')
            }
            if (this.isTagContext) {
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
        window.addEventListener('resize', this.measureTabs)
        this.measureTabs()
    },
    beforeUnmount () {
        window.removeEventListener('beforeinstallprompt', this.onBeforeInstallPrompt)
        window.removeEventListener('resize', this.measureTabs)
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
        tabActive (id) {
            return this.activeTab === id || (id === 'more' && this.activeTab === 'settings')
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
        measureTabs () {
            const tabs = this.$refs.tabs
            this.tabsHeight = tabs ? Math.round(tabs.getBoundingClientRect().height) : 0
        },
        selectTab (id) {
            if (id !== 'notes') {
                this.libraryOpen = false
            }
            if (id === 'search') {
                if (this.activeTab === 'search') {
                    this.closeSearch()
                } else {
                    this.activeTab = 'search'
                    this.$nextTick(() => this.$refs.searchField && this.$refs.searchField.focus())
                }
                return
            }
            if (id === 'notes') {
                this.toggleLibrary()
                return
            }

            this.activeTab = id
            if (id === 'overview') {
                if (this.$route.name !== 'qvApp') {
                    this.$router.push('/')
                }
            }
        },
        // The "Notes" tab toggles the notebook picker: the first tap opens it,
        // the next one closes it. Two fast taps are a shortcut to All Notes,
        // which is why the closing tap only counts after a short delay. The
        // sheet opens immediately, so the normal flow has no lag.
        toggleLibrary () {
            const now = Date.now()
            const isDoubleTap = now - this.lastNotesTabTap < 320
            this.lastNotesTabTap = now

            // two fast taps always land on All Notes, whatever was open before
            if (isDoubleTap) {
                this.openNotebook('Allnotes')
                return
            }

            if (this.libraryOpen) {
                this.libraryOpen = false
                return
            }

            this.openLibrary()
        },
        // the "Notes" tab opens the notebook picker instead of the plain list
        openLibrary () {
            const onNotesScreen = this.$route.name === 'qvNote' ||
                this.$route.name === 'qvNotebooks' ||
                this.$route.name === 'qvTag' ||
                this.$route.name === 'qvTagsList'

            this.activeTab = 'notes'

            if (onNotesScreen) {
                this.showLibrary()
                return
            }

            const navigation = this.$router.push('/notes/' + (this.currentNotebookID || 'Allnotes') + '/')
            if (navigation && navigation.then) {
                navigation.then(() => {
                    this.$nextTick(() => {
                        this.showLibrary()
                    })
                }).catch(() => {})
            } else {
                this.$nextTick(() => {
                    this.showLibrary()
                })
            }
        },
        showLibrary () {
            this.libraryType = this.isTagContext ? 'tags' : 'notebooks'
            this.measureTabs()
            this.libraryOpen = true
            this.$nextTick(() => {
                const list = this.$refs.libraryList
                const current = list && list.querySelector('.mshell__item--current')
                if (current && current.scrollIntoView) {
                    current.scrollIntoView({ block: 'center' })
                }
            })
        },
        openNotebook (uuid) {
            const alreadyHere = this.$route.name === 'qvNotebooks' &&
                this.$route.params.nbUUID === uuid
            this.libraryOpen = false
            this.activeTab = 'notes'
            if (!alreadyHere) {
                this.$router.push('/notes/' + uuid + '/')
            }
        },
        openTag (name) {
            this.libraryOpen = false
            this.activeTab = 'notes'
            this.$router.push({ name: 'qvTagsList', params: { nbUUID: name } })
        },
        // a tag chip in the note view
        openTagByName (name) {
            this.openTag(name)
        },
        safeDecode (value) {
            try {
                return decodeURIComponent(value)
            } catch (e) {
                return value
            }
        },
        tagIsCurrent (tag) {
            return this.isTagContext && this.safeDecode(tag.url) === this.currentTagURL
        },
        openNote (uuid, notebookUUID) {
            this.searchText = ''
            this.searchResults = []
            this.activeTab = 'notes'

            // inside a tag keep the tag in the url, so "back" returns to the
            // tagged list and the header keeps showing the tag
            if (this.isTagContext && this.currentTagURL) {
                this.$router.push({ name: 'qvTag', params: { nbUUID: this.currentTagURL, noteUUID: uuid } })
                return
            }

            const nb = notebookUUID || this.currentNotebookID
            this.$router.push('/notes/' + nb + '/' + uuid + '/')
        },
        backFromFocus () {
            if (this.pageType === 'editor' && this.article.uuid) {
                this.$router.push('/notes/' + this.article.NoteBookUUID + '/' + this.article.uuid + '/')
                return
            }
            if (this.isNoteRoute && this.$route.name === 'qvTag') {
                this.$router.push({ name: 'qvTagsList', params: { nbUUID: this.currentTagURL || this.$route.params.nbUUID } })
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
            this.activeTab = 'settings'
        },
        openEditor () {
            this.noteStore.doEmptyCurrentArticle()
            this.noteStore.setCurrentNotebookID('')
            this.noteStore.setPageType('editor')
            this.$router.push({ name: 'qvEditor' })
        },
        closeSearch () {
            this.searchText = ''
            this.searchResults = []
            this.activeTab = 'notes'

            // the notes tab renders whatever qvApp loaded for the current
            // route, so make sure a list route is active (All Notes by
            // default) — otherwise the list stays empty
            const onNotesRoute = this.$route.name === 'qvNote' ||
                this.$route.name === 'qvNotebooks' ||
                this.$route.name === 'qvTag' ||
                this.$route.name === 'qvTagsList'
            if (!onNotesRoute) {
                this.$router.push('/notes/' + (this.currentNotebookID || 'Allnotes') + '/')
            }
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
    background: var(--m-bg);
}

.mshell__bar {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: .35rem;
    min-height: 3.2rem;
    padding: .25rem .75rem;
    background: var(--m-bg);
    border-bottom: 1px solid var(--m-border);
}

.mshell__grow {
    flex: 1 1 auto;
}

.mshell__title {
    font-weight: 600;
    font-size: 1.05rem;
    color: var(--m-text);
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
    color: var(--m-text);
    font-size: 1.05rem;
    font-weight: 600;
}

.mshell__library > i:first-child {
    flex: 0 0 auto;
    color: var(--m-accent);
    font-size: 1.05rem;
}

.mshell__library > i:last-child {
    flex: 0 0 auto;
    color: var(--m-text-muted);
    font-size: .8rem;
}

.mshell__lead {
    display: flex;
    align-items: center;
    gap: .5rem;
    min-width: 0;
}

.mshell__crumb {
    display: flex;
    align-items: center;
    gap: .35rem;
    min-width: 0;
    color: var(--m-text-muted);
    font-size: .82rem;
}

.mshell__crumb i {
    flex: 0 0 auto;
    color: var(--m-accent);
    font-size: .9rem;
}

.mshell__btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 2.5rem;
    height: 2.5rem;
    padding: 0;
    border: 1px solid var(--m-border);
    border-radius: .35rem;
    background: var(--m-bg);
    color: var(--m-text-muted);
    font-size: 1.1rem;
    text-decoration: none;
}

.mshell__btn:active {
    background: var(--m-surface-2);
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
    /* long unbreakable strings (raw URLs in note titles) must not create a
       horizontal scrollbar */
    overflow-wrap: anywhere;
    word-break: break-word;
    padding: .8rem 1rem;
    border: 0;
    border-bottom: 1px solid var(--m-surface-2);
    background: none;
    color: var(--m-text);
    font-size: 1rem;
    line-height: 1.35;
    text-align: left;
}

.mshell__item:active {
    background: var(--m-surface-2);
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
    gap: .5rem;
}

.mshell__item--action i {
    flex: 0 0 auto;
    width: 1.1rem;
    text-align: center;
}

.mshell__item--selected {
    font-weight: 600;
}

.mshell__danger,
.mshell__danger i {
    color: var(--m-danger);
}

.mshell__hint {
    padding: 1.25rem 1rem;
    color: var(--m-text-muted);
    overflow-wrap: anywhere;
}

.mshell__group {
    padding: .9rem 1rem .35rem;
    color: var(--m-text-muted);
    font-size: .8rem;
    text-transform: uppercase;
    letter-spacing: .03em;
}

.mshell__search {
    padding: .75rem .75rem 0;
}

.mshell__search-field {
    width: 100%;
    padding: .6rem .75rem;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius);
    background: var(--m-bg);
    color: var(--m-text);
    font-size: 1rem;
    outline: none;
}

.mshell__search-field:focus {
    border-color: var(--m-accent);
    box-shadow: 0 0 0 .2rem var(--m-accent-soft);
}

.mshell__tabs {
    flex: 0 0 auto;
    display: flex;
    border-top: 1px solid var(--m-border);
    background: var(--m-surface);
    box-shadow: 0 -2px 10px rgba(46, 52, 64, .06);
    padding-bottom: env(safe-area-inset-bottom);
}

.mshell__tab {
    position: relative;
    flex: 1 1 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: .1rem;
    padding: .45rem .25rem .35rem;
    border: 0;
    background: none;
    color: var(--m-text-muted);
    font-size: .72rem;
    transition: color .15s ease;
}

.mshell__tab i {
    padding: .1rem .85rem;
    border-radius: 999px;
    font-size: 1.3rem;
    transition: background-color .15s ease, transform .15s ease;
}

.mshell__tab--active {
    color: var(--m-accent);
    font-weight: 600;
}

.mshell__tab--active i {
    background: var(--m-accent-soft);
    transform: translateY(-1px);
}

.mshell__backdrop {
    position: fixed;
    inset: 0;
    z-index: 1100;
    background: rgba(46, 52, 64, .38);
}

/* the library leaves the bottom tabs uncovered */
.mshell__backdrop--above-tabs {
    bottom: var(--mshell-tabs, 0px);
}

/* the notebook picker is a large bottom sheet: notebooks scroll above,
   fast access to the main sections is pinned below */
/* the sheet ends above the tab bar so the bar stays visible and usable */
.mshell__sheet.mshell__sheet--library {
    inset: auto 0 var(--mshell-tabs, 0px) 0;
    width: min(100%, 36rem);
    height: auto;
    max-height: calc(100vh - var(--mshell-tabs, 0px) - 1.5rem);
    margin: 0 auto;
    border-radius: 1rem 1rem 0 0;
    box-shadow: 0 -8px 24px rgba(0, 0, 0, .18);
    animation: mshell-sheet-up .2s ease-out;
}

@keyframes mshell-sheet-up {
    from {
        transform: translateY(14px);
        opacity: .6;
    }

    to {
        transform: translateY(0);
        opacity: 1;
    }
}

.mshell__grip {
    width: 2.25rem;
    height: .25rem;
    margin: .45rem auto .1rem;
    border-radius: 999px;
    background: var(--m-border);
}

.mshell__sheet-title {
    padding-left: .25rem;
    color: var(--m-text);
    font-size: 1.02rem;
    font-weight: 700;
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
    background: var(--m-bg);
}

.mshell__sheet-bar {
    display: flex;
    align-items: center;
    gap: .5rem;
    padding: .4rem .5rem;
    border-bottom: 1px solid var(--m-border);
}

.mshell__sheet-switch {
    display: flex;
    gap: .25rem;
    margin: .55rem .75rem;
    padding: .2rem;
    border-radius: .6rem;
    background: var(--m-surface);
}

.mshell__seg {
    flex: 1 1 0;
    padding: .45rem .5rem;
    border: 0;
    border-radius: .45rem;
    background: none;
    color: var(--m-text-muted);
    font-size: .9rem;
    font-weight: 600;
}

.mshell__seg--active {
    background: var(--m-bg);
    color: var(--m-text);
    box-shadow: 0 1px 3px rgba(46, 52, 64, .12);
}

.mshell__sheet-list {
    flex: 1 1 auto;
    min-height: 0;
    overflow-y: auto;
}

.mshell__item--library {
    align-items: center;
    padding: .75rem 1rem;
}

.mshell__entry {
    display: flex;
    align-items: center;
    gap: .6rem;
    min-width: 0;
}

.mshell__entry i {
    flex: 0 0 auto;
    color: var(--m-accent);
    font-size: 1rem;
}

.mshell__ellipsis {
    min-width: 0;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}

.mshell__item-count {
    flex: 0 0 auto;
    color: var(--m-text-muted);
    font-size: .85rem;
    font-variant-numeric: tabular-nums;
}

.mshell__item--current {
    background: var(--m-accent-soft);
    color: var(--m-accent);
    font-weight: 600;
}

.mshell__shortcuts {
    display: grid;
    flex: 0 0 auto;
    grid-template-columns: 1fr 1fr;
    gap: .5rem;
    padding: .65rem .75rem .8rem;
    border-top: 1px solid var(--m-surface-2);
    background: var(--m-surface);
}

.mshell__shortcut {
    display: flex;
    align-items: center;
    gap: .45rem;
    min-width: 0;
    padding: .55rem .6rem;
    border: 1px solid var(--m-border);
    border-radius: .65rem;
    background: var(--m-bg);
    color: var(--m-text);
    font-size: .88rem;
    text-align: left;
}

.mshell__shortcut i {
    flex: 0 0 auto;
    color: var(--m-accent);
    font-size: 1rem;
}

.mshell__shortcut--active {
    border-color: var(--m-accent);
    background: var(--m-accent-soft);
    color: var(--m-accent);
    font-weight: 600;
}

.mshell__shortcut-count {
    flex: 0 0 auto;
    margin-left: auto;
    color: var(--m-text-muted);
    font-size: .78rem;
    font-variant-numeric: tabular-nums;
}
</style>
