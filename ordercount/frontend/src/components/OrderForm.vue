<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>手动添加订单</template>
    <el-form :model="form" :rules="rules" ref="formRef" label-width="90px" @submit.prevent>
      <el-row :gutter="10">
        <el-col :span="12">
          <el-form-item label="国家" prop="country">
            <el-select v-model="form.country" placeholder="请选择国家">
              <el-option label="菲律宾" value="菲律宾" />
              <el-option label="印尼" value="印尼" />
              <el-option label="马来西亚" value="马来西亚" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12"><el-form-item label="平台" prop="platform"><el-input v-model="form.platform" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="订单号" prop="order_no"><el-input v-model="form.order_no" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="商品名" prop="product_name"><el-input v-model="form.product_name" /></el-form-item></el-col>
        <el-col :span="12"><el-form-item label="SKU" prop="sku"><el-input v-model="form.sku" /></el-form-item></el-col>
        <el-col :span="6"><el-form-item label="数量" prop="quantity"><el-input-number v-model="form.quantity" :min="1" /></el-form-item></el-col>
        <el-col :span="6"><el-form-item label="总额" prop="total_amount"><el-input-number v-model="form.total_amount" :min="0" :step="0.01" /></el-form-item></el-col>
      </el-row>
      <el-form-item>
        <el-button type="primary" @click="onSubmit">提交</el-button>
        <el-button @click="onReset">重置</el-button>
      </el-form-item>
    </el-form>
    <el-alert v-if="msg" :title="msg" type="success" show-icon style="margin-top:10px;" />
  </el-card>
</template>
<script setup>
import { ref } from 'vue'
import axios from 'axios'
const emit = defineEmits(['refresh'])
const formRef = ref()
const form = ref({ country:'', platform:'', order_no:'', product_name:'', sku:'', quantity:1, total_amount:0 })
const rules = { country:[{required:true,message:'必填'}], platform:[{required:true,message:'必填'}], order_no:[{required:true,message:'必填'}], product_name:[{required:true,message:'必填'}], sku:[{required:true,message:'必填'}], quantity:[{required:true,type:'number',min:1,message:'必填'}], total_amount:[{required:true,type:'number',min:0,message:'必填'}] }
const msg = ref('')
async function onSubmit() {
  await formRef.value.validate()
  const res = await axios.post('/api/order', form.value).catch(e=>({data:{error:e.message}}))
  if(res.data && !res.data.error){
    msg.value = '添加成功！'
    emit('refresh')
    onReset()
  }else{
    msg.value = res.data.error || '添加失败'
  }
}
function onReset(){
  form.value = { country:'', platform:'', order_no:'', product_name:'', sku:'', quantity:1, total_amount:0 }
  msg.value = ''
}
</script>
