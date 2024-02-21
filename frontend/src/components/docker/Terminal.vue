<template>
    <el-dialog class="terminal-dlg" :title="props.title" v-model="show" :fullscreen="true" destroy-on-close @open="onOpen"
        @close="onClose">
        <div class="h-full" ref="xtermDiv"></div>
    </el-dialog>
</template>

<script setup lang="ts">
import { WBSocket } from '@/lib/websocket'
import { ref, computed, nextTick } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'

const xtermDiv = ref()
let wbs: WBSocket | null = null
let resizeObserver!: ResizeObserver

const props = defineProps({
    modelValue: Boolean,
    url: {
        type: String,
        required: true
    },
    title: {
        type: String,
        default: ''
    },
    onlyRead: {
        type: Boolean,
        default: true
    },
    autoExit: {
        type: Boolean,
        default: false
    }
})

const emit = defineEmits<{
    'update:modelValue': [boolean],
    open: [Terminal, WBSocket],
    close: [],
}>()


const show = computed({
    get() {
        return props.modelValue
    },
    set(value) {
        emit('update:modelValue', value)
    }
})

function onClose() {
    if (wbs) {
        wbs.Disconn()
    }

    if (resizeObserver) {
        resizeObserver.disconnect()
    }

    emit('close')
}

function onOpen() {
    const terminal = new Terminal()
    const fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)
    terminal.open(xtermDiv.value)

    fitAddon.fit()

    resizeObserver = new ResizeObserver(entries => {
        fitAddon.fit()
    })

    resizeObserver.observe(xtermDiv.value)

    wbs = new WBSocket(0)

    wbs.SetMsgFun((event: MessageEvent) => {
        terminal.write(event.data.toString())
    })

    const urlObj = new URL(props.url)
    urlObj.searchParams.append("rows", terminal.rows.toString())
    urlObj.searchParams.append("cols", terminal.cols.toString())

    wbs.Conn(urlObj.href)

    wbs.SetOpenFun(() => {
        terminal.onResize((event) => {
            const data = {
                cmd: "resize",
                rows: event.rows,
                cols: event.cols
            }

            wbs?.WebSocketObj()?.send(JSON.stringify(data))
        })

        if (!props.onlyRead) {
            terminal.onData(msg => {
                const data = {
                    cmd: "data",
                    data: msg
                }

                wbs?.WebSocketObj()?.send(JSON.stringify(data))
            })
        }

        terminal.focus()
    })

    wbs.SetCloseFun((event: Event, reconnTime: number) => {
        if(props.autoExit){
            show.value = false
        }        
        
        return false
    })

    emit('open', terminal, wbs)
}

</script>

<style>
.terminal-dlg {
    padding: 6px;
    overflow: hidden !important;
}

.terminal-dlg .el-dialog__header {
    height: 30px;
    margin: 0px;
    padding: 0px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.terminal-dlg .el-dialog__title {
    font-size: large;
    font-weight: bold;
}

.terminal-dlg .el-dialog__headerbtn {
    height: 30px;
    width: 30px;
    margin-right: 10px;
}

.terminal-dlg .el-dialog__body {
    height: calc(100% - 30px);
    margin: 0px;
    padding: 0px;
}
</style>