<template>
    <el-dialog v-model="runDlgIsShow">
    </el-dialog>
    <el-dialog class="image_info_dlg h-2/3" v-model="image_infos_dlg_show">
        <el-table class="!h-full" :data="image_infos" stripe v-loading="image_infos_loading">
            <el-table-column label="加星" :width="100" prop="star_count"></el-table-column>
            <el-table-column label="官方" :width="100" prop="is_official"></el-table-column>
            <el-table-column label="名称" :width="200" prop="name">
                <template #default="scope">
                    <el-button type="warning" link @click="image_name = scope.row.name">{{ scope.row.name }}</el-button>
                </template>
            </el-table-column>
            <el-table-column label="描述" prop="description"></el-table-column>
            <el-table-column label="操作" :width="160" prop="tags">
                <template #default="scope">
                    <el-button link type="primary" @click="onPull(scope.row.name)">拉取</el-button>
                </template>
            </el-table-column>
        </el-table>
    </el-dialog>

    <el-dialog v-model="detailIsShow">

        <template #header>
            <h1 class="text-gray-500 text-center font-bold">{{ detailTitle }}</h1>
        </template>
        <el-descriptions title="基本信息" direction="horizontal" :column="4" border>
            <el-descriptions-item :width="150" label="操作系统" :span="4">{{ imageDetails.os }}</el-descriptions-item>
            <el-descriptions-item label="工作路径" :span="4">{{ imageDetails.working_dir }}</el-descriptions-item>
            <el-descriptions-item label="启动命令" :span="4">{{ imageDetails.cmd }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions class="mt-4" title="端口映射" direction="horizontal" :column="4" border>
            <el-descriptions-item v-for="item in imageDetails.exposed_ports" :width="150" :label="item.proto"
                :span="4">{{
        item.port }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions class="mt-4" title="目录映射" direction="horizontal" :column="4" border>
            <el-descriptions-item v-for="item in imageDetails.volumes" :width="10" label="" :span="4">{{ item
                }}</el-descriptions-item>
        </el-descriptions>
        <el-descriptions class="mt-4" title="环境变量" direction="horizontal" :column="4" border>
            <el-descriptions-item v-for="item in imageDetails.env" :width="10" label="" :span="4">{{ item
                }}</el-descriptions-item>
        </el-descriptions>
    </el-dialog>

    <el-card class="h-full" body-class="h-full !pb-1" :width="160">
        <el-table table-layout="auto" class="!h-2/3" :data="imageDatas" stripe empty-text=" ">
            <el-table-column prop="id" label="ID" />
            <el-table-column prop="repostitory" label="仓库" />
            <el-table-column prop="tag" label="版本" />
            <el-table-column label="大小">

                <template #default="scope">
                    <el-text>{{ showSize(scope.row) }}</el-text>
                </template>
            </el-table-column>
            <el-table-column prop="create_time" label="创建时间" />
            <el-table-column label="详细信息">

                <template #default="scope">
                    <el-button type="primary" link size="default" @click='onDetails(scope.row)'>
                        信息
                    </el-button>
                </template>
            </el-table-column>
            <el-table-column fixed="right" label="操作">

                <template #header>
                    <div class="flex">
                        <el-input input-style="color:#909399" v-model="image_name" :prefix-icon="Search" size="small"
                            placeholder="查询官网镜像" @keyup.enter="onQuery" />
                        <el-button class="ml-2" type="primary" size="small" @click='onPull(image_name)'>
                            拉取
                        </el-button>
                    </div>
                </template>

                <template #default="scope">
                    <el-button type="success" link size="default" @click='onRun(scope.row)'>
                        运行
                    </el-button>
                    <el-button type="danger" link size="default" @click='onDel(scope.row)'>
                        删除
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-table class="!h-1/3" empty-text=" " :span-method="SpanMethod"  :data="pullLogData.layer">
            <el-table-column label="镜像" :width="200">
                <template #default="scope">
                    {{ pullLogData.name }}
                </template>
            </el-table-column>
            <el-table-column label="ID" :width="200" prop="id"></el-table-column>
            <el-table-column label="类型" :width="100" prop="status">
                <template #default="scope">
                    <el-icon v-if="scope.row.status === 'Downloading'">
                        <Download />
                    </el-icon>
                    <el-icon v-else-if="scope.row.status === 'Extracting'">
                        <Files />
                    </el-icon>
                </template>
            </el-table-column>
            <el-table-column label="进度">
                <template #default="scope">
                    <el-progress v-if="scope.row.total_size > 0"
                        :percentage="Math.floor((scope.row.cur_size / scope.row.total_size) * 100)" />
                </template>
            </el-table-column>
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { AsyncFetch } from '@/lib/comm';
import { WBSocket } from '@/lib/websocket'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

interface ImageInfo {
    repostitory: string
    tag: string
    id: string
    create_time: string
    size: number
}

interface ExportPortInfo {
    proto: string
    port: string
}

interface ImageDetailInfo {
    os: string
    osVersion: string
    size: number
    exposed_ports: ExportPortInfo[]
    volumes: string[]
    working_dir: string
    env: string[]
    cmd: string
}

interface ImagePullInfo {
    star_count: number
    is_official: boolean
    name: string
    is_automated: boolean
    description: string
}

interface PullLayerInfo {
    id: string
    status: string
    cur_size: number
    total_size: number
}

interface PullLogInfo {
    name: string
    err: string
    layer: PullLayerInfo[]
}

const imageDatas = ref<ImageInfo[]>([])
const imageDetails = ref<ImageDetailInfo>({} as ImageDetailInfo)

const detailIsShow = ref(false)
const detailTitle = ref<string>('')

const runDlgIsShow = ref(false)
const dataLoding = ref(false)

const image_name = ref('')
const image_infos = ref<ImagePullInfo[]>([])
const image_infos_dlg_show = ref(false)
const image_infos_loading = ref(false)

const pullLogData = ref<PullLogInfo>({
    name: '',
    err: '',
    layer: [] as PullLayerInfo[],
})

let websocket: WBSocket | null = null

const props = defineProps<{
    group: string
}>()

function SpanMethod(data:{row: any, column: any, rowIndex: number, columnIndex: number}) : number[]{
    if (data.columnIndex == 0) {
        if (data.rowIndex != 0) {
            return [0, 0]
        }

        return [pullLogData.value.layer.length, 1]
    }

    return [1, 1]
}

function onQuery() {
    let name = image_name.value.trim()

    if (name.length == 0) {
        ElMessage.error("请输入镜像")
        return
    }

    image_infos.value.length = 0
    image_infos_loading.value = true
    image_infos_dlg_show.value = true

    AsyncFetch<ImagePullInfo[]>(`${props.group}queryImage?name=${name}`, null).then((infos) => {
        image_infos.value = infos
        image_infos_loading.value = false
    }).catch(() => {
        image_infos_loading.value = false
    })
}

function onPull(name: string) {
    name = image_name.value.trim()

    if (name.length == 0) {
        ElMessage.error("请输入镜像")
        return
    }

    AsyncFetch<void>(`${props.group}pullImage?name=${name}`, null).then((infos) => {
    })
}

function onDel(row: ImageInfo) {
    ElMessageBox.confirm(
        `是否删除镜像：${row.repostitory}`,
        `删除`,
        {
            confirmButtonText: '删除',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {
        dataLoding.value = true
        AsyncFetch(`${props.group}delImage?id=${row.id}`, null).then(() => {
            ElMessage.success(`删除镜像 ${row.repostitory} 成功`)
            imageDatas.value = imageDatas.value.filter(item => item.id != row.id)
            dataLoding.value = false
        }).catch(err => {
            dataLoding.value = false
        })
    }).catch(() => {
    })
}

function onRun(row: ImageInfo) {
    runDlgIsShow.value = true
}

function onDetails(row: ImageInfo) {
    AsyncFetch<ImageDetailInfo>(`${props.group}getImageDetails?id=${row.id}`, null).then((infos) => {
        detailTitle.value = row.repostitory
        imageDetails.value = infos
        detailIsShow.value = true
    })
}

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

function getPullLog() {
    websocket = new WBSocket(6)

    websocket.SetMsgFun((event: MessageEvent) => {
        let msg = event.data.toString()
        pullLogData.value = JSON.parse(msg)
    })

    websocket.Conn(`ws://${window.location.host}/${props.group}getPullImageLog`)
}

onMounted(() => {
    getImages()
    getPullLog()
})

</script>

<style>
.image_info_dlg .el-dialog__body {
    height: calc(100% - 30px);
}
</style>