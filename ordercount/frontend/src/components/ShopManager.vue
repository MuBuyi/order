<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>店铺管理</template>
    <!-- 只有超级管理员可以新增/修改店铺 -->
    <el-form v-if="isSuperAdmin" :model="form" label-width="100px" @submit.prevent style="margin-bottom:10px;">
      <el-row :gutter="10">
        <el-col :span="6">
          <el-form-item label="国家">
            <el-select v-model="form.country" placeholder="请选择国家" style="width:100%;">
              <el-option
                v-for="c in countriesForForm"
                :key="c"
                :label="c"
                :value="c"
              />
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
        <el-col :span="6">
          <el-form-item label="备注">
            <el-input v-model="form.remark" placeholder="填写备注（可选）" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item>
        <el-button type="primary" @click="onSubmit">{{ form.id ? '保存修改' : '新增店铺' }}</el-button>
        <el-button @click="onReset">重置</el-button>
        <span v-if="msg" style="margin-left:10px;font-size:12px;" :style="{ color: msgOk ? '#67C23A' : '#F56C6C' }">{{ msg }}</span>
      </el-form-item>
    </el-form>

    <div style="margin:10px 0; display:flex; align-items:center; justify-content:flex-end; gap:12px; flex-wrap:wrap;">
      <div>按国家筛选：</div>
      <el-select v-model="selectedCountry" placeholder="全部国家" style="width:220px;">
        <el-option
          v-for="c in countries"
          :key="c"
          :label="c"
          :value="c"
        />
      </el-select>
      <div>按状态筛选：</div>
      <el-select v-model="selectedStatus" placeholder="全部状态" style="width:160px;">
        <el-option label="全部" value="all" />
        <el-option label="已启用" value="enabled" />
        <el-option label="已停用" value="disabled" />
      </el-select>
    </div>

    <div v-if="showEnabledSection" style="margin-bottom:14px;">
      <div style="font-size:13px; color:#606266; margin:4px 0 8px;">已启用店铺（{{ filteredEnabledStores.length }}）</div>
      <el-table :data="filteredEnabledStores" size="small" border table-layout="auto" class="shop-table" style="width:100%;">
        <el-table-column prop="country" label="国家" min-width="90" show-overflow-tooltip />
        <el-table-column prop="platform" label="平台" min-width="90" show-overflow-tooltip />
        <el-table-column prop="name" label="店铺名称" min-width="100" show-overflow-tooltip />
        <el-table-column prop="login_account" label="登录账号" min-width="120" show-overflow-tooltip />
        <el-table-column v-if="isSuperAdmin" label="登录密码" min-width="120" show-overflow-tooltip>
          <template #default="scope">
            <span>{{ scope.row.login_password || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="绑定手机号" min-width="110" show-overflow-tooltip />
        <el-table-column prop="email" label="绑定邮箱" min-width="140" show-overflow-tooltip />
        <el-table-column label="负责人" min-width="120" show-overflow-tooltip>
          <template #default="scope">
            {{ ownerText(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" min-width="140" show-overflow-tooltip>
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column v-if="isSuperAdmin" label="状态" min-width="80">
          <template #default="scope">
            <el-switch
              v-model="scope.row.is_blocked"
              :active-value="false"
              :inactive-value="true"
              @change="onToggleBlocked(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column v-if="isSuperAdmin" label="操作" min-width="150">
          <template #default="scope">
            <div class="op-actions">
              <el-button type="primary" link size="small" @click="onEdit(scope.row)">编辑</el-button>
              <el-button type="danger" link size="small" @click="onDelete(scope.row)">删除</el-button>
              <el-button type="warning" link size="small" @click="onOpenAuth(scope.row)">授权</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="showDisabledSection">
      <div style="font-size:13px; color:#606266; margin:4px 0 8px;">已停用店铺（{{ filteredDisabledStores.length }}）</div>
      <el-table :data="filteredDisabledStores" size="small" border table-layout="auto" class="shop-table" style="width:100%;">
        <el-table-column prop="country" label="国家" min-width="90" show-overflow-tooltip />
        <el-table-column prop="platform" label="平台" min-width="90" show-overflow-tooltip />
        <el-table-column prop="name" label="店铺名称" min-width="100" show-overflow-tooltip />
        <el-table-column prop="login_account" label="登录账号" min-width="120" show-overflow-tooltip />
        <el-table-column v-if="isSuperAdmin" label="登录密码" min-width="120" show-overflow-tooltip>
          <template #default="scope">
            <span>{{ scope.row.login_password || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="phone" label="绑定手机号" min-width="110" show-overflow-tooltip />
        <el-table-column prop="email" label="绑定邮箱" min-width="140" show-overflow-tooltip />
        <el-table-column label="负责人" min-width="120" show-overflow-tooltip>
          <template #default="scope">
            {{ ownerText(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column prop="remark" label="备注" min-width="120" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" min-width="140" show-overflow-tooltip>
          <template #default="scope">
            {{ formatTime(scope.row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column v-if="isSuperAdmin" label="状态" min-width="80">
          <template #default="scope">
            <el-switch
              v-model="scope.row.is_blocked"
              :active-value="false"
              :inactive-value="true"
              @change="onToggleBlocked(scope.row)"
            />
          </template>
        </el-table-column>
        <el-table-column v-if="isSuperAdmin" label="操作" min-width="150">
          <template #default="scope">
            <div class="op-actions">
              <el-button type="primary" link size="small" @click="onEdit(scope.row)">编辑</el-button>
              <el-button type="danger" link size="small" @click="onDelete(scope.row)">删除</el-button>
              <el-button type="warning" link size="small" @click="onOpenAuth(scope.row)">授权</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-empty v-if="!showEnabledSection && !showDisabledSection" description="暂无符合筛选条件的店铺" />
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
const allStores = ref([])
// 国家默认印尼，is_blocked 默认 false（未封禁）
const form = ref({ id: 0, platform: '', country: '印尼', name: '', login_account: '', login_password: '', phone: '', email: '', remark: '', is_blocked: false })
const msg = ref('')
const msgOk = ref(false)

const users = ref([])
const authDialogVisible = ref(false)
const authStore = ref(null)
const authUserIds = ref([])

const isSuperAdmin = computed(() => props.currentUser && props.currentUser.role === 'superadmin')

// 国家筛选与动态国家列表
const selectedCountry = ref('印尼')
const selectedStatus = ref('enabled')
const countries = computed(() => {
  const set = new Set()
  allStores.value.forEach((s) => { if (s && s.country) set.add(s.country) })
  // Ensure commonly used countries are present in the list
  const required = ['菲律宾', '印尼', '马来西亚']
  required.forEach(r => set.add(r))
  const arr = Array.from(set).sort()
  if (arr.length === 0) return ['菲律宾', '印尼', '马来西亚', '其他']
  return ['全部', ...arr]
})

const countriesForForm = computed(() => {
  const arr = countries.value.filter((c) => c !== '全部')
  return arr.length ? arr : ['菲律宾', '印尼', '马来西亚', '其他']
})

const filteredStores = computed(() => {
  return (stores.value || []).slice().sort((a, b) => {
    const an = (a && a.name) ? String(a.name) : ''
    const bn = (b && b.name) ? String(b.name) : ''
    return an.localeCompare(bn, 'zh-CN', { numeric: true })
  })
})

const filteredEnabledStores = computed(() => {
  const base = filteredStores.value || []
  if (selectedStatus.value === 'disabled') return []
  return base.filter((s) => !s?.is_blocked)
})

const filteredDisabledStores = computed(() => {
  const base = filteredStores.value || []
  if (selectedStatus.value === 'enabled') return []
  return base.filter((s) => Boolean(s?.is_blocked))
})

const showEnabledSection = computed(() => filteredEnabledStores.value.length > 0)
const showDisabledSection = computed(() => filteredDisabledStores.value.length > 0)

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


function ownerText(row) {
  const names = Array.isArray(row?.user_names) ? row.user_names.filter(Boolean) : []
  if (names.length > 0) return names.join("、")
  const ids = Array.isArray(row?.user_ids) ? row.user_ids.filter(Boolean) : []
  if (ids.length === 0) return "未授权"
  const userMap = new Map((users.value || []).map((u) => [u.id, u.username]))
  return ids.map((id) => userMap.get(id) || `用户${id}`).join("、")
}

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  if (Number.isNaN(d.getTime())) return t
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
}

function normalizeStores(items) {
  const mapped = (items || []).map((s) => ({
    ...s,
    is_blocked: s === null || s === undefined ? false : (s.is_blocked === true || s.is_blocked === 'true' || s.is_blocked === 1 || s.is_blocked === '1')
  }))
  mapped.sort((a, b) => {
    const an = (a && a.name) ? String(a.name) : ''
    const bn = (b && b.name) ? String(b.name) : ''
    return an.localeCompare(bn, 'zh-CN', { numeric: true })
  })
  return mapped
}

function buildShopParams() {
  const params = {}
  if (selectedCountry.value && selectedCountry.value !== '全部') {
    params.country = selectedCountry.value
  }
  if (selectedStatus.value && selectedStatus.value !== 'all') {
    params.status = selectedStatus.value
  }
  return params
}

async function loadAllStores() {
  try {
    const res = await axios.get('/api/shops')
    allStores.value = normalizeStores(res.data?.items || [])
  } catch (e) {
    allStores.value = []
  }
}

async function load() {
  try {
    const res = await axios.get('/api/shops', { params: buildShopParams() })
    stores.value = normalizeStores(res.data?.items || [])
  } catch (e) {
    stores.value = []
  }
}

watch([selectedCountry, selectedStatus], () => {
  load()
})

function onReset() {
  form.value = { id: 0, platform: '', country: '印尼', name: '', login_account: '', login_password: '', phone: '', email: '', remark: '', is_blocked: false }
  msg.value = ''
}

async function loadUsers() {
  if (!isSuperAdmin.value) return
  try {
    const res = await axios.get('/api/user-options')
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
    await loadAllStores()
    await load()
    return res
  } catch (e) {
    msg.value = e?.response?.data?.error || e?.message || '保存失败'
    msgOk.value = false
  }
}
 
async function onToggleBlocked(row) {
  if (!row || !row.id) return
  try {
    await axios.post('/api/shops', { ...row, is_blocked: Boolean(row.is_blocked) })
    await loadAllStores()
    await load()
  } catch (e) {
    const status = e?.response?.status
    if (status === 401 || status === 403) {
      ElMessage.error('无操作权限，请以超级管理员登录')
    } else {
      ElMessage.error(e?.response?.data?.error || e?.message || '操作失败')
    }
    // 回滚状态
    row.is_blocked = !row.is_blocked
  }
}

function onEdit(row) {
  form.value = { ...row, is_blocked: (row.is_blocked === true || row.is_blocked === 'true' || row.is_blocked === 1 || row.is_blocked === '1') }
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
    await loadAllStores()
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
    await loadAllStores()
    await load()
  } catch (e) {
    ElMessage.error(e?.response?.data?.error || e?.message || '保存授权失败')
  }
}

onMounted(() => {
  loadAllStores()
  load()
  loadUsers()
})
</script>

<style scoped>
.shop-table {
  width: 100%;
}

.op-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 2px 6px;
}

@media (max-width: 992px) {
  .shop-table :deep(.el-table__cell) {
    padding: 6px 4px;
    font-size: 12px;
  }

  .shop-table :deep(.el-button--small) {
    padding: 2px 4px;
  }
}

@media (max-width: 768px) {
  .shop-table :deep(.cell) {
    white-space: normal;
    word-break: break-word;
    line-height: 1.35;
  }
}
</style>
