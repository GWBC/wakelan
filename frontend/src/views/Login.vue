<template>
    <div class="flex items-center justify-center h-full">
        <el-card class="flex items-center justify-center w-[650px] h-[300px]">
            <el-form class="w-[400px] pr-[60px]" label-position="right" label-width="100px" :model="formData">
                <el-form-item label="密钥">
                    <el-input ref="code" placeholder="请输入密钥" @keydown.enter.prevent="onLogin" v-model="formData.code" />
                </el-form-item>
                <el-form-item label="">
                    <el-button class="w-full" type="primary" @click="onLogin">登录</el-button>
                </el-form-item>
            </el-form>
        </el-card>
    </div>
</template>
  
<script setup lang="ts">
import { Fetch } from '@/lib/comm';
import { onMounted, ref } from 'vue'
import router from '@/router'

const code = ref()
const formData = ref({
    code: ""
})

function onLogin() {
    Fetch<number>(`/api/login?code=${formData.value.code}`, null, secretLen => {
        if (secretLen == 0) {
            router.push("/config")
        } else {
            router.push("/")
        }
    })
}

onMounted(function () {
    code.value.focus()
})

</script>
