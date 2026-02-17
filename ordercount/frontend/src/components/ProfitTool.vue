<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      结账工具（当天利润计算）
    </template>
    <el-form label-width="120px" @submit.prevent>
      <el-row :gutter="10">
        <el-col :span="12">
          <el-form-item label="国家">
            <div style="display:inline-block;width:60%;">
              <el-select
                v-model="country"
                placeholder="请选择国家"
                style="width:100%;"
                @change="onCountryChange"
              >
                <el-option label="菲律宾" value="菲律宾" />
                <el-option label="印尼" value="印尼" />
                <el-option label="马来西亚" value="马来西亚" />
              </el-select>
            </div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="当天销售总额">
            <el-input-number
              v-model="saleTotal"
              :min="0"
              :step="0.01"
              style="width:100%;"
              :disabled="true"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="广告费">
            <div style="position:relative;display:inline-block;width:60%;padding-bottom:14px;">
              <el-input-number
                v-model="adCost"
                :min="0"
                :step="0.01"
                style="width:100%;"
                placeholder="请输入本国货币数值"
              />
              <div v-if="adCostLevel" style="position:absolute;right:4px;bottom:0;font-size:10px;color:#909399;line-height:1;">
                {{ adCostLevel }}
              </div>
            </div>
          </el-form-item>
        </el-col>

        <el-col :span="12">
          <el-form-item label="货款成本">
            <el-input-number
              v-model="goodsCost"
              :min="0"
              :step="0.01"
              style="width:100%;"
              :disabled="true"
            />
          </el-form-item>
        </el-col>

        <el-col :span="12">
          <el-form-item label="刷单费用">
            <div style="display:flex;align-items:flex-start;gap:8px;">
              <div style="position:relative;display:inline-block;width:28%;padding-bottom:14px;">
                <el-input-number
                  v-model="shuaDanFee"
                  :min="0"
                  :step="0.01"
                  style="width:100%;"
                  placeholder="本国货币/美元"
                />
                <div
                  v-if="shuaDanLevel"
                  style="position:absolute;right:4px;bottom:0;font-size:10px;color:#909399;line-height:1;"
                >
                  {{ shuaDanLevel }}
                </div>
              </div>
              <div style="transform:scale(0.85);transform-origin:left center;margin-top:2px;">
                <el-switch
                  v-model="shuaDanUseUSD"
                  active-text="美元"
                  inactive-text="本国货币"
                  size="small"
                />
              </div>
              <div style="width:18%;">
                <el-input-number
                  v-model="shuaDanCount"
                  :min="0"
                  :step="1"
                  controls-position="right"
                  style="width:100%;font-size:12px;"
                  placeholder="刷单数"
                />
              </div>
            </div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="对应汇率">
            <el-input-number
              v-model="exchangeRate"
              :min="0"
              :step="0.0001"
              style="width:100%;"
              :disabled="autoFollowRate"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="固定成本">
            <div style="position:relative;display:inline-block;width:60%;padding-bottom:14px;">
              <el-input-number v-model="fixedCost" :min="0" :step="0.01" style="width:100%;" />
              <div v-if="fixedCostLevel" style="position:absolute;right:4px;bottom:0;font-size:10px;color:#909399;line-height:1;">
                {{ fixedCostLevel }}
              </div>
            </div>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="汇率模式">
            <el-switch
              v-model="autoFollowRate"
              active-text="自动跟随实时汇率"
              inactive-text="手动输入并锁定"
            />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注">
            <el-input v-model="remark" placeholder="可记录活动、说明等" />
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <el-divider />

    <div style="text-align:right;margin-bottom:10px;">
      <el-button type="primary" @click="onSave" :loading="saving">计算</el-button>
      <span v-if="saveMsg" style="margin-left:10px;font-size:12px;" :style="{color: saveOk ? '#67C23A' : '#F56C6C'}">{{ saveMsg }}</span>
    </div>

    <el-descriptions title="计算结果" :column="1" size="small" border>
      <el-descriptions-item label="选择国家">
        {{ country || '未选择' }} ({{ currentCurrency || '-' }})
      </el-descriptions-item>
      <el-descriptions-item label="广告费折算">
        - ( (广告费 + 广告费 × 11%) × 汇率 ) = -{{ adDeduction.toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="平台手续费">
        - 当天销售总额 × 7% = -{{ platformFee.toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="货款成本">
        -{{ goodsCost.toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="刷单费用">
        - ( 输入货币 × 汇率 × 7% + 刷单数 × 2 ) = -{{ shuaDanCost.toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="固定成本">
        -{{ (Number(fixedCost) || 0).toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="当天销售总额">
        {{ saleTotal.toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="当天利润">
        <span :style="{ color: profit >= 0 ? '#67C23A' : '#F56C6C', 'font-weight': 'bold' }">
          {{ profit.toFixed(2) }}
        </span>
      </el-descriptions-item>
    </el-descriptions>

    <el-divider />

    <div style="margin-top:10px;">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:4px;">
        <span style="font-size:14px;">广告费折算趋势</span>
        <div style="display:flex;align-items:center;gap:8px;">
          <el-radio-group v-model="trendMode" size="small">
            <el-radio-button label="daily">近7天</el-radio-button>
            <el-radio-button label="monthly">按月</el-radio-button>
          </el-radio-group>
          <el-date-picker
            v-if="trendMode === 'monthly'"
            v-model="trendYear"
            type="year"
            size="small"
            value-format="YYYY"
            placeholder="选择年份"
          />
        </div>
      </div>
      <div ref="adTrendRef" style="height:260px;width:100%;"></div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'

// 国家与币种映射
const countryCurrencyMap = {
  '菲律宾': 'PHP',
  '印尼': 'IDR',
  '马来西亚': 'MYR',
}

// 当前选择的国家和币种（默认印尼）
const country = ref('印尼')
const currentCurrency = ref('')

// 当天销售总额
const saleTotal = ref(0)
// 广告费
const adCost = ref(null)
// 当天对应货币汇率
const exchangeRate = ref(1)
// 货款成本
const goodsCost = ref(0)
// 刷单费用
const shuaDanFee = ref(null)
// 刷单数
const shuaDanCount = ref(null)
// 刷单费用是否使用美元
const shuaDanUseUSD = ref(false)
// 固定成本
const fixedCost = ref(null)

// 备注
const remark = ref('')

// 是否自动跟随实时汇率
const autoFollowRate = ref(true)

// 汇率表（1 人民币 ≈ rates[币种] 外币），需取倒数后用于“1 外币 ≈ ? 人民币”
const rates = ref({ PHP: 0, IDR: 0, MYR: 0 })

const saving = ref(false)
const saveMsg = ref('')
const saveOk = ref(false)

const adTrendRef = ref(null)
let adTrendChart

const trendMode = ref('daily') // daily: 近7天, monthly: 按月
const trendYear = ref(new Date().getFullYear().toString())

const STORAGE_KEY = 'ordercount_profit_tool_last_v1'

function getLevelLabel (value) {
  const n = Math.abs(Number(value) || 0)
  if (n >= 10000000) return '千万'
  if (n >= 1000000) return '百万'
  if (n >= 100000) return '十万'
  if (n >= 10000) return '万'
  if (n >= 1000) return '千'
  if (n >= 100) return '百'
  return ''
}

async function loadRates() {
  try {
    const res = await axios.get('/api/exchange/rates')
    rates.value = {
      PHP: res.data?.rates?.PHP || 0,
      IDR: res.data?.rates?.IDR || 0,
      MYR: res.data?.rates?.MYR || 0,
    }
    // 如果已经选了国家，根据最新汇率同步
    if (country.value && autoFollowRate.value) {
      onCountryChange(country.value)
    }
  } catch (e) {
    // 失败时保留现有数值，不中断使用
  }
}

function onCountryChange(val) {
  const cur = countryCurrencyMap[val] || ''
  currentCurrency.value = cur
  if (autoFollowRate.value && cur && rates.value[cur]) {
    // 后端返回：1 CNY ≈ rates[cur] 外币
    // 我们需要：1 外币 ≈ ? CNY，因此在这里取倒数
    exchangeRate.value = 1 / rates.value[cur]
  }
}

async function loadTodayGoodsCost () {
  try {
    const res = await axios.get('/api/costs/today')
    goodsCost.value = res.data?.total_cost || 0
  } catch (e) {
    // 失败时保持手动输入的值
  }
}

function loadLastInput () {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw)

    if (typeof data.autoFollowRate === 'boolean') {
      autoFollowRate.value = data.autoFollowRate
    }

    if (data.country) {
      country.value = data.country
      onCountryChange(data.country)
    }

    if (typeof data.adCost === 'number') {
      adCost.value = data.adCost
    }
    if (typeof data.shuaDanFee === 'number') {
      shuaDanFee.value = data.shuaDanFee
    }
    if (typeof data.shuaDanCount === 'number') {
      shuaDanCount.value = data.shuaDanCount
    }
    if (typeof data.shuaDanUseUSD === 'boolean') {
      shuaDanUseUSD.value = data.shuaDanUseUSD
    }
    if (typeof data.fixedCost === 'number') {
      fixedCost.value = data.fixedCost
    }

    if (!autoFollowRate.value && typeof data.exchangeRate === 'number' && data.exchangeRate > 0) {
      exchangeRate.value = data.exchangeRate
    }

    if (typeof data.remark === 'string') {
      remark.value = data.remark
    }
  } catch (e) {
    // 忽略本地存储解析错误
  }
}

onMounted(() => {
  loadRates()
  loadTodayGoodsCost()
  loadTodaySales()
		loadAdTrend()
  loadLastInput()
})

async function loadTodaySales () {
  try {
    const res = await axios.get('/api/sales/today')
    saleTotal.value = res.data?.total_amount || 0
  } catch (e) {
    // 失败时保持手动输入的值
  }
}

async function loadAdTrend () {
  try {
    if (trendMode.value === 'daily') {
      const res = await axios.get('/api/stats/ad-deduction/daily', { params: { days: 7 } })
      const data = Array.isArray(res.data) ? res.data : []
      const labels = data.map(i => (i.Day || '').split('T')[0] || i.Day || '')
      const values = data.map(i => Number(i.Total || 0))
      renderAdTrendChart(labels, values, '近7天广告费折算（人民币）')
    } else {
      const year = trendYear.value || new Date().getFullYear().toString()
      const res = await axios.get('/api/stats/ad-deduction/monthly', { params: { year } })
      const months = Array.from({ length: 12 }, (_, i) => `${i + 1}月`)
      const values = Array.isArray(res.data?.monthly) ? res.data.monthly : []
      renderAdTrendChart(months, months.map((_, idx) => Number(values[idx] || 0)), `按月广告费折算（${year}）`)
    }
  } catch (e) {
    renderAdTrendChart([], [], '广告费数据加载失败')
  }
}

function renderAdTrendChart (labels, values, title) {
  if (!adTrendChart && adTrendRef.value) {
    adTrendChart = echarts.init(adTrendRef.value)
  }
  if (!adTrendChart) return

  adTrendChart.setOption({
    title: { text: title, left: 'center' },
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: labels },
    yAxis: { type: 'value' },
    series: [{
      type: 'bar',
      data: values,
      label: { show: true, position: 'top' },
      itemStyle: { color: '#E6A23C' }
    }]
  })
  adTrendChart.resize()
}

watch(trendMode, () => {
  loadAdTrend()
})

watch(trendYear, () => {
  if (trendMode.value === 'monthly') {
    loadAdTrend()
  }
})

async function onSave() {
  saveMsg.value = ''
  saveOk.value = false
  saving.value = true
  try {
    const today = new Date()
    const dateStr = today.toISOString().slice(0, 10)
    const payload = {
      date: dateStr,
      country: country.value,
      currency: currentCurrency.value,
      sale_total: Number(saleTotal.value) || 0,
      ad_cost: Number(adCost.value) || 0,
      exchange: Number(exchangeRate.value) || 0,
      goods_cost: Number(goodsCost.value) || 0,
      shua_dan_fee: Number(shuaDanFee.value) || 0,
      fixed_cost: Number(fixedCost.value) || 0,
      remark: remark.value,
    }
    const res = await axios.post('/api/settlement', payload)
    if (res.data && !res.data.error) {
      saveOk.value = true
      saveMsg.value = '结算记录已保存'

      // 保存本次输入，供下次自动回显
      try {
        const toSave = {
          country: country.value,
          adCost: Number(adCost.value) || 0,
          shuaDanFee: Number(shuaDanFee.value) || 0,
          shuaDanCount: Number(shuaDanCount.value) || 0,
          shuaDanUseUSD: !!shuaDanUseUSD.value,
          fixedCost: Number(fixedCost.value) || 0,
          autoFollowRate: !!autoFollowRate.value,
          exchangeRate: Number(exchangeRate.value) || 0,
          remark: remark.value || '',
        }
        localStorage.setItem(STORAGE_KEY, JSON.stringify(toSave))
      } catch (e) {
        // 忽略本地存储失败
      }
    } else {
      saveMsg.value = res.data.error || '保存失败'
    }
  } catch (e) {
    // 优先展示后端返回的 error，便于排查企微推送失败等问题
    const msg = e?.response?.data?.error || e?.message || '保存失败'
    saveMsg.value = msg
  } finally {
    saving.value = false
  }
}

// 广告费折算：广告费 = (广告费 + 广告费 * 11%) * 汇率
const adDeduction = computed(() => {
  const a = Number(adCost.value) || 0
  const r = Number(exchangeRate.value) || 0
  return a * 1.11 * r
})

// 平台手续费：当天销售总额 * 7%
const platformFee = computed(() => {
  const s = Number(saleTotal.value) || 0
  return s * 0.07
})

// 刷单费用：(输入货币 × 汇率 × 7%) + 输入的刷单数量 × 2
const shuaDanCost = computed(() => {
  const fee = Number(shuaDanFee.value) || 0
  const r = Number(exchangeRate.value) || 0
  const count = Number(shuaDanCount.value) || 0
  return fee * r * 0.07 + count * 2
})

// 当天利润 = 当天销售总额
//          - (广告费 + 广告费 * 11%) * 当天对应货币的汇率
//          - 货款成本
//          - 当天销售总额 * 7%
//          - 刷单费用
//          - 固定成本
const profit = computed(() => {
  const s = Number(saleTotal.value) || 0
  const g = Number(goodsCost.value) || 0
  const f = Number(fixedCost.value) || 0
  return s - adDeduction.value - g - platformFee.value - shuaDanCost.value - f
})

const adCostLevel = computed(() => getLevelLabel(adCost.value))
const shuaDanLevel = computed(() => getLevelLabel(shuaDanFee.value))
const fixedCostLevel = computed(() => getLevelLabel(fixedCost.value))
</script>
