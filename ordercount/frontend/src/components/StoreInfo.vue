<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      现有店铺信息
    </template>

    <div style="margin-bottom:10px;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <span>统计日期：</span>
      <el-date-picker
        v-model="date"
        type="date"
        placeholder="选择日期"
        format="YYYY-MM-DD"
        value-format="YYYY-MM-DD"
        @change="onDateChange"
      />
      <el-button size="small" @click="loadAll">刷新</el-button>
      <span v-if="loading" style="font-size:12px;color:#909399;">加载中...</span>
    </div>

    <el-table :data="rows" size="small" border style="width:100%;">
      <el-table-column prop="country" label="国家" width="100" />
      <el-table-column prop="platform" label="平台" width="120" />
      <el-table-column prop="name" label="店铺名称" width="200" />
      <el-table-column label="每天广告费用(本国货币)">
        <template #default="scope">
          <template v-if="isSuperAdmin">
            <div style="display:flex;align-items:center;">
              <el-input-number
                v-model="scope.row.ad_cost"
                :min="0"
                :step="0.01"
                controls-position="right"
                style="width:140px;"
              />
              <el-button
                type="primary"
                link
                size="small"
                style="margin-left:8px;"
                @click="onSaveStat(scope.row)"
              >保存</el-button>
            </div>
            <div style="margin-top:2px;font-size:11px;color:#909399;min-height:16px;">
              {{ scaleHint(scope.row.ad_cost) }}
            </div>
          </template>
          <template v-else>
            <span>{{ Number(scope.row.ad_cost || 0).toFixed(2) }}</span>
          </template>
        </template>
      </el-table-column>
      <el-table-column label="折算成人民币" width="150">
        <template #default="scope">
          <span>￥{{ formatAdCostCny(scope.row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="140">
        <template #default="scope">
          <el-button type="primary" link size="small" @click="onShowDetail(scope.row)">
            查看详情
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <!-- 店铺基本信息详情弹窗，仅展示，不在此编辑 -->
  <el-dialog v-model="detailVisible" title="店铺基本信息" width="500px">
    <div v-if="detailStore">
      <el-descriptions :column="1" size="small" border>
        <el-descriptions-item label="国家">{{ detailStore.country }}</el-descriptions-item>
        <el-descriptions-item label="平台">{{ detailStore.platform }}</el-descriptions-item>
        <el-descriptions-item label="店铺名称">{{ detailStore.name }}</el-descriptions-item>
        <el-descriptions-item label="登录账号">{{ detailStore.login_account }}</el-descriptions-item>
        <el-descriptions-item label="登录密码">{{ detailStore.login_password }}</el-descriptions-item>
        <el-descriptions-item label="绑定手机号">{{ detailStore.phone }}</el-descriptions-item>
        <el-descriptions-item label="绑定邮箱">{{ detailStore.email }}</el-descriptions-item>
      </el-descriptions>
    </div>
    <template #footer>
      <el-button @click="detailVisible = false">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})

const date = ref('')
const stores = ref([])
const statsMap = ref({}) // { [storeId]: { ad_cost, sale_total } }
const rows = ref([])
const loading = ref(false)

// 汇率：后端返回含义为 1 CNY ≈ rates[币种] 外币
const rates = ref({ PHP: 0, IDR: 0, MYR: 0, USD: 0, CNY: 1 })

// 店铺国家与币种映射（与结算工具保持一致）
const countryCurrencyMap = {
  '菲律宾': 'PHP',
  '印尼': 'IDR',
  '马来西亚': 'MYR',
}

const detailVisible = ref(false)
const detailStore = ref(null)

const isSuperAdmin = computed(() => props.currentUser && props.currentUser.role === 'superadmin')

function rebuildRows () {
  const map = statsMap.value || {}
  rows.value = (stores.value || [])
    .map(s => ({
      // 兼容后端返回的 is_blocked 可能为 boolean、0/1 或字符串 '0'/'1'/'false'/'true'
      ...s,
      is_blocked: (s.is_blocked === true) || (s.is_blocked === 1) || (s.is_blocked === '1') || (s.is_blocked === 'true')
    }))
    .filter(s => !s.is_blocked)
    .map(s => {
      const stat = map[s.id] || {}
      return {
        ...s,
        ad_cost: typeof stat.ad_cost === 'number' ? stat.ad_cost : 0,
        sale_total: typeof stat.sale_total === 'number' ? stat.sale_total : 0,
      }
    })
}

async function loadStores () {
  try {
    const devBypass = (typeof window !== 'undefined') && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')
    const opts = devBypass ? { headers: { 'X-Bypass-Admin': '1' } } : {}
    const res = await axios.get('/api/shops', opts)
    stores.value = res.data?.items || []
  } catch (e) {
    stores.value = []
  }
}

async function loadRates () {
  try {
    const res = await axios.get('/api/exchange/rates')
    rates.value = {
      PHP: res.data?.rates?.PHP || 0,
      IDR: res.data?.rates?.IDR || 0,
      MYR: res.data?.rates?.MYR || 0,
      USD: res.data?.rates?.USD || 0,
      CNY: res.data?.rates?.CNY || 1,
    }
  } catch (e) {
    // 失败时保留已有数值，不影响页面使用
  }
}

async function loadStats () {
  if (!date.value) {
    const today = new Date()
    const y = today.getFullYear()
    const m = String(today.getMonth() + 1).padStart(2, '0')
    const d = String(today.getDate()).padStart(2, '0')
    date.value = `${y}-${m}-${d}`
  }
  try {
    const devBypass = (typeof window !== 'undefined') && (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')
    const opts = devBypass ? { headers: { 'X-Bypass-Admin': '1' }, params: { date: date.value } } : { params: { date: date.value } }
    const res = await axios.get('/api/store_stats', opts)
    const list = res.data?.items || []
    const map = {}
    for (const it of list) {
      if (!it || !it.store_id) continue
      map[it.store_id] = { ad_cost: it.ad_cost, sale_total: it.sale_total }
    }
    statsMap.value = map
  } catch (e) {
    statsMap.value = {}
  }
}

async function loadAll () {
  loading.value = true
  try {
    await Promise.all([
      loadStores(),
      loadStats(),
      loadRates(),
    ])
    rebuildRows()
  } finally {
    loading.value = false
  }
}

function onDateChange () {
  loadAll()
}

async function onSaveStat (row) {
  if (!row || !row.id) return
  try {
    await axios.post('/api/store_stats', {
      store_id: row.id,
      date: date.value,
      ad_cost: Number(row.ad_cost) || 0,
      // sale_total 字段预留，这里先固定为 0
      sale_total: Number(row.sale_total) || 0,
    })
    ElMessage.success('已保存当日广告费用')
    await loadStats()
    rebuildRows()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || e?.message || '保存失败')
  }
}

function onShowDetail (row) {
  if (!row) return
  detailStore.value = { ...row }
  detailVisible.value = true
}

// 计算某一行广告费折算成人民币
function calcAdCostCny (row) {
  if (!row) return 0
  const amount = Number(row.ad_cost) || 0
  if (!amount) return 0
  const cur = countryCurrencyMap[row.country]
  if (!cur) {
    // 未映射的国家，默认视为人民币
    return amount
  }
  const r = Number(rates.value[cur]) || 0
  if (!r) return 0
  // 后端返回：1 CNY ≈ r 外币，这里需要 1 外币 ≈ 1/r CNY
  return amount / r
}

function formatAdCostCny (row) {
  const v = calcAdCostCny(row)
  return v.toFixed(2)
}

function scaleHint (val) {
  const v = Number(val) || 0
  if (!v) return ''
  if (v >= 100000000000) return '千亿'
  if (v >= 10000000000) return '百亿'
  if (v >= 1000000000) return '十亿'
  if (v >= 100000000) return '亿'
  if (v >= 10000000) return '千万'
  if (v >= 1000000) return '百万'
  if (v >= 100000) return '十万'
  if (v >= 10000) return '万'
  if (v >= 1000) return '千'
  return ''
}

onMounted(() => {
  loadAll()
})
</script>

<style scoped>
</style>
