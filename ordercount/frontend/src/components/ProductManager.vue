<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>商品管理</template>
    <el-form
      v-if="canCreateProducts"
      :model="form"
      label-width="90px"
      @submit.prevent
      style="margin-bottom:10px;"
    >
      <el-row :gutter="10">
        <el-col :span="6">
          <el-form-item label="商品SKU">
            <el-input v-model="form.sku" placeholder="SKU" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="商品名称">
            <el-select v-model="form.name" placeholder="请选择商品名称" style="width:100%;">
              <el-option label="固态硬盘" value="固态硬盘" />
              <el-option label="机械硬盘" value="机械硬盘" />
              <el-option label="U盘" value="U盘" />
              <el-option label="内存卡" value="内存卡" />
            </el-select>
          </el-form-item>
        </el-col>
        <!-- 超级管理员：可以为不同角色配置不同的成本 -->
        <template v-if="isSuperAdmin">
          <el-col :span="6">
            <el-form-item label="基础成本">
              <el-input
                v-model="form.cost"
                placeholder="超管视角的基础成本"
                type="number"
                step="0.01"
                min="0"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="管理员成本">
              <el-input
                v-model="form.cost_admin"
                placeholder="管理员看到的成本"
                type="number"
                step="0.01"
                min="0"
              />
            </el-form-item>
          </el-col>
          <el-col :span="6">
            <el-form-item label="员工成本">
              <el-input
                v-model="form.cost_staff"
                placeholder="员工看到的成本"
                type="number"
                step="0.01"
                min="0"
              />
            </el-form-item>
          </el-col>
        </template>
        <!-- 管理员：只维护自己角色对应的成本 -->
        <el-col v-else-if="canEditCost" :span="6">
          <el-form-item label="成本">
            <el-input
              v-model="form.cost"
              placeholder="请输入成本"
              type="number"
              step="0.01"
              min="0"
            />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="商品图片">
            <el-upload
              class="product-uploader"
              action="/api/products/upload"
              :headers="uploadHeaders"
              :show-file-list="false"
              :on-success="onUploadSuccess"
              :on-error="onUploadError"
              accept="image/*"
            >
              <img v-if="form.image_url" :src="form.image_url" class="product-image" />
              <el-icon v-else class="product-uploader-icon"><Plus /></el-icon>
            </el-upload>
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item>
        <el-button type="primary" @click="onSubmit">{{ form.id ? '保存修改' : '新增商品' }}</el-button>
        <el-button @click="onReset">重置</el-button>
        <span v-if="msg" style="margin-left:10px;font-size:12px;" :style="{color: msgOk ? '#67C23A' : '#F56C6C'}">{{ msg }}</span>
      </el-form-item>
    </el-form>

    <el-table :data="products" size="small" border style="width:100%;">
      <el-table-column label="图片" width="100">
        <template #default="scope">
          <el-popover
            v-if="scope.row.image_url"
            placement="right"
            trigger="hover"
            popper-class="image-popover"
          >
            <template #reference>
              <img :src="scope.row.image_url" alt="thumb" class="product-thumb" />
            </template>
            <img :src="scope.row.image_url" alt="preview" class="product-preview" />
          </el-popover>
        </template>
      </el-table-column>
      <el-table-column prop="sku" label="SKU" width="160">
        <template #default="scope">
          <span v-if="!canInlineEdit || editingProductId !== scope.row.id">{{ scope.row.sku }}</span>
          <el-input
            v-else
            v-model="editingProduct.sku"
            size="small"
          />
        </template>
      </el-table-column>
      <el-table-column prop="name" label="商品名称">
        <template #default="scope">
          <span v-if="!canInlineEdit || editingProductId !== scope.row.id">{{ scope.row.name }}</span>
          <el-select
            v-else
            v-model="editingProduct.name"
            placeholder="请选择商品名称"
            size="small"
            style="width: 140px;"
          >
            <el-option label="固态硬盘" value="固态硬盘" />
            <el-option label="机械硬盘" value="机械硬盘" />
            <el-option label="U盘" value="U盘" />
            <el-option label="内存卡" value="内存卡" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column v-if="canViewCost" prop="cost" label="成本" width="180">
        <template #default="scope">
          <span v-if="!canInlineEdit || editingProductId !== scope.row.id">
            <span v-if="scope.row.cost != null">{{ Number(scope.row.cost).toFixed(2) }}</span>
          </span>
          <template v-else>
            <!-- 超级管理员：行内同时编辑三种角色成本 -->
            <div v-if="isSuperAdmin" class="inline-costs">
              <div class="inline-cost-row">
                <span class="inline-cost-label">基础</span>
                <el-input
                  v-model="editingProduct.cost"
                  size="small"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="基础成本"
                />
              </div>
              <div class="inline-cost-row">
                <span class="inline-cost-label">管理员</span>
                <el-input
                  v-model="editingProduct.cost_admin"
                  size="small"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="管理员成本"
                />
              </div>
              <div class="inline-cost-row">
                <span class="inline-cost-label">员工</span>
                <el-input
                  v-model="editingProduct.cost_staff"
                  size="small"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="员工成本"
                />
              </div>
            </div>
            <!-- 管理员：只编辑自己角色对应的成本 -->
            <el-input
              v-else
              v-model="editingProduct.cost"
              size="small"
              type="number"
              step="0.01"
              min="0"
            />
          </template>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="scope">
          {{ formatTime(scope.row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column v-if="canManageProducts" label="操作" width="180">
        <template #default="scope">
          <template v-if="canInlineEdit && editingProductId === scope.row.id">
            <el-button type="primary" link size="small" @click="onInlineSave(scope.row)">保存</el-button>
            <el-button link size="small" @click="onInlineCancel">取消</el-button>
          </template>
          <template v-else>
            <el-button type="primary" link size="small" @click="onEditClick(scope.row)">编辑</el-button>
            <el-button
              v-if="isSuperAdmin"
              type="danger"
              link
              size="small"
              @click="onDelete(scope.row)"
            >删除</el-button>
          </template>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'
import { Plus } from '@element-plus/icons-vue'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})

const products = ref([])
// 表单同时预留多角色成本字段：
// - cost:      超级管理员视角的基础成本
// - cost_admin: 管理员看到/使用的成本
// - cost_staff: 员工看到/使用的成本
const form = ref({ id: 0, sku: '', name: '', image_url: '', cost: '', cost_admin: '', cost_staff: '' })
const msg = ref('')
const msgOk = ref(false)
const editingProductId = ref(null)
const editingProduct = ref({ id: 0, sku: '', name: '', cost: '', cost_admin: '', cost_staff: '' })

const TOKEN_KEY = 'ordercount-token'

const uploadHeaders = computed(() => {
  if (typeof window === 'undefined') return {}
  const token = window.localStorage.getItem(TOKEN_KEY)
  return token ? { Authorization: `Bearer ${token}` } : {}
})

const isSuperAdmin = computed(() => props.currentUser && props.currentUser.role === 'superadmin')
const isAdmin = computed(() => props.currentUser && props.currentUser.role === 'admin')
const isStaff = computed(() => props.currentUser && props.currentUser.role === 'staff')
// 所有角色都能看到“自己角色对应的成本”
const canViewCost = computed(() => true)
// 只有管理员或超级管理员可以维护成本
const canEditCost = computed(() => isSuperAdmin.value || isAdmin.value)
// 仅超级管理员和管理员可以新增商品
const canCreateProducts = computed(() => isSuperAdmin.value || isAdmin.value)
// 仅超级管理员和管理员可以在表格中看到“操作”列（编辑）
const canManageProducts = computed(() => isSuperAdmin.value || isAdmin.value)
// 管理员和超级管理员在表格中使用同行编辑
const canInlineEdit = computed(() => isSuperAdmin.value || isAdmin.value)

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  const res = await axios.get('/api/products')
  products.value = res.data || []
}

