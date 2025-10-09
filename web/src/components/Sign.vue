<script setup lang="ts">
import { reactive, ref, computed, watch } from 'vue'
import { NButton, NCard, NSpace, NAlert, NInput, NInputGroup, NSelect, NIcon, NInputNumber, NDivider } from 'naive-ui'
import { Eye, EyeOff } from '@vicons/ionicons5'
import type { Rules } from 'async-validator'
import { useAsyncValidator } from '@vueuse/integrations/useAsyncValidator'
import { useDebounceFn, useFetch } from '@vueuse/core'
import { SignJWT } from 'jose';

import { CopyButton, ExcuteButton } from './MyButtons';

const domainOptions = ["geektr.co", "geektr.cloud"].map(i => ({ label: i, value: i }))

const form = reactive({
  jwtSecret: '',
  host: '',
  domain: 'geektr.cloud',
  slot: undefined,
})

const rules: Rules = {
  jwtSecret: {
    type: 'string',
    required: true,
  },
  host: {
    type: 'string',
    required: true,
  },
  slot: {
    type: 'number',
  },
}

const { errorFields, execute } = useAsyncValidator(form, rules, { immediate: false })

const token = ref('')
const showToken = ref(false)
const safeToken = computed(() => {
  if (!token.value) return ""
  if (showToken.value) return token.value
  const [header, payload, sig] = token.value.split('.')
  return `${header}.${payload}.${'*'.repeat(sig?.length || 0)}`
})

watch(form, useDebounceFn(async () => {
  const { pass } = await execute()
  if (!pass) {
    token.value = ""
    return
  }

  const secret = new TextEncoder().encode(form.jwtSecret)
  const payload: any = { host: form.host, domain: form.domain }

  if (form.slot != undefined) payload.slot = form.slot

  token.value = await new SignJWT(payload)
    .setProtectedHeader({ alg: "HS256" })
    .sign(secret)
}, 300))

const url = computed(() => `${location.protocol}//${location.host}/ddns/v1/${token.value}`)
const rosCommand = computed(() => `:log info ([/tool fetch mode=https url="${url.value}" output=user as-value]->"data");`)

import { useRosScript } from './config-helpers';
const showRosScript = useRosScript();
const helpers = [
  {
    name: "RouterOS Script",
    action: () => showRosScript({ refreshUrl: url.value, domain: `${form.host}.${form.domain}` }),
  },
]
</script>

<template>
  <n-card title="DDNS Sign" hoverable>
    <n-space vertical>
      <n-input type="password" show-password-on="mousedown" placeholder="Your JWT Secret" v-model:value="form.jwtSecret"
        :status="errorFields?.jwtSecret?.length ? 'error' : 'success'" />
      <p v-if="errorFields?.jwtSecret?.length" class="text-red-300">
        Error: {{ errorFields?.jwtSecret[0]?.message }}
      </p>
      <n-input-group>
        <n-input type="text" placeholder="Domain Prefix" v-model:value="form.host"
          :status="errorFields?.host?.length ? 'error' : 'success'" />
        <n-select placeholder="Domain" v-model:value="form.domain" :options="domainOptions" />
      </n-input-group>
      <p v-if="errorFields?.host?.length" class="text-red-300">
        Error: {{ errorFields?.host[0]?.message }}
      </p>
      <n-input-number placeholder="Slot" v-model:value="form.slot" />
      <p v-if="errorFields?.slot?.length" class="text-red-300">
        Error: {{ errorFields?.slot[0]?.message }}
      </p>
    </n-space>
    <template #footer>
      <n-space vertical v-if="token">
        <n-alert type="success">
          <template #header>
            <n-space class="items-center cursor-pointer">
              <div>Token</div>
              <n-icon @click="showToken = !showToken">
                <Eye v-if="showToken" />
                <EyeOff v-else />
              </n-icon>
            </n-space>
          </template>
          <pre class="whitespace-pre-wrap">{{ safeToken }}</pre>
        </n-alert>
      </n-space>
    </template>
    <template #action>
      <div v-show="token">
        <n-space>
          <CopyButton tiny secondary size="small" type="info" label="Copy Token" :content="token" />
          <CopyButton tiny secondary size="small" type="info" label="Copy URL" :content="url" />
          <CopyButton tiny secondary size="small" type="info" label="Copy RouterOS Command" :content='rosCommand' />
          <n-button tiny secondary size="small" type="info" v-for="helper in helpers" @click="helper.action()">
            {{ helper.name }}
          </n-button>
        </n-space>
        <n-divider />
        <n-space>
          <ExcuteButton tiny secondary size="small" type="primary" :exec="() => useFetch(url, { immediate: false })"
            label="Update DNS Record with Current IP" loaddingLabel="Updating..." />
        </n-space>
      </div>
    </template>
  </n-card>
</template>
