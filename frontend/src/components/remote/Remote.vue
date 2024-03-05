<template>
    <el-dialog ref="fsDlg" @keydown="onFSKeyDown" @dragover.prevent v-model="fsDlgShow" :append-to-body="true">
        <template #header>
            <el-breadcrumb class="breadcrumb" separator=">">
                <el-breadcrumb-item v-for="(item, index) in fsCurPathObj" @click='changeDir(item.path)'>
                    <a href="#">{{ item.name }}</a>
                </el-breadcrumb-item>
            </el-breadcrumb>
        </template>
        <el-table v-loading="fsDlgLoading" :data="fsDlgDataObj" @row-dblclick="rowFSDBClick"
            :default-sort="{ prop: 'type', order: 'ascending' }" stripe :max-width="fsDlgMaxWidth"
            :max-height="fsDlgMaxHeight" empty-text=" ">
            <el-table-column min-width="260" prop="name" label="名称" sortable>
                <template #default="scope">
                    <el-icon v-if="scope.row.type == 'file'" size="18" color="blue">
                        <Tickets />
                    </el-icon>
                    <el-icon v-else size="18" color="green">
                        <FolderOpened />
                    </el-icon>
                    {{ scope.row.name }}
                </template>
            </el-table-column>
            <el-table-column prop="type" label="类型" sortable>
                <template #default="scope">
                    <span>{{ scope.row.type == 'dir' ? '目录' : '文件' }}</span>
                </template>
            </el-table-column>
        </el-table>
        <el-collapse v-model="fsDlgLog" accordion>
            <el-collapse-item title="日志" name="log">
                <el-table :data="fsDlgLogData" stripe :max-width="fsDlgMaxWidth" :max-height=150 empty-text=" ">
                    <el-table-column width="200" prop="time" label="时间" />
                    <el-table-column width="100" prop="cmd" label="操作" />
                    <el-table-column prop="msg" label="结果" />
                </el-table>
            </el-collapse-item>
        </el-collapse>
    </el-dialog>

    <ContrlButton v-if="showRemote">
        <el-button class="ctrl-button" type="info" @click="onSendCtrlAltDel">
            <el-icon size="20">
                <SwitchFilled />
            </el-icon>
            <span>CTRL+ALT+DEL</span>
        </el-button>
        <el-button v-show="showFilesystem" class="ctrl-button" type="info" @click="onFilesystem">
            <el-icon size="20">
                <Folder />
            </el-icon>
        </el-button>
        <el-button class="ctrl-button" type="info" @click="onFullScreenOrRecover">
            <el-icon size="20">
                <FullScreen />
            </el-icon>
        </el-button>
        <el-button class="ctrl-button" type="danger" @click="onCloseConn">
            <el-icon size="20">
                <Close />
            </el-icon>
        </el-button>
    </ContrlButton>

    <el-dialog class="remoteDlg" v-model="showRemote" :close-on-press-escape="false" :fullscreen="true" :show-close="false"
        :append-to-body="true" destroy-on-close @open="onOpen" @close="onClose">
        <div ref="viewport" class="viewport" v-loading="isLoading" element-loading-text="连接中..."
            :element-loading-spinner="loadingSVG" element-loading-svg-view-box="-10, -10, 50, 50"
            element-loading-background="rgba(0, 0, 0, 0.7)">
            <div :style="drawerStyle" ref="drawer" @mouseenter="onDrawerEnter" @mouseleave="onDrawerLeave">
                <el-button-group>
                    <el-button class="ctrl-button" type="info" @click="onSendCtrlAltDel">
                        <el-icon size="20">
                            <SwitchFilled />
                        </el-icon>
                        <span>CTRL+ALT+DEL</span>
                    </el-button>
                    <el-button v-show="showFilesystem" class="ctrl-button" type="info" @click="onFilesystem">
                        <el-icon size="20">
                            <Folder />
                        </el-icon>
                    </el-button>
                    <el-button class="ctrl-button" type="info" @click="onFullScreenOrRecover">
                        <el-icon size="20">
                            <FullScreen />
                        </el-icon>
                    </el-button>
                    <el-button class="ctrl-button" type="danger" @click="onCloseConn">
                        <el-icon size="20">
                            <Close />
                        </el-icon>
                    </el-button>
                </el-button-group>
            </div>

            <!--远程绘制div，必须使用tabindex，否则元素不支持焦点 -->
            <div ref="display" class="display" :style="displayStyle" />
        </div>
    </el-dialog>
</template>
  
<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'
import { FullScreenOrRecover, ExitFullScreen, Now2Str } from '@/lib/comm';
import { GuacdClient, GuacdViewPort } from "@/lib/guacd/guacd"
import { onMounted, ref, reactive, onUnmounted, computed } from 'vue'
import { Close, FullScreen, Folder, FolderOpened, Tickets, SwitchFilled } from "@element-plus/icons-vue"

