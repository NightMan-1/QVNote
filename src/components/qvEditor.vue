<template>
    <div id="qv-editor">
        <div id="qv-editor-header">
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

        <div id="qv-editor-main">
                <input type="text" class="form-control mb-2 mt-3 text-dark font-size-normal"
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
                <div class="editor mt-2" v-if="articleCurrentEditable.type === 'text'">
                    <quill-editor v-model="articleCurrentEditable.content" :options="editorSettings"></quill-editor>
                </div>
                <div class="editor prism mt-2" v-if="articleCurrentEditable.type === 'code'">
                    <prism-editor v-model="articleCurrentEditable.content" language="html" :line-numbers="true"></prism-editor>
                </div>
                <div class="clearfix"></div>
        </div>
    </div>
</template>

<script>
import mixin from './mixins'
import 'prismjs'
import 'prismjs/themes/prism.css'
import PrismEditor from 'vue-prism-editor'
import { quillEditor } from 'vue-quill-editor'
import Multiselect from 'vue-multiselect'
import Quill from 'quill'
import '../style/quill.snow.css'
import { ImageDrop } from 'quill-image-drop-module'
import ImageResize from 'quill-image-resize-module'
Quill.register('modules/imageDrop', ImageDrop)
Quill.register('modules/imageResize', ImageResize)
let BeautifyHtml = require('js-beautify').html

export default {
    name: 'qvEditor',
    mixins: [mixin],
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
            editorSettings: {
                modules: {
                    toolbar: {
                        container: [
                            [{ 'header': [1, 2, 3, 4, 5, 6, false] }],
                            ['bold', 'italic', 'underline'],
                            ['blockquote', 'code-block'],
                            [{ 'list': 'ordered' }, { 'list': 'bullet' }],
                            [{ 'color': [] }, { 'background': [] }],
                            [{ 'align': [] }],
                            ['clean'],
                            ['link', 'image'], [], ['eraser', 'fullscreen']
                        ],
                        handlers: {
                            'eraser': () => {
                                fetch(this.$store.getters.apiFolder + '/cleanup_html.json',
                                    { method: 'POST',
                                        body: JSON.stringify({ 'content': this.articleCurrentEditable.content }) }).then(response => { return response.json() })
                                    .then(jsonData => {
                                        this.articleCurrentEditable.content = jsonData.content
                                    })
                                    .catch(error => {
                                        console.error('Error cleanup html:', error)
                                    })
                            },
                            'fullscreen': () => {
                                if (!document.fullscreenElement) {
                                    // console.log('requesting fullscreen')
                                    document.querySelector('.quill-editor').requestFullscreen()
                                } else {
                                    if (document.exitFullscreen) {
                                        // console.log('disable fullscreen')
                                        document.exitFullscreen()
                                    }
                                }
                            }
                        }
                    },
                    imageDrop: true,
                    imageResize: { modules: [ 'Resize', 'DisplaySize' ] } // + 'Toolbar'
                }
            }
        }
    },
    components: {
        quillEditor,
        Multiselect,
        PrismEditor
    },
    mounted () {
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
        saveData () {
            fetch(this.$store.getters.apiFolder + '/note_edit.json',
                { method: 'POST',
                    body: JSON.stringify({
                        'title': this.articleCurrentEditable.title,
                        'url': this.articleCurrentEditable.url_src,
                        'uuid': this.articleCurrentEditable.uuid,
                        'type': this.articleCurrentEditable.editorType,
                        'tags': this.articleCurrentEditable.tags,
                        'content': this.articleCurrentEditable.content
                    }) }).then(response => { return response.json() })
                .then(jsonData => {
                    this.articleCurrentEditable.uuid = jsonData.uuid
                    this.articleCurrentEditable.NoteBookUUID = jsonData.NoteBookUUID
                    // this.articleCurrentEditable.content = jsonData.html // slow
                    this.$store.dispatch('getAllData')
                })
                .catch(error => {
                    console.error('Error save note data:', error)
                    this.$store.commit('setStatus', { errorType: 2, errorText: this.$t('editor.errorSave') })
                })
        },
        addTag (newTag) {
            this.articleCurrentEditable.tags.push(newTag)
            this.$refs.editorTags.$el.focus()
        }
    },
    watch: {
        'articleCurrentEditable.type' () {
            if (this.articleCurrentEditable.type === 'code') {
                this.articleCurrentEditable.content = BeautifyHtml(this.articleCurrentEditable.content)
            }
        }
    }

}
</script>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>

