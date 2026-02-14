<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      每日结算记录
    </template>
    <div style="margin-bottom:10px;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <span>选择日期：</span>
      <el-date-picker
        v-model="date"
        type="date"
        placeholder="选择日期"
        format="YYYY-MM-DD"
        value-format="YYYY-MM-DD"
        @change="load"
      />
      <span>国家：</span>
      <el-select v-model="country" placeholder="全部国家" style="width:140px;" @change="load" clearable>
        <el-option label="全部" :value="''" />
        <el-option label="菲律宾" value="菲律宾" />
        <el-option label="印尼" value="印尼" />
        <el-option label="马来西亚" value="马来西亚" />
      </el-select>
      <el-button size="small" @click="load">刷新</el-button>
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

const date = ref('')
const country = ref('')
const items = ref([])
const loading = ref(false)

const totalProfit = computed(() => {
  return items.value.reduce((sum, it) => sum + (Number(it.profit) || 0), 0)
})

function formatTime(t) {
  if (!t) return ''
  // 兼容字符串时间
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  if (!date.value) return
  loading.value = true
  try {
    const params = { date: date.value }
    if (country.value) {
      params.country = country.value
    }
    const res = await axios.get('/api/settlements', { params })
    items.value = res.data?.items || []
  } catch (e) {
    items.value = []
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  const today = new Date()
  const yyyy = today.getFullYear()
  const mm = String(today.getMonth() + 1).padStart(2, '0')
  const dd = String(today.getDate()).padStart(2, '0')
  date.value = `${yyyy}-${mm}-${dd}`
  load()
})
</script>
