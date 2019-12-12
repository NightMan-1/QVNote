const mixin = {
    computed: {
        gridClass () {
            return this.$store.state.gridClass
        },
        pageType () {
            return this.$store.state.pageType
        },
        sidebarType () {
            return this.$store.state.sidebarType
        },
        notesCountInbox () {
            return this.$store.state.notesCountInbox
        },
        notesCountTrash () {
            return this.$store.state.notesCountTrash
        },
        notesCountTotal () {
            return this.$store.state.notesCountTotal
        },
        notesCountFavorites () {
            return this.$store.state.notesCountFavorites
        },
        currentNotebookID () {
            return this.$store.state.currentNotebookID
        },
        notebooksList () {
            return this.$store.state.notebooksList
        },
        notesList () {
            return this.$store.state.notesList
        },
        tagsList () {
            return this.$store.state.tagsList
        },
        currentTagURL () {
            return this.$store.state.currentTagURL
        },
        articleCurrent () {
            return this.$store.state.currentArticle
        },
        showAdvancedInfo () {
            return this.$store.state.showAdvancedNoteInfo
        },
        readerMode () {
            return this.$store.state.readerMode
        },
        layoutBig () {
            return this.$store.state.layoutBig
        },
        notebookCount () {
            return this.$store.getters.getNotebooksCount
        },
        tagsCount () {
            return this.$store.getters.getTagsCount
        }
    }
}

export default mixin
