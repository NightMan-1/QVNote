<template>
    <div class="container h-100">
        <div class="row h-100 justify-content-center align-items-center">
            <div class="col-12 col-sm-10 col-md-8 col-lg-6">
                <div class="alert alert-danger text-center p-4 p-md-5 mb-0">
                    <h4>{{ $t('general.serverOfflineTitle') }}</h4>
                    <p class="text-dark mb-0 mt-4" v-if="noteStore.status.errorText">
                        {{ noteStore.status.errorText }}
                    </p>
                </div>
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
