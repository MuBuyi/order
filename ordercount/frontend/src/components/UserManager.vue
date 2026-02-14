<template>
  <el-card shadow="hover" style="margin-bottom:20px;">
    <template #header>
      <div style="display:flex;justify-content:space-between;align-items:center;">
        <span>用户与角色管理</span>
      </div>
    </template>

    <el-form :model="form" label-width="80px" @submit.prevent style="margin-bottom:16px;">
      <el-row :gutter="10">
        <el-col :span="6">
          <el-form-item label="用户名">
            <el-input v-model="form.username" placeholder="新用户名" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" placeholder="登录密码" />
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="角色">
            <el-select v-model="form.role" placeholder="选择角色" style="width:100%;">
              <el-option label="超级管理员" value="superadmin" />
              <el-option label="管理员" value="admin" />
              <el-option label="员工" value="staff" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="6">
          <el-form-item label="页面权限">
            <el-checkbox-group v-model="form.permissions">
              <el-checkbox label="settlement">结账工具</el-checkbox>
              <el-checkbox label="product">商品管理</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </el-col>
        <el-col :span="6" style="display:flex;align-items:flex-end;">
          <el-form-item>
            <el-button type="primary" @click="onCreate">创建用户</el-button>
            <span v-if="msg" style="margin-left:10px;font-size:12px;" :style="{color: ok ? '#67C23A' : '#F56C6C'}">{{ msg }}</span>
          </el-form-item>
        </el-col>
      </el-row>
    </el-form>

    <el-table :data="users" size="small" border style="width:100%;">
      <el-table-column prop="username" label="用户名" width="180" />
      <el-table-column prop="role" label="角色" width="200">
        <template #default="scope">
          <el-select v-model="scope.row.role" size="small" style="width:140px;" @change="(v) => onRoleChange(scope.row, v)">
            <el-option label="超级管理员" value="superadmin" />
            <el-option label="管理员" value="admin" />
            <el-option label="员工" value="staff" />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="页面权限" width="260">
        <template #default="scope">
          <el-checkbox-group
            v-model="scope.row.permissions"
            :disabled="currentUser && scope.row.id === currentUser.id"
            @change="(vals) => onPermChange(scope.row, vals)"
          >
            <el-checkbox label="settlement">结账工具</el-checkbox>
            <el-checkbox label="product">商品管理</el-checkbox>
          </el-checkbox-group>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="180" />
    </el-table>
  </el-card>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
})

const users = ref([])
const form = ref({ username: '', password: '', role: 'staff', permissions: [] })
const msg = ref('')
const ok = ref(true)

async function loadUsers() {
  const res = await axios.get('/api/users')
  users.value = res.data || []
}

async function onCreate() {
  msg.value = ''
  ok.value = true
  if (!form.value.username || !form.value.password || !form.value.role) {
    msg.value = '请填写完整的用户名、密码和角色'
    ok.value = false
    return
  }
  try {
    await axios.post('/api/users', {
      username: form.value.username,
      password: form.value.password,
      role: form.value.role,
      permissions: form.value.permissions,
    })
    msg.value = '创建成功'
    ok.value = true
    form.value = { username: '', password: '', role: 'staff', permissions: [] }
    loadUsers()
  } catch (e) {
    ok.value = false
    if (e && e.response && e.response.data && e.response.data.error) {
      msg.value = e.response.data.error
    } else {
      msg.value = '创建失败，请稍后重试'
    }
  }
}

async function onRoleChange(row, newRole) {
  try {
    await axios.put(`/api/users/${row.id}/role`, { role: newRole })
  } catch (e) {
    // 恢复原角色
    msg.value = (e && e.response && e.response.data && e.response.data.error) || '更新角色失败'
    ok.value = false
    loadUsers()
  }
}

async function onPermChange(row, vals) {
  try {
    await axios.put(`/api/users/${row.id}/permissions`, { permissions: vals })
    msg.value = '更新权限成功'
    ok.value = true
  } catch (e) {
    msg.value = (e && e.response && e.response.data && e.response.data.error) || '更新权限失败'
    ok.value = false
    loadUsers()
  }
}

onMounted(loadUsers)
</script>
