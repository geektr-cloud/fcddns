<script setup lang="ts">
import { NButton, NCard, NSpace, NAlert, NGradientText, NIcon } from 'naive-ui'
import { Copy } from '@vicons/ionicons5'
import { useFetch, useClipboard } from '@vueuse/core'

const { isFetching, error, data, execute } = useFetch<string>("/myip", { immediate: false })
const { copy, copied, isSupported } = useClipboard()
</script>

<template>
  <n-card title="My IP Address" hoverable>
    <n-space class="items-center">
      <n-gradient-text type="info" class="font-bold" :class="data?.includes(':') ? 'text-lg' : 'text-3xl'">
        {{ data || '-' }}
      </n-gradient-text>
      <NButton size="small" secondary type="primary" v-if="isSupported" v-show="data" @click="data && copy(data)">
        <template #icon>
          <NIcon>
            <Copy />
          </NIcon>
        </template>
        {{ copied ? 'copied!' : 'Copy' }}
      </NButton>
    </n-space>
    <template #footer>
      <n-alert v-if="error" title="Error" type="error">
        {{ error }}
      </n-alert>
    </template>
    <template #action>
      <NButton :loading="isFetching" @click="execute()" type="primary">Get My IP</NButton>
    </template>
  </n-card>
</template>
