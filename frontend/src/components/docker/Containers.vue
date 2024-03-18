<template>
    <el-dialog v-model="editDlg" @open="onDlgOpen">
        <el-form :model="editDlgData">
            <el-form-item label="容器名">
                <el-text>{{ editDlgData.name }}</el-text>
            </el-form-item>
            <el-form-item label="新名称">
                <el-input ref="editDlgNewName" :formatter="newFormat" placeholder="只能输入数字、字母、下划线"
                    v-model="editDlgData.new_name" />
            </el-form-item>
            <el-form-item>
                <el-button class="ml-auto" type="primary"
                    @click="onEditModify(editDlgData.name, editDlgData.new_name)">修改</el-button>
            </el-form-item>
        </el-form>
    </el-dialog>

    <el-dialog v-model="details" :title="detailsInfo.name + ' 详情'" class="container_details_dlg !h-1/2">
        <el-tabs class="ml-10 mr-10 h-full" tab-position="left">
            <el-tab-pane class="!h-full" label="网络配置">
                <el-table class="!h-full" table-layout="auto" :data="detailsInfo.networks" empty-text=" ">
                    <el-table-column prop="name" label="名称" />
                    <el-table-column prop="mac" label="硬件地址" />
                    <el-table-column prop="ip" label="地址" />
                    <el-table-column prop="gateway" label="网关" />
                    <el-table-column prop="dns" label="域名" />
                </el-table>
            </el-tab-pane>
            <el-tab-pane class="h-full" label="目录映射">
                <el-table class="!h-full" table-layout="auto" show-overflow-tooltip :data="detailsInfo.volume_info"
                    empty-text=" ">
                    <el-table-column prop="type" label="类型" />
                    <el-table-column prop="name" label="名称" />
                    <el-table-column prop="source" label="宿主目录" />
                    <el-table-column prop="destination" label="容器目录" />
                </el-table>
            </el-tab-pane>
            <el-tab-pane class="h-full" label="IPv4映射">
                <el-table class="!h-full" table-layout="auto" :data="detailsInfo.v4ports" empty-text=" ">
                    <el-table-column prop="type" label="协议" />
                    <el-table-column prop="ip" label="宿主地址" />
                    <el-table-column prop="public_port" label="宿主端口" />
                    <el-table-column prop="private_port" label="容器端口" />
                    <el-table-column label="">
                        <template #default="scope">
                            <div class="flex gap-2">
                                <el-link type="primary" target="_blank"
                                    :href="calcURL(detailsInfo.server_ip, scope.row.public_port, 'http')">
                                    HTTP
                                </el-link>
                                <el-link type="primary" target="_blank"
                                    :href="calcURL(detailsInfo.server_ip, scope.row.public_port, 'https')">
                                    HTTPS
                                </el-link>
                            </div>
                        </template>
                    </el-table-column>
                </el-table>
            </el-tab-pane>
            <el-tab-pane class="h-full" label="IPv6映射">
                <el-table class="!h-full" table-layout="auto" :data="detailsInfo.v6ports" empty-text=" ">
                    <el-table-column prop="type" label="协议" />
                    <el-table-column prop="ip" label="宿主地址" />
                    <el-table-column prop="public_port" label="宿主端口" />
                    <el-table-column prop="private_port" label="容器端口" />
                    <el-table-column label="">
                        <template #default="scope">
                            <div class="flex gap-2">
                                <el-link type="primary" target="_blank"
                                    :href="calcURL(detailsInfo.server_ip, scope.row.public_port, 'http')">
                                    HTTP
                                </el-link>
                                <el-link type="primary" target="_blank"
                                    :href="calcURL(detailsInfo.server_ip, scope.row.public_port, 'https')">
                                    HTTPS
                                </el-link>
                            </div>
                        </template>
                    </el-table-column>
                </el-table>
            </el-tab-pane>
        </el-tabs>
    </el-dialog>

    <Terminal v-model="logShow" :title="logTitle" :url="logUrl" />
    <Terminal v-model="termShow" :title="termTitle" :url="termUrl" :only-read="false" :auto-exit="true" />

    <el-card v-loading="dataLoding" element-loading-background="rgba(255, 255, 255, 0.5)" :element-loading-svg="dataSvg"
        element-loading-svg-view-box="-10, -10, 50, 50" class="h-full" body-class="h-full !pb-1">
        <el-table table-layout="auto" class="!h-full" :data="containerDatas" :row-key="(row: ContainerInfo) => row.id"
            empty-text=" " stripe>
            <el-table-column label="" width="40">

                <template #default="scope">
                    <div v-if="scope.row.state != 'running'" class="w-3 h-3 rounded-full bg-red-400"></div>
                    <div v-else class="w-3 h-3 rounded-full bg-green-400"></div>
                </template>
            </el-table-column>
            <el-table-column prop="id" label="ID" />
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="image" label="镜像" />
            <el-table-column prop="cmd" label="命令" />
            <el-table-column prop="run_time" label="运行时间" />
            <el-table-column prop="create_time" label="创建时间" />

            <el-table-column label="详细信息">

                <template #default="scope">
                    <div class="container_op_btn">
                        <el-button type="primary" link size="default" @click='onDetails(scope.row)'>
                            信息
                        </el-button>
                    </div>
                </template>
            </el-table-column>

            <el-table-column label="操作" fixed="right">

                <template #default="scope">
                    <div class="container_op_btn">
                        <el-button :type="scope.row.state == 'running' ? 'success' : 'info'" link size="default"
                            :disabled="scope.row.state != 'running'" @click="onTerminal(scope.row)">
                            终端
                        </el-button>
                        <el-button type="success" link size="default" @click="onLog(scope.row)">
                            日志
                        </el-button>
                        <el-button v-if="scope.row.state == 'running'" type="success" link size="default"
                            @click="onStop(scope.row)">
                            停止
                        </el-button>
                        <el-button v-else type="success" link size="default" @click="onStart(scope.row)">
                            启动
                        </el-button>
                        <el-button type="success" link size="default" @click="onRestart(scope.row)">
                            重启
                        </el-button>
                        <el-button type="success" link size="default" @click="onEdit(scope.row)">
                            修改
                        </el-button>
                        <el-button type="success" link size="default" @click="onBackup(scope.row)">
                            备份
                        </el-button>
                        <el-button type="danger" link size="default" @click="onDelete(scope.row)">
                            删除
                        </el-button>
                    </div>
                </template>
            </el-table-column>
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, nextTick } from 'vue'
import { AsyncFetch } from '@/lib/comm'
import Terminal from '@/components/docker/Terminal.vue'
import { ElMessageBox, ElMessage } from 'element-plus'

