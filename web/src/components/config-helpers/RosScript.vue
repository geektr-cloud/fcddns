<script setup lang="ts">
import { reactive, computed, watch, ref } from 'vue'
import { NCard, NSpace, NInput, NCode, NCollapse, NCollapseItem, NButton, NSwitch, NScrollbar } from 'naive-ui'
import type { Rules } from 'async-validator'
import { useAsyncValidator } from '@vueuse/integrations/useAsyncValidator'
import { pascalCase } from 'es-toolkit'
import { useClipboard, useDebounceFn } from '@vueuse/core'

const props = defineProps<{ domain: string, refreshUrl: string }>()

const form = reactive({
  domain: props.domain,
  interface: '',
  comment: '',
  addressList: '',
  updateLocalDns: false,
})

const defaultComment = computed(() => form.interface ? `nic:${form.interface}` : '')

const rules: Rules = {
  interface: {
    type: 'string',
    required: true,
  },
}
const { pass, errorFields, execute } = useAsyncValidator(form, rules, { immediate: false })
const code = ref('')
const getCode = () => {
  const globalVariable = `FcddnsNicIpOf${pascalCase(form.interface)}`
  const comment = form.comment || defaultComment.value

  const trimIndent = (str: string, indent: number) => str.trim().split('\n').map((line, index) => index == 0 ? line : ' '.repeat(indent) + line).join('\n')
  const formatCode = (str: string) => str.replace(/\n\s+\n/g, '\n\n').trim() + '\n'

  const dnsScript = form.updateLocalDns ? `
:if ([:len [/ip dns static find where comment="${comment}"]] = 0) do={
  /ip dns static add name="${form.domain}" address=$NewIP comment="${comment}";
  :log info "create dns record ${form.domain}";
} else={
  /ip dns static set [find comment="${comment}"] address=$NewIP;
}` : ''

  const firewallScript = form.addressList ? `
:if ([:len [/ip firewall address-list find where list=${form.addressList} comment="${comment}"]] = 0) do={
  /ip firewall address-list add list=${form.addressList} comment="${comment}" address=$NewIP;
  :log info "add ip to firewall address list ${form.addressList}";
} else={
  /ip firewall address-list set [find list=${form.addressList} comment="${comment}"] address=$NewIP;
}` : ''

  return formatCode(`{
  :global ${globalVariable}
  :local NewIpCidr [/ip address get [find interface="${form.interface}"] address];
  :local NewIP [:pick $NewIpCidr 0 [:find $NewIpCidr "/"]]
  :if ($${globalVariable} = $NewIP) do={ :return true }

  :log info ([/tool fetch mode=https url="${props.refreshUrl}/$NewIP" output=user as-value]->"data");
  :set ${globalVariable} $NewIP;

  ${trimIndent(dnsScript, 2)}

  ${trimIndent(firewallScript, 2)}
}`)
}

watch([form], useDebounceFn(() => code.value = getCode(), 300))

const { copy, copied } = useClipboard({ source: code })

const copyScript = async () => {
  const { pass } = await execute()
  if (!pass) return
  await copy()
}
</script>

<template>
  <div class="w-2xl">
    <n-card title="RouterOS Script" hoverable>
      <n-space vertical>
        <n-space vertical>
          <p>Domain</p>
          <n-input :value="form.domain" readonly />
        </n-space>

        <n-space vertical>
          <div>
            <p>Interface <span class="text-red-400">*</span></p>
            <p class="text-sm text-gray-500">
              interface name that your public ip bind to, maybe your pppoe interface name
            </p>
          </div>
          <n-input v-model:value="form.interface" :status="errorFields?.interface?.length ? 'error' : 'success'" />
          <div v-if="errorFields?.interface?.length" class="text-sm text-red-300">
            {{ errorFields?.interface[0]?.message }}
          </div>
        </n-space>

        <n-space vertical>
          <div>
            <p>Comment</p>
            <p class="text-sm text-gray-500">
              comment is used to identify the dns record and firewall address list, defaults to
              <span class="text-zinc-400 font-mono">nic:&lt;interface name&gt;</span>
            </p>
          </div>
          <n-input v-model:value="form.comment" :placeholder="defaultComment" />
        </n-space>

        <n-space vertical>
          <div>
            <p>Address List</p>
            <p class="text-sm text-gray-500">
              if not empty, script will create ip in
              <span class="text-zinc-400 font-mono">/ip firewall address-list</span>
            </p>
          </div>
          <n-input v-model:value="form.addressList" placeholder="eg. WanAddresses" />
        </n-space>

        <n-space vertical>
          <div>
            <p>Update Local DNS</p>
            <p class="text-sm text-gray-500">
              if true, script will update dns record in
              <span class="text-zinc-400 font-mono">/ip dns static</span>
            </p>
          </div>
          <n-switch v-model:value="form.updateLocalDns" />
        </n-space>
      </n-space>
      <template #footer>
        <n-collapse>
          <n-collapse-item title="Show Script" name="script">
            <n-scrollbar style="max-height: 360px" trigger="none">
              <n-code :code=code language="routeros" word-wrap />
            </n-scrollbar>
          </n-collapse-item>
        </n-collapse>
      </template>
      <template #action>
        <n-button :disabled="!pass" @click="copyScript" type="primary">
          {{ copied ? 'copied!' : 'Copy Script' }}
        </n-button>
      </template>
    </n-card>
  </div>
</template>
