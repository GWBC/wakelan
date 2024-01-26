<template>
    <el-container class="sub-container">
        <el-aside width="50%">
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
                <el-table-column prop="time" label="时间" min-width="150" />
                <el-table-column prop="name" label="文件名" min-width="280" />
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
                                <el-button class="button" v-if="scope.row.index != scope.row.size" disabled size="small"
                                    type="info">下载</el-button>
                                <el-button class="button" v-else size="small" type="warning"
                                    @click="download(scope.row)">下载</el-button>
                            </el-col>
                        </el-row>
                    </template>
                </el-table-column>
            </el-table>
            </el-card>          
        </el-aside>
        <el-main class="wakelan-main">
            <el-form :model="sharedInfo">
                <el-form-item label="分享地址">
                    <el-input readonly ref="sharedInput" @focus="sharedFocus" v-model="sharedInfo.path"></el-input>
                </el-form-item>
                <el-form-item>
                    <qrcode-vue :value="sharedInfo.path" :options="{
                        width: 200,
                        height: 200,
                    }" />
                </el-form-item>
            </el-form>
        </el-main>
    </el-container>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import SparkMD5 from 'spark-md5'
import { ElMessage } from 'element-plus'
import QrcodeVue from 'qrcode.vue'
import { Fetch, DownloadFileFromURL } from '@/lib/comm'
import { UploadFilled } from '@element-plus/icons-vue'

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

interface SharedInfo {
    path: string
}

type Awaitable<T> = Promise<T> | T

const uploadObj = ref()

const metaDatas = ref<FileMeta[]>([])

const sharedInput = ref()
const sharedInfo = ref<SharedInfo>({} as SharedInfo)

let group = "/api/file"
let fileUpload = new Map()

async function fileMeta(md5: string): Promise<FileMeta[]> {
    return new Promise((resolve, reject) => {
        fetch(`${group}/meta?md5=${md5}`).
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

        const reader = new FileReader();
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

function sharedFocus() {
    sharedInput.value.select()
}

function remove(uploadFile: UploadFile, uploadFiles: UploadFiles): Awaitable<boolean> {
    fileUpload.delete(uploadFile.name)
    pullMetaData()
    return true
}

function download(row: FileMeta) {
    DownloadFileFromURL(`${group}/download?file=${row.md5}`, row.name)
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

                fetch(opt.action, {
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
    Fetch<FileMeta[]>(`${group}/meta`, null, (infos: FileMeta[]) => {
        metaDatas.value = infos
    })
}

onMounted(() => {
    pullMetaData()
})

</script>

<style scoped>
.upload-card{
    margin: 0px 20px 0px 20px; 
    height: 35%;
}

.upload {
    height: 100%;
}

.upload-process-card{
    margin: 10px 20px 0px 20px;
    height: 60%;
}

.upload-process {
    height: 100%;
}

.tag {
    width: 60px;
}

.button {
    width: 60px;
}

.sub-container {
    height: 100%;
}

</style>
