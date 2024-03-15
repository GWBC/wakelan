<template>
    <MainPage>
        <template #header />
        <template #main>
            <el-tabs class="navigation" v-model="activeName" type="border-card">
                <el-tab-pane class="flex justify-center" label="系统设置" name="系统设置">
                    <el-card class="min-w-[50%]">
                        <el-form label-position="left" label-width="100px" :model="formData">
                            <el-form-item label="公网地址">
                                <el-text style="font-weight: bold;" type="danger">{{ formData.ip }}</el-text>
                            </el-form-item>
                            <el-form-item label="获取公网地址">
                                <el-input v-model="formData.check_ip_addr" />
                            </el-form-item>
                            <el-form-item label="Guacd主机">
                                <el-input v-model="formData.guacd_host" />
                            </el-form-item>
                            <el-form-item label="Guacd端口">
                                <el-input v-model.number="formData.guacd_port" />
                            </el-form-item>
                            <el-form-item label="分享时限">
                                <el-select v-model.number="formData.shared_limit" placeholder="文件、消息分享保留天数">
                                    <el-option label="1天" :value="1">1天</el-option>
                                    <el-option label="3天" :value="3">3天</el-option>
                                    <el-option label="7天" :value="7">7天</el-option>
                                    <el-option label="15天" :value="15">15天</el-option>
                                </el-select>
                            </el-form-item>
                            <el-form-item label="调试模式">
                                <el-switch v-model="formData.debug" />
                            </el-form-item>
                            <el-form-item label="动态密码">
                                <el-text v-if="formData.auth_url.length == 0"
                                    class="text-red-600 text-lg font-bold">请生成动态密码，使用手机小程序【动态密码】</el-text>
                                <qrcode-vue v-if="formData.auth_url.length != 0" :value="formData.auth_url"
                                    :size=qrCodeSize />
                            </el-form-item>
                            <el-form-item label="">
                                <div class="ml-auto">
                                    <el-button type="danger" @click="onGenPassowrd">生成动态密码</el-button>
                                    <el-button type="primary" @click="onModify()">提交</el-button>
                                </div>
                            </el-form-item>
                        </el-form>
                    </el-card>
                </el-tab-pane>
                <el-tab-pane class="flex justify-center" label="推送设置" name="推送设置">
                    <el-card class="min-w-[50%]">
                        <el-form label-position="left" label-width="100px" :model="formData">
                            <el-form-item label="爱语飞飞">
                                <el-input v-model="formData.ayff_token" />
                            </el-form-item>
                            <el-form-item label="微信推送">
                                <el-input v-model="formData.wxpusher_token" />
                            </el-form-item>
                            <el-form-item label="微信推送主题">
                                <el-input v-model.number="formData.wxpusher_topicid" />
                            </el-form-item>
                            <el-form-item label="">
                                <div class="ml-auto">
                                    <el-button type="primary" @click="onModify()">提交</el-button>
                                </div>
                            </el-form-item>
                        </el-form>
                    </el-card>
                </el-tab-pane>
                <el-tab-pane class="flex justify-center" label="容器设置" name="容器设置">
                    <el-card class="min-w-[50%]">
                        <el-form label-position="left" label-width="100px" :model="formData">
                            <el-form-item label="启动TCP">
                                <el-switch v-model="formData.docker_enable_tcp" />
                            </el-form-item>
                            <el-form-item label="服务地址">
                                <el-input :disabled="!formData.docker_enable_tcp" v-model="formData.docker_svr_ip" />
                            </el-form-item>
                            <el-form-item label="服务端口">
                                <el-input :disabled="!formData.docker_enable_tcp"
                                    v-model.number="formData.docker_svr_port" />
                            </el-form-item>
                            <el-form-item label="推送用户">
                                <el-input v-model="formData.docker_user" />
                            </el-form-item>
                            <el-form-item label="推送密码">
                                <el-input type="password" v-model="formData.docker_passwd" />
                            </el-form-item>
                            <el-form-item>
                                <div class="ml-auto">
                                    <el-button type="primary" @click="onModify()">提交</el-button>
                                </div>
                            </el-form-item>
                        </el-form>
                    </el-card>
                </el-tab-pane>
            </el-tabs>
        </template>
    </MainPage>
</template>

<script setup lang="ts">
import { AsyncFetch, DeleteCookie, AESEncrypt } from '@/lib/comm'
import { ref, onMounted } from 'vue'
import QrcodeVue from 'qrcode.vue'
import router from '@/router'
import { ElMessage } from 'element-plus'
import MainPage from '@/components/MainPage.vue'

interface AuthPwd {
    auth_url: string
    secret: string
}

interface GuacdInfo {
    debug: boolean
    guacd_host: string
    guacd_port: number
    auth_url: string
    secret: string
    ayff_token: string
    wxpusher_token: string
    wxpusher_topicid: number
    shared_limit: number
    check_ip_addr: string
    docker_enable_tcp: boolean
    docker_svr_ip: string
    docker_svr_port: number
    container_root_path: string
    docker_user: string
    docker_passwd: string
    ip: string
}

const activeName = ref('系统设置')

const formData = ref<GuacdInfo>({
    debug: false,
    guacd_host: '127.0.0.1',
    guacd_port: 4822,
    auth_url: '',
    secret: '',
    ayff_token: '',
    wxpusher_token: '',
    wxpusher_topicid: 0,
    shared_limit: 7,
    check_ip_addr: '',
    docker_enable_tcp: false,
    docker_svr_ip: '127.0.0.1',
    docker_svr_port: 2375,
    container_root_path: '/opt/container-root',
    docker_user: '',
    docker_passwd: '',
    ip: '',
})

const group: string = 'api/system/'
const qrCodeSize = ref(200)

let secret: string
let randKey = ''
let iv = 'FF9B491CE5EE6BAF'

function getRandKey() {
    return new Promise<boolean>((resolve, reject) => {
        AsyncFetch<string>(`/api/public/getRandKey`, null).then(infos => {
            randKey = infos
            resolve(true)
        }).catch(error => {
            reject(error)
        })
    })
}

function getData() {
    getRandKey().then(() => {
        AsyncFetch<GuacdInfo>(`${group}configinfo`, null).then(info => {            
            formData.value = info
            secret = info.secret
        })
    })
}

function onModify() {
    let data = formData.value
    if (data.docker_passwd != '******'){
        data.docker_passwd = AESEncrypt(data.docker_passwd, randKey, iv)
    }
    
    AsyncFetch(`${group}setconfig?info=${encodeURIComponent(JSON.stringify(data))}`, null).then(info => {
        if (secret != formData.value.secret) {
            DeleteCookie('token')
            router.push("/login")
        } else {
            ElMessage.success(`修改成功`)
        }
    })
}

function onGenPassowrd() {
    AsyncFetch<AuthPwd>(`${group}genpwd`, null).then(info => {
        formData.value.auth_url = info.auth_url
        formData.value.secret = info.secret
    })
}

onMounted(function () {
    getData()
})
</script>
