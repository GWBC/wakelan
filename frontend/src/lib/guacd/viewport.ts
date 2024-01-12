import type { GuacdClient } from './client'
import type { Ref } from 'vue'

export class GuacdViewPort {
    private viewPortEl: Element
    private displayEl!: Element
    private drawerEl!: Element
    private displayStyle!: Ref<Record<string, string | number>>
    private drawerStyle!: Ref<Record<string, string | number>>
    private client: GuacdClient
    private resizeObserver!: ResizeObserver
    private drawerTimerID!: number
    private onMouseMove: any
    private onTouchStart: any


    constructor(viewPortEl: Element, client: GuacdClient) {
        this.client = client
        this.viewPortEl = viewPortEl
    }

    SetDisplay(displayEl: Element, displayStyle: Ref<{}>) {
        this.displayEl = displayEl
        this.displayStyle = displayStyle
    }

    SetDrawer(drawerEl: Element, drawerStyle: Ref<{}>) {
        this.drawerEl = drawerEl
        this.drawerStyle = drawerStyle
    }

    Install() {
        this.onMouseMove = (e: MouseEvent) => {
            let top = e.clientY - this.displayEl.getBoundingClientRect().top
            if (top <= 4) {
                if (this.drawerStyle.value.visibility == "hidden") {
                    this.ShowDrawer(true)
                    this.CloseDrawer(2000)  //延迟2s关闭
                }
            }
        }

        this.onTouchStart = (e: TouchEvent) => {
            var touch = e.touches[0];
            let top = touch.clientY - this.displayEl.getBoundingClientRect().top
            if (top <= 4) {
                if (this.drawerStyle.value.visibility == "hidden") {
                    this.ShowDrawer(true)
                    this.CloseDrawer(6000)  //延迟6s关闭
                }
            }
        }

        if (this.displayEl && this.drawerEl) {
            this.displayEl.addEventListener('mousemove', this.onMouseMove);
            this.displayEl.addEventListener('touchstart', this.onTouchStart, { passive: true });
        }

        this.resizeObserver = new ResizeObserver(entries => {
            let size = this.client.GetSize()
            if (size.width == 0 || size.height == 0) {
                //必须设置缩放比例，否则无法显示
                this.client.Scale(1)
                return
            }

            //隐藏控制按钮
            this.ShowDrawer(false)

            for (const entry of entries) {
                const scale = Math.min(
                    entry.contentRect.width / size.width,
                    entry.contentRect.height / size.height
                )

                //先缩放
                this.client.Scale(scale)

                //再调整位置
                if (this.displayEl) {
                    this.displayStyle.value.top = (entry.contentRect.height - this.displayEl.clientHeight) / 2 + "px"
                    this.displayStyle.value.left = (entry.contentRect.width - this.displayEl.clientWidth) / 2 + "px"
                }
            }
        });

        this.resizeObserver.observe(this.viewPortEl);
    }

    UnInstall() {
        if (this.displayEl) {
            this.displayEl.removeEventListener('mousemove', this.onMouseMove);
            this.displayEl.removeEventListener('touchstart', this.onTouchStart)
        }

        if (this.resizeObserver) {
            this.resizeObserver.disconnect()
        }

        this.CloseDrawer(0)
    }

    ShowDrawer(isShow: boolean) {
        if (!this.drawerEl ||
            !this.viewPortEl ||
            !this.displayEl) {
            return
        }

        const rDrawer = this.drawerEl.getBoundingClientRect();
        const rViewport = this.viewPortEl.getBoundingClientRect();
        const left = Math.floor((rViewport.width - rDrawer.width) / 2)

        this.drawerStyle.value["top"] = this.displayStyle.value.top
        this.drawerStyle.value["left"] = left + 'px'
        this.drawerStyle.value["visibility"] = isShow ? 'visible' : 'hidden'
        this.drawerStyle.value["z-index"] = 1000
    }

    CloseDrawer(ms: number) {
        this.StopCloseDrawer()

        if (ms == 0) {
            this.ShowDrawer(false)
            return
        }

        this.drawerTimerID = setTimeout(() => {
            this.ShowDrawer(false)
        }, ms);
    }

    StopCloseDrawer() {
        if (this.drawerTimerID) {
            clearTimeout(this.drawerTimerID)
            this.drawerTimerID = 0
        }
    }
}



