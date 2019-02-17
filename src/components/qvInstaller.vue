<template>
	<div>
		<qv-loader v-if="loader"></qv-loader>
		<div class="container">
			<div class="row justify-content-center">
				<div class="col-md-8">
					<div class="card-group">
						<div class="card p-4 animated- fadeIn-">
							<div class="card-body">
								<form content="multipart/form-data" id="formInstaller">
									<h1>{{$t('installer.title')}}</h1>
									<div class="alert alert-danger" v-if="errorData.error">{{ errorData.errorText }}</div>
									<p class="text-muted">{{$t('installer.selectDataFolder')}}:</p>

									<div class="input-group mb-3">
										<div class="input-group-prepend">
											<span class="input-group-text">
												<i class="far fa-folder-open"></i>
											</span>
										</div>
										<input class="form-control" name="sourceFolder" v-model="formData.sourceFolder">
										<!--
										<div class="input-group-append">
											<button class="btn btn-primary" type="button" onclick="document.querySelector('#selectFolderDialog').click();"><i class="far fa-folder-open"></i></button>
										</div>
										-->
									</div>
									<!--
									<input id="selectFolderDialog" v-on:change="folderSelected" ref="selectFolderDialog" type="file" style="display: none;" webkitdirectory mozdirectory msdirectory odirectory directory>
									-->

									<div class="form-group form-check">
										<input type="checkbox" class="form-check-input" id="sourceFolderCreateIfNotExist" name="sourceFolderCreateIfNotExist" v-model="formData.sourceFolderCreateIfNotExist">
										<label class="form-check-label" for="sourceFolderCreateIfNotExist">{{$t('installer.sourceFolderCreateIfNotExist')}}</label>
									</div>

									<div class="row">
										<div class="col-6">
											<button type="button" class="btn btn-primary px-4" @click="saveChanges">
												{{$t('general.buttonSave')}}
											</button>
										</div>
									</div>
								</form>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
export default {
  name: 'qvInstaller',
  created: function () {
    // прямой переход, еще нечего не инициализировано
    if (Object.keys(this.$route.params).length === 0) {
      this.$router.push('/')
    } else {
      // console.log(this.$route.params);
    }
  },
  mounted: function () {
    document.body.classList.add('app', 'flex-row', 'align-items-center')
  },
  destroyed: function () {
    document.body.className = ''
  },
  data () {
    return {
      errorData: {
        error: false,
        errorText: ''
      },
      loader: false,
      // files: [],
      formData: {
        sourceFolder: this.$store.getters.getConfig.sourceFolder,
        sourceFolderCreateIfNotExist: true
      }

    }
  },
  methods: {
    folderSelected (files) {
      // console.log(files);
      // console.log(this.$refs.selectFolderDialog.files);

    },
    saveChanges (index) {
      this.loader = true
      this.$http.post(this.$store.getters.apiFolder + '/config.write', this.formData, { method: 'PUT' }).then(response => {
        const responseData = response.body
        this.errorData.error = responseData.error
        this.errorData.errorText = responseData.errorText
        // console.log(responseData);
        if (!this.errorData.error) {
          this.$router.push('/app')
        }
      }, response => {
      })

      this.loader = false
    }
  }
}
</script>
