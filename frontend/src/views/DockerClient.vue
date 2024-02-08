<template>
  <MainPage>
    <template #header>

    </template>

    <template #main>
      <el-tabs v-model="curSelect" class="docker navigation relative" type="border-card">
        <el-tab-pane class="h-full" name="image" label="镜像">
          <Images v-if="curSelect=='image'" :group="group" />
        </el-tab-pane>
        <el-tab-pane class="h-full" name="container" label="容器">
          <Containers v-if="curSelect=='container'" :group="group" />
        </el-tab-pane>
        <el-tab-pane class="h-full" name="network" label="网络">

        </el-tab-pane>
      </el-tabs>
    </template>
  </MainPage>
</template>
  
<script setup lang="ts">
import MainPage from '@/components/MainPage.vue'
import Images from '@/components/docker/Images.vue'
import Containers from '@/components/docker/Containers.vue'
import { computed, onMounted, ref } from 'vue'
import { AsyncFetch } from '@/lib/comm';

const group: string = 'api/docker/'

const curSelect = ref('container')

onMounted(() => {
})

</script>
  
<style>
.docker .el-tabs__content {
  height: calc(100% - 30px);
}
</style>