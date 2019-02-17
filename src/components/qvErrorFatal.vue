<template>
    <div class="justify-content-center d-flex">
        <div class="alert alert-danger w-50 text-center p-5">
            <h4>{{$t('general.messageErrorFatal')}}</h4>
            <p class="text-dark mb-0 mt-4" v-if="errorText">
                {{ errorText }}
            </p>
        </div>
    </div>
</template>

<script>
export default {
  name: 'qvErrorFatal',
  data () {
    return {
      errorText: ''
    }
  },
  mounted: function () {
    document.body.classList.add('app', 'flex-row', 'align-items-center')
    this.errorText = this.$store.getters.getStatus.errorText
  },
  destroyed: function () {
    document.body.className = ''
  },
  watch: {
    '$route' (to, from) {
      this.errorText = this.$store.getters.getStatus.errorText
    }
  }
}
</script>
