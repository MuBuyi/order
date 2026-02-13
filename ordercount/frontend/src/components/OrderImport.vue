<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>Excel 批量导入</template>
    <el-upload
      class="upload-demo"
      drag
      action="/api/orders/import"
      :show-file-list="false"
      :on-success="onSuccess"
      :on-error="onError"
      accept=".xlsx,.xls"
    >
      <el-icon><upload-filled /></el-icon>
      <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
      <template #tip>
        <div class="el-upload__tip">仅支持 .xlsx/.xls，表头需包含：国家、平台、订单号、商品名、SKU、数量、总额</div>
      </template>
    </el-upload>
    <el-alert v-if="msg" :title="msg" type="success" show-icon style="margin-top:10px;" />
  </el-card>
</template>
<script setup>
import { ref } from 'vue'
import { UploadFilled } from '@element-plus/icons-vue'
const emit = defineEmits(['refresh'])
const msg = ref('')
function onSuccess(res){
  msg.value = `成功导入 ${res.created||0} 条订单！`
  emit('refresh')
}
function onError(err){
  msg.value = '导入失败：' + (err.message||err)
}
</script>
