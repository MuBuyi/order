<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>今日站点出单详情</template>
    <el-form :model="form" ref="formRef" label-width="90px" @submit.prevent>
      <el-row :gutter="10">
        <!-- 第一行：国家 -->
        <el-col :span="12">
          <el-form-item label="国家">
            <el-select v-model="form.country" placeholder="请选择国家">
              <el-option label="菲律宾" value="菲律宾" />
              <el-option label="印尼" value="印尼" />
              <el-option label="马来西亚" value="马来西亚" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12" />

        <!-- 第二行：商品SKU + 数量（数量紧跟在 SKU 后面） -->
        <el-col :span="12">
          <el-form-item label="商品SKU">
            <el-select
              v-model="form.sku"
              placeholder="请选择商品SKU"
              filterable
              @change="onSkuChange"
            >
              <el-option
                v-for="p in products"
                :key="p.id"
                :label="p.sku"
                :value="p.sku"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="数量">
            <el-input-number v-model="form.quantity" :min="1" />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 操作按钮：按 SKU 多次提交出单明细 -->
      <el-form-item>
        <el-button type="primary" @click="onSubmitDetail">提交出单明细</el-button>
        <el-button @click="onReset">重置</el-button>
      </el-form-item>

      <!-- 总额：在录完所有 SKU 之后再填写并单独保存 -->
      <el-form-item label="今日总额">
        <el-input-number v-model="form.total_amount" :min="0" :step="0.01" />
        <el-button type="success" style="margin-left:10px;" @click="onSubmitTotal">
          保存今日总额
        </el-button>
      </el-form-item>
    </el-form>

    <el-alert v-if="msg" :title="msg" type="success" show-icon style="margin-top:10px;" />

    <!-- 本次会话提交的出单详情列表 -->
    <el-divider content-position="left" style="margin-top:20px;">今日已提交出单记录（当前页面）</el-divider>
    <el-table
      v-if="submittedOrders.length"
      :data="submittedOrders"
      size="small"
      style="margin-top:10px;"
      border
    >
      <el-table-column prop="created_at" label="时间" width="180" />
      <el-table-column prop="country" label="国家" width="80" />
      <el-table-column prop="sku" label="商品SKU">
        <template #default="scope">
          <span v-if="editingId !== scope.row.id">{{ scope.row.sku }}</span>
          <el-select
            v-else
            v-model="editForm.sku"
            placeholder="请选择商品SKU"
            size="small"
            style="width: 160px;"
            filterable
          >
            <el-option
              v-for="p in products"
              :key="p.id"
              :label="p.sku"
              :value="p.sku"
            />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column prop="product_name" label="商品名称" />
      <el-table-column prop="quantity" label="数量" width="80">
        <template #default="scope">
          <span v-if="editingId !== scope.row.id">{{ scope.row.quantity }}</span>
          <el-input-number
            v-else
            v-model="editForm.quantity"
            :min="0"
            size="small"
          />
        </template>
      </el-table-column>
      <el-table-column prop="total_amount" label="总额" width="100" />
      <el-table-column label="操作" width="160">
        <template #default="scope">
          <template v-if="editingId === scope.row.id">
            <el-button type="primary" size="small" @click="onSaveEdit(scope.row)">
              保存
            </el-button>
            <el-button size="small" @click="onCancelEdit" style="margin-left:8px;">
              取消
            </el-button>
          </template>
          <template v-else>
            <el-button type="primary" size="small" @click="onStartEdit(scope.row)">
              修改
            </el-button>
            <el-popconfirm
              title="确定要删除这条出单记录吗？"
              confirm-button-text="删除"
              cancel-button-text="取消"
              @confirm="onDelete(scope.row)"
            >
              <template #reference>
                <el-button type="danger" size="small" style="margin-left:8px;">
                  删除
                </el-button>
              </template>
            </el-popconfirm>
          </template>
        </template>
      </el-table-column>
    </el-table>
    <el-empty v-else description="暂无本次会话出单记录" style="margin-top:10px;" />

    <!-- 按商品名汇总的数量统计 -->
    <el-divider content-position="left" style="margin-top:20px;">今日商品汇总（按商品名合计数量）</el-divider>
    <el-table
      v-if="productSummary.length"
      :data="productSummary"
      size="small"
      style="margin-top:10px;"
      border
    >
      <el-table-column prop="product_name" label="商品名称" />
      <el-table-column prop="quantity" label="总数量" width="100" />
    </el-table>
    <el-empty v-else description="暂无可汇总的商品" style="margin-top:10px;" />
  </el-card>
</template>
<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
const emit = defineEmits(['refresh'])
const formRef = ref()
// 平台、订单号、SKU 文本输入去掉，保留 SKU 下拉与数量，总额改为汇总后单独填写
const form = ref({ country: '印尼', product_name: '', sku: '', quantity: 1, total_amount: 0 })
const msg = ref('')
const products = ref([])
const submittedOrders = ref([])
const editingId = ref(null)
const editForm = ref({ sku: '', quantity: 0 })

