<template>
  <el-card class="return-page" shadow="hover" style="margin-bottom:20px;">
    <template #header>
      退货管理
    </template>

    <div style="margin-bottom:10px;display:flex;justify-content:flex-end;">
      <el-button type="primary" @click="openCreateDialog">添加退货</el-button>
    </div>

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑退货' : '添加退货'" :width="dialogWidth" destroy-on-close>
      <el-form class="dialog-vertical-form" :model="form" label-width="130px" size="small" @submit.prevent>
        <el-form-item label="国家">
          <el-select v-model="form.country" placeholder="请选择国家" clearable filterable style="width: 100%;" @change="onCountryChange">
            <el-option
              v-for="c in countries"
              :key="c"
              :label="c"
              :value="c"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="店铺">
          <el-select v-model="form.store_name" placeholder="请选择店铺" filterable clearable style="width: 100%;">
            <el-option
              v-for="shop in enabledShops"
              :key="shop.id"
              :label="shop.name"
              :value="shop.name"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="商品SKU">
          <el-select v-model="form.sku" placeholder="请选择商品SKU" filterable clearable style="width: 100%;" @change="onSkuChange">
            <el-option
              v-for="p in products"
              :key="p.id"
              :label="p.sku"
              :value="p.sku"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="数量">
          <el-input-number v-model="form.quantity" :min="1" style="width:100%;" />
        </el-form-item>

        <el-form-item label="退货日期" class="return-date-item">
          <el-date-picker v-model="form.return_date" type="date" format="YYYY-MM-DD" value-format="YYYY-MM-DD" style="width:100%;" />
        </el-form-item>

        <el-form-item label="订单号">
          <el-input v-model="form.order_id" maxlength="100" placeholder="请输入纯数字或数字+字母组合，如 12345 / A12b" style="width:100%;" />
        </el-form-item>

        <el-form-item label="平台退款金额">
          <div class="amount-field">
            <div class="amount-inline">
              <el-input-number v-model="form.refund_amount" :step="0.01" :min="0" :controls="false" style="width:100%;" />
              <span class="currency-tag">{{ localCurrencyCode }}</span>
            </div>
            <div class="scale-hint">{{ scaleHint(form.refund_amount) }}</div>
            <div class="rate-hint">当前汇率（CNY/{{ localCurrencyCode }}）：{{ currentRateText }}</div>
          </div>
        </el-form-item>

        <el-form-item label="利润亏损金额（CNY）">
          <div class="amount-field">
            <div class="calc-hint">{{ lossCalcDetail }}</div>
            <el-input-number v-model="form.loss_amount" :step="0.01" :controls="false" :disabled="true" style="width:100%;" />
          </div>
        </el-form-item>

        <el-form-item label="处理人">
          <el-input v-model="form.handler" disabled />
        </el-form-item>

        <el-form-item label="原因/备注">
          <el-input type="textarea" v-model="form.remark" rows="2" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button @click="onReset">重置</el-button>
        <el-button type="primary" @click="onSave">{{ form.id ? '保存修改' : '添加退货' }}</el-button>
      </template>
    </el-dialog>

    <div style="margin-bottom:10px;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <span style="font-size:13px;color:#606266;">时间筛选：</span>
      <el-date-picker
        v-model="filters.dateRange"
        type="daterange"
        value-format="YYYY-MM-DD"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        clearable
      />
      <el-button type="primary" size="small" @click="load">查询</el-button>
      <el-button size="small" @click="resetFilters">重置筛选</el-button>
    </div>

    <el-table :data="items" size="small" border style="width:100%;" v-loading="loading">
      <el-table-column type="index" label="#" width="50" />
      <el-table-column prop="country" label="国家" width="100" />
      <el-table-column prop="store_name" label="店铺" width="130" />
      <el-table-column prop="sku" label="商品SKU" width="180" />
      <el-table-column prop="quantity" label="数量" width="80" />
      <el-table-column prop="order_id" label="订单号" width="120" />
      <el-table-column prop="refund_amount" label="平台退款金额（当地货币）">
        <template #default="{ row }">{{ Number(row.refund_amount).toFixed(2) }} {{ getCurrencyCodeByCountry(row.country) }}</template>
      </el-table-column>
      <el-table-column prop="loss_amount" label="利润亏损金额">
        <template #default="{ row }">
          <span :style="{ color: row.loss_amount >= 0 ? '#F56C6C' : '#67C23A' }">￥{{ Number(row.loss_amount).toFixed(2) }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="return_date" label="退货日期" width="120" />
      <el-table-column prop="handler" label="处理人" width="120">
        <template #default="{ row }">{{ displayHandlerName(row.handler) }}</template>
      </el-table-column>
      <el-table-column v-if="isSuperAdmin" label="提交人" width="180">
        <template #default="{ row }">
          <span>{{ row.creator_username || '未知' }}</span>
          <span style="color:#909399;">{{ row.creator_role ? '（' + row.creator_role + '）' : '' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="160">
        <template #default="{ row }">
          <el-button size="small" @click="editRow(row)">编辑</el-button>
          <el-button v-if="isSuperAdmin" size="small" type="danger" @click="removeRow(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div style="margin-top:8px;display:flex;justify-content:space-between;align-items:center;">
      <div>共 {{ items.length }} 条退货记录，累计亏损： <strong style="color:#F56C6C;">￥{{ totalLoss.toFixed(2) }}</strong></div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})

const DEFAULT_COUNTRY = '印尼'

const items = ref([])
const dialogVisible = ref(false)
const dialogWidth = ref('70vw')
const filters = ref({
  dateRange: [],
})
const countries = ref([])
const enabledShops = ref([])
const products = ref([])
const rates = ref({ PHP: 0, IDR: 0, MYR: 0, USD: 0, CNY: 1 })
const loading = ref(false)

const countryCurrencyMap = {
  '菲律宾': 'PHP',
  '印尼': 'IDR',
  '马来西亚': 'MYR',
}

function getCurrentHandler() {
  const role = String(props.currentUser?.role || '').trim().toLowerCase()
  const username = String(props.currentUser?.username || '').trim()
  const usernameLower = username.toLowerCase()
  if (role === 'admin' || role === 'root' || usernameLower === 'root' || usernameLower === 'admin') {
    return 'yangjian'
  }
  return username
}

const form = ref({
  id: 0,
  country: DEFAULT_COUNTRY,
  store_name: '',
  order_id: '',
  product_name: '',
  sku: '',
  quantity: 1,
  refund_amount: 0,
  loss_amount: 0,
  return_date: '',
  handler: getCurrentHandler(),
  remark: '',
})

const localCurrencyCode = computed(() => getCurrencyCodeByCountry(form.value.country))

const currentRate = computed(() => {
  const cur = localCurrencyCode.value
  if (cur === 'CNY') return 1
  const foreignPerCny = Number(rates.value[cur]) || 0
  if (!foreignPerCny) return 0
  // 后端汇率含义为 1 CNY ≈ foreignPerCny 外币，换算成人民币/外币需要取倒数。
  return 1 / foreignPerCny
})

const currentRateText = computed(() => {
  const r = Number(currentRate.value) || 0
  if (!r) return '-'
  return r.toFixed(6)
})

const selectedSkuCost = computed(() => {
  const sku = String(form.value.sku || '').trim()
  if (!sku) return 0
  const p = (products.value || []).find(x => String(x?.sku || '').trim() === sku)
  return Number(p?.cost) || 0
})

const lossCalcDetail = computed(() => {
  const refund = Number(form.value.refund_amount) || 0
  const rate = Number(currentRate.value) || 0
  const cost = Number(selectedSkuCost.value) || 0
  const qty = Number(form.value.quantity) || 0
  const refundCny = refund * rate
  const totalCost = cost * qty
  return `计算：${refund.toFixed(2)} × ${rate.toFixed(6)} - ${cost.toFixed(2)} × ${qty} = ${refundCny.toFixed(2)} - ${totalCost.toFixed(2)} = ${(refundCny - totalCost).toFixed(2)} CNY`
})

const totalLoss = computed(() => items.value.reduce((s, it) => s + (Number(it.loss_amount) || 0), 0))
const isSuperAdmin = computed(() => String(props.currentUser?.role || '').trim().toLowerCase() === 'superadmin')

function displayHandlerName(name) {
  const raw = String(name || '').trim()
  return raw === 'root' || raw === 'admin' ? 'yangjian' : raw
}

function onReset() {
  form.value = { id: 0, country: DEFAULT_COUNTRY, store_name: '', order_id: '', product_name: '', sku: '', quantity: 1, refund_amount: 0, loss_amount: 0, return_date: '', handler: getCurrentHandler(), remark: '' }
  recalcLossAmount()
  loadEnabledShops()
}

function openCreateDialog() {
  onReset()
  dialogVisible.value = true
}

function updateDialogWidth() {
  if (typeof window === 'undefined') return
  dialogWidth.value = window.innerWidth <= 768 ? '92vw' : '560px'
}

function scaleHint(val) {
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

function getCurrencyCodeByCountry(country) {
  return countryCurrencyMap[String(country || '').trim()] || 'CNY'
}

function recalcLossAmount() {
  const refund = Number(form.value.refund_amount) || 0
  const rate = Number(currentRate.value) || 0
  const cost = Number(selectedSkuCost.value) || 0
  const qty = Number(form.value.quantity) || 0
  const result = refund * rate - cost * qty
  form.value.loss_amount = Number(result.toFixed(2))
}

async function load() {
  loading.value = true
  try {
    const params = {}
    if (Array.isArray(filters.value.dateRange) && filters.value.dateRange.length === 2) {
      params.start_date = filters.value.dateRange[0]
      params.end_date = filters.value.dateRange[1]
    }
    const res = await axios.get('/api/returns', { params })
    items.value = Array.isArray(res.data) ? res.data : []
  } catch (e) {
    items.value = []
  } finally {
    loading.value = false
  }
}

function resetFilters() {
  filters.value.dateRange = []
  load()
}

async function loadEnabledShops() {
  try {
    const params = { status: 'enabled' }
    if (form.value.country) {
      params.country = form.value.country
    }
    const res = await axios.get('/api/shops', { params })
    const list = Array.isArray(res.data?.items) ? res.data.items : []
    enabledShops.value = list
      .filter(it => it && it.name)
      .sort((a, b) => String(a.name).localeCompare(String(b.name), 'zh-CN', { numeric: true }))

    if (form.value.store_name && !enabledShops.value.some(it => it.name === form.value.store_name)) {
      form.value.store_name = ''
    }
  } catch (e) {
    enabledShops.value = []
  }
}

async function loadCountries() {
  try {
    const res = await axios.get('/api/countries')
    const list = Array.isArray(res.data?.items) ? res.data.items : []
    countries.value = list.includes(DEFAULT_COUNTRY) ? list : [DEFAULT_COUNTRY, ...list]
  } catch (e) {
    countries.value = [DEFAULT_COUNTRY]
  }
}

async function loadProducts() {
  try {
    const res = await axios.get('/api/products')
    products.value = Array.isArray(res.data) ? res.data : []
  } catch (e) {
    products.value = []
  }
}

async function loadRates() {
  try {
    const res = await axios.get('/api/exchange/rates')
    rates.value = {
      PHP: res.data?.rates?.PHP || 0,
      IDR: res.data?.rates?.IDR || 0,
      MYR: res.data?.rates?.MYR || 0,
      USD: res.data?.rates?.USD || 0,
      CNY: 1,
    }
    recalcLossAmount()
  } catch (e) {
  }
}

function onSkuChange(val) {
  const p = products.value.find(x => x.sku === val)
  form.value.product_name = p ? p.name : ''
}

function onCountryChange() {
  loadEnabledShops()
  recalcLossAmount()
}

function editRow(row) {
  form.value = Object.assign({}, row)
  form.value.order_id = String(form.value.order_id || '').trim()
  recalcLossAmount()
  form.value.handler = getCurrentHandler()
  loadEnabledShops()
  dialogVisible.value = true
}

async function onSave() {
  const orderID = String(form.value.order_id || '').trim()
  const orderIDPattern = /^(?=.*\d)[A-Za-z0-9]+$/
  if (!orderID) {
    ElMessage.error('订单号：不能为空')
    return
  }
  if (!orderIDPattern.test(orderID)) {
    ElMessage.error('订单号：需为纯数字或数字+字母组合（仅限大小写字母和数字）')
    return
  }

  try {
    const payload = Object.assign({}, form.value)
    payload.order_id = orderID
    payload.handler = displayHandlerName(payload.handler)
    await axios.post('/api/returns', payload)
    await load()
    dialogVisible.value = false
    onReset()
  } catch (e) {
    const field = e?.response?.data?.field
    const msg = e?.response?.data?.error || '保存失败'
    if (field === 'order_id') {
      ElMessage.error('订单号：' + msg)
      return
    }
    ElMessage.error(msg)
  }
}

async function removeRow(row) {
  try {
    await axios.delete('/api/returns/' + row.id)
    await load()
  } catch (e) {
  }
}

onMounted(() => {
  updateDialogWidth()
  if (typeof window !== 'undefined') {
    window.addEventListener('resize', updateDialogWidth)
  }
  loadCountries()
  loadEnabledShops()
  loadProducts()
  loadRates()
  form.value.handler = getCurrentHandler()
  recalcLossAmount()
  load()
})

watch(
  () => [form.value.refund_amount, form.value.country, form.value.sku, form.value.quantity, rates.value.PHP, rates.value.IDR, rates.value.MYR, rates.value.USD, products.value.length],
  () => {
    recalcLossAmount()
  },
)

onBeforeUnmount(() => {
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', updateDialogWidth)
  }
})
</script>

<style scoped>
.return-page {
  font-size: 14px;
}

.return-page :deep(.el-form-item__label),
.return-page :deep(.el-input__inner),
.return-page :deep(.el-select__selected-item),
.return-page :deep(.el-button),
.return-page :deep(.el-table),
.return-page :deep(.el-date-editor .el-input__inner) {
  font-size: 14px;
}

.amount-field {
  width: 280px;
  max-width: 100%;
}

.amount-inline {
  display: flex;
  align-items: center;
  gap: 8px;
}

.currency-tag {
  flex-shrink: 0;
  color: #606266;
  font-size: 12px;
}

.dialog-vertical-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

.dialog-vertical-form :deep(.el-input-number),
.dialog-vertical-form :deep(.el-select),
.dialog-vertical-form :deep(.el-date-editor.el-input),
.dialog-vertical-form :deep(.el-input),
.dialog-vertical-form :deep(.el-textarea) {
  width: 280px !important;
  max-width: 100%;
}

.dialog-vertical-form :deep(.return-date-item .el-date-editor.el-input) {
  width: 280px !important;
  max-width: 100%;
}

.scale-hint {
  margin-top: 1px;
  min-height: 16px;
  font-size: 11px;
  color: #909399;
  text-align: right;
  line-height: 16px;
}

.rate-hint {
  margin-top: 1px;
  min-height: 16px;
  font-size: 11px;
  color: #909399;
  text-align: right;
  line-height: 16px;
}

.calc-hint {
  margin-top: 1px;
  min-height: 16px;
  font-size: 11px;
  color: #606266;
  text-align: right;
  line-height: 16px;
}
</style>
