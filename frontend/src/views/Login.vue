<template>
    <div class="cfg_main">
        <el-form class="cfg" label-position="right" label-width="100px" :model="formData">
            <el-form-item label="密钥">
                <el-input ref="code" placeholder="请输入密钥" v-model="formData.code" />
            </el-form-item>
            <el-form-item label="">
                <el-button class="btn full-width" type="primary" @click="onLogin()">登录</el-button>
            </el-form-item>
        </el-form>
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
    Fetch(`/api/login?code=${formData.value.code}`, null, info => {
        router.push("/")
    })
}

onMounted(function () {
    code.value.focus()
})

</script>
  
<style scoped>
.btn {
    margin-left: auto;
}

.cfg_main {
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
}

.cfg {
    width: 400px
}
</style>
 