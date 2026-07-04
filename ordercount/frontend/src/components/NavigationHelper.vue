<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center;gap:10px;flex-wrap:wrap;">
        <span>导航助手</span>
        <span style="font-size:12px;color:#909399;">可保存常用网站及账号信息，点击一键跳转</span>
      </div>
    </template>

    <!-- 快速导航：先选类别，再从下拉中选站点直接打开 -->
    <div style="margin-bottom:10px;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <span style="font-size:12px;color:#606266;">快速导航：</span>
      <el-select
        v-model="quickCategory"
        size="small"
        placeholder="选择类别"
        style="width:160px;"
        clearable
      >
        <el-option
          v-for="c in categoryOptions"
          :key="c"
          :label="c"
          :value="c"
        />
      </el-select>
      <el-select
        v-model="quickLinkId"
        size="small"
        placeholder="选择站点"
        style="width:220px;"
        :disabled="!quickCategory"
        @change="onQuickOpen"
        clearable
      >
        <el-option
          v-for="link in quickLinks"
          :key="link.id"
          :label="link.title + (link.account ? '（' + link.account + '）' : '')"
          :value="link.id"
        />
      </el-select>
    </div>

    <el-form :model="form" label-width="70px" size="small" @submit.prevent>
      <el-row :gutter="10">
        <el-col :span="6">
          <el-form-item label="类别">
            <el-select
              v-model="form.category"
              filterable
              allow-create
              default-first-option
              placeholder="例如：电商后台/工具"
              style="width:100%;"
            >
              <el-option
                v-for="c in categoryOptions"
                :key="c"
                :label="c"
                :value="c"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="名称">
            <el-input v-model="form.title" placeholder="例如：Shopee 后台" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="网址">
            <el-input v-model="form.url" placeholder="例如：https://xxx" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="账号">
            <el-input v-model="form.account" placeholder="登录账号或备注" />
          </el-form-item>
        </el-col>
      </el-row>
      <el-form-item label="备注">
        <el-input v-model="form.remark" placeholder="可填写用途说明，不建议填写密码" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" size="small" @click="onSave">{{ form.id ? '保存修改' : '添加导航' }}</el-button>
        <el-button size="small" @click="onReset">重置</el-button>
      </el-form-item>
    </el-form>

    <div style="margin:10px 0;display:flex;align-items:center;gap:10px;flex-wrap:wrap;">
      <span style="font-size:12px;color:#606266;">按类别查看：</span>
      <el-select
        v-model="filterCategory"
        clearable
        size="small"
        placeholder="全部类别"
        style="width:160px;"
      >
        <el-option
          v-for="c in categoryOptions"
          :key="c"
          :label="c"
          :value="c"
        />
      </el-select>
    </div>

    <el-table :data="filteredItems" size="small" border style="margin-top:10px;" v-loading="loading">
      <el-table-column type="index" label="#" width="50" />
      <el-table-column prop="title" label="名称" width="200">
        <template #default="scope">
          <a href="javascript:void(0)" style="color:#409EFF;" @click="openLink(scope.row)">
            {{ scope.row.title }}
          </a>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="类别" width="140" />
      <el-table-column prop="account" label="账号" width="160" />
      <el-table-column v-if="isSuperAdmin" label="添加人" width="180">
        <template #default="scope">
          <span>{{ scope.row.creator_username || '未知' }}</span>
          <span style="color:#909399;">{{ scope.row.creator_role ? '（' + scope.row.creator_role + '）' : '' }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="140">
        <template #default="scope">
          <el-button size="small" @click="editRow(scope.row)">编辑</el-button>
          <el-button type="danger" size="small" @click="removeRow(scope.row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import axios from 'axios'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})

const items = ref([])
const loading = ref(false)
const form = ref({
  id: null,
  category: '',
  title: '',
  url: '',
  account: '',
  remark: '',
})

const filterCategory = ref('')
const quickCategory = ref('')
const quickLinkId = ref(null)

const categoryOptions = computed(() => {
  const set = new Set()
  for (const it of items.value) {
    if (it.category) set.add(it.category)
  }
  return Array.from(set)
})

const filteredItems = computed(() => {
  if (!filterCategory.value) return items.value
  return items.value.filter(it => it.category === filterCategory.value)
})

const quickLinks = computed(() => {
  if (!quickCategory.value) return []
  return items.value.filter(it => it.category === quickCategory.value)
})

const isSuperAdmin = computed(() => props.currentUser && props.currentUser.role === 'superadmin')

async function load() {
  loading.value = true
  try {
    const { data } = await axios.get('/api/nav_links')
    items.value = Array.isArray(data) ? data : []
  } catch (e) {
    items.value = []
  } finally {
    loading.value = false
  }
}

function onReset() {
  form.value = { id: null, category: '', title: '', url: '', account: '', remark: '' }
}

async function onSave() {
  if (!form.value.title || !form.value.url) return
  const payload = {
    id: form.value.id || 0,
    category: form.value.category,
    title: form.value.title,
    url: form.value.url,
    account: form.value.account,
    remark: form.value.remark,
  }
  await axios.post('/api/nav_links', payload)
  await load()
  onReset()
}

function editRow(row) {
  form.value = {
    id: row.id,
    category: row.category || '',
    title: row.title,
    url: row.url,
    account: row.account,
    remark: row.remark,
  }
}

async function removeRow(row) {
  await axios.delete(`/api/nav_links/${row.id}`)
  await load()
  if (form.value.id === row.id) {
    onReset()
  }
}

function openLink(row) {
  if (!row || !row.url) return
  window.open(row.url, '_blank', 'noopener')
}

function onQuickOpen(id) {
  if (!id) return
  const row = items.value.find(it => it.id === id)
  if (row) {
    openLink(row)
  }
}

onMounted(load)
</script>
