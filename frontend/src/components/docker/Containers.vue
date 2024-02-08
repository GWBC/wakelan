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
    <Terminal v-model="logShow" :title="logTitle" :url="logUrl" :only-read="true"/>
    <Terminal v-model="termShow" :title="termTitle" :url="termUrl" :only-read="false"/>
    <el-card class="h-full" body-class="h-full !pb-1">
        <el-table class="!h-full" :data="containerDatas" :row-key="(row: ContainerInfo) => row.id"
            :expand-row-keys="expandRow" empty-text=" " stripe @expand-change="onClick" @row-click="onClick">
            <el-table-column type="expand">
                <template #default="scope">
                    <div class="ml-16 mr-16">
                        <el-collapse v-model="detailsShow">
                            <el-collapse-item name="operation">
                                <template #title>
                                    <span class="text-gray-500 text-sm font-bold">操作</span>
                                </template>
                                <el-button type="warning" @click="onTerminal(scope.row)">
                                    <el-icon size="16">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="ionicon" viewBox="0 0 512 512">
                                            <rect x="32" y="48" width="448" height="416" rx="48" ry="48" fill="none"
                                                stroke="currentColor" stroke-linejoin="round" stroke-width="32" />
                                            <path fill="none" stroke="currentColor" stroke-linecap="round"
                                                stroke-linejoin="round" stroke-width="32"
                                                d="M96 112l80 64-80 64M192 240h64" />
                                        </svg>
                                    </el-icon>
                                </el-button>
                                <el-button type="success" @click="onLog(scope.row)">
                                    <el-icon size="16">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="ionicon" viewBox="0 0 512 512">
                                            <circle cx="128" cy="256" r="96" fill="none" stroke="currentColor"
                                                stroke-linecap="round" stroke-linejoin="round" stroke-width="32" />
                                            <circle cx="384" cy="256" r="96" fill="none" stroke="currentColor"
                                                stroke-linecap="round" stroke-linejoin="round" stroke-width="32" />
                                            <path fill="none" stroke="currentColor" stroke-linecap="round"
                                                stroke-linejoin="round" stroke-width="32" d="M128 352h256" />
                                        </svg>
                                    </el-icon>
                                </el-button>
                                <el-button type="primary" @click="onEdit(scope.row)">
                                    <el-icon size="16">
                                        <Edit />
                                    </el-icon>
                                </el-button>
                                <el-button type="danger" @click="onDelete(scope.row)">
                                    <el-icon size="16">
                                        <Delete />
                                    </el-icon>
                                </el-button>
                            </el-collapse-item>
                            <el-collapse-item name="network">
                                <template #title>
                                    <span class="text-gray-500 text-sm font-bold">网络</span>
                                </template>
                                <el-table :data="scope.row.networks" empty-text=" ">
                                    <el-table-column prop="name" label="名称" />
                                    <el-table-column prop="mac" label="硬件地址" />
                                    <el-table-column prop="ip" label="地址" />
                                    <el-table-column prop="gateway" label="网关" />
                                    <el-table-column prop="dns" label="域名" />
                                </el-table>
                            </el-collapse-item>
                            <el-collapse-item name="volume">
                                <template #title>
                                    <span class="text-gray-500 text-sm font-bold">目录映射</span>
                                </template>
                                <el-table :data="scope.row.volume_info" empty-text=" ">
                                    <el-table-column prop="type" label="类型" />
                                    <el-table-column prop="name" label="名称" />
                                    <el-table-column prop="source" label="宿主目录" />
                                    <el-table-column prop="destination" label="容器目录" />
                                </el-table>
                            </el-collapse-item>
                            <el-collapse-item v-if="scope.row.v4ports.length > 0" name="v4">
                                <template #title>
                                    <span class="text-gray-500 text-sm font-bold">v4端口映射</span>
                                </template>
                                <el-table :data="scope.row.v4ports" empty-text=" ">
                                    <el-table-column prop="type" label="协议" />
                                    <el-table-column prop="ip" label="宿主地址" />
                                    <el-table-column prop="public_port" label="宿主端口" />
                                    <el-table-column prop="private_port" label="容器端口" />
                                </el-table>
                            </el-collapse-item>
                            <el-collapse-item v-if="scope.row.v6ports.length > 0" name="v6">
                                <template #title>
                                    <span class="text-gray-500 text-sm font-bold">v6端口映射</span>
                                </template>
                                <el-table :data="scope.row.v6ports" empty-text=" ">
                                    <el-table-column prop="type" label="协议" />
                                    <el-table-column prop="ip" label="宿主地址" />
                                    <el-table-column prop="public_port" label="宿主端口" />
                                    <el-table-column prop="private_port" label="容器端口" />
                                </el-table>
                            </el-collapse-item>
                        </el-collapse>
                    </div>
                </template>
            </el-table-column>
            <el-table-column label="" width="40">
                <template #default="scope">
                    <div v-if="scope.row.state != 'running'" class="w-3 h-3 rounded-full bg-red-400"></div>
                    <div v-else class="w-3 h-3 rounded-full bg-green-400"></div>
                </template>
            </el-table-column>
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="image" label="镜像" />
            <el-table-column prop="cmd" label="命令" />
            <el-table-column prop="run_time" label="运行时间" />
            <el-table-column prop="create_time" label="创建时间" />
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, nextTick } from 'vue'
import { AsyncFetch } from '@/lib/comm'
import Terminal from '@/components/docker/Terminal.vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { Edit, Delete } from '@element-plus/icons-vue'

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

const expandRow = ref<string[]>([])
const detailsShow = ref(["network", "operation", "volume", "v4", "v6"])

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
    termUrl.value = `ws://${window.location.host}/${props.group}enterContainer?name=${row.name}&rows=100&cols=200`
    termShow.value = true
}

function onLog(row: ContainerInfo) {
    logTitle.value = `${row.name} 日志`
    logUrl.value = `ws://${window.location.host}/${props.group}getLogsContainer?name=${row.name}&rows=100&cols=200`
    logShow.value = true
}

function onEditModify(name: string, newName: string) {
    if (name == newName) {
        return
    }

    if (containerDatas.value.find(item => item.name == newName)) {
        ElMessage.error(`容器名 ${newName} 已存在`)
        return
    }

    AsyncFetch(`${props.group}renameContainer?old=${name}&new=${newName}`, null).then(() => {
        ElMessage.success(`容器 ${name} 重命名成功`)
        containerDatas.value = containerDatas.value.map(item => {
            if (item.name == name) {
                item.name = newName
            }
            return item
        })
        editDlg.value = false
    }).catch(err => {
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
        `是否删除容：${row.name}`,
        `删除`,
        {
            confirmButtonText: '删除',
            cancelButtonText: '取消',
            type: 'warning',
        }
    ).then(() => {
        AsyncFetch(`${props.group}delContainer?name=${row.name}`, null).then(() => {
            ElMessage.success(`删除容器 ${row.name} 成功`)
            containerDatas.value = containerDatas.value.filter(item => item.name != row.name)
        }).catch(err => {
        })
    }).catch(() => {
    })
}

function onClick(row: ContainerInfo) {
    const index = expandRow.value.indexOf(row.id)
    if (index != -1) {
        //收缩
        expandRow.value.length = 0
        return
    }

    //展开
    expandRow.value.length = 0
    expandRow.value.push(row.id)
}

function getDatas() {
    AsyncFetch<ContainerInfo[]>(`${props.group}getContainers`, null).then((infos) => {
        containerDatas.value = infos
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
</style>
