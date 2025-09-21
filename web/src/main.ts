import "./style.css";
import { createApp, h } from "vue";

import App from "./App.vue";

import { NConfigProvider, NModalProvider, darkTheme } from "naive-ui";

import hljs from "highlight.js/lib/core";
import routeros from "highlight.js/lib/languages/routeros";
hljs.registerLanguage("routeros", routeros);

const nModel = () => h(NModalProvider, {}, () => h(App));
const nConfig = () => h(NConfigProvider, { theme: darkTheme, hljs }, nModel);

createApp(nConfig).mount("#app");
