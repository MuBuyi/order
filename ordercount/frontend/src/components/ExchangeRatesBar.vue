<template>
  <el-card shadow="never" body-style="padding: 10px 20px;">
    <div style="display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;">
      <div style="font-weight:bold;">当前主要汇率（1 人民币 ≈ ? 外币）</div>
      <div v-if="loading" style="color:#909399;font-size:12px;">正在获取最新汇率...</div>
      <div v-else-if="error" style="color:#F56C6C;font-size:12px;">{{ error }}</div>
      <div v-else style="display:flex;gap:16px;font-size:13px;flex-wrap:wrap;">
        <span>菲律宾 (PHP)：1 CNY ≈ {{ formatRate(rates.PHP) }} PHP</span>
        <span>印尼 (IDR)：1 CNY ≈ {{ formatRate(rates.IDR) }} IDR</span>
        <span>马来西亚 (MYR)：1 CNY ≈ {{ formatRate(rates.MYR) }} MYR</span>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

// 后端返回的 rates 含义：1 人民币 ≈ rates[币种] 外币
// 例如：rates.PHP = 8.37 表示 1 CNY ≈ 8.37 PHP
const rates = ref({ PHP: 0, IDR: 0, MYR: 0 })
const loading = ref(false)
const error = ref('')

function formatRate(v) {
  const n = Number(v) || 0
  if (n === 0) return '-'
  // 印尼盾一般数值较大，但小数部分较小，可适当多保留几位
  if (n < 0.01) return n.toFixed(6)
  return n.toFixed(4)
}

async function loadRates() {
  loading.value = true
  error.value = ''
  try {
    const res = await axios.get('/api/exchange/rates')
    rates.value = {
      PHP: res.data?.rates?.PHP || 0,
      IDR: res.data?.rates?.IDR || 0,
      MYR: res.data?.rates?.MYR || 0,
    }
  } catch (e) {
    error.value = '汇率获取失败，请稍后重试'
  } finally {
    loading.value = false
  }
}

onMounted(loadRates)
</script>
