<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      订单记录
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
    <el-table :data="items" size="small" border style="width:100%;">
      <el-table-column prop="created_at" label="时间" width="170">
        <template #default="scope">
          {{ formatTime(scope.row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column prop="country" label="国家" width="80" />
      <el-table-column prop="platform" label="平台" width="100" />
      <el-table-column prop="order_no" label="订单号" width="160" />
      <el-table-column prop="product_name" label="商品名" min-width="150" />
      <el-table-column prop="sku" label="SKU" width="120" />
      <el-table-column prop="quantity" label="数量" width="70" />
      <el-table-column prop="currency" label="币种" width="80" />
      <el-table-column prop="total_amount" label="总额" />
    </el-table>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const date = ref('')
const country = ref('')
const items = ref([])
const loading = ref(false)

function formatTime(t) {
  if (!t) return ''
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
    if (country.value) params.country = country.value
    const res = await axios.get('/api/orders', { params })
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
