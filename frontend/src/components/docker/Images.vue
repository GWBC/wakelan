<template>
    <el-dialog v-model="runDlgIsShow" append-to-body title="创建容器" @open="onRunDlgOpen">
        <el-form v-loading="loading" :model="dockerContainerCreate">
            <el-form-item label="容器名称">
                <el-input placeholder="容器名为空，则使用随机名称" v-model="dockerContainerCreate.name" />
            </el-form-item>
            <el-form-item label="重启策略">
                <el-select v-model="dockerContainerCreate.restart_policy" placeholder="选择异常退出后重启策略" size="large"
                    style="width: 240px">
                    <el-option key="no" label="no" value="no" />
                    <el-option key="always" label="always" value="always" />
                    <el-option key="on-failure" label="on-failure" value="on-failure" />
                    <el-option key="unless-stopped" label="unless-stopped" value="unless-stopped" />
                </el-select>
            </el-form-item>
            <el-form-item label="镜像名称">
                <el-input readonly v-model="dockerContainerCreate.image" />
            </el-form-item>
            <el-form-item label="执行命令">
                <el-input v-model="dockerContainerCreate.cmd" />
            </el-form-item>
            <el-form-item label="网络名称">
                <el-select v-model="dockerContainerCreate.net_name" placeholder="请选择网络">
                    <el-option v-for="item in dockerNetworkName" :key="item.name" :label="item.name"
                        :value="item.name"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="端口映射">
                <div class="flex flex-wrap gap-2 w-full">
                    <el-tag v-for="tag in dockerContainerCreate.ports" :key="tag" closable :disable-transitions="false"
                        @close="onDlgPortsClose(tag)">
                        {{ tag }}
                    </el-tag>
                    <el-tooltip content="主机:容器/协议，主机-主机:容器-容器/协议" placement="top-start">
                        <el-input v-if="dlgPortInputShow" class="!w-40" ref="dlgPortInputRef" v-model="dlgPortInput"
                            size="small" @keyup.enter="onPortInputConfirm" @blur="onPortInputConfirm" />
                        <el-button type="primary" v-else size="small" @click="showPortInput">
                            <el-icon>
                                <Plus />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-button class="!ml-0" type="danger" size="small" @click="onPortClear">
                        <el-icon>
                            <Close />
                        </el-icon>
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item label="目录映射">
                <div class="flex flex-wrap gap-2 w-full">
                    <el-tag v-for="tag in dockerContainerCreate.mounts" :key="tag" closable :disable-transitions="false"
                        @close="onDlgMountClose(tag)">
                        {{ tag }}
                    </el-tag>
                    <el-tooltip content="主机:容器" placement="top-start">
                        <el-input v-if="dlgMountInputShow" class="!w-40" ref="dlgMountInputRef" v-model="dlgMountInput"
                            size="small" @keyup.enter="onMountInputConfirm" @blur="onMountInputConfirm" />

                        <el-button type="primary" v-else size="small" @click="showMountInput">
                            <el-icon>
                                <Plus />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-button class="!ml-0" type="danger" size="small" @click="onMountClear">
                        <el-icon>
                            <Close />
                        </el-icon>
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item label="开启特权">
                <el-switch v-model="dockerContainerCreate.privileged" />
            </el-form-item>
            <el-form-item label="自动删除">
                <el-switch v-model="dockerContainerCreate.auto_remove" />
            </el-form-item>
            <el-form-item>
                <div class="flex justify-end">
                    <el-button @click="runDlgIsShow = false">取消</el-button>
                    <el-button type="primary" @click="onRunDlgOk">
                        运行
                    </el-button>
                </div>
            </el-form-item>
        </el-form>
    </el-dialog>

    <el-dialog class="image_info_dlg h-2/3" v-model="image_infos_dlg_show">
        <el-table v-loading="loading" class="!h-full" :data="image_infos" stripe>
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

    <el-dialog v-model="detailIsShow" :title="detailTitle + ' 详情'">
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
                    <!-- <el-button type="success" link size="default" @click='onRun(scope.row)'>
                        推送
                    </el-button> -->
                    <el-button type="danger" link size="default" @click='onDel(scope.row)'>
                        删除
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
        <el-table class="!h-1/3" empty-text=" " :span-method="SpanMethod" :data="pullLogData.layer">
            <el-table-column label="镜像" :width="200">
                <template #default="scope">
                    {{ pullLogData.name }}
                </template>
            </el-table-column>
            <el-table-column label="ID" :width="200" prop="id"></el-table-column>
            <el-table-column label="类型" :width="100" prop="status">
                <template #default="scope">
                    <span class="text-red-500 font-bold" v-if="scope.row.id === 'Error'">
                        {{ scope.row.status }}
                    </span>
                    <span class="text-green-500 font-bold" v-else-if="scope.row.id === 'Success'">
                        {{ scope.row.status }}
                    </span>
                    <el-icon v-else-if="scope.row.status === 'Downloading'">
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
import { onMounted, ref, nextTick } from 'vue'
import { AsyncFetch } from '@/lib/comm';
import { WBSocket } from '@/lib/websocket'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'

const xx = ref(true)

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
    refresh: boolean
    name: string
    layer: PullLayerInfo[]
}

interface DockerContainerCreate {
    name: string                        //容器名称
    restart_policy: string              //重启策略
    image: string                       //镜像名称
    cmd: string                         //执行命令
    privileged: boolean                 //开启特权
    net_name: string                    //网络名称
    ports: string[]                     //端口映射 public:private/proto，public-public:private-private/proto
    mounts: string[]                    //目录映射 public:private
    auto_remove: boolean                //自动删除
}

