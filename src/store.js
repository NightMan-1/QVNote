import { defineStore } from 'pinia'

function lsGet (key, defaultValue) {
    try {
        const item = localStorage.getItem(key)
        return item !== null ? JSON.parse(item) : defaultValue
    } catch (e) {
        return defaultValue
    }
}

function lsSet (key, value) {
    localStorage.setItem(key, JSON.stringify(value))
}

export const useNoteStore = defineStore('note', {
    state: () => ({
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
        pageType: 'dashboard',
        sidebarType: 'notebooksList',
        settingsPageType: 'global',
        currentNotebookID: '',
        currentTagURL: '',
        showAdvancedNoteInfo: lsGet('showAdvancedNoteInfo', false),
        readerMode: lsGet('readerMode', false),
        layoutBig: lsGet('layoutBig', false),
        gridClass: 'grid-v1',
        localesList: {
            'en-US': 'English',
            'ru-RU': 'Русский'
        },
        editorsList: {
            'hugerte': 'HugeRTE'
        }
    }),

    getters: {
        apiFolder () {
            return '/api'
        },
        getNotebooksCount () {
            return Object.keys(this.notebooksList).length
        },
        getTagsCount () {
            if (this.tagsList === null) {
                return 0
            } else {
                return Object.keys(this.tagsList).length
            }
        },
        getStatus () {
            return () => this.status.errorType
        }
    },

    actions: {
        setConfig (config) {
            this.config = config
        },
        setGridClass (config) {
            this.gridClass = config
        },
        setFavoritesStatus (config) {
            this.currentArticle.favorites = config
        },
        setFavoritesCount (config) {
            this.notesCountFavorites = config
        },
        setStatus (data) {
            this.status.errorType = data.errorType
            this.status.errorText = data.errorText
        },
        setNotebooksList (data) {
            this.notebooksList = data
        },
        setNotesList (data) {
            this.notesList = data
        },
        setTags (data) {
            this.tagsList = data
        },
        setNotesCountInbox (data) {
            this.notesCountInbox = data
        },
        setNotesCountTrash (data) {
            this.notesCountTrash = data
        },
        setNotesCountTotal (data) {
            this.notesCountTotal = data
        },
        setPageType (data) {
            this.pageType = data
        },
        setSidebarType (data) {
            this.sidebarType = data
        },
        setSettingsPageType (data) {
            this.settingsPageType = data
        },
        setCurrentNotebookID (data) {
            this.currentNotebookID = data
        },
        setCurrentArticle (data) {
            this.currentArticle = data
        },
        setCurrentTagURL (data) {
            this.currentTagURL = data
        },
        doEmptyCurrentArticle () {
            this.currentArticle = this.emptyArticle
        },
        setShowAdvancedNoteInfo (data) {
            lsSet('showAdvancedNoteInfo', data)
            this.showAdvancedNoteInfo = data
        },
        setReaderMode (data) {
            lsSet('readerMode', data)
            this.readerMode = data
        },
        setLayoutMode (data) {
            lsSet('layoutBig', data)
            this.layoutBig = data
        },
        toggleShowAdvancedNoteInfo () {
            if (this.showAdvancedNoteInfo === true) {
                lsSet('showAdvancedNoteInfo', false)
                this.showAdvancedNoteInfo = false
            } else {
                lsSet('showAdvancedNoteInfo', true)
                this.showAdvancedNoteInfo = true
            }
        },
        toggleReaderMode () {
            if (this.readerMode === true) {
                lsSet('readerMode', false)
                this.readerMode = false
            } else {
                lsSet('readerMode', true)
                this.readerMode = true
            }
        },
        toggleLayoutMode () {
            if (this.layoutBig === true) {
                lsSet('layoutBig', false)
                this.layoutBig = false
            } else {
                lsSet('layoutBig', true)
                this.layoutBig = true
            }
        },
        getAllData () {
            fetch(this.apiFolder + '/notebooks.json').then((response) => { return response.json() })
                .then((jsonData) => {
                    this.setNotebooksList(jsonData)
                    this.setNotesCountTotal(0)

                    for (const value in this.notebooksList) {
                        let countTMP = this.notesCountTotal + this.notebooksList[value].notesCount
                        this.setNotesCountTotal(countTMP)
                        if (this.notebooksList[value].name === 'Inbox') {
                            this.setNotesCountInbox(this.notebooksList[value].notesCount)
                        } else if (this.notebooksList[value].name === 'Trash') {
                            this.setNotesCountTrash(this.notebooksList[value].notesCount)
                        }
                    }
                })
                .catch((error) => {
                    this.setStatus({ errorType: 2, errorText: 'Error downloading notebooks list...' })
                    console.error('Error fetching notebooks.json:', error)
                })

            fetch(this.apiFolder + '/tags.json').then((response) => { return response.json() })
                .then((jsonData) => {
                    this.setTags(jsonData)
                })
                .catch((error) => {
                    this.setStatus({ errorType: 2, errorText: 'Error downloading tags list...' })
                    console.error('Error fetching tags.json:', error)
                })
        },
        getArticle (noteUUID) {
            if (this.currentArticle.content !== undefined) {
                this.setCurrentArticle({})
            }
            fetch(this.apiFolder + '/note.json', { method: 'POST', body: JSON.stringify({ NoteID: noteUUID }) }).then((response) => { return response.json() })
                .then((jsonData) => {
                    this.setCurrentArticle(jsonData)
                })
                .catch((error) => {
                    console.error('Error fetching note.json:', error)
                    this.setStatus({ errorType: 2, errorText: 'Error downloading note...' })
                })
        },
        getFavoritesCount () {
            fetch(this.apiFolder + '/favorites.json').then((response) => { return response.json() })
                .then((jsonData) => {
                    this.setFavoritesCount(jsonData.length)
                })
                .catch(() => {
                    this.setFavoritesCount(0)
                })
        },
        getConfig () {
            fetch(this.apiFolder + '/config.json').then(response => { return response.json() })
                .then(jsonData => {
                    this.setConfig(jsonData)
                })
                .catch(error => {
                    console.error('Error fetching config.json:', error)
                })
        }
    }
})
