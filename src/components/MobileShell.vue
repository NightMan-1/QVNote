<template>
    <div class="mshell">
        <header class="mshell__bar">
            <button class="mshell__btn" :aria-label="$t('general.menu')" @click="drawerOpen = true">
                <i class="bi bi-list"></i>
            </button>
            <div class="mshell__title">{{ pageTitle }}</div>
            <button class="mshell__btn" :aria-label="$t('articleList.searchPlaceholder')" @click="toggleSearch">
                <i class="bi" :class="searchOpen ? 'bi-x-lg' : 'bi-search'"></i>
            </button>
        </header>

        <div v-if="searchOpen" class="mshell__search">
            <input
                ref="searchField"
                v-model="searchText"
                type="search"
                class="form-control"
                :placeholder="$t('articleList.searchPlaceholder')"
            >
        </div>

        <main class="mshell__content">
            <qv-dashboard v-if="isDashboard" />

            <qv-editor v-else-if="pageType === 'editor'" />

            <div v-else-if="showSearchResults" class="mshell__list">
                <div v-if="searchResults.length === 0" class="mshell__empty">{{ $t('articleList.searchNothing') }}</div>
                <button
                    v-for="note in searchResults"
                    :key="note.uuid"
                    class="mshell__item"
                    @click="openNote(note.uuid, note.NoteBookUUID)"
                >{{ note.title }}</button>
            </div>

            <NoteArticle v-else-if="isNoteRoute && article.uuid" :article="article" />

            <div v-else class="mshell__list">
                <button
                    v-for="note in notes"
                    :key="note.uuid"
                    class="mshell__item"
                    @click="openNote(note.uuid, note.NoteBookUUID)"
                >{{ note.title }}</button>
            </div>
        </main>

        <div class="mshell__note-bar" v-if="isNoteRoute && article.uuid">
            <button class="mshell__btn" :aria-label="$t('articleList.backToList')" @click="backToList">
                <i class="bi bi-arrow-left"></i>
            </button>
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
        </div>

        <div v-if="drawerOpen" class="mshell__backdrop" @click="drawerOpen = false"></div>
        <aside class="mshell__drawer" :class="{ 'mshell__drawer--open': drawerOpen }">
            <div class="mshell__drawer-scroll"><qv-sidebar /></div>
            <div class="mshell__drawer-footer">
                <div class="btn-group w-100">
                    <button
                        class="btn text-white"
                        :class="sidebarType === 'notebooksList' ? 'btn-primary' : 'btn-outline-primary'"
                        @click="noteStore.setSidebarType('notebooksList')"
                    >
                        <i class="bi bi-journal-text me-1"></i>{{ $t('general.sidebarSwitchNotebooks') }}
                    </button>
                    <button
                        class="btn text-white"
                        :class="sidebarType === 'tagsList' ? 'btn-primary' : 'btn-outline-primary'"
                        @click="noteStore.setSidebarType('tagsList')"
                    >
                        <i class="bi bi-tags me-1"></i>{{ $t('general.sidebarSwitchTags') }}
                    </button>
                </div>
            </div>
        </aside>
    </div>
</template>

<script>
import { defineAsyncComponent } from 'vue'
import { useNoteStore } from '../store'
import qvSidebar from './qvSidebar.vue'
import NoteArticle from './NoteArticle.vue'

const qvDashboard = defineAsyncComponent(() => import('./qvDashboard.vue'))
const qvEditor = defineAsyncComponent(() => import('./qvEditor.vue'))