function onUploadSuccess(res) {
  if (res && res.url) {
    form.value.image_url = res.url
    msgOk.value = true
    msg.value = '图片上传成功'
  } else {
    msgOk.value = false
    msg.value = '图片上传返回异常'
  }
}

function onUploadError(err) {
  msgOk.value = false
  msg.value = (err && err.message) ? `图片上传失败：${err.message}` : '图片上传失败'
}

async function onSubmit() {
  msg.value = ''
  msgOk.value = false
  // 前端基础校验：SKU 和 名称不能都为空
  const sku = (form.value.sku || '').trim()
  const name = (form.value.name || '').trim()
  if (!sku && !name) {
    msg.value = '请先填写商品SKU或选择商品名称'
    msgOk.value = false
    return
  }
  try {
    const payload = {
      id: form.value.id,
      sku: form.value.sku,
      name: form.value.name,
      image_url: form.value.image_url,
      // cost 始终表示“当前角色看到/维护的成本”；
      // 对于超级管理员，它是基础成本；对于管理员，它是管理员成本。
      cost: form.value.cost ? Number(form.value.cost) : 0,
    }

    // 超级管理员可以一次性配置多角色成本
    if (isSuperAdmin.value) {
      payload.cost_admin = form.value.cost_admin ? Number(form.value.cost_admin) : 0
      payload.cost_staff = form.value.cost_staff ? Number(form.value.cost_staff) : 0
    }
    const res = await axios.post('/api/products', payload)
    if (res.data && !res.data.error) {
      msgOk.value = true
      msg.value = '保存成功'
      onReset()
      load()
    } else {
      msg.value = res.data.error || '保存失败'
    }
  } catch (e) {
    // 优先显示后端返回的业务错误，其次显示通用错误
    if (e && e.response && e.response.data && e.response.data.error) {
      msg.value = e.response.data.error
    } else {
      msg.value = '保存失败，请稍后重试'
    }
  }
}

