<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      每日结算记录
    </template>
    <div style="margin-bottom:10px;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <span>选择日期：</span>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        format="YYYY-MM-DD"
        value-format="YYYY-MM-DD"
        @change="onFilterChange"
      />
      <span>国家：</span>
      <el-select v-model="country" placeholder="全部国家" style="width:140px;" @change="onFilterChange" clearable>
        <el-option label="全部" :value="''" />
        <el-option v-for="c in countries" :key="c" :label="c" :value="c" />
      </el-select>
      <el-button size="small" @click="onRefresh">刷新</el-button>
      <el-button
        v-if="isSuperAdmin"
        type="primary"
        size="small"
        @click="onManualPush"
        :loading="pushing"
      >
        主动推送
      </el-button>
      <span v-if="loading" style="font-size:12px;color:#909399;">加载中...</span>
    </div>
    <el-table :data="items" size="small" border style="width:100%;margin-bottom:10px;">
      <el-table-column prop="created_at" label="时间" width="160">
        <template #default="scope">
          {{ formatTime(scope.row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column prop="country" label="国家" width="80" />
      <el-table-column prop="currency" label="币种" width="80" />
      <el-table-column prop="sale_total" label="销售额" />
      <el-table-column prop="ad_cost" label="广告费(原币)" />
      <el-table-column prop="exchange" label="汇率(1外币≈?CNY)" />
      <el-table-column prop="ad_deduction" label="广告成本(￥)">
        <template #default="scope">
          ￥{{ Number(scope.row.ad_deduction).toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column prop="platform_fee" label="平台手续费(￥)">
        <template #default="scope">
          ￥{{ Number(scope.row.platform_fee).toFixed(2) }}
        </template>
      </el-table-column>
      <el-table-column prop="goods_cost" label="货款成本" />
      <el-table-column prop="shua_dan_fee" label="刷单费用" />
      <el-table-column prop="fixed_cost" label="固定成本" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column prop="profit" label="利润(￥)">
        <template #default="scope">
          <span :style="{color: scope.row.profit >= 0 ? '#67C23A' : '#F56C6C', 'font-weight':'bold'}">
            ￥{{ Number(scope.row.profit).toFixed(2) }}
          </span>
        </template>
      </el-table-column>
    </el-table>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:4px;" v-if="!loading">
      <div style="font-size:12px;color:#909399;">
        共 {{ total }} 条记录，每页 {{ pageSize }} 条
      </div>
      <el-pagination
        background
        layout="prev, pager, next"
        :page-size="pageSize"
        :current-page="currentPage"
        :total="total"
        @current-change="onPageChange"
      />
    </div>
    <div v-if="!loading" style="text-align:right;font-weight:bold;">
      当日利润汇总：
      <span :style="{color: totalProfit >= 0 ? '#67C23A' : '#F56C6C'}">
        ￥{{ totalProfit.toFixed(2) }}
      </span>
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { fetchCountries } from '../utils/countries'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})
const dateRange = ref([])
const country = ref('')
const countries = ref([])
const items = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)
const pushing = ref(false)

const totalProfit = computed(() => {
  return items.value.reduce((sum, it) => sum + (Number(it.profit) || 0), 0)
})

const isSuperAdmin = computed(() => props.currentUser && props.currentUser.role === 'superadmin')

function formatTime(t) {
  if (!t) return ''
  // 兼容字符串时间
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      page_size: pageSize.value,
    }
    if (Array.isArray(dateRange.value) && dateRange.value.length === 2) {
      const [start, end] = dateRange.value
      if (start && end) {
        params.start_date = start
        params.end_date = end
      }
    }
    if (country.value) {
      params.country = country.value
    }
    const res = await axios.get('/api/settlements', { params })
    items.value = res.data?.items || []
    // 后端返回 total 时使用后端的；否则退回前端计算
    total.value = typeof res.data?.total === 'number' ? res.data.total : items.value.length
  } catch (e) {
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function onManualPush() {
  if (pushing.value) return
  pushing.value = true
  try {
    const payload = {}
    // 兼容：当选择区间为同一天时，仍然向后端传递单一 date 进行推送
    if (Array.isArray(dateRange.value) && dateRange.value.length === 2) {
      const [start, end] = dateRange.value
      if (start && end && start === end) {
        payload.date = start
      }
    }
    await axios.post('/api/settlements/push', payload)
    // 推送成功后不强制刷新列表，只在需要时手动刷新
  } catch (e) {
    // 这里可以根据需要加上消息提示，目前先静默失败
  } finally {
    pushing.value = false
  }
}

function onFilterChange() {
  currentPage.value = 1
  load()
}

function onRefresh() {
  currentPage.value = 1
  load()
}

function onPageChange(page) {
  currentPage.value = page
  load()
}

onMounted(() => {
  // 默认不限定日期，加载所有结算记录的第 1 页
  load()
  ;(async () => {
    const list = await fetchCountries()
    countries.value = list || []
  })()
})
</script>
