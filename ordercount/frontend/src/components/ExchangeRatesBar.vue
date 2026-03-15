<template>
  <el-card shadow="never" body-style="padding: 10px 20px;">
    <div class="exchange-bar-header">
      <div class="exchange-bar-title">
        <span class="title-main">当前主要汇率</span>
        <span class="title-sub">（1 人民币 ≈ ? 外币）</span>
      </div>
      <div class="exchange-bar-status">
        <template v-if="loading">
          <span class="status-loading">正在获取最新汇率...</span>
        </template>
        <template v-else-if="error">
          <span class="status-error">{{ error }}</span>
        </template>
        <template v-else-if="lastUpdated">
          <span class="status-updated">更新于 {{ lastUpdated }}</span>
        </template>
      </div>
    </div>

    <div class="exchange-bar-content" :class="{ 'exchange-bar--dim': !!error }">
      <div
        v-for="item in displayRates"
        :key="item.code"
        class="exchange-rate-item"
      >
        <span class="rate-label">{{ item.label }} ({{ item.code }})：</span>
        <span class="rate-value">
          1 CNY ≈
          <span class="rate-number">{{ formatRate(item.value) }}</span>
          <span class="rate-code">{{ item.code }}</span>
        </span>
      </div>
    </div>
  </el-card>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import axios from 'axios'

// 后端返回的 rates 含义：1 人民币 ≈ rates[币种] 外币
// 例如：rates.PHP = 8.37 表示 1 CNY ≈ 8.37 PHP
const rates = ref({ PHP: 0, IDR: 0, MYR: 0 })
const loading = ref(false)
const error = ref('')
const lastUpdated = ref('')

const displayRates = computed(() => [
	{ code: 'PHP', label: '菲律宾', value: rates.value.PHP },
	{ code: 'IDR', label: '印尼', value: rates.value.IDR },
	{ code: 'MYR', label: '马来西亚', value: rates.value.MYR },
])

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
    lastUpdated.value = new Date().toLocaleString()
  } catch (e) {
    if (!rates.value.PHP && !rates.value.IDR && !rates.value.MYR) {
      error.value = '汇率获取失败，请稍后重试'
    } else {
      error.value = '汇率更新失败，已展示上次成功的数据'
    }
  } finally {
    loading.value = false
  }
}
let timer = null

onMounted(() => {
  loadRates()
  timer = setInterval(loadRates, 10 * 60 * 1000)
})

onBeforeUnmount(() => {
  if (timer) {
    clearInterval(timer)
    timer = null
  }
})
</script>

<style scoped>
.exchange-bar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 6px;
}

.exchange-bar-title {
  display: flex;
  align-items: baseline;
  gap: 4px;
}

.title-main {
  font-weight: 600;
  font-size: 14px;
}

.title-sub {
  font-size: 12px;
  color: #909399;
}

.exchange-bar-status {
  font-size: 12px;
}

.status-loading {
  color: #909399;
}

.status-error {
  color: #f56c6c;
}

.status-updated {
  color: #909399;
}

.exchange-bar-content {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 24px;
  font-size: 13px;
}

.exchange-bar--dim {
  opacity: 0.8;
}

.exchange-rate-item {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.rate-label {
  color: #606266;
}

.rate-value {
  color: #303133;
}

.rate-number {
  font-weight: 600;
  margin: 0 2px;
}

.rate-code {
  color: #909399;
  margin-left: 2px;
}
</style>
