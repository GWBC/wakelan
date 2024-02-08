
type WSOpen = (event: Event) => void
type WSMsg = (event: MessageEvent) => void
type WSError = (event: Event) => void
type WSClose = (event: Event, reconnTime: number) => boolean

export class WBSocket {
    private websocket: WebSocket | null
    private reconnTime: number
    private reconnTimer: number
    private url: string
    private reconnDefauteTime: number
    private wsOpenFun: WSOpen | null
    private wsMsgFun: WSMsg | null
    private wsErrorFun: WSError | null
    private wsCloseFun: WSClose | null

    //重连默认时间为0，则不重连
    constructor(reconnDefauteTime: number) {
        this.reconnDefauteTime = reconnDefauteTime
        this.websocket = null
        this.reconnTime = this.reconnDefauteTime
        this.reconnTimer = 0
        this.url = ''
        this.wsOpenFun = null
        this.wsMsgFun = null
        this.wsErrorFun = null
        this.wsCloseFun = null       
    }

    private wsOpen(event: Event) {
        //重置重连时间
        this.reconnTime = this.reconnDefauteTime

        if (this.wsOpenFun) {
            this.wsOpenFun(event)
        }
    }

    private wsMsg(event: MessageEvent) {
        if (this.wsMsgFun) {
            this.wsMsgFun(event)
        }
    }

    private wsError(event: Event) {
        this.websocket?.close()

        if (this.wsErrorFun) {
            this.wsErrorFun(event)
        }
    }

    private wsClose(event: Event) {
        if (this.wsCloseFun) {
            if (!this.wsCloseFun(event, this.reconnTime)) {
                return
            }
        }

        if (this.reconnDefauteTime != 0) {
            this.reconnTimer = setTimeout(() => {
                this.reconnTime *= 1.5
                if (this.reconnTime > 300) {
                    this.reconnTime = 300
                } else {
                    this.Conn(this.url)
                }
            }, this.reconnTime * 1000)
        }
    }

    SetOpenFun(fun: WSOpen) {
        this.wsOpenFun = fun
    }

    SetCloseFun(fun: WSClose) {
        this.wsCloseFun = fun
    }

    SetErrorFun(fun: WSError) {
        this.wsErrorFun = fun
    }

    SetMsgFun(fun: WSMsg) {
        this.wsMsgFun = fun
    }

    //连接服务
    Conn(url: string) {
        this.url = url

        //先移除之前的事件
        this.removeEvent(this.websocket)

        this.reconnTimer = 0
        this.websocket = new WebSocket(url)

        this.websocket.addEventListener("open", this.wsOpen.bind(this))
        this.websocket.addEventListener("message", this.wsMsg.bind(this))
        this.websocket.addEventListener("error", this.wsError.bind(this))
        this.websocket.addEventListener("close", this.wsClose.bind(this))
    }

    //断开连接
    Disconn() {
        if (this.reconnTimer != 0) {
            clearTimeout(this.reconnTimer)
            this.reconnTimer = 0
        }

        if (this.websocket != null) {
            this.removeEvent(this.websocket)
            this.websocket.close()
            this.websocket = null
        }
    }

    //获取WebSocket对象
    WebSocketObj(): WebSocket | null {
        return this.websocket
    }

    private removeEvent(wbsocket: WebSocket | null) {
        if (wbsocket != null) {
            wbsocket.removeEventListener("open", this.wsOpen)
            wbsocket.removeEventListener("message", this.wsMsg)
            wbsocket.removeEventListener("error", this.wsError)
            wbsocket.removeEventListener("close", this.wsClose)
        }
    }

}

