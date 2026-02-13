<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>数据可视化图表</template>
    <div ref="trendRef" style="height:220px;width:100%;margin-bottom:20px;"></div>
    <div ref="topRef" style="height:220px;width:100%;"></div>
  </el-card>
</template>
<script setup>
import { ref, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import axios from 'axios'
const trendRef = ref(null)
const topRef = ref(null)
let trendChart, topChart
function renderTrend(data){
  nextTick(()=>{
    if(!trendChart) trendChart = echarts.init(trendRef.value)
    if(!data || data.length === 0){
      trendChart.clear()
      trendChart.setOption({title:{text:'近30天销售趋势',left:'center'},xAxis:{type:'category',data:[]},yAxis:{type:'value'},series:[],graphic:{type:'text',left:'center',top:'middle',style:{text:'暂无数据',fontSize:18,fill:'#888'}}})
      return
    }
    trendChart.setOption({
      title:{text:'近30天销售趋势',left:'center'},
      xAxis:{type:'category',data:data.map(i=>i.Day)},
      yAxis:{type:'value'},
      series:[{type:'line',data:data.map(i=>i.Total),smooth:true,areaStyle:{}}]
    })
    trendChart.resize()
  })
}
function renderTop(data){
  nextTick(()=>{
    if(!topChart) topChart = echarts.init(topRef.value)
    if(!data || data.length === 0){
      topChart.clear()
      topChart.setOption({title:{text:'商品销售排行',left:'center'},xAxis:{type:'category',data:[]},yAxis:{type:'value'},series:[],graphic:{type:'text',left:'center',top:'middle',style:{text:'暂无数据',fontSize:18,fill:'#888'}}})
      return
    }
    topChart.setOption({
      title:{text:'商品销售排行',left:'center'},
      xAxis:{type:'category',data:data.map(i=>i.ProductName)},
      yAxis:{type:'value'},
      series:[{type:'bar',data:data.map(i=>i.Total),itemStyle:{color:'#67C23A'}}]
    })
    topChart.resize()
  })
}
function load(){
  axios.get('/api/stats/sales-trend').then(res=>renderTrend(res.data||[]))
  axios.get('/api/stats/top-products').then(res=>renderTop(res.data||[]))
}
onMounted(load)
defineExpose({load})
</script>
