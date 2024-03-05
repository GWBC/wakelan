<template>
    <el-dialog v-model="editDlgShow" @open="onDlgOpen" @close="onDlgClose">
        <template #header>
            <h1 class=" text-gray-500 text-center">添加网络</h1>
        </template>
        <el-form :model="editFormData" label-width="80px" status-icon class="!mb-10" :rules="rules">
            <el-form-item label="名称" prop="name">
                <el-input ref="formName" v-model="editFormData.name" placeholder="请输入网络名称" />
            </el-form-item>
            <el-form-item label="驱动" prop="driver">
                <el-select v-model="editFormData.driver" placeholder="请选择驱动" @change="onDriverChange">
                    <el-option v-for="item in driverList" :key="item.value" :label="item.label" :value="item.value">
                        {{ item.label }}
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item v-if="editFormData.driver === 'macvlan'" label="网卡" prop="parent">
                <el-select v-model="editFormData.parent" placeholder="请选择关联网卡" allow-create filterable 
                    clearable @change="onParentChange">
                    <el-option v-for="item in localNetworkCard" :key="item.name" :label="item.name" :value="item.name">
                        {{ item.name }}
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="子网" prop="subnet">
                <el-input v-model="editFormData.subnet" placeholder="请输入子网，如：192.168.0.0/16" />
            </el-form-item>
            <el-form-item label="网关" prop="gateway">
                <el-input v-model="editFormData.gateway" placeholder="请输入网关，如：192.168.0.1" />
            </el-form-item>
        </el-form>

        <template #footer>
            <div class="dialog-footer">
                <el-button @click="editDlgShow = false">取消</el-button>
                <el-button type="primary" @click="onAddSubmit">添加</el-button>
            </div>
        </template>
    </el-dialog>

    <el-card class="h-full" body-class="h-full !pb-1">
        <el-table class="!h-full" :data="cardInfos" stripe empty-text=" " :row-key="(row: NetworkCardInfo) => row.id"
            :expand-row-keys="expandRow" @expand-change="onClick" @row-click="onClick" table-layout="auto">
            <el-table-column type="expand">

                <template #default="scope">
                    <div v-if="scope.row.configs.length > 0" class="ml-6 mb-2 text-gray-500 font-bold">
                        <span class="h-10 flex items-center">配置</span>
                        <hr />
                        <div class="mt-2 ml-4" v-for="v in scope.row.configs">
                            <div>
                                <el-tag type="primary" size="large">子网</el-tag>&nbsp=>
                                <el-tag type="success" size="large">{{ v.subnet }}</el-tag>
                            </div>

                            <div>
                                <el-tag type="primary" size="large">网关</el-tag>&nbsp=>
                                <el-tag type="success" size="large">{{ v.gateway }}</el-tag>
                            </div>

                            <div v-if="v.iprange.length > 0">
                                <el-tag type="primary" size="large">网络范围</el-tag>&nbsp=>
                                <el-tag type="success" size="large">{{ v.iprange }}</el-tag>
                            </div>

                        </div>
                    </div>

                    <div v-if="Object.keys(scope.row.options).length > 0" class="ml-6 mb-2 text-gray-500 font-bold">
                        <div class="h-10 flex items-center">
                            <span class="flex-1">选项</span>
                        </div>
                        <hr />
                        <div class="mt-2 ml-4" v-for="v, k in scope.row.options">
                            <el-tag type="primary" size="large">{{ k }}</el-tag>&nbsp=>
                            <el-tag type="success" size="large">{{ v }}</el-tag>
                        </div>
                    </div>

                </template>
            </el-table-column>
            <el-table-column prop="id" label="ID" />
            <el-table-column prop="name" label="名称" />
            <el-table-column prop="driver" label="驱动" />
            <el-table-column prop="scope" label="范围" />
            <el-table-column prop="created" label="创建时间" />
            <el-table-column fixed="right" width="120">

                <template #header>
                    <el-button type="primary" link @click.stop="onAdd">
                        <el-icon :size="20">
                            <Edit />
                        </el-icon>
                    </el-button>
                </template>

                <template #default="scope">
                    <el-button v-if="scope.row.name == 'host' || scope.row.name == 'bridge' || scope.row.name == 'none'"
                        disabled type="info" link size="small">删除</el-button>
                    <el-button v-else link type="danger" size="small" @click.stop="onDel(scope.row)">
                        删除
                    </el-button>
                </template>
            </el-table-column>
        </el-table>
    </el-card>
</template>

<script setup lang="ts">
import { onMounted, ref, reactive, nextTick } from 'vue'
import { ElMessageBox, FormRules } from 'element-plus'
import { AsyncFetch } from '@/lib/comm'

interface IPAMConfig {
    subnet: string
    iprange: string
    gateway: string
}

interface NetworkCardInfo {
    name: string
    id: string
    created: string
    scope: string
    driver: string
    options: { [key: string]: string }
    configs: IPAMConfig[]
}

interface AddrNet {
    P: string
    subnet: string
}

interface LocalNetworkInfo {
    name: string
    addrs: AddrNet[]
}

interface NetworkCardEditForm {
    name: string
    driver: string
    subnet: string
    gateway: string
    parent: string
}

interface FormItem {
    required: boolean
    validator: (rule: any, value: any, callback: any) => void
    trigger: string
}

const props = defineProps<{
    group: string
}>()

