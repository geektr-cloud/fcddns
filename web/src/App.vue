<script setup lang="ts">
import { NPageHeader } from 'naive-ui'
import MyIP from '@/components/MyIP.vue'
import Sign from '@/components/Sign.vue'
</script>

<template>
  <n-page-header subtitle="......" class="mb-10">
    <template #title>
      <p>FC DDNS</p>
    </template>
  </n-page-header>

  <div class="container grid grid-cols-2 gap-4">
    <MyIP />
    <Sign />
  </div>
</template>