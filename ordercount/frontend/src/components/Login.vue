<template>
  <div class="login-wrapper">
    <div class="login-bg"></div>
    <div class="login-bg-overlay"></div>
    <div class="login-content">
      <el-card class="login-card" shadow="hover">
        <template #header>
          <div class="login-title">订单统计管理后台</div>
        </template>
        <el-form :model="form" label-width="72px" @submit.prevent>
          <el-form-item label="用户名">
            <el-input v-model="form.username" autocomplete="off" />
          </el-form-item>
          <el-form-item label="密码">
            <el-input v-model="form.password" type="password" show-password autocomplete="off" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="loading" style="width:100%;" @click="onLogin">登 录</el-button>
          </el-form-item>
          <div class="login-msg-wrapper" v-if="msg">
            <span class="login-msg" :class="{ error: !ok }">{{ msg }}</span>
          </div>
          <div class="login-hint">
            默认账户：超级管理员 root / root123；管理员 admin / admin123（建议登录后尽快修改密码）
          </div>
        </el-form>
      </el-card>
    </div>
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
  position: relative;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.login-bg {
  position: absolute;
  top: -20%;
  left: -20%;
  width: 140%;
  height: 140%;
  background: radial-gradient(circle at 0% 0%, #409eff 0, transparent 55%),
    radial-gradient(circle at 100% 0%, #67c23a 0, transparent 55%),
    radial-gradient(circle at 0% 100%, #e6a23c 0, transparent 55%),
    radial-gradient(circle at 100% 100%, #f56c6c 0, transparent 55%);
  filter: blur(40px);
  opacity: 0.9;
  animation: float-bg 18s linear infinite alternate;
}

.login-bg-overlay {
  position: absolute;
  inset: 0;
  background: rgba(10, 15, 30, 0.6);
}

.login-content {
  position: relative;
  z-index: 1;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
  box-sizing: border-box;
}

.login-card {
  width: 380px;
  border-radius: 14px;
  box-shadow: 0 18px 45px rgba(0, 0, 0, 0.45);
  background: rgba(255, 255, 255, 0.97);
}

.login-title {
  font-size: 20px;
  font-weight: 600;
  text-align: center;
}

.login-msg-wrapper {
  text-align: center;
  margin-bottom: 4px;
}

.login-msg {
  font-size: 12px;
}

.login-msg.error {
  color: #f56c6c;
}

.login-hint {
  margin-top: 4px;
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
}

@keyframes float-bg {
  0% {
    transform: translate3d(0, 0, 0) scale(1);
  }
  100% {
    transform: translate3d(-30px, -30px, 0) scale(1.05);
  }
}
</style>
