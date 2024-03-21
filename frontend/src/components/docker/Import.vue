<template>
    <el-card class="h-full" body-class="h-full !pb-1" :width="160">
        <el-table v-loading="loading" table-layout="auto" class="!h-full" :data="backupData" stripe empty-text=" ">
            <el-table-column label="名称" prop="name"></el-table-column>
            <el-table-column label="类型" prop="type">
                <template #default="scope">
                    <el-tag class="w-32" :type="scope.row.type == 'image' ? 'success' : 'primary'">
                        {{ scope.row.type }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column label="大小" prop="size">
                <template #default="scope">{{ (scope.row.size / 1024 / 1024).toFixed(2) }} MB</template>
            </el-table-column>
            <el-table-column label="修改时间" prop="modify_time"></el-table-column>
            <el-table-column label="操作" width="400">
                <template #default="scope">
                    <el-button type="primary" link @click="onDownload(scope.row)">下载</el-button>
                    <el-button type="primary" link @click="onRestore(scope.row)">恢复</el-button>
                    <el-button type="danger" link @click="onDelete(scope.row)">删除</el-button>
                </template>
            </el-table-column>
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref, nextTick } from 'vue'
import { AsyncFetch, DownloadFileFromURL } from '@/lib/comm'
import { ElMessage, ElMessageBox } from 'element-plus';

interface BackupInfos {
    name: string
    type: string
    size: number
    modify_time: string
}

const loading = ref(false)
const backupData = ref<BackupInfos[]>([])

const props = defineProps<{
    group: string
}>()

function onDelete(row: BackupInfos) {
    ElMessageBox.confirm(
        `是否删除备份：${row.name}`,
        `删除`,
        {
            confirmButtonText: '删除',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {
        AsyncFetch<BackupInfos[]>(`${props.group}delete?name=${row.name}&type=${row.type}`, null).then((infos) => {
            ElMessage.success(`${row.name} 删除成功`)
            backupData.value.splice(backupData.value.indexOf(row), 1)
        })
    })
}

function onRestore(row: BackupInfos) {
    loading.value = true
    AsyncFetch<BackupInfos[]>(`${props.group}restore?name=${row.name}&type=${row.type}`, null).then((infos) => {
        loading.value = false
        ElMessage.success('恢复成功')
    }).catch(() => {
        loading.value = false
    })
}

function onDownload(row: BackupInfos) {
    let fileName = 'docker_backup_' + row.type + '_' + row.name 
    DownloadFileFromURL(`${props.group}/download?file=${row.name}&name=${fileName}&type=${row.type}`, fileName)
}

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