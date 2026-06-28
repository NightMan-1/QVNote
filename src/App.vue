<template>
    <div class="app">
        <router-view/>

    </div>
</template>

<script>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useHead } from '@unhead/vue'
import { useToast } from 'vue-toastification'
import { useI18n } from 'vue-i18n'
import { useNoteStore } from './store'

const lastRouteKey = 'qvnote-last-route'

const saveLastRoute = (route) => {
    if (route.name === 'qvOffline' || route.name === 'qvErrorFatal' || route.name === 'qvInstaller' || route.name === 'qvNotFound') {
        return
    }
    try {
        localStorage.setItem(lastRouteKey, route.fullPath)
    } catch (e) {
        // ignore localStorage errors (e.g. private mode)
    }
}

const getLastRoute = () => {
    try {
        return localStorage.getItem(lastRouteKey) || '/'
    } catch (e) {
        return '/'
    }
}

export default {
    name: 'App',
    setup () {
        const noteStore = useNoteStore()
        const router = useRouter()
        const toast = useToast()
        const { t } = useI18n()
        const isOnline = ref(true)
        let pingInterval = null

        const checkServer = async () => {
            try {
                const response = await fetch(noteStore.apiFolder + '/ping')
                if (!response.ok) throw new Error('Server error')
                if (!isOnline.value) {
                    isOnline.value = true
                }
            } catch (e) {
                if (isOnline.value) {
                    isOnline.value = false
                    noteStore.setStatus({ errorType: 2, errorText: t('general.serverDownMessage') })
                }
            }
        }

        watch(isOnline, (online) => {
            if (!online && router.currentRoute.value.name !== 'qvOffline') {
                router.push({ name: 'qvOffline' })
            } else if (online && router.currentRoute.value.name === 'qvOffline') {
                toast.success(t('general.serverBackMessage'), { timeout: 5000 })
                // Full page reload guarantees that qvApp re-mounts and loads
                // the notebook/note data for the saved route.
                window.location.assign(getLastRoute())
            }
        })

        onMounted(() => {
            // global warning
            watch(() => noteStore.status.errorType, (errorType) => {
                if (errorType === 1) {
                    toast.error(noteStore.status.errorText, { timeout: 15000 })
                    router.push({ name: 'qvErrorFatal' })
                } else if (errorType === 2) {
                    toast.warning(noteStore.status.errorText, { timeout: 7000 })
                } else if (errorType === 5) {
                    toast.success(noteStore.status.errorText, { timeout: 5000 })
                }
            })

            router.afterEach((to) => {
                saveLastRoute(to)
            })

            checkServer().then(() => {
                // If user opened /offline/ directly while the server is actually up,
                // reload the last working route so that qvApp mounts correctly.
                if (isOnline.value && router.currentRoute.value.name === 'qvOffline') {
                    window.location.assign(getLastRoute())
                }
            })
            pingInterval = setInterval(checkServer, 60000)
        })

        onUnmounted(() => {
            clearInterval(pingInterval)
        })

        useHead({ title: 'QVNote' })
        return { noteStore }
    }
}
</script>
