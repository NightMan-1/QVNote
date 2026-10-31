<template>
    <article class="m-article">
        <h1 class="text-success">{{ article.title }}</h1>
        <div v-if="article.tags && article.tags.length" class="m-article__tags">
            <span v-for="tag in article.tags" :key="tag" class="badge text-bg-primary fw-normal">#{{ tag }}</span>
        </div>
        <div class="clearfix"></div>
        <div
            ref="contentEl"
            class="articleCell"
            :class="'cellType_' + article.type"
            v-html="html"
            @click="onContentClick"
        ></div>
    </article>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { highlightCodeBlocks, renderNoteContent } from '../utils/noteContent'

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

// CodePen embeds can be refreshed in place (the iframe is proxied by the server)
function onContentClick (event) {
    const btn = event.target.closest('.codepen-refresh')
    if (!btn) {
        return
    }
    event.preventDefault()
    const src = btn.getAttribute('data-codepen-refresh')
    if (!src) {
        return
    }
    const iframe = btn.closest('.codepen-figure') ? btn.closest('.codepen-figure').querySelector('iframe') : null
    if (!iframe) {
        return
    }
    btn.classList.add('codepen-refresh--loading')
    const url = src + (src.indexOf('?') >= 0 ? '&' : '?') + 'refresh=1'
    fetch(url, { method: 'POST' })
        .catch(() => {})
        .then(() => {
            btn.classList.remove('codepen-refresh--loading')
            iframe.src = src + (src.indexOf('?') >= 0 ? '&' : '?') + 't=' + Date.now()
        })
}
</script>

<style scoped>
.m-article {
    padding: .25rem 1rem 3rem;
    overflow-wrap: anywhere;
}

.m-article h1 {
    font-size: 1.6rem;
    line-height: 1.3;
    margin: .25rem 0 .75rem;
}

.m-article__tags {
    display: flex;
    flex-wrap: wrap;
    gap: .35rem;
    margin-bottom: .75rem;
}

.m-article :deep(img),
.m-article :deep(video),
.m-article :deep(iframe) {
    max-width: 100%;
    height: auto;
}

.m-article :deep(pre) {
    overflow-x: auto;
}
</style>
