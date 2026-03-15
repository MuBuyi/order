<template>
  <div style="padding:16px;background:#fff;border-radius:8px;">
    <h3 style="margin:0 0 8px 0">汇率换算</h3>

    <div v-if="loading" style="color:#666;margin-bottom:12px">正在加载汇率...</div>

    <div v-else>
      <div style="font-size:14px;color:#333">
        <div>
          1 {{ labels[from] || from }} ≈ <strong style="font-size:18px">{{ fmt(rateFromTo) }}</strong> {{ labels[to] || to }}
        </div>
        <div style="margin-top:6px;color:#666">
          1 {{ labels[to] || to }} ≈ <strong style="font-size:18px">{{ fmt(rateToFrom) }}</strong> {{ labels[from] || from }}
        </div>
      </div>

      <div class="cc-main-row">
        <div class="cc-card">
          <el-input v-model="amount" type="number" class="cc-input">
            <template #prepend>
              <el-select v-model="from" class="cc-select" placeholder="从">
                <el-option v-for="c in currencies" :key="c" :label="optionLabel(c)" :value="c" />
              </el-select>
            </template>
          </el-input>
          <div class="cc-scale">{{ scaleHint(amount) }}</div>
        </div>

        <span class="cc-swap" @click="swap">⇄</span>

        <div class="cc-card">
          <el-input :model-value="displayResult" readonly class="cc-input">
            <template #prepend>
              <el-select v-model="to" class="cc-select" placeholder="到">
                <el-option v-for="c in currencies" :key="c" :label="optionLabel(c)" :value="c" />
              </el-select>
            </template>
          </el-input>
        </div>
      </div>

      <div style="margin-top:12px;display:flex;align-items:center;gap:12px">
        <el-button plain @click="refreshRates">刷新汇率</el-button>
        <div style="margin-left:8px;color:#999">更新时间：{{ lastUpdated }} 数据仅供参考</div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import axios from 'axios'

const amount = ref(1)
const from = ref('CNY')
const to = ref('USD')
const rates = ref({})
const labels = ref({})
const loading = ref(false)
const result = ref(null)
const lastUpdated = ref('')

const flagMap = {
  PHP: '🇵🇭',
  CNY: '🇨🇳',
  USD: '🇺🇸',
  IDR: '🇮🇩',
  MYR: '🇲🇾',
}

const currencies = computed(() => Object.keys(rates.value).sort())

function optionLabel(code) {
  return `${flagMap[code] || ''} ${labels.value[code] || code}`.trim()
}

function fmt(v) {
  if (v === null || v === undefined || Number.isNaN(v)) return '-'
  // show up to 4 decimal places, but trim trailing zeros
  const s = Number(v).toFixed(4)
  return s.replace(/\.0+$|(?<=\.[0-9]*?)0+$/,'')
}

const rateFromTo = computed(() => {
  if (!rates.value || !rates.value[from.value] || !rates.value[to.value]) return null
  return rates.value[to.value] / rates.value[from.value]
})

const rateToFrom = computed(() => {
  if (!rateFromTo.value) return null
  return 1 / rateFromTo.value
})

const displayResult = computed(() => {
  if (result.value === null || result.value === undefined) return ''
  return fmt(result.value) + ' ' + (labels.value[to.value] || to.value)
})

function scaleHint (val) {
  const v = Number(val) || 0
  if (!v) return ''
  if (v >= 100000000) return '亿'
  if (v >= 10000000) return '千万'
  if (v >= 1000000) return '百万'
  if (v >= 100000) return '十万'
  if (v >= 10000) return '万'
  if (v >= 1000) return '千'
  if (v >= 100) return '百'
  return ''
}

async function loadRates() {
  loading.value = true
  try {
    const res = await axios.get('/api/exchange/rates')
    if (res.data && res.data.rates) {
      rates.value = res.data.rates
      labels.value = res.data.labels || {}
      if (!rates.value['CNY']) rates.value['CNY'] = 1
      if (!labels.value['CNY']) labels.value['CNY'] = '人民币'
      lastUpdated.value = new Date().toLocaleString()
    }
  } catch (e) {
    console.error('load rates error', e)
  }
    // 在成功加载汇率后自动计算当前转换结果
    try {
      doConvert()
    } catch (e) {
      // ignore
    }
    loading.value = false
}

function doConvert() {
  if (!rates.value || !rates.value[from.value] || !rates.value[to.value]) {
    result.value = null
    return
  }
  const factor = rates.value[to.value] / rates.value[from.value]
  const amt = Number(amount.value) || 0
  result.value = Number((amt * factor).toFixed(4))
}

function swap() {
  const t = from.value
  from.value = to.value
  to.value = t
  doConvert()
}

function refreshRates() {
  loadRates()
}

onMounted(async () => {
  await loadRates()
  doConvert()
})

// 当 amount/from/to 变化时自动转换
watch([amount, from, to], () => {
  doConvert()
})
</script>

<style scoped>
h3 { margin: 0; }
.cc-main-row {
  margin-top: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
}
.cc-card {
  width: 300px;
}
.cc-input :deep(.el-input__wrapper) {
  height: 36px;
  border-radius: 6px;
}
.cc-input :deep(.el-input-group__prepend) {
  padding: 0 8px;
}
.cc-select {
  width: 120px;
}
.cc-scale {
  margin-top: 2px;
  font-size: 11px;
  color: #909399;
  min-height: 16px;
}
.cc-swap {
  display: inline-block;
  font-size: 18px;
  color: #409EFF;
  cursor: pointer;
  user-select: none;
}
.cc-swap:hover {
  color: #66b1ff;
}
</style>
