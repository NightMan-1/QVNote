<template>
    <div class="container h-100">
        <div class="row h-100 justify-content-center align-items-center">
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
                                        </div>

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
    data () {
        return {
            errorData: {
                error: false,
                errorText: ''
            },
            loader: false,
            formData: {
                sourceFolder: this.$store.state.config.sourceFolder,
                sourceFolderCreateIfNotExist: true
            }

        }
    },
    methods: {
        saveChanges (index) {
            this.loader = true

            fetch(this.$store.getters.apiFolder + '/config.write', { method: 'POST', body: JSON.stringify(this.formData) }).then(response => { return response.json() })
                .then(jsonData => {
                    this.errorData.error = jsonData.error
                    this.errorData.errorText = jsonData.errorText
                    if (!this.errorData.error) {
                        this.$router.push('/')
                    }
                })
                .catch(error => {
                    console.error('Error save installation data:', error)
                })

            this.loader = false
        }
    }
}
</script>
