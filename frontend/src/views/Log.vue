<template>
    <MainPage>
        <template #header />
        <template #main>
            <el-card class="navigation" body-class="flex flex-col !p-1 h-full">
                <el-table style="flex: 1;" :data="table_data" empty-text=" " :show-overflow-tooltip="false" stripe v-loading="table_loading">
                    <el-table-column prop="time" label="时间" width="180" />
                    <el-table-column prop="cmd" label="动作" width="180" />
                    <el-table-column prop="msg" label="信息" />
                </el-table>
                <el-pagination style="margin: 10px 0px 10px 10px;" :page-sizes="[18, 40, 80, 100]"
                    layout="total, sizes, prev, pager, next, jumper" background :total="total" :default-page-size="pageSize"
                    @current-change="Change" @size-change="SizeChange">
                </el-pagination>
            </el-card>
        </template>
    </MainPage>
</template>
 
<script setup lang="ts">
import { Fetch } from '@/lib/comm'
import MainPage from '@/components/MainPage.vue'
import { ref, reactive, onMounted, onUnmounted } from 'vue'

//日志信息
interface LogInfo {
    cmd: string
    msg: string
    time: string
}

//日志总数
interface LogSizeInfo {
    total: number
}

const total = ref(0)
const pageSizes = ref([20, 40, 60, 80, 100])
const pageSize = ref(pageSizes.value[0])

const table_loading = ref(false)

const table_data = reactive<LogInfo[]>([])

const group: string = 'api/system/'

function getData(page: number) {
    table_loading.value = true
    table_data.length = 0
    Fetch<LogInfo[]>(`${group}log?pageSize=${pageSize.value}&page=${page}`, null, infos => {
        for (let i = 0; i < infos.length; ++i) {
            table_data.push(infos[i])
        }

        table_loading.value = false
    })
}

function initTotal() {
    Fetch<LogSizeInfo>(`${group}logsize`, null, info => {
        total.value = info.total
    })
}

function Change(value: number) {
    getData(value)
}

function SizeChange(value: number) {
    console.log(value)
    pageSize.value = value
    getData(1)
}

onMounted(function () {
    initTotal()
    getData(1)
})
</script>
