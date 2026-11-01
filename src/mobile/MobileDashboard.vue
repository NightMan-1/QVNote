<template>
    <div class="mdash">
        <section class="mdash__card">
            <h2 class="mdash__title">{{ $t('dashboard.infoTitle') }}</h2>

            <div class="mdash__stats">
                <div class="mdash__stat">
                    <span class="mdash__stat-value">{{ notebookCount }}</span>
                    <span class="mdash__stat-label">{{ $t('dashboard.infoTotalNotebooks') }}</span>
                </div>
                <div class="mdash__stat">
                    <span class="mdash__stat-value">{{ notesCountTotal }}</span>
                    <span class="mdash__stat-label">{{ $t('dashboard.infoTotalNotes') }}</span>
                </div>
                <div class="mdash__stat">
                    <span class="mdash__stat-value">{{ tagsCount }}</span>
                    <span class="mdash__stat-label">{{ $t('dashboard.infoTotalTags') }}</span>
                </div>
            </div>

            <dl class="mdash__meta">
                <div class="mdash__meta-row">
                    <dt>{{ $t('dashboard.infoFirstNote') }}</dt>
                    <dd>{{ $filters.formatDate(statistic.dateFirst) }}</dd>
                </div>
                <div class="mdash__meta-row">
                    <dt>{{ $t('dashboard.infoLastChanges') }}</dt>
                    <dd>{{ $filters.formatDate(statistic.dateLast) }}</dd>
                </div>
                <div class="mdash__meta-row">
                    <dt>{{ $t('dashboard.infoSearchIndexSize') }}</dt>
                    <dd>{{ $filters.formatBytes(statistic.dataSize) }}</dd>
                </div>
            </dl>
        </section>

        <section v-if="tagRows.length" class="mdash__card">
            <h2 class="mdash__title">{{ $t('dashboard.tagsPieTitle') }}</h2>
            <div class="mdash__bars">
                <div v-for="row in tagRows" :key="row.name" class="mdash__bar">
                    <div class="mdash__bar-head">
                        <span>{{ row.count }} {{ plural(row.count, tagForms) }}</span>
                        <span class="mdash__bar-value">{{ row.notes }} {{ plural(row.notes, noteForms) }}</span>
                    </div>
                    <div class="mdash__bar-track">
                        <span :style="{ width: row.percent + '%' }"></span>
                    </div>
                </div>
            </div>
        </section>

        <section class="mdash__card mdash__card--flat">
            <h2 class="mdash__title">{{ $t('dashboard.activityTitle') }}</h2>
            <calendar-heatmap
                :values="statistic.calendarData"
                :end-date="new Date()"
                :range-color="['#ebedf0', '#c6e48b', '#7bc96f', '#24a53e', '#1f7a31', '#0e5b1d']"
                :locale="calendarLocale"
                :tooltip-unit="$t('dashboard.activityGraph.tooltipUnit')"
                :tooltip-formatter="formatTooltip"
            />
        </section>
    </div>
</template>

<script>
import { push } from 'notivue'
import { useNoteStore } from '../store'
import CalendarHeatmap from '../components/CalendarHeatmap.vue'

/**
 * Mobile overview: own layout and styles on top of the mobile tokens. The
 * data comes from the same /statistic.json as the desktop dashboard.
 */
