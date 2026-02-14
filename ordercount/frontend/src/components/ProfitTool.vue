<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      结账工具（当天利润计算）
    </template>
    <el-form label-width="120px" @submit.prevent>
      <el-row :gutter="10">
        <el-col :span="12">
          <el-form-item label="国家">
            <el-select v-model="country" placeholder="请选择国家" style="width:100%;" @change="onCountryChange">
              <el-option label="菲律宾" value="菲律宾" />
              <el-option label="印尼" value="印尼" />
              <el-option label="马来西亚" value="马来西亚" />
            </el-select>
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
        <el-col :span="12">
          <el-form-item label="当天销售总额">
            <el-input-number v-model="saleTotal" :min="0" :step="0.01" style="width:100%;" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="广告费">
            <el-input-number v-model="adCost" :min="0" :step="0.01" style="width:100%;" />
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
          <el-form-item label="货款成本">
            <el-input-number v-model="goodsCost" :min="0" :step="0.01" style="width:100%;" />
          </el-form-item>
        </el-col>

        <el-col :span="12">
          <el-form-item label="刷单费用">
            <el-input-number v-model="shuaDanFee" :min="0" :step="0.01" style="width:100%;" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="固定成本">
            <el-input-number v-model="fixedCost" :min="0" :step="0.01" style="width:100%;" />
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
      <el-button type="primary" @click="onSave" :loading="saving">保存本次结算</el-button>
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
        -{{ shuaDanFee.toFixed(2) }}
      </el-descriptions-item>
      <el-descriptions-item label="固定成本">
        -{{ fixedCost.toFixed(2) }}
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
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

// 国家与币种映射
const countryCurrencyMap = {
  '菲律宾': 'PHP',
  '印尼': 'IDR',
  '马来西亚': 'MYR',
}

// 当前选择的国家和币种
const country = ref('')
const currentCurrency = ref('')

// 当天销售总额
const saleTotal = ref(0)
// 广告费
const adCost = ref(0)
// 当天对应货币汇率
const exchangeRate = ref(1)
// 货款成本
const goodsCost = ref(0)
// 刷单费用（默认 0）
const shuaDanFee = ref(0)
// 固定成本（给一个默认值，可按需调整）
const fixedCost = ref(1000)

// 备注
const remark = ref('')

// 是否自动跟随实时汇率
const autoFollowRate = ref(true)

// 汇率表（1 人民币 ≈ rates[币种] 外币），需取倒数后用于“1 外币 ≈ ? 人民币”
const rates = ref({ PHP: 0, IDR: 0, MYR: 0 })

const saving = ref(false)
const saveMsg = ref('')
const saveOk = ref(false)

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

onMounted(loadRates)

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
    } else {
      saveMsg.value = res.data.error || '保存失败'
    }
  } catch (e) {
    saveMsg.value = e?.message || '保存失败'
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
  const sd = Number(shuaDanFee.value) || 0
  return s - adDeduction.value - g - platformFee.value - sd - f
})
</script>
