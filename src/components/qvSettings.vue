<template>
    <div :class="gridClass">
    	<div class="grid-head-1">
            <qv-header-logo></qv-header-logo>
        </div>
        <div class="grid-sidebar-wrap"></div>
    	<div class="grid-sidebar-1">
            <nav class="sidebar-nav">
            <ul class="nav">
                <li class="nav-title text-primary mb-2">{{$t('general.sidebarSettings')}}</li>
                <li class="nav-item mb-1">
                    <button class="nav-link"
                            @click="noteStore.setSettingsPageType('global')" :class="{'active':settingsPageType === 'global'}">
                        <i class="bi bi-gear-fill me-1"></i>
                        {{$t('general.sidebarSettingsGeneral')}}
                    </button>
                </li>
                <li class="nav-item mb-1">
                    <button class="nav-link"
                            @click="noteStore.setSettingsPageType('notebooks')"
                            :class="{'active':settingsPageType === 'notebooks'}">
                        <i class="bi bi-journal-text me-1"></i>
                        {{$t('general.sidebarSettingsNotebooks')}}
                    </button>
                </li>
                <li class="nav-item">
                    <button class="nav-link" @click="noteStore.setSettingsPageType('tags')"
                            :class="{'active':settingsPageType === 'tags'}">
                        <i class="bi bi-tags-fill me-1"></i>
                        {{$t('general.sidebarSettingsTags')}}
                    </button>
                </li>
            </ul>
            </nav>
        </div>
    	<div class="grid-footer-1 bg-dark"></div>
    	<div class="grid-head-2"></div>
    	<div class="grid-body-1">
            <div class="main-content pt-0" :class="{'d-none':settingsPageType !== 'global'}">
                <div class="card mb-4">
                <div class="card-header">
                    <h5 class="m-0 ">{{$t('setting.global.settingsTitle')}}</h5>
                </div>
                <div class="card-body">
                    <p>
                        <span>{{$t('setting.global.folder')}}:</span> {{config.sourceFolder}}
                    </p>

                    <div class="mb-2">
                    <label for="StartingMode">{{$t('setting.global.runningMode')}}</label>
                    <select class="form-select w-25 select-css ms-3" v-model="startingMode" id="StartingMode">
                        <option value="independent">{{$t('setting.global.runningModeIndependent')}}</option>
                        <option value="browser">{{$t('setting.global.runningModeBrowser')}}</option>
                    </select>
                    </div>

                    <div class="form-check form-switch mb-3" v-if="startingMode === 'browser'">
                    <input type="checkbox" class="form-check-input" id="OpenBrowserSwitch" v-model="checkboxOpenBrowser">
                    <label class="form-check-label" for="OpenBrowserSwitch">{{$t('setting.global.switchOpenBrowser')}}</label>
                    </div>
                    <div class="form-check form-switch mb-3">
                    <input type="checkbox" class="form-check-input" id="CheckNewSwitch" v-model="checkboxCheckNew">
                    <label class="form-check-label" for="CheckNewSwitch">{{$t('setting.global.switchCheckNew')}}</label>
                    </div>
                    <div class="form-check form-switch mb-2" v-if="consolePresent">
                    <input type="checkbox" class="form-check-input" id="ShowConsoleSwitch" v-model="checkboxShowConsole">
                    <label class="form-check-label" for="ShowConsoleSwitch">{{$t('setting.global.switchShowConsole')}}</label>
                    </div>
                    <div class="clearfix"></div>
                    <select class="form-select mt-4 w-25 select-css" v-model="langSelected" id="localeSelect">
                    <option v-for="(lang, i) in localesList" :key="`Lang${i}`" :value="i">{{ lang }}</option>
                    </select>
                </div>
                </div>

                <div class="card mb-4">
                    <div class="card-header">
                        <h5 class="m-0 ">{{$t('setting.global.actionsTitle')}}</h5>
                    </div>
                    <div class="card-body">
                        <p>
                            <span>{{$t('setting.global.requestIndexing')}}:</span>&nbsp; <span class="text-success" v-if="!config.requestIndexing">{{$t('general.no')}}</span><span class="text-danger" v-if="config.requestIndexing">{{$t('general.yes')}}</span>
                        </p>
                        <button class="btn btn-primary me-2" v-bind:disabled="searchStatus.status !== 'idle' && searchStatus.status !== 'done'" @click="refreshData('reload')">{{$t('setting.global.btnRefreshData')}}</button>
                        <button class="btn btn-success me-2" v-bind:disabled="searchStatus.status !== 'idle' && searchStatus.status !== 'done'" @click="indexingStart">{{$t('setting.global.btnIndexChanges')}}</button>
                        <button class="btn btn-warning" v-bind:disabled="searchStatus.status !== 'idle' && searchStatus.status !== 'done'" @click="refreshData('reloadAll')">{{$t('setting.global.btnFullReload')}}</button>

                        <br>
                        <button class="btn btn-dark mt-4" v-bind:disabled="searchStatus.status !== 'idle' && searchStatus.status !== 'done'" @click="optimizationStart">{{$t('setting.global.btnDownloadImages')}}</button>
                        <br>
                        <i>{{$t('setting.global.downloadWarnLine1')}}<br><span class="text-danger">{{$t('setting.global.downloadWarnLine2')}}</span></i>

                        <div v-if="searchStatus.status === 'indexing'">
                            <p class="mt-3"><b>{{$t('setting.global.msgIndexing', [searchStatus.notesCurrent, searchStatus.notesTotal])}}:</b></p>
                            <div class="progress">
                                <div class="progress-bar progress-bar-striped bg-info progress-bar-animated" role="progressbar" v-bind:style="{ width: searchStatus.persent + '%' }" aria-valuenow="25" aria-valuemin="0" aria-valuemax="100">{{searchStatus.persent}}%</div>
                            </div>
                        </div>
                        <div v-if="searchStatus.status === 'refresh'">
                            <p class="mt-3"><b>{{$t('setting.global.msgSearchNewData')}}</b></p>
                        </div>
                        <div v-if="searchStatus.status === 'done'">
                            <p class="mt-3"><b>{{$t('setting.global.msgSearchComplete')}}</b></p>
                        </div>

                        <div v-if="optimizationStatus.status === 'processing'">
                            <p class="mt-3"><b>{{$t('setting.global.msgOptimizationStatus', [optimizationStatus.notesCurrent, optimizationStatus.notesTotal])}}:</b></p>
                            <div class="progress">
                                <div class="progress-bar progress-bar-striped bg-info progress-bar-animated" role="progressbar" v-bind:style="{ width: optimizationStatus.persent + '%' }" aria-valuenow="25" aria-valuemin="0" aria-valuemax="100">{{optimizationStatus.persent}}%</div>
                            </div>
                        </div>
                        <div v-if="optimizationStatus.status === 'done'">
                            <p class="mt-3"><b>{{$t('setting.global.msgOptimizationComplete')}}</b></p>
                        </div>

                    </div>
                </div>

                <div class="card">
                    <div class="card-header">
                        <h5 class="m-0 ">{{$t('setting.global.favorites')}}</h5>
                    </div>
                    <div class="card-body">
                        <a :href="noteStore.apiFolder + '/favorites.json'" download="favorites.json" class="btn btn-primary me-2">
                            <i class="bi bi-file-earmark-arrow-down me-1"></i> {{$t('setting.global.btnFavoritesExport')}}
                        </a>
                        <label for="favorites-upload" class="btn btn-success me-2 mb-0">
                            <i class="bi bi-file-earmark-arrow-up me-1"></i> {{$t('setting.global.btnFavoritesImport')}}
                        </label>
                        <input id="favorites-upload" type="file" v-on:change="favoritesImportSelected"/>
                    </div>
                </div>

            </div>
            <div class="main-content pt-0" :class="{'d-none':settingsPageType !== 'notebooks'}">
                <h3 class="mt-2">{{$t('setting.notebooks.title')}}</h3>
                <p class="text-muted"><i>{{$t('setting.notebooks.tips')}}</i></p>
                <div class="row" id="notebooksList">
                    <div
                            class="col-4"
                            v-for="item in notebooksList"
                            :key="item.uuid"
                            @click="notebookEdit(item.uuid, item.name)"
                    >
                        <div class="notebook-edit-link">
                            {{item.name}}
                            <i class="bi bi-gear-fill float-end"></i>
                        </div>
                    </div>
                </div>

            </div>
            <div class="main-content pt-0" :class="{'d-none':settingsPageType !== 'tags'}">
                <h3 class="mt-2">{{$t('setting.tags.title')}}</h3>
                <p class="text-muted"><i>{{$t('setting.tags.tips')}}</i></p>
                <div class="row" id="tagsList">
                    <div
                            class="col-4"
                            v-for="item in tagsList"
                            :key="item.url"
                            @click="tagEdit(item.url, item.name)"
                    >
                        <div class="notebook-edit-link">
                            {{item.name}}
                            <i class="bi bi-gear-fill float-end"></i>
                        </div>
                    </div>
                </div>
            </div>
      </div>
    </div>
