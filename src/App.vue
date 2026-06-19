<template>
    <div class="app">
        <router-view/>

    </div>
</template>

<script>
import { watch } from 'vue'
import { useHead } from '@unhead/vue'
import { useNoteStore } from './store'

export default {
    name: 'App',
    setup () {
        const noteStore = useNoteStore()
        useHead({ title: 'QVNote' })
        return { noteStore }
    },
    mounted: function () {
        // global warning
        watch(() => this.noteStore.status.errorType, (errorType) => {
            if (errorType === 1) {
                this.$toast.error(this.noteStore.status.errorText, { timeout: 15000 })
                this.$router.push({ name: 'qvErrorFatal' })
            } else if (errorType === 2) {
                this.$toast.warning(this.noteStore.status.errorText, { timeout: 7000 })
            } else if (errorType === 5) {
                this.$toast.success(this.noteStore.status.errorText, { timeout: 5000 })
            }
        })
        //
    }
}
</script>
