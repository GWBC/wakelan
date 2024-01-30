<template>
    <el-dialog v-model="qrcodeShow">
        <div class="qrcode_container">
            <qrcode-vue :value="sharedInfo.path" :size="400" />
        </div>
    </el-dialog>
    <el-container class="sub-container">
        <el-aside width="50%">
            <div class="upload_container">
                <el-card class="upload-card">
                    <el-upload ref="uploadObj" class="upload" :show-file-list=true drag action="/api/file/upload"
                        :http-request="upload" :before-remove="remove" multiple>
                        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                        <div class="el-upload__text">
                            将文件放到这里或 <em>点击上传</em>
                        </div>
                    </el-upload>
                </el-card>

                <el-card class="upload-process-card" body-class="full-height">
                    <el-table class="upload-process" :data="metaDatas" empty-text=" " stripe>
                        <el-table-column prop="time" label="时间" min-width="100" />
                        <el-table-column prop="name" label="文件名" min-width="200" />
                        <el-table-column label="进度" min-width="100">
                            <template #default="scope">
                                <el-tag class="tag" v-if="scope.row.index == scope.row.size" type="success"
                                    effect="dark">100%</el-tag>
                                <el-tag class="tag" v-else type="info" effect="dark">{{ Math.floor(100 * (scope.row.index /
                                    scope.row.size)) }}%</el-tag>
                            </template>
                        </el-table-column>
                        <el-table-column fixed="right" label="操作" min-width="100">
                            <template #default="scope">
                                <el-row>
                                    <el-col :span="8">
                                        <el-button class="button" v-if="scope.row.index != scope.row.size" disabled
                                            size="small" type="info">下载</el-button>
                                        <el-button class="button" v-else size="small" type="warning"
                                            @click="download(scope.row)">下载</el-button>
                                    </el-col>
                                </el-row>
                            </template>
                        </el-table-column>
                    </el-table>
                </el-card>
            </div>
        </el-aside>
        <el-main class="wakelan-main">
            <div class="right_container">
                <el-card v-if="sharedKey.length == 0" class="addr_shared">
                    <el-text class="shared_link" truncated>
                        分享链接：<a :href="sharedInfo.path" target="_blank">{{ sharedInfo.path }}</a>
                    </el-text>
                    <el-button style="padding: 0px;" link :icon="DocumentCopy" @click="sharedCopy">复制</el-button>
                    <el-button style="padding: 0px;" link @click="shardQRCode">
                        <svg style="width: 16px; height: 16px; margin-right: 5px;" xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 512 512">
                            <rect x="336" y="336" width="80" height="80" rx="8" ry="8" />
                            <rect x="272" y="272" width="64" height="64" rx="8" ry="8" />
                            <rect x="416" y="416" width="64" height="64" rx="8" ry="8" />
                            <rect x="432" y="272" width="48" height="48" rx="8" ry="8" />
                            <rect x="272" y="432" width="48" height="48" rx="8" ry="8" />
                            <rect x="336" y="96" width="80" height="80" rx="8" ry="8" />
                            <rect x="288" y="48" width="176" height="176" rx="16" ry="16" fill="none" stroke="currentColor"
                                stroke-linecap="round" stroke-linejoin="round" stroke-width="32" />
                            <rect x="96" y="96" width="80" height="80" rx="8" ry="8" />
                            <rect x="48" y="48" width="176" height="176" rx="16" ry="16" fill="none" stroke="currentColor"
                                stroke-linecap="round" stroke-linejoin="round" stroke-width="32" />
                            <rect x="96" y="336" width="80" height="80" rx="8" ry="8" />
                            <rect x="48" y="288" width="176" height="176" rx="16" ry="16" fill="none" stroke="currentColor"
                                stroke-linecap="round" stroke-linejoin="round" stroke-width="32" />
                        </svg>
                        二维码
                    </el-button>
                </el-card>
                <el-card class="message" body-class="full-height">
                    <div class="message_container">
                        <el-input v-model="msgData" :rows="3" type="textarea" @keydown.enter.prevent="sendMsg"
                            placeholder="请输入消息，按回车发送消息" />
                        <el-table :data="msgDatas" class="message_table" empty-text=" " stripe>
                            <el-table-column label="时间" prop="time"></el-table-column>
                            <el-table-column label="消息">
                                <template #default="scope">
                                    <el-text truncated>
                                        {{ scope.row.msg }}
                                    </el-text>
                                </template>
                            </el-table-column>
                            <el-table-column label="操作" fixed="right" min-width="100px">
                                <template #default="scope">
                                    <el-button type="success" size="small" @click="msgCopy(scope.row)">复制</el-button>
                                </template>
                            </el-table-column>
                        </el-table>
                    </div>

                </el-card>

            </div>
        </el-main>
    </el-container>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import SparkMD5 from 'spark-md5'