export default {
    name: 'MobileDashboard',
    components: { CalendarHeatmap },
    data () {
        return {
            statistic: {
                dateFirst: 0,
                dateLast: 0,
                dataSize: 0,
                tagsCount: {},
                calendarData: []
            }
        }
    },
    setup () {
        return { noteStore: useNoteStore() }
    },
    computed: {
        notebookCount () { return this.noteStore.getNotebooksCount },
        notesCountTotal () { return this.noteStore.notesCountTotal },
        tagsCount () { return this.noteStore.getTagsCount },
        tagForms () { return this.$tm('dashboard.tagForms') },
        noteForms () { return this.$tm('dashboard.noteForms') },
        tagRows () {
            const entries = Object.entries(this.statistic.tagsCount || {})
                .map(([count, notes]) => ({ name: String(count), count: Number(count), notes: Number(notes) }))
                .sort((a, b) => a.count - b.count)
            const max = entries.reduce((acc, row) => Math.max(acc, row.notes), 0)
            return entries.map(row => ({ ...row, percent: max > 0 ? Math.round((row.notes * 100) / max) : 0 }))
        },
        calendarLocale () {
            const messages = this.$tm('dashboard.activityGraph')
            return {
                months: messages.months,
                days: messages.days,
                on: this.$t('dashboard.activityGraph.on'),
                less: this.$t('dashboard.activityGraph.less'),
                more: this.$t('dashboard.activityGraph.more')
            }
        }
    },
    created () {
        fetch(this.noteStore.apiFolder + '/statistic.json')
            .then(response => response.json())
            .then(jsonData => {
                this.statistic.tagsCount = jsonData.tagsCount || {}
                this.statistic.calendarData = Object.keys(jsonData.chartsUpdatedDate || {}).map(date => ({
                    date,
                    count: jsonData.chartsUpdatedDate[date]
                }))
                this.statistic.dateFirst = jsonData.dateFirst
                this.statistic.dateLast = jsonData.dateLast
                this.statistic.dataSize = jsonData.dataSize
            })
            .catch(error => {
                console.error('Request failed', error)
                push.error({ message: this.$t('dashboard.errorDownloadingStatistic'), duration: 7000 })
            })
    },
    methods: {
        // russian needs three forms, english only two — forms come from i18n
        plural (count, forms) {
            const list = Array.isArray(forms) ? forms : []
            if (list.length === 0) {
                return ''
            }
            if (this.$i18n.locale === 'ru') {
                const mod10 = count % 10
                const mod100 = count % 100
                if (mod10 === 1 && mod100 !== 11) {
                    return list[0]
                }
                if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) {
                    return list[1] || list[0]
                }
                return list[2] || list[1] || list[0]
            }
            return count === 1 ? list[0] : (list[2] || list[1] || list[0])
        },
        formatTooltip (value) {
            const date = new Date(value.date)
            const formattedDate = date.toLocaleDateString(this.$i18n.locale, {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            })
            return `${formattedDate}: ${value.count} ${this.$t('dashboard.activityGraph.tooltipUnit')}`
        }
    }
}
</script>

<style scoped>
.mdash {
    padding: .75rem;
}

.mdash__card {
    margin-bottom: .75rem;
    padding: .9rem 1rem 1rem;
    border: 1px solid var(--m-border);
    border-radius: var(--m-radius);
    background: var(--m-bg);
}

.mdash__card--flat {
    padding-bottom: .25rem;
}

.mdash__title {
    margin: 0 0 .75rem;
    color: var(--m-text);
    font-size: 1.02rem;
    font-weight: 700;
}

.mdash__stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: .5rem;
    margin-bottom: .9rem;
}

.mdash__stat {
    display: flex;
    flex-direction: column;
    gap: .1rem;
    padding: .6rem .5rem;
    border-radius: var(--m-radius-sm);
    background: var(--m-surface);
    text-align: center;
}

.mdash__stat-value {
    color: var(--m-accent);
    font-size: 1.4rem;
    font-weight: 700;
    font-variant-numeric: tabular-nums;
}

.mdash__stat-label {
    color: var(--m-text-muted);
    font-size: .72rem;
    line-height: 1.25;
}

.mdash__meta {
    margin: 0;
}

.mdash__meta-row {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: .75rem;
    padding: .35rem 0;
    border-top: 1px solid var(--m-surface);
}

.mdash__meta-row dt {
    color: var(--m-text-muted);
    font-size: .82rem;
    font-weight: 400;
}

.mdash__meta-row dd {
    margin: 0;
    color: var(--m-text);
    font-size: .85rem;
    text-align: right;
}

.mdash__bars {
    display: flex;
    flex-direction: column;
    gap: .6rem;
}

.mdash__bar-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: .75rem;
    margin-bottom: .25rem;
    font-size: .85rem;
    color: var(--m-text);
}

.mdash__bar-value {
    color: var(--m-text-muted);
    font-size: .8rem;
}

.mdash__bar-track {
    height: .5rem;
    border-radius: 999px;
    background: var(--m-surface-2);
    overflow: hidden;
}

.mdash__bar-track span {
    display: block;
    height: 100%;
    border-radius: 999px;
    background: var(--m-accent);
}

.mdash :deep(.qvh) {
    font-size: .68rem;
}
</style>
