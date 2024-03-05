<template>
    <div ref="container" class="ctrl_button_container" :style="containerStyle">
        <div class="container_x">
            <div class="ctrl_button" :class="ctrlBtnClass" :style="ctrlBtnStyle" @mouseenter.prevent="onEnter"
                @mouseleave.prevent="onLeave" @mousemove.prevent @mousedown.prevent="onMouseDown"
                @mouseup.prevent="onMouseUp">
                <span></span>
                <span></span>
            </div>
            <div class="container_back_btn" :style="containerBackBtnStyle">
                <div ref="back_btn" @click="onBackClick" class="contrl_btn_back" :style="ctrlBackBtnStyle" v-if="showBackBtn">
                    <slot></slot>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onUnmounted } from "vue";
import { onMounted, ref, nextTick } from "vue"

const ctrlSize = 40

const container = ref()
const containerStyle = ref({
    top: "calc(100% - 80px)",
    left: "calc(100% - 100px)"
})

const ctrlBtnClass = ref()
const ctrlBtnStyle = ref({
    width: ctrlSize + "px",
    height: ctrlSize + "px"
})

const containerBackBtnStyle = ref({
    left: "0px"
})

const ctrlBackBtnStyle = ref({
    height: ctrlSize + "px"
})

const showBackBtn = ref(false)

const back_btn = ref()

let leftDown = false
let downPoint = { x: 0, y: 0 }

function onEnter() {
    ctrlBtnClass.value = "animate_ratate"
    ctrlBtnStyle.value = {
        width: ctrlSize + 2 + "px",
        height: ctrlSize + 2 + "px"
    }
}

function onLeave() {
    ctrlBtnClass.value = ""
    ctrlBtnStyle.value = {
        width: ctrlSize + "px",
        height: ctrlSize + "px"
    }
}

function onMouseDown(e: MouseEvent) {
    downPoint = { x: e.clientX, y: e.clientY }

    leftDown = true
    ctrlBtnStyle.value = {
        width: ctrlSize + "px",
        height: ctrlSize + "px"
    }
}

function onBackClick(){
    showBackBtn.value = false
}

function onMouseUp(e: MouseEvent) {
    //判断是否为点击
    if (e.clientX == downPoint.x && e.clientY == downPoint.y) {
        showBackBtn.value = !showBackBtn.value

        nextTick(() => {
            let totalWidth = 0
            for (let c of back_btn.value?.children) {
                totalWidth += c.clientWidth
                totalWidth += 3
            }
            containerBackBtnStyle.value.left = totalWidth * -1 + "px"
        })
    }

    leftDown = false
    ctrlBtnStyle.value = {
        width: ctrlSize + 2 + "px",
        height: ctrlSize + 2 + "px"
    }
}

function onMouseMove(e: MouseEvent) {
    if (!leftDown) return

    const { clientX, clientY } = e

    containerStyle.value.top = clientY - ctrlSize / 2 + "px"
    containerStyle.value.left = clientX - ctrlSize / 2 + "px"
}

function onResize() {
    if (window.innerWidth < parseInt(containerStyle.value.left) ||
        window.innerHeight < parseInt(containerStyle.value.top)) {
        containerStyle.value.top = "calc(100% - 80px)"
        containerStyle.value.left = "calc(100% - 100px)"
    }
}

onMounted(() => {
    nextTick(() => {
        document.querySelector("body")?.append(container.value)
    })

    window.addEventListener("mousemove", onMouseMove)
    window.addEventListener("resize", onResize)
})

onBeforeUnmount(()=>{
    window.removeEventListener("mousemove", onMouseMove)
    window.removeEventListener("resize", onResize)
    document.querySelector("body")?.removeChild(container.value)
})

</script>

<style scoped>
.ctrl_button_container {
    position: absolute;
}

.container_x {
    position: relative;
}

.ctrl_button {
    position: absolute;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    cursor: pointer;
    z-index: 9999;
    background: linear-gradient(#14ffe9, #ffeb3b, #ff00e0);
}

.animate_ratate {
    animation: animate 1s linear infinite;
}

@keyframes animate {
    0% {
        transform: rotate(0deg);
    }

    100% {
        transform: rotate(360deg);
    }
}

.ctrl_button span {
    position: absolute;
    width: 100%;
    height: 100%;
    border-radius: 50%;
    background: linear-gradient(#14ffe9, #ffeb3b, #ff00e0);
}

.ctrl_button span:nth-child(1) {
    /* filter: blur(10px); */
}

.ctrl_button span:nth-child(2) {
    /* filter: blur(20px); */
}

.ctrl_button::after {
    content: "";
    position: absolute;
    top: 5px;
    left: 5px;
    right: 5px;
    bottom: 5px;
    background: #781188;
    border-radius: 50%;
}

.container_back_btn {
    position: absolute;
    overflow: hidden;
    z-index: 9998;
}

.contrl_btn_back {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 40px;
}

@keyframes animate_back {
    0% {
        transform: translateX(290px);
    }

    100% {
        transform: translateX(0);
    }
}
</style>

<style>
.contrl_btn_back button {
    height: 30px;
    margin: 0px !important;
}
</style>
