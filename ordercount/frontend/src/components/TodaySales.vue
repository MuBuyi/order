<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      今日销售额汇总
    </template>
    <el-table :data="currencies" size="small" border style="width:100%;margin-bottom:10px;">
      <el-table-column prop="Currency" label="币种" width="100" />
      <el-table-column prop="Sum" label="原币金额" />
      <el-table-column prop="cny_amount" label="人民币金额">
        <template #default="scope">
          ￥{{ Number(scope.row.cny_amount).toFixed(2) }}
        </template>
      </el-table-column>
    </el-table>
    <div style="text-align:right;font-weight:bold;">人民币总计：￥{{ total.toFixed(2) }}</div>
  </el-card>
</template>
<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const total = ref(0)
const currencies = ref([])
function load(){
  axios.get('/api/sales/today', { params: { cny: 1 } }).then(res=>{
    total.value = res.data.total_amount || 0
    currencies.value = res.data.currencies || []
  })
}
onMounted(load)
defineExpose({load})
</script>
