<template>
  <div class="container-fluid pt-5 pb-5">
      <div :class="{'d-none':settingsPageType !== 'global'}">
        <div class="card">
          <div class="card-header">
            <h5 class="m-0 font-weight-bold">{{$t('setting.global.settingsTitle')}}</h5>
          </div>
          <div class="card-body">
            <p>
                <span>{{$t('setting.global.folder')}}:</span> {{config.sourceFolder}}
            </p>
            <div class="custom-control custom-switch mb-2">
              <input type="checkbox" class="custom-control-input" id="OpenBrowserSwitch" v-model="checkboxOpenBrowser">
              <label class="custom-control-label" for="OpenBrowserSwitch">{{$t('setting.global.switchOpenBrowser')}}</label>
            </div>
            <div class="custom-control custom-switch mb-2">
              <input type="checkbox" class="custom-control-input" id="CheckNewSwitch" v-model="checkboxCheckNew">
              <label class="custom-control-label" for="CheckNewSwitch">{{$t('setting.global.switchCheckNew')}}</label>
            </div>
            <div class="custom-control custom-switch mb-2">
              <input type="checkbox" class="custom-control-input" id="ShowConsoleSwitch" v-model="checkboxShowConsole">
              <label class="custom-control-label" for="ShowConsoleSwitch">{{$t('setting.global.switchShowConsole')}}</label>
            </div>
            <!--
            <div class="mb-2- mt-3">
              <b class="d-block float-left mr-2 mt-1">{{$t('setting.global.editor')}}:</b>
              <select class="form-control w-25 select-css float-left" v-model="editorSelected" id="editorSelect">
                <option v-for="(name, i) in editorsList" :key="`editor${i}`" :value="i">{{ name }}</option>
              </select>
            </div>
            -->
            <div class="clearfix"></div>
            <select class="custom-select mt-4 w-25 select-css" v-model="langSelected" id="localeSelect">
              <option v-for="(lang, i) in localesList" :key="`Lang${i}`" :value="i">{{ lang }}</option>
            </select>
          </div>
        </div>

        <div class="card">
          <div class="card-header">
            <h5 class="m-0 font-weight-bold">{{$t('setting.global.actionsTitle')}}</h5>
          </div>
          <div class="card-body">
            <p>
                <span>{{$t('setting.global.requestIndexing')}}:</span>&nbsp; <span class="text-success" v-if="!config.requestIndexing">{{$t('general.no')}}</span><span class="text-danger" v-if="config.requestIndexing">{{$t('general.yes')}}</span>
            </p>
            <button class="btn btn-primary mr-2" v-bind:disabled="searchStatus.status !== 'idle' && searchStatus.status !== 'done'" @click="refreshData('reload')">{{$t('setting.global.btnRefreshData')}}</button>
            <button class="btn btn-success mr-2" v-bind:disabled="searchStatus.status !== 'idle' && searchStatus.status !== 'done'" @click="indexingStart">{{$t('setting.global.btnIndexChanges')}}</button>
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

      </div>
      <div :class="{'d-none':settingsPageType !== 'notebooks'}">
          <h2>{{$t('setting.notebooks.title')}}</h2>
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
                      <i class="fas fa-cog float-right"></i>
                  </div>
              </div>
          </div>

      </div>
      <div :class="{'d-none':settingsPageType !== 'tags'}">
          <h2>{{$t('setting.tags.title')}}</h2>
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
                      <i class="fas fa-cog float-right"></i>
                  </div>
              </div>
          </div>
      </div>
  </div>
</template>

<script>
import tingle from 'tingle.js'

