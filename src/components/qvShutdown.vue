<template>
    <div class="justify-content-center d-flex">
        <div class="alert alert-info w-50 text-center p-3">
            <h4 class="m-0">{{$t('general.shutdownMessage')}}</h4>
        </div>
    </div>
</template>

<script>
export default {
  name: 'qvShutdown',
  mounted: function () {
    document.body.classList.add('app', 'flex-row', 'align-items-center')
    let thisGlobal = this
    setInterval(function () {
      thisGlobal.$http.get(thisGlobal.$store.getters.apiFolder + '/ping').then(response => {
        // console.log(response.body.result)
        if (response.body.result === 'pong') {
          thisGlobal.$router.push('/app')
        }
      }, response => {
      })
    }, 2000)
  },
  destroyed: function () {
    document.body.className = ''
  },
  watch: {
  }
}
</script>
