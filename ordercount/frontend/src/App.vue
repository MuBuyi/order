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
          <el-menu-item v-if="canAccessMenu('stats')" index="stats">订单统计</el-menu-item>
          <el-menu-item v-if="canAccessMenu('shop-info')" index="shop-info">店铺广告</el-menu-item>
          <el-menu-item v-if="canAccessMenu('settlement')" index="settlement">结账工具</el-menu-item>
          <el-menu-item v-if="canAccessMenu('returns')" index="returns">退货管理</el-menu-item>
          <el-menu-item v-if="canAccessMenu('product')" index="product">商品管理</el-menu-item>
          <el-menu-item v-if="canAccessMenu('shop-manage')" index="shop-manage">店铺管理</el-menu-item>
          <el-menu-item v-if="canAccessMenu('exchange-tool')" index="exchange-tool">汇率小工具</el-menu-item>
          <el-menu-item v-if="canAccessMenu('charts')" index="charts">图表统计</el-menu-item>
          <el-menu-item v-if="canAccessMenu('users')" index="users">用户管理</el-menu-item>
          <div style="flex:1;" />
          <el-menu-item v-if="canAccessMenu('nav-helper')" index="nav-helper">导航助手</el-menu-item>
        </el-menu>
      </el-aside>
      <el-main class="layout-main">
        <ExchangeRatesBar style="margin-bottom:10px;" />

        <!-- 首页概览视图 -->
        <template v-if="activeMenu === 'home'">
          <HomeDashboard />
        </template>

        <!-- 订单统计视图 -->
        <template v-else-if="activeMenu === 'stats' && canAccessMenu('stats')">
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
        <template v-else-if="activeMenu === 'charts' && canAccessMenu('charts')">
          <StatsDashboard />
        </template>

        <!-- 结账工具视图（根据权限控制） -->
        <template v-else-if="activeMenu === 'settlement' && canAccessMenu('settlement')">
          <ProfitTool />
          <SettlementList :current-user="currentUser" />
        </template>

        <!-- 退货管理（与结账工具同级） -->
        <template v-else-if="activeMenu === 'returns' && canAccessMenu('returns')">
          <ReturnManagement :current-user="currentUser" />
        </template>

        <!-- 商品管理视图（根据权限控制；编辑权限由内部控制） -->
        <template v-else-if="activeMenu === 'product' && canAccessMenu('product')">
          <ProductManager :current-user="currentUser" />
        </template>

        <!-- 店铺管理视图（根据权限控制） -->
        <template v-else-if="activeMenu === 'shop-manage' && canAccessMenu('shop-manage')">
          <ShopManager :current-user="currentUser" />
        </template>

        <!-- 现有店铺信息视图：展示店铺每日广告费用等信息 -->
        <template v-else-if="activeMenu === 'shop-info' && canAccessMenu('shop-info')">
          <StoreInfo :current-user="currentUser" />
        </template>

        <!-- 用户管理视图（根据权限控制） -->
        <template v-else-if="activeMenu === 'users' && canAccessMenu('users')">
          <UserManager :current-user="currentUser" />
        </template>
        
        <!-- 汇率小工具页面 -->
        <template v-else-if="activeMenu === 'exchange-tool' && canAccessMenu('exchange-tool')">
          <CurrencyConverter />
        </template>

        <!-- 导航助手 -->
        <template v-else-if="activeMenu === 'nav-helper' && canAccessMenu('nav-helper')">
          <NavigationHelper :current-user="currentUser" />
        </template>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { ref, computed } from 'vue'
import axios from 'axios'
import { MENU_ORDER, MENU_PERMISSION_MAP } from './utils/pagePermissions'
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
import ReturnManagement from './components/ReturnManagement.vue'

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

function canAccessMenu(menuKey) {
  if (!currentUser.value) return false
  if (isSuperAdmin.value) return true
  const permKey = MENU_PERMISSION_MAP[menuKey]
  if (!permKey) return true
  return hasPerm(permKey)
}

function firstAccessibleMenu() {
  for (const key of MENU_ORDER) {
    if (canAccessMenu(key)) {
      return key
    }
  }
  return 'home'
}

// 记住上次选中的菜单，刷新后保持在同一页面
const ACTIVE_MENU_STORAGE_KEY = 'ordercount-active-menu'
const savedMenu = typeof window !== 'undefined'
  ? window.localStorage.getItem(ACTIVE_MENU_STORAGE_KEY)
  : null
// 兼容旧版本中使用的 'shop' 菜单索引，统一映射到新的 'shop-manage'
// 默认首页作为登录后的第一个页面
const initialMenu = savedMenu === 'shop' ? 'shop-manage' : (savedMenu || 'home')
const activeMenu = ref(initialMenu)

// 如果当前用户对上次记住的页面无权限，则跳转到第一个有权限的页面
if (currentUser.value && !canAccessMenu(activeMenu.value)) {
  activeMenu.value = firstAccessibleMenu()
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, activeMenu.value)
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
  if (!canAccessMenu(key)) {
    activeMenu.value = firstAccessibleMenu()
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, activeMenu.value)
    }
    return
  }
  activeMenu.value = key
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, key)
  }
}

function onLoggedIn(user) {
	currentUser.value = user
  // 登录后进入第一个有权限的页面
  activeMenu.value = firstAccessibleMenu()
  if (typeof window !== 'undefined') {
    window.localStorage.setItem(ACTIVE_MENU_STORAGE_KEY, activeMenu.value)
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
