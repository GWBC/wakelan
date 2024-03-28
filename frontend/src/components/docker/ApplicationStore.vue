<template>
    
    <!-- <el-dialog v-model="runDlgIsShow" append-to-body title="创建应用" @open="onRunDlgOpen">
        <el-form v-loading="loading" :model="dockerContainerCreate">
            <el-form-item label="容器名称">
                <el-input ref="containerNameInput" placeholder="容器名为空，则使用随机名称" v-model="dockerContainerCreate.name" />
            </el-form-item>
            <el-form-item label="重启策略">
                <el-select v-model="dockerContainerCreate.restart_policy" placeholder="选择异常退出后重启策略" size="large"
                    style="width: 240px">
                    <el-option key="no" label="no" value="no" />
                    <el-option key="always" label="always" value="always" />
                    <el-option key="on-failure" label="on-failure" value="on-failure" />
                    <el-option key="unless-stopped" label="unless-stopped" value="unless-stopped" />
                </el-select>
            </el-form-item>
            <el-form-item label="镜像名称">
                <el-input readonly v-model="dockerContainerCreate.image" />
            </el-form-item>
            <el-form-item label="执行命令">
                <el-input v-model="dockerContainerCreate.cmd" />
            </el-form-item>
            <el-form-item label="网络名称">
                <el-select v-model="dockerContainerCreate.net_name" placeholder="请选择网络">
                    <el-option v-for="item in dockerNetworkName" :key="item.name" :label="item.name"
                        :value="item.name"></el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="环境变量">
                <div class="flex flex-wrap gap-2 w-full">
                    <el-tag v-for="tag in dockerContainerCreate.env" :key="tag" closable :disable-transitions="false"
                        @close="onDlgEnvClose(tag)">
                        {{ tag }}
                    </el-tag>
                    <el-tooltip content="变量=值" placement="top-start">
                        <el-input v-if="dlgEnvInputShow" class="!w-40" ref="dlgEnvInputRef" v-model="dlgEnvInput"
                            size="small" @keyup.enter="onEnvInputConfirm" @blur="onEnvInputConfirm" />
                        <el-button type="primary" v-else size="small" @click="showEnvInput">
                            <el-icon>
                                <Plus />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-button class="!ml-0" type="danger" size="small" @click="onEnvClear">
                        <el-icon>
                            <Close />
                        </el-icon>
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item label="端口映射">
                <div class="flex flex-wrap gap-2 w-full">
                    <el-tag v-for="tag in dockerContainerCreate.ports" :key="tag" closable :disable-transitions="false"
                        @close="onDlgPortsClose(tag)">
                        {{ tag }}
                    </el-tag>
                    <el-tooltip content="主机:容器/协议，主机-主机:容器-容器/协议" placement="top-start">
                        <el-input v-if="dlgPortInputShow" class="!w-40" ref="dlgPortInputRef" v-model="dlgPortInput"
                            size="small" @keyup.enter="onPortInputConfirm" @blur="onPortInputConfirm" />
                        <el-button type="primary" v-else size="small" @click="showPortInput">
                            <el-icon>
                                <Plus />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-button class="!ml-0" type="danger" size="small" @click="onPortClear">
                        <el-icon>
                            <Close />
                        </el-icon>
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item label="目录映射">
                <div class="flex flex-wrap gap-2 w-full">
                    <el-tag v-for="tag in dockerContainerCreate.mounts" :key="tag" closable :disable-transitions="false"
                        @close="onDlgMountClose(tag)">
                        {{ tag }}
                    </el-tag>
                    <el-tooltip content="主机:容器" placement="top-start">
                        <el-input v-if="dlgMountInputShow" class="!w-40" ref="dlgMountInputRef" v-model="dlgMountInput"
                            size="small" @keyup.enter="onMountInputConfirm" @blur="onMountInputConfirm" />

                        <el-button type="primary" v-else size="small" @click="showMountInput">
                            <el-icon>
                                <Plus />
                            </el-icon>
                        </el-button>
                    </el-tooltip>
                    <el-button class="!ml-0" type="danger" size="small" @click="onMountClear">
                        <el-icon>
                            <Close />
                        </el-icon>
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item label="开启特权">
                <el-switch v-model="dockerContainerCreate.privileged" />
            </el-form-item>
            <el-form-item label="自动删除">
                <el-switch v-model="dockerContainerCreate.auto_remove" />
            </el-form-item>
            <el-form-item>
                <div class="w-full flex justify-end">
                    <el-button @click="runDlgIsShow = false">取消</el-button>
                    <el-button type="primary" @click="onRunDlgOk">
                        运行
                    </el-button>
                </div>
            </el-form-item>
        </el-form>
    </el-dialog> -->

    <el-scrollbar>
        <div class="app_container">
            <el-button class="app_card !flex !items-end !pb-6" v-for="(item, index) in store_list">
                <span class="!font-bold">
                    {{ item }}
                </span>
            </el-button>
            <el-button class="app_card" @click="onAddStore">
                <el-icon size="60">
                    <Plus />
                </el-icon>
            </el-button>
        </div>
    </el-scrollbar>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'

const group: string = 'api/docker/'
const store_list = ref<string[]>([])

function onAddStore() {
    store_list.value.push('士大夫')
}

onMounted(() => {
})

</script>

<style scoped>
.app_container {
    display: flex;
    flex-wrap: wrap;
}

.app_card {
    margin: 0 !important;
    max-width: 200px;
    height: 300px;
    flex-grow: 1;
    border-radius: 5px;
}
</style>