import { ElMessage } from 'element-plus'
import QrcodeVue from 'qrcode.vue'
import router from '@/router'
import { UploadFilled, DocumentCopy } from '@element-plus/icons-vue'
import { Fetch, AsyncFetch, DownloadFileFromURL, SetLocalClipboard } from '@/lib/comm'

interface UploadRequestOptions {
    action: string
    method: string
    data: Record<string, string | Blob | [string | Blob, string]>
    filename: string
    file: File
    headers: Headers | Record<string, string | number | null | undefined>
    onError: (evt: any) => void
    onProgress: (evt: UploadProgressEvent) => void
    onSuccess: (response: any) => void
    withCredentials: boolean
}

interface UploadRawFile extends File {
    uid: number
}

type UploadStatus = 'ready' | 'uploading' | 'success' | 'fail'

interface UploadFile {
    name: string
    percentage?: number
    status: UploadStatus
    size?: number
    response?: unknown
    uid: number
    url?: string
    raw?: UploadRawFile
}

type UploadFiles = UploadFile[]

interface UploadProgressEvent extends ProgressEvent {
    percent: number
}

interface FileMeta {
    md5: string
    name: string
    size: number
    index: number
    time: string
}

interface Message {
    msg: string
    time: string
}

interface SharedInfo {
    path: string
}

type Awaitable<T> = Promise<T> | T

const uploadObj = ref()

const metaDatas = ref<FileMeta[]>([])
const sharedInfo = ref<SharedInfo>({
    path: 'https://www.baidu.com'
} as SharedInfo)

const msgData = ref('')
const msgDatas = ref<Message[]>()

const sharedKey = ref('')
const qrcodeShow = ref(false)

let group = "/api/file"
let fileUpload = new Map()

async function fileMeta(md5: string): Promise<FileMeta[]> {
    return new Promise((resolve, reject) => {
        fetch(`${group}/meta?md5=${md5}&key=${sharedKey.value}`).
            then(response => {
                if (!response.ok) {
                    throw response.statusText
                }

                try {
                    return response.json()
                } catch (errors) {
                    throw errors
                }
            }).then(data => {
                if (!data) {
                    throw new Error("unknown error")
                }

                if (data.err.length != 0) {
                    throw data.err
                }

                resolve(data.infos)
            }).catch(errors => {
                reject(`get fileMeta err: ${errors}`)
            })
    })
}

async function CalcMd5(file: File): Promise<FileMeta> {
    return new Promise((resolve, reject) => {
        let index = 0
        let block = 100 * 1024 * 1024
        let count = Math.ceil(file.size / block)
        let spark = new SparkMD5.ArrayBuffer()

        const reader = new FileReader()
        let meta = {} as FileMeta

        reader.onload = (e: ProgressEvent<FileReader>) => {
            spark.append(e.target?.result as ArrayBuffer)
            ++index

            if (index >= count) {
                meta.md5 = spark.end()
                spark.destroy()
                fileMeta(meta.md5).then(metas => {
                    if (metas.length == 0) {
                        meta.index = 0
                        meta.name = file.name
                        meta.size = file.size
                        resolve(meta)
                    } else {
                        resolve(metas[0])
                    }
                }).catch(err => {
                    reject(err)
                })
            } else {
                nextRead()
            }
        }

        reader.onerror = (ev: ProgressEvent<FileReader>) => {
            spark.destroy()
            reject("MD5计算失败")
        }

        let nextRead = () => {
            let start = index * block
            let end = ((start + block) >= file.size) ? file.size : start + block
            reader.readAsArrayBuffer(file.slice(start, end))
        }

        nextRead()
    })
}

function sendMsg() {
    AsyncFetch<Message>(`${group}/addMsg?key=${sharedKey.value}`, { "msg": msgData.value }).then((msg) => {
        msgData.value = ''
        console.log(msg)
        msgDatas.value?.unshift(msg)
        ElMessage.success("消息发送成功")
    }).catch(err => {
        ElMessage.success(`消息发送失败，${err.toString()}`)
    })
}

function msgCopy(msg:Message){
    SetLocalClipboard(msg.msg).then(()=>{
        ElMessage.success("复制消息成功")
    }).catch(err => {
        ElMessage.error(`复制消息失败，${err.toString()}`)
    })
}

function remove(uploadFile: UploadFile, uploadFiles: UploadFiles): Awaitable<boolean> {
    fileUpload.delete(uploadFile.name)
    pullMetaData()
    return true
}

function download(row: FileMeta) {
    DownloadFileFromURL(`${group}/download?file=${row.md5}&key=${sharedKey.value}`, row.name)
}

