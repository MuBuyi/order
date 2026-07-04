<template>
  <div>
    <el-card shadow="hover" style="margin-bottom:20px;">
      <template #header>
        <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:8px;">
          <span>首页概览</span>
          <div style="display:flex;align-items:center;gap:8px;flex-wrap:wrap;justify-content:flex-end;">
            <el-date-picker
              v-model="selectedMonth"
              type="month"
              placeholder="选择月份"
              format="YYYY-MM"
              value-format="YYYY-MM"
              size="small"
              style="width:160px;"
              :clearable="false"
              :disabled-date="disableFutureMonth"
              @change="onMonthChange"
            />
            <el-button size="small" @click="resetToCurrentMonth">本月</el-button>
            <span v-if="monthLabel" style="font-size:12px;color:#909399;">当前统计月份：{{ monthLabel }}</span>
          </div>
        </div>
      </template>
      <div class="overview-cards">
        <div class="overview-card-item">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">{{ selectedMonthText }}销售额（人民币）</div>
            <div style="font-size:22px;font-weight:bold;color:#409EFF;margin-top:4px;">￥{{ formatNumber(metrics.sale_total) }}</div>
          </el-card>
        </div>
        <div class="overview-card-item">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">{{ selectedMonthText }}货款成本（人民币）</div>
            <div style="font-size:22px;font-weight:bold;color:#E6A23C;margin-top:4px;">￥{{ formatNumber(metrics.goods_cost) }}</div>
          </el-card>
        </div>
        <div class="overview-card-item">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">{{ selectedMonthText }}广告费折算（人民币）</div>
            <div style="font-size:22px;font-weight:bold;color:#F56C6C;margin-top:4px;">￥{{ formatNumber(metrics.ad_cost) }}</div>
          </el-card>
        </div>
        <div class="overview-card-item">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">{{ selectedMonthText }}退货利润亏损总和（CNY）</div>
            <div style="font-size:22px;font-weight:bold;color:#909399;margin-top:4px;">￥{{ formatNumber(metrics.return_loss) }}</div>
          </el-card>
        </div>
        <div class="overview-card-item">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">{{ selectedMonthText }}利润（人民币）</div>
            <div :style="profitStyle">￥{{ formatNumber(metrics.profit) }}</div>
          </el-card>
        </div>
        <div class="overview-card-item">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">{{ selectedMonthText }}日均利润（人民币）</div>
            <div :style="avgProfitStyle">￥{{ formatNumber(avgDailyProfit) }}</div>
          </el-card>
        </div>
      </div>
    </el-card>

    <el-card shadow="hover" style="margin-bottom:20px;">
      <template #header>{{ selectedMonthText }}每天利润（人民币）</template>
      <div ref="profitTrendRef" style="height:260px;width:100%;"></div>
    </el-card>

    <el-row :gutter="20" class="product-panels-row">
      <el-col :xs="24" :md="14">
        <el-card shadow="hover" class="panel-card" style="margin-bottom:20px;">
          <template #header>商品销售情况（{{ selectedMonthText }}）</template>
          <el-table :data="productSales" size="small" border height="320">
            <el-table-column prop="product_name" label="商品名称" min-width="200" />
            <el-table-column prop="quantity" label="已销售数量" width="140" />
          </el-table>
          <div v-if="!loading && productSales.length === 0" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
            {{ selectedMonthText }}暂无商品销售数据
          </div>
        </el-card>

        <el-card shadow="hover" class="panel-card" style="margin-bottom:20px;">
          <template #header>商品退货率（{{ selectedMonthText }}）</template>
          <el-table :data="productReturnRates" size="small" border height="220">
            <el-table-column prop="product_name" label="商品名称" min-width="140" />
            <el-table-column prop="sold_quantity" label="销售数量" width="88" />
            <el-table-column prop="return_quantity" label="退货数量" width="88" />
            <el-table-column prop="return_rate" label="退货率" width="90" />
          </el-table>
          <div v-if="!loading && productReturnRates.length === 0" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
            {{ selectedMonthText }}暂无可计算退货率数据
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="hover" class="panel-card" style="margin-bottom:20px;">
          <template #header>商品退货情况（{{ selectedMonthText }}）</template>
          <el-table :data="productReturns" size="small" border height="320">
            <el-table-column prop="product_name" label="商品名称" min-width="180" />
            <el-table-column prop="quantity" label="已退货数量" width="120" />
          </el-table>
          <div v-if="!loading && productReturns.length === 0" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
            {{ selectedMonthText }}暂无商品退货数据
          </div>
        </el-card>

        <el-card shadow="hover" class="panel-card" style="margin-bottom:20px;">
          <template #header>
            <div style="display:flex;align-items:center;justify-content:space-between;gap:8px;">
              <span>店铺数量概览</span>
              <span style="font-size:12px;color:#606266;">总数：<strong>{{ storeStats.total }}</strong></span>
            </div>
          </template>
          <el-table :data="storeStats.by_country" size="small" border height="220">
            <el-table-column prop="country" label="国家" />
            <el-table-column prop="count" label="店铺数量" width="100" />
          </el-table>
          <div v-if="!loading && (!storeStats.by_country || storeStats.by_country.length === 0)" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
            暂无店铺数据
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="hover" style="margin-bottom:20px;">
      <template #header>历史月度数据（销售额 / 广告 / 利润）</template>
      <div ref="monthlyMetricsRef" style="height:300px;width:100%;"></div>
    </el-card>

    <el-card shadow="hover" style="margin-bottom:20px;">
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center;gap:8px;flex-wrap:wrap;">
          <span>{{ selectedMonthText }}详情</span>
          <span style="font-size:12px;color:#909399;">详情数据与顶部月份选择保持同步</span>
        </div>
      </template>
      <el-table :data="detailRows" size="small" border v-loading="detailLoading" style="width:100%;">
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column prop="sale_total" label="销售额" min-width="120">
          <template #default="scope">￥{{ formatNumber(scope.row.sale_total) }}</template>
        </el-table-column>
        <el-table-column prop="goods_cost" label="货款成本" min-width="120">
          <template #default="scope">￥{{ formatNumber(scope.row.goods_cost) }}</template>
        </el-table-column>
        <el-table-column prop="ad_deduction" label="广告费折算" min-width="120">
          <template #default="scope">￥{{ formatNumber(scope.row.ad_deduction) }}</template>
        </el-table-column>
        <el-table-column prop="platform_fee" label="平台费" min-width="120">
          <template #default="scope">￥{{ formatNumber(scope.row.platform_fee) }}</template>
        </el-table-column>
        <el-table-column prop="profit" label="利润" min-width="120">
          <template #default="scope">￥{{ formatNumber(scope.row.profit) }}</template>
        </el-table-column>
        <el-table-column prop="avg_daily_profit" label="当月日平均利润" min-width="140">
          <template #default="scope">￥{{ formatNumber(scope.row.avg_daily_profit) }}</template>
        </el-table-column>
      </el-table>
      <div v-if="!detailLoading && detailRows.length === 0" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
        暂无数据
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'
import { ElMessage } from 'element-plus'

