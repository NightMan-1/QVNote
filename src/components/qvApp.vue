<template>
    <div :class="gridClass" id="grid">
    	<div class="grid-head-1">
            <qv-header-logo></qv-header-logo>
        </div>
    	<div class="grid-sidebar-wrap"></div>
    	<div class="grid-sidebar-1"><qv-sidebar/></div>
    	<div class="grid-footer-1">
                <div class="btn-group w-100">
                    <button class="btn text-white" @click="noteStore.setSidebarType('notebooksList')"
                            :class="{'btn-outline-primary': sidebarType !== 'notebooksList', 'btn-primary': sidebarType === 'notebooksList' }">
                        <i class="bi bi-journal-text me-1" :class="{'text-nord2': sidebarType === 'notebooksList', 'text-success': sidebarType !== 'notebooksList' }"></i>
                        {{$t('general.sidebarSwitchNotebooks')}}
                    </button>
                    <button class="btn text-white" @click="noteStore.setSidebarType('tagsList')"
                            :class="{'btn-outline-primary': sidebarType !== 'tagsList', 'btn-primary': sidebarType === 'tagsList' }">
                        <i class="bi bi-tags me-1" :class="{'text-nord2': sidebarType === 'tagsList', 'text-success': sidebarType !== 'tagsList' }"></i>
                        {{$t('general.sidebarSwitchTags')}}
                    </button>
                </div>
        </div>
    	<div class="grid-head-2">
                <div v-if="pageType === 'articleList'">
                <div class="input-group">
                    <span class="input-group-text"><i class="bi bi-search"></i></span>
                    <input type="text" class="form-control" :placeholder="$t('articleList.searchPlaceholder')" v-model="searchInput">
                    <button class="input-group-text" v-if="searchInput" @click="searchInput = ''"><i class="bi bi-eraser text-info"></i></button>
                </div>
            </div>
        </div>
    	<div class="grid-body-1"><div class="scrooll-wrap">
            <qv-dashboard v-if="pageType === 'dashboard'" />

            <qv-editor v-if="pageType === 'editor'"></qv-editor>

            <div v-if="pageType === 'articleList'">
                <div class="alert alert-danger me-5 ms-4 mt-3" v-if="mutableNotesList === null">
                    {{$t('articleList.searchNothing')}}
                </div>
                <div v-if="mutableNotesList !== null">
                    <ul class="nav article-title-list" v-if="pageType === 'articleList'">
                    <li class="nav-item" v-for="item in mutableNotesList" :key="item.uuid">
                        <button class="nav-link"
                            @click="openArticle(item.uuid, item.NoteBookUUID)"
                            :class="{ 'active': item.uuid === articleCurrent.uuid }"
                            :title="item.title"
                        >{{item.title}}
                        </button>
                    </li>

                    </ul>
                </div>
            </div>
        </div></div>
    	<div class="grid-head-3 text-end" v-if="pageType === 'articleList'">
            <button class="btn btn-outline-secondary float-start" :title="$t('articleList.btnHideSidebar')" @click="gridShow = !gridShow"><i class="bi bi-chevron-left text-black-50" v-if="gridShow"></i><i class="bi bi-chevron-right text-black-50" v-if="!gridShow"></i></button>

            <a v-bind:href="articleCurrent.url_src" v-if="articleCurrent.url_src"
                target="_blank" class="btn btn-outline-secondary me-2"><i class="bi bi-box-arrow-up-right text-dark"></i></a>

            <div class="btn-group me-2" role="group">
                <button class="btn btn-outline-secondary" :title="$t('articleList.btnInfo')" @click="doShowAdvancedInfo"><i class="bi bi-info-circle-fill text-info"></i></button>
                <button class="btn btn-outline-secondary" :title="$t('articleList.btnEdit')" @click="$router.push({name: 'qvEditor'})"><i class="bi bi-pencil-square text-success"></i></button>
                <button class="btn btn-outline-secondary" :title="$t('articleList.btnDelete')" @click="deleteArticle"><i class="bi bi-trash-fill text-danger"></i></button>
                <button class="btn btn-outline-secondary" :title="$t('articleList.btnMove')" @click="moveArticle"><i class="bi bi-arrow-left-right text-black-50"></i></button>
            </div>
            <div class="btn-group me-2" role="group">
                <button class="btn btn-outline-secondary" :class="{'btn-secondary':readerMode}" :title="$t('articleList.btnReaderMode')" @click="noteStore.toggleReaderMode()">
                    <i class="bi text-black-50 bi-book-half"></i>
                </button>
                <button class="btn btn-outline-secondary" :class="{'btn-secondary-':layoutBig, 'btn-disabled':readerMode}" :title="$t('articleList.btnReaderMode')" @click="noteStore.toggleLayoutMode()">
                    <i class="bi text-black-50" :class="{'bi-arrows-angle-expand':layoutBig, 'bi-arrows-angle-contract':!layoutBig}"></i>
                </button>
            </div>
            <button class="btn btn-outline-secondary" :title="$t('articleList.btnFavorites')" @click="addToFavorites">
                <i class="bi text-black-50" :class="articleCurrent.favorites ? 'bi-star-fill' : 'bi-star'"></i>
            </button>
            <template v-if="articleCurrent.uuid && !articleCurrent.content_state">
                <template v-if="!refetchPreview">
                    <button class="btn btn-outline-secondary ms-2" :class="{'btn-secondary':showOriginal}"
                            :title="$t('articleList.btnShowOriginal')" @click="toggleOriginal">
                        <i class="bi bi-file-earmark-code text-black-50"></i>
                    </button>
                    <button v-if="articleCurrent.url_src" class="btn btn-outline-secondary ms-2"
                            :title="$t('articleList.btnRefetch')" @click="refetchFromSource" :disabled="refetchLoading">
                        <i class="bi text-black-50" :class="refetchLoading ? 'bi-arrow-repeat spin' : 'bi-cloud-arrow-down'"></i>
                    </button>
                </template>
                <template v-else>
                    <button class="btn btn-outline-success ms-2" :title="$t('articleList.btnRefetchConfirm')" @click="confirmRefetch">
                        <i class="bi bi-check-lg"></i>
                    </button>
                    <button class="btn btn-outline-danger ms-2" :title="$t('articleList.btnRefetchCancel')" @click="cancelRefetch">
                        <i class="bi bi-x-lg"></i>
                    </button>
                </template>
            </template>
        </div>
    	<div class="grid-body-2 bg-white" v-if="pageType === 'articleList'"><div class="scrooll-wrap">
                <div class="justify-content-center article-info"
                     :class="{'d-block':showAdvancedInfo === true, 'd-none':showAdvancedInfo === false }">
                    <b>{{$t('articleList.infoDateCreate')}}:</b> {{ $filters.formatDate(articleCurrent.created_at) }}<br>
                    <b>{{$t('articleList.infoDateModify')}}:</b> {{ $filters.formatDate(articleCurrent.updated_at) }}<br>
                    <div
                        v-if="articleCurrent.tags !== null && articleCurrent.tags !== undefined && articleCurrent.tags.length > 0">
                        <b>{{$t('articleList.infoTags')}}: </b>
                        <button class="btn badge text-bg-primary me-1 fw-normal"
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
                <div :class="{'article-main':readerMode, 'article-text-big':layoutBig, 'article-text':!layoutBig,}">
                    <article>
                        <h1 class="text-success mb-3 mt-2">{{ articleCurrent.title }}</h1>
                        <div class="clearfix"></div>
                        <div class="articleCell"
                             :class="'cellType_' + articleCurrent.type"
                             v-html="displayedContent"
                        ></div>
                    </article>
                </div>
        </div></div>
    </div>
