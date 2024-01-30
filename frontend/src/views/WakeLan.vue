
<template>
  <Remote v-model="remoteShow" :conn-info="remoteConnInfo" />
  <RemoteConfig v-model="remoteCfgShow" :host="remoteHost" :data="remoteInfo" :edit="remoteEdit" :rand-key="remoteRandKey"
    @submit="onRemoteConfig" />
  <el-dialog v-model="addPCInfoDlgShow" :append-to-body=true @close="closeAddPCInfoDlg" @open="openAddPCInfoDlg">
    <el-form :model="addPCInfoDlgData" label-width="auto">
      <el-form-item label="地址">
        <el-input v-model="addPCInfoDlgData.ip" placeholder="请输入Host" />
      </el-form-item>
      <el-form-item label="地址">
        <el-input v-model="addPCInfoDlgData.mac" @blur="checkMacAddress" placeholder="请输入Mac" />
      </el-form-item>
      <el-form-item label="描述">
        <el-input v-model="addPCInfoDlgData.manuf" placeholder="请Host的描述" />
      </el-form-item>
    </el-form>
    <el-row>
      <el-col :span="21" />
      <el-col :span="3">
        <el-button style="width: 100%;" type="primary" @click="addPCInfoDlgSubmit">确认</el-button>
      </el-col>
    </el-row>
  </el-dialog>
  <MainPage>
    <template #header>
      <el-row :gutter="5">
        <el-col :xs="5" :sm="6" :md="3">
          <el-input v-model="searchIP" placeholder="请输入筛选的IP地址" clearable />
        </el-col>
        <el-col :xs="16" :sm="16" :md="8">
          <el-row :gutter="5">
            <el-col :xs="10" :span="8">
              <el-select v-model="netcard_select" placeholder="选择网卡" @change="onOpenNetcard">
                <el-option v-for="item in netcards" :key="item.name" :label="item.desc" :value="item.name" />
              </el-select>
            </el-col>
            <el-col :xs="4" :span="3">
              <el-button class="full-width" type="primary" @click="probeNetwork">探测</el-button>
            </el-col>
            <el-col :xs="4" :span="3">
              <el-button class="full-width" @click="pingAllPC" type="primary">检测</el-button>
            </el-col>
            <el-col :xs="4" :span="3">
              <el-button class="full-width" @click="addPCInfoDlgShow = true" type="primary">添加</el-button>
            </el-col>
          </el-row>
        </el-col>
      </el-row>
    </template>
    <template #main>
      <el-card class="wakelan_card" body-class="wakelan_card_body">
        <el-table class="wakelan_table disable-text-selection" element-loading-background="rgba(255, 255, 255, 20)"
          v-loading="table_loading" :data="table_data_filter" stripe @row-dblclick="onOpenRemote" empty-text=" "
          :default-sort="{ prop: 'ip', order: 'ascending' }" @sort-change="customSort">
          <el-table-column width="48">
            <template #default="scope">
              <el-button v-if="scope.row.edit" type="danger" size="small" :icon="Delete" circle
                @click="deletePC(scope.row)" />
              <el-button v-else="scope.row.edit" type="info" size="small" disabled :icon="Delete" circle />
            </template>
          </el-table-column>
          <el-table-column width="40">
            <template #default="scope">
              <el-icon v-show="scope.row.attach_info.star">
                <StarFilled color="#67C23A" />
              </el-icon>
              <el-icon v-show="!scope.row.attach_info.star">
                <StarFilled color="#b1b3b8" />
              </el-icon>
            </template>
          </el-table-column>
          <el-table-column prop="ip" label="地址" width="180" sortable="custom" />
          <el-table-column prop="mac" label="硬件地址" width="180" />
          <el-table-column label="描述" width="380">
            <template #default="scope">
              <el-input v-if="scope.row.edit" placeholder="编辑描述" v-model="scope.row.attach_info.describe" />
              <el-text v-else-if="scope.row.attach_info.describe"> {{ scope.row.attach_info.describe }} </el-text>
              <el-text v-else> {{ scope.row.manuf }}</el-text>
            </template>
          </el-table-column>
          <el-table-column label="在线" width="60">
            <template #default="scope">
              <el-icon v-show="scope.row.online">
                <CircleCheck color="#529b2e" />
              </el-icon>
              <el-icon v-show="!scope.row.online">
                <CircleClose color="red" />
              </el-icon>
            </template>
          </el-table-column>
          <el-table-column label="编辑" width="80">
            <template #default="scope">
              <el-switch @change="(val: boolean) => { editChange(val, scope.row) }" v-model="scope.row.edit"></el-switch>
            </template>
          </el-table-column>
          <el-table-column width="180" fixed="right" label="操作">
            <template #default="scope">
              <el-row>
                <el-col :span="8">
                  <el-button size="small" type="success" @click.stop="onRemoteCfg(scope.row)">配置</el-button>
                </el-col>
                <el-col :span="8">
                  <el-button v-if="!scope.row.attach_info.star" type="warning" size="small"
                    @click.stop="addStar(scope.$index, scope.row)">收藏</el-button>
                  <el-button v-else="scope.row.attach_info.star" type="info" size="small"
                    @click.stop="cancelStar(scope.$index, scope.row)">取消</el-button>
                </el-col>
                <el-col :span="8">
                  <el-button size="small" type="danger" @click.stop="wakeUP(scope.$index, scope.row)">唤醒</el-button>
                </el-col>
              </el-row>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>
  </MainPage>
