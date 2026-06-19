import { computed } from 'vue'
import { useNoteStore } from '../store'

export function useNoteMixin () {
    const store = useNoteStore()

    return {
        gridClass: computed(() => store.gridClass),
        pageType: computed(() => store.pageType),
        sidebarType: computed(() => store.sidebarType),
        notesCountInbox: computed(() => store.notesCountInbox),
        notesCountTrash: computed(() => store.notesCountTrash),
        notesCountTotal: computed(() => store.notesCountTotal),
        notesCountFavorites: computed(() => store.notesCountFavorites),
        currentNotebookID: computed(() => store.currentNotebookID),
        notebooksList: computed(() => store.notebooksList),
        notesList: computed(() => store.notesList),
        tagsList: computed(() => store.tagsList),
        currentTagURL: computed(() => store.currentTagURL),
        articleCurrent: computed(() => store.currentArticle),
        showAdvancedInfo: computed(() => store.showAdvancedNoteInfo),
        readerMode: computed(() => store.readerMode),
        layoutBig: computed(() => store.layoutBig),
        notebookCount: computed(() => store.getNotebooksCount),
        tagsCount: computed(() => store.getTagsCount),
        config: computed(() => store.config)
    }
}
