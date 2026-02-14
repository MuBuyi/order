<template>
  <Login v-if="!currentUser" @logged-in="onLoggedIn" />
  <el-container v-else style="height:100%;">
		<el-header class="layout-header" style="background:#409EFF;color:#fff;font-size:22px;display:flex;align-items:center;">
      <span style="font-size:22px;flex:1;">订单统计管理后台</span>
      <div v-if="currentUser" style="font-size:14px;display:flex;align-items:center;gap:10px;">
        <span>当前用户：{{ currentUser.username }}</span>
        <el-button size="small" type="danger" @click="onLogout">退出登录</el-button>
      </div>
    </el-header>
    <el-container style="height:100%;">
      <el-aside class="layout-aside" width="200px" style="background:#fff;border-right:1px solid #ebeef5;">
        <el-menu :default-active="activeMenu" @select="onSelect" router="false">
          <el-menu-item index="stats">订单统计</el-menu-item>
          <el-menu-item v-if="canSeeSettlement" index="settlement">结账工具</el-menu-item>
          <el-menu-item v-if="canSeeProduct" index="product">商品管理</el-menu-item>
          <el-menu-item v-if="isSuperAdmin" index="users">用户管理</el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="layout-main">
        <ExchangeRatesBar style="margin-bottom:10px;" />

        <!-- 订单统计视图 -->
        <template v-if="activeMenu === 'stats'">
          <el-row :gutter="20">
            <el-col :span="12">
              <OrderForm @refresh="refreshAll" />
            </el-col>
            <el-col :span="12">
              <TodaySales ref="todaySales" />
              <TodayGoodsCost ref="todayGoodsCost" />
              <OrderCharts ref="orderCharts" />
            </el-col>
          </el-row>
          <el-divider />
          <OrderList />
        </template>

        <!-- 结账工具视图（根据权限控制） -->
        <template v-else-if="activeMenu === 'settlement' && canSeeSettlement">
          <ProfitTool />
          <SettlementList />
        </template>

        <!-- 商品管理视图（根据权限控制；编辑权限由内部控制） -->
        <template v-else-if="activeMenu === 'product' && canSeeProduct">
          <ProductManager :current-user="currentUser" />
        </template>

        <!-- 用户管理视图（仅超级管理员可见） -->
        <template v-else-if="activeMenu === 'users' && isSuperAdmin">
          <UserManager :current-user="currentUser" />
        </template>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from 'axios'
import OrderForm from './components/OrderForm.vue'
import TodaySales from './components/TodaySales.vue'
import TodayGoodsCost from './components/TodayGoodsCost.vue'
import OrderCharts from './components/OrderCharts.vue'
import ProfitTool from './components/ProfitTool.vue'
import ExchangeRatesBar from './components/ExchangeRatesBar.vue'
import SettlementList from './components/SettlementList.vue'
import OrderList from './components/OrderList.vue'
import ProductManager from './components/ProductManager.vue'
import Login from './components/Login.vue'
import UserManager from './components/UserManager.vue'

const todaySales = ref(null)
const todayGoodsCost = ref(null)
const orderCharts = ref(null)

// 登录用户
const USER_KEY = 'ordercount-user'
const TOKEN_KEY = 'ordercount-token'
const savedUser = typeof window !== 'undefined' ? window.localStorage.getItem(USER_KEY) : null
const currentUser = ref(savedUser ? JSON.parse(savedUser) : null)

// 角色辅助判断
const isSuperAdmin = computed(() => currentUser.value && currentUser.value.role === 'superadmin')
const isAdminLike = computed(() => currentUser.value && (currentUser.value.role === 'admin' || currentUser.value.role === 'superadmin'))

// 页面权限辅助函数（permissions 可以是逗号分隔字符串或数组）
function hasPerm(key) {
  if (!currentUser.value) return false
  const raw = currentUser.value.permissions
  if (!raw) return false
  if (Array.isArray(raw)) {
    return raw.includes(key)
  }
  return String(raw)
    .split(',')
    .map(s => s.trim())
    .filter(Boolean)
    .includes(key)
}

const canSeeSettlement = computed(() => isSuperAdmin.value || hasPerm('settlement'))
const canSeeProduct = computed(() => isSuperAdmin.value || hasPerm('product'))

// 记住上次选中的菜单，刷新后保持在同一页面
const ACTIVE_MENU_STORAGE_KEY = 'ordercount-active-menu'
const savedMenu = typeof window !== 'undefined'
  ? window.localStorage.getItem(ACTIVE_MENU_STORAGE_KEY)
  : null
const activeMenu = ref(savedMenu || 'stats')

// 如果当前用户无权限，但上次记住的是结账工具/商品管理/用户管理，则强制回到订单统计
if (currentUser.value && (
  (!canSeeSettlement.value && activeMenu.value === 'settlement') ||
  (!canSeeProduct.value && activeMenu.value === 'product') ||
  (!isSuperAdmin.value && activeMenu.value === 'users')
)) {
  activeMenu.value = 'stats'
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, 'stats')
  }
}

function refreshAll() {
  todaySales.value && todaySales.value.load()
  todayGoodsCost.value && todayGoodsCost.value.load()
  orderCharts.value && orderCharts.value.load()
}

function onSelect(key) {
  // 无结账工具权限的用户禁止切换到结账工具
  if (!canSeeSettlement.value && key === 'settlement') {
    activeMenu.value = 'stats'
    return
  }
  // 非超级管理员禁止进入用户管理
  if (!isSuperAdmin.value && key === 'users') {
    activeMenu.value = 'stats'
    return
  }
  activeMenu.value = key
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, key)
  }
}

function onLoggedIn(user) {
	currentUser.value = user
	// 登录后根据角色调整当前菜单
  if ((!canSeeSettlement.value && activeMenu.value === 'settlement') ||
      (!isSuperAdmin.value && activeMenu.value === 'users')) {
    activeMenu.value = 'stats'
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, 'stats')
    }
  }
}

function onLogout() {
  // 清除本地登录状态
  if (typeof window !== 'undefined') {
    window.localStorage.removeItem(USER_KEY)
    window.localStorage.removeItem(TOKEN_KEY)
  }
  delete axios.defaults.headers.common.Authorization
  // 重置当前用户和菜单
  currentUser.value = null
}
</script>

<style>
body {
  margin: 0;
  background: #f5f7fa;
}
</style>
