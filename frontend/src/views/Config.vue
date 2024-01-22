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
        <el-main class="wakelan-main cfg_main">
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
                    <qrcode-vue v-if="formData.auth_url.length != 0" :value="formData.auth_url" :options="qrCodeOptions" />
                </el-form-item>
                <el-form-item label="">
                    <div class="btn_container">
                        <el-button type="danger" @click="onGenPassowrd">生成动态密码</el-button>
                        <el-button type="primary" @click="onModify()">提交</el-button>
                    </div>
                </el-form-item>
            </el-form>
        </el-main>
    </el-container>
</template>
 
<script setup lang="ts">
import { Fetch, DeleteCookie } from '@/lib/comm';
import Navigation from './Navigation.vue'
import { Menu } from '@element-plus/icons-vue'
import { ref, onMounted } from 'vue'
import QrcodeVue from 'qrcode.vue';
import router from '@/router';

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

const navigationShow = ref(false)

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
    Fetch<GuacdInfo>(`${group}configinfo`, null, info => {
        formData.value = info
        secret = info.secret
    })
}

function onModify() {
    Fetch(`${group}setconfig?info=${encodeURIComponent(JSON.stringify(formData.value))}`, null, info => {
        if (secret != formData.value.secret) {
            DeleteCookie('token')
            router.push("/login")
        }
    })
}

function onGenPassowrd() {
    Fetch<AuthPwd>(`${group}genpwd`, null, info => {
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

.cfg_main {
    display: flex;
    justify-content: center;
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
 