const mixin = {
    computed: {
        gridClass () {
            return this.$store.getters.getGridClass
        },
        pageType () {
            return this.$store.getters.getPageType
        },
        sidebarType () {
            return this.$store.getters.getSidebarType
        },
        notesCountInbox () {
            return this.$store.getters.getNotesCountInbox
        },
        notesCountTrash () {
            return this.$store.getters.getNotesCountTrash
        },
        notesCountTotal () {
            return this.$store.getters.getNotesCountTotal
        },
        notesCountFavorites () {
            return this.$store.getters.getNotesCountFavorites
        },
        currentNotebookID () {
            return this.$store.getters.getCurrentNotebookID
        },
        notebooksList () {
            return this.$store.getters.getNotebooksList
        },
        notesList () {
            return this.$store.getters.getNotesList
        },
        tagsList () {
            return this.$store.getters.getTagsList
        },
        currentTagURL () {
            return this.$store.getters.getCurrentTagURL
        },
        articleCurrent () {
            return this.$store.getters.getCurrentArticle
        },
        showAdvancedInfo () {
            return this.$store.getters.getShowAdvancedNoteInfo
        },
        readerMode () {
            return this.$store.getters.getReaderMode
        }
    }
}

export default mixin
