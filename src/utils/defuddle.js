import MarkdownIt from 'markdown-it'
import Defuddle from 'defuddle'

const mdRenderer = new MarkdownIt({ html: true, linkify: true, breaks: false })

// Fetches a page through the server-side proxy (/api/fetch.json, direct
// browser fetch is blocked by CORS on most sites), extracts the article
// with defuddle and returns { title, description, html }.
// Throws Error with a human-readable message on fetch/parse errors.
export async function fetchDefuddleArticle (url) {
    const target = String(url || '').trim()
    if (!target) {
        throw new Error('empty url')
    }
    const resp = await fetch('/api/fetch.json', {
        method: 'POST',
        body: JSON.stringify({ url: target }),
        signal: AbortSignal.timeout(90000)
    })
    const data = await resp.json()
    if (data.error) {
        throw new Error(data.error)
    }
    const doc = new DOMParser().parseFromString(data.html, 'text/html')
    const article = new Defuddle(doc, { url: data.url || target, markdown: true }).parse()
    if (!article.content || !article.content.trim()) {
        throw new Error('no content extracted')
    }
    return {
        title: normalizeTitle(article.title || ''),
        description: article.description || '',
        html: mdRenderer.render(article.content)
    }
}

// Trims and collapses whitespace runs into single spaces.
export function normalizeTitle (t) {
    return String(t || '').replace(/\s+/g, ' ').trim()
}
