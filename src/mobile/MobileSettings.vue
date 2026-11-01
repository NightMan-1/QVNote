<template>
    <div class="mset">
        <nav class="mset__tabs">
            <button
                v-for="item in sections"
                :key="item.id"
                class="mset__tab"
                :class="{ 'mset__tab--active': section === item.id }"
                @click="section = item.id"
            >
                <i class="bi" :class="item.icon"></i>
                <span>{{ item.label }}</span>
            </button>
        </nav>

        <div class="mset__body">
            <!-- ---------------------------------------------------------- -->
            <template v-if="section === 'global'">
                <section class="mset__card">
                    <h2 class="mset__title">{{ $t('setting.global.settingsTitle') }}</h2>

                    <div class="mset__stack">
                        <span class="mset__label">{{ $t('setting.global.folder') }}</span>
                        <span class="mset__path">{{ config.sourceFolder }}</span>
                    </div>

                    <label class="mset__switch">
                        <input v-model="checkNewNotes" type="checkbox">
                        <span>{{ $t('setting.global.switchCheckNew') }}</span>
                    </label>

                    <select v-model="langSelected" class="mset__select">
                        <option v-for="(lang, i) in localesList" :key="i" :value="i">{{ lang }}</option>
                    </select>
                </section>

                <section class="mset__card">
                    <h2 class="mset__title">{{ $t('setting.global.actionsTitle') }}</h2>

                    <div class="mset__row">
                        <span>{{ $t('setting.global.requestIndexing') }}:</span>
                        <strong :class="config.requestIndexing ? 'mset__danger' : 'mset__ok'">
                            {{ config.requestIndexing ? $t('general.yes') : $t('general.no') }}
                        </strong>
                    </div>

                    <button class="mset__btn mset__btn--primary" :disabled="busy" @click="refreshData('reload')">
                        {{ $t('setting.global.btnRefreshData') }}
                    </button>
                    <button class="mset__btn mset__btn--success" :disabled="busy" @click="indexingStart">
                        {{ $t('setting.global.btnIndexChanges') }}
                    </button>
                    <button class="mset__btn mset__btn--warning" :disabled="busy" @click="refreshData('reloadAll')">
                        {{ $t('setting.global.btnFullReload') }}
                    </button>

                    <div v-if="searchStatus.status === 'indexing'" class="mset__progress-block">
                        <p class="mset__hint">
                            {{ $t('setting.global.msgIndexing', [searchStatus.notesCurrent, searchStatus.notesTotal]) }}
                        </p>
                        <div class="mset__progress"><span :style="{ width: searchStatus.persent + '%' }"></span></div>
                    </div>
                    <p v-else-if="searchStatus.status === 'refresh'" class="mset__hint">
                        {{ $t('setting.global.msgSearchNewData') }}
                    </p>
                    <p v-else-if="searchStatus.status === 'done'" class="mset__hint mset__ok">
                        {{ $t('setting.global.msgSearchComplete') }}
                    </p>

                    <hr class="mset__hr">

                    <button class="mset__btn mset__btn--dark" :disabled="busy" @click="optimizationStart">
                        {{ $t('setting.global.btnDownloadImages') }}
                    </button>
                    <p class="mset__hint">
                        {{ $t('setting.global.downloadWarnLine1') }}
                        <span class="mset__danger">{{ $t('setting.global.downloadWarnLine2') }}</span>
                    </p>

                    <div v-if="optimizationStatus.status === 'processing'" class="mset__progress-block">
                        <p class="mset__hint">
                            {{ $t('setting.global.msgOptimizationStatus', [optimizationStatus.notesCurrent, optimizationStatus.notesTotal]) }}
                        </p>
                        <div class="mset__progress"><span :style="{ width: optimizationStatus.persent + '%' }"></span></div>
                    </div>
                    <p v-else-if="optimizationStatus.status === 'done'" class="mset__hint mset__ok">
                        {{ $t('setting.global.msgOptimizationComplete') }}
                    </p>

                    <hr class="mset__hr">

                    <button class="mset__btn mset__btn--outline" :disabled="webpCacheStatus === 'processing'" @click="webpCacheClear">
                        {{ $t('setting.global.btnWebpCacheClear') }}
                    </button>
                    <p class="mset__hint">{{ $t('setting.global.webpCacheHint') }}</p>
                    <p v-if="webpCacheStatus === 'done'" class="mset__hint mset__ok">{{ $t('setting.global.msgWebpCacheCleared') }}</p>
                    <p v-if="webpCacheStatus === 'error'" class="mset__hint mset__danger">{{ $t('setting.global.msgWebpCacheError') }}</p>
                </section>

                <section class="mset__card">
                    <h2 class="mset__title">{{ $t('setting.global.favorites') }}</h2>
                    <div class="mset__buttons">
                        <a
                            class="mset__btn mset__btn--primary"
                            :href="apiFolder + '/favorites.json'"
                            download="favorites.json"
                        >
                            <i class="bi bi-file-earmark-arrow-down"></i>
                            {{ $t('setting.global.btnFavoritesExport') }}
                        </a>
                        <label class="mset__btn mset__btn--success">
                            <i class="bi bi-file-earmark-arrow-up"></i>
                            {{ $t('setting.global.btnFavoritesImport') }}
                            <input type="file" accept="application/json" class="mset__file" @change="favoritesImportSelected">
                        </label>
                    </div>
                </section>
            </template>

            <!-- ---------------------------------------------------------- -->
            <template v-else-if="section === 'notebooks'">
                <h2 class="mset__title mset__title--page">{{ $t('setting.notebooks.title') }}</h2>
                <p class="mset__hint">{{ $t('setting.notebooks.tips') }}</p>

                <label class="mset__search">
                    <i class="bi bi-search"></i>
                    <input v-model="notebookFilter" type="search" :placeholder="$t('setting.filterPlaceholder')">
                </label>

                <div class="mset__list">
                    <button
                        v-for="item in filteredNotebooks"
                        :key="item.uuid"
                        class="mset__item"
                        @click="openEditor('notebook', item)"
                    >
                        <span class="mset__item-name">{{ item.name }}</span>
                        <i class="bi bi-gear"></i>
                    </button>
                    <p v-if="filteredNotebooks.length === 0" class="mset__hint">{{ $t('general.libraryEmpty') }}</p>
                </div>
            </template>

            <!-- ---------------------------------------------------------- -->
            <template v-else-if="section === 'tags'">
                <h2 class="mset__title mset__title--page">{{ $t('setting.tags.title') }}</h2>
                <p class="mset__hint">{{ $t('setting.tags.tips') }}</p>

                <label class="mset__search">
                    <i class="bi bi-search"></i>
                    <input v-model="tagFilter" type="search" :placeholder="$t('setting.filterPlaceholder')">
                </label>

                <div class="mset__list">
                    <button
                        v-for="item in filteredTags"
                        :key="item.url"
                        class="mset__item"
                        @click="openEditor('tag', item)"
                    >
                        <span class="mset__item-name">#{{ item.name }}</span>
                        <i class="bi bi-gear"></i>
                    </button>
                </div>
            </template>

            <!-- ---------------------------------------------------------- -->
            <template v-else-if="section === 'mcp'">
                <section class="mset__card">
                    <h2 class="mset__title">{{ $t('setting.mcp.title') }}</h2>
                    <p class="mset__hint">{{ $t('setting.mcp.tips') }}</p>

                    <label class="mset__switch">
                        <input v-model="mcpEnabled" type="checkbox" @change="saveMcpSettings">
                        <span>{{ $t('setting.mcp.enabled') }}</span>
                    </label>

                    <label class="mset__switch" :class="{ 'mset__switch--disabled': !mcpEnabled }">
                        <input v-model="mcpAllowWrite" type="checkbox" :disabled="!mcpEnabled" @change="saveMcpSettings">
                        <span>{{ $t('setting.mcp.allowWrite') }}</span>
                    </label>
                    <p class="mset__hint">{{ $t('setting.mcp.allowWriteHint') }}</p>

                    <div class="mset__stack" :class="{ 'mset__stack--disabled': !mcpEnabled }">
                        <span class="mset__label">{{ $t('setting.mcp.token') }}</span>
                        <input class="mset__input mset__input--mono" :value="mcpToken" readonly>
                        <button class="mset__btn mset__btn--warning" :disabled="!mcpEnabled" @click="mcpRegenToken">
                            {{ $t('setting.mcp.btnRegenToken') }}
                        </button>
                    </div>
                    <p class="mset__hint">{{ $t('setting.mcp.tokenHint') }}</p>

                    <div class="mset__stack">
                        <span class="mset__label">{{ $t('setting.mcp.endpointLabel') }}</span>
                        <code class="mset__code">{{ mcpEndpoint }}</code>
                    </div>
                    <p class="mset__hint">{{ $t('setting.mcp.endpointHint') }}</p>

                    <p v-if="mcpSaveStatus === 'done'" class="mset__hint mset__ok">{{ $t('setting.mcp.msgSaved') }}</p>
                    <p v-if="mcpSaveStatus === 'error'" class="mset__hint mset__danger">{{ $t('setting.mcp.msgError') }}</p>
                </section>
            </template>
        </div>

        <!-- rename / delete sheet -->
        <div v-if="editor" class="mset__backdrop" @click="closeEditor"></div>
        <div v-if="editor" class="mset__sheet">
            <div class="mset__sheet-grip"></div>
            <header class="mset__sheet-head">
                <div class="mset__sheet-head-text">
                    <div class="mset__sheet-kicker">
                        {{ editor.kind === 'notebook' ? $t('setting.notebooks.modalTitle') : $t('setting.tags.modalTitle') }}
                    </div>
                    <div class="mset__sheet-title">{{ editor.original }}</div>
                </div>
                <button class="mset__icon-btn" @click="closeEditor"><i class="bi bi-x-lg"></i></button>
            </header>

            <label class="mset__field">
                <span class="mset__label">
                    {{ editor.kind === 'notebook' ? $t('setting.notebooks.modalNewTitle') : $t('setting.tags.modalNewTitle') }}
                </span>
                <input v-model="editor.name" class="mset__input" @keyup.enter="saveEditor">
            </label>

            <div class="mset__sheet-actions">
                <button class="mset__btn mset__btn--danger" :disabled="editor.saving" @click="deleteEdited">
                    {{ editor.kind === 'notebook' ? $t('setting.notebooks.modalBtnDelete') : $t('setting.tags.modalBtnDelete') }}
                </button>
                <button class="mset__btn mset__btn--primary" :disabled="editor.saving" @click="saveEditor">
                    {{ editor.kind === 'notebook' ? $t('setting.notebooks.modalBtnSave') : $t('setting.tags.modalBtnSave') }}
                </button>
            </div>
        </div>
    </div>
