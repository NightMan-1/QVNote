<template>
    <div class="app">
        <vue-headful
            title="QVNote"
            description=""
            keywords=""
        />
        <router-view/>

    </div>
</template>

<script>
export default {
    name: 'App',
    mounted: function () {
    // global warning
        this.$store.watch(this.$store.getters.getStatus, errorType => {
            if (errorType === 1) {
                this.$toast.error({
                    title: 'Error!',
                    message: this.$store.state.status.errorText,
                    closeButton: true,
                    progressBar: true,
                    timeOut: 15000
                })
                this.$router.push({ name: 'qvErrorFatal' })
            } else if (errorType === 2) {
                this.$toast.error({
                    title: 'Warning!',
                    message: this.$store.state.status.errorText,
                    closeButton: true,
                    progressBar: true,
                    timeOut: 7000
                })
            } else if (errorType === 5) {
                this.$toast.success({
                    title: '',
                    message: this.$store.state.status.errorText,
                    closeButton: true,
                    progressBar: true,
                    timeOut: 5000
                })
            }
        })
    //
    }
}
</script>
