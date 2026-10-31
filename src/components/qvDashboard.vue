<template>
  <div class="main-content">
    <div class="dashboard-block dashboard-info">
      <h4 class="mt-2"><b>{{$t('dashboard.infoTitle')}}:</b></h4>
      <p>
        <b>{{$t('dashboard.infoTotalNotebooks')}}:</b> {{notebookCount}}<br>
        <b>{{$t('dashboard.infoTotalNotes')}}:</b> {{notesCountTotal}}<br>
        <b>{{$t('dashboard.infoTotalTags')}}:</b> {{tagsCount}}
      </p>
      <p>
        <b>{{$t('dashboard.infoFirstNote')}}:</b> {{ $filters.formatDate(statistic.dateFirst) }}<br>
        <b>{{$t('dashboard.infoLastChanges')}}:</b> {{ $filters.formatDate(statistic.dateLast) }}
      </p>
      <p>
        <b>{{$t('dashboard.infoSearchIndexSize')}}:</b> {{ $filters.formatBytes(statistic.dataSize) }}
      </p>
    </div>
    <div class="dashboard-block dashboard-chart">
      <h4 class="mt-2"><b>{{$t('dashboard.tagsPieTitle')}}:</b></h4>
      <VChart
        :option="statistic.chartOption"
        class="dashboard-chart-canvas"
        autoresize
      />

    </div>
    <div class="clearfix"></div>
    <div class="dashboard-block dashboard-activity">
      <h4 class="mt-4"><b>{{$t('dashboard.activityTitle')}}:</b></h4>
      <calendar-heatmap
        :values="statistic.calendarData"
        :end-date="new Date()"
        :range-color="['#ebedf0', '#c6e48b', '#7bc96f', '#24a53e', '#1f7a31', '#0e5b1d']"
        :locale="calendarLocale"
        :tooltip-unit="$t('dashboard.activityGraph.tooltipUnit')"
        :tooltip-formatter="formatTooltip"
      />
    </div>
  </div>
</template>

<script>
import { useNoteStore } from '../store'
import { push } from 'notivue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { LegendComponent, TooltipComponent } from 'echarts/components'
import CalendarHeatmap from './CalendarHeatmap.vue'

use([CanvasRenderer, PieChart, LegendComponent, TooltipComponent])

export default {
    name: 'qvDashboard',
    components: { VChart, CalendarHeatmap },
    data () {
        return {
            statistic: {
                dateFirst: 0,
                dateLast: 0,
                // 'chartsCreatedDate': {},
                dataSize: 0,
                chartOption: {
                    legend: {
                        left: 0,
                        top: 'middle',
                        orient: 'vertical'
                    },
                    tooltip: {
                        trigger: 'item'
                    },
                    series: [{
                        type: 'pie',
                        radius: 80,
                        center: ['50%', '50%'],
                        label: { show: false },
                        data: []
                    }]
                },
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
        calendarLocale () {
            const activityMessages = this.$tm('dashboard.activityGraph')
            return {
                months: activityMessages.months,
                days: activityMessages.days,
                on: this.$t('dashboard.activityGraph.on'),
                less: this.$t('dashboard.activityGraph.less'),
                more: this.$t('dashboard.activityGraph.more')
            }
        }
    },
    methods: {
        formatTooltip (value) {
            const date = new Date(value.date)
            const formattedDate = date.toLocaleDateString(this.$i18n.locale, {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
            })
            return `${formattedDate}: ${value.count} ${this.$t('dashboard.activityGraph.tooltipUnit')}`
        }
    },
    created: function () {
        fetch(this.noteStore.apiFolder + '/statistic.json')
            .then(response => response.json())
            .then(jsonData => {
                var pieData = []
                for (var elementC in jsonData.tagsCount) {
                    pieData.push({
                        name: elementC,
                        value: jsonData.tagsCount[elementC]
                    })
                }
                this.statistic.chartOption.series[0].data = pieData
                this.statistic.calendarData = Object.keys(jsonData.chartsUpdatedDate).map(date => ({
                    date,
                    count: jsonData.chartsUpdatedDate[date]
                }))
                this.statistic.dateFirst = jsonData.dateFirst
                this.statistic.dateLast = jsonData.dateLast
                this.statistic.dataSize = jsonData.dataSize
            })
            .catch(error => {
                console.error('Request failed', error)
                this.status = {
                    errorType: 2,
                    errorText: this.$t('dashboard.errorDownloadingStatistic')
                }
                push.error({ message: this.$t('dashboard.errorDownloadingStatistic'), duration: 7000 })
            })
    }
}
</script>
