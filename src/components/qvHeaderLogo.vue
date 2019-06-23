<template>
    <div>
      <button class="dashboard-button" @click="goHome">
        <img style="width:1.4rem; margin:-0.15rem 0.2rem 0 0;" src="data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGRhdGEtbmFtZT0iTGF5ZXIgMSIgdmlld0JveD0iMCAwIDY0IDY0Ij4gIDxyZWN0IHdpZHRoPSIzOSIgaGVpZ2h0PSI0OCIgeD0iMTQiIHk9IjUiIGZpbGw9IiNmYWVmZGUiIHJ4PSIyIiByeT0iMiIvPiAgPHBhdGggZmlsbD0iI2NkYTFhNyIgZD0iTTEyIDVoN3Y0OEg5VjhhMyAzIDAgMCAxIDMtM3oiLz4gIDxyZWN0IHdpZHRoPSI0MiIgaGVpZ2h0PSI2IiB4PSI5IiB5PSI1MyIgZmlsbD0iI2VmZDhiZSIgcng9IjIiIHJ5PSIyIi8+ICA8cGF0aCBmaWxsPSIjOGQ2YzlmIiBkPSJNMzggMjVhMSAxIDAgMCAwLTEtMUgyN2ExIDEgMCAwIDAgMCAyaDEwYTEgMSAwIDAgMCAxLTF6TTQ1IDI0aC00YTEgMSAwIDAgMCAwIDJoNGExIDEgMCAwIDAgMC0yek00MSAyOEgzMWExIDEgMCAwIDAgMCAyaDEwYTEgMSAwIDAgMCAwLTJ6TTE5IDhhMSAxIDAgMCAwLTEgMXY0YTEgMSAwIDAgMCAyIDBWOWExIDEgMCAwIDAtMS0xeiIvPiAgPHBhdGggZmlsbD0iIzhkNmM5ZiIgZD0iTTUxIDRIMTJhNCA0IDAgMCAwLTQgNHY0OGE0IDQgMCAwIDAgNCA0aDM3YTMgMyAwIDAgMCAzLTN2LTMuMThBMyAzIDAgMCAwIDU0IDUxVjdhMyAzIDAgMCAwLTMtM3ptLTIgNTRIMTJhMiAyIDAgMCAxLTItMiAyLjI2IDIuMjYgMCAwIDEgMi0yaDM4djNhMSAxIDAgMCAxLTEgMXptMy03YTEgMSAwIDAgMS0xIDFIMjBWMTdhMSAxIDAgMCAwLTIgMHYzNWgtNmEzLjk0IDMuOTQgMCAwIDAtMiAuNjNWOGEyIDIgMCAwIDEgMi0yaDM5YTEgMSAwIDAgMSAxIDF6Ii8+ICA8cGF0aCBmaWxsPSIjOGQ2YzlmIiBkPSJNMTUgOGgtMmExIDEgMCAwIDAgMCAyaDJhMSAxIDAgMCAwIDAtMnpNMTUgMTNoLTJhMSAxIDAgMCAwIDAgMmgyYTEgMSAwIDAgMCAwLTJ6TTE1IDE4aC0yYTEgMSAwIDAgMCAwIDJoMmExIDEgMCAwIDAgMC0yek0xNSAyM2gtMmExIDEgMCAwIDAgMCAyaDJhMSAxIDAgMCAwIDAtMnpNMTUgMjhoLTJhMSAxIDAgMCAwIDAgMmgyYTEgMSAwIDAgMCAwLTJ6TTE1IDMzaC0yYTEgMSAwIDAgMCAwIDJoMmExIDEgMCAwIDAgMC0yek0xNSAzOGgtMmExIDEgMCAwIDAgMCAyaDJhMSAxIDAgMCAwIDAtMnpNMTUgNDNoLTJhMSAxIDAgMCAwIDAgMmgyYTEgMSAwIDAgMCAwLTJ6TTE1IDQ4aC0yYTEgMSAwIDAgMCAwIDJoMmExIDEgMCAwIDAgMC0yeiIvPjwvc3ZnPg==">
        <span class="text-dark-">QVNote</span>
      </button>

      <div class="dropdown btn-group settings-button">
        <button class="btn btn-outline-secondary btn-sm" title="Создать запись" @click="openEditor"><i class="fas fa-edit text-dark"></i></button>

        <button class="btn btn-outline-secondary btn-sm dropdown-toggle" type="button"
                data-toggle="dropdown" aria-haspopup="true" aria-expanded="false"
                @click.stop="showSettingsMenu = !showSettingsMenu">
            <i class="fas fa-cog text-dark"></i>
        </button>
        <div
            class="dropdown-menu"
            :class="{'show':showSettingsMenu}"
        >
            <button class="dropdown-item" @click="openEditor"><i
                class="fas fa-pencil-alt- fa-edit mr-2 text-nord3"></i> {{$t('general.addNewNote')}}
            </button>
            <button class="dropdown-item" @click="addNotebook"><i class="fas fa-book mr-2 text-nord3"></i>
                {{$t('general.addNewNotebook')}}
            </button>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item" @click="openSettings"><i class="fas fa-cog mr-2 text-nord3"></i>
                {{$t('general.buttonSettings')}}
            </button>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item" @click="powerOFF"><i class="fas fa-power-off mr-2 text-nord3"></i>
                {{$t('general.buttonExit')}}
            </button>
        </div>
    </div>
  </div>
</template>

<script>
import mixin from './mixins'
import tingle from 'tingle.js'

export default {
    name: 'qvHeaderLogo',
    mixins: [mixin],
    data () {
        return {
            showSettingsMenu: false
        }
    },
    watch: {
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
            fetch(this.$store.getters.apiFolder + '/exit')
            this.$router.push('/shutdown')
        },
        goHome (index) {
            this.$store.commit('setCurrentNotebookID', '')
            this.$store.commit('setPageType', 'dashboard')
            this.$store.commit('setSidebarType', 'notebooksList')
            this.$router.push('/')
        },
        openEditor (index) {
            this.$store.commit('doEmptyCurrentArticle')
            this.$store.commit('setCurrentNotebookID', '')
            this.$store.commit('setPageType', 'editor')
            this.$router.push({ name: 'qvNotes' })
        },
        openSettings () {
            this.$store.commit('setPageType', 'settings')
            this.$router.push('/settings')
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
                fetch(thisGlobal.$store.getters.apiFolder + '/notebook_edit.json',
                    { method: 'POST',
                        body: JSON.stringify({
                            'action': 'new',
                            'uuid': '',
                            'title': document.getElementById('notebook-new').value
                        })
                    })
                    .then(response => { return response.json() })
                    .then(jsonData => {
                        modal.destroy()
                        thisGlobal.$store.dispatch('getAllData')
                    })
                    .catch(error => {
                        console.error('Error add new notebook:', error)
                        modal.destroy()
                        thisGlobal.$store.commit('setStatus', { errorType: 2, errorText: this.$t('general.messageCanNotAddNewNotebook') })
                    })
            })
            modal.open()
        },
        toggleSettingsMenu () {
            this.showSettingsMenu = !this.showSettingsMenu
        }
    }
}
</script>
