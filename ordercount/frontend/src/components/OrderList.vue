<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      订单记录
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
        @change="load"
      />
        <span>国家：</span>
        <el-select v-model="country" placeholder="全部国家" style="width:140px;" @change="load" clearable>
          <el-option label="全部" :value="''" />
          <el-option v-for="c in countries" :key="c" :label="c" :value="c" />
        </el-select>
      <el-button size="small" @click="load">刷新</el-button>
      <span v-if="loading" style="font-size:12px;color:#909399;">加载中...</span>
    </div>
    <el-table :data="items" size="small" border style="width:100%;margin-bottom:10px;">
      <el-table-column prop="created_at" label="时间" width="180">
        <template #default="scope">
          <template v-if="editingId === scope.row.id">
            <div style="display:flex;align-items:center;gap:6px;">
              <el-date-picker
                v-model="editDate"
                type="date"
                placeholder="选择日期"
                format="YYYY-MM-DD"
                value-format="YYYY-MM-DD"
                size="small"
              />
              <el-button type="primary" link size="small" @click="onSaveDate(scope.row)">
                保存
              </el-button>
              <el-button link size="small" @click="onCancelEdit">取消</el-button>
            </div>
          </template>
          <template v-else>
            <div style="display:flex;align-items:center;gap:6px;">
              <span>{{ formatDateOnly(scope.row.created_at) }}</span>
            </div>
          </template>
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
    <div v-if="!loading" style="display:flex;justify-content:space-between;align-items:center;">
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
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { fetchCountries } from '../utils/countries'

const dateRange = ref([])
const country = ref('')
const countries = ref([])
const items = ref([])
const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const editingId = ref(null)
const editDate = ref('')

function formatDateOnly(t) {
  if (!t) return ''
  // 后端返回 "YYYY-MM-DD" 或 "YYYY-MM-DD HH:MM:SS"，这里统一只取日期部分
  const s = String(t)
  const idx = s.indexOf(' ')
  return idx >= 0 ? s.slice(0, idx) : s
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
    if (country.value) params.country = country.value
    const res = await axios.get('/api/orders', { params })
    items.value = res.data?.items || []
    total.value = typeof res.data?.total === 'number' ? res.data.total : items.value.length
  } catch (e) {
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function onPageChange(page) {
  currentPage.value = page
  load()
}

function onEditDate(row) {
  if (!row || !row.id) return
  editingId.value = row.id
  // 预填当前日期部分，方便修改
  const t = row.created_at
  if (t) {
    const d = new Date(t)
    if (!Number.isNaN(d.getTime())) {
      const yyyy = d.getFullYear()
      const mm = String(d.getMonth() + 1).padStart(2, '0')
      const dd = String(d.getDate()).padStart(2, '0')
      editDate.value = `${yyyy}-${mm}-${dd}`
    }
  }
}

function onCancelEdit() {
  editingId.value = null
  editDate.value = ''
}

async function onSaveDate(row) {
  if (!row || !row.id || !editDate.value) return
  try {
    await axios.put(`/api/orders/${row.id}/date`, { date: editDate.value })
    editingId.value = null
    editDate.value = ''
    await load()
  } catch (e) {
    // 简单忽略错误，必要时可加提示
  }
}

onMounted(() => {
  const today = new Date()
  const yyyy = today.getFullYear()
  const mm = String(today.getMonth() + 1).padStart(2, '0')
  const dd = String(today.getDate()).padStart(2, '0')
  const todayStr = `${yyyy}-${mm}-${dd}`
  dateRange.value = [todayStr, todayStr]
  load()
  ;(async () => {
    const list = await fetchCountries()
    countries.value = list || []
  })()
})
</script>