function upload(opt: UploadRequestOptions): any {
    return new Promise((resolve, reject) => {

        let file = opt.file
        if (fileUpload.get(file.name)) {
            let error = `文件：${file.name}，正在上传`
            ElMessage.warning(error)
            reject(error)
            return
        }

        fileUpload.set(file.name, true)

        CalcMd5(file).then(meta => {
            const block = 10 * 1024 * 1024;
            let start = meta.index
            let md5 = meta.md5

            function uploadChunk() {
                if (!fileUpload.get(file.name)) {
                    let error = `文件：${file.name}，取消上传`
                    ElMessage.warning(error)
                    reject(error)
                    return
                }

                const end = Math.min(start + block, file.size)
                const chunk = file.slice(start, end)

                const formData = new FormData();
                formData.append('md5', md5)
                formData.append('file', chunk)
                formData.append('name', file.name)
                formData.append('index', start.toString())
                formData.append('size', file.size.toString())

                fetch(`${opt.action}?key=${sharedKey.value}`, {
                    method: 'POST',
                    body: formData,
                }).then(response => {
                    if (!response.ok) {
                        throw response.statusText
                    }

                    return response.json()
                }).then(data => {
                    if (!data) {
                        throw new Error("unknown error")
                    }

                    if (data.err.length != 0) {
                        throw data.err
                    }
                    start = end;

                    if (start < file.size) {
                        //更新进度
                        opt.onProgress({ percent: 100 * (start / file.size) } as UploadProgressEvent)
                        uploadChunk()
                    } else {
                        pullMetaData()
                        fileUpload.delete(file.name)
                        opt.onSuccess(true)
                        ElMessage.success(`文件：${file.name}，上传完成`)
                        uploadObj.value.clearFiles('success')
                        resolve(true)
                    }
                }).catch(error => {
                    pullMetaData()
                    fileUpload.delete(file.name)
                    opt.onError(error)
                    ElMessage.error(`文件：${file.name}，上传失败，${error}`)
                    reject(error)
                });
            }

            uploadChunk()
        }).catch(error => {
            pullMetaData()
            fileUpload.delete(file.name)
            ElMessage.error(`文件：${file.name}，${error}`)
            reject(error)
        })
    })
}

function pullMetaData() {
    Fetch<FileMeta[]>(`${group}/meta?key=${sharedKey.value}`, null, (infos: FileMeta[]) => {
        metaDatas.value = infos
    })
}

function pullSharedKey() {
    return new Promise((resolve, reject) => {
        AsyncFetch<string>(`${group}/genkey`, null).then(infos => {
            resolve(infos)
        }).catch(err => {
            reject(err)
        })
    })
}

function sharedCopy() {
    SetLocalClipboard(sharedInfo.value.path).then(ret => {
        ElMessage.success("复制成功")
    }).catch(err => {
        ElMessage.error(`复制失败，${err.toString()}`)
    })
}

function shardQRCode() {
    qrcodeShow.value = true
}

function getMsg() {
    return new Promise((resolve, reject) => {
        AsyncFetch<Message[]>(`${group}/getMsg?key=${sharedKey.value}`, null).then(infos => {
            msgDatas.value = infos
            resolve(infos)
        }).catch(err => {
            reject(err)
        })
    })
}

onMounted(() => {
    let curPage = router.currentRoute.value
    if (curPage.path == "/shared") {
        //共享页面
        sharedKey.value = curPage.query["key"] as string
        if (!sharedKey) {
            //throw new Error("缺少参数Key")     //可以中止后续
            ElMessage.error("缺少参数Key")
        } else {
            getMsg().then(() => {
                pullMetaData()
            })
        }
    } else {
        pullSharedKey().then(key => {
            sharedInfo.value.path = `${window.location.protocol}//${window.location.host}/shared?key=${key}`
            getMsg().then(() => {
                pullMetaData()
            })
        })
    }
})

</script>

<style scoped>
.tag {
    width: 60px;
}

.button {
    width: 60px;
}

.sub-container {
    height: 100%;
}

.upload_container {
    display: flex;
    flex-direction: column;
    height: calc(100% - 10px);
}

.upload-card {
    flex: 4;
    margin: 10px 20px 0px 20px;
}

.upload {
    height: 100%;
}

.upload-process-card {
    flex: 5;
    margin: 10px 20px 0px 20px;

}

.upload-process {
    height: calc(100% - 20px);
}

.qrcode_container {
    display: flex;
    justify-content: center;
    align-items: center;
}

.right_container {
    display: flex;
    flex-direction: column;
    height: calc(100% - 10px);
}

.shared_link {
    display: block;
    margin-bottom: 8px;
}

.addr_shared {
    min-width: 200px;
    margin: 10px 20px 0px 20px;
}

.message {
    flex: 1;
    min-width: 200px;
    margin: 10px 20px 0px 20px;
}

.message_container {
    display: flex;
    flex-direction: column;
    height: calc(100% - 20px);
}

.message_table {
    flex: 1;
    margin-top: 20px;
}
</style>
