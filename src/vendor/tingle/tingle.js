/*!
 * tingle.js 0.16.0 — vendored into QVNote
 * https://github.com/robinparisi/tingle — MIT, (c) 2015 Robin Parisi
 * See LICENSE in this directory.
 *
 * Local modifications (QVNote):
 *  - the "busy" flag is per instance (upstream had one module-level flag, so
 *    overlapping modals could silently swallow open()/close());
 *  - destroy() is idempotent and safe to call from onClose;
 *  - body scroll lock is reference-counted, so nested modals do not unlock
 *    the page when the top one closes;
 *  - accessibility: role="dialog"/aria-modal, the label is taken from the
 *    first heading of the content, focus moves into the dialog, Tab is
 *    trapped inside it and focus returns to the trigger on close;
 *  - Escape is handled via event.key (event.which is deprecated);
 *  - button labels are set with textContent instead of innerHTML.
 */
/* global define, module */
(function (root, factory) {
  if (typeof define === 'function' && define.amd) {
    define(factory)
  } else if (typeof exports === 'object') {
    module.exports = factory()
  } else {
    root.tingle = factory()
  }
}(this, function () {
  /* ----------------------------------------------------------- */
  /* == modal */
  /* ----------------------------------------------------------- */

  var bodyLock = {
    count: 0,
    position: 0
  }

  function Modal (options) {
    var defaults = {
      onClose: null,
      onOpen: null,
      beforeOpen: null,
      beforeClose: null,
      stickyFooter: false,
      footer: false,
      cssClass: [],
      closeLabel: 'Close',
      ariaLabel: null,
      closeMethods: ['overlay', 'button', 'escape']
    }

    // extends config
    this.opts = extend({}, defaults, options)

    // per-instance state (upstream used one module-level busy flag)
    this.busy = false
    this._destroyed = false

    // init modal
    this.init()
  }

  Modal.prototype.init = function () {
    if (this.modal) {
      return this
    }

    _build.call(this)
    _bindEvents.call(this)

    // insert modal in dom
    document.body.appendChild(this.modal, document.body.firstChild)

    if (this.opts.footer) {
      this.addFooter()
    }

    return this
  }

  Modal.prototype._busy = function (state) {
    this.busy = state
  }

  Modal.prototype._isBusy = function () {
    return this.busy === true
  }

  Modal.prototype.destroy = function () {
    // safe to call twice, and safe to call from an onClose handler
    if (!this.modal || this._destroyed) {
      return
    }
    this._destroyed = true

    // restore scrolling
    if (this.isOpen()) {
      this.close(true)
    }

    // unbind all events
    _unbindEvents.call(this)

    // remove modal from dom
    if (this.modal.parentNode) {
      this.modal.parentNode.removeChild(this.modal)
    }

    this.modal = null
  }

  Modal.prototype.isOpen = function () {
    return !!this.modal && this.modal.classList.contains('tingle-modal--visible')
  }

  Modal.prototype.open = function () {
    if (this._isBusy()) return
    this._busy(true)

    var self = this

    // before open callback
    if (typeof self.opts.beforeOpen === 'function') {
      self.opts.beforeOpen()
    }

    if (this.modal.style.removeProperty) {
      this.modal.style.removeProperty('display')
    } else {
      this.modal.style.removeAttribute('display')
    }

    // prevent text selection when opening multiple times
    document.getSelection().removeAllRanges()

    // remember the trigger so focus can go back there on close
    this._previousFocus = document.activeElement

    // prevent double scroll (reference-counted for nested modals)
    _lockBodyScroll()

    // sticky footer
    this.setStickyFooter(this.opts.stickyFooter)

    // show modal
    this.modal.classList.add('tingle-modal--visible')

    // onOpen callback
    if (typeof self.opts.onOpen === 'function') {
      self.opts.onOpen.call(self)
    }

    self._busy(false)

    // move focus into the dialog (on the next frame, so it does not jump
    // while the opening animation runs)
    _focusFirst.call(self)

    // check if modal is bigger than screen height
    this.checkOverflow()

    return this
  }

  Modal.prototype.close = function (force) {
    if (this._isBusy()) return
    this._busy(true)
    force = force || false

    //  before close
    if (typeof this.opts.beforeClose === 'function') {
      var close = this.opts.beforeClose.call(this)
      if (!close) {
        this._busy(false)
        return
      }
    }

    _unlockBodyScroll()

    this.modal.classList.remove('tingle-modal--visible')

    // using similar setup as onOpen
    var self = this

    self.modal.style.display = 'none'

    // give focus back to the element that opened the dialog
    if (this._previousFocus && document.contains(this._previousFocus) && typeof this._previousFocus.focus === 'function') {
      this._previousFocus.focus({ preventScroll: true })
    }

    // onClose callback
    if (typeof self.opts.onClose === 'function') {
      self.opts.onClose.call(this)
    }

    // release modal
    self._busy(false)
  }

  Modal.prototype.setContent = function (content) {
    // check type of content : String or Node
    if (typeof content === 'string') {
      this.modalBoxContent.innerHTML = content
      _applyAriaLabel.call(this)
    } else {
      this.modalBoxContent.innerHTML = ''
      this.modalBoxContent.appendChild(content)
      _applyAriaLabel.call(this)
    }

    if (this.isOpen()) {
      // check if modal is bigger than screen height
      this.checkOverflow()
    }

    return this
  }

  Modal.prototype.getContent = function () {
    return this.modalBoxContent
  }

  Modal.prototype.addFooter = function () {
    // add footer to modal
    _buildFooter.call(this)

    return this
  }

  Modal.prototype.setFooterContent = function (content) {
    // set footer content
    this.modalBoxFooter.innerHTML = content

    return this
  }

  Modal.prototype.getFooterContent = function () {
    return this.modalBoxFooter
  }

  Modal.prototype.setStickyFooter = function (isSticky) {
    // if the modal is smaller than the viewport height, we don't need sticky
    if (!this.isOverflow()) {
      isSticky = false
    }

    if (isSticky) {
      if (this.modalBox.contains(this.modalBoxFooter)) {
        this.modalBox.removeChild(this.modalBoxFooter)
        this.modal.appendChild(this.modalBoxFooter)
        this.modalBoxFooter.classList.add('tingle-modal-box__footer--sticky')
        _recalculateFooterPosition.call(this)
        this.modalBoxContent.style['padding-bottom'] = this.modalBoxFooter.clientHeight + 20 + 'px'
      }
    } else if (this.modalBoxFooter) {
      if (!this.modalBox.contains(this.modalBoxFooter)) {
        this.modal.removeChild(this.modalBoxFooter)
        this.modalBox.appendChild(this.modalBoxFooter)
        this.modalBoxFooter.style.width = 'auto'
        this.modalBoxFooter.style.left = ''
        this.modalBoxContent.style['padding-bottom'] = ''
        this.modalBoxFooter.classList.remove('tingle-modal-box__footer--sticky')
      }
    }

    return this
  }

  Modal.prototype.addFooterBtn = function (label, cssClass, callback) {
    var btn = document.createElement('button')

    // set label (plain text: labels come from i18n strings)
    btn.textContent = label

    // bind callback
    btn.addEventListener('click', callback)

    if (typeof cssClass === 'string' && cssClass.length) {
      // add classes to btn
      cssClass.split(' ').forEach(function (item) {
        btn.classList.add(item)
      })
    }

    this.modalBoxFooter.appendChild(btn)

    return btn
  }

  Modal.prototype.resize = function () {
    // eslint-disable-next-line no-console
    console.warn('Resize is deprecated and will be removed in version 1.0')
  }

  Modal.prototype.isOverflow = function () {
    var viewportHeight = window.innerHeight
    var modalHeight = this.modalBox.clientHeight

    return modalHeight >= viewportHeight
  }

  Modal.prototype.checkOverflow = function () {
    // only if the modal is currently shown
    if (this.modal.classList.contains('tingle-modal--visible')) {
      if (this.isOverflow()) {
        this.modal.classList.add('tingle-modal--overflow')
      } else {
        this.modal.classList.remove('tingle-modal--overflow')
      }

      // tODO: remove offset
      // _offset.call(this);
      if (!this.isOverflow() && this.opts.stickyFooter) {
        this.setStickyFooter(false)
      } else if (this.isOverflow() && this.opts.stickyFooter) {
        _recalculateFooterPosition.call(this)
        this.setStickyFooter(true)
      }
    }
  }

  /* ----------------------------------------------------------- */
  /* == private methods */
  /* ----------------------------------------------------------- */

  function closeIcon () {
    return '<svg viewBox="0 0 10 10" xmlns="http://www.w3.org/2000/svg"><path d="M.3 9.7c.2.2.4.3.7.3.3 0 .5-.1.7-.3L5 6.4l3.3 3.3c.2.2.5.3.7.3.2 0 .5-.1.7-.3.4-.4.4-1 0-1.4L6.4 5l3.3-3.3c.4-.4.4-1 0-1.4-.4-.4-1-.4-1.4 0L5 3.6 1.7.3C1.3-.1.7-.1.3.3c-.4.4-.4 1 0 1.4L3.6 5 .3 8.3c-.4.4-.4 1 0 1.4z" fill="#000" fill-rule="nonzero"/></svg>'
  }

  function _recalculateFooterPosition () {
    if (!this.modalBoxFooter) {
      return
    }
    this.modalBoxFooter.style.width = this.modalBox.clientWidth + 'px'
    this.modalBoxFooter.style.left = this.modalBox.offsetLeft + 'px'
  }

  function _build () {
    // wrapper
    this.modal = document.createElement('div')
    this.modal.classList.add('tingle-modal')
    this.modal.setAttribute('role', 'dialog')
    this.modal.setAttribute('aria-modal', 'true')
    if (this.opts.ariaLabel) {
      this.modal.setAttribute('aria-label', this.opts.ariaLabel)
    }

    // remove cusor if no overlay close method
    if (this.opts.closeMethods.length === 0 || this.opts.closeMethods.indexOf('overlay') === -1) {
      this.modal.classList.add('tingle-modal--noOverlayClose')
    }

    this.modal.style.display = 'none'

    // custom class
    this.opts.cssClass.forEach(function (item) {
      if (typeof item === 'string') {
        this.modal.classList.add(item)
      }
    }, this)

    // close btn
    if (this.opts.closeMethods.indexOf('button') !== -1) {
      this.modalCloseBtn = document.createElement('button')
      this.modalCloseBtn.type = 'button'
      this.modalCloseBtn.classList.add('tingle-modal__close')

      this.modalCloseBtnIcon = document.createElement('span')
      this.modalCloseBtnIcon.classList.add('tingle-modal__closeIcon')
      this.modalCloseBtnIcon.innerHTML = closeIcon()

      this.modalCloseBtnLabel = document.createElement('span')
      this.modalCloseBtnLabel.classList.add('tingle-modal__closeLabel')
      this.modalCloseBtnLabel.textContent = this.opts.closeLabel

      this.modalCloseBtn.appendChild(this.modalCloseBtnIcon)
      this.modalCloseBtn.appendChild(this.modalCloseBtnLabel)
    }

    // modal
    this.modalBox = document.createElement('div')
    this.modalBox.classList.add('tingle-modal-box')
    this.modalBox.setAttribute('tabindex', '-1')

    // modal box content
    this.modalBoxContent = document.createElement('div')
    this.modalBoxContent.classList.add('tingle-modal-box__content')

    this.modalBox.appendChild(this.modalBoxContent)

    if (this.opts.closeMethods.indexOf('button') !== -1) {
      this.modal.appendChild(this.modalCloseBtn)
    }

    this.modal.appendChild(this.modalBox)
  }

  function _buildFooter () {
    this.modalBoxFooter = document.createElement('div')
    this.modalBoxFooter.classList.add('tingle-modal-box__footer')
    this.modalBox.appendChild(this.modalBoxFooter)
  }

  function _bindEvents () {
    this._events = {
      clickCloseBtn: this.close.bind(this),
      clickOverlay: _handleClickOutside.bind(this),
      resize: this.checkOverflow.bind(this),
      keyboardNav: _handleKeyboardNav.bind(this)
    }

    if (this.opts.closeMethods.indexOf('button') !== -1) {
      this.modalCloseBtn.addEventListener('click', this._events.clickCloseBtn)
    }

    this.modal.addEventListener('mousedown', this._events.clickOverlay)
    window.addEventListener('resize', this._events.resize)
    document.addEventListener('keydown', this._events.keyboardNav)
  }

  function _handleKeyboardNav (event) {
    if (!this.isOpen()) {
      return
    }

    // escape key
    var isEscape = event.key === 'Escape' || event.key === 'Esc' || event.which === 27
    if (isEscape && this.opts.closeMethods.indexOf('escape') !== -1) {
      event.preventDefault()
      this.close()
      return
    }

    // keep Tab inside the dialog
    var isTab = event.key === 'Tab' || event.keyCode === 9
    if (!isTab) {
      return
    }

    var focusable = _focusableElements(this.modal)
    if (!focusable.length) {
      event.preventDefault()
      return
    }

    var first = focusable[0]
    var last = focusable[focusable.length - 1]

    if (!this.modal.contains(document.activeElement)) {
      event.preventDefault()
      first.focus()
      return
    }

    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault()
      last.focus()
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault()
      first.focus()
    }
  }

  var FOCUSABLE_SELECTOR = [
    'a[href]',
    'area[href]',
    'button:not([disabled])',
    'input:not([disabled]):not([type="hidden"])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    'iframe',
    '[tabindex]:not([tabindex="-1"])',
    '[contenteditable="true"]'
  ].join(',')

  function _focusableElements (root) {
    return Array.prototype.filter.call(root.querySelectorAll(FOCUSABLE_SELECTOR), function (el) {
      return !!(el.offsetWidth || el.offsetHeight || el.getClientRects().length)
    })
  }

  function _focusFirst () {
    var target = _focusableElements(this.modalBoxContent)[0] ||
      (this.modalBoxFooter ? _focusableElements(this.modalBoxFooter)[0] : null) ||
      this.modalBox

    window.requestAnimationFrame(function () {
      // the modal may have been closed again before the frame fires
      if (target && document.contains(target) && typeof target.focus === 'function') {
        target.focus({ preventScroll: true })
      }
    })
  }

  function _applyAriaLabel () {
    if (this.opts.ariaLabel || !this.modal) {
      return
    }
    var heading = this.modalBoxContent.querySelector('h1, h2, h3, h4, h5, h6')
    var label = heading ? heading.textContent.replace(/\s+/g, ' ').trim() : ''
    if (label) {
      this.modal.setAttribute('aria-label', label)
    }
  }

  function _lockBodyScroll () {
    if (bodyLock.count === 0) {
      bodyLock.position = window.pageYOffset
      document.body.classList.add('tingle-enabled')
      document.body.style.top = -bodyLock.position + 'px'
    }
    bodyLock.count++
  }

  function _unlockBodyScroll () {
    if (bodyLock.count === 0) {
      return
    }
    bodyLock.count--
    if (bodyLock.count === 0) {
      document.body.classList.remove('tingle-enabled')
      document.body.style.top = null
      window.scrollTo({ top: bodyLock.position, behavior: 'instant' })
    }
  }

  function _handleClickOutside (event) {
    // on macOS, click on scrollbar (hidden mode) will trigger close event so we need to bypass this behavior by detecting scrollbar mode
    var scrollbarWidth = this.modal.offsetWidth - this.modal.clientWidth
    var clickedOnScrollbar = event.clientX >= this.modal.offsetWidth - 15 // 15px is macOS scrollbar default width
    var isScrollable = this.modal.scrollHeight !== this.modal.offsetHeight
    if (navigator.platform === 'MacIntel' && scrollbarWidth === 0 && clickedOnScrollbar && isScrollable) {
      return
    }

    // if click is outside the modal
    if (this.opts.closeMethods.indexOf('overlay') !== -1 && !_findAncestor(event.target, 'tingle-modal') &&
        event.clientX < this.modal.clientWidth) {
      this.close()
    }
  }

  function _findAncestor (el, cls) {
    while ((el = el.parentElement) && !el.classList.contains(cls));
    return el
  }

  function _unbindEvents () {
    if (this.opts.closeMethods.indexOf('button') !== -1) {
      this.modalCloseBtn.removeEventListener('click', this._events.clickCloseBtn)
    }
    this.modal.removeEventListener('mousedown', this._events.clickOverlay)
    window.removeEventListener('resize', this._events.resize)
    document.removeEventListener('keydown', this._events.keyboardNav)
  }

  /* ----------------------------------------------------------- */
  /* == helpers */
  /* ----------------------------------------------------------- */

  function extend () {
    for (var i = 1; i < arguments.length; i++) {
      for (var key in arguments[i]) {
        if (arguments[i].hasOwnProperty(key)) {
          arguments[0][key] = arguments[i][key]
        }
      }
    }
    return arguments[0]
  }

  /* ----------------------------------------------------------- */
  /* == return */
  /* ----------------------------------------------------------- */

  return {
    modal: Modal
  }
}))
