<template>
    <div class="w-100">
        <vue-headful
            title="QVNote"
            description=""
            keywords=""
        />
        <router-view/>
    </div>
</template>

<script>
export default {
  name: 'App',
  mounted: function () {
    // global warning
    this.$store.watch(this.$store.getters.getStatus, errorType => {
      if (errorType === 1) {
        this.$toast.error({
          title: 'Error!',
          message: this.$store.getters.getStatusText,
          closeButton: true,
          progressBar: true,
          timeOut: 15000
        })
        this.$router.push({ name: 'qvErrorFatal' })
      } else if (errorType === 2) {
        this.$toast.error({
          title: 'Warning!',
          message: this.$store.getters.getStatusText,
          closeButton: true,
          progressBar: true,
          timeOut: 7000
        })
      } else if (errorType === 5) {
        this.$toast.success({
          title: '',
          message: this.$store.getters.getStatusText,
          closeButton: true,
          progressBar: true,
          timeOut: 5000
        })
      }
    })
  }
}
</script>

<style>
    @import url('https://fonts.googleapis.com/css?family=Roboto:400,400i,500,500i,700,700i&subset=cyrillic,cyrillic-ext,latin-ext');
    /* purgecss start ignore */
    html, body {
        font-family: 'Roboto', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
        font-kerning: normal;
        font-variant-ligatures: common-ligatures contextual;
        font-feature-settings: "kern", "liga", "clig", "calt";
        font-size: 14px;
        font-weight: 400;
        line-height: 1.42857143;
        font-display: swap; /* or block */
        background-color: white;
        height: 100vh;
        /*
        font-smooth:always;
        -webkit-font-smoothing:subpixel-antialiased;
        -moz-osx-font-smoothing:inherit;
        */
    }

    b, strong {
        font-weight: 500;
    }

    p {
        font-kerning: normal;
        font-variant-ligatures: common-ligatures contextual;
        font-feature-settings: "kern", "liga", "clig", "calt";
    }

    a, button {
        outline: 0 !important;
    }
    a {
        display: inline-block;
        position: relative;
    }
    a::after {
        content: '';
        position: absolute;
        width: 100%;
        transform: scaleX(0);
        height: 1px;
        bottom: 3px;
        left: 0;
        background-color: #167495;
        transform-origin: bottom right;
        transition: transform 0.15s ease-out;
    }
    a:hover {
        text-decoration: none;
    }
    a:hover::after {
        transform: scaleX(1);
        transform-origin: bottom left;
    }
    img {
        image-rendering: optimizeQuality;
        max-width: 100%;
        height: auto !important;
        margin: 0.5rem 0;
    }

    code {
        background-color: #f0f3f5 !important;
        padding: 0.1rem 0.5rem;
    }

    pre {
        background-color: #f0f3f5 !important;
        padding: 0.5rem;
        display: block;
        white-space: pre;
        color: #363636; /* e83e8c */
    }
    .h1, .h2, .h3, .h4, .h5, .h6, h1, h2, h3, h4, h5, h6 {
        margin-bottom: .5em;
    }
    p + .h1, .h2, .h3, .h4, .h5, .h6, h1, h2, h3, h4, h5, h6 {
        margin-top: 1.1em;
    }

    .article-main {
        font-size: 18px;
        line-height: 24px;
        max-width: 600px;
        margin: 0 auto;
    }

    .ml--1 {
        margin-left: -1rem;
    }
    .b-0 {border: none;}

    .app {
        background-color: rgba(228, 229, 230, 0.11);
        height: 100vh;
    }
    .app-body{
        margin-top: 0;
    }

    .header-fixed .app-header {
        z-index: 1;
    }

    .dropdown-toggle::after {
        margin-left: .4em;
        vertical-align: .2em;
    }

    .dropleft .dropdown-toggle::before {
        margin-right: .4em;
        vertical-align: 0.1rem;
        display: none;
    }

    .h1, h1 {
        font-weight: 500;
    }

    .progress {
        font-size: .9rem;
        height: 1.3rem;
    }

    .dashboard-button {
        width: 100%;
        text-align: left;
        border: 0;
        border-right: 1px solid #c8ced3;
        background: white;
        color: #20a8d8 !important;
        font-size: 1.1em;
        font-weight: 700;
        padding: .75em 1em .75em 0.5em;
        cursor: pointer;
    }

    .settings-button {
        position: absolute;
        right: .75em;
        top: .75rem;
    }

    .dropdown-divider {
        border-top: 1px solid #c8ced3;
        margin: 0;
        background: #c8ced3;
        height: 1px;
    }

    .header-fixed .app-header {
        top: 0;
    }

    .sidebar-nav .nav-item {
        width: 100%;
        /*padding-right: 0.5em;*/
    }
    .sidebar, .main {
        height: 100% !important;
        /* overflow: hidden; */
    }
    .sidebar .nav-link {
        padding: 0.5rem 1rem;
        text-overflow: ellipsis;
        overflow: hidden;
        white-space: nowrap;
    }
    .sidebar .nav-link.nav-link-notebook {
        font-size: 0.9em;
        padding: 0.3rem 1rem;
    }
    .sidebar-nav .nav-link i {
        padding-right: 0.2rem
    }

    .sidebar .nav-link i {
        color: #73818f;
    }

    .sidebar .nav-link:hover i {
        color: #ffffff;
    }

    .sidebar-minimized .sidebar .nav-link i {
        display: block;
        float: left;
        width: 50px;
        font-size: 18px;
    }

    .cxlt-toastr-container > div {
        width: 400px;
    }
    .cxlt-toastr-container > div, .cxlt-toastr-container > :hover {
        box-shadow: none;
        border-radius: 0.25rem;
    }
    .cxlt-toastr-container .toast {border: none; box-shadow:0 .25rem .75rem rgba(0,0,0,.1) !important;}
    .cxlt-toastr-container > :hover{
        box-shadow: 0 0 12px #999 !important;
    }
    .cxlt-toastr-container .toast-icon {
        top: 0.6rem !important;
    }

    .scrolly-col {
        height: calc(100vh); /* TODO here problem with url control on iOS*/
        overflow: hidden;
    }

    .scrolly-col .scrolly-col-wrap {
        height: 100%;
        padding-bottom: 4.3em;
    }
    .simplebarHeight {
      height: 100%;
      padding-bottom: 4.3em;
    }
    .sidebar .simplebarHeight {
      height: 100%;
    }
    .sidebar .simplebar-scrollbar::before {
        background: rgba(255, 255, 255, 0.65);
    }
    .btn-to-top {
        width: 60px;
        height: 60px;
        padding: 10px 16px;
        border-radius: 50%;
        font-size: 22px;
        line-height: 22px;
    }
