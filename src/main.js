import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { createHead } from '@unhead/vue/client'
import Toast from 'vue-toastification'
import 'vue-toastification/dist/index.css'
import 'vue3-calendar-heatmap/dist/style.css'

import 'bootstrap-icons/font/bootstrap-icons.css'
import './style/nord.css'
import './style/bootstrap/bootstrap.scss'
import './style/grid.css'
import './style/main.css'
import './style/typography.scss'
import 'tingle.js/dist/tingle.css'

import App from './App.vue'
import router from './router'
import en from './i18n/en.js'
import ru from './i18n/ru.js'

const isDev = import.meta.env.DEV

let savedLocale = localStorage.getItem('locale') || 'ru'
// migrate old locale keys to match the i18n locale codes
if (savedLocale === 'ru-RU') {
    savedLocale = 'ru'
    localStorage.setItem('locale', 'ru')
} else if (savedLocale === 'en-US') {
    savedLocale = 'en'
    localStorage.setItem('locale', 'en')
}
const i18n = createI18n({
    legacy: true,
    locale: savedLocale,
    fallbackLocale: 'en',
    messages: { en, ru }
})

const pinia = createPinia()
const app = createApp(App)

app.config.globalProperties.$filters = {
    formatDate (value) {
        if (value) {
            const date = new Date(Number(value) * 1000)
            const locale = i18n.global.locale.value || i18n.global.locale || 'ru'
            const dateOptions = {
                year: 'numeric', month: 'long', day: 'numeric',
                hour: '2-digit', minute: '2-digit'
            }
            return new Intl.DateTimeFormat(locale, dateOptions).format(date)
        }
    },
    formatBytes (value) {
        if (!value || value === 0) return '0 B'
        const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
        const i = Math.min(
            Math.floor(Math.log(value) / Math.log(1024)),
            units.length - 1
        )
        const formatted = value / Math.pow(1024, i)
        return parseFloat(formatted.toFixed(2)) + ' ' + units[i]
    }
}

app.use(router)
app.use(i18n)
app.use(pinia)

const head = createHead()
app.use(head)
app.use(Toast, {
    position: 'top-right',
    timeout: 5000,
    closeOnClick: true,
    pauseOnHover: true,
    hideProgressBar: false,
    newestOnTop: true
})

app.mount('#app')
