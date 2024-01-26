<template>
    <MainPage>
        <template #header />
        <template #main>
            <FileShared />
        </template>
    </MainPage>
</template>

<script setup lang="ts">
import MainPage from '@/components/MainPage.vue'
import FileShared from '@/components/FileShared.vue'
</script>