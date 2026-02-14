<template>
  <div class="login-wrapper">
    <el-card class="login-card" shadow="hover">
      <template #header>
        <div class="login-title">登录系统</div>
      </template>
      <el-form :model="form" label-width="80px" @submit.prevent>
        <el-form-item label="用户名">
          <el-input v-model="form.username" autocomplete="off" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" show-password autocomplete="off" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" @click="onLogin">登录</el-button>
          <span v-if="msg" class="login-msg" :class="{ error: !ok }">{{ msg }}</span>
        </el-form-item>
          <div class="login-hint">
            超级管理员：root / root123；管理员：admin / admin123（请尽快修改默认密码）
          </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'

const emit = defineEmits(['logged-in'])

  const form = ref({ username: 'root', password: '' })
const loading = ref(false)
const msg = ref('')
const ok = ref(true)

const TOKEN_KEY = 'ordercount-token'
const USER_KEY = 'ordercount-user'

async function onLogin() {
  msg.value = ''
  ok.value = true
  if (!form.value.username || !form.value.password) {
    msg.value = '请输入用户名和密码'
    ok.value = false
    return
  }
  loading.value = true
  try {
    const res = await axios.post('/api/login', {
      username: form.value.username,
      password: form.value.password,
    })
    const { token, user } = res.data || {}
    if (!token || !user) {
      msg.value = '登录返回数据异常'
      ok.value = false
      return
    }
    // 保存本地并设置 axios 默认头
    window.localStorage.setItem(TOKEN_KEY, token)
    window.localStorage.setItem(USER_KEY, JSON.stringify(user))
    axios.defaults.headers.common.Authorization = `Bearer ${token}`
    emit('logged-in', user)
  } catch (e) {
    ok.value = false
    if (e && e.response && e.response.data && e.response.data.error) {
      msg.value = e.response.data.error
    } else {
      msg.value = '登录失败，请稍后重试'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-wrapper {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f7fa;
}

.login-card {
  width: 360px;
}

.login-title {
  font-size: 18px;
  font-weight: 600;
}

.login-msg {
  margin-left: 12px;
  font-size: 12px;
}

.login-msg.error {
  color: #f56c6c;
}

.login-hint {
  margin-top: 8px;
  font-size: 12px;
  color: #909399;
}
</style>
