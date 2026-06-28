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
                router.push('/')
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

            checkServer()
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
