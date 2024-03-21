<template>
    <el-dialog :append-to-body=true @open="onOpen">
        <RemoteForm v-model="cfgDlgShow" :host="host" :rand-key="props.randKey" :data="cfgDlgData" @submit="onSubmit" />
        <el-container>
            <el-main>
                <el-table height="200" :data="datas" empty-text=" " @row-dblclick="onDBClick">
                    <el-table-column label="协议" prop="remote.type">
                        <template #default="scope">
                            <el-tag v-if="scope.row.remote.type == RemoteType.RDP" effect="dark">{{
        p2s[scope.row.remote.type] }}</el-tag>
                            <el-tag v-if="scope.row.remote.type == RemoteType.VNC" effect="dark" type="info">{{
        p2s[scope.row.remote.type] }}</el-tag>
                            <el-tag v-if="scope.row.remote.type == RemoteType.SSH" effect="dark" type="warning">{{
        p2s[scope.row.remote.type] }}</el-tag>
                            <el-tag v-if="scope.row.remote.type == RemoteType.TELNET" effect="dark" type="danger">{{
        p2s[scope.row.remote.type] }}</el-tag>
                            <el-tag v-if="scope.row.remote.type == RemoteType.HTTP" effect="dark" type="success">{{
        p2s[scope.row.remote.type] }}</el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column label="主机" prop="remote.host"></el-table-column>
                    <el-table-column label="端口" prop="remote.port"></el-table-column>
                    <el-table-column v-if="props.edit" width="180" fixed="right" label="操作">
                        <template #default="scope">
                            <el-button size="small" type="danger" @click="delConfig(scope.row)">删除</el-button>
                        </template>
                    </el-table-column>
                </el-table>
            </el-main>
            <el-footer v-show="props.edit">
                <div class="w-full flex justify-end">
                    <el-button class="w-36" @click="addConfig" type="primary">添加</el-button>
                    <el-button class="w-36" @click="onSetting" type="success">提交</el-button>
                </div>
            </el-footer>
        </el-container>
    </el-dialog>
</template>

<script setup lang="ts">
import { DeepCopy } from '@/lib/comm'
import RemoteForm from './RemoteForm.vue'
import { RemoteType } from '@/lib/guacd/client'
import type { RemoteConfigInfo } from '@/lib/guacd/client';
import { ref, reactive } from 'vue'

const props = defineProps<{
    host: string,
    randKey: string,
    data: RemoteConfigInfo[],
    edit: boolean
}>()

const emit = defineEmits<{
    submit: [edit: boolean, host: string, datas: RemoteConfigInfo[]]
}>()

const host = ref('')
const cfgDlgShow = ref(false)
const cfgDlgData = ref<RemoteConfigInfo | null>({} as RemoteConfigInfo)

const datas = reactive<RemoteConfigInfo[]>([])

let p2s = ['RDP', 'VNC', 'SSH', 'TELNET', 'HTTP']

function findDatas(info: RemoteConfigInfo) {
    let i = 0
    for (; i < datas.length; ++i) {
        if (datas[i].id == info.id) {
            break
        }
    }

    if (i == datas.length) {
        return -1
    }

    return i
}

function onOpen() {
    datas.length = 0
    for (let i = 0; i < props.data.length; ++i) {
        datas.push(props.data[i])
    }
}

function delConfig(info: RemoteConfigInfo) {
    let i = findDatas(info)
    if (i < 0) {
        return
    }

    datas.splice(i, 1)
}

function addConfig() {
    host.value = props.host
    cfgDlgData.value = null
    cfgDlgShow.value = true
}

function onDBClick(info: RemoteConfigInfo) {
    if (props.edit) {
        host.value = props.host
        cfgDlgData.value = info
        cfgDlgShow.value = true
    } else {
        emit('submit', props.edit, props.host, [info])
    }
}

function onSetting() {
    emit('submit', props.edit, props.host, datas)
}

function onSubmit(oldCfg: RemoteConfigInfo | null, newCfg: RemoteConfigInfo) {
    //此处需要深拷贝    
    newCfg = DeepCopy(newCfg)

    if (oldCfg) {
        //编辑
        let i = findDatas(oldCfg)
        if (i >= 0) {
            datas[i] = newCfg
        }
    } else {
        //添加
        let i = findDatas(newCfg)
        if (i < 0) {
            datas.push(newCfg)
        } else {
            datas[i] = newCfg
        }
    }

    cfgDlgShow.value = false
}

</script>