</template>

<script>
import { useNoteStore } from '../store'
import { push } from 'notivue'

/**
 * Mobile settings screen: own markup, own styles (tokens from
 * src/mobile/tokens.css) and its own section state, so the phone settings can
 * evolve without touching the desktop qvSettings.
 */
export default {
    name: 'MobileSettings',
    data () {
        return {
            section: 'global',
            notebookFilter: '',
            tagFilter: '',
            checkNewNotes: false,
            langSelected: localStorage.getItem('locale') || 'ru',
            searchStatus: {
                notesCurrent: 0,
                notesTotal: 0,
                status: 'idle',
                persent: 0
            },
            optimizationStatus: {
                notesCurrent: 0,
                notesTotal: 0,
                status: 'idle',
                persent: 0
            },
            webpCacheStatus: 'idle',
            mcpEnabled: false,
            mcpAllowWrite: true,
            mcpToken: '',
            mcpSaveStatus: 'idle',
            editor: null
        }
    },
    setup () {
        return { noteStore: useNoteStore() }
    },
    computed: {
        apiFolder () { return this.noteStore.apiFolder },
        config () { return this.noteStore.config },
        localesList () { return this.noteStore.localesList },
        notebooksList () {
            const list = this.noteStore.notebooksList
            return Array.isArray(list) ? list : []
        },
        tagsList () {
            const list = this.noteStore.tagsList
            return Array.isArray(list) ? list : []
        },
        filteredNotebooks () {
            const query = this.notebookFilter.trim().toLowerCase()
            return query ? this.notebooksList.filter(n => n.name.toLowerCase().includes(query)) : this.notebooksList
        },
        filteredTags () {
            const query = this.tagFilter.trim().toLowerCase()
            return query ? this.tagsList.filter(t => t.name.toLowerCase().includes(query)) : this.tagsList
        },
        busy () {
            const status = this.searchStatus.status
            return status !== 'idle' && status !== 'done'
        },
        mcpEndpoint () { return window.location.origin + '/mcp' },
        sections () {
            return [
                { id: 'global', icon: 'bi-gear', label: this.$t('general.sidebarSettingsGeneral') },
                { id: 'notebooks', icon: 'bi-journal-text', label: this.$t('general.sidebarSettingsNotebooks') },
                { id: 'tags', icon: 'bi-tags', label: this.$t('general.sidebarSettingsTags') },
                { id: 'mcp', icon: 'bi-plug', label: this.$t('general.sidebarSettingsMCP') }
            ]
        }
    },
    watch: {
        checkNewNotes () {
            this.saveSettings()
        },
        langSelected () {
            localStorage.setItem('locale', this.langSelected)
            this.$i18n.locale = this.langSelected
        },
        'noteStore.config' (config) {
            this.mcpEnabled = !!config.mcpEnabled
            this.mcpAllowWrite = config.mcpAllowWrite !== false
            this.mcpToken = config.mcpToken || ''
        }
    },
    created () {
        this.checkNewNotes = !!this.noteStore.config.atStartCheckNewNotes
        this.mcpEnabled = !!this.noteStore.config.mcpEnabled
        this.mcpAllowWrite = this.noteStore.config.mcpAllowWrite !== false
        this.mcpToken = this.noteStore.config.mcpToken || ''
    },
    mounted () {
        this.loadData()
        this.optimizationGet()
        this.configTimer = setInterval(() => this.loadData(), 10000)
        this.optimizationTimer = setInterval(() => this.optimizationGet(), 3000)
    },
    beforeUnmount () {
        clearInterval(this.configTimer)
        clearInterval(this.optimizationTimer)
    },
    methods: {
        loadData () {
            fetch(this.apiFolder + '/search_index.json')
                .then(response => response.json())
                .then(jsonData => {
                    this.searchStatus.status = jsonData.status
                    this.searchStatus.notesTotal = jsonData.notesTotal
                    this.searchStatus.notesCurrent = jsonData.notesCurrent
                    this.searchStatus.persent = jsonData.notesTotal > 0
                        ? parseInt((jsonData.notesCurrent * 100) / jsonData.notesTotal)
                        : 0
                    this.noteStore.getConfig()
                })
                .catch(error => console.error('Error fetching search_index.json:', error))
        },
        optimizationGet () {
            fetch(this.apiFolder + '/optimization.json')
                .then(response => response.json())
                .then(jsonData => {
                    this.optimizationStatus.status = jsonData.status
                    this.optimizationStatus.notesTotal = jsonData.notesTotal
                    this.optimizationStatus.notesCurrent = jsonData.notesCurrent
                    this.optimizationStatus.persent = jsonData.notesTotal > 0
                        ? parseInt((jsonData.notesCurrent * 100) / jsonData.notesTotal)
                        : 0
                })
                .catch(error => console.error('Error fetching optimization.json:', error))
        },
        saveSettings () {
            fetch(this.apiFolder + '/config.json', {
                method: 'POST',
                body: JSON.stringify({ atStartCheckNewNotes: this.checkNewNotes.toString() })
            })
                .then(() => fetch(this.apiFolder + '/config.json'))
                .then(response => response.json())
                .then(jsonData => this.noteStore.setConfig(jsonData))
                .catch(error => console.error('Error saving config.json:', error))
        },
        saveMcpSettings () {
            this.mcpSaveStatus = 'idle'
            fetch(this.apiFolder + '/config.json', {
                method: 'POST',
                body: JSON.stringify({ mcpEnabled: this.mcpEnabled, mcpAllowWrite: this.mcpAllowWrite })
            })
                .then(response => response.json())
                .then(jsonData => {
                    this.noteStore.setConfig(jsonData)
                    this.mcpSaveStatus = 'done'
                })
                .catch(error => {
                    console.error('Error saving MCP config:', error)
                    this.mcpSaveStatus = 'error'
                })
        },
        mcpRegenToken () {
            this.mcpSaveStatus = 'idle'
            fetch(this.apiFolder + '/config.json', {
                method: 'POST',
                body: JSON.stringify({
                    mcpEnabled: this.mcpEnabled,
                    mcpAllowWrite: this.mcpAllowWrite,
                    mcpRegenerateToken: true
                })
            })
                .then(response => response.json())
                .then(jsonData => {
                    this.noteStore.setConfig(jsonData)
                    this.mcpSaveStatus = 'done'
                })
                .catch(error => {
                    console.error('Error regenerating MCP token:', error)
                    this.mcpSaveStatus = 'error'
                })
        },
        refreshData (action) {
            fetch(this.apiFolder + '/refresh_data.json', {
                method: 'POST',
                body: JSON.stringify({ action })
            })
                .then(() => fetch(this.apiFolder + '/config.json'))
                .then(response => response.json())
                .then(jsonData => {
                    this.noteStore.setConfig(jsonData)
                    if (action === 'reload') {
                        push.success({ message: this.$t('setting.global.notificationAddDataRefreshed') })
                    }
                })
                .catch(error => {
                    console.error('Error fetching refresh_data.json:', error)
                    push.error({ message: this.$t('setting.global.notificationErrorGetSearchStatus') })
                })
        },
        indexingStart () {
            fetch(this.apiFolder + '/search_index.json', {
                method: 'POST',
                body: JSON.stringify({ action: 'start' })
            })
        },
        optimizationStart () {
            fetch(this.apiFolder + '/optimization.json', {
                method: 'POST',
                body: JSON.stringify({ action: 'start' })
            })
        },
        webpCacheClear () {
            this.webpCacheStatus = 'processing'
            fetch(this.apiFolder + '/webp_cache_clear.json', { method: 'POST', body: '{}' })
                .then(response => response.json())
                .then(jsonData => { this.webpCacheStatus = jsonData.status === 'done' ? 'done' : 'error' })
                .catch(error => {
                    console.error('Error fetching webp_cache_clear.json:', error)
                    this.webpCacheStatus = 'error'
                })
        },
        favoritesImportSelected (event) {
            const file = event.target.files && event.target.files[0]
            if (!file) {
                return
            }
            if (file.type !== 'application/json') {
                push.error({ message: this.$t('setting.global.favoritesImportWrongType') })
                return
            }
            const reader = new FileReader()
            reader.onload = () => {
                try {
                    const list = JSON.parse(reader.result)
                    list.forEach(uuid => {
                        fetch(this.apiFolder + '/favorites.json', {
                            method: 'POST',
                            body: JSON.stringify({ action: 'add', UUID: uuid })
                        })
                    })
                    push.success({ message: this.$t('setting.global.favoritesImportDone') })
                } catch (e) {
                    push.error({ message: this.$t('setting.global.favoritesImportWrongData') })
                }
            }
            reader.readAsText(file)
        },
        openEditor (kind, item) {
            this.editor = {
                kind,
                uuid: item.uuid || '',
                url: item.url || '',
                original: item.name,
                name: item.name,
                saving: false
            }
        },
        closeEditor () {
            this.editor = null
        },
        saveEditor () {
            if (!this.editor || this.editor.saving) {
                return
            }
            const name = this.editor.name.trim()
            if (!name) {
                return
            }
            const isNotebook = this.editor.kind === 'notebook'
            const endpoint = isNotebook ? '/notebook_edit.json' : '/tag_edit.json'
            const body = isNotebook
                ? { action: 'rename', uuid: this.editor.uuid, title: name }
                : { action: 'rename', url: this.editor.url, title: name }

            this.editor.saving = true
            fetch(this.apiFolder + endpoint, { method: 'POST', body: JSON.stringify(body) })
                .then(() => {
                    this.noteStore.getAllData()
                    this.closeEditor()
                })
                .catch(error => {
                    this.editor.saving = false
                    console.error('Error renaming:', error)
                    push.error({
                        message: isNotebook
                            ? this.$t('setting.notebooks.statusErrorUpdate')
                            : this.$t('setting.tags.statusErrorUpdate')
                    })
                })
        },
        deleteEdited () {
            if (!this.editor || this.editor.saving) {
                return
            }
            const isNotebook = this.editor.kind === 'notebook'
            const endpoint = isNotebook ? '/notebook_edit.json' : '/tag_edit.json'
            const body = isNotebook
                ? { action: 'remove', uuid: this.editor.uuid }
                : { action: 'remove', url: this.editor.url }

            this.editor.saving = true
            fetch(this.apiFolder + endpoint, { method: 'POST', body: JSON.stringify(body) })
                .then(() => {
                    this.noteStore.getAllData()
                    this.closeEditor()
                })
                .catch(error => {
                    this.editor.saving = false
                    console.error('Error deleting:', error)
                    push.error({
                        message: isNotebook
                            ? this.$t('setting.notebooks.statusErrorDelete')
                            : this.$t('setting.tags.statusErrorDelete')
                    })
                })
        }
    }
}
</script>

