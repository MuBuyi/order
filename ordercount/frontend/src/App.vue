<template>
  <el-container style="min-height:100vh;">
    <el-header style="background:#409EFF;color:#fff;font-size:22px;">订单统计管理后台</el-header>
    <el-main>
      <el-row :gutter="20">
        <el-col :span="12">
          <OrderForm @refresh="refreshAll" />
          <OrderImport @refresh="refreshAll" />
        </el-col>
        <el-col :span="12">
          <TodaySales ref="todaySales" />
          <OrderCharts ref="orderCharts" />
        </el-col>
      </el-row>
      <el-divider />
      <OrderExport />
    </el-main>
  </el-container>
</template>

<script setup>
import { ref } from 'vue'
import OrderForm from './components/OrderForm.vue'
import OrderImport from './components/OrderImport.vue'
import TodaySales from './components/TodaySales.vue'
import OrderCharts from './components/OrderCharts.vue'
import OrderExport from './components/OrderExport.vue'

const todaySales = ref(null)
const orderCharts = ref(null)
function refreshAll() {
  todaySales.value && todaySales.value.load()
  orderCharts.value && orderCharts.value.load()
}
</script>

<style>
body {
  margin: 0;
  background: #f5f7fa;
}
</style>
