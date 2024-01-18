<template>
    <el-dialog :append-to-body=true @open="onOpen" @close="onClose">
        <el-container>
            <el-main>
                <el-tabs>
                    <el-tab-pane label="远程服务">
                        <el-form :model="data" label-width="auto">
                            <el-form-item label="主机">
                                <el-input v-model="data.remote.host" disabled />
                            </el-form-item>
                            <el-form-item label="端口">
                                <el-input v-model="data.remote.port" />
                            </el-form-item>
                            <el-form-item label="协议">
                                <el-select v-model="data.remote.type" placeholder="选择协议" @change="onSelectChange">
                                    <el-option label="RDP" :value=0 />
                                    <el-option label="VNC" :value=1 />
                                    <el-option label="SSH" :value=2 />
                                    <el-option label="TELNET" :value=3 />
                                </el-select>
                            </el-form-item>
                            <el-form-item label="用户">
                                <el-input v-model="data.remote.user" />
                            </el-form-item>
                            <el-form-item label="密码">
                                <el-input type="password" placeholder="请输入密码" v-model="data.remote.pwd" />
                            </el-form-item>
                        </el-form>
                    </el-tab-pane>
                    <el-tab-pane label=" SFTP服务">
                        <el-form :model="data" label-width="auto">
                            <el-form-item label="启动">
                                <el-switch v-model="data.sftp.enable" />
                            </el-form-item>
                            <el-form-item label="权限">
                                <el-checkbox label="上传" v-model="data.sftp.up" :disabled="!data.sftp.enable" />
                                <el-checkbox label="下载" v-model="data.sftp.down" :disabled="!data.sftp.enable" />
                            </el-form-item>
                            <el-form-item label="主机">
                                <el-input v-model="data.sftp.host" :disabled="!data.sftp.enable" />
                            </el-form-item>
                            <el-form-item label="端口">
                                <el-input v-model="data.sftp.port" :disabled="!data.sftp.enable" />
                            </el-form-item>
                            <el-form-item label="用户">
                                <el-input v-model="data.sftp.user" :disabled="!data.sftp.enable" />
                            </el-form-item>
                            <el-form-item label="密码">
                                <el-input type="password" placeholder="请输入密码" v-model="data.sftp.pwd"
                                    :disabled="!data.sftp.enable" />
                            </el-form-item>
                            <el-form-item label="根路径">
                                <el-input v-model="data.sftp.rootPath" :disabled="!data.sftp.enable" />
                            </el-form-item>
                        </el-form>
                    </el-tab-pane>
                </el-tabs>
            </el-main>
            <el-footer>
                <el-row>
                    <el-col :span="21" />
                    <el-col :span="3">
                        <el-button style="width: 100%;" type="primary" @click="onSubmit">确认</el-button>
                    </el-col>
                </el-row>
            </el-footer>
        </el-container>
    </el-dialog>
</template>

<script setup lang="ts">
import { reactive } from 'vue'
import { AESEncrypt } from '@/lib/comm'
import { RemoteType } from '@/lib/guacd/client'
import type { RemoteConfigInfo, RemoteInfo, SFTPInfo } from '@/lib/guacd/client';

const props = defineProps<{
    host: string
    data: RemoteConfigInfo | null
}>()

const emit = defineEmits<{
    submit: [oldCfg: RemoteConfigInfo | null, newCfg: RemoteConfigInfo]
}>()

const data = reactive<RemoteConfigInfo>({
    id: '',
    remote: {} as RemoteInfo,
    sftp: {} as SFTPInfo
})

const t2p = [3389, 5900, 22, 23]
const t2u = ['Administrator', 'Administrator', 'root', 'Administrator']

const iv = 'aaaaaaaaaabbbbbb'
const key = '111111111122222222223333333333aa'

function onOpen() {
    if (props.data != null) {
        data.remote.host = props.data.remote.host
        data.remote.port = props.data.remote.port
        data.remote.user = props.data.remote.user
        data.remote.pwd = props.data.remote.pwd
        data.remote.type = props.data.remote.type

        data.sftp.enable = props.data.sftp.enable
        data.sftp.up = props.data.sftp.up
        data.sftp.down = props.data.sftp.down
        data.sftp.rootPath = props.data.sftp.rootPath
        data.sftp.keepalive = props.data.sftp.keepalive
        data.sftp.host = props.data.sftp.host
        data.sftp.port = props.data.sftp.port
        data.sftp.user = props.data.sftp.user
        data.sftp.pwd = props.data.sftp.pwd
    } else {
        data.remote.host = props.host
        data.remote.type = RemoteType.RDP
        data.remote.port = t2p[data.remote.type]
        data.remote.user = t2u[data.remote.type]
        data.remote.pwd = ''

        data.sftp.enable = false
        data.sftp.up = true
        data.sftp.down = true
        data.sftp.rootPath = '/'
        data.sftp.keepalive = 10
        data.sftp.host = props.host
        data.sftp.port = 22
        data.sftp.user = 'root'
        data.sftp.pwd = ''
    }
}

function onClose() {
    data.remote.host = ''
    data.remote.type = RemoteType.RDP
    data.remote.port = t2p[data.remote.type]
    data.remote.user = t2u[data.remote.type]
    data.remote.pwd = ''

    data.sftp.enable = false
    data.sftp.up = true
    data.sftp.down = true
    data.sftp.rootPath = '/'
    data.sftp.keepalive = 10
    data.sftp.host = ''
    data.sftp.port = 22
    data.sftp.user = 'root'
    data.sftp.pwd = ''
}

function onSelectChange() {
    data.remote.port = t2p[data.remote.type]
    data.remote.user = t2u[data.remote.type]

    if (data.remote.type == RemoteType.SSH) {
        data.sftp.port = data.remote.port
        data.sftp.user = data.remote.user
        data.sftp.pwd = data.remote.pwd
    }
}

function onSubmit() {
    data.id = data.remote.host + data.remote.port
    data.remote.pwd = AESEncrypt(data.remote.pwd, key, iv)
    data.sftp.pwd = AESEncrypt(data.sftp.pwd, key, iv)
    emit('submit', props.data, data)
}

</script>