<style scoped>
.mset {
    display: flex;
    flex-direction: column;
    height: 100%;
    background: var(--m-bg);
}

.mset__tabs {
    display: flex;
    flex: 0 0 auto;
    gap: .25rem;
    padding: .55rem .65rem;
    overflow-x: auto;
    border-bottom: 1px solid var(--m-border);
    background: var(--m-surface);
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
}

.mset__tabs::-webkit-scrollbar {
    display: none;
}

.mset__tab {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    gap: .35rem;
    padding: .45rem .75rem;
    border: 1px solid var(--m-border);
    border-radius: 999px;
    background: var(--m-bg);
    color: var(--m-text-muted);
    font-size: .88rem;
    white-space: nowrap;
}

.mset__tab--active {
    border-color: var(--m-accent);
    background: var(--m-accent-soft);
    color: var(--m-accent);
    font-weight: 600;
}

.mset__body {
    flex: 1 1 auto;
    min-height: 0;
    padding: .75rem;
    overflow-y: auto;
}

.mset__card {
    margin-bottom: .75rem;
    padding: .9rem 1rem 1rem;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius);
    background: var(--m-bg);
}

.mset__title {
    margin: 0 0 .75rem;
    color: var(--m-text);
    font-size: 1.02rem;
    font-weight: 700;
}

