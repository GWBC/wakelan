import { GuacdClipboard } from './clipboard'
import Guacamole from 'guacamole-common-js'
import { ElMessageBox } from 'element-plus'
import { GuacdFilesystem } from './filesystem'

type GuacdStartConnFun = () => void
type GuacdCancelConnFun = () => void
type GuacdConnFinFun = () => void
type GuacdDisconnFinFun = () => void

interface GuacdSize {
    width: number
    height: number
}

export enum RemoteType {
    RDP,
    VNC,
    SSH,
    TELNET
}

export interface RemoteInfo {
    host: string
    port: number
    user: string
    pwd: string
    type: RemoteType
    width: string
    height: string
}

export interface SFTPInfo {
    enable: boolean
    up: boolean
    down: boolean
    rootPath: string
    keepalive: number
    host: string
    port: number
    user: string
    pwd: string
}

export interface RemoteConfigInfo {
    id: string
    remote: RemoteInfo
    sftp: SFTPInfo
}

//guacd客户端
export class GuacdClient {
    private displayEl: HTMLElement
    private startConnFun: GuacdStartConnFun | null
    private cancelConnFun: GuacdCancelConnFun | null
    private connFinFun: GuacdConnFinFun | null

    private connInfo!: RemoteConfigInfo
    private client!: Guacamole.Client | null
    private clientDisplay!: Guacamole.Display | null
    private clipboard!: GuacdClipboard | null
    private filesystem!: GuacdFilesystem | null
    private mouse!: Guacamole.Mouse | null
    private touch!: Guacamole.Mouse.Touchscreen | null
    private keyboard!: Guacamole.Keyboard | null
    private inputSink!: Guacamole.InputSink | null
    private mouseProc: any
    private touchProc: any
    private inputSinkProc: any

    constructor(displayEl: HTMLElement,
        onStartConn: GuacdStartConnFun,
        onCancelConn: GuacdCancelConnFun,
        onConnFinish: GuacdConnFinFun) {
        this.displayEl = displayEl
        this.startConnFun = onStartConn
        this.cancelConnFun = onCancelConn
        this.connFinFun = onConnFinish
    }

    //连接服务
    Conn(connInfo: RemoteConfigInfo) {
        //创建客户端连接对象
        const tunnel = new Guacamole.WebSocketTunnel("api/remote/conn")
        this.client = new Guacamole.Client(tunnel);

        //添加guacd绘图对象到display中
        this.clientDisplay = this.client.getDisplay()
        this.displayEl.appendChild(this.clientDisplay.getElement())

        //WS错误事件
        tunnel.onerror = (status: Guacamole.Status) => {
            ElMessageBox.confirm(
                status.message ? `服务异常，连接失败 => ${status.message}` : `服务异常，连接失败`,
                {
                    confirmButtonText: '重连',
                    cancelButtonText: '取消',
                    type: 'error',
                }
            ).then(() => {
                this.Reconn(connInfo)
            }).catch(() => {
                this.Disconn()

                if (this.cancelConnFun) {
                    this.cancelConnFun()
                }
            })
        }

        //错误事件
        this.client.onerror = (status: Guacamole.Status) => {
            ElMessageBox.confirm(
                status.message ? `服务异常，连接失败 => ${status.message}` : `服务异常，连接失败`,
                {
                    confirmButtonText: '重连',
                    cancelButtonText: '取消',
                    type: 'error',
                }
            ).then(() => {
                this.Reconn(connInfo)
            }).catch(() => {
                this.Disconn()

                if (this.cancelConnFun) {
                    this.cancelConnFun()
                }
            })
        }

        //状态变化事件
        this.client.onstatechange = (state: Guacamole.Client.State) => {
            switch (state) {
                case Guacamole.Client.State.CONNECTED:  //连接完成
                    {
                        this.initMouse()
                        this.initTouchScreen()
                        this.initKeyboard()

                        if (this.connFinFun) {
                            this.connFinFun()
                        }
                    }
                    break
                case Guacamole.Client.State.DISCONNECTED: //断开连接
                    {
                        ElMessageBox.confirm(
                            `连接断开，是否重连`,
                            {
                                confirmButtonText: '重连',
                                cancelButtonText: '取消',
                                type: 'info',
                            }
                        ).then(() => {
                            this.Reconn(connInfo)
                        }).catch(() => {
                            this.Disconn()

                            if (this.cancelConnFun) {
                                this.cancelConnFun()
                            }
                        })
                    }
                    break
            }
        }

        this.clipboard = new GuacdClipboard(this.client)
        this.clipboard.Install()
        this.client.onclipboard = (stream, mimeType) => {
            this.clipboard!.RemoteToLocalClipboard(stream, mimeType)
        }

        if (this.startConnFun) {
            this.startConnFun()
        }

        //创建文件系统对象
        this.filesystem = new GuacdFilesystem()
        this.client.onfilesystem = (object: Guacamole.Object, name: string) => {
            this.filesystem?.onFileSystem(object, name)
        }

        const param: string = 'info=' + encodeURIComponent(JSON.stringify(connInfo))

        //连接到远程服务器
        this.client.connect(param)

        this.connInfo = connInfo
    }

