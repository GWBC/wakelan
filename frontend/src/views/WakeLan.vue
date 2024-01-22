
<template>
  <Navigation v-model="navigationShow" />
  <Remote v-model="remoteShow" :conn-info="remoteConnInfo" />
  <RemoteConfig v-model="remoteCfgShow" :host="remoteHost" :data="remoteInfo" :edit="remoteEdit"
    @submit="onRemoteConfig" />
  <el-container class="wakelan-layout">
    <el-header class="wakelan-header">
      <el-row :gutter="10">
        <el-col :xs="2" :sm="2" :md="1">
          <el-button :icon="Menu" @click="navigationShow = true" />
        </el-col>
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
              <el-button class="full-width" type="success" @click="probeNetwork">探测</el-button>
            </el-col>
            <el-col :xs="4" :span="3">
              <el-button class="full-width" @click="pingAllPC" type="warning">Ping</el-button>
            </el-col>
          </el-row>
        </el-col>
      </el-row>
    </el-header>
    <el-main ref=table_ref class="wakelan-main">
      <el-table class="disable-text-selection" :max-height="table_max_height" v-loading="table_loading"
        :data="table_data_filter" stripe @row-dblclick="onOpenRemote" style="width: 100%" empty-text=" "
        :default-sort="{ prop: 'ip', order: 'ascending' }" @sort-change="customSort">
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
        <el-table-column label="在线" width="80">
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
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import '@/assets/wakelan.css'
import { Fetch, AsyncFetch } from '@/lib/comm'
import { Menu } from '@element-plus/icons-vue'
import Remote from '@/components/Remote.vue'
import Navigation from './Navigation.vue'
import RemoteConfig from '@/components/RemoteConfig.vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { RemoteType } from '@/lib/guacd/client'
import type { RemoteConfigInfo } from '@/lib/guacd/client'

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

const navigationShow = ref(false)

const table_ref = ref()
const table_max_height = ref(0)
const table_loading = ref(false)
const table_data: PCInfo[] = reactive([])

const searchIP = ref('')
const netcard_select = ref('')
const netcards = reactive<NetcardInfo[]>([])

const remoteCfgShow = ref(false)
const remoteHost = ref('')
const remoteInfo = ref<RemoteConfigInfo[]>([])
const remoteEdit = ref(false)

const remoteShow = ref(false)
const remoteConnInfo = ref<RemoteConfigInfo>({} as RemoteConfigInfo)

const group: string = 'api/wake/'
let resizeObserver: ResizeObserver | null = null

const websocket: WebsocketInfo = { s: null, conn_timer: 0 }

const table_data_filter = computed(() => {
  try {
    if (searchIP.value.length == 0) {
      return table_data
    }

    let regex = new RegExp(searchIP.value)
    return table_data.filter(
      (data) => {
        if (!searchIP.value) {
          return true
        }

        return regex.test(data.ip)
      }
    )
  } catch (error) {
    searchIP.value = ''
    return table_data
  }
})

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
  for (let i = 0; i < table_data.length; ++i) {
    if (table_data[i].mac == a[1]) {
      table_data[i].online = true
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

function initObserverSize() {
  resizeObserver = new ResizeObserver(entries => {
    table_max_height.value = table_ref.value.$el.offsetHeight
  })

  resizeObserver.observe(table_ref.value.$el)
}

function getData(isAes: number) {
  table_loading.value = true
  table_data.length = 0
  Fetch<PCInfo[]>(`${group}getnetworklist?aes=${isAes}`, null, infos => {
    for (let i = 0; i < infos.length; ++i) {
      infos[i].online = false
      table_data.push(infos[i])
    }

    table_loading.value = false
  })
}

function getNetCard() {
  netcards.length = 0
  Fetch<NetcardInfo[]>(`${group}getinterfaces`, null, infos => {
    for (let i = 0; i < infos.length; ++i) {
      netcards.push(infos[i])
    }
  })
}

function getNetcardSelect() {
  Fetch<NetcardInfo>(`${group}getselectnetcard`, null, info => {
    if (info.name) {
      if (info.desc.length == 0) {
        netcard_select.value = info.name
      } else {
        netcard_select.value = info.desc
      }
    }
  })
}

function onOpenNetcard(name: string) {
  Fetch(`${group}opencard?name=${name}`, null, infos => {
  })
}

function probeNetwork(name: string) {
  ElMessageBox.confirm('探测将清除之前的数据，是否继续', '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  ).then(() => {
    table_loading.value = true
    Fetch(`${group}cleannetworklist`, null, infos => {
      Fetch(`${group}probenetwork`, null, infos => {
        getData(1)
      })
    })
  }).catch(() => {

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

  for (let i = 0; i < table_data.length; ++i) {
    table_data[i].online = false
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

    for (let i = 0; i < table_data.length; ++i) {
      if (table_data[i].ip == host) {
        Fetch(`api/remote/setting?mac=${table_data[i].mac}`, datas, infos => {
          table_data[i].attach_info.mac = host
          table_data[i].attach_info.remote = JSON.stringify(datas)
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

onMounted(function () {
  initWebsocket()
  initObserverSize()
  getNetcardSelect()
  getNetCard()
  getData(1)
})

onUnmounted(function () {
  uninitWebsocket()
  resizeObserver!.disconnect()
})
</script>