.mset__title--page {
    margin: .25rem 0 .35rem;
}

.mset__row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: .75rem;
    margin-bottom: .75rem;
}

.mset__stack {
    display: flex;
    flex-direction: column;
    gap: .35rem;
    margin-bottom: .75rem;
}

.mset__stack--disabled {
    opacity: .5;
}

.mset__label {
    color: var(--m-text-muted);
    font-size: .85rem;
}

.mset__path,
.mset__code {
    color: var(--m-text);
    font-family: var(--m-font-mono);
    font-size: .82rem;
    overflow-wrap: anywhere;
}

.mset__code {
    display: block;
    padding: .45rem .55rem;
    border-radius: var(--m-radius-sm);
    background: var(--m-surface);
}

.mset__switch {
    display: flex;
    align-items: center;
    gap: .65rem;
    padding: .5rem 0;
    color: var(--m-text);
    font-size: .95rem;
    line-height: 1.35;
}

.mset__switch input {
    position: relative;
    flex: 0 0 auto;
    width: 2.6rem;
    height: 1.5rem;
    margin: 0;
    border-radius: 999px;
    background: var(--m-surface-2);
    appearance: none;
    transition: background .15s ease;
}

.mset__switch input::after {
    content: '';
    position: absolute;
    top: .18rem;
    left: .18rem;
    width: 1.14rem;
    height: 1.14rem;
    border-radius: 50%;
    background: var(--m-bg);
    box-shadow: 0 1px 2px rgba(0, 0, 0, .2);
    transition: transform .15s ease;
}

