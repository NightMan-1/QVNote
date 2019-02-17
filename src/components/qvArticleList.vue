<template>
    <div class="row scrolly">
        <div class="col-4 col-xl-3 pr-0 b-r-1 scrolly-col">
            <div class="card mb-0 b-0" ref="notesListHead">
                <div class="card-header pl-3 pt-2 pb-2 pr-3">
                    <div class="input-group">
                        <div class="input-group-prepend">
                            <span class="input-group-text"><i class="fas fa-search"></i></span>
                        </div>
                        <input type="text" class="form-control" :placeholder="$t('articleList.searchPlaceholder')" v-model="searchInput">
                    </div>
                </div>
            </div>
            <div class="scrolly-col-wrap pt-2">
              <simplebar class="simplebarHeight" data-simplebar-auto-hide="true">
                <div class="card" id="notesList">
                    <div class="alert alert-danger mr-5 ml-4 mt-3" v-if="mutableNotesList === null">
                      {{$t('articleList.searchNothing')}}
                    </div>
                    <div class="card-body p-0" v-if="mutableNotesList !== null">
                        <button
                            class="pl-3"
                            v-for="item in mutableNotesList"
                            :key="item.uuid"
                            @click="openArticle(item.uuid, item.NoteBookUUID)"
                            :class="{ 'active': item.uuid === articleCurrent.uuid }"
                            :title="item.title"
                        >{{item.title}}
                        </button>
                    </div>
                </div>
              </simplebar>
            </div>

        </div>
        <div class="col-8 col-xl-9 scrolly-col pl-0">
            <div class="breadcrumb- d-flex p-2 bg-light b-b-1 mb-0">
                <div class="ml-auto"></div>
                <a v-bind:href="articleCurrent.url_src" v-if="articleCurrent.url_src"
                   target="_blank" class="btn btn-outline-secondary mr-2"><i class="fas fa-external-link-alt text-dark"></i></a>

                <div class="btn-group mr-2" role="group">
                  <button class="btn btn-outline-secondary" :title="$t('articleList.btnInfo')" @click="doShowAdvancedInfo"><i class="fas fa-info-circle text-info"></i></button>
                  <button class="btn btn-outline-secondary" :title="$t('articleList.btnEdit')" @click="$router.push({name: 'qvEditor'})"><i class="fas fa-edit text-success"></i></button>
                  <button class="btn btn-outline-secondary" :title="$t('articleList.btnDelete')" @click="deleteArticle"><i class="fas fa-trash text-danger"></i></button>
                  <button class="btn btn-outline-secondary" :title="$t('articleList.btnMove')" @click="moveArticle"><i class="fas fa-people-carry- fa-exchange-alt text-black-50"></i></button>
                </div>
                <button class="btn btn-outline-secondary" :title="$t('articleList.btnFavorites')" @click="addToFavorites">
                  <i class="far fa-star text-black-50" :class="{'fas':articleCurrent.favorites}"></i>
                </button>

                <div class="dropdown note-move-button dropleft d-none"> <!-- !!! disabled -->
                    <button class="btn btn-outline-secondary dropdown-toggle" type="button" data-toggle="dropdown"
                            aria-haspopup="true" aria-expanded="false" @click="showActionMenu = true">
                        <i class="fas fa-bars text-dark"></i>
                    </button>
                    <div
                        class="dropdown-menu"
                        :class="{'show':showActionMenu}"
                    >
                        <button class="dropdown-item" @click="doShowAdvancedInfo"><i
                            class="fas fa-info text-dark mr-2"></i> {{$t('articleList.btnInfo')}}
                        </button>
                        <div class="dropdown-divider"></div>
                        <button class="dropdown-item"
                                @click="$router.push({name: 'qvEditor'})">
                            <i class="fas fa-edit text-dark mr-2"></i> {{$t('articleList.btnEdit')}}
                        </button>
                        <button class="dropdown-item" @click="deleteArticle"><i class="fas fa-trash text-dark mr-2"></i>
                            {{$t('articleList.btnDelete')}}
                        </button>
                        <button class="dropdown-item" @click="moveArticle"><i
                            class="fas fa-people-carry text-dark mr-2"></i> {{$t('articleList.btnMove')}}
                        </button>
                        <!--<button class="dropdown-item" @click="doNothing"><i class="fas fa-eraser text-dark mr-2"></i> Очистить html</button>-->
                        <!--<div class="dropdown-divider" v-if="$qvGlobalData.articleCurrent.SearchIndex === false"></div>
                        <button class="dropdown-item" @click="doNothing" v-if="$qvGlobalData.articleCurrent.SearchIndex === false"><i class="fas fa-search-plus text-dark mr-2"></i> Индексировать</button>-->
                    </div>
                </div>
                <!--
                <div class="btn-group ml-auto" role="group" aria-label="Button group">
                    <button class="btn btn-outline-secondary" @click="doshowAdvancedInfo" v-html="showAnvancedIcon"></button>
                </div>
                -->
            </div>
            <div class="scrolly-col-wrap pt-2- bg-white">
              <simplebar class="simplebarHeight" data-simplebar-auto-hide="true">
                <div class="justify-content-center pt-2 pb-2 pl-3 pr-3 b-b-1"
                     :class="{'d-block':showAdvancedInfo === true, 'd-none':showAdvancedInfo === false }">
                    <b>{{$t('articleList.infoDateCreate')}}:</b> {{ articleCurrent.created_at | formatDate}}<br>
                    <b>{{$t('articleList.infoDateModify')}}:</b> {{ articleCurrent.updated_at | formatDate}}<br>
                    <div
                        v-if="articleCurrent.tags !== null && articleCurrent.tags !== undefined && articleCurrent.tags.length > 0">
                        <b>{{$t('articleList.infoTags')}}: </b>
                        <button class="btn badge badge-primary mr-1 font-weight-normal"
                                v-for="tag in articleCurrent.tags"
                                :key="tag" @click="$router.push('/tags/'+tag+'/'+articleCurrent.uuid)">
                            {{tag}}
                        </button>
                        <br>
                    </div>
                    <!--
                    Статус:{{ $qvGlobalData.articleCurrent.Status }}<br>
                    Поисковые индекс:{{ $qvGlobalData.articleCurrent.SearchIndex }}<br>
                    -->
                </div>
                <div class="pr-3 pl-3 pb-2 mt-3">
                    <article>
                        <h1 class="text-success mb-3 mt-2">{{ articleCurrent.title }}</h1>
                        <div class="clearfix"></div>
                        <div class="articleCell"
                             :class="'cellType_' + articleCurrent.type"
                             v-html="articleCurrent.content"
                        ></div>
                    </article>
                </div>
              </simplebar>
            </div>
        </div>

    </div>