import type { SFTPFileInfo } from '@/lib/guacd/filesystem'
import type { RemoteConfigInfo } from '@/lib/guacd/client'
import ContrlButton from '@/components/ContrlButton.vue'

interface FSInfo {
    type: string
    name: string
    path: string
}

interface FSLogInfo {
    cmd: string
    time: string
    msg?: string
}

const props = defineProps<{
    modelValue: boolean,
    connInfo: RemoteConfigInfo
}>()

const emit = defineEmits(['update:modelValue'])

const showRemote = computed({
    get() {
        return props.modelValue
    },
    set(value) {
        emit('update:modelValue', value)
    }
})

const viewport = ref()
const isLoading = ref(false)
const loadingSVG = `<path class="path" d="
    M 30 15
    L 28 17
    M 25.61 25.61
    A 15 15, 0, 0, 1, 15 30
    A 15 15, 0, 1, 1, 27.99 7.5
    L 15 15
  " style="stroke-width: 4px; fill: rgba(0, 0, 0, 0)"/>`

const display = ref()
const displayStyle = ref({})

const drawer = ref()
const showFilesystem = ref(false)
const drawerStyle = ref<Record<string, string | number>>({
    left: '0px',
    visibility: 'hidden',
    position: "absolute",
})

const fsDlg = ref()
const fsCurPath = ref("/")
const fsDlgShow = ref(false)
const fsDlgLoading = ref(false)
const fsDlgData = reactive<SFTPFileInfo[]>([])
const fsDlgMaxWidth = ref(0)
const fsDlgMaxHeight = ref(0)
const fsDlgLog = ref(['log'])
const fsDlgLogData = reactive<FSLogInfo[]>([])
const fsDlgSearchName = ref('')

//控制按键
const theme = ref('#F56C6C');
const position = ref('top right');
const column = ref(1);
const events = ref([
    {
        icon: '[ ]',
        text: '全屏',
        handle: (e: any) => onFullScreenOrRecover(),
    },
    {
        icon: 'M',
        text: '管理',
        handle: (e: any) => onSendCtrlAltDel(),
    },
    {
        icon: 'F',
        text: '文件',
        handle: (e: any) => onFilesystem(),
    },
    {
        icon: 'X',
        text: '关闭',
        handle: (e: any) => onCloseConn(),
    },
])

const fsDlgDataObj = computed(function () {
    try {
        let regex = new RegExp(fsDlgSearchName.value)
        return fsDlgData.filter(
            (data) => {
                if (!fsDlgSearchName.value) {
                    return true
                }

                return regex.test(data.name)
            }
        )
    } catch (error) {
        fsDlgSearchName.value = ''
        return fsDlgData
    }
})

const buttons = ref([
    {
        html: '<i class="fas fa-bars" />'
    },
    {
        html: '<i class="fas fa-compass" />',
        click: () => alert('hello')
    }
])

const fsCurPathObj = computed(function () {
    let ret: FSInfo[] = []
    let paths = fsCurPath.value.split("/")
    for (let i = 0; i < paths.length; ++i) {
        if (paths[i].length == 0) {
            if (i == 0) {
                ret.push({ name: "/", path: "/", type: "" })
            }

            continue
        }

        let fsInfo: FSInfo = {} as FSInfo
        fsInfo.type = ''
        fsInfo.name = paths[i]
        fsInfo.path = paths.slice(0, i + 1).join('/')
        ret.push(fsInfo)
    }

    return ret
})

let client: GuacdClient | null
let clientViewPort: GuacdViewPort | null

function onOpen() {
    windowResize()
    window.addEventListener('resize', windowResize);

    client = new GuacdClient(display.value, function () {
        isLoading.value = true
    }, function () {
        onClose()
    }, function () {
        isLoading.value = false
        fsDlgShow.value = false
        clientViewPort = new GuacdViewPort(viewport.value, client!)
        clientViewPort.SetDisplay(display.value, displayStyle)
        //clientViewPort.SetDrawer(drawer.value, drawerStyle)
        clientViewPort.Install()
    })

    if (props.connInfo.sftp) {
        showFilesystem.value = props.connInfo.sftp.enable
        client?.GetFileSystem()?.Install(fsDlg.value, upFile)
    }

    let tConnInfo = props.connInfo
    tConnInfo.remote.width = window.screen.width.toString()
    tConnInfo.remote.height = window.screen.height.toString()

    client.Reconn(tConnInfo)
}

function onClose() {
    showFilesystem.value = false
    window.removeEventListener('resize', windowResize)
    onCloseConn()
}

onMounted(function () {
})

onUnmounted(function () {
    onClose()
})

function onFSKeyDown(e: KeyboardEvent) {
    if (e.key == 'Backspace') {
        fsDlgSearchName.value = ''
        return
    }

    if (/[A-Z]/.test(e.key) || /[a-z]/.test(e.key)) {
        fsDlgSearchName.value += e.key
    }
}

