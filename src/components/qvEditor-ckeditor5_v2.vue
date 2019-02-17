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
                        <ckeditor :editor="ckEditor" v-model="articleCurrentEditable.content" :config="ckEditorConfig"></ckeditor>
                        <vue-editor
                            :customModules="customModulesForEditor"
                            :editorOptions="quillEditorSettings"
                            v-model="articleCurrentEditable.content" v-if="articleCurrentEditable.type === 'text1'"></vue-editor>
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
/*
import CKEditor from '@ckeditor/ckeditor5-vue'
Vue.use(CKEditor)

        "@ckeditor/ckeditor5-alignment": "^10.0.3",
        "@ckeditor/ckeditor5-autoformat": "^10.0.3",
        "@ckeditor/ckeditor5-basic-styles": "^10.0.3",
        "@ckeditor/ckeditor5-block-quote": "^10.1.0",
        "@ckeditor/ckeditor5-build-classic": "^11.2.0",
        "@ckeditor/ckeditor5-dev-utils": "^11.0.1",
        "@ckeditor/ckeditor5-dev-webpack-plugin": "^7.0.1",
        "@ckeditor/ckeditor5-easy-image": "^10.0.3",
        "@ckeditor/ckeditor5-editor-classic": "^11.0.1",
        "@ckeditor/ckeditor5-essentials": "^10.1.2",
        "@ckeditor/ckeditor5-heading": "^10.1.0",
        "@ckeditor/ckeditor5-image": "^11.0.0",
        "@ckeditor/ckeditor5-link": "^10.0.4",
        "@ckeditor/ckeditor5-list": "^11.0.2",
        "@ckeditor/ckeditor5-media-embed": "^10.0.0",
        "@ckeditor/ckeditor5-paragraph": "^10.0.3",
        "@ckeditor/ckeditor5-table": "^11.0.0",
        "@ckeditor/ckeditor5-theme-lark": "^11.1.0",
        "@ckeditor/ckeditor5-vue": "^1.0.0-beta.1",
*/

import { VueEditor } from 'vue2-editor'
import { ImageDrop } from 'quill-image-drop-module'
import ImageResize from 'quill-image-resize-module'
import Multiselect from 'vue-multiselect'
import 'prismjs'
import 'prismjs/themes/prism.css'
import PrismEditor from 'vue-prism-editor'

import ClassicEditor from '@ckeditor/ckeditor5-build-classic'
/*
import ClassicEditor from '@ckeditor/ckeditor5-editor-classic/src/classiceditor'

import Essentials from '@ckeditor/ckeditor5-essentials/src/essentials'
import Autoformat from '@ckeditor/ckeditor5-autoformat/src/autoformat'
import Bold from '@ckeditor/ckeditor5-basic-styles/src/bold'
import Italic from '@ckeditor/ckeditor5-basic-styles/src/italic'
import BlockQuote from '@ckeditor/ckeditor5-block-quote/src/blockquote'
import Subscript from '@ckeditor/ckeditor5-basic-styles/src/subscript'
import Superscript from '@ckeditor/ckeditor5-basic-styles/src/superscript'
import Code from '@ckeditor/ckeditor5-basic-styles/src/code'
import Underline from '@ckeditor/ckeditor5-basic-styles/src/underline'
import Strikethrough from '@ckeditor/ckeditor5-basic-styles/src/strikethrough'
import Heading from '@ckeditor/ckeditor5-heading/src/heading'
import Image from '@ckeditor/ckeditor5-image/src/image'
import ImageCaption from '@ckeditor/ckeditor5-image/src/imagecaption'
import ImageStyle from '@ckeditor/ckeditor5-image/src/imagestyle'
import ImageToolbar from '@ckeditor/ckeditor5-image/src/imagetoolbar'
import Link from '@ckeditor/ckeditor5-link/src/link'
import List from '@ckeditor/ckeditor5-list/src/list'
import Paragraph from '@ckeditor/ckeditor5-paragraph/src/paragraph'
import Alignment from '@ckeditor/ckeditor5-alignment/src/alignment'
import Table from '@ckeditor/ckeditor5-table/src/table'
import TableToolbar from '@ckeditor/ckeditor5-table/src/tabletoolbar'
import MediaEmbed from '@ckeditor/ckeditor5-media-embed/src/mediaembed'
*/
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
      quillEditorSettings: {
        modules: {
          imageDrop: true,
          imageResize: { modules: [ 'Resize', 'DisplaySize' ] } // + 'Toolbar'
        }
      },
      ckContent: '',
      ckEditor: ClassicEditor,
      ckEditorConfig: {},
      ckEditorConfigFull: {
        plugins: [
          Table,
          TableToolbar,
          Essentials,
          Autoformat,
          Bold,
          Italic,
          BlockQuote,
          Heading,
          Image,
          ImageCaption,
          ImageStyle,
          ImageToolbar,
          Link,
          List,
          Paragraph,
          Alignment,
          MediaEmbed,
          Code,
          Underline,
          Strikethrough,
          Subscript,
          Superscript
        ],

        toolbar: {
          items: [
            'heading',
            '|',
            'bold',
            'italic',
            'underline',
            'strikethrough',
            'subscript',
            'superscript',
            'link',
            'bulletedList',
            'numberedList',
            '|',
            'alignment:left', 'alignment:right', 'alignment:center', 'alignment:justify',
            '|',
            'code',
            'blockQuote',
            'insertTable',
            'mediaEmbed',
            '|',
            'undo',
            'redo'
          ]
        },
        language: 'ru',
        image: {
          toolbar: [
            'imageStyle:full',
            'imageStyle:side',
            '|',
            'imageTextAlternative'
          ]
        },
        table: {
          contentToolbar: ['tableColumn', 'tableRow', 'mergeTableCells']
        },
        heading: {
          options: [
            { model: 'paragraph', title: 'Paragraph', class: 'ck-heading_paragraph' },
            { model: 'heading1', view: 'h1', title: 'Heading 1', class: 'ck-heading_heading1' },
            { model: 'heading2', view: 'h2', title: 'Heading 2', class: 'ck-heading_heading2' },
            { model: 'heading3', view: 'h3', title: 'Heading 3', class: 'ck-heading_heading3' },
            { model: 'heading4', view: 'h4', title: 'Heading 4', class: 'ck-heading_heading4' }
          ]
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
    // this.ckContent = '<p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p><p>test</p>'
    // console.log(this.articleCurrentEditable.content.toString())
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

    .ck-content code{
        white-space: pre;
        display: inline-block;
    }

    /* purgecss end ignore */
</style>
