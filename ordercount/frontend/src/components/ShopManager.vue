<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>店铺管理</template>
    <!-- 只有超级管理员可以新增/修改店铺 -->
    <el-form v-if="isSuperAdmin" :model="form" label-width="100px" @submit.prevent style="margin-bottom:10px;">
      <el-row :gutter="10">
        <el-col :span="6">
          <el-form-item label="国家">
            <el-select v-model="form.country" placeholder="请选择国家" style="width:100%;">
              <el-option label="菲律宾" value="菲律宾" />
              <el-option label="印尼" value="印尼" />
              <el-option label="马来西亚" value="马来西亚" />
              <el-option label="其他" value="其他" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="所属平台">
            <el-select v-model="form.platform" placeholder="请选择平台" style="width:100%;">
              <el-option label="Shopee" value="Shopee" />
              <el-option label="Lazada" value="Lazada" />
              <el-option label="TikTok" value="TikTok" />
              <el-option label="其他" value="其他" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="店铺名称">
            <el-input v-model="form.name" placeholder="请输入店铺名称" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="登录账号">
            <el-input v-model="form.login_account" placeholder="店铺登录账号" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="登录密码">
            <el-input v-model="form.login_password" type="password" show-password placeholder="店铺登录密码" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="绑定手机号">
            <el-input v-model="form.phone" placeholder="店铺绑定手机号" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="绑定邮箱">
            <el-input v-model="form.email" placeholder="店铺绑定邮箱" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item>
        <el-button type="primary" @click="onSubmit">{{ form.id ? '保存修改' : '新增店铺' }}</el-button>
        <el-button @click="onReset">重置</el-button>
        <span v-if="msg" style="margin-left:10px;font-size:12px;" :style="{ color: msgOk ? '#67C23A' : '#F56C6C' }">{{ msg }}</span>
      </el-form-item>
    </el-form>

    <el-table :data="stores" size="small" border style="width:100%;">
      <el-table-column prop="country" label="国家" width="100" />
      <el-table-column prop="platform" label="平台" width="120" />
      <el-table-column prop="name" label="店铺名称" width="180" />
      <el-table-column prop="login_account" label="登录账号" width="180" />
      <el-table-column label="登录密码" width="160">
        <template #default="scope">
          <el-input v-model="scope.row.login_password" type="password" show-password size="small" disabled />
        </template>
      </el-table-column>
      <el-table-column prop="phone" label="绑定手机号" width="150" />
      <el-table-column prop="email" label="绑定邮箱" width="200" />
      <el-table-column prop="created_at" label="创建时间" width="180">
        <template #default="scope">
          {{ formatTime(scope.row.created_at) }}
        </template>
      </el-table-column>
      <el-table-column v-if="isSuperAdmin" label="操作" width="220">
        <template #default="scope">
          <el-button type="primary" link size="small" @click="onEdit(scope.row)">编辑</el-button>
          <el-button type="danger" link size="small" @click="onDelete(scope.row)">删除</el-button>
          <el-button type="warning" link size="small" @click="onOpenAuth(scope.row)">授权</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <!-- 店铺授权对话框，仅超级管理员使用 -->
  <el-dialog v-model="authDialogVisible" title="店铺授权" width="400px">
    <div v-if="authStore">
      <div style="margin-bottom:10px;">店铺：<b>{{ authStore.name }}</b></div>
      <el-form label-width="90px">
        <el-form-item label="可见用户">
          <el-select
            v-model="authUserIds"
            multiple
            filterable
            clearable
            placeholder="请选择可见的用户"
            style="width:100%;"
          >
            <el-option
              v-for="u in users"
              :key="u.id"
              :label="u.username + ' (' + u.role + ')'"
              :value="u.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
    </div>
    <template #footer>
      <el-button @click="authDialogVisible = false">取 消</el-button>
      <el-button type="primary" @click="onSaveAuth">保 存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import axios from 'axios'
import { ElMessageBox, ElMessage } from 'element-plus'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})

const stores = ref([])
// 国家默认印尼
const form = ref({ id: 0, platform: '', country: '印尼', name: '', login_account: '', login_password: '', phone: '', email: '' })
const msg = ref('')
const msgOk = ref(false)

const users = ref([])
const authDialogVisible = ref(false)
const authStore = ref(null)
const authUserIds = ref([])

const isSuperAdmin = computed(() => props.currentUser && props.currentUser.role === 'superadmin')

// 在输入绑定邮箱时，如果只输入了前缀、不包含 @，自动补全为 xxx@radiant-ec.com
watch(
  () => form.value.email,
  (val) => {
    const trimmed = (val || '').trim()
    if (!trimmed) return
    // 已经包含 @ 的情况，认为用户手动输入完整邮箱，不做改动
    if (trimmed.includes('@')) return
    form.value.email = `${trimmed}@radiant-ec.com`
  }
)

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

async function load() {
  try {
    const res = await axios.get('/api/shops')
    stores.value = res.data?.items || []
  } catch (e) {
    stores.value = []
  }
}

function onReset() {
  form.value = { id: 0, platform: '', country: '印尼', name: '', login_account: '', login_password: '', phone: '', email: '' }
  msg.value = ''
}

async function loadUsers() {
  if (!isSuperAdmin.value) return
  try {
    const res = await axios.get('/api/users')
    users.value = res.data || []
  } catch {
    users.value = []
  }
}

async function onSubmit() {
  msg.value = ''
  msgOk.value = false
  try {
    const payload = { ...form.value }
    const res = await axios.post('/api/shops', payload)
    msg.value = form.value.id ? '保存成功' : '新增成功'
    msgOk.value = true
    onReset()
    await load()
    return res
  } catch (e) {
    msg.value = e?.response?.data?.error || e?.message || '保存失败'
    msgOk.value = false
  }
}

function onEdit(row) {
  form.value = { ...row }
  msg.value = ''
}

async function onDelete(row) {
  if (!row || !row.id) return
  try {
    await ElMessageBox.confirm(`确定要删除店铺【${row.name}】吗？`, '提示', {
      type: 'warning',
    })
  } catch {
    return
  }
  try {
    await axios.delete(`/api/shops/${row.id}`)
    ElMessage.success('删除成功')
    await load()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || e?.message || '删除失败')
  }
}

async function onOpenAuth(row) {
  if (!row || !row.id) return
  authStore.value = row
  authUserIds.value = []
  authDialogVisible.value = true
  try {
    const res = await axios.get(`/api/shops/${row.id}/users`)
    authUserIds.value = res.data?.user_ids || []
  } catch {
    authUserIds.value = []
  }
}

async function onSaveAuth() {
  if (!authStore.value) return
  try {
    await axios.post(`/api/shops/${authStore.value.id}/users`, { user_ids: authUserIds.value })
    ElMessage.success('授权已保存')
    authDialogVisible.value = false
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || e?.message || '保存授权失败')
  }
}

onMounted(() => {
  load()
  loadUsers()
})
</script>

<style scoped>
</style>
