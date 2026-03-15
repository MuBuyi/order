<template>
  <el-card shadow="hover">
    <template #header>
      <div class="stats-header">
        <span class="stats-header-title">图表统计</span>
        <div class="stats-header-filters">
          <span>时间范围：</span>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            size="small"
            value-format="YYYY-MM-DD"
            :unlink-panels="true"
            :clearable="false"
            :disabled-date="disableOverOneMonth"
          />
        </div>
      </div>
    </template>

    <div style="display:flex;flex-direction:column;gap:24px;">
      <!-- 1-3: 销售总额、广告费用折算、利润（折线/柱状合并图） -->
      <div>
        <div ref="metricsRef" style="width:100%;height:320px;"></div>
      </div>

      <!-- 整体指标柱状图（当前时间范围汇总） -->
      <div>
        <div ref="metricsOverallRef" style="width:100%;height:260px;"></div>
      </div>

      <!-- 3: 商品销售排行 -->
      <div>
        <div ref="productsRef" style="width:100%;height:280px;"></div>
      </div>

      <!-- 5: 店铺每日广告费用 -->
      <div>
        <div ref="storeAdsRef" style="width:100%;height:320px;"></div>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted, watch, nextTick, onBeforeUnmount, computed } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'

const dateRange = ref([])

const metricsRef = ref(null)
const metricsOverallRef = ref(null)
const productsRef = ref(null)
const storeAdsRef = ref(null)

let metricsChart
let metricsOverallChart
let productsChart
let storeAdsChart

const isSingleDay = computed(() => {
  return Array.isArray(dateRange.value) && dateRange.value.length === 2 && dateRange.value[0] === dateRange.value[1]
})

function disableOverOneMonth(date) {
  // 限制选择范围不超过 31 天；Element Plus 的 disabled-date 针对单个面板日期，这里仅做简单限制：不允许选择今天之后
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  return date.getTime() > today.getTime()
}

function initDefaultRange() {
  if (!Array.isArray(dateRange.value) || dateRange.value.length !== 2) {
    const end = new Date()
    const start = new Date()
    start.setDate(end.getDate() - 29)

    const fmt = d => {
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      return `${y}-${m}-${day}`
    }
    dateRange.value = [fmt(start), fmt(end)]
  }
}

function getRangeParams() {
  initDefaultRange()
  if (!Array.isArray(dateRange.value) || dateRange.value.length !== 2) return {}
  const [start, end] = dateRange.value
  return { start_date: start, end_date: end }
}

function ensureMetricsChart() {
  if (!metricsChart && metricsRef.value) {
    metricsChart = echarts.init(metricsRef.value)
  }
  return metricsChart
}

function ensureMetricsOverallChart() {
  if (!metricsOverallChart && metricsOverallRef.value) {
    metricsOverallChart = echarts.init(metricsOverallRef.value)
  }
  return metricsOverallChart
}

function ensureProductsChart() {
  if (!productsChart && productsRef.value) {
    productsChart = echarts.init(productsRef.value)
  }
  return productsChart
}

function ensureStoreAdsChart() {
  if (!storeAdsChart && storeAdsRef.value) {
    storeAdsChart = echarts.init(storeAdsRef.value)
  }
  return storeAdsChart
}