</template>

<script>
import { defineAsyncComponent } from 'vue'
import { useNoteStore } from '../store'
import qvHeaderLogo from './qvHeaderLogo.vue'
import qvDashboard from './qvDashboard.vue'
import qvSidebar from './qvSidebar.vue'
import { useModal } from '../composables/useModal'
import Prism from 'prismjs'
import 'prismjs/components/prism-clike'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-css'
import 'prismjs/components/prism-markup'
import 'prismjs/components/prism-python'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-sql'
import 'prismjs/components/prism-java'
import 'prismjs/components/prism-c'
import 'prismjs/components/prism-cpp'
import 'prismjs/components/prism-csharp'
import 'prismjs/components/prism-markup-templating'
import 'prismjs/components/prism-php'
import 'prismjs/components/prism-ruby'
import { useToast } from 'vue-toastification'
import { fetchDefuddleArticle } from '../utils/defuddle'

const qvEditor = defineAsyncComponent(() => import('./qvEditor.vue'))

export default {
    name: 'qvApp',
    components: { qvHeaderLogo, qvDashboard, qvSidebar, qvEditor },
    data () {
        return {
            articleListType: 'notes', // tags
            searchInput: '',
            notesListBackup: {},
            mutableNotesList: {},
            gridShow: true,
            contentOverride: null,
            showOriginal: false,
            refetchLoading: false,
            refetchPreview: false
        }
    },
    setup () {
        return { noteStore: useNoteStore(), modal: useModal(), toast: useToast() }
    },
    computed: {
        gridClass () { return this.noteStore.gridClass },
        pageType () { return this.noteStore.pageType },
        sidebarType () { return this.noteStore.sidebarType },
        currentNotebookID () { return this.noteStore.currentNotebookID },
        currentTagURL () { return this.noteStore.currentTagURL },
        notebooksList () { return this.noteStore.notebooksList },
        notesList () { return this.noteStore.notesList },
        articleCurrent () { return this.noteStore.currentArticle },
        showAdvancedInfo () { return this.noteStore.showAdvancedNoteInfo },
        readerMode () { return this.noteStore.readerMode },
        layoutBig () { return this.noteStore.layoutBig },
        displayedContent () {
            var content = this.contentOverride !== null
                ? this.contentOverride
                : (this.articleCurrent ? this.articleCurrent.content : '')
            content = content || ''
            // local .png links are rendered through the webp converter
            // (the /resources handler serves a cached webp for *.webp URLs);
            // the editor still works with original .png links
            return content.replace(/(\/resources\/[^"'\s]+?)\.png/gi, '$1.webp')
        }
    },
    beforeMount: function () {
        this.noteStore.getConfig()
        this.noteStore.getAllData()
        this.noteStore.getFavoritesCount()
    },
    mounted: function () {
        // Load Prism syntax-highlighting theme once
        if (!document.querySelector('link[data-prism-theme="atom-one-light"]')) {
            var prismLink = document.createElement('link')
            prismLink.rel = 'stylesheet'
            prismLink.href = '/static/prism/prism-atom-one-light.css'
            prismLink.setAttribute('data-prism-theme', 'atom-one-light')
            document.head.appendChild(prismLink)
        }
        this.highlightCodeBlocks()
        this.noteStore.setGridClass('grid-v1')
        if (this.$route.name === 'qvNote' || this.$route.name === 'qvNotebooks') {
            this.noteStore.setSidebarType('notebooksList')
            this.notebookSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
        } else if (this.$route.name === 'qvTag' || this.$route.name === 'qvTagsList') {
            this.noteStore.setSidebarType('tagsList')
            this.tagSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
        }
    },
    watch: {
        'gridShow' () {
            const root = document.documentElement
            root.style.setProperty('--sidebar-width', document.getElementsByClassName('grid-sidebar-1')[0].offsetWidth + 'px')
            root.style.setProperty('--menu-width', document.getElementsByClassName('grid-body-1')[0].offsetWidth + 'px')
            window.onresize = function (event) {
                root.style.setProperty('--sidebar-width', document.getElementsByClassName('grid-sidebar-1')[0].offsetWidth + 'px')
                root.style.setProperty('--menu-width', document.getElementsByClassName('grid-body-1')[0].offsetWidth + 'px')
            }

            const gridContent = document.querySelector('#grid')
            if (gridContent.classList.contains('hidden')) {
                gridContent.classList.remove('hidden')
            } else {
                gridContent.classList.add('hidden')
            }
        },
        '$route' (to, from) {
            // console.log('from ', from.name, 'to ', to.name)
            // console.log('route.name ', this.$route.name)
            if (from.name === 'qvEditor' && to.name === 'qvNote') {
                this.noteStore.getAllData()
                this.noteStore.setCurrentNotebookID('')
                if (this.articleCurrent.uuid !== '') {
                    this.notebookSelect(this.articleCurrent.NoteBookUUID, this.articleCurrent.uuid)
                    if (this.articleCurrent.uuid !== '') {
                        this.noteStore.getArticle(this.articleCurrent.uuid)
                    }
                }
            }
            if (this.$route.name === 'qvNote' || this.$route.name === 'qvNotebooks') {
                if (this.$route.params.nbUUID !== '') {
                    this.notebookSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
                }
            } else if (this.$route.name === 'qvApp') {
                this.noteStore.setCurrentNotebookID('')
            } else if (this.$route.name === 'qvTagsList' || this.$route.name === 'qvTag') {
                this.noteStore.setCurrentNotebookID('')
                if (this.$route.params.nbUUID !== '') {
                    this.tagSelect(this.$route.params.nbUUID, this.$route.params.noteUUID)
                }
            } else if (this.$route.name === 'qvEditor') {
                this.noteStore.setPageType('editor')
            }
        },
        'pageType' () {
            // console.log('pageType ', this.pageType)
            if (this.pageType === 'articleList') {
                this.noteStore.setGridClass('grid-v2')
            } else {
                this.noteStore.setGridClass('grid-v1')
            }
        },
        'searchInput' () {
            if (this.searchInput.length >= 3) {
                this.notesListBackup = this.notesList

                fetch(this.noteStore.apiFolder + '/search.json', { method: 'POST', body: JSON.stringify({ 'text': this.searchInput }) }).then(response => { return response.json() })
                    .then(jsonData => {
                        this.mutableNotesList = jsonData
                    })
                    .catch(error => {
                        console.error('Searching error:', error)
                        this.mutableNotesList = this.notesListBackup
                    })
            } else {
                if (this.notesListBackup !== null && this.notesListBackup.length >= 1) {
                    if (this.articleCurrent.uuid !== null && this.articleCurrent.NoteBookUUID !== null) {
                        // console.log('restore search v1')
                        this.noteStore.setCurrentNotebookID('')
                        this.notesListBackup = null
                        this.notebookSelect(this.articleCurrent.NoteBookUUID, this.articleCurrent.uuid)
                    } else {
                        // console.log('restore search v2')
                        this.mutableNotesList = this.notesListBackup
                        this.notesListBackup = null
                    }
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
        'articleCurrent.content' () {
            this.highlightCodeBlocks()
        },
        'articleCurrent.uuid' () {
            this.contentOverride = null
            this.showOriginal = false
            this.refetchPreview = false
            this.refetchLoading = false
        }
    },
    methods: {
        highlightCodeBlocks () {
            var self = this
            this.$nextTick(function () {
                var el = self.$el.querySelector('.articleCell')
                if (el) {
                    Prism.highlightAllUnder(el)
                }
            })
        },
        toggleOriginal () {
            if (this.showOriginal) {
                this.contentOverride = null
                this.showOriginal = false
                this.highlightCodeBlocks()
                return
            }
            fetch(this.noteStore.apiFolder + '/note.json', { method: 'POST', body: JSON.stringify({ NoteID: this.articleCurrent.uuid, raw: true }) })
                .then((response) => { return response.json() })
                .then((jsonData) => {
                    if (jsonData && jsonData.content !== undefined) {
                        this.contentOverride = jsonData.content
                        this.showOriginal = true
                        this.highlightCodeBlocks()
                    }
                })
                .catch((error) => {
                    console.error('Error fetching raw note:', error)
                    this.toast.error(this.$t('articleList.refetchError'))
                })
        },
        async refetchFromSource () {
            this.refetchLoading = true
            try {
                const article = await fetchDefuddleArticle(this.articleCurrent.url_src)
                this.contentOverride = article.html
                this.refetchPreview = true
                this.highlightCodeBlocks()
            } catch (e) {
                console.error('Defuddle refetch error:', e)
                this.toast.error(this.$t('articleList.refetchError') + ': ' + e.message)
            } finally {
                this.refetchLoading = false
            }
        },
        cancelRefetch () {
            this.contentOverride = null
            this.refetchPreview = false
            this.highlightCodeBlocks()
        },
        confirmRefetch () {
            fetch(this.noteStore.apiFolder + '/note_edit.json', {
                method: 'POST',
                body: JSON.stringify({
                    title: this.articleCurrent.title,
                    url: this.articleCurrent.url_src,
                    uuid: this.articleCurrent.uuid,
                    type: this.articleCurrent.type === 'code' ? 'code' : 'text',
                    content: this.contentOverride,
                    tags: this.articleCurrent.tags,
                    content_state: 'refetched'
                })
            }).then((response) => { return response.json() })
                .then(() => {
                    this.contentOverride = null
                    this.refetchPreview = false
                    this.noteStore.getAllData()
                    this.noteStore.getArticle(this.articleCurrent.uuid)
                    this.toast.success(this.$t('articleList.refetchDone'))
                })
                .catch((error) => {
                    console.error('Error saving refetched note:', error)
                    this.toast.error(this.$t('editor.errorSave'))
                })
        },
        tagSelect (nbUUID, noteUUID) {
            this.articleListType = 'tags'
            this.noteStore.setPageType('articleList')
            this.noteStore.setSidebarType('tagsList')
            if (nbUUID !== undefined && this.currentTagURL !== nbUUID) {
                this.noteStore.setCurrentTagURL(nbUUID)
                this.noteStore.setNotesList({}) // нужно для скрола списка вверх, иначе будет на предыдущей позиции

                fetch(this.noteStore.apiFolder + '/notes_with_tag.json', { method: 'POST', body: JSON.stringify({ tag: this.currentTagURL }) }).then(response => { return response.json() })
                    .then(jsonData => {
                        this.noteStore.setNotesList(jsonData)
                        if (this.notesList !== null && Object.keys(this.notesList).length > 0) {
                            let articleCurrentUUID = this.notesList[0].uuid
                            if (noteUUID !== undefined) {
                                articleCurrentUUID = noteUUID
                            }
                            this.$router.push('/tags/' + this.currentTagURL + '/' + articleCurrentUUID + '/') // сразу более правильные ссылки
                        }
                    })
                    .catch(error => {
                        console.error('Error fetching notes_with_tag.json:', error)
                        this.noteStore.setStatus({ errorType: 2, errorText: this.$t('general.messageErrorDownloadNotesWithTag') })
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
                this.noteStore.getArticle(articleCurrentUUID)
            }
        },
        notebookSelect (nbUUID, noteUUID) {
            this.noteStore.setPageType('articleList')
            this.articleListType = 'notes'
            this.noteStore.setSidebarType('notebooksList')
            if (noteUUID === undefined) {
                this.noteStore.setCurrentNotebookID('')
            }
            if (nbUUID !== undefined && this.currentNotebookID !== nbUUID) {
                this.noteStore.setCurrentNotebookID(nbUUID)
                this.noteStore.setNotesList({}) // нужно для скрола списка вверх, иначе будет на предыдущей позиции

                fetch(this.noteStore.apiFolder + '/notes_at_notebook.json', { method: 'POST', body: JSON.stringify({ NotebookID: this.currentNotebookID }) }).then(response => { return response.json() })
                    .then(jsonData => {
                        this.noteStore.setNotesList(jsonData)
                        if (this.notesList !== null && Object.keys(this.notesList).length > 0) {
                            let articleCurrentUUID = this.notesList[0].uuid
                            if (noteUUID !== undefined) {
                                articleCurrentUUID = noteUUID
                            }
                            this.$router.push('/notes/' + this.currentNotebookID + '/' + articleCurrentUUID + '/').catch(() => {}) // сразу более правильные ссылки
                        }
                    })
                    .catch(error => {
                        console.error('Error fetching notes_at_notebook.json:', error)
                        this.noteStore.setStatus({ errorType: 2, errorText: this.$t('general.messageErrorDownloadNotesList') })
                    })
            }

            let articleCurrentUUID = ''
            if (noteUUID !== undefined) {
                articleCurrentUUID = noteUUID
            } else if (this.notesList[0] !== undefined) {
                articleCurrentUUID = this.notesList[0].uuid
            }
            if (this.articleCurrent.uuid !== articleCurrentUUID) {
                this.noteStore.getArticle(articleCurrentUUID)
            }
        },
        openArticle (UUID, nbUUID) {
            if (this.$route.params.nbUUID === 'Allnotes') {
                this.$router.push('/notes/Allnotes' + '/' + UUID + '/')
            } else if (this.articleListType === 'tags' && this.currentTagURL !== '') {
                this.$router.push('/tags/' + this.currentTagURL + '/' + UUID + '/')
            } else {
                this.$router.push('/notes/' + nbUUID + '/' + UUID + '/')
            }
        },
        doShowAdvancedInfo () {
            this.noteStore.toggleShowAdvancedNoteInfo()
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
            let modal = this.modal.createModal(this.$t('general.modalClose'))

            modal.setContent('<h4 class="ml--1">' + this.$t('articleList.modalMoveTitle') + '</h4>' +
                        '<div class="row mt-4 mb-0 bg-light pt-2 pb-1">' +
                        '<label class="col-3 col-form-label"><b>' + this.$t('articleList.modalMoveNotebook') + ':</b></label>' +
                        '<div class="col-9">' + selectRAW + '</div>' +
                        '</div>' +
                        '')
            modal.addFooterBtn(this.$t('articleList.modalMoveBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
                modal.destroy()
            })
            modal.addFooterBtn(this.$t('articleList.modalMoveBtnMove'), 'tingle-btn tingle-btn--warning tingle-btn--pull-right me-3', function () {
                fetch(thisGlobal.noteStore.apiFolder + '/note_move.json',
                    { method: 'POST',
                        body: JSON.stringify({
                            'action': 'move',
                            'uuid': thisGlobal.articleCurrent.uuid,
                            'target': document.getElementById('notebookTarget').value
                        }) })
                    .then(() => {
                        thisGlobal.noteStore.getAllData()
                        thisGlobal.$router.push('/notes/' + document.getElementById('notebookTarget').value + '/' + thisGlobal.articleCurrent.uuid + '/')
                        modal.destroy()
                    })
                    .catch(error => {
                        modal.destroy()
                        console.error('Error moving note:', error)
                        thisGlobal.noteStore.setStatus({ errorType: 2, errorText: this.$t('articleList.notificationErrorMove') })
                    })
            })
            modal.open()
        },
        addToFavorites () {
            if (this.articleCurrent.favorites) {
                fetch(this.noteStore.apiFolder + '/favorites.json', { method: 'POST', body: JSON.stringify({ 'action': 'remove', 'UUID': this.articleCurrent.uuid }) })
                this.noteStore.setFavoritesStatus(false)
            } else {
                fetch(this.noteStore.apiFolder + '/favorites.json', { method: 'POST', body: JSON.stringify({ 'action': 'add', 'UUID': this.articleCurrent.uuid }) })
                this.noteStore.setFavoritesStatus(true)
            }
            this.noteStore.getFavoritesCount()
        },
        deleteArticle () {
            let thisGlobal = this
            let modal = this.modal.createModal(this.$t('general.modalClose'))

            modal.setContent('<h4 class="ml--1">' + this.$t('articleList.modalDeleteTitle') + '</h4>')
            modal.addFooterBtn(this.$t('general.noBig'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
                modal.destroy()
            })
            modal.addFooterBtn(this.$t('general.yesBig'), 'tingle-btn tingle-btn--danger tingle-btn--pull-right me-3', function () {
                fetch(thisGlobal.noteStore.apiFolder + '/note_move.json',
                    { method: 'POST',
                        body: JSON.stringify({
                            'action': 'delete',
                            'uuid': thisGlobal.articleCurrent.uuid
                        }) })
                    .then(() => {
                        modal.destroy()
                        thisGlobal.noteStore.getAllData()
                        thisGlobal.$router.push('/notes/' + thisGlobal.articleCurrent.NoteBookUUID + '/')
                    })
                    .catch(error => {
                        modal.destroy()
                        console.error('Error deleting note :', error)
                        thisGlobal.noteStore.setStatus({ errorType: 2, errorText: 'Error deleting note ...' })
                    })
            })
            modal.open()
        }
    }
}
</script>