    //断开连接
    Disconn() {
        if (this.clipboard) {
            this.clipboard.UnInstall()
            this.clipboard = null
        }

        if (this.filesystem) {
            this.filesystem.UnInstall()
            this.filesystem = null
        }

        if (this.client) {
            this.client.onclipboard = null
            this.client.onerror = null
            this.client.onstatechange = null
            this.client.disconnect()
            this.client = null
        }

        if (this.mouse) {
            this.mouse = null
        }

        if (this.touch) {
            this.touch = null
        }

        if (this.keyboard) {
            this.keyboard.onkeyup = null
            this.keyboard.onkeydown = null
            this.keyboard = null
        }

        if (this.clientDisplay && this.displayEl) {
            if (this.inputSink) {
                this.displayEl.removeChild(this.inputSink.getElement())
                this.inputSink = null
            }

            if (this.inputSinkProc) {
                this.displayEl.removeEventListener("keydown", this.inputSinkProc)
                this.inputSinkProc = null
            }

            this.displayEl.removeChild(this.clientDisplay.getElement())
            this.clientDisplay = null
        }
    }

    //重连    
    Reconn(connInfo: RemoteConfigInfo) {
        this.Disconn()
        this.Conn(connInfo)
    }

    //获取文件系统对象
    GetFileSystem(): GuacdFilesystem | null {
        return this.filesystem
    }

    //发送
    SendCtrlAltDel() {
        this.client!.sendKeyEvent(1, 0xFFE3); // Ctrl key down
        this.client!.sendKeyEvent(1, 0xFFE9); // Alt key down
        this.client!.sendKeyEvent(1, 0xFFFF); // Delete key down

        this.client!.sendKeyEvent(0, 0xFFFF); // Delete key up
        this.client!.sendKeyEvent(0, 0xFFE9); // Alt key up
        this.client!.sendKeyEvent(0, 0xFFE3); // Ctrl key up
    }

    //获取display窗口大小
    GetSize(): GuacdSize {
        let size: GuacdSize = { width: 0, height: 0 }
        if (!this.clientDisplay) {
            return size
        }

        size.width = this.clientDisplay.getWidth()
        size.height = this.clientDisplay.getHeight()

        return size
    }

    //获取display窗口大小
    Scale(scale: number) {
        if (!this.clientDisplay) {
            return
        }

        this.clientDisplay.scale(scale)
    }

    //初始化鼠标事件
    private initMouse() {
        this.mouse = new Guacamole.Mouse(this.displayEl);
        this.mouseProc = (e: Guacamole.Mouse.Event, target: Guacamole.Event.Target) => {
            this.clientDisplay?.showCursor(false)
            this.client?.sendMouseState(e.state, true);
        }

        this.mouse.onEach(['mousedown', 'mousemove', 'mouseup'], this.mouseProc)

        this.clientDisplay!.showCursor(false)                      //初始化不显示鼠标指针
        this.clientDisplay!.oncursor = this.mouse.setCursor        //使用客户端绘制鼠标
    }

    //初始化触摸屏事件
    private initTouchScreen() {
        this.touch = new Guacamole.Mouse.Touchscreen(this.clientDisplay!.getElement());
        this.touchProc = (e: Guacamole.Mouse.Event, target: Guacamole.Event.Target) => {
            this.clientDisplay!.showCursor(false)
            this.client!.sendMouseState(e.state, true);
        }

        this.touch.longPressThreshold = 1000
        this.touch.onEach(['mousedown', 'mousemove', 'mouseup'], this.touchProc);
    }

    //初始化键盘事件
    private initKeyboard() {
        this.keyboard = new Guacamole.Keyboard(this.displayEl)
        this.keyboard.onkeydown = (keysym) => {
            this.client!.sendKeyEvent(1, keysym)
        }

        this.keyboard.onkeyup = (keysym) => {
            this.client!.sendKeyEvent(0, keysym)
        }

        //让display具备获取焦点的功能，这样才能使用键盘
        this.displayEl.setAttribute("tabindex", "0")

        if (this.connInfo.remote.type == RemoteType.SSH ||
            this.connInfo.remote.type == RemoteType.TELNET) {
            //使用文本输入模式，这样才能支持中文
            this.inputSink = new Guacamole.InputSink()
            this.displayEl.appendChild(this.inputSink.getElement())
            this.keyboard.listenTo(this.inputSink.getElement())

            this.inputSinkProc = () => {
                this.inputSink?.getElement().focus()
            }

            //必须添加tabindex，否则无法触发focus
            this.displayEl.addEventListener("focus", this.inputSinkProc, true)
        }
    }
}


