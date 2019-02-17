<template>
    <div class="row scrolly">
        <div class="col-12 scrolly-col">
            <div class="breadcrumb- d-flex p-2 bg-light b-b-1 mb-2">
                <h4 class="mt-1 mb-0">Создание заметки</h4>
                <button class="btn btn-outline-primary ml-auto"
                        @click="$router.push('/notes/' + notebook_uuid + '/' + note_uuid)"
                        v-if="(note_uuid !== '' && notebook_uuid !== '')"><i class="fas fa-eye"></i></button>
                <button class="btn btn-outline-success ml-2" @click="saveData"
                        :class="{'ml-auto':note_uuid === '' && notebook_uuid === ''}"><i class="fas fa-save"></i>
                </button>
                <button class="btn btn-outline-secondary ml-2"><i class="fas fa-eraser text-muted"></i></button>
                <div class="btn-group ml-4" role="group" aria-label="Button group">
                    <button class="btn btn-outline-secondary" :class="{'active':editorType === 'tinymce'}"
                            @click="editorType = 'tinymce'"><i class="fas fa-edit"></i></button>
                    <button class="btn btn-outline-secondary" :class="{'active':editorType === 'markdown'}"
                            @click="editorType = 'markdown'"><i class="fas fa-columns"></i></button>
                    <button class="btn btn-outline-secondary" :class="{'active':editorType === 'code'}"
                            @click="editorType = 'code'"><i class="fas fa-code"></i></button>
                </div>

            </div>

            <div class="pr-3 pl-3 pb-5">
                <input type="text" class="form-control mb-3 mt-3 text-dark" placeholder="Заголовок записи..."
                       id="editorTitle" v-model="title"/>
                <div class="input-group mb-3 mt-3">
                    <div class="input-group-prepend">
                        <span class="input-group-text"><i class="fas fa-external-link-alt"></i></span>
                    </div>
                    <input type="text" class="form-control text-dark" placeholder="Ссылка..." v-model="src_url"/>
                </div>
                <!--<div class="well bg-light p-3 mt-3 mb-3">{{content}}</div>-->
                <div class="editor">
                    <textarea id="tinyEditor" class="d-none"></textarea>
                    <editor v-model="content" @init="editorInit" lang="html" theme="crimson_editor" height="657"
                            class="mt-3" v-if="editorType === 'code'"></editor>
                    <vue-editor v-model="content" v-if="editorType === 'markdown'"></vue-editor>
                </div>

                <div class="clearfix"></div>
            </div>

        </div>
    </div>

</template>

