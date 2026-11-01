import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import { createHead } from '@unhead/vue/client'
import { createNotivue } from 'notivue'
import 'notivue/notification.css'
import 'notivue/notification-progress.css'
import 'notivue/animations.css'

import 'bootstrap-icons/font/bootstrap-icons.css'
import './style/nord-root.scss'
import './style/bootstrap/bootstrap.scss'
import './style/grid.css'
import './style/main.css'
import './style/typography.scss'

// mobile shell: own tokens, base and reader styles (independent of desktop)
import './mobile/tokens.css'
import './mobile/base.css'
import './mobile/content.css'
import './vendor/tingle/tingle.css'

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
    legacy: false,
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
const notivue = createNotivue({
    position: 'top-right',
    pauseOnHover: true,
    pauseOnTouch: true,
    pauseOnTabChange: false,
    limit: 4,
    notifications: {
        global: { duration: 5000 }
    }
})
app.use(notivue)

app.mount('#app')

// PWA: cache the shell and static assets for a fast, installable app. The dev
// server is left alone so hot reload is not affected.
if (import.meta.env.PROD && 'serviceWorker' in navigator) {
    window.addEventListener('load', () => {
        navigator.serviceWorker.register('/sw.js').catch((error) => {
            console.warn('Service worker registration failed:', error)
        })
    })
}