</template>

<script>
import tingle from 'tingle.js'

export default {
  name: 'qvArticleList',
  components: {},
  data () {
    return {
      showActionMenu: false,
      showAnvancedIcon: '<i class="fas fa-angle-double-down text-dark"></i>',
      urlPrefix: '/notes/',
      searchInput: '',
      notesListBackup: {},
      mutableNotesList: {}
    }
  },
  mounted: function () {
    if (this.$route.name === 'qvNotebooks' || this.$route.name === 'qvNote') {
      this.urlPrefix = '/notes/'
    } else if (this.$route.name === 'qvTagsList' || this.$route.name === 'qvTag') {
      this.urlPrefix = '/tags/'
    }
  },
  destroyed: function () {
  },
  watch: {
    '$route' (to, from) {
      if (this.$route.name === 'qvNotebooks' || this.$route.name === 'qvNote') {
        this.urlPrefix = '/notes/'
      } else if (this.$route.name === 'qvTagsList' || this.$route.name === 'qvTag') {
        this.urlPrefix = '/tags/'
      }
      if (this.searchInput.length >= 3 && to.name === 'qvNote' && from.name === 'qvNote') {
        // сохраняем список поиска неизменным
      } else {
        this.searchInput = ''
        this.mutableNotesList = this.notesList
      }
      // console.log(this.searchInput)
    },
    'searchInput' () {
      if (this.searchInput.length >= 3) {
        this.notesListBackup = this.notesList
        this.$http.post(this.$store.getters.apiFolder + '/search.json', { 'text': this.searchInput }, { method: 'PUT' }).then(response => {
          this.mutableNotesList = response.body
        }, response => {
          this.mutableNotesList = this.notesListBackup
        })
      } else {
        // console.log('restore search')
        if (this.notesListBackup.length >= 1) {
          this.mutableNotesList = this.notesListBackup
        }
      }
    },
    'notesList' () {
      if (this.searchInput.length >= 3) {
        // сохраняем список поиска неизменным
      } else {
        this.mutableNotesList = this.notesList
      }
    },
    'showActionMenu' () {
      if (this.showActionMenu === true) {
        document.addEventListener('click', this.toggleMenu)
      } else {
        document.removeEventListener('click', this.toggleMenu)
      }
    }
  },
  methods: {
    openArticle (UUID, nbUUID) {
      if (this.$route.params.nbUUID === 'Allnotes') {
        this.$router.push(this.urlPrefix + 'Allnotes' + '/' + UUID)
      } else if (this.pageType === 'tags' && this.currentTagURL !== '') {
        this.$router.push(this.urlPrefix + this.currentTagURL + '/' + UUID)
      } else {
        this.$router.push(this.urlPrefix + nbUUID + '/' + UUID)
      }
    },
    doShowAdvancedInfo () {
      this.$store.commit('toggleShowAdvancedNoteInfo')
      if (this.showAdvancedInfo) {
        this.showAnvancedIcon = '<i class="fas fa-angle-double-up text-dark"></i>'
      } else {
        this.showAnvancedIcon = '<i class="fas fa-angle-double-down text-dark"></i>'
      }
    },
    moveArticle () {
      let thisGlobal = this
      let selectRAW = '<select class="form-control" id="notebookTarget">'
      for (let i in this.notebooksList) {
        if (this.articleCurrent.NoteBookUUID === this.notebooksList[i].uuid) {
          selectRAW += '<option value="' + this.notebooksList[i].uuid + '" selected>' + this.notebooksList[i].name + '</option>'
        } else {
          selectRAW += '<option value="' + this.notebooksList[i].uuid + '">' + this.notebooksList[i].name + '</option>'
        }
      }
      selectRAW += '</select>'
      let modal = new tingle.modal({
        footer: true,
        stickyFooter: false,
        closeMethods: ['overlay', 'button', 'escape'],
        closeLabel: this.$t('general.modalClose'),
        onClose: function () {
          modal.destroy()
        }
      })

      modal.setContent('<h4 class="ml--1">' + this.$t('articleList.modalMoveTitle') + '</h4>' +
                        '<div class="form-group row mt-4 mb-0 bg-light pt-2 pb-1">' +
                        '<label class="col-3 col-form-label"><b>' + this.$t('articleList.modalMoveNotebook') + ':</b></label>' +
                        '<div class="col-9">' + selectRAW + '</div>' +
                        '</div>' +
                        '')
      modal.addFooterBtn(this.$t('articleList.modalMoveBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
        modal.destroy()
      })
      modal.addFooterBtn(this.$t('articleList.modalMoveBtnMove'), 'tingle-btn tingle-btn--warning tingle-btn--pull-right mr-3', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/note_move.json', {
          'action': 'move',
          'uuid': thisGlobal.articleCurrent.uuid,
          'target': document.getElementById('notebookTarget').value
        }, { method: 'PUT' }).then(response => {
          thisGlobal.$store.dispatch('getAllData')
          thisGlobal.$router.push('/notes/' + document.getElementById('notebookTarget').value + '/' + thisGlobal.articleCurrent.uuid)
          modal.destroy()
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('articleList.notificationErrorMove') })
        })
      })
      modal.open()
    },
    addToFavorites () {
      if (this.articleCurrent.favorites) {
        this.$http.post(this.$store.getters.apiFolder + '/favorites.json', { 'action': 'remove', 'UUID': this.articleCurrent.uuid }, { method: 'PUT' })
        this.$store.commit('setFavoritesStatus', false)
      } else {
        this.$http.post(this.$store.getters.apiFolder + '/favorites.json', { 'action': 'add', 'UUID': this.articleCurrent.uuid }, { method: 'PUT' })
        this.$store.commit('setFavoritesStatus', true)
      }
      this.$store.dispatch('getFavoritesCount')
    },
    deleteArticle () {
      let thisGlobal = this
      // console.log(thisGlobal.$qvGlobalData.articleCurrent.uuid)
      let modal = new tingle.modal({
        footer: true,
        stickyFooter: false,
        closeMethods: ['overlay', 'button', 'escape'],
        closeLabel: this.$t('general.modalClose')
      })

      modal.setContent('<h4 class="ml--1">' + this.$t('articleList.modalDeleteTitle') + '</h4>')
      modal.addFooterBtn(this.$t('general.noBig'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
        modal.destroy()
      })
      modal.addFooterBtn(this.$t('general.yesBig'), 'tingle-btn tingle-btn--danger tingle-btn--pull-right mr-3', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/note_move.json', {
          'action': 'delete',
          'uuid': thisGlobal.articleCurrent.uuid
        }, { method: 'PUT' }).then(response => {
          modal.destroy()
          thisGlobal.$store.dispatch('getAllData')
          thisGlobal.$router.push('/notes/' + thisGlobal.articleCurrent.NoteBookUUID)
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: 'Error deleting note ...' })
        })
      })
      modal.open()
    },
    toggleMenu () {
      this.showActionMenu = !this.showActionMenu
    }
  },
  computed: {
    notesList () {
      return this.$store.getters.getNotesList
    },
    notebooksList () {
      return this.$store.getters.getNotebooksList
    },
    currentTagURL () {
      return this.$store.getters.getCurrentTagURL
    },
    articleCurrent () {
      return this.$store.getters.getCurrentArticle
    },
    pageType () {
      return this.$store.getters.getPageType
    },
    showAdvancedInfo () {
      return this.$store.getters.getShowAdvancedNoteInfo
    }
  }
}
</script>

<style>
    #notesList {
        border: 0;
        background: transparent;
        overflow: hidden !important;
        border-radius: 0;
    }

    #notesList button {
        overflow-x: hidden;
        text-overflow: ellipsis;
        text-align: left;
        width: 100%;
        white-space: nowrap;
        background: transparent;
        border: 0;
        margin: 0;
        margin-right: 10em;
        /* border-bottom: 1px solid #ccc; */
        padding: .45rem 1em .45rem 0;
        outline: 0;
        font-size: 0.9rem;
    }
    .platform-general #notesList button{
        padding: .4rem 1em .4rem 0;
    }

    #notesList button:last-of-type {
        border-bottom: none;
        padding-bottom: 0;
    }

    #notesList button.active {
        color: #4dbd74;
        background-color: #f0f3f5 !important;
    }

    #notesList button:hover {
        cursor: pointer;
        color: #20a8d8;
        background-color: #f0f3f5 !important;
        box-shadow: 0 0px 2px 0 rgba(0,0,0,.2);
    }
</style>