const expandRow = ref<string[]>([])
const cardInfos = ref<NetworkCardInfo[]>([])
const editDlgShow = ref(false)
const driverList = ref([
    {
        value: 'bridge',
        label: 'bridge',
    },
    {
        value: 'macvlan',
        label: 'macvlan',
    }
])
const editFormData = ref<NetworkCardEditForm>({
    name: '',
    driver: 'bridge',
    subnet: '',
    gateway: '',
    parent: '',
})
const formName = ref()

const localNetworkCard = ref<LocalNetworkInfo[]>([])

const rules = reactive<FormRules<NetworkCardEditForm>>({
    name: { required: true, validator: nameCheck, trigger: 'change' },
    subnet: { required: false, validator: subnetCheck, trigger: 'change' },
    gateway: { required: false, validator: gatewayCheck, trigger: 'change' },
    parent: { required: false, message: '请选择关联网卡' },
    driver: { required: true, message: '请选择驱动' },
})

function onParentChange() {
    if (editFormData.value.driver == 'macvlan' && editFormData.value.parent != '') {
        let i = localNetworkCard.value.findIndex((v) => {
            return v.name == editFormData.value.parent
        })

        if(i < 0){
            return
        }
        
        let val = localNetworkCard.value[i]
        editFormData.value.subnet = val.addrs.length > 0 ? val.addrs[0].subnet : editFormData.value.subnet
        editFormData.value.gateway = editFormData.value.subnet.split('/')[0]
    } else {
        editFormData.value.subnet = ""
        editFormData.value.gateway = ""
    }
}

function onDriverChange() {
    if (editFormData.value.driver == 'macvlan') {
        if (rules.subnet) {
            let r = rules.subnet as FormItem
            r.required = true
        }

        if (rules.gateway) {
            let r = rules.gateway as FormItem
            r.required = true
        }

        if (rules.parent) {
            let r = rules.parent as FormItem
            r.required = true
        }
    } else {
        if (rules.subnet) {
            let r = rules.subnet as FormItem
            r.required = false
        }

        if (rules.gateway) {
            let r = rules.gateway as FormItem
            r.required = false
        }

        if (rules.parent) {
            let r = rules.parent as FormItem
            r.required = false
        }
    }
}

function nameCheck(rule: any, value: any, callback: any) {
    const regex = /^[^\d].{0,20}$/

    if (value == '') {
        callback(new Error('网络名不能为空'))
    } else if (!regex.test(value)) {
        callback(new Error('网络名不能数字开头，长度1-20'))
    } else {
        callback()
    }
}

function subnetCheck(rule: any, value: any, callback: any) {
    if (value == '' && !rule.required) {
        callback()
        return
    }

    const regex = /^(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\/([0-9]|[1-2][0-9]|3[0-2]+)$/
    if (value == '') {
        callback(new Error('子网不能为空'))
    } else if (!regex.test(value)) {
        callback(new Error('子网格式错误'))
    } else {
        callback()
    }
}

function gatewayCheck(rule: any, value: any, callback: any) {
    if (value == '' && !rule.required) {
        callback()
        return
    }

    const regex = /^(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)\.(25[0-5]|2[0-4]\d|[01]?\d\d?)$/
    if (value == '') {
        callback(new Error('网关不能为空'))
    } else if (!regex.test(value)) {
        callback(new Error('网关格式错误'))
    } else {
        callback()
    }
}

function onDlgClose() {
    editFormData.value = {
        name: '',
        driver: 'bridge',
        subnet: '',
        gateway: '',
        parent: '',
    }
}

function onDlgOpen() {
    nextTick(() => {
        //由于dlg自动焦点在标题上，需要自己实现获取焦点
        formName.value?.focus()
    })
}

function onAddSubmit() {
    const url = `${props.group}addNetworkCard?name=${editFormData.value.name}&` +
        `driver=${editFormData.value.driver}&` +
        `subnet=${editFormData.value.subnet}&` +
        `gateway=${editFormData.value.gateway}&` +
        `parent=${editFormData.value.parent}`

    AsyncFetch(url, null).then(() => {
        getNetworks()
        editDlgShow.value = false
    })
}

function onAdd(row: NetworkCardInfo) {
    editDlgShow.value = true
}

function onDel(row: NetworkCardInfo) {
    ElMessageBox.confirm(`确定删除 ${row.name}?`, '警告', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
        AsyncFetch(`${props.group}delNetworkCard?name=${row.name}`, null).then(() => {
            let index = cardInfos.value.indexOf(row)
            if (index != -1) {
                cardInfos.value.splice(index, 1)
            }
        })
    })
}

function onClick(row: NetworkCardInfo) {
    const index = expandRow.value.indexOf(row.id)
    if (index != -1) {
        //收缩
        expandRow.value.length = 0
        return
    }

    //展开
    expandRow.value.length = 0
    expandRow.value.push(row.id)
}

function getLocalNetwork() {
    AsyncFetch<LocalNetworkInfo[]>(`${props.group}localNetworkCard`, null).then((infos) => {
        localNetworkCard.value = infos
    })
}

function getNetworks() {
    AsyncFetch<NetworkCardInfo[]>(`${props.group}getNetworkCards`, null).then((infos) => {
        cardInfos.value = infos
        getLocalNetwork()
    })
}

onMounted(() => {
    getNetworks()
})


</script>

<style scoped>
.el-tag {
    border: none;
    aspect-ratio: 1;
}
</style>