</template>

<script>
import { useNoteStore } from '../store'
import tingle from 'tingle.js'
import qvHeaderLogo from './qvHeaderLogo.vue'

export default {
    name: 'qvSettings',
    components: { qvHeaderLogo },
    setup () {
        return { noteStore: useNoteStore() }
    },
    data () {
        return {
            checkboxOpenBrowser: false,
            checkboxCheckNew: false,
            checkboxShowConsole: false,
            consolePresent: false,
            startingMode: 'browser',
            searchStatus: {
                'notesCurrent': 0,
                'notesTotal': 0,
                'status': 'idle',
                'persent': 0
            },
            optimizationStatus: {
                'notesCurrent': 0,
                'notesTotal': 0,
                'status': 'idle',
                'persent': 0
            },
            langSelected: localStorage.getItem('locale') || 'ru',
            editorSelected: 'hugerte'
        }
    },
    created () {
        this.checkboxOpenBrowser = this.noteStore.config.atStartOpenBrowser
        this.checkboxCheckNew = this.noteStore.config.atStartCheckNewNotes
        this.checkboxShowConsole = this.noteStore.config.atStartShowConsole
        this.consolePresent = this.noteStore.config.consolePresent
        this.startingMode = this.noteStore.config.startingMode
        this.editorSelected = this.noteStore.config.postEditor
    },
    watch: {
        'checkboxOpenBrowser' () {
            this.saveSettings()
        },
        'checkboxCheckNew' () {
            this.saveSettings()
        },
        'checkboxShowConsole' () {
            this.saveSettings()
        },
        'startingMode' () {
            this.saveSettings()
        },
        'langSelected' () {
            localStorage.setItem('locale', this.langSelected)
            this.$i18n.locale = this.langSelected
        },
        'editorSelected' () {
            this.saveSettings()
        }

    },
    methods: {
        favoritesImportSelected: function (event) {
            if (event.target.files[0].type === 'application/json') {
                let reader = new FileReader()
                reader.onload = () => {
                    let dataFavoritesRaw = reader.result
                    try {
                        let dataFavorites = JSON.parse(dataFavoritesRaw)
                        dataFavorites.forEach((element) => {
                            fetch(this.noteStore.apiFolder + '/favorites.json', { method: 'POST', body: JSON.stringify({ 'action': 'add', 'UUID': element }) })
                        })
                        console.log(this.$t('setting.global.favoritesImportDone'))
                        this.noteStore.setStatus({ errorType: 5, errorText: this.$t('setting.global.favoritesImportDone') })
                    } catch (e) {
                        this.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.global.favoritesImportWrongData') })
                    }
                }
                reader.readAsText(event.target.files[0])
            } else {
                this.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.global.favoritesImportWrongType') })
            }
        },
        saveSettings: function () {
            var newConfig = { 'startingMode': this.startingMode.toString(), 'postEditor': this.editorSelected.toString(), 'atStartOpenBrowser': this.checkboxOpenBrowser.toString(), 'atStartShowConsole': this.checkboxShowConsole.toString(), 'atStartCheckNewNotes': this.checkboxCheckNew.toString() }
            fetch(this.noteStore.apiFolder + '/config.json', { method: 'POST', body: JSON.stringify(newConfig) }).then(response => { return response.text() })
                .then(() => {
                    fetch(this.noteStore.apiFolder + '/config.json').then(response => { return response.json() })
                        .then(jsonData => {
                            this.noteStore.setConfig(jsonData)
                        })
                        .catch(error => {
                            console.error('Error fetching config.json:', error)
                        })
                })
                .catch(error => {
                    console.error('Error fetching config.json:', error)
                    this.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.global.notificationErrorGetSearchStatus') })
                })
        },
        notebookEdit: function (uuid, title) {
            let thisGlobal = this
            let modal = new tingle.modal({
                footer: true,
                stickyFooter: false,
                closeMethods: ['overlay', 'button', 'escape'],
                closeLabel: this.$t('general.modalClose'),
                onClose: function () {
                    modal.destroy()
                }
            })
            modal.setContent('<h4 class="ml--1">' + this.$t('setting.notebooks.modalTitle') + ' <span class="text-primary">"' + title + '"</span></h4>' +
          '<div class="row mt-4 mb-0 bg-light pt-2 pb-1"><label class="col-3 col-form-label"><b>' + this.$t('setting.notebooks.modalNewTitle') + ':</b></label><div class="col-9"><input id="notebook-new-name" type="text" value="' + title + '" class="form-control"></div></div>' +
          '')
            modal.addFooterBtn(this.$t('setting.notebooks.modalBtnDelete'), 'tingle-btn tingle-btn--danger tingle-btn--pull-left', function () {
                fetch(thisGlobal.noteStore.apiFolder + '/notebook_edit.json', { method: 'POST', body: JSON.stringify({ 'action': 'remove', 'uuid': uuid }) }).then(response => { return response.text() })
                    .then(() => {
                        modal.destroy()
                        thisGlobal.noteStore.getAllData()
                    })
                    .catch(error => {
                        modal.destroy()
                        console.error('Error fetching notebook_edit.json:', error)
                        thisGlobal.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.notebooks.statusErrorDelete') })
                    })
            })
            modal.addFooterBtn(this.$t('setting.notebooks.modalBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
                modal.destroy()
            })
            modal.addFooterBtn(this.$t('setting.notebooks.modalBtnSave'), 'tingle-btn tingle-btn--warning tingle-btn--pull-right me-3', function () {
                fetch(thisGlobal.noteStore.apiFolder + '/notebook_edit.json', { method: 'POST', body: JSON.stringify({ 'action': 'rename', 'uuid': uuid, 'title': document.getElementById('notebook-new-name').value }) }).then(response => { return response.text() })
                    .then(() => {
                        modal.destroy()
                        thisGlobal.noteStore.getAllData()
                    })
                    .catch(error => {
                        modal.destroy()
                        console.error('Error fetching notebook_edit.json:', error)
                        thisGlobal.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.notebooks.statusErrorUpdate') })
                    })
            })
            modal.open()
        },
        tagEdit: function (url, title) {
            let thisGlobal = this
            let modal = new tingle.modal({
                footer: true,
                stickyFooter: false,
                closeMethods: ['overlay', 'button', 'escape'],
                closeLabel: this.$t('general.modalClose'),
                onClose: function () {
                    modal.destroy()
                }
            })
            modal.setContent('<h4 class="ml--1">' + this.$t('setting.tags.modalTitle') + ' <span class="text-primary">"' + title + '"</span></h4>' +
          '<div class="row mt-4 mb-0 bg-light pt-2 pb-1"><label class="col-3 col-form-label"><b>' + this.$t('setting.tags.modalNewTitle') + ':</b></label><div class="col-9"><input id="tag-new-name" type="text" value="' + title + '" class="form-control"></div></div>' +
          '')
            modal.addFooterBtn(this.$t('setting.tags.modalBtnDelete'), 'tingle-btn tingle-btn--danger tingle-btn--pull-left', function () {
                fetch(thisGlobal.noteStore.apiFolder + '/tag_edit.json', { method: 'POST', body: JSON.stringify({ 'action': 'remove', 'url': url }) }).then(response => { return response.text() })
                    .then(() => {
                        modal.destroy()
                        thisGlobal.noteStore.getAllData()
                    })
                    .catch(error => {
                        modal.destroy()
                        console.error('Error fetching tag_edit.json:', error)
                        thisGlobal.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.tags.statusErrorDelete') })
                    })
            })
            modal.addFooterBtn(this.$t('setting.tags.modalBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
                modal.destroy()
            })
            modal.addFooterBtn(this.$t('setting.tags.modalBtnSave'), 'tingle-btn tingle-btn--warning tingle-btn--pull-right me-3', function () {
                fetch(thisGlobal.noteStore.apiFolder + '/tag_edit.json', { method: 'POST', body: JSON.stringify({ 'action': 'rename', 'url': url, 'title': document.getElementById('tag-new-name').value }) }).then(response => { return response.text() })
                    .then(() => {
                        modal.destroy()
                        thisGlobal.noteStore.getAllData()
                    })
                    .catch(error => {
                        modal.destroy()
                        console.error('Error fetching tag_edit.json:', error)
                        thisGlobal.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.tags.statusErrorUpdate') })
                    })
            })
            modal.open()
        },
        loadData: function () {
            fetch(this.noteStore.apiFolder + '/search_index.json').then(response => { return response.json() })
                .then(jsonData => {
                    this.searchStatus.status = jsonData.status
                    this.searchStatus.notesTotal = jsonData.notesTotal
                    this.searchStatus.notesCurrent = jsonData.notesCurrent
                    if (jsonData.notesTotal > 0) {
                        this.searchStatus.persent = parseInt((jsonData.notesCurrent * 100) / jsonData.notesTotal)
                    }

                    this.noteStore.getConfig()
                })
                .catch(error => {
                    console.error('Error fetching search_index.json:', error)
                    this.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.global.notificationErrorGetSearchStatus') })
                })
        },
        indexingStart () {
            fetch(this.noteStore.apiFolder + '/search_index.json', { method: 'POST', body: JSON.stringify({ 'action': 'start' }) })
        },
        optimizationStart () {
            fetch(this.noteStore.apiFolder + '/optimization.json', { method: 'POST', body: JSON.stringify({ 'action': 'start' }) })
        },
        optimizationGet () {
            fetch(this.noteStore.apiFolder + '/optimization.json').then(response => { return response.json() })
                .then(jsonData => {
                    this.optimizationStatus.status = jsonData.status
                    this.optimizationStatus.notesTotal = jsonData.notesTotal
                    this.optimizationStatus.notesCurrent = jsonData.notesCurrent
                    if (jsonData.notesTotal > 0) {
                        this.optimizationStatus.persent = parseInt((jsonData.notesCurrent * 100) / jsonData.notesTotal)
                    }
                })
                .catch(error => {
                    console.error('Error fetching optimization.json:', error)
                })
        },
        refreshData (action) {
            fetch(this.noteStore.apiFolder + '/refresh_data.json', { method: 'POST', body: JSON.stringify({ 'action': action }) }).then(response => { return response.json() })
                .then(() => {
                    fetch(this.noteStore.apiFolder + '/config.json').then(response => { return response.json() })
                        .then(jsonData => {
                            this.noteStore.setConfig(jsonData)
                            if (action === 'reload') {
                                this.noteStore.setStatus({ errorType: 5, errorText: this.$t('setting.global.notificationAddDataRefreshed') })
                            }
                        })
                        .catch(error => {
                            console.error('Error fetching config.json:', error)
                        })
                })
                .catch(error => {
                    console.error('Error fetching refresh_data.json:', error)
                    this.noteStore.setStatus({ errorType: 2, errorText: this.$t('setting.global.notificationErrorGetSearchStatus') })
                })
        }
    },
    beforeMount: function () {
        if (this.notebooksList.length === undefined) {
            this.$router.push('/')
        }
        this.noteStore.setGridClass('grid-v1')
    },
    mounted: function () {
        this.noteStore.setSettingsPageType('global')
        this.loadData()
        this.optimizationGet()
        this.intervalConfig = setInterval(() => {
            this.loadData()
        }, 10000)
        this.intervalOptimizationStatus = setInterval(() => {
            this.optimizationGet()
        }, 3000)
    },
    beforeUnmount: function () {
        clearInterval(this.intervalConfig)
        clearInterval(this.intervalOptimizationStatus)
    },
    computed: {
        gridClass () { return this.noteStore.gridClass },
        settingsPageType () { return this.noteStore.settingsPageType },
        localesList () { return this.noteStore.localesList },
        editorsList () { return this.noteStore.editorsList },
        notebooksList () { return this.noteStore.notebooksList },
        tagsList () { return this.noteStore.tagsList },
        config () { return this.noteStore.config }
    }
}
</script>
