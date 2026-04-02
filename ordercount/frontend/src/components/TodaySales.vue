<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      今日销售额汇总
    </template>
    <div style="text-align:center;font-size:24px;font-weight:bold;margin:10px 0;">
      人民币金额：￥{{ total.toFixed(2) }}
    </div>
  </el-card>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const total = ref(0)

function load (date) {
  const params = {}
  if (date) params.date = date
  axios.get('/api/sales/today', { params }).then(res => {
    total.value = res.data?.total_amount || 0
  }).catch(() => {
    // 出错时保持当前值
  })
}

onMounted(() => load())

defineExpose({ load })
</script>
