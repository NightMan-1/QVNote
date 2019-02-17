<template>
    <div id="qv-editor">
        <div class="pt-2 pb-2 pl-3 pr-3 bg-light b-b-1 mb-2" id="qv-editor-header">
            <h4 class="mt-1 mb-0 float-left" v-if="(articleCurrentEditable.uuid === '')">{{$t('editor.titleNew')}}</h4>
            <h4 class="mt-1 mb-0 float-left" v-if="(articleCurrentEditable.uuid !== '')">{{$t('editor.titleExist')}}</h4>
            <div class="float-right">
                <button class="btn btn-outline-primary ml-auto"
                        @click="$router.push('/notes/' + articleCurrentEditable.NoteBookUUID + '/' + articleCurrentEditable.uuid)"
                        v-if="(articleCurrentEditable.uuid !== '' && articleCurrentEditable.NoteBookUUID !== '')"><i class="fas fa-eye"></i></button>
                <button class="btn btn-outline-success ml-2" @click="saveData"
                        :class="{'ml-auto':articleCurrentEditable.uuid === '' && articleCurrentEditable.NoteBookUUID === ''}"><i class="fas fa-save"></i>
                </button>
                <!--<button class="btn btn-outline-secondary ml-2"><i class="fas fa-eraser text-muted"></i></button>-->
                <div class="btn-group ml-4" role="group" aria-label="Button group">
                    <button class="btn btn-outline-secondary" :class="{'active':articleCurrentEditable.type === 'text'}"
                            @click="articleCurrentEditable.type = 'text'"><i class="fas fa-edit"></i></button>
                    <!--<button class="btn btn-outline-secondary" :class="{'active':editorType === 'markdown'}" @click="editorType = 'markdown'"><i class="fas fa-columns"></i></button>-->
                    <button class="btn btn-outline-secondary" :class="{'active':articleCurrentEditable.type === 'code'}"
                            @click="articleCurrentEditable.type = 'code'"><i class="fas fa-code"></i></button>
                </div>
            </div>
        </div>

        <div class="pl-0" id="qv-editor-main">
            <simplebar class="simplebarHeight" data-simplebar-auto-hide="true">
                <div class="pl-3 pr-4 mb-4">
                    <input type="text" class="form-control mb-3 mt-3 text-dark font-size-normal"
                           :placeholder="$t('editor.inputTitlePlaceholder')" v-model="articleCurrentEditable.title" ref='editorTitle'/>
                    <div class="row">
                        <div class="col-6">
                            <div class="form-group">
                                <label><b>{{$t('editor.titleURL')}}</b></label>
                                <div class="input-group">
                                    <div class="input-group-prepend">
                                        <span class="input-group-text"><i class="fas fa-external-link-alt"></i></span>
                                    </div>
                                    <input type="text" class="form-control text-dark font-size-normal"
                                           :placeholder="$t('editor.inputURLPlaceholder')" v-model="articleCurrentEditable.url_src"/>
                                </div>
                            </div>
                        </div>
                        <div class="col-6">
                            <div class="form-group">
                                <label><b>{{$t('editor.titleTags')}}</b></label>
                                <multiselect
                                    ref='editorTags'
                                    v-model="articleCurrentEditable.tags"
                                    :placeholder="$t('editor.inputTagsPlaceholder')"
                                    :options="tagsListFormatted"
                                    :multiple="true"
                                    :taggable="true"
                                    @tag="addTag"
                                    :selectLabel="multiselectLang.selectLabel"
                                    :deselectLabel="multiselectLang.deselectLabel"
                                    :selectedLabel="multiselectLang.selectedLabel"></multiselect>
                            </div>

                        </div>
                    </div>
                    <div class="editor mt-3">
                        <vue-editor :customModules="customModulesForEditor" :editorOptions="editorSettings" v-model="articleCurrentEditable.content" v-if="articleCurrentEditable.type === 'text'"></vue-editor>
                        <prism-editor v-model="articleCurrentEditable.content" language="html" :line-numbers="true"
                            class="mt-3"
                            v-if="articleCurrentEditable.type === 'code'"></prism-editor>
                        <div class="clearfix"></div>
                    </div>
                </div>
            </simplebar>
        </div>
    </div>
</template>

<script>
import { VueEditor } from 'vue2-editor'
import { ImageDrop } from 'quill-image-drop-module'
import ImageResize from 'quill-image-resize-module'
import Multiselect from 'vue-multiselect'
import 'prismjs'
import 'prismjs/themes/prism.css'
import PrismEditor from 'vue-prism-editor'

// const SimpleScrollbar = require('simple-scrollbar')
// const sanitizeHtml = require('sanitize-html') // https://www.npmjs.com/package/sanitize-html
// const ting = require('ting') // https://www.npmjs.com/package/ting

