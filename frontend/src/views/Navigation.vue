<template>
  <el-drawer :with-header="false" direction="ltr" size="300">
    <el-menu @select="Select">
      <el-menu-item index="pc_manager">
        <el-icon>
          <Memo />
        </el-icon>
        <span>机器管理</span>
      </el-menu-item>
      <el-menu-item index="filetransfer">
        <el-icon>
          <Folder />
        </el-icon>
        <span>文件传输</span>
      </el-menu-item>
      <el-sub-menu index="setting">
        <template #title>
          <el-icon>
            <setting />
          </el-icon>
          <span>设置</span>
        </template>
        <el-menu-item index="system_cfg">系统设置</el-menu-item>
        <el-menu-item index="log_query">日志查看</el-menu-item>
      </el-sub-menu>
      <el-menu-item index="about">
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
import { Memo, Folder, ChatDotRound } from '@element-plus/icons-vue'

function Select(index: any) {
  if (index == 'pc_manager') {
    router.push('/')
  } else if (index == 'system_cfg') {
    router.push('/config')
  } else if (index == 'log_query') {
    router.push('/log')
  } else if (index == 'filetransfer') {
    router.push('/filetransfer')
  } else if (index == 'about') {
    router.push('/about')
  }
}
</script>