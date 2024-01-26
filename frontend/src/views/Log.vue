<template>
    <MainPage>
        <template #header />
        <template #main>
            <el-card class="log-card" body-class="log-card-body">
                <div class="log-main" ref="table_ref">
                    <el-table :data="table_data" empty-text=" " :height="table_height" stripe v-loading="table_loading">
                        <el-table-column prop="time" label="时间" width="180" />
                        <el-table-column prop="cmd" label="动作" width="180" />
                        <el-table-column prop="msg" label="信息" />
                    </el-table>
                    <el-pagination class="page" :page-sizes="[18, 40, 80, 100]"
                        layout="total, sizes, prev, pager, next, jumper" background :total="total"
                        :default-page-size="pageSize" @current-change="Change" @size-change="SizeChange">
                    </el-pagination>
                </div>
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
const pageSizes = ref([18, 20, 40, 60, 80, 100])
const pageSize = ref(pageSizes.value[0])

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

<style>
.log-card-body {
    height: 98%;
}
</style>

<style scoped>
.log-card {
    margin: 0px 20px 0px 20px;
    height: 98%;
}

.log-main {
    height: 100%;
}

.page {
    margin-top: 10px;
    margin-left: 10px;
}
</style>
 