function windowResize() {
    const w = document.documentElement.clientWidth || document.body.clientWidth;
    const h = document.documentElement.clientHeight || document.body.clientHeight;
    fsDlgMaxWidth.value = w * 0.5
    fsDlgMaxHeight.value = h * 0.3
}

function changeDir(path: string) {
    fsCurPath.value = path
    onFilesystem()
}

function rowFSDBClick(fsInfo: FSInfo) {
    if (fsInfo.type != 'dir') {
        if (!props.connInfo.sftp.down) {
            ElMessage.error('没有下载权限')
            return
        }

        ElMessageBox.confirm(
            `是否下载：${fsInfo.name}`,
            `下载`,
            {
                confirmButtonText: '下载',
                cancelButtonText: '取消',
                type: 'warning'
            }
        ).then(() => {
            const logMsg: FSLogInfo = { cmd: "下载", time: Now2Str() }
            let i = fsDlgLogData.push(logMsg) - 1

            client?.GetFileSystem()?.Down(fsInfo.path, function (recvSize, isEnd, err) {
                let size = recvSize / 1024
                let unit = "KB"
                if (size >= 1024) {
                    size /= 1024
                    unit = "MB"
                }

                if (err) {
                    fsDlgLogData[i].msg = `${fsInfo.name}下载失败，原因：${err}`
                    return
                }

                if (isEnd) {
                    fsDlgLogData[i].msg = `${fsInfo.name}下载完成，文件大小：${size.toFixed(2)}${unit}`
                    return
                }

                fsDlgLogData[i].msg = `${fsInfo.name}下载中，大小：${size.toFixed(2)}${unit}`
            })
        }).catch(() => {
        })

        return
    }

    changeDir(fsInfo.path)
}

function upFile(filename: string, data: Blob, type: string) {
    if (!props.connInfo.sftp.up) {
        ElMessage.error('没有上传权限')
        return
    }

    let logMsg: FSLogInfo = { cmd: "上传", time: Now2Str() }
    let i = fsDlgLogData.push(logMsg) - 1

    client?.GetFileSystem()?.Upload(data, type, "/" + filename, function (size, offset, err) {
        let upSize = offset / 1024
        let unit = "KB"
        if (upSize >= 1024) {
            upSize /= 1024
            unit = "MB"
        }

        if (err) {
            fsDlgLogData[i].msg = `${filename}上传失败，原因：${err}`
            return
        }

        if (size == offset) {
            onFilesystem()
            fsDlgLogData[i].msg = `${filename}上传完成，文件大小：${upSize.toFixed(2)}${unit}`
            return
        }

        fsDlgLogData[i].msg = `${filename}上传中，大小：${upSize.toFixed(2)}${unit}`
    })
}

//进入控制面板
function onDrawerEnter() {
    if (clientViewPort) {
        clientViewPort.StopCloseDrawer()
    }
}

//离开控制面板
function onDrawerLeave() {
    if (clientViewPort) {
        clientViewPort.CloseDrawer(2000)
    }
}

//发送ctrl+alt+del
function onSendCtrlAltDel() {
    client?.SendCtrlAltDel()
}

//文件系统
function onFilesystem() {
    fsDlgLoading.value = true
    client?.GetFileSystem()?.CD(fsCurPath.value, function (fileList) {
        fsDlgLoading.value = false
        fsDlgSearchName.value = ''
        fsDlgData.length = 0
        for (let i = 0; i < fileList.length; ++i) {
            fsDlgData.push(fileList[i])
        }
    })

    fsDlgShow.value = true
}

//全屏
function onFullScreenOrRecover() {
    //此处必须使用document.documentElement，否则其他元素无法置顶
    FullScreenOrRecover(document.documentElement)
}

//关闭连接
function onCloseConn() {
    ExitFullScreen()
    client?.GetFileSystem()?.UnInstall()
    client?.Disconn()
    client = null

    clientViewPort?.UnInstall()
    clientViewPort = null

    showRemote.value = false
}

</script>
  
<style scoped>
.el-dialog {
    /* 必须设置绝对定位的锚点 */
    position: relative;
}

.viewport {
    /* 相对el-dialog的绝对定位，这样可以使用它的大小 */
    position: absolute;
    background-color: #73767a;
    width: 100%;
    height: 100%;
}

.display {
    position: absolute;
}

/* 去掉获取焦点的边框 */
.display:focus {
    outline: none;
}

.ctrl-button {
    border-radius: 0;
}

.breadcrumb {
    margin-top: 5px;
}
</style>
  
<style>
/* 修改el-dialog的样式不能使用scoped */
.remoteDlg>.el-dialog__header {
    padding: 0px;
}

.remoteDlg>.el-dialog__body {
    padding: 0px;
}
</style>