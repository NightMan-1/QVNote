<template>
    <div>

        <div class="app-body">
            <div class="sidebar">
                <button
                    class="dashboard-button"
                    :class="{'active': pageType === 'dashboard'}"
                    @click="goHome">

                    <img style="width:1.4rem; margin:-0.15rem 0.2rem 0 0;" src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGRhdGEtbmFtZT0iTGF5ZXIgMSIgdmlld0JveD0iMCAwIDY0IDY0Ij4gIDxyZWN0IHdpZHRoPSIzOSIgaGVpZ2h0PSI0OCIgeD0iMTQiIHk9IjUiIGZpbGw9IiNmYWVmZGUiIHJ4PSIyIiByeT0iMiIvPiAgPHBhdGggZmlsbD0iI2NkYTFhNyIgZD0iTTEyIDVoN3Y0OEg5VjhhMyAzIDAgMCAxIDMtM3oiLz4gIDxyZWN0IHdpZHRoPSI0MiIgaGVpZ2h0PSI2IiB4PSI5IiB5PSI1MyIgZmlsbD0iI2VmZDhiZSIgcng9IjIiIHJ5PSIyIi8+ICA8cGF0aCBmaWxsPSIjOGQ2YzlmIiBkPSJNMzggMjVhMSAxIDAgMCAwLTEtMUgyN2ExIDEgMCAwIDAgMCAyaDEwYTEgMSAwIDAgMCAxLTF6TTQ1IDI0aC00YTEgMSAwIDAgMCAwIDJoNGExIDEgMCAwIDAgMC0yek00MSAyOEgzMWExIDEgMCAwIDAgMCAyaDEwYTEgMSAwIDAgMCAwLTJ6TTE5IDhhMSAxIDAgMCAwLTEgMXY0YTEgMSAwIDAgMCAyIDBWOWExIDEgMCAwIDAtMS0xeiIvPiAgPHBhdGggZmlsbD0iIzhkNmM5ZiIgZD0iTTUxIDRIMTJhNCA0IDAgMCAwLTQgNHY0OGE0IDQgMCAwIDAgNCA0aDM3YTMgMyAwIDAgMCAzLTN2LTMuMThBMyAzIDAgMCAwIDU0IDUxVjdhMyAzIDAgMCAwLTMtM3ptLTIgNTRIMTJhMiAyIDAgMCAxLTItMiAyLjI2IDIuMjYgMCAwIDEgMi0yaDM4djNhMSAxIDAgMCAxLTEgMXptMy03YTEgMSAwIDAgMS0xIDFIMjBWMTdhMSAxIDAgMCAwLTIgMHYzNWgtNmEzLjk0IDMuOTQgMCAwIDAtMiAuNjNWOGEyIDIgMCAwIDEgMi0yaDM5YTEgMSAwIDAgMSAxIDF6Ii8+ICA8cGF0aCBmaWxsPSIjOGQ2YzlmIiBkPSJNMTUgOGgtMmExIDEgMCAwIDAgMCAyaDJhMSAxIDAgMCAwIDAtMnpNMTUgMTNoLTJhMSAxIDAgMCAwIDAgMmgyYTEgMSAwIDAgMCAwLTJ6TTE1IDE4aC0yYTEgMSAwIDAgMCAwIDJoMmExIDEgMCAwIDAgMC0yek0xNSAyM2gtMmExIDEgMCAwIDAgMCAyaDJhMSAxIDAgMCAwIDAtMnpNMTUgMjhoLTJhMSAxIDAgMCAwIDAgMmgyYTEgMSAwIDAgMCAwLTJ6TTE1IDMzaC0yYTEgMSAwIDAgMCAwIDJoMmExIDEgMCAwIDAgMC0yek0xNSAzOGgtMmExIDEgMCAwIDAgMCAyaDJhMSAxIDAgMCAwIDAtMnpNMTUgNDNoLTJhMSAxIDAgMCAwIDAgMmgyYTEgMSAwIDAgMCAwLTJ6TTE1IDQ4aC0yYTEgMSAwIDAgMCAwIDJoMmExIDEgMCAwIDAgMC0yeiIvPjwvc3ZnPg==">
                    <span class="text-dark-">QVNote</span>
                </button>

                <div class="dropdown  btn-group settings-button">
                    <button class="btn btn-outline-secondary btn-sm" title="Создать запись" @click="openEditor"><i class="fas fa-edit text-dark"></i></button>

                    <button class="btn btn-outline-secondary btn-sm dropdown-toggle" type="button"
                            data-toggle="dropdown" aria-haspopup="true" aria-expanded="false"
                            @click.stop="showSettingsMenu = true">
                        <i class="fas fa-cog text-dark"></i>
                    </button>
                    <div
                        class="dropdown-menu"
                        :class="{'show':showSettingsMenu}"
                    >
                        <button class="dropdown-item" @click="openEditor"><i
                            class="fas fa-pencil-alt- fa-edit text-dark mr-2"></i> {{$t('general.addNewNote')}}
                        </button>
                        <!--<button class="dropdown-item"><i class="fas fa-link text-dark mr-2"></i> Импортировать ссылку</button>-->
                        <button class="dropdown-item" @click="addNotebook"><i class="fas fa-book text-dark mr-2"></i>
                            {{$t('general.addNewNotebook')}}
                        </button>
                        <div class="dropdown-divider"></div>
                        <button class="dropdown-item" @click="openSettings"><i class="fas fa-cog text-dark mr-2"></i>
                            {{$t('general.buttonSettings')}}
                        </button>
                        <div class="dropdown-divider"></div>
                        <button class="dropdown-item" @click="powerOFF"><i class="fas fa-power-off text-dark mr-2"></i>
                            {{$t('general.buttonExit')}}
                        </button>
                    </div>
                </div>

                <nav class="sidebar-nav">
                  <simplebar class="simplebarHeight" data-simplebar-auto-hide="true">
                    <ul class="nav"
                        :class="{'d-none':(pageType !== 'dashboard' && pageType !==  'articleList' && pageType !==  'editor')}">
                        <li class="nav-title text-success">
                            {{$t('general.sidebarLibrary')}}
                        </li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left"
                                    :class="{ 'active': currentNotebookID === 'Inbox' }"
                                    @click="$router.push('/notes/Inbox')">
                                <span class="badge badge-primary">{{notesCountInbox}}</span>
                                <i class="fas fa-inbox mr-1"></i> {{$t('general.sidebarInbox')}}
                            </button>
                        </li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left"
                                    :class="{ 'active': currentNotebookID === 'Favorites' }"
                                    @click="$router.push('/notes/Favorites')">
                                <span class="badge badge-primary">{{notesCountFavorites}}</span>
                                <i class="fas fa-star mr-1"></i> {{$t('general.sidebarFavorites')}}
                            </button>
                        </li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left"
                                    :class="{ 'active': currentNotebookID === 'Trash' }"
                                    @click="$router.push('/notes/Trash')">
                                <span class="badge badge-primary">{{notesCountTrash}}</span>
                                <i class="fas fa-trash-alt mr-1"></i> {{$t('general.sidebarTrash')}}
                            </button>
                        </li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left"
                                    :class="{ 'active': currentNotebookID === 'Allnotes' }"
                                    @click="$router.push('/notes/Allnotes')">
                                <span class="badge badge-primary" v-if="notesCountTotal > 0">{{notesCountTotal}}</span>
                                <i class="fas fa-archive mr-1"></i> {{$t('general.sidebarAllNotes')}}
                            </button>
                        </li>

                        <li class="nav-title text-primary">{{$t('general.sidebarNotebooks')}}</li>
                        <li
                            class="nav-item"
                            v-for="item in notebooksList" v-if="item.name !== 'Inbox' && item.name !== 'Trash'"
                            :key="item.uuid"
                        >
                            <button class="nav-link nav-link-notebook bg-primary- border-0 w-100 text-left"
                                    @click="$router.push('/notes/' + item.uuid)"
                                    :class="{ 'active': item.uuid === currentNotebookID }">
                                <span class="badge badge-primary">{{item.notesCount}}</span>
                                {{item.name}}
                            </button>
                        </li>
                    </ul>

                    <ul class="nav" :class="{'d-none':pageType !== 'tags'}">
                        <li class="nav-title text-primary">{{$t('general.sidebarTags')}}</li>
                        <li
                            class="nav-item"
                            v-for="item in tagsList"
                            :key="item.url"
                        >
                            <button class="nav-link nav-link-notebook bg-primary- border-0 w-100 text-left"
                                    @click="$router.push('/tags/'+item.url)"
                                    :class="{ 'active': item.name === currentTagURL }">
                                <span class="badge badge-primary">{{item.count}}</span>
                                {{item.name}}
                            </button>
                        </li>

                    </ul>

                    <ul class="nav" :class="{'d-none':pageType !== 'settings'}">
                        <li class="nav-title text-primary">{{$t('general.sidebarSettings')}}</li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left"
                                    @click="$store.commit('setSettingsPageType', 'global')" :class="{'active':settingsPageType === 'global'}">
                                <i class="fas fa-cog mr-1"></i>
                                {{$t('general.sidebarSettingsGeneral')}}
                            </button>
                        </li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left"
                                    @click="$store.commit('setSettingsPageType', 'notebooks')"
                                    :class="{'active':settingsPageType === 'notebooks'}">
                                <i class="fas fa-book mr-1"></i>
                                {{$t('general.sidebarSettingsNotebooks')}}
                            </button>
                        </li>
                        <li class="nav-item">
                            <button class="nav-link bg-primary- border-0 w-100 text-left" @click="$store.commit('setSettingsPageType', 'tags')"
                                    :class="{'active':settingsPageType === 'tags'}">
                                <i class="fas fa-tags mr-1"></i>
                                {{$t('general.sidebarSettingsTags')}}
                            </button>
                        </li>

                    </ul>
                  </simplebar>
                </nav>

                <div id="mainactions" class="btn-group justify-content-center p-2 pt-3" role="group"
                     :class="{'d-none':pageType === 'settings'}">
                    <button class="btn text-white" @click="$store.commit('setPageType', 'articleList')"
                            :class="{'btn-outline-primary': pageType !== 'articleList', 'btn-primary': pageType === 'articleList' }">
                        <i class="fas fa-book text-success"></i> {{$t('general.sidebarSwitchNotebooks')}}
                    </button>
                    <button class="btn text-white" @click="$store.commit('setPageType', 'tags')"
                            :class="{'btn-outline-primary': pageType !== 'tags', 'btn-primary': pageType === 'tags' }">
                        <i class="fas fa-tags text-success"></i> {{$t('general.sidebarSwitchTags')}}
                    </button>
                </div>
            </div>
            <main class="main">

                <transition name="fade">
                    <qv-editor v-if="pageType === 'editor'"></qv-editor>
                </transition>
                <transition name="fade">
                    <qv-settings v-if="pageType === 'settings'"></qv-settings>
                </transition>
                <transition name="fade">
                    <qv-dashboard v-if="pageType === 'dashboard'" ></qv-dashboard>
                </transition>
                <transition name="fade">
                    <qv-article-list v-if="pageType === 'articleList' || pageType === 'tags'" ></qv-article-list>
                </transition>

            </main>
        </div>
    </div>