</template>

<script setup lang="ts">
import '@/assets/wakelan.css'
import { Fetch, AsyncFetch } from '@/lib/comm'
import { Delete } from '@element-plus/icons-vue'
import Remote from '@/components/remote/Remote.vue'
import RemoteConfig from '@/components/remote/RemoteConfig.vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { RemoteType } from '@/lib/guacd/client'
import type { RemoteConfigInfo } from '@/lib/guacd/client'
import MainPage from '@/components/MainPage.vue'

interface AttachInfo {
  mac: string
  star: boolean
  describe: string
  remote: string
}

//PC信息
interface PCInfo {
  ip: string
  mac: string
  manuf: string
  attach_info: AttachInfo
  online: boolean
  edit: boolean
}

//网卡信息
interface NetcardInfo {
  name: string
  desc: string
  ips: string[]
}

interface WebsocketInfo {
  s: WebSocket | null
  conn_timer: number
}

const table_loading = ref(false)
const table_data = ref<PCInfo[]>([])

const searchIP = ref('')
const netcard_select = ref('')
const netcards = reactive<NetcardInfo[]>([])

const remoteCfgShow = ref(false)
const remoteHost = ref('')
const remoteRandKey = ref('')
const remoteInfo = ref<RemoteConfigInfo[]>([])
const remoteEdit = ref(false)

const remoteShow = ref(false)
const remoteConnInfo = ref<RemoteConfigInfo>({} as RemoteConfigInfo)

const addPCInfoDlgShow = ref(false)
const addPCInfoDlgData = ref<PCInfo>({} as PCInfo)

const group: string = 'api/wake/'

const websocket: WebsocketInfo = { s: null, conn_timer: 0 }

const table_data_filter = computed(() => {
  try {
    if (searchIP.value.length == 0) {
      return table_data.value
    }

    let regex = new RegExp(searchIP.value)
    return table_data.value.filter(
      (data) => {
        if (!searchIP.value) {
          return true
        }

        return regex.test(data.ip)
      }
    )
  } catch (error) {
    searchIP.value = ''
    return table_data.value
  }
})

function checkMacAddress() {
  if (addPCInfoDlgData.value.mac.length == 0) {
    return
  }

  const macRegex = /^([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})$/
  if (!macRegex.test(addPCInfoDlgData.value.mac)) {
    addPCInfoDlgData.value.mac = ''
    ElMessage.error('请输入有效的Mac地址，例如：11:22:33:44:55:66')
  }
}

function openAddPCInfoDlg() {
  addPCInfoDlgData.value = {
    ip: "",
    mac: "",
    manuf: "",
    attach_info: {
      mac: "",
      star: false,
      describe: "",
      remote: "",
    },
    online: false,
    edit: false
  } as PCInfo
}

function closeAddPCInfoDlg() {
  addPCInfoDlgData.value = {} as PCInfo
}

function addPCInfoDlgSubmit() {
  if (addPCInfoDlgData.value.mac.length == 0 || addPCInfoDlgData.value.ip.length == 0) {
    ElMessage.error('Host和Mac都不能为空')
    return
  }

  AsyncFetch(`${group}addnetworklist?ip=${addPCInfoDlgData.value.ip}&mac=${addPCInfoDlgData.value.mac}&`, null).then(() => {
    table_data.value.push(addPCInfoDlgData.value)
    addPCInfoDlgShow.value = false
  })
}