.mset__switch input:checked {
    background: var(--m-accent);
}

.mset__switch input:checked::after {
    transform: translateX(1.1rem);
}

.mset__switch--disabled {
    opacity: .5;
}

.mset__select,
.mset__input {
    width: 100%;
    padding: .55rem .65rem;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius-sm);
    background: var(--m-bg);
    color: var(--m-text);
    font-size: .95rem;
}

.mset__input--mono {
    font-family: var(--m-font-mono);
    font-size: .8rem;
}

.mset__btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: .4rem;
    width: 100%;
    min-height: var(--m-tap);
    margin-bottom: .5rem;
    padding: .5rem .75rem;
    border: 1px solid transparent;
    border-radius: var(--m-radius-sm);
    color: #fff;
    font-size: .92rem;
    font-weight: 600;
    text-align: center;
    text-decoration: none;
}

.mset__btn:disabled {
    opacity: .5;
}

.mset__btn--primary { background: var(--m-accent); }
.mset__btn--success { background: #8fbc8f; color: #1f2a1f; }
.mset__btn--warning { background: #ebcb8b; color: #3b2f14; }
.mset__btn--dark { background: #2e3440; }
.mset__btn--danger { background: var(--m-danger); }
.mset__btn--outline {
    border-color: var(--m-danger);
    background: transparent;
    color: var(--m-danger);
}

.mset__buttons {
    display: flex;
    gap: .5rem;
}

.mset__buttons .mset__btn {
    margin-bottom: 0;
}

.mset__file {
    display: none;
}

.mset__hint {
    margin: .25rem 0 .6rem;
    color: var(--m-text-muted);
    font-size: .8rem;
    line-height: 1.4;
}

.mset__ok { color: #6d8f4e; }
.mset__danger { color: var(--m-danger); }

.mset__hr {
    margin: 1rem 0;
    border: 0;
    border-top: 1px solid var(--m-border);
}

.mset__progress-block {
    margin: .5rem 0 .75rem;
}

.mset__progress {
    height: .55rem;
    border-radius: 999px;
    background: var(--m-surface-2);
    overflow: hidden;
}

.mset__progress span {
    display: block;
    height: 100%;
    background: var(--m-accent);
    transition: width .2s ease;
}

.mset__search {
    display: flex;
    align-items: center;
    gap: .5rem;
    padding: .5rem .65rem;
    margin-bottom: .65rem;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius);
    background: var(--m-bg);
    color: var(--m-text-muted);
}

.mset__search input {
    flex: 1 1 auto;
    min-width: 0;
    border: 0;
    outline: none;
    background: none;
    font-size: .95rem;
    color: var(--m-text);
}

.mset__list {
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius);
    overflow: hidden;
    background: var(--m-bg);
}

.mset__item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: .75rem;
    width: 100%;
    min-height: var(--m-tap);
    padding: .7rem .9rem;
    border-bottom: 1px solid var(--m-surface);
    color: var(--m-text);
    font-size: .95rem;
    text-align: left;
}

.mset__item:last-child {
    border-bottom: 0;
}

.mset__item:active {
    background: var(--m-surface);
}

.mset__item-name {
    min-width: 0;
    overflow-wrap: anywhere;
}

.mset__item i {
    flex: 0 0 auto;
    color: var(--m-accent);
}

.mset__backdrop {
    position: fixed;
    inset: 0;
    z-index: 1100;
    background: rgba(46, 52, 64, .38);
}

.mset__sheet {
    position: fixed;
    inset: auto 0 var(--mshell-tabs, 0px) 0;
    z-index: 1200;
    width: min(100%, 36rem);
    max-height: calc(100vh - var(--mshell-tabs, 0px) - 2rem);
    margin: 0 auto;
    padding: .4rem .9rem 1rem;
    overflow-y: auto;
    border-radius: var(--m-radius-lg) var(--m-radius-lg) 0 0;
    background: var(--m-bg);
    box-shadow: var(--m-shadow-sheet);
}

.mset__sheet-grip {
    width: 2.25rem;
    height: .25rem;
    margin: .15rem auto .5rem;
    border-radius: 999px;
    background: var(--m-border);
}

.mset__sheet-head {
    display: flex;
    align-items: flex-start;
    gap: .5rem;
    margin-bottom: .75rem;
}

.mset__sheet-head-text {
    min-width: 0;
}

.mset__sheet-kicker {
    color: var(--m-text-muted);
    font-size: .78rem;
}

.mset__sheet-title {
    color: var(--m-text);
    font-size: 1.02rem;
    font-weight: 700;
    overflow-wrap: anywhere;
}

.mset__icon-btn {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    justify-content: center;
    width: var(--m-tap);
    height: var(--m-tap);
    margin-left: auto;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius-sm);
    color: var(--m-text-muted);
    font-size: 1.05rem;
}

.mset__field {
    display: flex;
    flex-direction: column;
    gap: .35rem;
    margin-bottom: .85rem;
}

.mset__sheet-actions {
    display: flex;
    gap: .5rem;
}

.mset__sheet-actions .mset__btn {
    margin-bottom: 0;
}
</style>
