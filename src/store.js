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
    settingsPageType: 'global',
    currentNotebookID: '',
    currentTagURL: '',
    showAdvancedNoteInfo: false,
    localesList: {
      'en-US': 'English',
      'ru-RU': 'Русский'
    },
    editorsList: {
      'quill': 'Quill',
      'ckeditor5': 'CKEditor 5'
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
    toggleShowAdvancedNoteInfo: (state) => {
      if (Vue.ls.get('showAdvancedNoteInfo', false) === true) {
        Vue.ls.set('showAdvancedNoteInfo', false)
        state.showAdvancedNoteInfo = false
      } else {
        Vue.ls.set('showAdvancedNoteInfo', true)
        state.showAdvancedNoteInfo = true
      }
    }
  },
  actions: {
    getAllData (store) {
      Vue.http.get(this.getters.apiFolder + '/notebooks.json').then(response => {
        this.commit('setNotebooksList', response.body)
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
      }, response => {
        this.commit('setStatus', { errorType: 2, errorText: 'Error downloading notebooks list...' })
        console.error('Status:', response.status)
        console.error('Status text:', response.statusText)
      })

      Vue.http.get(this.getters.apiFolder + '/tags.json').then(response => {
        this.commit('setTags', response.body)
      }, response => {
        this.commit('setStatus', { errorType: 2, errorText: 'Error downloading tags list...' })
        console.error('Status:', response.status)
        console.error('Status text:', response.statusText)
      })
    },
    getArticle (store, noteUUID) {
      if (this.getters.getCurrentArticle.content !== undefined) {
        this.commit('setCurrentArticle', {})
      } // нужно для скрола списка вверх, иначе будет на предыдущей позиции
      Vue.http.post(this.getters.apiFolder + '/note.json', { 'NoteID': noteUUID }, { method: 'PUT' }).then(response => {
        this.commit('setCurrentArticle', response.body)
      }, response => {
        console.error('Status:', response.status)
        console.error('Status text:', response.statusText)
        this.commit('setStatus', { errorType: 2, errorText: 'Не удается скачать заметку...' })
      })
    },
    getFavoritesCount (store, noteUUID) {
      Vue.http.post(this.getters.apiFolder + '/favorites.json', { 'NoteID': noteUUID }, { method: 'PUT' }).then(response => {
        if (response.body !== null) {
          this.commit('setFavoritesCount', response.body.length)
        } else {
          this.commit('setFavoritesCount', 0)
        }
      })
    }
  },
  strict: process.env.NODE_ENV !== 'production'

})