<script>
// const beautify = require('beautify')
// const sanitizeHtml = require('sanitize-html') // https://www.npmjs.com/package/sanitize-html
// const ting = require('ting') // https://www.npmjs.com/package/ting
// const xss = require('xss')
export default {
  name: 'qvEditor',
  props: ['noteUUID'],
  data () {
    return {
      content: ' ',
      title: '',
      src_url: '',
      note_uuid: '',
      notebook_uuid: '',
      editorType: 'tinymce'
    }
  },
  components: {
    editor: require('vue2-ace-editor')
  },
  created () {
  },
  mounted () {
    if (this.$route.params.noteUUID !== '') {
      this.content = this.$qvGlobalData.articleCurrent.content
      // this.content = xss(this.$qvGlobalData.articleCurrent.content)
      // this.content = beautify(sanitizeHtml(this.$qvGlobalData.articleCurrent.content, { allowedTags: sanitizeHtml.defaults.allowedTags.concat([ 'img' ]) }), {format: 'html'})
      // this.content = beautify(sanitizeHtml(this.$qvGlobalData.articleCurrent.content, { allowedTags: sanitizeHtml.defaults.allowedTags.concat([ 'img' ]) }), {format: 'html'})
      // this.content = beautify(this.$qvGlobalData.articleCurrent.content, {format: 'html'})
      // this.content = beautify(ting.sanitize(this.$qvGlobalData.articleCurrent.content), {format: 'html'})
      this.title = this.$qvGlobalData.articleCurrent.title
      this.note_uuid = this.$qvGlobalData.articleCurrent.uuid
      this.notebook_uuid = this.$qvGlobalData.articleCurrent.NoteBookUUID
      this.src_url = this.$qvGlobalData.articleCurrent.src_url
      if (this.$qvGlobalData.articleCurrent.type === 'text') {
        this.editorType = 'tinymce'
      } else if (this.$qvGlobalData.articleCurrent.type === 'code') {
        this.editorType = 'code'
      } else {
        this.editorType = 'markdown'
      }
    }
    if (this.editorType === 'tinymce') {
      this.tinyInit()
    }
  },
  beforeDestroy () {
    if (tinymce.get('tinyEditor') != null) {
      tinymce.get('tinyEditor').destroy()
      // document.getElementById('tinyEditor').classList.add('d-none')
    }
  },
  watch: {
    'editorType' () {
      if (this.editorType === 'tinymce') {
        this.tinyInit()
      } else {
        if (tinymce.get('tinyEditor') != null) {
          tinymce.get('tinyEditor').destroy()
          document.getElementById('tinyEditor').classList.add('d-none')
        }
      }
    }
  },
  computed: {},
  methods: {
    editorInit: function () {
      require('brace/ext/language_tools') // language extension prerequsite...
      require('brace/mode/html')
      require('brace/mode/php')
      require('brace/mode/markdown')
      require('brace/mode/javascript') // language
      require('brace/mode/less')
      require('brace/theme/crimson_editor')
      require('brace/snippets/javascript') // snippet
    },
    tinyInit () {
      /* eslint no-undef: "off" */
      document.getElementById('tinyEditor').classList.remove('d-none')
      let self = this
      tinymce.init({
        selector: '#tinyEditor',
        language: 'ru',
        image_advtab: true,
        browser_spellcheck: true,
        plugins: 'advlist anchor autolink charmap colorpicker contextmenu directionality help hr image imagetools insertdatetime legacyoutput link lists media nonbreaking noneditable paste searchreplace tabfocus table textcolor textpattern visualblocks visualchars',
        toolbar: ['undo redo | formatselect | bold italic | forecolor backcolor | template codesample | alignleft aligncenter alignright alignjustify | bullist numlist | outdent indent | link image'],
        menubar: 'edit insert view format table tools help',
        min_height: 557,
        // paste_use_dialog: true,
        // paste_auto_cleanup_on_paste: true,
        // paste_convert_headers_to_strong: false,
        // paste_strip_class_attributes: 'all',
        // paste_remove_spans: true,
        // paste_remove_styles: true,
        // paste_retain_style_properties: '',
        setup: function (editor) {
          editor.on('init', function () {
            tinymce.get('tinyEditor').setContent(self.content)
          })
          editor.on('keyup', function () {
            // self.content = tinymce.get('tinyEditor').getContent()
            // console.log('Editor contents was changed v2')
          })
          editor.on('Change', function () {
            self.content = tinymce.get('tinyEditor').getContent()
            // console.log('Editor contents was changed.')
          })
        }
      })
    },
    saveData () {
      var thisGlobal = this
      this.$http.post(this.$store.getters.apiFolder + '/note_edit.json', {
        'title': this.title,
        'url': this.src_url,
        'uuid': this.note_uuid,
        'type': this.editorType,
        'content': this.content
      }, { method: 'PUT' }).then(response => {
        const thisResponse = response.body
        thisGlobal.note_uuid = thisResponse.uuid
        thisGlobal.notebook_uuid = thisResponse.NoteBookUUID
      }, response => {
        this.$toast.error({
          title: 'Error!',
          message: 'Error save note...',
          closeButton: true,
          progressBar: true,
          timeOut: 7000
        })
      })
    }
  }

}
</script>

<style>
    #editorTitle {
        font-size: 1.2em;
    }

    .editor {
        min-height: 40rem !important;
    }

    .vmd-body {
        min-height: 40rem !important;
    }
</style>
