---
name: qvnote-qveditor
description: Working knowledge for QVNote's qvEditor.vue component. Use when modifying src/components/qvEditor.vue, HugeRTE/TinyMCE editor, paste_from_word, supercode-tinymce-plugin, PrismJS syntax highlighting, or custom toolbar buttons.
---

# QVNote qvEditor.vue Skill

Component path: `src/components/qvEditor.vue`

## Architecture at a glance

- Vue 3 Options API component using `@hugerte/hugerte-vue` (TinyMCE fork).
- Self-hosted HugeRTE: core bundled by Vite, static assets served from `public/static/hugerte/`.
- Two editor modes: `text` (HugeRTE) and `code` (`vue-prism-editor`).
- `editorReady` gate: the `<Editor>` only renders after `paste_from_word` library is loaded in `mounted()`.

## Critical global aliases

HugeRTE exposes `window.hugerte`, but TinyMCE plugins expect `window.tinymce`. Set before editor init:

```js
if (typeof window.tinymce === 'undefined' && typeof window.hugerte !== 'undefined') {
    window.tinymce = window.hugerte
}
```

Also expose Ace + beautify globals for Supercode:

```js
window.html_beautify = htmlBeautify
window.ace.config.set('basePath', '/static/ace/')
```

## External plugin loading

- Built-in HugeRTE plugins are imported at module top (`import 'hugerte/plugins/...'`).
- `paste_from_word` is NOT a built-in plugin. It is a UMD library (`@pangaeatech/tinymce-paste-from-word-lib`) loaded via a custom wrapper at `public/static/hugerte/plugins/paste_from_word/plugin.min.js`.
- The wrapper aliases `window.tinymce = window.hugerte`, then registers `PastePreProcess` to call the library.

## Custom toolbar buttons

Registered in the `setup(editor)` callback:

| Button | ID | Behavior |
|--------|----|----------|
| Fullscreen | `fullscreen` | Toggles browser fullscreen on `.editor` |
| Clean HTML | `eraser` | POSTs content to `/api/cleanup_html.json`, replaces editor content |
| Inline code | `inlinecode` | Wraps selection in `<code>`; multi-block selection becomes `<pre class="language-plaintext"><code>...` |
| Code block | `codeblock` | Wraps selection or opens HugeRTE `codesample` dialog |
| Source Code Editor | `supercode` | Toggles Ace-based source-code view |

Custom icons are registered via `editor.ui.registry.addIcon('name', svgString)` before `addButton`.

## Syntax highlighting

- Editor live preview: PrismJS highlights `<pre class="language-*">` blocks inside the iframe on `SetContent`/`NodeChange`.
- Note reading view: `qvApp.vue` loads `prism-atom-one-light.css` and calls `Prism.highlightAllUnder('.articleCell')`.
- Code mode: `vue-prism-editor` + `Prism.languages.markup`.

## Supercode (Ace source editor)

- Plugin files: `public/static/supercode/plugin.js` and `plugin.min.js` (copies of each other).
- Plugin was patched to remove the TinyMCE majorVersion ≤ 5 auto-fallback so Custom View works with HugeRTE.
- Config uses `fallbackModal: false` (Custom View inline) and theme `crimson_editor`.
- Ace dependencies: `ace-builds/src-min-noconflict/{ace,theme-crimson_editor,mode-html,ext-language_tools}` imported in the component.

## Common pitfalls

1. **HugeRTE plugin listed in both `plugins` and `external_plugins`** — causes silent failure. Use only `external_plugins` for non-built-in plugins.
2. **UMD plugin sees wrong global** — always set `window.tinymce = window.hugerte` before init.
3. **Ace worker/theme 404** — set `window.ace.config.set('basePath', '/static/ace/')` and ensure `public/static/ace/` is populated by postinstall.
4. **Prism CSS `!important` on `code`/`pre` backgrounds** — breaks `vue-prism-editor` cursor/selection; avoid `!important`.
5. **`codesample` plugin conflicts** — it also registers a `code` toolbar button; use distinct button names (`inlinecode`, `codeblock`).

## Build / verify

```bash
rtk npm run build          # Vite → templates/
rtk go build               # embeds templates/
rtk lint src/components/qvEditor.vue
```

For visual verification use the Playwright skill (`/playwright`) against `http://localhost:8080/editor`.

## See also

- Detailed integration history and debugging notes: [REFERENCE.md](REFERENCE.md)
