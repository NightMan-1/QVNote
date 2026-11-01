<template>
    <article class="m-note">
        <header class="m-note__head">
            <h1 class="m-note__title">{{ article.title }}</h1>
            <div v-if="article.tags && article.tags.length" class="m-note__tags">
                <span v-for="tag in article.tags" :key="tag" class="m-note__tag">#{{ tag }}</span>
            </div>
        </header>

        <div
            ref="contentEl"
            class="m-note__content"
            :class="'cellType_' + article.type"
            v-html="html"
            @click="onContentClick"
        ></div>
    </article>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { handleCodePenRefresh, highlightCodeBlocks, renderNoteContent } from '../utils/noteContent'

// Mobile reader. Deliberately independent of the desktop NoteArticle: its own
// markup, its own styles (src/mobile/content.css) and its own reading rhythm.
const props = defineProps({
    article: { type: Object, required: true }
})

const contentEl = ref(null)
const html = computed(() => renderNoteContent(props.article.content))

function highlight () {
    nextTick(() => highlightCodeBlocks(contentEl.value))
}

onMounted(highlight)
watch(html, highlight)

function onContentClick (event) {
    handleCodePenRefresh(event)
}
</script>

<style scoped>
.m-note {
    padding: .25rem 1rem 3rem;
    background: var(--m-bg);
}

.m-note__head {
    margin: .25rem 0 .9rem;
}

.m-note__title {
    margin: 0 0 .55rem;
    color: var(--m-title);
    font-size: 1.6rem;
    font-weight: 500;
    line-height: 1.3;
    overflow-wrap: anywhere;
}

.m-note__tags {
    display: flex;
    flex-wrap: wrap;
    gap: .35rem;
}

.m-note__tag {
    padding: .22rem .45rem;
    border-radius: var(--m-radius-sm);
    background: var(--m-chip-bg);
    color: var(--m-chip-text);
    font-size: .75rem;
    line-height: 1.3;
}
</style>
