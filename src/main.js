import Vue from 'vue'
import '@coreui/coreui/dist/css/coreui.min.css'
import '@fortawesome/fontawesome-free/css/all.css'
import VueResource from 'vue-resource'
import App from './App'
import store from './store'
import router from './router'
import vueHeadful from 'vue-headful'
import qvLoader from './components/qvLoader.vue'
import CxltToastr from 'cxlt-vue2-toastr'
import 'cxlt-vue2-toastr/dist/css/cxlt-vue2-toastr.css'
import SimpleBar from 'simplebar-vue'
import 'simplebar/dist/simplebar.min.css'

import VueI18n from 'vue-i18n'
import translationsEn from './i18n/en.js'
import translationsRu from './i18n/ru.js'

Vue.use(VueI18n)

// configure localization
if (Vue.ls.get('locale', false) === false) {
  let lang = navigator.language || navigator.userLanguage
  if (lang === 'ru-RU') {
    Vue.ls.set('locale', 'ru-RU')
  } else {
    Vue.ls.set('locale', 'en-US')
  }
}

const i18n = new VueI18n({
  locale: Vue.ls.get('locale'),
  fallbackLocale: 'en-US',
  messages: {
    'en-US': translationsEn,
    'ru-RU': translationsRu
  }
})

const isDev = process.env.NODE_ENV !== 'production'
Vue.config.performance = isDev

Vue.config.productionTip = false

Vue.component('simplebar', SimpleBar)

Vue.use(VueResource)
Vue.component('vue-headful', vueHeadful)

Vue.filter('formatDate', function (value) {
  if (value) {
    // https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/DateTimeFormat
    let dateT = new Date(String(value) * 1000)
    let dateOptions = { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' }
    return new Intl.DateTimeFormat(Vue.ls.get('locale'), dateOptions).format(dateT)
  }
})

const toastrConfigs = {
  position: 'top right',
  showDuration: 5000
}
Vue.use(CxltToastr, toastrConfigs)

Vue.component('qv-loader', qvLoader)

/* eslint-disable no-new */
new Vue({
  el: '#app',
  data: {},
  i18n,
  router,
  store,
  components: { App },
  template: '<App/>',
  mounted: function () {
    this.$store.commit('setShowAdvancedNoteInfo', Vue.ls.get('showAdvancedNoteInfo', false))
    this.$store.commit('setReaderMode', Vue.ls.get('readerMode', false))
    this.$http.get(this.$store.getters.apiFolder + '/config.json').then(response => {
      this.$store.commit('setConfig', response.body)
      if (!this.$store.getters.getConfig.installed) {
        this.$router.push({ name: 'qvInstaller', params: { initialized: true } })
      } else if (this.$route.params.nbUUID !== undefined && this.$route.params.noteUUID !== undefined) {
        // this.$router.push('/app/' + this.$route.params.nbUUID + '/' + this.$route.params.noteUUID)
      } else if (this.$route.name === 'qvSettings') {
        // nothing
      } else {
        this.$router.push({ name: 'qvApp' })
      }
    }, response => {
      this.$store.commit('setStatus', { errorType: 1, errorText: this.$t('setting.global.messageErrorDownloadConfiguration') })
      console.error('Status:', response.status)
      console.error('Status text:', response.statusText)
    })
  },
  render: h => h(App)
})
