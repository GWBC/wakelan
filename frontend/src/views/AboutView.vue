<template>
  <Navigation v-model="navigationShow" />
  <el-container class="wakelan-layout">
    <el-header class="wakelan-header">
      <el-row :gutter="10">
        <el-col :xs="2" :sm="2" :md="1">
          <el-button :icon="Menu" @click="navigationShow = true" />
        </el-col>
      </el-row>
    </el-header>
    <el-main class="wakelan-main">
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Navigation from './Navigation.vue'
import { Menu } from '@element-plus/icons-vue'

const navigationShow = ref(false)
</script>
