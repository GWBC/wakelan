<template>
    <Navigation v-model="navigationShow" />
    <el-container class="wakelan-layout">
        <el-header class="wakelan-header">
            <el-row :gutter="10">
                <el-col :xs="2" :sm="2" :md="1">
                    <el-button :icon="Menu" @click="navigationShow = true" />
                </el-col>
            </el-row>
        </el-header>
        <el-main class="wakelan-main">
            <el-upload :show-file-list=false drag action="/api/file/upload" :before-upload="beforeUpload"
                :http-request="upload">
                <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                <div class="el-upload__text">
                    将文件放到这里或 <em>点击上传</em>
                </div>
            </el-upload>
        </el-main>
    </el-container>
</template>
  
<script setup lang="ts">
import { ref } from 'vue'
import SparkMD5 from 'spark-md5'
import Navigation from './Navigation.vue'
import { Menu } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { UploadFilled } from '@element-plus/icons-vue'

interface UploadRequestOptions {
    action: string
    method: string
    data: Record<string, string | Blob | [string | Blob, string]>
    filename: string
    file: File
    headers: Headers | Record<string, string | number | null | undefined>
    withCredentials: boolean
}

interface UploadRawFile extends File {
    uid: number
}

interface FileMeta {
    md5: string
    name: string
    size: number
    index: number
}

type Awaitable<T> = Promise<T> | T

const navigationShow = ref(false)

let fileMd5 = ''
let fileIndex = 0

async function fileMeta(md5: string): Promise<FileMeta[]> {
    return new Promise((resolve, reject) => {
        fetch(`/api/file/meta?md5=${fileMd5}`).
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

function beforeUpload(file: UploadRawFile): Awaitable<void | undefined | null | boolean | File | Blob> {
    fileMd5 = ''
    fileIndex = 0
    return new Promise((resolve, reject) => {
        let index = 0
        let block = 20 * 1024 * 1024
        let count = Math.ceil(file.size / block)
        let spark = new SparkMD5.ArrayBuffer()

        const reader = new FileReader();

        reader.onload = (e: ProgressEvent<FileReader>) => {
            spark.append(e.target?.result as ArrayBuffer)
            ++index

            if (index >= count) {
                fileMd5 = spark.end()
                spark.destroy()
                fileMeta(fileMd5).then(metas => {
                    fileIndex = metas[0].index
                    resolve(true)
                }).catch(err => {
                    ElMessage.error(`文件：${file.name}，上传失败，${err}`)
                    reject(err)
                })
            } else {
                nextRead()
            }
        }

        reader.onerror = (ev: ProgressEvent<FileReader>) => {
            ElMessage.error(`文件：${file.name}，上传失败，MD5计算失败`)
            spark.destroy()
            reject("md5 error")
        }

        let nextRead = () => {
            let start = index * block
            let end = ((start + block) >= file.size) ? file.size : start + block
            reader.readAsArrayBuffer(file.slice(start, end))
        }

        nextRead()
    });
}

function upload(opt: UploadRequestOptions): any {
    return new Promise((resolve, reject) => {
        const block = 10 * 1024 * 1024;
        let start = fileIndex
        let file = opt.file
        let md5 = fileMd5

        function uploadChunk() {
            const end = Math.min(start + block, file.size)
            const chunk = file.slice(start, end)

            const formData = new FormData();
            formData.append('md5', md5)
            formData.append('file', chunk)
            formData.append('index', start.toString())
            formData.append('size', file.size.toString())

            fetch(opt.action, {
                method: 'POST',
                body: formData,
            })
                .then(response => {
                    if (!response.ok) {
                        throw response.statusText
                    }

                    return response.json()
                })
                .then(data => {
                    if (!data) {
                        throw new Error("unknown error")
                    }

                    if (data.err.length != 0) {
                        throw data.err
                    }
                    start = end;

                    if (start < file.size) {
                        uploadChunk()
                    } else {
                        ElMessage.success(`文件：${file.name}，上传完成`)
                    }
                })
                .catch(error => {
                    ElMessage.error(`文件：${file.name}，上传失败，${error}`)
                });
        }

        uploadChunk()
    })
}

</script>
  