</template>

<script>
import qvEditor from './qvEditor.vue'
import qvDashboard from './qvDashboard.vue'
import qvArticleList from './qvArticleList.vue'
import qvSettings from './qvSettings.vue'
import tingle from 'tingle.js'

export default {
  name: 'qvApp',
  data () {
    return {
      showSettingsMenu: false
    }
  },
  components: {
    qvEditor,
    qvDashboard,
    qvArticleList,
    qvSettings
  },
  mounted: function () {
    document.body.className = ''
    document.body.classList.add('app', 'sidebar-fixed', 'sidebar-show')

    this.$store.dispatch('getAllData')
    this.$store.dispatch('getFavoritesCount')

    if (this.$route.name === 'qvNote' && this.$route.params.nbUUID !== undefined && this.$route.params.noteUUID !== undefined) {
      this.notebookSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
    } else if (this.$route.name === 'qvTag' && this.$route.params.nbUUID !== undefined && this.$route.params.noteUUID !== undefined) {
      this.tagSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
    } else if (this.$route.name === 'qvSettings') {
      this.show = 'settings'
      this.$router.push('/settings')
    }

    let thisGlobal = this
    setInterval(function () {
      thisGlobal.$http.get(thisGlobal.$store.getters.apiFolder + '/ping').then(response => {
        // console.log(response.body.result)
        if (response.body.result !== 'pong') {
          thisGlobal.$router.push('/shutdown')
        }
      }, response => {
        thisGlobal.$router.push('/shutdown')
      })
    }, 2000)
  },
  destroyed: function () {
    document.body.className = ''
  },
  watch: {
    '$route' (to, from) {
      // console.log('from ', from.name, 'to ', to.name)
      if (from.name === 'qvEditor' && to.name === 'qvNote') {
        this.$store.dispatch('getAllData')
        this.$store.commit('setCurrentNotebookID', '')

        if (this.articleCurrent.uuid !== '') {
          this.notebookSelect(this.articleCurrent.NoteBookUUID, this.articleCurrent.uuid)
          if (this.articleCurrent.uuid !== '') {
            this.$store.dispatch('getArticle', this.articleCurrent.uuid)
          }
        }
      }

      if (this.$route.name === 'qvNote' || this.$route.name === 'qvNotebooks') {
        if (this.$route.params.nbUUID !== '') {
          this.notebookSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
        }
      } else if (this.$route.name === 'qvApp') {
        this.$store.commit('setCurrentNotebookID', '')
      } else if (this.$route.name === 'qvTagsList' || this.$route.name === 'qvTag') {
        this.$store.commit('setCurrentNotebookID', '')
        if (this.$route.params.nbUUID !== '') {
          this.tagSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
        }
      } else if (this.$route.name === 'qvSettings') {
        this.$store.commit('setPageType', 'settings')
      } else if (this.$route.name === 'qvEditor') {
        this.$store.commit('setPageType', 'editor')
      }
    },
    'showSettingsMenu' () {
      if (this.showSettingsMenu === true) {
        document.addEventListener('click', this.toggleSettingsMenu)
      } else {
        document.removeEventListener('click', this.toggleSettingsMenu)
      }
    }
  },
  methods: {
    powerOFF () {
      this.$http.get(this.$store.getters.apiFolder + '/exit')
      this.$router.push('/shutdown')
    },
    goHome (index) {
      this.$store.commit('setCurrentNotebookID', '')
      this.$store.commit('setPageType', 'dashboard')
      this.$router.push('/app')
    },
    openEditor (index) {
      this.$store.commit('doEmptyCurrentArticle')
      this.$store.commit('setCurrentNotebookID', '')
      this.$store.commit('setPageType', 'editor')
      this.$router.push({ name: 'qvNotes' })
    },
    tagSelect (nbUUID, noteUUID) {
      this.$store.commit('setPageType', 'tags')
      if (nbUUID !== undefined && this.currentTagURL !== nbUUID) {
        this.$store.commit('setCurrentTagURL', nbUUID)
        this.$store.commit('setNotesList', {}) // нужно для скрола списка вверх, иначе будет на предыдущей позиции
        this.$http.post(this.$store.getters.apiFolder + '/notes_with_tag.json', { 'tag': this.currentTagURL }, { method: 'PUT' }).then(response => {
          this.$store.commit('setNotesList', response.body)
          if (this.notesList !== null && Object.keys(this.notesList).length > 0) {
            let articleCurrentUUID = this.notesList[0].uuid
            if (noteUUID !== undefined) {
              articleCurrentUUID = noteUUID
            }
            this.$router.push('/tags/' + this.currentTagURL + '/' + articleCurrentUUID) // сразу более правильные ссылки
          }

          // articleCurrent
        }, response => {
          this.$store.commit('setStatus', { errorType: 2, errorText: this.$t('general.messageErrorDownloadNotesWithTag') })
          console.error('Status:', response.status)
          console.error('Status text:', response.statusText)
        })
      }

      let articleCurrentUUID = ''
      if (this.notesList[0] !== undefined) {
        articleCurrentUUID = this.notesList[0].uuid
      }
      if (noteUUID !== undefined) {
        articleCurrentUUID = noteUUID
      }
      if (this.articleCurrent.uuid !== articleCurrentUUID) {
        this.$store.dispatch('getArticle', articleCurrentUUID)
      }
    },
    notebookSelect (nbUUID, noteUUID) {
      if (noteUUID === undefined) {
        this.$store.commit('setCurrentNotebookID', '')
      }
      this.$store.commit('setPageType', 'articleList')
      if (nbUUID !== undefined && this.currentNotebookID !== nbUUID) {
        this.$store.commit('setCurrentNotebookID', nbUUID)
        this.$store.commit('setNotesList', {}) // нужно для скрола списка вверх, иначе будет на предыдущей позиции
        const global = this
        this.$http.post(this.$store.getters.apiFolder + '/notes_at_notebook.json', { 'NotebookID': global.currentNotebookID }, { method: 'PUT' }).then(response => {
          this.$store.commit('setNotesList', response.body)
          if (this.notesList !== null && Object.keys(this.notesList).length > 0) {
            let articleCurrentUUID = this.notesList[0].uuid
            if (noteUUID !== undefined) {
              articleCurrentUUID = noteUUID
            }
            this.$router.push('/notes/' + global.currentNotebookID + '/' + articleCurrentUUID) // сразу более правильные ссылки
          }
        }, response => {
          global.$store.commit('setStatus', { errorType: 2, errorText: this.$t('general.messageErrorDownloadNotesList') })
          console.error('Status:', response.status)
          console.error('Status text:', response.statusText)
        })
      }

      let articleCurrentUUID = ''
      if (this.notesList[0] !== undefined) {
        articleCurrentUUID = this.notesList[0].uuid
      }
      if (noteUUID !== undefined) {
        articleCurrentUUID = noteUUID
      }
      if (this.articleCurrent.uuid !== articleCurrentUUID) {
        this.$store.dispatch('getArticle', articleCurrentUUID)
      }
    },
    openSettings () {
      if (this.pageType === 'settings') {
        this.$store.commit('setPageType', 'dashboard')
        this.$router.push('/app')
      } else {
        this.$store.commit('setPageType', 'settings')
        this.$router.push('/settings')
      }
    },
    addNotebook () {
      let thisGlobal = this
      let modal = new tingle.modal({
        footer: true,
        stickyFooter: false,
        closeMethods: ['overlay', 'button', 'escape'],
        closeLabel: this.$t('general.modalClose')
      })
      modal.setContent('<h4 class="ml--1">' + this.$t('general.modalNewNotebookTitle') + ':</h4>' +
                    '<div class="form-group row mt-4 mb-0 bg-light pt-2 pb-1"><label class="col-sm-4 col-form-label"><b>' + this.$t('general.modalNewNotebookText') + '</b></label><div class="col-sm-8"><input id="notebook-new" type="text" class="form-control"></div></div>' +
                    '')
      modal.addFooterBtn(this.$t('general.modalNewNotebookBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () { modal.destroy() })
      modal.addFooterBtn(this.$t('general.modalNewNotebookBtnAdd'), 'tingle-btn tingle-btn--default tingle-btn--pull-right mr-3', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/notebook_edit.json', {
          'action': 'new',
          'uuid': '',
          'title': document.getElementById('notebook-new').value
        }, { method: 'PUT' }).then(response => {
          modal.destroy()
          thisGlobal.$store.dispatch('getAllData')
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('general.messageCanNotAddNewNotebook') })
        })
      })
      modal.open()
    },
    toggleSettingsMenu () {
      this.showSettingsMenu = !this.showSettingsMenu
    }
  },
  computed: {
    pageType () {
      return this.$store.getters.getPageType
    },
    notesCountInbox () {
      return this.$store.getters.getNotesCountInbox
    },
    notesCountTrash () {
      return this.$store.getters.getNotesCountTrash
    },
    notesCountTotal () {
      return this.$store.getters.getNotesCountTotal
    },
    notesCountFavorites () {
      return this.$store.getters.getNotesCountFavorites
    },
    currentNotebookID () {
      return this.$store.getters.getCurrentNotebookID
    },
    notebooksList () {
      return this.$store.getters.getNotebooksList
    },
    notesList () {
      return this.$store.getters.getNotesList
    },
    tagsList () {
      return this.$store.getters.getTagsList
    },
    currentTagURL () {
      return this.$store.getters.getCurrentTagURL
    },
    articleCurrent () {
      return this.$store.getters.getCurrentArticle
    },
    settingsPageType () {
      return this.$store.getters.getSettingsPageType
    }
  }
}
</script>

<style scoped>
    .fade-enter-active, .fade-leave-active {
        transition: opacity .2s;
    }

    .fade-enter, .fade-leave-to /* .fade-leave-active below version 2.1.8 */
    {
        opacity: 0;
    }
</style>