function onReset() {
  form.value = { id: 0, sku: '', name: '', image_url: '', cost: '', cost_admin: '', cost_staff: '' }
}

function onEdit(row) {
  form.value.id = row.id
  form.value.sku = row.sku
  form.value.name = row.product_name || row.name
  form.value.image_url = row.image_url || ''
  // 对于非超级管理员，这里的 cost 已经是“当前角色可见的成本”
  form.value.cost = row.cost != null ? String(row.cost) : ''
  if (isSuperAdmin.value) {
    form.value.cost_admin = row.cost_admin != null ? String(row.cost_admin) : ''
    form.value.cost_staff = row.cost_staff != null ? String(row.cost_staff) : ''
  } else {
    form.value.cost_admin = ''
    form.value.cost_staff = ''
  }
}

async function onDelete(row) {
  try {
    await axios.delete(`/api/products/${row.id}`)
    load()
  } catch (e) {
    // ignore
  }
}

function onEditClick(row) {
  // 管理员和超级管理员使用表格行内编辑；其他角色沿用顶部表单编辑
  if (canInlineEdit.value) {
    editingProductId.value = row.id
    editingProduct.value = {
      id: row.id,
      sku: row.sku || '',
      name: row.name || row.product_name || '',
      cost: row.cost != null ? String(row.cost) : '',
      cost_admin: row.cost_admin != null ? String(row.cost_admin) : '',
      cost_staff: row.cost_staff != null ? String(row.cost_staff) : '',
    }
    msg.value = ''
  } else {
    onEdit(row)
  }
}

function onInlineCancel() {
  editingProductId.value = null
  editingProduct.value = { id: 0, sku: '', name: '', cost: '', cost_admin: '', cost_staff: '' }
}

async function onInlineSave(row) {
  if (!row || !row.id) return
  msg.value = ''
  msgOk.value = false
  const sku = (editingProduct.value.sku || '').trim()
  const name = (editingProduct.value.name || '').trim()
  if (!sku && !name) {
    msg.value = '请先填写商品SKU或选择商品名称'
    return
  }
  try {
    const payload = {
      id: editingProduct.value.id,
      sku: editingProduct.value.sku,
      name: editingProduct.value.name,
      image_url: row.image_url || '',
      cost: editingProduct.value.cost ? Number(editingProduct.value.cost) : 0,
    }
    if (isSuperAdmin.value) {
      // 超级管理员可以同时调整管理员/员工成本
      payload.cost_admin = editingProduct.value.cost_admin ? Number(editingProduct.value.cost_admin) : 0
      payload.cost_staff = editingProduct.value.cost_staff ? Number(editingProduct.value.cost_staff) : 0
    }
    const res = await axios.post('/api/products', payload)
    if (res.data && !res.data.error) {
      msgOk.value = true
      msg.value = '保存成功'
      editingProductId.value = null
      editingProduct.value = { id: 0, sku: '', name: '', cost: '', cost_admin: '', cost_staff: '' }
      load()
    } else {
      msg.value = res.data.error || '保存失败'
    }
  } catch (e) {
    if (e && e.response && e.response.data && e.response.data.error) {
      msg.value = e.response.data.error
    } else {
      msg.value = '保存失败，请稍后重试'
    }
  }
}

onMounted(load)
</script>

<style scoped>
.product-uploader {
  display: inline-block;
}
.product-image {
  width: 30px;
  height: 30px;
  object-fit: cover;
}
.product-uploader-icon {
  font-size: 32px;
  color: #8c939d;
  width: 30px;
  height: 30px;
  border: 1px dashed #d9d9d9;
  display: flex;
  align-items: center;
  justify-content: center;
}

.product-thumb {
  width: 30px;
  height: 30px;
  object-fit: cover;
}

.product-preview {
  max-width: 260px;
  max-height: 260px;
  object-fit: contain;
}

.inline-costs {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.inline-cost-row {
  display: flex;
  align-items: center;
}

.inline-cost-label {
  width: 50px;
  font-size: 12px;
  color: #606266;
  text-align: right;
  margin-right: 4px;
}

/* 去掉图片悬浮弹窗的白底、边框和阴影，仅保留图片 */
:deep(.image-popover),
:deep(.image-popover.el-popover),
:deep(.image-popover .el-popover__content) {
  background-color: transparent !important;
  border: none !important;
  box-shadow: none !important;
  padding: 0 !important;
  border-radius: 0 !important;
}

/* 隐藏 Popover 的小箭头 */
:deep(.image-popover .el-popper__arrow) {
  display: none !important;
}
</style>
