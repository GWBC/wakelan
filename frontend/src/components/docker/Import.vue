<template>
    <el-card class="h-full" body-class="h-full !pb-1" :width="160">
        <el-table table-layout="auto" class="!h-full" :data="backupData" stripe empty-text=" ">
            <el-table-column label="名称" prop="name"></el-table-column>
            <el-table-column label="大小" prop="size">
                <template #default="scope">{{ (scope.row.size / 1024 / 1024).toFixed(2) }} MB</template>
            </el-table-column>
            <el-table-column label="修改时间" prop="modify_time"></el-table-column>
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref, nextTick } from 'vue'
import { AsyncFetch } from '@/lib/comm'
import { WBSocket } from '@/lib/websocket'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

interface BackupInfos {
    name: string
    size: number
    modify_time: string
}

const backupData = ref<BackupInfos[]>([])

const props = defineProps<{
    group: string
}>()

function getInfos() {
    AsyncFetch<BackupInfos[]>(`${props.group}getBackupInfos`, null).then((infos) => {
        backupData.value = infos
    })
}

onMounted(() => {
    getInfos()
})

</script>

<style>
.image_info_dlg .el-dialog__body {
    height: calc(100% - 30px);
}
</style>