<template>

  <Login v-if="!currentUser" @logged-in="onLoggedIn" />
  <!-- 主布局使用整屏高度，配合左侧菜单冻结 -->
  <el-container v-else style="height:100vh;">
		<el-header class="layout-header" style="background:#409EFF;color:#fff;font-size:22px;display:flex;align-items:center;">
      <span style="font-size:22px;flex:1;">订单统计管理后台</span>
      <div v-if="currentUser" style="font-size:14px;display:flex;align-items:center;gap:10px;">
        <span>当前用户：{{ currentUser.username }}</span>
        <el-button size="small" type="danger" @click="onLogout">退出登录</el-button>
      </div>
    </el-header>
    <el-container style="height:100%;">
      <el-aside class="layout-aside" width="200px" style="background:#fff;border-right:1px solid #ebeef5;">
        <el-menu :default-active="activeMenu" @select="onSelect" router="false" style="height:100%;display:flex;flex-direction:column;">
          <el-menu-item index="home">首页</el-menu-item>
          <el-menu-item index="stats">订单统计</el-menu-item>
          <el-menu-item v-if="canSeeSettlement" index="settlement">结账工具</el-menu-item>
          <el-menu-item v-if="canSeeProduct" index="product">商品管理</el-menu-item>
          <el-sub-menu v-if="canSeeShop" index="shop">
            <template #title>店铺管理</template>
            <el-menu-item index="shop-manage">店铺管理</el-menu-item>
            <el-menu-item index="shop-info">现有店铺信息</el-menu-item>
          </el-sub-menu>
          <el-menu-item index="exchange-tool">汇率小工具</el-menu-item>
          <el-menu-item index="charts">图表统计</el-menu-item>
          <el-menu-item v-if="isSuperAdmin" index="users">用户管理</el-menu-item>
          <div style="flex:1;" />
          <el-menu-item index="nav-helper">导航助手</el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="layout-main">
        <ExchangeRatesBar style="margin-bottom:10px;" />

        <!-- 首页概览视图 -->
        <template v-if="activeMenu === 'home'">
          <HomeDashboard />
        </template>

        <!-- 订单统计视图 -->
        <template v-else-if="activeMenu === 'stats'">
          <el-row :gutter="20">
            <el-col :span="12">
              <OrderForm @refresh="refreshAll" @go-shop-info="goShopInfo" />
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

        <!-- 图表统计视图 -->
        <template v-else-if="activeMenu === 'charts'">
          <StatsDashboard />
        </template>

        <!-- 结账工具视图（根据权限控制） -->
        <template v-else-if="activeMenu === 'settlement' && canSeeSettlement">
          <ProfitTool />
          <SettlementList :current-user="currentUser" />
        </template>

        <!-- 商品管理视图（根据权限控制；编辑权限由内部控制） -->
        <template v-else-if="activeMenu === 'product' && canSeeProduct">
          <ProductManager :current-user="currentUser" />
        </template>

        <!-- 店铺管理视图（根据权限控制） -->
        <template v-else-if="activeMenu === 'shop-manage' && canSeeShop">
          <ShopManager :current-user="currentUser" />
        </template>

        <!-- 现有店铺信息视图：展示店铺每日广告费用等信息 -->
        <template v-else-if="activeMenu === 'shop-info' && canSeeShop">
          <StoreInfo :current-user="currentUser" />
        </template>

        <!-- 用户管理视图（仅超级管理员可见） -->
        <template v-else-if="activeMenu === 'users' && isSuperAdmin">
          <UserManager :current-user="currentUser" />
        </template>
        
        <!-- 汇率小工具页面 -->
        <template v-else-if="activeMenu === 'exchange-tool'">
          <CurrencyConverter />
        </template>

        <!-- 导航助手 -->
        <template v-else-if="activeMenu === 'nav-helper'">
          <NavigationHelper />
        </template>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from 'axios'
import OrderForm from './components/OrderForm.vue'
import HomeDashboard from './components/HomeDashboard.vue'
import TodaySales from './components/TodaySales.vue'
import TodayGoodsCost from './components/TodayGoodsCost.vue'
import OrderCharts from './components/OrderCharts.vue'
import StatsDashboard from './components/StatsDashboard.vue'
import ProfitTool from './components/ProfitTool.vue'
import ExchangeRatesBar from './components/ExchangeRatesBar.vue'
import CurrencyConverter from './components/CurrencyConverter.vue'
import NavigationHelper from './components/NavigationHelper.vue'
import SettlementList from './components/SettlementList.vue'
import OrderList from './components/OrderList.vue'
import ProductManager from './components/ProductManager.vue'
import ShopManager from './components/ShopManager.vue'
import StoreInfo from './components/StoreInfo.vue'
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
const canSeeShop = computed(() => isSuperAdmin.value || hasPerm('shop'))

// 记住上次选中的菜单，刷新后保持在同一页面
const ACTIVE_MENU_STORAGE_KEY = 'ordercount-active-menu'
const savedMenu = typeof window !== 'undefined'
  ? window.localStorage.getItem(ACTIVE_MENU_STORAGE_KEY)
  : null
// 兼容旧版本中使用的 'shop' 菜单索引，统一映射到新的 'shop-manage'
// 默认首页作为登录后的第一个页面
const initialMenu = savedMenu === 'shop' ? 'shop-manage' : (savedMenu || 'home')
const activeMenu = ref(initialMenu)

// 如果当前用户无权限，但上次记住的是结账工具/商品管理/用户管理，则强制回到订单统计
if (currentUser.value && (
  (!canSeeSettlement.value && activeMenu.value === 'settlement') ||
  (!canSeeProduct.value && activeMenu.value === 'product') ||
  (!canSeeShop.value && (activeMenu.value === 'shop-manage' || activeMenu.value === 'shop-info')) ||
  (!isSuperAdmin.value && activeMenu.value === 'users')
)) {
  activeMenu.value = 'stats'
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, 'stats')
  }
}

function refreshAll(date) {
  todaySales.value && todaySales.value.load(date)
  todayGoodsCost.value && todayGoodsCost.value.load(date)
  orderCharts.value && orderCharts.value.load()
}

function goShopInfo() {
  // 复用菜单选择逻辑及权限判断
  onSelect('shop-info')
}

function onSelect(key) {
  // 无结账工具权限的用户禁止切换到结账工具
  if (!canSeeSettlement.value && key === 'settlement') {
    activeMenu.value = 'stats'
    return
  }
  // 无店铺管理权限的用户禁止切换到店铺管理
  if (!canSeeShop.value && (key === 'shop-manage' || key === 'shop-info')) {
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
  // 登录后优先进入首页
  activeMenu.value = 'home'
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, 'home')
  }
  // 再根据角色校验当前菜单是否允许访问（如果后续手动切换）
  if ((!canSeeSettlement.value && activeMenu.value === 'settlement') ||
  (!canSeeShop.value && (activeMenu.value === 'shop-manage' || activeMenu.value === 'shop-info')) ||
      (!isSuperAdmin.value && activeMenu.value === 'users')) {
    activeMenu.value = 'home'
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, 'home')
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
html, body, #app {
  height: 100%;
}

body {
  margin: 0;
  background: #f5f7fa;
}

/* 冻结左侧菜单栏：占满视口高度，内部滚动 */
.layout-aside {
  height: 100vh;
  overflow-y: auto;
}

/* 右侧内容区域随页面滚动，保持与左侧同高 */
.layout-main {
  height: 100vh;
  overflow-y: auto;
}
</style>
