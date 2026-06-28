<template>
    <div class="container h-100">
        <div class="row h-100 justify-content-center align-items-center">
            <div class="alert alert-danger w-50 text-center p-5">
                <h4>{{ $t('general.serverOfflineTitle') }}</h4>
                <p class="text-dark mb-0 mt-4" v-if="noteStore.status.errorText">
                    {{ noteStore.status.errorText }}
                </p>
            </div>
        </div>
    </div>
</template>

<script>
import { useNoteStore } from '../store'

export default {
    name: 'qvOffline',
    setup () {
        return { noteStore: useNoteStore() }
    }
}
</script>