<style scope>
    /* purgecss start ignore */
    #qv-editor{
        padding: .5rem 1.5rem 2rem;
    }
    #qv-editor-header{
        position: fixed;
        width: calc(100% - var(--sidebar-width));
        top: 0;
        left:var(--sidebar-width);
        z-index: 1000;
        background-color: var(--nord6);
        border-bottom: 1px solid var(--nord4);
        padding: .5rem .75rem;
    }
    #qv-editor-main{
        overflow: hidden;
    }

    .font-size-normal {
        font-size: 1em;
    }

    .editor {
      min-height: 40rem !important;
    }

    .editor.prism {
      background-color: rgb(240, 243, 245);
      height: 70vh;
      overflow: auto;
    }

    .vmd-body {
        min-height: 40rem !important;
    }

    .quill-editor:-moz-full-screen {padding: 1rem;background: white;}
    .quill-editor:-webkit-full-screen {padding: 1rem;background: white; }
    .quill-editor:-ms-fullscreen {padding: 1rem;background: white; }

    .quill-editor:-moz-full-screen .ql-container {max-height: calc(100vh - 5rem);}
    .quill-editor:-webkit-full-screen .ql-container {max-height: calc(100vh - 5rem);}
    .quill-editor:-ms-full-screen .ql-container {max-height: calc(100vh - 5rem);}

    .quill-editor:-moz-full-screen .ql-editor {max-height: 100%;}
    .quill-editor:-webkit-full-screen .ql-editor {max-height: 100%;}
    .quill-editor:-ms-full-screen .ql-editor {max-height: 100%;}

    .ql-editor {
        /*white-space: normal !important;*/
        max-height: calc(100vh - 18.5rem);
    }
    .ql-container{
        font-family: 'Roboto', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
        font-kerning: normal;
        font-variant-ligatures: common-ligatures contextual;
        font-feature-settings: "kern", "liga", "clig", "calt";
        font-size: 16px;
        font-weight: 400;
        line-height: 1.42857143;
        font-display: swap; /* or block */
        border-radius: 0 0 .25rem .25rem !important;
    }
    .ql-toolbar.ql-snow{border-radius: .25rem .25rem 0 0 !important;}

    .ql-snow .ql-editor pre.ql-syntax {
        color: black !important;
    }

    .ql-fullscreen, .ql-eraser { padding: 0 !important; }
    .ql-fullscreen:after, .ql-eraser:after {
        font-family: 'Font Awesome 5 Free';
        content: "\f31e";
        font-weight: 900;
        font-size: .9rem;
    }
    .ql-eraser:after{
        content: "\f12d";
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
        /*background: var(--nord14);*/
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
    .prism-editor__code, pre[class*="language-"] { overflow: inherit; margin: 0;}
    .prism-editor__code code { padding: 0; overflow: inherit;}
    .prism-editor__line-numbers {
        width: 2rem;
        float: left;
        margin-top: 0 !important;
        text-align: right;
        margin-right: 0.5rem;
    }

    .ql-editor blockquote, .ql-editor h1, .ql-editor h2, .ql-editor h3, .ql-editor h4, .ql-editor h5, .ql-editor h6, .ql-editor ol, .ql-editor p, .ql-editor pre, .ql-editor ul {
        margin-bottom: 1em !important;
    }

    /* purgecss end ignore */
</style>
