<template>
    <MainPage>
        <template #header />
        <template #main>
            <div style="height: 100%;" ref="table_ref">
                <el-table :data="table_data" empty-text=" " :height="table_height" stripe v-loading="table_loading">
                    <el-table-column prop="time" label="时间" width="180" />
                    <el-table-column prop="cmd" label="动作" width="180" />
                    <el-table-column prop="msg" label="信息" />
                </el-table>
                <el-pagination class="page" :page-sizes="[20, 30, 40, 60, 80]"
                    layout="total, sizes, prev, pager, next, jumper" background :total="total" :default-page-size="pageSize"
                    @current-change="Change" @size-change="SizeChange">
                </el-pagination>
            </div>
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
const pageSize = ref(20)

const table_ref = ref()
const table_height = ref(0)

const table_loading = ref(false)

const table_data = reactive<LogInfo[]>([])

const group: string = 'api/system/'
let resizeObserver: ResizeObserver | null = null

function initObserverSize() {
    resizeObserver = new ResizeObserver(entries => {
        table_height.value = table_ref.value.offsetHeight - 50
    })

    resizeObserver.observe(table_ref.value)
}

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
    initObserverSize()
    initTotal()
    getData(1)
})

onUnmounted(function () {
    resizeObserver!.disconnect()
})

</script>

<style scoped>
.page {
    margin-top: 10px;
    margin-left: 10px;
}
</style>
 