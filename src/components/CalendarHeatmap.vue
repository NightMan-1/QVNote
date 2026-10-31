<template>
    <div class="qvh" :class="{ 'qvh--scroll': needsScroll }">
        <div v-if="needsScroll" class="qvh__daylabels" aria-hidden="true">
            <span
                v-for="d in dayLabels"
                :key="d.index"
                :style="{ top: (top + d.index * step - 2) + 'px' }"
            >{{ d.label }}</span>
        </div>
        <div ref="scrollHost" class="qvh__scroll">
        <svg class="qvh__svg" :viewBox="`0 0 ${viewWidth} ${viewHeight}`" :style="needsScroll ? { width: viewWidth + 'px' } : null" role="img">
            <g class="qvh__months">
                <text v-for="m in monthLabels" :key="m.key" :x="m.x" y="10">{{ m.label }}</text>
            </g>
            <g v-if="!needsScroll" class="qvh__days">
                <text v-for="d in dayLabels" :key="d.index" x="0" :y="top + d.index * step + square - 1">{{ d.label }}</text>
            </g>
            <g class="qvh__grid">
                <g v-for="(week, wi) in weeksData" :key="wi" :transform="`translate(${left + wi * step}, ${top})`">
                    <rect
                        v-for="(day, di) in week"
                        :key="di"
                        :y="di * step"
                        :width="square"
                        :height="square"
                        rx="2"
                        :fill="day.visible ? colors[day.level] : 'transparent'"
                    >
                        <title v-if="day.visible">{{ tooltip(day) }}</title>
                    </rect>
                </g>
            </g>
            <g class="qvh__legend" :transform="`translate(${legendX}, ${viewHeight - 14})`">
                <text x="0" y="8">{{ locale.less }}</text>
                <rect
                    v-for="(color, ci) in colors"
                    :key="ci"
                    :x="legendLessWidth + ci * 12"
                    y="0"
                    width="10"
                    height="10"
                    rx="2"
                    :fill="color"
                />
                <text :x="legendLessWidth + colors.length * 12 + 6" y="8">{{ locale.more }}</text>
            </g>
        </svg>
        </div>
    </div>
</template>

<script>
/**
 * Compact calendar heatmap (replaces vue3-calendar-heatmap).
 *
 * The library hardcodes a 365-day range and always scales the whole year into
 * the available width, which makes the cells unreadably small on phones. This
 * component renders the same data as a plain SVG grid at its natural size and
 * switches to horizontal scrolling when the container is narrower than the
 * grid (scrolled to the most recent days). Native <title> elements provide
 * tooltips, which also removes the tippy.js dependency.
 */