export default {
  name: 'qvSettings',
  data () {
    return {
      checkboxOpenBrowser: this.$store.getters.getConfig.atStartOpenBrowser,
      checkboxCheckNew: this.$store.getters.getConfig.atStartCheckNewNotes,
      checkboxShowConsole: this.$store.getters.getConfig.atStartShowConsole,
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
      langSelected: this.$ls.get('locale', false),
      editorSelected: this.$store.getters.getConfig.postEditor
    }
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
    'langSelected' () {
      this.$ls.set('locale', this.langSelected)
      this.$i18n.locale = this.langSelected
    },
    'editorSelected' () {
      this.saveSettings()
    }

  },
  methods: {
    saveSettings: function () {
      var newConfig = { 'postEditor': this.editorSelected.toString(), 'atStartOpenBrowser': this.checkboxOpenBrowser.toString(), 'atStartShowConsole': this.checkboxShowConsole.toString(), 'atStartCheckNewNotes': this.checkboxCheckNew.toString() }
      this.$http.post(this.$store.getters.apiFolder + '/config.json', newConfig, { method: 'PUT' }).then(response => {
        this.$http.get(this.$store.getters.apiFolder + '/config.json').then(response => {
          this.$store.commit('setConfig', response.body)
        }, response => {})
      }, response => {
        this.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.global.notificationErrorGetSearchStatus') })
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
          '<div class="form-group row mt-4 mb-0 bg-light pt-2 pb-1"><label class="col-3 col-form-label"><b>' + this.$t('setting.notebooks.modalNewTitle') + ':</b></label><div class="col-9"><input id="notebook-new-name" type="text" value="' + title + '" class="form-control"></div></div>' +
          '')
      modal.addFooterBtn(this.$t('setting.notebooks.modalBtnDelete'), 'tingle-btn tingle-btn--danger tingle-btn--pull-left', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/notebook_edit.json', { 'action': 'remove', 'uuid': uuid }, { method: 'PUT' }).then(response => {
          modal.destroy()
          thisGlobal.$store.dispatch('getAllData')
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.notebooks.statusErrorDelete') })
        })
      })
      modal.addFooterBtn(this.$t('setting.notebooks.modalBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
        modal.destroy()
      })
      modal.addFooterBtn(this.$t('setting.notebooks.modalBtnSave'), 'tingle-btn tingle-btn--warning tingle-btn--pull-right mr-3', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/notebook_edit.json', { 'action': 'rename', 'uuid': uuid, 'title': document.getElementById('notebook-new-name').value }, { method: 'PUT' }).then(response => {
          modal.destroy()
          thisGlobal.$store.dispatch('getAllData')
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.notebooks.statusErrorUpdate') })
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
          '<div class="form-group row mt-4 mb-0 bg-light pt-2 pb-1"><label class="col-3 col-form-label"><b>' + this.$t('setting.tags.modalNewTitle') + ':</b></label><div class="col-9"><input id="tag-new-name" type="text" value="' + title + '" class="form-control"></div></div>' +
          '')
      modal.addFooterBtn(this.$t('setting.tags.modalBtnDelete'), 'tingle-btn tingle-btn--danger tingle-btn--pull-left', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/tag_edit.json', { 'action': 'remove', 'url': url }, { method: 'PUT' }).then(response => {
          modal.destroy()
          thisGlobal.$store.dispatch('getAllData')
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.tags.statusErrorDelete') })
        })
      })
      modal.addFooterBtn(this.$t('setting.tags.modalBtnCancel'), 'tingle-btn tingle-btn--primary tingle-btn--pull-right', function () {
        modal.destroy()
      })
      modal.addFooterBtn(this.$t('setting.tags.modalBtnSave'), 'tingle-btn tingle-btn--warning tingle-btn--pull-right mr-3', function () {
        thisGlobal.$http.post(thisGlobal.$store.getters.apiFolder + '/tag_edit.json', { 'action': 'rename', 'url': url, 'title': document.getElementById('tag-new-name').value }, { method: 'PUT' }).then(response => {
          modal.destroy()
          thisGlobal.$store.dispatch('getAllData')
        }, response => {
          modal.destroy()
          thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.tags.statusErrorUpdate') })
        })
      })
      modal.open()
    },
    loadData: function () {
      this.$http.get(this.$store.getters.apiFolder + '/search_index.json').then(response => {
        const thisResponse = response.body
        this.searchStatus.status = thisResponse.status
        this.searchStatus.notesTotal = thisResponse.notesTotal
        this.searchStatus.notesCurrent = thisResponse.notesCurrent
        if (thisResponse.notesTotal > 0) {
          this.searchStatus.persent = parseInt((thisResponse.notesCurrent * 100) / thisResponse.notesTotal)
        }

        this.$http.get(this.$store.getters.apiFolder + '/config.json').then(response => {
          this.$store.commit('setConfig', response.body)
        }, response => {})
      }, response => {
        this.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.global.notificationErrorGetSearchStatus') })
      }).bind(this)
    },
    indexingStart () {
      this.$http.post(this.$store.getters.apiFolder + '/search_index.json', { 'action': 'start' }, { method: 'PUT' })
    },
    optimizationStart () {
      this.$http.post(this.$store.getters.apiFolder + '/optimization.json', { 'action': 'start' }, { method: 'PUT' })
    },
    optimizationGet () {
      this.$http.get(this.$store.getters.apiFolder + '/optimization.json').then(response => {
        const thisResponse = response.body
        this.optimizationStatus.status = thisResponse.status
        this.optimizationStatus.notesTotal = thisResponse.notesTotal
        this.optimizationStatus.notesCurrent = thisResponse.notesCurrent
        if (thisResponse.notesTotal > 0) {
          this.optimizationStatus.persent = parseInt((thisResponse.notesCurrent * 100) / thisResponse.notesTotal)
        }
      }).bind(this)
    },
    refreshData (action) {
      this.$http.post(this.$store.getters.apiFolder + '/refresh_data.json', { 'action': action }, { method: 'PUT' }).then(response => {
        this.$http.get(this.$store.getters.apiFolder + '/config.json').then(response => {
          this.$store.commit('setConfig', response.body)
          if (action === 'reload') {
            this.$store.commit('setStatus', { errorType: 5, errorText: this.$t('setting.global.notificationAddDataRefreshed') })
          }
        }, response => {})
      }, response => {
        this.$store.commit('setStatus', { errorType: 2, errorText: this.$t('setting.global.notificationErrorGetSearchStatus') })
      })
    }
  },
  mounted: function () {
    this.loadData()
    this.optimizationGet()
    this.intervalConfig = setInterval(() => {
      this.loadData()
    }, 10000)
    this.intervalOptimizationStatus = setInterval(() => {
      this.optimizationGet()
    }, 3000)
  },
  beforeDestroy: function () {
    clearInterval(this.intervalConfig)
    clearInterval(this.intervalOptimizationStatus)
  },
  computed: {
    notebooksList () {
      return this.$store.getters.getNotebooksList
    },
    tagsList () {
      return this.$store.getters.getTagsList
    },
    settingsPageType () {
      return this.$store.getters.getSettingsPageType
    },
    config () {
      return this.$store.getters.getConfig
    },
    localesList () {
      return this.$store.getters.getLocalesList
    },
    editorsList () {
      return this.$store.getters.getEditorsList
    }
  }
}
</script>

<style scoped>
    .notebook-edit-link{
        text-decoration: none;
        border-bottom: 1px dotted #ccc;
        padding: 0.5rem 0;
        cursor: pointer;
        color: #73818f;
    }
    .notebook-edit-link:hover{
        border-bottom: 1px solid #20a8d8;
        color: black;
    }
    .notebook-edit-link:hover svg{
        color: #20a8d8;
    }
</style>
