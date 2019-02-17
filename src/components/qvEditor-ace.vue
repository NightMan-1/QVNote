<template>
    <div class="qv-editor">
        <div class="p-2 bg-light b-b-1 mb-2" id="qv-editor-header">
            <h4 class="mt-1 mb-0 float-left">Создание заметки</h4>
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
                <div class="pl-2 pr-2 mb-4">
                    <input type="text" class="form-control mb-3 mt-3 text-dark font-size-normal"
                           placeholder="Заголовок записи..." v-model="articleCurrentEditable.title" ref='editorTitle'/>
                    <div class="row">
                        <div class="col-6">
                            <div class="form-group">
                                <label><b>Ссылка</b></label>
                                <div class="input-group">
                                    <div class="input-group-prepend">
                                        <span class="input-group-text"><i class="fas fa-external-link-alt"></i></span>
                                    </div>
                                    <input type="text" class="form-control text-dark font-size-normal"
                                           placeholder="Добавьте ссылку тут..." v-model="articleCurrentEditable.url_src"/>
                                </div>
                            </div>
                        </div>
                        <div class="col-6">
                            <div class="form-group">
                                <label><b>Теги</b></label>
                                <multiselect
                                    v-model="articleCurrentEditable.tags"
                                    tag-placeholder="Дабавьте новый тег..."
                                    placeholder="Поиск или добавление тега"
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
                    <!--<div class="well bg-light p-3 mt-3 mb-3">{{content}}</div>-->
                    <div class="editor mt-3">
                        <vue-editor v-model="articleCurrentEditable.content" v-if="articleCurrentEditable.type === 'text'"></vue-editor>
                        <editor v-model="articleCurrentEditable.content" @init="editorInit" lang="html" theme="crimson_editor" height="657"
                                class="mt-3" v-if="articleCurrentEditable.type === 'code'"></editor>
                        <div class="clearfix"></div>
                    </div>
                </div>

        </div>
    </div>
</template>

<script>
import { VueEditor } from 'vue2-editor'
import Multiselect from 'vue-multiselect'

// const SimpleScrollbar = require('simple-scrollbar')
// const sanitizeHtml = require('sanitize-html') // https://www.npmjs.com/package/sanitize-html
// const ting = require('ting') // https://www.npmjs.com/package/ting

export default {
  name: 'qvEditor',
  props: ['noteUUID'],
  data () {
    return {
      multiselectLang: {
        selectLabel: 'Нажмите для выбора',
        deselectLabel: 'Нажмите для удаление',
        selectedLabel: 'Выбрано'
      },
      articleCurrentEditable: { title: '', uuid: '', NoteBookUUID: '', status: '', tags: [], CreatedDate: '', UpdatedDate: '', cells: {}, content: '', type: 'text', url_src: '' },
      tagsListFormatted: []
    }
  },
  components: {
    editor: require('vue2-ace-editor'),
    VueEditor,
    Multiselect
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
      require('brace/ext/language_tools') // language extension prerequsite...
      require('brace/mode/html')
      require('brace/mode/php')
      require('brace/mode/markdown')
      require('brace/mode/javascript') // language
      require('brace/mode/less')
      require('brace/theme/crimson_editor')
      require('brace/snippets/javascript') // snippet
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
        this.$store.commit('setStatus', { errorType: 2, errorText: 'Error save note...' })
      })
    },
    addTag (newTag) {
      this.articleCurrentEditable.tags.push(newTag)
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
    #qv-editor-header{
        position: fixed;
        width: calc(100% - 200px);
        top: 0;
        z-index: 1000;
    }
    #qv-editor-main{
        margin-top: 4.5rem;
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
        padding: 0.425em 3rem 0 .375em;
        min-height: 1rem;
        border-color: #dcdfe2;
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
        padding-left: 0.375em;
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

    /* purgecss end ignore */
</style>