export default {
    name: 'MobileShell',
    components: { qvSidebar, NoteArticle, qvDashboard, qvEditor },
    data () {
        return {
            drawerOpen: false,
            searchOpen: false,
            searchText: '',
            searchResults: [],
            searchPending: false,
            searchTimer: null
        }
    },
    setup () {
        return { noteStore: useNoteStore() }
    },
    computed: {
        pageType () { return this.noteStore.pageType },
        sidebarType () { return this.noteStore.sidebarType },
        notes () {
            const list = this.noteStore.notesList
            return Array.isArray(list) ? list : []
        },
        article () { return this.noteStore.currentArticle || {} },
        currentNotebookID () { return this.noteStore.currentNotebookID },
        currentTagURL () { return this.noteStore.currentTagURL },
        isDashboard () { return this.$route.name === 'qvApp' },
        isNoteRoute () { return this.$route.name === 'qvNote' || this.$route.name === 'qvTag' },
        showSearchResults () { return this.searchText.trim().length >= 3 },
        pageTitle () {
            if (this.isDashboard) {
                return this.$t('dashboard.infoTitle')
            }
            if (this.$route.name === 'qvTag' || this.$route.name === 'qvTagsList') {
                return this.currentTagURL || this.$t('general.sidebarSwitchTags')
            }
            if (this.$route.name === 'qvEditor') {
                return this.article.uuid ? this.$t('editor.titleExist') : this.$t('editor.titleNew')
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
            const notebooks = this.noteStore.notebooksList
            const list = Array.isArray(notebooks) ? notebooks : []
            const current = list.find(n => n.uuid === id)
            return current ? current.name : 'QVNote'
        }
    },
    watch: {
        '$route' () {
            this.drawerOpen = false
            this.searchOpen = false
            this.searchText = ''
            this.searchResults = []
        },
        searchText () {
            this.runSearch()
        }
    },
    methods: {
        toggleSearch () {
            this.searchOpen = !this.searchOpen
            if (this.searchOpen) {
                this.$nextTick(() => this.$refs.searchField && this.$refs.searchField.focus())
            } else {
                this.searchText = ''
                this.searchResults = []
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
            this.searchPending = true
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
                    .finally(() => { this.searchPending = false })
            }, 250)
        },
        openNote (uuid, notebookUUID) {
            const nb = notebookUUID || this.currentNotebookID
            this.searchText = ''
            this.searchResults = []
            this.searchOpen = false
            this.$router.push('/notes/' + nb + '/' + uuid + '/')
        },
        backToList () {
            if (this.$route.name === 'qvTag') {
                this.$router.push('/tags/' + this.$route.params.nbUUID + '/')
            } else {
                this.$router.push('/notes/' + this.currentNotebookID + '/')
            }
        },
        editNote () {
            this.$router.push({ name: 'qvEditor' })
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
    height: 3.4rem;
    padding: 0 .4rem;
    background: #fff;
    border-bottom: 1px solid var(--nord4);
}

.mshell__title {
    flex: 1;
    min-width: 0;
    font-weight: 600;
    font-size: 1.05rem;
    color: var(--nord0);
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
    font-size: 1.15rem;
    text-decoration: none;
}

.mshell__btn:active {
    background: var(--nord5);
}

.mshell__search {
    flex: 0 0 auto;
    padding: .5rem .75rem;
    border-bottom: 1px solid var(--nord4);
    background: var(--nord6);
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

.mshell__empty {
    padding: 1.5rem 1rem;
    color: var(--nord3);
}

.mshell__note-bar {
    flex: 0 0 auto;
    display: flex;
    gap: .35rem;
    padding: .4rem .5rem;
    border-top: 1px solid var(--nord4);
    background: var(--nord6);
}

.mshell__backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, .45);
    z-index: 1190;
}

.mshell__drawer {
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    width: min(85vw, 20rem);
    display: flex;
    flex-direction: column;
    background: var(--nord0);
    transform: translateX(-100%);
    transition: transform .25s ease;
    z-index: 1200;
}

.mshell__drawer--open {
    transform: none;
}

.mshell__drawer-scroll {
    flex: 1 1 auto;
    overflow-y: auto;
}

.mshell__drawer-footer {
    flex: 0 0 auto;
    padding: .6rem;
    border-top: 1px solid var(--nord3);
}
</style>
