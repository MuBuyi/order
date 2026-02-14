<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>数据可视化图表</template>
    <div style="margin-bottom:10px;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <el-radio-group v-model="mode" size="small">
        <el-radio-button label="daily">近7天</el-radio-button>
        <el-radio-button label="monthly">按月</el-radio-button>
      </el-radio-group>
      <el-date-picker
        v-if="mode === 'monthly'"
        v-model="selectedYear"
        type="year"
        size="small"
        value-format="YYYY"
        placeholder="选择年份"
      />
    </div>
    <div ref="trendRef" style="height:260px;width:100%;margin-bottom:20px;"></div>
    <div ref="topRef" style="height:220px;width:100%;"></div>
  </el-card>
</template>
<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'

const trendRef = ref(null)
const topRef = ref(null)

let trendChart
let topChart

const mode = ref('daily') // daily: 近7天, monthly: 按月
const selectedYear = ref('')

function formatDate(d) {
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function initDefaultDates() {
  const now = new Date()
  if (!selectedYear.value) {
    selectedYear.value = String(now.getFullYear())
  }
}

function renderTrend(option) {
  nextTick(() => {
    if (!trendChart && trendRef.value) {
      trendChart = echarts.init(trendRef.value)
    }
    if (!trendChart) return
    trendChart.setOption(option, true)
    trendChart.resize()
  })
}

function renderTop(data) {
  nextTick(() => {
    if (!topChart && topRef.value) {
      topChart = echarts.init(topRef.value)
    }
    if (!topChart) return
    if (!data || data.length === 0) {
      topChart.clear()
      topChart.setOption({
        title: { text: '商品销售排行', left: 'center' },
        xAxis: { type: 'category', data: [] },
        yAxis: { type: 'value' },
        series: [],
        graphic: {
          type: 'text',
          left: 'center',
          top: 'middle',
          style: { text: '暂无数据', fontSize: 18, fill: '#888' }
        }
      })
      return
    }
    topChart.setOption({
      title: { text: '商品销售排行', left: 'center' },
      xAxis: { type: 'category', data: data.map(i => i.ProductName) },
      yAxis: { type: 'value' },
      series: [{
        type: 'bar',
        data: data.map(i => i.Total),
        itemStyle: { color: '#67C23A' },
        label: { show: true, position: 'top' }
      }]
    })
    topChart.resize()
  })
}

async function loadTrend() {
  initDefaultDates()
  try {
    if (mode.value === 'daily') {
      const res = await axios.get('/api/stats/daily', { params: { days: 7 } })
      const data = Array.isArray(res.data) ? res.data : []
      const labels = data.map(i => (i.Day || '').split('T')[0] || i.Day || '')
      const values = data.map(i => Number(i.Total || 0))
      renderTrend({
        title: { text: '近7天销售金额', left: 'center' },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: labels },
        yAxis: { type: 'value' },
        series: [{
          type: 'line',
          smooth: true,
          data: values,
          areaStyle: {},
          label: { show: true, position: 'top' }
        }]
      })
    } else if (mode.value === 'monthly') {
      const year = selectedYear.value || String(new Date().getFullYear())
      const res = await axios.get('/api/stats/monthly', { params: { year } })
      const months = Array.from({ length: 12 }, (_, i) => `${i + 1}月`)
      const values = Array.isArray(res.data) ? res.data : []
      const seriesData = months.map((_, idx) => Number(values[idx] || 0))
      renderTrend({
        title: { text: `按月销售金额（${year}）`, left: 'center' },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: months },
        yAxis: { type: 'value' },
        series: [{
          type: 'line',
          smooth: true,
          data: seriesData,
          areaStyle: {},
          label: { show: true, position: 'top' }
        }]
      })
    }
  } catch (e) {
    // 接口异常时简单清空图表
    renderTrend({
      title: { text: '数据加载失败', left: 'center' },
      xAxis: { type: 'category', data: [] },
      yAxis: { type: 'value' },
      series: []
    })
  }
}

function loadTop() {
  axios.get('/api/stats/top-products').then(res => renderTop(res.data || [])).catch(() => {
    renderTop([])
  })
}

function load() {
  initDefaultDates()
  loadTrend()
  loadTop()
}

watch(mode, () => {
  loadTrend()
})

watch(selectedYear, () => {
  if (mode.value === 'monthly') {
    loadTrend()
  }
})

onMounted(load)

defineExpose({ load })
</script>
