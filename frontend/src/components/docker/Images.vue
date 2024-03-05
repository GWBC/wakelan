<template>
    <el-card class="h-full" body-class="h-full !pb-1">
        <el-table table-layout="auto" class="!h-full" :data="imageDatas" stripe empty-text=" ">
            <el-table-column prop="id" label="ID" />
            <el-table-column prop="repostitory" label="仓库" />
            <el-table-column prop="tag" label="版本" />
            <el-table-column label="大小">
                <template #default="scope">
                    <el-text>{{ showSize(scope.row) }}</el-text>
                </template>
            </el-table-column>
            <el-table-column prop="create_time" label="创建时间" />
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { AsyncFetch } from '@/lib/comm';

interface ImageInfo {
    repostitory: string
    tag: string
    id: string
    create_time: string
    size: number
}

const imageDatas = ref<ImageInfo[]>([])
const props = defineProps<{
    group: string
}>()

function showSize(row: ImageInfo): string {
    if (row.size < 1024 * 1024) {
        return (row.size / 1024).toFixed(2) + "KB"
    }

    if (row.size >= 1024 * 1024 && row.size <= 1024 * 1024 * 1024) {
        return (row.size / 1024 / 1024).toFixed(2) + "MB"
    }

    return (row.size / 1024 / 1024 / 1024).toFixed(2) + "GB"
}

function getImages() {
    AsyncFetch<ImageInfo[]>(`${props.group}getImages`, null).then((infos) => {
        imageDatas.value = infos
    })
}

onMounted(() => {
    getImages()
})


</script>