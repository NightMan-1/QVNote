/**
 * Shared Ace loader.
 *
 * Ace is shipped as static assets (see the `postinstall` script in
 * package.json) instead of being bundled into the editor chunk: it is ~470 KB
 * of prebuilt UMD code and is only needed on the editor page.
 *
 * Besides QVNote's own code editor (AceCodeEditor.vue), the HugeRTE
 * "supercode" plugin uses the global `window.ace`, so this module must be
 * loaded as soon as the editor page opens.
 */

export const ACE_THEME = 'ace/theme/ayu-light'

// Ayu Light, adapted from https://github.com/ayu-theme/ayu-ace
// MIT License, Copyright (c) 2016 Ike Ku.
// Changes: the upstream selectors are kept, plus .ace_meta.ace_tag /
// .ace_tag-name rules so HTML tags are highlighted by the current html mode
// tokenization (upstream only styles .ace_entity.ace_name.ace_tag).
export const ACE_THEME_CSS = `
.ace-ayu-light .ace_gutter { background: #FAFAFA; color: #B4B8BD; }
.ace-ayu-light .ace_print-margin { width: 1px; background: #E8E8E8; }
.ace-ayu-light { background-color: #FAFAFA; color: #6E7580; }
.ace-ayu-light .ace_cursor { color: #FF6A00; }
.ace-ayu-light .ace_marker-layer .ace_selection { background: #F0EEE4; }
.ace-ayu-light.ace_multiselect .ace_selection.ace_start { box-shadow: 0 0 3px 0 #FAFAFA; border-radius: 2px; }
.ace-ayu-light .ace_marker-layer .ace_step { background: #C6DBAE; }
.ace-ayu-light .ace_marker-layer .ace_bracket { margin: -1px 0 0 -1px; border: 1px solid #FF0000; }
.ace-ayu-light .ace_marker-layer .ace_active-line { background: #F2F2F2; }
.ace-ayu-light .ace_gutter-active-line { background-color: #F2F2F2; }
.ace-ayu-light .ace_marker-layer .ace_selected-word { border: 1px solid #F0EEE4; }
.ace-ayu-light .ace_fold { background-color: #F29718; border-color: #6E7580; }
.ace-ayu-light .ace_keyword.ace_operator { color: #ED9366; }
.ace-ayu-light .ace_constant.ace_language,
.ace-ayu-light .ace_constant.ace_numeric,
.ace-ayu-light .ace_constant.ace_character,
.ace-ayu-light .ace_constant.ace_other { color: #A37ACC; }
.ace-ayu-light .ace_constant.ace_character.ace_escape { color: #4CBF99; }
.ace-ayu-light .ace_support.ace_function { color: #F07171; }
.ace-ayu-light .ace_support.ace_constant { font-style: italic; color: #ED9366; }
.ace-ayu-light .ace_support.ace_class,
.ace-ayu-light .ace_support.ace_type { font-style: italic; color: #55B4D4; }
.ace-ayu-light .ace_storage,
.ace-ayu-light .ace_storage.ace_type { color: #FA6E32; }
.ace-ayu-light .ace_invalid { color: #F51818; }
.ace-ayu-light .ace_invalid.ace_deprecated { color: #FFFFFF; background-color: #FA6E32; }
.ace-ayu-light .ace_string { color: #86B300; }
.ace-ayu-light .ace_string.ace_regexp { color: #4CBF99; }
.ace-ayu-light .ace_comment { font-style: italic; color: #ABB0B6; }
.ace-ayu-light .ace_variable { color: #6E7580; }
.ace-ayu-light .ace_variable.ace_language { font-style: italic; color: #55B4D4; }
.ace-ayu-light .ace_variable.ace_parameter { color: #A37ACC; }
.ace-ayu-light .ace_entity.ace_other.ace_attribute-name { color: #F29718; }
.ace-ayu-light .ace_entity.ace_name.ace_function { color: #F29718; }
.ace-ayu-light .ace_entity.ace_name.ace_tag,
.ace-ayu-light .ace_meta.ace_tag,
.ace-ayu-light .ace_tag-name { color: #55B4D4; }
`

let acePromise = null
let themeDefined = false

function defineTheme (ace) {
    if (themeDefined) {
        return
    }
    themeDefined = true

    ace.define(ACE_THEME, ['require', 'exports', 'module', 'ace/lib/dom'], function (require, exports) {
        exports.isDark = false
        exports.cssClass = 'ace-ayu-light'
        exports.cssText = ACE_THEME_CSS

        const dom = require('../lib/dom')
        dom.importCssString(exports.cssText, exports.cssClass, false)
    })
}

export function loadAce () {
    if (window.ace && window.ace.edit) {
        defineTheme(window.ace)
        return Promise.resolve(window.ace)
    }

    if (!acePromise) {
        acePromise = new Promise((resolve, reject) => {
            const script = document.createElement('script')
            script.src = '/static/ace/ace.js'
            script.async = true
            script.onload = () => {
                if (!window.ace || !window.ace.edit) {
                    reject(new Error('Ace failed to initialise'))
                    return
                }
                window.ace.config.set('basePath', '/static/ace/')
                defineTheme(window.ace)
                resolve(window.ace)
            }
            script.onerror = () => reject(new Error('Failed to load /static/ace/ace.js'))
            document.head.appendChild(script)
        })
    }

    return acePromise
}