export default {
  name: 'qvEditor',
  props: ['noteUUID'],
  data () {
    return {
      multiselectLang: {
        selectLabel: this.$t('editor.multiselectLang.selectLabel'),
        deselectLabel: this.$t('editor.multiselectLang.deselectLabel'),
        selectedLabel: this.$t('editor.multiselectLang.selectedLabel')
      },
      articleCurrentEditable: { title: '', uuid: '', NoteBookUUID: '', status: '', tags: [], CreatedDate: '', UpdatedDate: '', cells: {}, content: '', type: 'text', url_src: '' },
      tagsListFormatted: [],
      customModulesForEditor: [
        { alias: 'imageDrop', module: ImageDrop },
        { alias: 'imageResize', module: ImageResize }
      ],
      editorSettings: {
        modules: {
          imageDrop: true,
          imageResize: { modules: [ 'Resize', 'DisplaySize' ] } // + 'Toolbar'
        }
      }
    }
  },
  components: {
    VueEditor,
    Multiselect,
    PrismEditor
  },
  created () {
  },
  mounted () {
    if (this.BrowserPlatform === 'general') {
      // SimpleScrollbar.initAll()
    }
    this.$refs.editorTitle.focus()
    // this.articleCurrentEditable = Object.assign({}, this.articleCurrent)
    this.articleCurrentEditable = JSON.parse(JSON.stringify(this.articleCurrent))
    if (this.articleCurrentEditable.type === '') {
      this.articleCurrentEditable.type = 'text'
    }

    for (const tag in this.tagsList) {
      if (this.tagsList[tag].name !== '') {
        this.tagsListFormatted.push(this.tagsList[tag].name)
      }
    }
  },
  methods: {
    editorInit: function () {
    },
    saveData () {
      this.$http.post(this.$store.getters.apiFolder + '/note_edit.json', {
        'title': this.articleCurrentEditable.title,
        'url': this.articleCurrentEditable.url_src,
        'uuid': this.articleCurrentEditable.uuid,
        'type': this.articleCurrentEditable.editorType,
        'tags': this.articleCurrentEditable.tags,
        'content': this.articleCurrentEditable.content
      }, { method: 'PUT' }).then(response => {
        const thisResponse = response.body
        this.articleCurrentEditable.uuid = thisResponse.uuid
        this.articleCurrentEditable.NoteBookUUID = thisResponse.NoteBookUUID
      }, response => {
        this.$store.commit('setStatus', { errorType: 2, errorText: this.$t('editor.errorSave') })
      })
    },
    addTag (newTag) {
      this.articleCurrentEditable.tags.push(newTag)
      this.$refs.editorTags.$el.focus()
    }
  },
  computed: {
    articleCurrent () {
      return this.$store.getters.getCurrentArticle
    },
    tagsList () {
      return this.$store.getters.getTagsList
    },
    BrowserPlatform () {
      return this.$store.getters.getBrowserPlatform
    }
  }

}
</script>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>

<style>
    /* purgecss start ignore */
    #qv-editor{
        height: 100vh;
        overflow: hidden;
    }
    #qv-editor-header{
        position: fixed;
        width: calc(100% - 200px);
        top: 0;
        z-index: 1000;
    }
    #qv-editor-main{
        margin-top: 3.5rem;
        height: calc(100vh - 3.5rem);
    }

    .font-size-normal {
        font-size: 1em;
    }

    .editor {
        min-height: 40rem !important;
    }

    .vmd-body {
        min-height: 40rem !important;
    }

    .ql-editor- {
        white-space: normal !important;
    }

    .ql-snow .ql-editor pre.ql-syntax {
        color: black !important;
    }

    .multiselect {
        font-size: 1rem;
        font-weight: 400;
        min-height: auto;
    }

    .multiselect__tags {
        border-radius: 0.25rem;
        font-size: 1rem;
        font-weight: 400;
        line-height: 1.2;
        padding: .5em 0 0 .7em;
        min-height: 1rem;
        border-color: #dcdfe2;
        margin-bottom: 4px;
    }

    .multiselect__tags:focus {
        border-color: #8ad4ee;
        box-shadow: 0 0 0 .2rem rgba(32, 168, 216, .25);
    }

    .multiselect__tag {
        border-radius: 0.25rem;
        font-size: 0.9rem;
    }

    .multiselect__input, .multiselect__single {
        padding-left: 0;
        margin-bottom: 0.425em;
        font-size: 1rem;
        font-weight: 400;
    }

    .multiselect__single {
        color: rgba(0, 0, 0, 0.5);
    }

    .multiselect__select {
        width: 2.85rem;
        height: 2.3rem;
    }
    .multiselect__placeholder{
        margin-bottom: 0.53rem;
    }
    .prism-editor__code code { padding: 0; }
    .prism-editor__line-numbers {
        width: 2rem;
        float: left;
        margin-top: 0 !important;
        text-align: right;
    }

    .ql-editor blockquote, .ql-editor h1, .ql-editor h2, .ql-editor h3, .ql-editor h4, .ql-editor h5, .ql-editor h6, .ql-editor ol, .ql-editor p, .ql-editor pre, .ql-editor ul {
        margin-bottom: 1em !important;
    }

    /* purgecss end ignore */
</style>
