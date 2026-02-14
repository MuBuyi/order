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
function load(){
  axios.get('/api/sales/today').then(res=>{
    total.value = res.data.total_amount || 0
  })
}
onMounted(load)
defineExpose({load})
</script>