function getCurrentMonth() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  return `${year}-${month}`
}

const loading = ref(false)
const monthLabel = ref('')
const selectedMonth = ref(getCurrentMonth())
const metrics = ref({ sale_total: 0, goods_cost: 0, ad_cost: 0, return_loss: 0, profit: 0 })
const productSales = ref([])
const productReturns = ref([])
const storeStats = ref({ total: 0, by_country: [] })
const selectedMonthText = computed(() => monthLabel.value || selectedMonth.value || '当月')

const productReturnRates = computed(() => {
  const soldMap = new Map()
  for (const item of productSales.value || []) {
    const name = String(item?.product_name || '').trim()
    if (!name) continue
    soldMap.set(name, Number(item?.quantity || 0))
  }

  const returnMap = new Map()
  for (const item of productReturns.value || []) {
    const name = String(item?.product_name || '').trim()
    if (!name) continue
    returnMap.set(name, Number(item?.quantity || 0))
  }

  const rows = []
  for (const [name, soldQtyRaw] of soldMap.entries()) {
    const soldQty = Number(soldQtyRaw || 0)
    const returnQty = Number(returnMap.get(name) || 0)
    const rate = soldQty > 0 ? (returnQty / soldQty) * 100 : 0

    rows.push({
      product_name: name,
      sold_quantity: soldQty,
      return_quantity: returnQty,
      return_rate: `${rate.toFixed(2)}%`,
      _rate_value: rate,
    })
  }

  return rows.sort((a, b) => b._rate_value - a._rate_value)
})

const profitTrendRef = ref(null)
let profitTrendChart

const monthlyMetricsRef = ref(null)
let monthlyMetricsChart

const detailRows = ref([])
const detailLoading = ref(false)
const detailDataDays = ref(0)

const profitStyle = computed(() => ({
  fontSize: '22px',
  fontWeight: 'bold',
  marginTop: '4px',
  color: Number(metrics.value.profit || 0) >= 0 ? '#67C23A' : '#F56C6C',
}))

