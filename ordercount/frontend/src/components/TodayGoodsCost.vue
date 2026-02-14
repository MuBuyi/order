<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      今日货款成本
    </template>
    <div style="font-size:14px;color:#606266;">
      <span style="font-weight:600;">今日货款成本（人民币）：</span>
      <span style="font-size:18px;color:#303133;">￥{{ total.toFixed(2) }}</span>
    </div>
  </el-card>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const total = ref(0)

function load () {
  axios.get('/api/costs/today').then(res => {
    total.value = res.data.total_cost || 0
  }).catch(() => {
    total.value = 0
  })
}

onMounted(load)

// 供父组件调用刷新
defineExpose({ load })
</script>