interface NetworkCardInfo {
    name: string
    id: string
}

const loading = ref(false)
const imageDatas = ref<ImageInfo[]>([])
const imageDetails = ref<ImageDetailInfo>({} as ImageDetailInfo)

const detailIsShow = ref(false)
const detailTitle = ref<string>('')

const runDlgIsShow = ref(false)
const dataLoding = ref(false)

const image_name = ref('')
const image_infos = ref<ImagePullInfo[]>([])
const image_infos_dlg_show = ref(false)

const dlgPortInputShow = ref(false)
const dlgPortInputRef = ref()
const dlgPortInput = ref('')

const dlgMountInputShow = ref(false)
const dlgMountInputRef = ref()
const dlgMountInput = ref('')


const dockerNetworkName = ref<NetworkCardInfo[]>([])
const dockerContainerCreate = ref<DockerContainerCreate>({} as DockerContainerCreate)

const pullLogData = ref<PullLogInfo>({
    refresh: false,
    name: '',
    layer: [] as PullLayerInfo[],
})

let websocket: WBSocket | null = null

const props = defineProps<{
    group: string
}>()

function onRunDlgOk() {
    loading.value = true
    AsyncFetch<NetworkCardInfo[]>(`${props.group}runContainer`, dockerContainerCreate.value).then((infos) => {
        ElMessage.success('创建容器成功')
        loading.value = false
        runDlgIsShow.value = false
    }).catch(() => {
        loading.value = false
    })
}

function onPortClear() {
    dockerContainerCreate.value.ports.length = 0
}

function onMountClear() {
    dockerContainerCreate.value.mounts.length = 0
}

function showPortInput() {
    dlgPortInputShow.value = true
    nextTick(() => {
        dlgPortInputRef.value!.input!.focus()
    })
}

function onPortInputConfirm() {
    if (dlgPortInput.value) {
        dockerContainerCreate.value.ports.push(dlgPortInput.value)
    }

    dlgPortInputShow.value = false
    dlgPortInput.value = ''
}

function onDlgPortsClose(mount: string) {
    let index = dockerContainerCreate.value.ports.indexOf(mount)
    if (index != -1) {
        dockerContainerCreate.value.ports.splice(index, 1)
    }
}

function showMountInput() {
    dlgMountInputShow.value = true
    nextTick(() => {
        dlgMountInputRef.value!.input!.focus()
    })
}

function onMountInputConfirm() {
    if (dlgMountInput.value) {
        dockerContainerCreate.value.mounts.push(dlgMountInput.value)
    }

    dlgMountInputShow.value = false
    dlgMountInput.value = ''
}

function onDlgMountClose(port: string) {
    let index = dockerContainerCreate.value.mounts.indexOf(port)
    if (index != -1) {
        dockerContainerCreate.value.mounts.splice(index, 1)
    }
}

function onRunDlgOpen() {
    AsyncFetch<NetworkCardInfo[]>(`${props.group}getNetworkCards`, null).then((infos) => {
        dockerNetworkName.value = infos
    })
}

function SpanMethod(data: { row: any, column: any, rowIndex: number, columnIndex: number }): number[] {
    if (data.columnIndex == 0) {
        if (data.rowIndex != 0) {
            return [0, 0]
        }

        return [pullLogData.value.layer.length, 1]
    }

    if (data.row.id == "Error" || data.row.id == "Success") {
        if (data.columnIndex == 0) {
            return [1, 1]
        } else if (data.columnIndex == 1 || data.columnIndex == 3) {
            return [0, 0]
        } else {
            return [1, 3]
        }
    }

    return [1, 1]
}

function onQuery() {
    let name = image_name.value.trim()

    if (name.length == 0) {
        ElMessage.error("请输入镜像")
        return
    }

    loading.value = true
    image_infos.value.length = 0
    image_infos_dlg_show.value = true

    AsyncFetch<ImagePullInfo[]>(`${props.group}queryImage?name=${name}`, null).then((infos) => {
        image_infos.value = infos
        loading.value = false
    }).catch(() => {
        loading.value = false
    })
}

function onPull(name: string) {
    let tName = name.trim()

    if (tName.length == 0) {
        ElMessage.error("请输入镜像")
        return
    }

    AsyncFetch<void>(`${props.group}pullImage?name=${tName}`, null).then((infos) => {
        ElMessage.success(`异步拉取镜像 ${tName} 成功`)
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
    dockerContainerCreate.value = {
        name: '',
        restart_policy: 'always',
        image: '',
        cmd: '',
        privileged: true,
        net_name: 'bridge',
        ports: [] as string[],
        mounts: [] as string[],
        auto_remove: false,
    }

    AsyncFetch<ImageDetailInfo>(`${props.group}getImageDetails?id=${row.id}`, null).then((infos) => {
        dockerContainerCreate.value.image = row.repostitory + ":" + row.tag
        dockerContainerCreate.value.cmd = infos.cmd
        infos.exposed_ports?.forEach((value, key) => {
            let pubPort = value.port
            if (pubPort == "80") {
                pubPort = "8080"
            }else if(pubPort == "443"){
                pubPort = "8443"
            }

            dockerContainerCreate.value.ports.push(`${pubPort}:${value.port}/${value.proto}`)
        })

        dockerContainerCreate.value.ports

        infos.volumes?.forEach((value, key) => {
            dockerContainerCreate.value.mounts.push(`${value}:${value}`)
        })

        runDlgIsShow.value = true
    })
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
        if (pullLogData.value.refresh) {
            getImages()
        }
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