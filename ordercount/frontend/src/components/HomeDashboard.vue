<template>
  <div>
    <el-card shadow="hover" style="margin-bottom:20px;">
      <template #header>
        <div style="display:flex;align-items:center;justify-content:space-between;flex-wrap:wrap;gap:8px;">
          <span>首页概览</span>
          <span v-if="monthLabel" style="font-size:12px;color:#909399;">当前统计月份：{{ monthLabel }}</span>
        </div>
      </template>
      <el-row :gutter="16">
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">当月销售额（人民币）</div>
            <div style="font-size:22px;font-weight:bold;color:#409EFF;margin-top:4px;">￥{{ formatNumber(metrics.sale_total) }}</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">当月货款成本（人民币）</div>
            <div style="font-size:22px;font-weight:bold;color:#E6A23C;margin-top:4px;">￥{{ formatNumber(metrics.goods_cost) }}</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">当月广告费折算（人民币）</div>
            <div style="font-size:22px;font-weight:bold;color:#F56C6C;margin-top:4px;">￥{{ formatNumber(metrics.ad_cost) }}</div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :md="6">
          <el-card shadow="never" style="margin-bottom:10px;">
            <div style="font-size:12px;color:#909399;">当月利润（人民币）</div>
            <div :style="profitStyle">￥{{ formatNumber(metrics.profit) }}</div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <el-card shadow="hover" style="margin-bottom:20px;">
      <template #header>本月每天利润（人民币）</template>
      <div ref="profitTrendRef" style="height:260px;width:100%;"></div>
    </el-card>

    <el-row :gutter="20">
      <el-col :xs="24" :md="14">
        <el-card shadow="hover" style="margin-bottom:20px;">
          <template #header>商品销售情况（当月）</template>
          <el-table :data="productSales" size="small" border height="320">
            <el-table-column prop="product_name" label="商品名称" min-width="200" />
            <el-table-column prop="quantity" label="已销售数量" width="140" />
          </el-table>
          <div v-if="!loading && productSales.length === 0" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
            当月暂无商品销售数据
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="hover" style="margin-bottom:20px;">
          <template #header>店铺数量概览</template>
          <div style="margin-bottom:10px;font-size:14px;">
            当前已录入店铺总数：
            <span style="font-weight:bold;">{{ storeStats.total }}</span>
          </div>
          <el-table :data="storeStats.by_country" size="small" border>
            <el-table-column prop="country" label="国家" />
            <el-table-column prop="count" label="店铺数量" width="100" />
          </el-table>
          <div v-if="!loading && (!storeStats.by_country || storeStats.by_country.length === 0)" style="margin-top:10px;font-size:12px;color:#909399;text-align:center;">
            暂无店铺数据
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const loading = ref(false)
const monthLabel = ref('')
const metrics = ref({ sale_total: 0, goods_cost: 0, ad_cost: 0, profit: 0 })
const productSales = ref([])
const storeStats = ref({ total: 0, by_country: [] })

const profitTrendRef = ref(null)
let profitTrendChart

const profitStyle = computed(() => ({
  fontSize: '22px',
  fontWeight: 'bold',
  marginTop: '4px',
  color: Number(metrics.value.profit || 0) >= 0 ? '#67C23A' : '#F56C6C',
}))

function formatNumber(val) {
  const n = Number(val || 0)
  return n.toFixed(2)
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
        title: { text: '本月每天利润（暂无数据）', left: 'center', top: 10 },
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
      title: { text: '本月每天利润', left: 'center', top: 10 },
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

async function load() {
  loading.value = true
  try {
    const res = await axios.get('/api/dashboard/home')
    const data = res.data || {}
    const m = data.month || {}
    if (m.start && m.end && m.start.slice(0, 7) === m.end.slice(0, 7)) {
      monthLabel.value = m.start.slice(0, 7)
    } else if (m.start || m.end) {
      monthLabel.value = `${m.start || ''} ~ ${m.end || ''}`
    } else {
      monthLabel.value = ''
    }

    const mm = data.monthly_metrics || {}
    metrics.value = {
      sale_total: Number(mm.sale_total || 0),
      goods_cost: Number(mm.goods_cost || 0),
      ad_cost: Number(mm.ad_cost || 0),
      profit: Number(mm.profit || 0),
    }

    productSales.value = Array.isArray(data.product_sales) ? data.product_sales : []
    const ss = data.store_stats || {}
    storeStats.value = {
      total: Number(ss.total || 0),
      by_country: Array.isArray(ss.by_country) ? ss.by_country : [],
    }

    // 使用后端返回的月份起止日期加载本月每天利润折线图
    await loadProfitTrend(m.start, m.end)
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