interface PortInfo {
    ip: string
    private_port: number
    public_port: number
    type: string
}

interface NetInfo {
    name: string
    mac: string
    gateway: string
    ip: string
    dns: string[]
}

interface VolumeInfo {
    type: string
    name: string
    source: string
    destination: string
}

interface ContainerInfo {
    id: string
    name: string
    image: string
    cmd: string
    v4ports: PortInfo[]
    v6ports: PortInfo[]
    networks: NetInfo[]
    volume_info: VolumeInfo[]
    run_time: string
    state: string
    create_time: string
    server_ip: string
}

interface EditData {
    name: string
    new_name: string
}

const editDlg = ref(false)
const editDlgNewName = ref()
const editDlgData = reactive<EditData>({} as EditData)
let editDlgPreInput = ''

const logShow = ref(false)
const logUrl = ref('')
const logTitle = ref('')

const termShow = ref(false)
const termUrl = ref('')
const termTitle = ref('')

const containerDatas = ref<ContainerInfo[]>([])
const props = defineProps<{ group: string }>()

const dataLoding = ref(false)
const dataSvg = ref(`
<path class="path" d="
          M 30 15
          L 28 17
          M 25.61 25.61
          A 15 15, 0, 0, 1, 15 30
          A 15 15, 0, 1, 1, 27.99 7.5
          L 15 15
        " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>
`)

const details = ref(false)
const detailsInfo = ref<ContainerInfo>({} as ContainerInfo)

function calcURL(host: string, pubPort: string, proto: string) {
    if (host.length == 0) {
        host = window.location.hostname
    }

    return `${proto}://${host}:${pubPort}`
}

function onDetails(row: ContainerInfo) {
    details.value = true
    detailsInfo.value = row
}

