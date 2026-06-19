// Bootstrap Icons SVG pack for HugeRTE.
// Imported SVGs are resized to 24x24 and stripped of Bootstrap-specific
// attributes so TinyMCE can treat them as regular icon SVGs.

import arrowCounterclockwise from 'bootstrap-icons/icons/arrow-counterclockwise.svg?raw'
import arrowClockwise from 'bootstrap-icons/icons/arrow-clockwise.svg?raw'
import textParagraph from 'bootstrap-icons/icons/text-paragraph.svg?raw'
import typeBold from 'bootstrap-icons/icons/type-bold.svg?raw'
import typeItalic from 'bootstrap-icons/icons/type-italic.svg?raw'
import typeUnderline from 'bootstrap-icons/icons/type-underline.svg?raw'
import blockquoteLeft from 'bootstrap-icons/icons/blockquote-left.svg?raw'
import code from 'bootstrap-icons/icons/code.svg?raw'
import fileCode from 'bootstrap-icons/icons/file-code.svg?raw'
import eraser from 'bootstrap-icons/icons/eraser.svg?raw'
import eraserFill from 'bootstrap-icons/icons/eraser-fill.svg?raw'
import listUl from 'bootstrap-icons/icons/list-ul.svg?raw'
import listOl from 'bootstrap-icons/icons/list-ol.svg?raw'
import palette from 'bootstrap-icons/icons/palette.svg?raw'
import paintBucket from 'bootstrap-icons/icons/paint-bucket.svg?raw'
import textLeft from 'bootstrap-icons/icons/text-left.svg?raw'
import textCenter from 'bootstrap-icons/icons/text-center.svg?raw'
import textRight from 'bootstrap-icons/icons/text-right.svg?raw'
import link45Deg from 'bootstrap-icons/icons/link-45deg.svg?raw'
import imageIcon from 'bootstrap-icons/icons/image.svg?raw'
import arrowsFullscreen from 'bootstrap-icons/icons/arrows-fullscreen.svg?raw'
import codeSquare from 'bootstrap-icons/icons/code-square.svg?raw'
import braces from 'bootstrap-icons/icons/braces.svg?raw'

// Common UI icons used by HugeRTE menus/dialogs
import xLg from 'bootstrap-icons/icons/x-lg.svg?raw'
import checkLg from 'bootstrap-icons/icons/check-lg.svg?raw'
import xCircle from 'bootstrap-icons/icons/x-circle.svg?raw'
import arrowLeft from 'bootstrap-icons/icons/arrow-left.svg?raw'
import arrowRight from 'bootstrap-icons/icons/arrow-right.svg?raw'
import caretDownFill from 'bootstrap-icons/icons/caret-down-fill.svg?raw'
import caretUpFill from 'bootstrap-icons/icons/caret-up-fill.svg?raw'
import caretLeftFill from 'bootstrap-icons/icons/caret-left-fill.svg?raw'
import caretRightFill from 'bootstrap-icons/icons/caret-right-fill.svg?raw'
import questionCircle from 'bootstrap-icons/icons/question-circle.svg?raw'
import infoCircle from 'bootstrap-icons/icons/info-circle.svg?raw'
import exclamationTriangle from 'bootstrap-icons/icons/exclamation-triangle.svg?raw'
import floppy from 'bootstrap-icons/icons/floppy.svg?raw'
import fileEarmarkPlus from 'bootstrap-icons/icons/file-earmark-plus.svg?raw'
import folder2Open from 'bootstrap-icons/icons/folder2-open.svg?raw'
import upload from 'bootstrap-icons/icons/upload.svg?raw'
import copy from 'bootstrap-icons/icons/copy.svg?raw'
import scissors from 'bootstrap-icons/icons/scissors.svg?raw'
import clipboard from 'bootstrap-icons/icons/clipboard.svg?raw'
import printer from 'bootstrap-icons/icons/printer.svg?raw'
import eye from 'bootstrap-icons/icons/eye.svg?raw'
import search from 'bootstrap-icons/icons/search.svg?raw'
import gear from 'bootstrap-icons/icons/gear.svg?raw'
import table from 'bootstrap-icons/icons/table.svg?raw'
import plus from 'bootstrap-icons/icons/plus.svg?raw'
import dash from 'bootstrap-icons/icons/dash.svg?raw'
import columns from 'bootstrap-icons/icons/columns.svg?raw'
import columnsGap from 'bootstrap-icons/icons/columns-gap.svg?raw'
import dashLg from 'bootstrap-icons/icons/dash-lg.svg?raw'
import clock from 'bootstrap-icons/icons/clock.svg?raw'
import keyboard from 'bootstrap-icons/icons/keyboard.svg?raw'
import type from 'bootstrap-icons/icons/type.svg?raw'
import typeStrikethrough from 'bootstrap-icons/icons/type-strikethrough.svg?raw'
// Bootstrap has no dedicated 'justify' text icon; reuse text-left for align-justify.
import textLeftForJustify from 'bootstrap-icons/icons/text-left.svg?raw'
import textIndentLeft from 'bootstrap-icons/icons/text-indent-left.svg?raw'
import textIndentRight from 'bootstrap-icons/icons/text-indent-right.svg?raw'
import threeDots from 'bootstrap-icons/icons/three-dots.svg?raw'
import film from 'bootstrap-icons/icons/film.svg?raw'
import emojiSmile from 'bootstrap-icons/icons/emoji-smile.svg?raw'
import boxArrowUpRight from 'bootstrap-icons/icons/box-arrow-up-right.svg?raw'

