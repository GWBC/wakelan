<template>
    <MainPage>
        <template #header />
        <template #main>
            <el-card class="cfg_card">
                <el-form class="cfg" label-position="left" label-width="100px" :model="formData">
                    <el-form-item label="Guacd主机">
                        <el-input v-model="formData.guacd_host" />
                    </el-form-item>
                    <el-form-item label="Guacd端口">
                        <el-input v-model.number="formData.guacd_port" />
                    </el-form-item>
                    <el-form-item label="爱语飞飞">
                        <el-input v-model="formData.ayff_token" />
                    </el-form-item>
                    <el-form-item label="微信推送">
                        <el-input v-model="formData.wxpusher_token" />
                    </el-form-item>
                    <el-form-item label="微信推送主题">
                        <el-input v-model.number="formData.wxpusher_topicid" />
                    </el-form-item>
                    <el-form-item label="调试模式">
                        <el-switch v-model="formData.debug" />
                    </el-form-item>
                    <el-form-item label="动态密码">
                        <el-text v-if="formData.auth_url.length == 0" class="text">请生成动态密码，使用手机小程序【动态密码】</el-text>
                        <qrcode-vue v-if="formData.auth_url.length != 0" :value="formData.auth_url"
                            :options="qrCodeOptions" />
                    </el-form-item>
                    <el-form-item label="">
                        <div class="btn_container">
                            <el-button type="danger" @click="onGenPassowrd">生成动态密码</el-button>
                            <el-button type="primary" @click="onModify()">提交</el-button>
                        </div>
                    </el-form-item>
                </el-form>
            </el-card>
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
}

const formData = ref<GuacdInfo>({
    debug: false,
    guacd_host: '127.0.0.1',
    guacd_port: 4822,
    auth_url: '',
    secret: '',
    ayff_token: '',
    wxpusher_token: '',
    wxpusher_topicid: 0,
})

const group: string = 'api/system/'

const qrCodeOptions = ref({
    width: 200,
    height: 200,
})

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

<style scoped>
.btn_container {
    margin-left: auto;
}

.cfg_card {
    display: flex;
    justify-content: center;
    margin: 10px 20px 0px 20px;
    padding: 60px 0px 0px 60px;
}

.cfg {
    width: 600px
}

.text {
    color: red;
    font-size: 16px;
    font-weight: bold;
}
</style>
 