function renderMetrics(data) {
  nextTick(() => {
    const chart = ensureMetricsChart()
    if (!chart) return

    const days = data.days || []
    const sale = data.sale_total || []
    const ad = data.ad_deduction || []
    const profit = data.profit || []

    if (!days.length) {
      chart.clear()
      chart.setOption({
        title: { text: '销售/广告/利润', left: 'center', top: 10 },
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
      chart.resize()
      return
    }

    if (isSingleDay.value) {
      const target = days[days.length - 1]
      const idx = days.indexOf(target)
      const vSale = idx >= 0 ? sale[idx] || 0 : 0
      const vAd = idx >= 0 ? ad[idx] || 0 : 0
      const vProfit = idx >= 0 ? profit[idx] || 0 : 0
      chart.setOption({
        title: { text: `${target} 指标对比`, left: 'center', top: 10 },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: ['销售总额', '广告费用折算', '利润'] },
        yAxis: { type: 'value' },
        series: [{
          name: '金额',
          type: 'bar',
          data: [vSale, vAd, vProfit],
          itemStyle: { color: '#409EFF' },
          label: { show: true, position: 'top' }
        }]
      })
    } else {
      chart.setOption({
        title: { text: '销售总额 / 广告费用折算 / 利润（按日）', left: 'center', top: 10 },
        tooltip: { trigger: 'axis' },
        legend: { top: 40 },
        grid: { top: 80, left: '8%', right: '4%', bottom: 40, containLabel: true },
        xAxis: { type: 'category', data: days },
        yAxis: { type: 'value' },
        series: [
          {
            name: '销售总额',
            type: 'line',
            smooth: true,
            data: sale,
            areaStyle: {},
          },
          {
            name: '广告费用折算',
            type: 'line',
            smooth: true,
            data: ad,
            areaStyle: {},
          },
          {
            name: '利润',
            type: 'line',
            smooth: true,
            data: profit,
            areaStyle: {},
          }
        ]
      })
    }
    chart.resize()

    // 同时渲染整体指标柱状图：当前时间范围内三项指标总和
    const sum = arr => (Array.isArray(arr) ? arr.reduce((acc, v) => acc + (Number(v) || 0), 0) : 0)
    const totalSale = sum(sale)
    const totalAd = sum(ad)
    const totalProfit = sum(profit)
    renderMetricsOverall({ totalSale, totalAd, totalProfit })
  })
}

function renderMetricsOverall(payload) {
  nextTick(() => {
    const chart = ensureMetricsOverallChart()
    if (!chart) return

    const data = [
      Number(payload.totalSale || 0),
      Number(payload.totalAd || 0),
      Number(payload.totalProfit || 0)
    ]

    if (data.every(v => !v)) {
      chart.clear()
      chart.setOption({
        title: { text: '整体指标汇总（当前时间范围）', left: 'center' },
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
      chart.resize()
      return
    }

    chart.setOption({
      title: { text: '整体指标汇总（当前时间范围）', left: 'center' },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: ['销售总额', '广告费用折算', '利润'] },
      yAxis: { type: 'value' },
      series: [{
        type: 'bar',
        data,
        itemStyle: { color: '#909399' },
        label: { show: true, position: 'top' }
      }]
    })
    chart.resize()
  })
}

function renderProducts(list) {
  nextTick(() => {
    const chart = ensureProductsChart()
    if (!chart) return

    if (!Array.isArray(list) || list.length === 0) {
      chart.clear()
      chart.setOption({
        title: { text: '商品销售排行', left: 'center', top: 10 },
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
      chart.resize()
      return
    }

    const names = list.map(i => i.ProductName || i.product_name)
    const values = list.map(i => Number(i.Total || i.total || 0))

    chart.setOption({
      title: { text: '商品销售排行', left: 'center', top: 10 },
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: names, axisLabel: { interval: 0, rotate: 30 } },
      yAxis: { type: 'value' },
      series: [{
        type: 'bar',
        data: values,
        itemStyle: { color: '#67C23A' },
        label: { show: true, position: 'top' }
      }]
    })
    chart.resize()
  })
}

function renderStoreAds(payload) {
  nextTick(() => {
    const chart = ensureStoreAdsChart()
    if (!chart) return

    const items = Array.isArray(payload.items) ? payload.items : []
    if (items.length === 0) {
      chart.clear()
      chart.setOption({
        title: { text: '店铺每日广告费用（按日）', left: 'center', top: 10 },
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
      chart.resize()
      return
    }

    const datesSet = new Set()
    const storeSet = new Set()
    items.forEach(i => {
      datesSet.add(i.date || i.Date)
      storeSet.add(i.store_name || i.StoreName || `店铺${i.store_id || i.StoreID}`)
    })

    const dates = Array.from(datesSet).sort()
    const stores = Array.from(storeSet)

    if (!isSingleDay.value) {
      const series = stores.map(storeName => {
        const data = dates.map(d => {
          const found = items.find(i => (i.date || i.Date) === d && (i.store_name || i.StoreName || `店铺${i.store_id || i.StoreID}`) === storeName)
          return found ? Number(found.ad_cost || found.AdCost || 0) : 0
        })
        return {
          name: storeName,
          type: 'line',
          smooth: true,
          data,
        }
      })

      chart.setOption({
        title: { text: '店铺每日广告费用（按日）', left: 'center', top: 10 },
        tooltip: { trigger: 'axis' },
        legend: { top: 40 },
        grid: { top: 80, left: '8%', right: '4%', bottom: 40, containLabel: true },
        xAxis: { type: 'category', data: dates },
        yAxis: { type: 'value' },
        series
      })
    } else {
      const target = dates.sort()[dates.length - 1]
      const filtered = items.filter(i => (i.date || i.Date) === target)
      const names = filtered.map(i => i.store_name || i.StoreName || `店铺${i.store_id || i.StoreID}`)
      const values = filtered.map(i => Number(i.ad_cost || i.AdCost || 0))
      chart.setOption({
        title: { text: `${target} 店铺广告费用`, left: 'center', top: 10 },
        tooltip: { trigger: 'axis' },
        xAxis: { type: 'category', data: names, axisLabel: { interval: 0, rotate: 30 } },
        yAxis: { type: 'value' },
        series: [{
          type: 'bar',
          data: values,
          itemStyle: { color: '#E6A23C' },
          label: { show: true, position: 'top' }
        }]
      })
    }
    chart.resize()
  })
}

async function loadMetrics() {
  try {
    const params = getRangeParams()
    const res = await axios.get('/api/stats/dashboard/summary', { params })
    renderMetrics(res.data || {})
  } catch (e) {
    renderMetrics({ days: [], sale_total: [], ad_deduction: [], profit: [] })
  }
}

async function loadProducts() {
  try {
    const params = getRangeParams()
    const res = await axios.get('/api/stats/top-products', { params })
    renderProducts(res.data || [])
  } catch (e) {
    renderProducts([])
  }
}

async function loadStoreAds() {
  try {
    const params = getRangeParams()
    const res = await axios.get('/api/store_stats/range', { params })
    renderStoreAds(res.data || { items: [] })
  } catch (e) {
    renderStoreAds({ items: [] })
  }
}

function loadAll() {
  initDefaultRange()
  loadMetrics()
  loadProducts()
  loadStoreAds()
}

watch(dateRange, () => {
  loadAll()
})

onMounted(() => {
  initDefaultRange()
  loadAll()

  window.addEventListener('resize', handleResize)
})

function handleResize() {
  metricsChart && metricsChart.resize()
  metricsOverallChart && metricsOverallChart.resize()
  productsChart && productsChart.resize()
  storeAdsChart && storeAdsChart.resize()
}

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  metricsChart && metricsChart.dispose()
  metricsOverallChart && metricsOverallChart.dispose()
  productsChart && productsChart.dispose()
  storeAdsChart && storeAdsChart.dispose()
  metricsChart = null
  metricsOverallChart = null
  productsChart = null
  storeAdsChart = null
})
</script>

<style scoped>
.stats-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.stats-header-title {
  font-size: 16px;
  font-weight: 600;
}

.stats-header-filters {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

@media (max-width: 768px) {
  .stats-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
