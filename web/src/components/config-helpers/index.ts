import { useModal } from "naive-ui";
import { defineAsyncComponent, h } from "vue";

export const useRosScript = () => {
  const modal = useModal();
  return (options: { domain: string; refreshUrl: string }) =>
    modal.create({ render: () => h(defineAsyncComponent(() => import("./RosScript.vue")), options) });
};