function normalizeSvg (svg) {
    return svg
        .replace(/\s*xmlns="http:\/\/www\.w3\.org\/2000\/svg"\s*/g, ' ')
        .replace(/fill="currentColor"/g, '')
        .replace(/class="bi [^"]*"/g, '')
        .replace(/\s{2,}/g, ' ')
        .trim()
}

const icons = {
    // Toolbar
    undo: normalizeSvg(arrowCounterclockwise),
    redo: normalizeSvg(arrowClockwise),
    blocks: normalizeSvg(textParagraph),
    bold: normalizeSvg(typeBold),
    italic: normalizeSvg(typeItalic),
    underline: normalizeSvg(typeUnderline),
    blockquote: normalizeSvg(blockquoteLeft),
    inlinecode: normalizeSvg(code),
    codeblock: normalizeSvg(fileCode),
    removeformat: normalizeSvg(eraser),
    bullist: normalizeSvg(listUl),
    numlist: normalizeSvg(listOl),
    forecolor: normalizeSvg(palette),
    backcolor: normalizeSvg(paintBucket),
    alignleft: normalizeSvg(textLeft),
    aligncenter: normalizeSvg(textCenter),
    alignright: normalizeSvg(textRight),
    link: normalizeSvg(link45Deg),
    image: normalizeSvg(imageIcon),
    fullscreen: normalizeSvg(arrowsFullscreen),
    eraser: normalizeSvg(eraserFill),
    sourcecode: normalizeSvg(codeSquare),
    codeblock: normalizeSvg(braces),

    // Common UI
    close: normalizeSvg(xLg),
    checkmark: normalizeSvg(checkLg),
    cancel: normalizeSvg(xCircle),
    'arrow-left': normalizeSvg(arrowLeft),
    'arrow-right': normalizeSvg(arrowRight),
    'chevron-down': normalizeSvg(caretDownFill),
    'chevron-up': normalizeSvg(caretUpFill),
    'chevron-left': normalizeSvg(caretLeftFill),
    'chevron-right': normalizeSvg(caretRightFill),
    help: normalizeSvg(questionCircle),
    info: normalizeSvg(infoCircle),
    warning: normalizeSvg(exclamationTriangle),
    save: normalizeSvg(floppy),
    'new-document': normalizeSvg(fileEarmarkPlus),
    browse: normalizeSvg(folder2Open),
    upload: normalizeSvg(upload),
    copy: normalizeSvg(copy),
    cut: normalizeSvg(scissors),
    paste: normalizeSvg(clipboard),
    print: normalizeSvg(printer),
    preview: normalizeSvg(eye),
    search: normalizeSvg(search),
    settings: normalizeSvg(gear),
    table: normalizeSvg(table),
    'table-insert-row-after': normalizeSvg(plus),
    'table-insert-row-before': normalizeSvg(plus),
    'table-delete-row': normalizeSvg(dash),
    'table-insert-col-after': normalizeSvg(plus),
    'table-delete-col': normalizeSvg(dash),
    'table-merge-cells': normalizeSvg(columns),
    'table-split-cells': normalizeSvg(columnsGap),
    line: normalizeSvg(dashLg),
    'horizontal-rule': normalizeSvg(dashLg),
    'insert-time': normalizeSvg(clock),
    'insert-character': normalizeSvg(keyboard),
    'code-sample': normalizeSvg(fileCode),
    format: normalizeSvg(type),
    strikethrough: normalizeSvg(typeStrikethrough),
    'align-justify': normalizeSvg(textLeftForJustify),
    indent: normalizeSvg(textIndentLeft),
    outdent: normalizeSvg(textIndentRight),
    'image-options': normalizeSvg(threeDots),
    media: normalizeSvg(film),
    emoji: normalizeSvg(emojiSmile),
    'open-link': normalizeSvg(boxArrowUpRight),
    unlink: normalizeSvg(link45Deg)
}

export function registerBootstrapIcons (editor) {
    for (const name in icons) {
        if (Object.prototype.hasOwnProperty.call(icons, name)) {
            editor.ui.registry.addIcon(name, icons[name])
        }
    }
}

export default icons
