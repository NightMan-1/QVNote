# QVNote qvEditor.vue — Detailed Reference

## 1. Component purpose

`src/components/qvEditor.vue` is the note editor for QVNote. It supports two content types:

- `text` — rich-text editing via HugeRTE (a community fork of TinyMCE 7).
- `code` — plain HTML editing via `vue-prism-editor`.

The component is written with Vue 3 Options API and uses the official HugeRTE Vue wrapper `@hugerte/hugerte-vue`.

## 2. HugeRTE integration

### Self-hosted setup

Core HugeRTE code is bundled by Vite through ES imports. Static assets (skins, themes, icons, plugins) are copied from `node_modules/hugerte/` to `public/static/hugerte/` by the `postinstall` script in `package.json`.

```js
import Editor from '@hugerte/hugerte-vue'
import 'hugerte'
import 'hugerte/themes/silver'
import 'hugerte/icons/default'
import 'hugerte/models/dom'
import 'hugerte/plugins/lists'
import 'hugerte/plugins/link'
import 'hugerte/plugins/image'
import 'hugerte/plugins/fullscreen'
import 'hugerte/plugins/codesample'
```

Config points to static assets:

```js
skin_url: '/static/hugerte/skins/ui/oxide',
content_css: ['/static/hugerte/skins/content/default/content.css', '/static/prism/prism-atom-one-light.css'],
theme_url: '/static/hugerte/themes/silver/theme.min.js',
```

### `window.tinymce` vs `window.hugerte`

HugeRTE registers itself as `window.hugerte`. Many TinyMCE plugins and libraries (including `tinymce-paste-from-word-lib` and `supercode`) reference `window.tinymce` in their UMD wrappers. If the alias is missing, these plugins load but receive `undefined` as their `tinymce` argument and fail silently — no console errors, no functionality.

The alias must be set **before** HugeRTE `init()` runs. In this component it is set in `mounted()`:

```js
mounted () {
    if (typeof window.tinymce === 'undefined' && typeof window.hugerte !== 'undefined') {
        window.tinymce = window.hugerte
    }
    // ... load external plugins, then set editorReady = true
}
```

## 3. paste_from_word integration

### What the npm package actually is

`@pangaeatech/tinymce-paste-from-word-lib` is a **library**, not a complete plugin. It exports a single cleanup function with the signature `function(tinymce, event)`. It does **not** call `PluginManager.add()` on its own.

### Why a custom wrapper is needed

1. The package expects `tinymce` as a global/module name.
2. It does not register event handlers.
3. It must be loaded before the editor starts processing paste.

The project therefore ships a custom wrapper at `public/static/hugerte/plugins/paste_from_word/plugin.min.js`.

### Wrapper responsibilities

1. Alias `window.tinymce = window.hugerte` (belt-and-suspenders; also done in component).
2. Wait for the library script (`plugin.min.js` from the same package) to load.
3. Register a `PastePreProcess` handler that invokes the library.

### Loading order

The component loads the wrapper in `mounted()` via a dynamic `<script>` tag. Only after `script.onload` fires does it set `editorReady = true`, guaranteeing the wrapper is ready before HugeRTE initializes.

```js
var script = document.createElement('script')
script.src = '/static/hugerte/plugins/paste_from_word/plugin.min.js'
script.onload = function () { self.editorReady = true }
script.onerror = function () { self.editorReady = true } // graceful degradation
document.head.appendChild(script)
```

### Gotchas

- Do **not** list `paste_from_word` in the `plugins:` config if it is not a built-in HugeRTE plugin. Use only `external_plugins`.
- The original attempt to rely on `external_plugins` alone failed because the library never self-registered.
- Loading the wrapper through `external_plugins` also failed because HugeRTE loaded it after `init()`, too late to attach event handlers.

## 4. Custom toolbar buttons

### Registration pattern

All custom buttons are registered inside the `setup(editor)` callback:

```js
editor.ui.registry.addButton('mybutton', {
    icon: 'icon-name',
    tooltip: 'Tooltip',
    onAction: () => { /* ... */ }
})
```

For custom SVG icons, register them first:

```js
editor.ui.registry.addIcon('eraser', '<svg ...>...</svg>')
editor.ui.registry.addButton('eraser', {
    icon: 'eraser',
    tooltip: 'Clean HTML',
    onAction: () => { /* ... */ }
})
```

### Fullscreen

Uses the browser Fullscreen API on the `.editor` DOM element, not the HugeRTE built-in fullscreen. This gives better control over the surrounding Vue layout.

### Clean HTML (`eraser`)

POSTs the current editor content to `/api/cleanup_html.json` and replaces the editor content with the server-cleaned HTML. Icon is a Bootstrap `eraser` SVG.

### Inline code (`inlinecode`)

Text-only button labeled `<>`. Behavior:

- If selection spans multiple blocks → insert `<pre class="language-plaintext"><code>...</code></pre>` with escaped HTML.
- Otherwise → `editor.execCommand('mceToggleFormat', false, 'code')`.

### Code block (`codeblock`)

Text-only button labeled `{ }`. Behavior:

- If cursor is inside an existing `<pre class="language-*">` → open HugeRTE `codesample` dialog for editing.
- If selection is non-empty → wrap it in `<pre class="language-plaintext"><code>...</code></pre>`.
- Otherwise → open `codesample` dialog to insert a new block.

### Why separate `inlinecode`/`codeblock` buttons exist

The HugeRTE `codesample` plugin provides a dialog for code blocks but no convenient inline-code button. The custom buttons unify the experience:

