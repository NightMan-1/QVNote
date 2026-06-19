<template>
  <div class="main-content">
    <div style="width:30rem; max-width:50%; float:left;">
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
    <div style="width:30rem; max-width:50%; float:left;">
      <h4 class="mt-2"><b>{{$t('dashboard.tagsPieTitle')}}:</b></h4>
      <VChart
        :option="statistic.chartOption"
        style="width: 100%; height: 13rem;"
        autoresize
      />

    </div>
    <div class="clearfix"></div>
    <div style="width:60rem; max-width:100%;">
      <h4 class="mt-4"><b>{{$t('dashboard.activityTitle')}}:</b></h4>
      <calendar-heatmap
        :values="statistic.calendarData"
        :end-date="new Intl.DateTimeFormat('en-US').format(new Date())"
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
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { PieChart } from 'echarts/charts'
import { LegendComponent, TooltipComponent } from 'echarts/components'
import { CalendarHeatmap } from 'vue3-calendar-heatmap'

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
                this.$toast.error(this.$t('dashboard.errorDownloadingStatistic'), { timeout: 7000 })
            })
    }
}
</script>
