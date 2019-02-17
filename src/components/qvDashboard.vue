<template>
  <div class="container-fluid pt-5">
    <div style="width:30rem; max-width:50%; float:left;">
      <h4><b>{{$t('dashboard.infoTitle')}}:</b></h4>
      <p>
        <b>{{$t('dashboard.infoTotalNotebooks')}}:</b> {{notebookCount}}<br>
        <b>{{$t('dashboard.infoTotalNotes')}}:</b> {{notesCountTotal}}<br>
        <b>{{$t('dashboard.infoTotalTags')}}:</b> {{tagsCount}}
      </p>
      <p>
        <b>{{$t('dashboard.infoFirstNote')}}:</b> {{statistic.dateFirst | formatDate}}<br>
        <b>{{$t('dashboard.infoLastChanges')}}:</b> {{statistic.dateLast | formatDate}}
      </p>
      <p>
        <b>{{$t('dashboard.infoSearchIndexSize')}}:</b> {{statistic.dataSize}}
      </p>
    </div>
    <div style="width:30rem; max-width:50%; float:left;">
      <h4 class="mb-3"><b>{{$t('dashboard.tagsPieTitle')}}:</b></h4>
      <ve-pie
        :data="statistic.tagsCountChartData"
        :settings="{radius:80,offsetY: '50%', label:{show:false}}"
        :extend="statistic.tagsCountchartSettings"
        width="100%"
        height="13rem"
      ></ve-pie>

    </div>
    <div class="clearfix"></div>
    <div style="width:60rem; max-width:100%;">
      <h4 class="mt-4"><b>{{$t('dashboard.activityTitle')}}:</b></h4>
      <calendar-heatmap
        :values="statistic.calendarData"
        :end-date="lastDay"
        :range-color="['#ebedf0', '#c6e48b', '#7bc96f', '#24a53e', '#1f7a31', '#0e5b1d']"
        :locale="{ months: $t('dashboard.activityGraph.months'), days: $t('dashboard.activityGraph.days'), on: $t('dashboard.activityGraph.on'), less: $t('dashboard.activityGraph.less'), more: $t('dashboard.activityGraph.more') }"
        :tooltip-unit="$t('dashboard.activityGraph.tooltipUnit')"
      />
    </div>
  </div>
</template>

<script>
import VePie from 'v-charts/lib/pie.common'
import { CalendarHeatmap } from 'vue-calendar-heatmap'

export default {
  name: 'qvDashboard',
  components: { VePie, CalendarHeatmap },
  data () {
    return {
      lastDay: new Intl.DateTimeFormat('en-US').format(new Date()),
      statistic: {
        dateFirst: 0,
        dateLast: 0,
        // 'chartsCreatedDate': {},
        dataSize: 0,
        tagsCountChartData: {
          columns: ['tags', 'cost'],
          rows: []
        },
        tagsCountchartSettings: {
          legend: {
            // 'right': 'left',
            left: 0,
            top: 'middle',
            orient: 'vertical',
            show: true
          }
        },
        calendarData: []
      }
    }
  },
  beforeCreate: function () {
    this.$http.get(this.$store.getters.apiFolder + '/statistic.json').then(
      response => {
        const tmpData = response.body
        for (var elementC in tmpData.tagsCount) {
          this.statistic.tagsCountChartData.rows.push({
            tags: elementC,
            cost: tmpData.tagsCount[elementC]
          })
        }
        for (var elementD in tmpData.chartsUpdatedDate) {
          this.statistic.calendarData.push({
            date: elementD,
            count: tmpData.chartsUpdatedDate[elementD]
          })
        }
        this.statistic.dateFirst = tmpData.dateFirst
        this.statistic.dateLast = tmpData.dateLast
        this.statistic.dataSize = tmpData.dataSize
      },
      response => {
        this.status = {
          errorType: 2,
          errorText: this.$t('dashboard.errorDownloadingStatistic')
        }
        console.info('Status:', response.status)
        console.info('Status text:', response.statusText)

        this.$toast.error({
          title: this.$t('dashboard.messageErrorTitle'),
          message: this.$t('dashboard.errorDownloadingStatistic'),
          closeButton: true,
          progressBar: true,
          timeOut: 7000
        })
      }
    )
  },
  computed: {
    notesCountTotal () {
      return this.$store.getters.getNotesCountTotal
    },
    notebookCount () {
      return this.$store.getters.getNotebooksCount
    },
    tagsCount () {
      return this.$store.getters.getTagsCount
    }
  }
}
</script>

<style>
</style>