- Inline code for short spans.
- Code block for multi-line/preformatted snippets.

## 5. Syntax highlighting with PrismJS

### Inside the editor

The `codesample` plugin loads PrismJS, but the custom code blocks (`inlinecode`/`codeblock`) need explicit highlighting. The component listens to `SetContent` and `NodeChange` inside the editor iframe and calls `Prism.highlightElement()` on `<pre class="language-*">` elements that do not already contain `.token` spans.

```js
editor.on('init', function () {
    var doc = editor.getDoc()
    var highlightCodeBlocks = function () {
        if (window.Prism && window.Prism.highlightElement) {
            var pres = doc.querySelectorAll('pre[class*="language-"]')
            for (var i = 0; i < pres.length; i++) {
                if (!pres[i].querySelector('.token')) {
                    try { window.Prism.highlightElement(pres[i]) } catch(e) {}
                }
            }
        }
    }
    editor.on('SetContent', highlightCodeBlocks)
    editor.on('NodeChange', highlightCodeBlocks)
    setTimeout(highlightCodeBlocks, 500)
})
```

### In note reading view (`qvApp.vue`)

PrismJS theme CSS is loaded once when the editor is used. After the article renders via `v-html`, `Prism.highlightAllUnder('.articleCell')` is called.

### CSS gotchas

- Do **not** use `!important` on global `code`/`pre` background colors. It makes the `vue-prism-editor` overlay textarea invisible (cursor/selection disappear).
- Use specific selectors and avoid `!important`.
- `<pre>` blocks for plain code use `#f0f3f5` background and `#363636` text; inline `<code>` uses `#f0f3f5` background and `#e83e8c` text.

## 6. Supercode plugin (Ace source editor)

### Plugin source

The project uses `supercode-tinymce-plugin` (not available on npm). Files are vendored in `public/static/supercode/`:

- `plugin.js` — readable source.
- `plugin.min.js` — identical copy, loaded by HugeRTE.

### Patches applied to upstream

1. Removed the `tinymce.majorVersion <= 5` auto-fallback so Custom View mode works with HugeRTE (which reports majorVersion 1).
2. Removed CDN dependency loading; dependencies are provided by the host app.

### Host app responsibilities

Import Ace modules and expose globals:

```js
import 'ace-builds/src-min-noconflict/ace'
import 'ace-builds/src-min-noconflict/theme-crimson_editor'
import 'ace-builds/src-min-noconflict/mode-html'
import 'ace-builds/src-min-noconflict/ext-language_tools'
import { html as htmlBeautify } from 'js-beautify'

if (window.ace && window.ace.config) {
    window.ace.config.set('basePath', '/static/ace/')
}
window.html_beautify = htmlBeautify
```

`public/static/ace/` is populated from `ace-builds/src-min-noconflict/` by `postinstall`.

### Configuration

```js
supercode: {
    theme: 'crimson_editor',
    language: 'html',
    fontSize: 14,
    wrap: true,
    fallbackModal: false,   // use Custom View inline
    shortcut: true
}
```

With `fallbackModal: false`, Supercode replaces the editor area with an Ace instance using HugeRTE's Custom View API. The toolbar is reused; only the Supercode button remains enabled.

### Modal fallback

If `fallbackModal: true` or inline mode is active, Supercode opens a full-page modal (`#supercode-modal-container`). The z-index was raised to `1500` in `src/style/main.css` so it appears above other UI layers:

```css
.tox .tox-dialog-wrap,
#supercode-modal-container {
    z-index: 1500;
}
```

## 7. Paste cleanup for plain `<pre><code>`

When pasting content from other sources, unclassified `<pre><code>` blocks are tagged as `language-plaintext` so PrismJS can highlight them consistently:

```js
editor.on('PastePreProcess', function (e) {
    // parse e.content, find <pre> with direct <code> child and no class,
    // set class="language-plaintext"
})
```

## 8. Content style

`content_style` is used instead of a separate stylesheet for editor-specific rules. It defines typography, heading sizes, margins, and `pre`/`code` colors. This keeps editor styling co-located with the component config.

## 9. Editor height

CSS in the `<style>` block sets HugeRTE height dynamically:

```css
.tox.tox-hugerte {
    height: calc(100vh - 270px) !important;
    min-height: 40rem;
}
```

## 10. Testing checklist

When changing this component:

1. `rtk lint src/components/qvEditor.vue` — must pass.
2. `rtk npm run build` — must produce `templates/`.
3. `rtk go build` — must produce binary with embedded templates.
4. Manual / Playwright verification:
   - Open an existing note in the editor.
   - Check toolbar buttons: fullscreen, eraser, inline code, code block, supercode.
   - Paste Word content (if `paste_from_word` changed) — verify cleanup.
   - Insert inline code and code block — verify Prism highlighting.
   - Toggle Supercode source view — verify Ace loads with `crimson_editor` theme and content round-trips.

## 11. Files related to the editor

- `src/components/qvEditor.vue` — main component.
- `src/components/qvApp.vue` — note reading view, Prism theme loading.
- `src/style/main.css` — global `pre`/`code` styles, z-index overrides.
- `public/static/hugerte/plugins/paste_from_word/plugin.min.js` — paste wrapper.
- `public/static/supercode/plugin.js` / `plugin.min.js` — Supercode plugin.
- `package.json` — `postinstall` copies HugeRTE and Ace assets.
- `qvnote-server.go` — `/api/cleanup_html.json` endpoint used by Clean HTML.