function onStop(row: ContainerInfo) {
    ElMessageBox.confirm(
        `是否停止容器：${row.name}`,
        `停止`,
        {
            confirmButtonText: '停止',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {
        dataLoding.value = true
        AsyncFetch(`${props.group}operContainer?name=${row.name}&oper=stop`, null).then(() => {
            ElMessage.success(`停止容器 ${row.name} 成功`)
            containerDatas.value.map(item => {
                if (item.name == row.name) {
                    item.state = 'exited'
                }
                return item
            })

            dataLoding.value = false
        }).catch(err => {
            dataLoding.value = false
        })
    }).catch(() => {
    })
}

function onStart(row: ContainerInfo) {
    dataLoding.value = true
    AsyncFetch(`${props.group}operContainer?name=${row.name}&oper=start`, null).then(() => {
        ElMessage.success(`启动容器 ${row.name} 成功`)
        containerDatas.value.map(item => {
            if (item.name == row.name) {
                item.state = 'running'
            }
            return item
        })
        dataLoding.value = false
    }).catch(err => {
        dataLoding.value = false
    })
}

function onRestart(row: ContainerInfo) {
    ElMessageBox.confirm(
        `是否重启容器：${row.name}`,
        `重启`,
        {
            confirmButtonText: '重启',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {
        dataLoding.value = true
        AsyncFetch(`${props.group}operContainer?name=${row.name}&oper=restart`, null).then(() => {
            ElMessage.success(`重启容器 ${row.name} 成功`)
            containerDatas.value.map(item => {
                if (item.name == row.name) {
                    item.state = 'running'
                }
                return item
            })
            dataLoding.value = false
        }).catch(err => {
            dataLoding.value = false
        })
    }).catch(() => {
    })
}

function onDlgOpen() {
    nextTick(() => {
        //由于dlg自动焦点在标题上，需要自己实现获取焦点
        editDlgNewName.value.focus()
    })
}

function newFormat(value: string | number): string {
    if (/^[a-zA-Z0-9_]+$/.test(value.toString())) {
        editDlgPreInput = value.toString()
        return editDlgPreInput
    }

    return editDlgPreInput
}

function onTerminal(row: ContainerInfo) {
    termTitle.value = `${row.name} 终端`
    termUrl.value = `ws://${window.location.host}/${props.group}enterContainer?name=${row.name}`
    termShow.value = true
}

function onLog(row: ContainerInfo) {
    logTitle.value = `${row.name} 日志`
    logUrl.value = `ws://${window.location.host}/${props.group}getLogsContainer?name=${row.name}`
    logShow.value = true
}

function onBackup(row: ContainerInfo) {
    let name = row.name
    dataLoding.value = true
    AsyncFetch(`${props.group}backupContainer?name=${name}`, null).then(() => {
        ElMessage.success(`备份容器 ${name} 成功`)
        dataLoding.value = false
    }).catch(err => {
        dataLoding.value = false
    })
}

function onEditModify(name: string, newName: string) {
    if (name == newName) {
        return
    }

    if (containerDatas.value.find(item => item.name == newName)) {
        ElMessage.error(`容器名 ${newName} 已存在`)
        return
    }

    dataLoding.value = true
    AsyncFetch(`${props.group}renameContainer?old=${name}&new=${newName}`, null).then(() => {
        ElMessage.success(`容器 ${name} 重命名成功`)
        containerDatas.value = containerDatas.value.map(item => {
            if (item.name == name) {
                item.name = newName
            }
            return item
        })
        editDlg.value = false
        dataLoding.value = false
    }).catch(err => {
        dataLoding.value = false
    })
}

function onEdit(row: ContainerInfo) {
    editDlgData.name = row.name
    editDlgData.new_name = ''
    editDlgPreInput = ''
    editDlg.value = true
}

function onDelete(row: ContainerInfo) {
    ElMessageBox.confirm(
        `是否删除容器：${row.name}`,
        `删除`,
        {
            confirmButtonText: '删除',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {
        dataLoding.value = true
        AsyncFetch(`${props.group}delContainer?name=${row.name}`, null).then(() => {
            ElMessage.success(`删除容器 ${row.name} 成功`)
            containerDatas.value = containerDatas.value.filter(item => item.name != row.name)
            dataLoding.value = false
        }).catch(err => {
            dataLoding.value = false
        })
    }).catch(() => {
    })
}

function getDatas() {
    dataLoding.value = true
    AsyncFetch<ContainerInfo[]>(`${props.group}getContainers`, null).then((infos) => {
        containerDatas.value = infos
        dataLoding.value = false
    }).catch(err => {
        dataLoding.value = false
    })
}

onMounted(() => {
    getDatas()
})

</script>

<style>
.port-dlg .el-dialog__body {
    height: calc(100% - 60px);
    padding: 0px;
}

.log-dlg {
    padding: 6px;
    overflow: hidden !important;
}

.log-dlg .el-dialog__header {
    height: 30px;
    margin: 0px;
    padding: 0px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.log-dlg .el-dialog__title {
    font-size: large;
    font-weight: bold;
}

.log-dlg .el-dialog__headerbtn {
    height: 30px;
    width: 30px;
    margin-right: 10px;
}

.log-dlg .el-dialog__body {
    height: calc(100% - 30px);
    margin: 0px;
    padding: 0px;
}

.el-tabs__item {
    color: gray;
}

.container_op_btn button {
    margin: 2px !important;
}

.container_details_dlg .el-dialog__body {
    padding-top: 0px;
    height: calc(100% - 60px);
}

.container_details_dlg thead {
    position: sticky;
    top: 0;
    z-index: calc(var(--el-table-index) + 2);
}
</style>