const avgDailyProfit = computed(() => {
  const dayCount = Number(detailDataDays.value || 0)
  if (dayCount <= 0) return 0
  return Number(metrics.value.profit || 0) / dayCount
})

const avgProfitStyle = computed(() => ({
  fontSize: '22px',
  fontWeight: 'bold',
  marginTop: '4px',
  color: Number(avgDailyProfit.value || 0) >= 0 ? '#67C23A' : '#F56C6C',
}))

function formatNumber(val) {
  const n = Number(val || 0)
  return n.toFixed(2)
}

function disableFutureMonth(date) {
  const currentMonth = new Date()
  currentMonth.setDate(1)
  currentMonth.setHours(0, 0, 0, 0)

  const candidate = new Date(date)
  candidate.setDate(1)
  candidate.setHours(0, 0, 0, 0)
  return candidate.getTime() > currentMonth.getTime()
}

function renderProfitTrend(days, profits) {
  nextTick(() => {
    if (!profitTrendChart && profitTrendRef.value) {
      profitTrendChart = echarts.init(profitTrendRef.value)
    }
    if (!profitTrendChart) return

    const hasData = Array.isArray(days) && days.length > 0 && Array.isArray(profits) && profits.length > 0
    if (!hasData) {
      profitTrendChart.clear()
      profitTrendChart.setOption({
        title: { text: `${selectedMonthText.value}每天利润（暂无数据）`, left: 'center', top: 10 },
        xAxis: { type: 'category', data: [] },
        yAxis: { type: 'value' },
        series: [],
        graphic: {
          type: 'text',
          left: 'center',
          top: 'middle',
          style: { text: '暂无数据', fontSize: 18, fill: '#888' },
        },
      })
      profitTrendChart.resize()
      return
    }

    const xData = days.map(d => (d || '').slice(5))
    const yData = profits.map(v => Number(v || 0))

    profitTrendChart.clear()
    profitTrendChart.setOption({
      title: { text: `${selectedMonthText.value}每天利润`, left: 'center', top: 10 },
      tooltip: {
        trigger: 'axis',
        valueFormatter: (v) => Math.round(Number(v || 0)).toString(),
      },
      grid: { left: 40, right: 20, top: 60, bottom: 40 },
      xAxis: { type: 'category', data: xData },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (val) => Math.round(Number(val || 0)),
        },
      },
      series: [{
        type: 'line',
        smooth: true,
        data: yData,
        areaStyle: {},
        label: {
          show: true,
          position: 'top',
          formatter: (params) => Math.round(Number(params.value || 0)),
        },
        itemStyle: { color: '#67C23A' },
      }],
    })
    profitTrendChart.resize()
  })
}

async function loadProfitTrend(start, end) {
  try {
    const params = {}
    if (start) params.start_date = start
    if (end) params.end_date = end
    const res = await axios.get('/api/stats/dashboard/summary', { params })
    const data = res.data || {}
    const days = Array.isArray(data.days) ? data.days : []
    const profits = Array.isArray(data.profit) ? data.profit : []
    renderProfitTrend(days, profits)
  } catch (e) {
    // 出错时清空并提示
    renderProfitTrend([], [])
  }
}

function renderMonthlyMetrics(saleArr, adArr, profitArr) {
  nextTick(() => {
    if (!monthlyMetricsChart && monthlyMetricsRef.value) {
      monthlyMetricsChart = echarts.init(monthlyMetricsRef.value)
    }
    if (!monthlyMetricsChart) return

    const months = ['1月', '2月', '3月', '4月', '5月', '6月', '7月', '8月', '9月', '10月', '11月', '12月']

    const hasData = (arr) => Array.isArray(arr) && arr.some(v => Number(v || 0) !== 0)
    if (!hasData(saleArr) && !hasData(adArr) && !hasData(profitArr)) {
      monthlyMetricsChart.clear()
      monthlyMetricsChart.setOption({
        title: { text: '历史月度数据（暂无数据）', left: 'center', top: 10 },
        xAxis: { type: 'category', data: months },
        yAxis: { type: 'value' },
        series: [],
        graphic: {
          type: 'text',
          left: 'center',
          top: 'middle',
          style: { text: '暂无数据', fontSize: 18, fill: '#888' },
        },
      })
      monthlyMetricsChart.resize()
      return
    }

    monthlyMetricsChart.clear()
    monthlyMetricsChart.setOption({
      title: { text: '历史月度数据（销售额 / 广告 / 利润）', left: 'center', top: 10 },
      tooltip: { trigger: 'axis' },
      legend: { top: 40 },
      grid: { top: 80, left: '8%', right: '4%', bottom: 40, containLabel: true },
      xAxis: { type: 'category', data: months },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (val) => Math.round(Number(val || 0)),
        },
      },
      series: [
        {
          name: '销售额',
          type: 'line',
          smooth: true,
          data: saleArr,
        },
        {
          name: '广告费用折算',
          type: 'line',
          smooth: true,
          data: adArr,
        },
        {
          name: '利润',
          type: 'line',
          smooth: true,
          data: profitArr,
        },
      ],
    })
    monthlyMetricsChart.resize()
  })
}

