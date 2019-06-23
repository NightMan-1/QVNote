import Vue from 'vue'
import Vuex from 'vuex'
import Storage from 'vue-ls'

let options = {
    namespace: 'vuejs__', // key prefix
    name: 'ls', // name variable Vue.[ls] or this.[$ls],
    storage: 'local' // storage name session, local, memory
}

Vue.use(Storage, options)

Vue.use(Vuex)

export default new Vuex.Store({
    state: {
        status: { errorType: 0, errorText: '' },
        config: {},
        currentArticle: { title: '', uuid: '', NoteBookUUID: '', status: '', tags: [], CreatedDate: '', UpdatedDate: '', cells: {}, content: '', type: '', url_src: '', favorites: false },
        emptyArticle: { title: '', uuid: '', NoteBookUUID: '', status: '', tags: [], CreatedDate: '', UpdatedDate: '', cells: {}, content: '', type: 'text', url_src: '', favorites: false },
        settingsReload: false,
        dataReload: false,
        notebooksList: {},
        notesList: {},
        tagsList: {},
        notesCountInbox: 0,
        notesCountTrash: 0,
        notesCountTotal: 0,
        notesCountFavorites: 0,
        pageType: 'dashboard', // article, dashboard, tags, articleList, settings
        sidebarType: 'notebooksList', // tagsList
        settingsPageType: 'global',
        currentNotebookID: '',
        currentTagURL: '',
        showAdvancedNoteInfo: false,
        readerMode: false,
        gridClass: 'grid-v1',
        localesList: {
            'en-US': 'English',
            'ru-RU': 'Русский'
        },
        editorsList: {
            'quill': 'Quill'
        }

    },
    getters: {
        apiFolder () {
            if (process.env.NODE_ENV === 'development') {
                return 'http://localhost:8000/api' // localhost
            } else {
                return '/api'
            }
        },
        getConfig: state => {
            return state.config
        },
        getGridClass: state => {
            return state.gridClass
        },
        getNotebooksList: state => {
            return state.notebooksList
        },
        getNotesList: state => {
            return state.notesList
        },
        getNotebooksCount: state => {
            return Object.keys(state.notebooksList).length
        },
        getTagsList: state => {
            return state.tagsList
        },
        getTagsCount: state => {
            if (state.tagsList == null) {
                return 0
            } else {
                return Object.keys(state.tagsList).length
            }
        },
        getNotesCountInbox: state => {
            return state.notesCountInbox
        },
        getNotesCountTrash: state => {
            return state.notesCountTrash
        },
        getNotesCountTotal: state => {
            return state.notesCountTotal
        },
        getNotesCountFavorites: state => {
            return state.notesCountFavorites
        },
        getStatusText: state => {
            return state.status.errorText
        },
        getPageType: state => {
            return state.pageType
        },
        getSidebarType: state => {
            return state.sidebarType
        },
        getSettingsPageType: state => {
            return state.settingsPageType
        },
        getCurrentNotebookID: state => {
            return state.currentNotebookID
        },
        getCurrentTagURL: state => {
            return state.currentTagURL
        },
        getCurrentArticle: state => {
            return state.currentArticle
        },
        getShowAdvancedNoteInfo: state => {
            return state.showAdvancedNoteInfo
        },
        getReaderMode: state => {
            return state.readerMode
        },
        getLocalesList: state => {
            return state.localesList
        },
        getEditorsList: state => {
            return state.editorsList
        },
        getStatus: state => () => state.status.errorType
    },
    mutations: {
        setConfig: (state, config) => {
            state.config = config
        },
        setGridClass: (state, config) => {
            state.gridClass = config
        },
        setFavoritesStatus: (state, config) => {
            state.currentArticle.favorites = config
        },
        setFavoritesCount: (state, config) => {
            state.notesCountFavorites = config
        },
        setStatus: (state, data) => {
            state.status.errorType = data.errorType
            state.status.errorText = data.errorText
        },
        setNotebooksList: (state, data) => {
            state.notebooksList = data
        },
        setNotesList: (state, data) => {
            state.notesList = data
        },
        setTags: (state, data) => {
            state.tagsList = data
        },
        setNotesCountInbox: (state, data) => {
            state.notesCountInbox = data
        },
        setNotesCountTrash: (state, data) => {
            state.notesCountTrash = data
        },
        setNotesCountTotal: (state, data) => {
            state.notesCountTotal = data
        },
        setPageType: (state, data) => {
            state.pageType = data
        },
        setSidebarType: (state, data) => {
            state.sidebarType = data
        },
        setSettingsPageType: (state, data) => {
            state.settingsPageType = data
        },
        setCurrentNotebookID: (state, data) => {
            state.currentNotebookID = data
        },
        setCurrentArticle: (state, data) => {
            state.currentArticle = data
        },
        setCurrentTagURL: (state, data) => {
            state.currentTagURL = data
        },
        doEmptyCurrentArticle: (state) => {
            state.currentArticle = state.emptyArticle
        },
        setShowAdvancedNoteInfo: (state, data) => {
            Vue.ls.set('showAdvancedNoteInfo', data)
            state.showAdvancedNoteInfo = data
        },
        setReaderMode: (state, data) => {
            Vue.ls.set('readerMode', data)
            state.readerMode = data
        },
        toggleShowAdvancedNoteInfo: (state) => {
            if (Vue.ls.get('showAdvancedNoteInfo', false) === true) {
                Vue.ls.set('showAdvancedNoteInfo', false)
                state.showAdvancedNoteInfo = false
            } else {
                Vue.ls.set('showAdvancedNoteInfo', true)
                state.showAdvancedNoteInfo = true
            }
        },
        toggleReaderMode: (state) => {
            if (Vue.ls.get('readerMode', false) === true) {
                Vue.ls.set('readerMode', false)
                state.readerMode = false
            } else {
                Vue.ls.set('readerMode', true)
                state.readerMode = true
            }
        }
    },
    actions: {
        getAllData (store) {
            fetch(this.getters.apiFolder + '/notebooks.json').then(response => { return response.json() })
                .then(jsonData => {
                    this.commit('setNotebooksList', jsonData)
                    this.commit('setNotesCountTotal', 0)

                    for (const value in this.getters.getNotebooksList) {
                        let countTMP = this.getters.getNotesCountTotal + this.getters.getNotebooksList[value].notesCount
                        this.commit('setNotesCountTotal', countTMP)
                        if (this.getters.getNotebooksList[value].name === 'Inbox') {
                            this.commit('setNotesCountInbox', this.getters.getNotebooksList[value].notesCount)
                        } else if (this.getters.getNotebooksList[value].name === 'Trash') {
                            this.commit('setNotesCountTrash', this.getters.getNotebooksList[value].notesCount)
                        }
                    }
                })
                .catch(error => {
                    this.commit('setStatus', { errorType: 2, errorText: 'Error downloading notebooks list...' })
                    console.error('Error fetching notebooks.json:', error)
                })

            fetch(this.getters.apiFolder + '/tags.json').then(response => { return response.json() })
                .then(jsonData => {
                    this.commit('setTags', jsonData)
                })
                .catch(error => {
                    this.commit('setStatus', { errorType: 2, errorText: 'Error downloading tags list...' })
                    console.error('Error fetching tags.json:', error)
                })
        },
        getArticle (store, noteUUID) {
            if (this.getters.getCurrentArticle.content !== undefined) {
                this.commit('setCurrentArticle', {})
            } // нужно для скрола списка вверх, иначе будет на предыдущей позиции
            fetch(this.getters.apiFolder + '/note.json', { method: 'POST', body: JSON.stringify({ NoteID: noteUUID }) }).then(response => { return response.json() })
                .then(jsonData => {
                    this.commit('setCurrentArticle', jsonData)
                })
                .catch(error => {
                    console.error('Error fetching note.json:', error)
                    this.commit('setStatus', { errorType: 2, errorText: 'Error downloading note...' })
                })
        },
        getFavoritesCount (store, noteUUID) {
            fetch(this.getters.apiFolder + '/favorites.json').then(response => { return response.json() })
                .then(jsonData => {
                    this.commit('setFavoritesCount', jsonData.length)
                })
                .catch(() => {
                    this.commit('setFavoritesCount', 0)
                })
        }
    },
    strict: process.env.NODE_ENV !== 'production'

})
