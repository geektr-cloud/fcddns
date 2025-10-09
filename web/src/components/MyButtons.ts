import { h, ref, watch } from "vue";
import { NButton } from "naive-ui";
import { useClipboard } from "@vueuse/core";
import type { UseFetchReturn } from "@vueuse/core";

export function CopyButton(
  props: { label: string; content: string; [key: string]: any },
  { attrs }: { attrs: any }
) {
  const { copy, copied } = useClipboard();

  return h(
    NButton,
    { ...props, ...attrs, onClick: () => copy(props.content) },
    () => (copied.value ? "copied!" : props.label)
  );
}

type ExecutionFunc<T> = () => UseFetchReturn<T>;

export function ExcuteButton<T>(
  props: {
    exec: ExecutionFunc<T>;
    label: string;
    loaddingLabel?: string;
    [key: string]: any;
  },
  { attrs }: { attrs: any }
) {
  const { execute, isFetching, error, response } = props.exec();

  const errorMsg = ref("");
  watch(error, (err: any) => {
    if (!err) {
      errorMsg.value = "";
      return;
    }
    errorMsg.value = err.value ?? response.value?.statusText ?? "Unknown Error";
    response.value?.text().then((v: string) => v && (errorMsg.value = v));
  });

  return h(() => [
    h(
      NButton,
      {
        ...props,
        ...attrs,
        loading: isFetching.value,
        onClick: () => execute(),
      },
      () =>
        isFetching.value
          ? props.loaddingLabel ?? props.label + "..."
          : props.label
    ),
    errorMsg.value &&
      h("p", {
        class: "text-red-300",
        innerText: errorMsg.value,
      }),
  ]);
}