export default {
    name: 'CalendarHeatmap',
    props: {
        values: { type: Array, default: () => [] },
        endDate: { type: [Date, String], default: () => new Date() },
        weeks: { type: Number, default: 52 },
        rangeColor: { type: Array, default: () => ['#ebedf0', '#c6e48b', '#7bc96f', '#24a53e', '#1f7a31'] },
        locale: { type: Object, default: () => ({ months: [], days: [], less: 'Less', more: 'More' }) },
        tooltipUnit: { type: String, default: 'notes' },
        tooltipFormatter: { type: Function, default: null }
    },
    data () {
        return {
            square: 10,
            gap: 3,
            left: 26,
            top: 18,
            legendLessWidth: 30,
            containerWidth: 0,
            observer: null
        }
    },
    computed: {
        step () {
            return this.square + this.gap
        },
        colors () {
            return this.rangeColor
        },
        viewWidth () {
            return this.left + this.weeks * this.step + 8
        },
        viewHeight () {
            // grid + a strip for the legend (it is drawn from the bottom up)
            return this.top + 7 * this.step + 30
        },
        needsScroll () {
            return this.containerWidth > 0 && this.containerWidth < this.viewWidth
        },
        legendX () {
            return Math.max(this.left, this.viewWidth - 8 - (this.legendLessWidth + this.colors.length * 12 + 34))
        },
        dayLabels () {
            const days = this.locale.days || []
            return [1, 3, 5]
                .filter(i => days[i] !== undefined)
                .map(i => ({ index: i, label: days[i] }))
        },
        valuesByDay () {
            const map = {}
            for (const item of this.values) {
                map[this.dayKey(this.parseDate(item.date))] = item.count
            }
            return map
        },
        maxCount () {
            return Math.max(1, ...this.values.map(v => v.count || 0))
        },
        weeksData () {
            const end = this.parseDate(this.endDate)
            const endDay = new Date(end.getFullYear(), end.getMonth(), end.getDate())
            // weeks start on Sunday, the last column contains the end date
            const lastWeekStart = new Date(endDay)
            lastWeekStart.setDate(lastWeekStart.getDate() - lastWeekStart.getDay())

            const weeks = []
            for (let w = this.weeks - 1; w >= 0; w--) {
                const week = []
                for (let d = 0; d < 7; d++) {
                    const date = new Date(lastWeekStart)
                    date.setDate(lastWeekStart.getDate() - w * 7 + d)
                    const visible = date <= endDay
                    const count = visible ? (this.valuesByDay[this.dayKey(date)] || 0) : 0
                    week.push({
                        date,
                        count,
                        visible,
                        level: count > 0 ? Math.min(this.colors.length - 1, Math.ceil(count / this.maxCount * (this.colors.length - 1))) : 0
                    })
                }
                weeks.push(week)
            }
            return weeks
        },
        monthLabels () {
            const months = this.locale.months || []
            const labels = []
            let lastMonth = -1
            this.weeksData.forEach((week, wi) => {
                for (const day of week) {
                    if (!day.visible) continue
                    const m = day.date.getMonth()
                    if (m !== lastMonth && day.date.getDate() <= 7 && wi < this.weeks - 1) {
                        labels.push({
                            key: `${day.date.getFullYear()}-${m}`,
                            x: this.left + wi * this.step,
                            label: months[m] || ''
                        })
                        lastMonth = m
                        break
                    }
                }
            })
            return labels
        }
    },
    watch: {
        weeksData () {
            this.scrollToEnd()
        }
    },
    mounted () {
        if (typeof ResizeObserver !== 'undefined') {
            this.observer = new ResizeObserver(entries => {
                const w = entries[0] && entries[0].contentRect ? entries[0].contentRect.width : 0
                this.containerWidth = Math.round(w)
                this.scrollToEnd()
            })
            this.observer.observe(this.$refs.scrollHost)
        }
        this.scrollToEnd()
    },
    beforeUnmount () {
        if (this.observer) {
            this.observer.disconnect()
            this.observer = null
        }
    },
    methods: {
        scrollToEnd () {
            this.$nextTick(() => {
                const el = this.$refs.scrollHost
                if (el && this.needsScroll) {
                    el.scrollLeft = el.scrollWidth
                }
            })
        },
        parseDate (value) {
            if (value instanceof Date) {
                return value
            }
            const [y, m, d] = String(value).split('-').map(Number)
            return new Date(y, (m || 1) - 1, d || 1)
        },
        dayKey (date) {
            return `${date.getFullYear()}-${date.getMonth() + 1}-${date.getDate()}`
        },
        tooltip (day) {
            if (this.tooltipFormatter) {
                return String(this.tooltipFormatter({ date: day.date, count: day.count })).replace(/<[^>]*>/g, '')
            }
            const month = (this.locale.months || [])[day.date.getMonth()] || ''
            return `${day.count} ${this.tooltipUnit} — ${day.date.getDate()} ${month} ${day.date.getFullYear()}`
        }
    }
}
</script>

<style scoped>
.qvh {
    max-width: 100%;
    position: relative;
}

.qvh__scroll {
    overflow-x: auto;
    overflow-y: hidden;
    -webkit-overflow-scrolling: touch;
    padding-bottom: .25rem;
}

/* weekday labels stay in place while the grid scrolls */
.qvh--scroll .qvh__scroll {
    padding-left: 0;
}

.qvh__daylabels {
    position: absolute;
    left: 0;
    top: 0;
    bottom: .25rem;
    width: 26px;
    background: #fff;
    z-index: 1;
}

.qvh__daylabels span {
    position: absolute;
    left: 0;
    font-size: 9px;
    color: #767676;
    line-height: 1;
}

.qvh__svg {
    width: 100%;
    max-width: 100%;
    height: auto;
    display: block;
}

.qvh--scroll .qvh__svg {
    max-width: none;
    margin-left: 26px;
}

.qvh__months text,
.qvh__days text,
.qvh__legend text {
    font-size: 9px;
    fill: #767676;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Oxygen, Ubuntu, Cantarell, "Open Sans", "Helvetica Neue", sans-serif;
}

.qvh__grid rect:hover {
    stroke: #555;
    stroke-width: 1px;
}
</style>