/*
    .articleCell ol {
        counter-reset: counter;
    }
    .articleCell ol li::before {
        counter-increment: counter;
        content: counters(counter, '.') ' ';
    }
*/
    /* https://filamentgroup.github.io/select-css/demo/ */
    /* class applies to select element itself, not a wrapper element */
    .select-css {
        color: #444;
        box-sizing: border-box;
        margin: 0;
        border: 1px solid #aaa;
        -moz-appearance: none;
        -webkit-appearance: none;
        appearance: none;
        background-color: #fff;
        /* note: bg image below uses 2 urls. The first is an svg data uri for the arrow icon, and the second is the gradient.
            for the icon, if you want to change the color, be sure to use `%23` instead of `#`, since it's a url. You can also swap in a different svg icon or an external image reference

        */
        background-image: url('data:image/svg+xml;charset=US-ASCII,%3Csvg%20xmlns%3D%22http%3A%2F%2Fwww.w3.org%2F2000%2Fsvg%22%20width%3D%22292.4%22%20height%3D%22292.4%22%3E%3Cpath%20fill%3D%22%23007CB2%22%20d%3D%22M287%2069.4a17.6%2017.6%200%200%200-13-5.4H18.4c-5%200-9.3%201.8-12.9%205.4A17.6%2017.6%200%200%200%200%2082.2c0%205%201.8%209.3%205.4%2012.9l128%20127.9c3.6%203.6%207.8%205.4%2012.8%205.4s9.2-1.8%2012.8-5.4L287%2095c3.5-3.5%205.4-7.8%205.4-12.8%200-5-1.9-9.2-5.5-12.8z%22%2F%3E%3C%2Fsvg%3E'),
        linear-gradient(to bottom, #ffffff 0%,#e5e5e5 100%);
        background-repeat: no-repeat, repeat;
        /* arrow icon position (1em from the right, 50% vertical) , then gradient position*/
        background-position: right .7em top 50%, 0 0;
        /* icon size, then gradient */
        background-size: .65em auto, 100%;
    }
    /* Hide arrow icon in IE browsers */
    .select-css::-ms-expand {
        display: none;
    }
    /* Hover style */
    .select-css:hover {
        border-color: #888;
    }
    /* Focus style */
    .select-css:focus {
        border-color: #aaa;
        /* It'd be nice to use -webkit-focus-ring-color here but it doesn't work on box-shadow */
        box-shadow: 0 0 1px 3px rgba(59, 153, 252, .7);
        box-shadow: 0 0 0 3px -moz-mac-focusring;
        color: #222;
        outline: none;
    }

    /* Set options to normal weight */
    .select-css option {
        font-weight:normal;
    }

    /* Support for rtl text, explicit support for Arabic and Hebrew */
    *[dir="rtl"] .select-css, :root:lang(ar) .select-css, :root:lang(iw) .select-css {
        background-position: left .7em top 50%, 0 0;
        padding: .6em .8em .5em 1.4em;
    }

    /* ----------------------------------------------------------- */
    /* == tingle v0.13.2 */
    /* ----------------------------------------------------------- */

    .tingle-modal * {
        box-sizing: border-box;
    }

    .tingle-modal {
        position: fixed;
        top: 0;
        right: 0;
        bottom: 0;
        left: 0;
        z-index: 1000;
        display: -webkit-box;
        display: -ms-flexbox;
        display: flex;
        visibility: hidden;
        -webkit-box-orient: vertical;
        -webkit-box-direction: normal;
        -ms-flex-direction: column;
        flex-direction: column;
        -webkit-box-align: center;
        -ms-flex-align: center;
        align-items: center;
        overflow: hidden;
        -webkit-overflow-scrolling: touch;
        background: rgba(0, 0, 0, .8);
        opacity: 0;
        cursor: pointer;
        -webkit-transition: -webkit-transform .2s ease;
        transition: -webkit-transform .2s ease;
        transition: transform .2s ease;
        transition: transform .2s ease, -webkit-transform .2s ease;
    }

    /* confirm and alerts
    -------------------------------------------------------------- */

    .tingle-modal--confirm .tingle-modal-box {
        text-align: center;
    }

    /* modal
    -------------------------------------------------------------- */

    .tingle-modal--noOverlayClose {
        cursor: default;
    }

    .tingle-modal--noClose .tingle-modal__close {
        display: none;
    }

    .tingle-modal__close {
        position: fixed;
        top: 10px;
        right: 28px;
        z-index: 1000;
        padding: 0;
        width: 5rem;
        height: 5rem;
        border: none;
        background-color: transparent;
        color: #f0f0f0;
        font-size: 6rem;
        font-family: monospace;
        line-height: 1;
        cursor: pointer;
        -webkit-transition: color .3s ease;
        transition: color .3s ease;
    }

    .tingle-modal__closeLabel {
        display: none;
    }

    .tingle-modal__close:hover {
        color: #fff;
    }

    .tingle-modal-box {
        position: relative;
        -ms-flex-negative: 0;
        flex-shrink: 0;
        margin-top: auto;
        margin-bottom: auto;
        width: 60%;
        border-radius: 4px;
        background: #fff;
        opacity: 1;
        cursor: auto;
        -webkit-transition: -webkit-transform .3s cubic-bezier(.175, .885, .32, 1.275);
        transition: -webkit-transform .3s cubic-bezier(.175, .885, .32, 1.275);
        transition: transform .3s cubic-bezier(.175, .885, .32, 1.275);
        transition: transform .3s cubic-bezier(.175, .885, .32, 1.275), -webkit-transform .3s cubic-bezier(.175, .885, .32, 1.275);
        -webkit-transform: scale(.8);
        -ms-transform: scale(.8);
        transform: scale(.8);
    }

    .tingle-modal-box__content {
        padding: 2rem 3rem;
    }

    .tingle-modal-box__footer {
        padding: 1.5rem 2rem;
        width: auto;
        border-bottom-right-radius: 4px;
        border-bottom-left-radius: 4px;
        background-color: #f5f5f5;
        cursor: auto;
    }

    .tingle-modal-box__footer::after {
        display: table;
        clear: both;
        content: "";
    }

    .tingle-modal-box__footer--sticky {
        position: fixed;
        bottom: -200px;
        z-index: 10001;
        opacity: 1;
        -webkit-transition: bottom .3s ease-in-out .3s;
        transition: bottom .3s ease-in-out .3s;
    }

    /* state
    -------------------------------------------------------------- */

    .tingle-enabled {
        position: fixed;
        overflow: hidden;
        left: 0;
        right: 0;
    }

    .tingle-modal--visible .tingle-modal-box__footer {
        bottom: 0;
    }

    .tingle-enabled .tingle-content-wrapper {
        -webkit-filter: blur(8px);
        filter: blur(8px);
    }

    .tingle-modal--visible {
        visibility: visible;
        opacity: 1;
    }

    .tingle-modal--visible .tingle-modal-box {
        -webkit-transform: scale(1);
        -ms-transform: scale(1);
        transform: scale(1);
    }

    .tingle-modal--overflow {
        overflow-y: scroll;
        padding-top: 8vh;
    }

    /* btn
    -------------------------------------------------------------- */

    .tingle-btn {
        display: inline-block;
        margin: 0;
        padding: 1rem 2rem;
        border: none;
        background-color: grey;
        box-shadow: none;
        color: #fff;
        vertical-align: middle;
        text-decoration: none;
        font-size: inherit;
        font-family: inherit;
        line-height: normal;
        cursor: pointer;
        -webkit-transition: background-color .4s ease;
        transition: background-color .4s ease;
    }

    .tingle-btn--primary {
        background-color: #3498db;
    }

    .tingle-btn--danger {
        background-color: #e74c3c;
    }

    .tingle-btn--default {
        background-color: #34495e;
    }

    .tingle-btn--pull-left {
        float: left;
    }

    .tingle-btn--pull-right {
        float: right;
    }

    /* responsive
    -------------------------------------------------------------- */

    @media (max-width: 540px) {
        .tingle-modal {
            top: 0px;
            display: block;
            padding-top: 60px;
            width: 100%;
        }

        .tingle-modal-box {
            width: auto;
            border-radius: 0;
        }

        .tingle-modal-box__content {
            overflow-y: scroll;
        }

        .tingle-modal--noClose {
            top: 0;
        }

        .tingle-modal--noOverlayClose {
            padding-top: 0;
        }

        .tingle-modal-box__footer .tingle-btn {
            display: block;
            float: none;
            margin-bottom: 1rem;
            width: 100%;
        }

        .tingle-modal__close {
            top: 0;
            right: 0;
            left: 0;
            display: block;
            width: 100%;
            height: 60px;
            border: none;
            background-color: #2c3e50;
            box-shadow: none;
            color: #fff;
            line-height: 55px;
        }

        .tingle-modal__closeLabel {
            display: inline-block;
            vertical-align: middle;
            font-size: 1.5rem;
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Roboto", "Oxygen", "Ubuntu", "Cantarell", "Fira Sans", "Droid Sans", "Helvetica Neue", sans-serif;
        }

        .tingle-modal__closeIcon {
            display: inline-block;
            margin-right: .5rem;
            vertical-align: middle;
            font-size: 4rem;
        }
    }

    @supports ((-webkit-backdrop-filter: blur(12px)) or (backdrop-filter: blur(12px))) {
        .tingle-modal {
            -webkit-backdrop-filter: blur(20px);
            backdrop-filter: blur(20px);
        }

        @media (max-width: 540px) {
            .tingle-modal {
                -webkit-backdrop-filter: blur(8px);
                backdrop-filter: blur(8px);
            }
        }

        .tingle-enabled .tingle-content-wrapper {
            -webkit-filter: none;
            filter: none;
        }
    }

    .tingle-modal--visible {
        z-index: 1500;
    }

    .tingle-btn--warning {
        background-color: rgb(255, 193, 7);
        color: black;
    }

    /* purgecss end ignore */
</style>
