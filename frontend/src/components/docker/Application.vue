<template>
    <el-scrollbar>
        <div class="app_container">
            <el-button class="app_card !flex !items-end !pb-6" v-for="(item, index) in app_list">
                <span class="!font-bold">
                    {{ item }}
                </span>
            </el-button>
        </div>
    </el-scrollbar>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

const group: string = 'api/docker/'
const app_list = ref(["xxx"])

onMounted(() => {
})

</script>

<style scoped>
.app_container {
    display: flex;
    flex-wrap: wrap;
}

.app_card {
    margin: 0 !important;
    max-width: 200px;
    height: 300px;
    flex-grow: 1;
    border-radius: 5px;
}
</style>