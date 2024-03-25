<template>
    <Navigation v-model="navigationShow" />
    <el-container class="wakelan-layout">
        <el-button link class="wakelan-header-button" @click="onHeaderButtonClick">
            <el-icon>
                <DArrowRight />
            </el-icon>
        </el-button>
        <el-header v-if="props.headerShow" class="wakelan-header">
            <el-row :gutter="10">
                <!-- <el-col :xs="2" :sm="2" :md="1">
                    <el-button :icon="Menu" @click="navigationShow = true" />
                </el-col> -->
                <el-col :xs="22" :sm="22" :md="23">
                    <slot name="header"></slot>
                </el-col>
            </el-row>
        </el-header>
        <el-main class="wakelan-main">
            <slot name="main"></slot>
        </el-main>
    </el-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import Navigation from './Navigation.vue'
import { Menu } from '@element-plus/icons-vue'

const navigationShow = ref(false)

const props = defineProps({
    headerShow: {
        type: Boolean,
        default: false
    }
})

function onHeaderButtonClick() {
    navigationShow.value = true
}

</script>

<style>
.wakelan-header-button {
    position: absolute;
    top: calc(50% - 100px);
    left: 4px;
    padding: 0 !important;
    margin: 0 !important;
    height: 200px !important;
    width: 16px !important;
    z-index: 1000;
}
</style>