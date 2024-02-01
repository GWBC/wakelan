<template>
    <MainPage>
        <template #header />
        <template #main>
            <el-tabs class="config_tabs" v-model="activeName" style="margin: 10px 20px 0px 20px; height: calc(100% - 20px);"
                type="border-card">
                <el-tab-pane class="hcenter" style="height: 100%;" label="系统设置" name="系统设置">
                    <el-card style="min-width:50%;">
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
                                <el-text v-if="formData.auth_url.length == 0" class="text">请生成动态密码，使用手机小程序【动态密码】</el-text>
                                <qrcode-vue v-if="formData.auth_url.length != 0" :value="formData.auth_url"
                                    :size=qrCodeSize />
                            </el-form-item>
                            <el-form-item label="">
                                <div class="btn_container">
                                    <el-button type="danger" @click="onGenPassowrd">生成动态密码</el-button>
                                    <el-button type="primary" @click="onModify()">提交</el-button>
                                </div>
                            </el-form-item>
                        </el-form>
                    </el-card>
                </el-tab-pane>
                <el-tab-pane class="hcenter" style="height: 100%;" label="推送设置" name="推送设置">
                    <el-card style="min-width:50%;">
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
                                <div class="btn_container">
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
import { AsyncFetch, DeleteCookie } from '@/lib/comm'
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
    ip: '',
})

const group: string = 'api/system/'
const qrCodeSize = ref(200)

let secret: string

function getData() {
    AsyncFetch<GuacdInfo>(`${group}configinfo`, null).then(info => {
        formData.value = info
        secret = info.secret
    })
}

function onModify() {
    AsyncFetch(`${group}setconfig?info=${encodeURIComponent(JSON.stringify(formData.value))}`, null).then(info => {
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

<style>
.config_tabs .el-tabs__content {
    height: calc(100% - 80px);
}
</style>

<style scoped>
.btn_container {
    margin-left: auto;
}

.text {
    color: red;
    font-size: 16px;
    font-weight: bold;
}
</style>
 