function deletePC(pcInfo: PCInfo) {
  ElMessageBox.confirm(`是否删除 ${pcInfo.ip} 的信息`, '警告',
    {
      confirmButtonText: '是',
      cancelButtonText: '否',
      type: 'warning',
      showClose: false
    }
  ).then(() => {
    AsyncFetch(`${group}delnetworklist?ip=${pcInfo.ip}`, null).then(() => {
      table_data.value = table_data.value.filter(item => item.ip != pcInfo.ip)
    })
  })
}

function editChange(val: boolean, pcInfo: PCInfo) {
  if (!val) {
    pcInfo.edit = true
    AsyncFetch(`${group}editpcinfo?mac=${pcInfo.mac}&describe=${pcInfo.attach_info.describe}`, null).then(() => {
      pcInfo.edit = false
    })
  }
}

function wsOpen(event: Event) {

}

function wsMsg(event: MessageEvent) {
  let a = event.data.toString().split(",")
  for (let i = 0; i < table_data.value.length; ++i) {
    if (table_data.value[i].mac == a[1]) {
      table_data.value[i].online = true
    }
  }
}

function wsError(event: Event) {
  websocket.s!.close()
}

function wsClose(event: Event) {
  websocket.conn_timer = setTimeout(() => {
    initWebsocket()
  }, 3000)
}

function initWebsocket() {
  websocket.conn_timer = 0
  websocket.s = new WebSocket(`ws://${window.location.host}/${group}pingpc`)

  websocket.s.addEventListener("open", wsOpen)
  websocket.s.addEventListener("message", wsMsg)
  websocket.s.addEventListener("error", wsError)
  websocket.s.addEventListener("close", wsClose)
}

function uninitWebsocket() {
  if (websocket.conn_timer != 0) {
    clearTimeout(websocket.conn_timer)
    websocket.conn_timer = 0
  }

  if (websocket.s != null) {
    websocket.s.removeEventListener("open", wsOpen)
    websocket.s.removeEventListener("message", wsMsg)
    websocket.s.removeEventListener("error", wsError)
    websocket.s.removeEventListener("close", wsClose)
    websocket.s.close()
    websocket.s = null
  }
}

function getData(isAes: number, showLoading: boolean = true): Promise<boolean> {
  return new Promise<boolean>((resolve, reject) => {
    table_loading.value = showLoading
    AsyncFetch<PCInfo[]>(`${group}getnetworklist?aes=${isAes}`, null).then(infos => {
      for (let i = 0; i < infos.length; ++i) {
        infos[i].online = false
      }

      table_data.value = infos
      table_loading.value = false
      resolve(true)
    }).catch(error => {
      table_loading.value = false
      reject(error)
    })
  })
}

function getNetCard(): Promise<boolean> {
  return new Promise<boolean>((resolve, reject) => {
    netcards.length = 0
    AsyncFetch<NetcardInfo[]>(`${group}getinterfaces`, null).then(infos => {
      for (let i = 0; i < infos.length; ++i) {
        netcards.push(infos[i])
      }

      resolve(true)
    }).catch(error => {
      reject(error)
    })
  })
}

function getNetcardSelect(): Promise<boolean> {
  return new Promise<boolean>((resolve, reject) => {
    AsyncFetch<NetcardInfo>(`${group}getselectnetcard`, null).then(info => {
      if (info.name) {
        if (info.desc.length == 0) {
          netcard_select.value = info.name
        } else {
          netcard_select.value = info.desc
        }
      }

      resolve(true)
    }).catch(error => {
      reject(error)
    })
  })
}

function onOpenNetcard(name: string) {
  Fetch(`${group}opencard?name=${name}`, null, infos => {
  })
}

function probeNetwork(name: string) {
  let proNetFun = () => {
    table_loading.value = true
    AsyncFetch(`${group}probenetwork`, null).then(() => {
      getData(1, false)
    })
  }

  ElMessageBox.confirm('是否清除之前的数据', '警告',
    {
      confirmButtonText: '是',
      cancelButtonText: '否',
      type: 'warning',
      showClose: false
    }
  ).then(() => {
    AsyncFetch(`${group}delnetworklist`, null).then(() => {
      proNetFun()
    })
  }).catch(() => {
    proNetFun()
  })
}

