<template>
  <el-drawer :with-header="false" direction="ltr" size="300">
    <el-menu>
      <el-menu-item index="pc_manager" @click="Select">
        <el-icon>
          <Memo />
        </el-icon>
        <span>机器管理</span>
      </el-menu-item>
      <el-sub-menu index="setting">
        <template #title>
          <el-icon>
            <setting />
          </el-icon>
          <span>设置</span>
        </template>
        <el-menu-item index="system_cfg" @click="Select">系统设置</el-menu-item>
        <el-menu-item index="log_query" @click="Select">日志查看</el-menu-item>
      </el-sub-menu>
      <el-menu-item index="about" @click="Select">
        <el-icon>
          <ChatDotRound />
        </el-icon>
        <span>关于</span>
      </el-menu-item>
    </el-menu>
  </el-drawer>
</template>
  
<script lang="ts" setup>
import router from '@/router'
import { Memo, ChatDotRound } from '@element-plus/icons-vue'

function Select(obj: any) {
  if (obj.index == 'pc_manager') {
    router.push('/')
  } else if (obj.index == 'system_cfg') {
    router.push('/config')
  } else if (obj.index == 'log_query') {
    router.push('/log')
  } else if (obj.index == 'about') {
    router.push('/about')
  }
}
</script>