async function loadMonthlyMetrics() {
  try {
    const res = await axios.get('/api/stats/monthly-metrics')
    const data = res.data || {}
    const saleArr = Array.isArray(data.sale_total) ? data.sale_total : []
    const adArr = Array.isArray(data.ad_deduction) ? data.ad_deduction : []
    const profitArr = Array.isArray(data.profit) ? data.profit : []
    renderMonthlyMetrics(saleArr, adArr, profitArr)
  } catch (e) {
    renderMonthlyMetrics([], [], [])
  }
}

async function loadMonthlyDetail() {
  detailLoading.value = true
  try {
    const params = {}
    if (selectedMonth.value) {
      const parts = String(selectedMonth.value).split('-')
      if (parts.length === 2) {
        params.year = parts[0]
        params.month = parts[1]
      }
    }
    const res = await axios.get('/api/stats/monthly-detail', { params })
    const data = res.data || {}
    const dayCount = Number(data.data_days || 0)
    detailDataDays.value = dayCount
    const items = Array.isArray(data.items) ? data.items : []
    detailRows.value = items.map(item => {
      const profit = Number(item?.profit || 0)
      return {
        ...item,
        avg_daily_profit: dayCount > 0 ? profit / dayCount : 0,
      }
    })
  } catch (e) {
    detailDataDays.value = 0
    detailRows.value = []
  } finally {
    detailLoading.value = false
  }
}

function syncMonthLabel(monthValue) {
  monthLabel.value = monthValue || ''
}

async function onMonthChange() {
  if (!selectedMonth.value) {
    selectedMonth.value = getCurrentMonth()
  }
  await load()
}

async function resetToCurrentMonth() {
  const currentMonth = getCurrentMonth()
  if (selectedMonth.value === currentMonth) {
    await load()
    return
  }
  selectedMonth.value = currentMonth
  await onMonthChange()
}

async function load() {
  loading.value = true
  try {
    const params = {}
    if (selectedMonth.value) {
      const parts = String(selectedMonth.value).split('-')
      if (parts.length === 2) {
        params.year = parts[0]
        params.month = parts[1]
      }
    }

    const res = await axios.get('/api/dashboard/home', { params })
    const data = res.data || {}
    const m = data.month || {}
    if (m.start && m.end && m.start.slice(0, 7) === m.end.slice(0, 7)) {
      syncMonthLabel(m.start.slice(0, 7))
      selectedMonth.value = m.start.slice(0, 7)
    } else if (m.start || m.end) {
      syncMonthLabel(`${m.start || ''} ~ ${m.end || ''}`)
    } else {
      syncMonthLabel(selectedMonth.value)
    }

    const mm = data.monthly_metrics || {}
    metrics.value = {
      sale_total: Number(mm.sale_total || 0),
      goods_cost: Number(mm.goods_cost || 0),
      ad_cost: Number(mm.ad_cost || 0),
      return_loss: Number(mm.return_loss || 0),
      profit: Number(mm.profit || 0),
    }

    productSales.value = Array.isArray(data.product_sales) ? data.product_sales : []
    productReturns.value = Array.isArray(data.product_returns) ? data.product_returns : []
    const ss = data.store_stats || {}
    storeStats.value = {
      total: Number(ss.total || 0),
      by_country: Array.isArray(ss.by_country) ? ss.by_country : [],
    }

    // 使用后端返回的月份起止日期加载本月每天利润折线图
    await loadProfitTrend(m.start, m.end)

    // 加载历史月度汇总数据
    await loadMonthlyMetrics()

    // 详情始终跟随当前选择月份
    await loadMonthlyDetail()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || e?.message || '加载首页数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(load)

// 暴露给父组件在需要时手动刷新
defineExpose({ load })
</script>

<style scoped>
.overview-cards {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 16px;
}

.overview-card-item :deep(.el-card) {
  margin-bottom: 0 !important;
}

.product-panels-row :deep(.el-col) {
  display: flex;
  flex-direction: column;
}

.panel-card {
  width: 100%;
}

@media (max-width: 1199px) {
  .overview-cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 767px) {
  .overview-cards {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
