<template>
  <el-card shadow="hover">
    <template #header>导出报表</template>
    <el-button type="success" @click="onExport">导出全部订单 (CSV)</el-button>
  </el-card>
</template>
<script setup>
function onExport(){
  window.open('/api/report/export')
}
</script>