// 按商品名汇总本次会话中各商品的总数量（不含“今日总额汇总”等数量为 0 的记录）
const productSummary = computed(() => {
  const map = {}
  for (const o of submittedOrders.value) {
    if (!o || !o.product_name) continue
    const qty = Number(o.quantity) || 0
    if (qty <= 0) continue
    const key = o.product_name
    if (!map[key]) {
      map[key] = { product_name: key, quantity: 0 }
    }
    map[key].quantity += qty
  }
  return Object.values(map)
})

async function loadProducts() {
  const res = await axios.get('/api/products')
  products.value = res.data || []
}

// 加载当天所有订单记录，用于“今日已提交出单记录”列表
async function loadTodayOrders() {
  try {
    const res = await axios.get('/api/orders')
    submittedOrders.value = res.data?.items || []
  } catch (e) {
    submittedOrders.value = []
  }
}

function onSkuChange(val) {
  const p = products.value.find(x => x.sku === val)
  form.value.product_name = p ? p.name : ''
}

onMounted(() => {
  loadProducts()
  loadTodayOrders()
})
// 提交单条 SKU 出单明细：不需要填写总额，总额统一在最后单独保存
async function onSubmitDetail() {
  // 手动校验，避免表单出现红色高亮
  if (!form.value.country) {
    msg.value = '请先选择国家'
    return
  }
  if (!form.value.sku) {
    msg.value = '请选择商品SKU'
    return
  }
  if (!form.value.quantity || form.value.quantity < 1) {
    msg.value = '请填写数量（至少 1）'
    return
  }
  const payload = {
    country: form.value.country,
    product_name: form.value.product_name,
    sku: form.value.sku,
    quantity: form.value.quantity,
    // 每条 SKU 明细不再录入总额，这里固定为 0
    total_amount: 0,
  }
  const res = await axios.post('/api/order', payload).catch(e=>({data:{error:e.message}}))
  if(res.data && !res.data.error){
    msg.value = '出单明细添加成功！'
    // 重新加载当天订单列表，确保与订单记录一致
    await loadTodayOrders()
    emit('refresh')
    // 只重置 SKU 与数量，国家与总额保留，方便连续录入
    form.value.sku = ''
    form.value.product_name = ''
    form.value.quantity = 1
  }else{
    msg.value = res.data.error || '添加失败'
  }
}

// 保存今日总额：在录完所有 SKU 之后，单独提交一条汇总记录
async function onSubmitTotal() {
  if (!form.value.total_amount || form.value.total_amount <= 0) {
    msg.value = '请先填写大于 0 的今日总额'
    return
  }
  // 至少需要国家信息，用于后续统计
  if (!form.value.country) {
    msg.value = '请先选择国家'
    return
  }
  const payload = {
    country: form.value.country,
    product_name: '今日总额汇总',
    sku: '',
    quantity: 0,
    total_amount: form.value.total_amount,
  }
  const res = await axios.post('/api/order', payload).catch(e=>({data:{error:e.message}}))
  if(res.data && !res.data.error){
    msg.value = '今日总额保存成功！'
    await loadTodayOrders()
    emit('refresh')
  }else{
    msg.value = res.data.error || '保存今日总额失败'
  }
}

function onReset(){
  form.value = { country:'印尼', product_name:'', sku:'', quantity:1, total_amount:0 }
  msg.value = ''
}

// 编辑、删除已提交订单
async function onDelete(row) {
  if (!row || !row.id) return
  const res = await axios.delete(`/api/orders/${row.id}`).catch(e => ({ data: { error: e.message } }))
  if (res.data && !res.data.error) {
    await loadTodayOrders()
    msg.value = '删除成功'
    emit('refresh')
  } else {
    msg.value = res.data.error || '删除失败'
  }
}

function onStartEdit(row) {
  if (!row || !row.id) return
  editingId.value = row.id
  editForm.value = {
    sku: row.sku || '',
    quantity: row.quantity ?? 0,
  }
}

function onCancelEdit() {
  editingId.value = null
  editForm.value = { sku: '', quantity: 0 }
}

async function onSaveEdit(row) {
  if (!row || !row.id) return
  const trimmedSku = (editForm.value.sku || '').trim()
  if (!trimmedSku) {
    msg.value = '商品SKU不能为空'
    return
  }
  const qtyNum = Number(editForm.value.quantity)
  if (!Number.isFinite(qtyNum) || qtyNum < 0) {
    msg.value = '数量必须是大于等于 0 的数字'
    return
  }
  const payload = { sku: trimmedSku, quantity: qtyNum }
  const res = await axios.put(`/api/orders/${row.id}`, payload).catch(e => ({ data: { error: e.message } }))
  if (res.data && !res.data.error) {
    await loadTodayOrders()
    msg.value = '修改成功'
    editingId.value = null
    editForm.value = { sku: '', quantity: 0 }
    emit('refresh')
  } else {
    msg.value = res.data.error || '修改失败'
  }
}
</script>