function customSort(oper: any) {
  if (!oper.order) {
    getData(1)
    return
  }

  getData(oper.order == 'ascending' ? 1 : 0)
}

function addStar(cloumn: any, row: any) {
  Fetch(`${group}operstar?star=1&mac=${row.mac}`, null, infos => {
    row.attach_info.star = true
  })
}

function cancelStar(cloumn: any, row: any) {
  Fetch(`${group}operstar?star=0&mac=${row.mac}`, null, infos => {
    row.attach_info.star = false
  })
}

function wakeUP(cloumn: any, row: any) {
  Fetch(`${group}wakeLan?&mac=${row.mac}`, null, infos => {
    ElMessageBox.alert('已发送唤醒', row.ip, {
      autofocus: false,
      confirmButtonText: 'OK',
    },
    )
  })
}

function pingAllPC() {
  if (!websocket.s) {
    ElMessage.error("websocket not connected")
    return
  }

  let cmd = {
    cmd: "ping",
    data: "",
  }

  for (let i = 0; i < table_data.value.length; ++i) {
    table_data.value[i].online = false
  }

  websocket.s.send(JSON.stringify(cmd))
}

function onOpenRemote(pcInfo: PCInfo) {
  if (pcInfo.edit) {
    return
  }

  let data: RemoteConfigInfo[] = []

  try {
    data = JSON.parse(pcInfo.attach_info.remote)
  } catch (error) {
    data = []
  }

  if (data.length == 0) {
    return
  }

  if (data.length == 1) {
    if (data[0].remote.type == RemoteType.HTTP) {
      let remote = data[0].remote
      let protocol = 'http'

      if (remote.https) {
        protocol = 'https'
      }

      window.open(`${protocol}://${remote.host}:${remote.port}/${remote.path}`, '_blank')
      return
    }

    remoteShow.value = true
    remoteConnInfo.value = data[0]
  } else {
    remoteCfgShow.value = true
    remoteEdit.value = false

    try {
      data = JSON.parse(pcInfo.attach_info.remote)
    } catch (error) {
      data = []
    }

    remoteInfo.value = data
  }
}

function onRemoteCfg(pcInfo: PCInfo) {
  remoteCfgShow.value = true
  if (pcInfo.attach_info.mac.length == 0) {
    //添加
    remoteEdit.value = true
    remoteHost.value = pcInfo.ip
    remoteInfo.value = []
  } else {
    //编辑
    remoteEdit.value = true
    remoteHost.value = pcInfo.ip
    let data = []
    try {
      data = JSON.parse(pcInfo.attach_info.remote)
    } catch (error) {
      data = []
    }

    remoteInfo.value = data
  }
}

function onRemoteConfig(edit: boolean, host: string, datas: RemoteConfigInfo[]) {
  if (edit) {
    remoteCfgShow.value = false

    for (let i = 0; i < table_data.value.length; ++i) {
      if (table_data.value[i].ip == host) {
        Fetch(`api/remote/setting?mac=${table_data.value[i].mac}`, datas, infos => {
          table_data.value[i].attach_info.mac = host
          table_data.value[i].attach_info.remote = JSON.stringify(datas)
        })

        return
      }
    }
  } else {
    if (datas[0].remote.type == RemoteType.HTTP) {
      let remote = datas[0].remote
      let protocol = 'http'

      if (remote.https) {
        protocol = 'https'
      }

      window.open(`${protocol}://${remote.host}:${remote.port}/${remote.path}`, '_blank')
      return
    }

    remoteShow.value = true
    remoteConnInfo.value = datas[0]
  }
}

function getRemoteRandKey() {
  return new Promise<boolean>((resolve, reject) => {
    AsyncFetch<string>(`${group}getRandKey`, null).then(infos => {
      remoteRandKey.value = infos
      resolve(true)
    }).catch(error => {
      reject(error)
    })
  })
}

onMounted(function () {
  initWebsocket()
  getNetcardSelect().then(ret => {
    getNetCard().then(ret => {
      getRemoteRandKey().then(ret => {
        getData(1)
      })
    })
  })
})

onUnmounted(function () {
  uninitWebsocket()
})
</script>

<style>
.wakelan_card_body {
  height: calc(100% - 20px);
}
</style>

<style scoped>
.wakelan_card {
  height: calc(100% - 20px);
  margin: 10px 20px 0px 20px;
}

.wakelan_table {
  height: 100%;
}
</style>
