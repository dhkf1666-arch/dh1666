import { createApp } from "vue";
import ElementPlus from "element-plus";
import "element-plus/dist/index.css";
import * as ElementPlusIconsVue from "@element-plus/icons-vue";
import zhCn from "element-plus/dist/locale/zh-cn.mjs";

import App from "./App.vue";
import router from "./router/index";
import pinia from "./store/index";
import { setupPermissionDirective } from "./store/permission";

const app = createApp(App);

// 注册所有 Element Plus 图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}

// 使用 Element Plus 并设置中文语言
app.use(ElementPlus, {
  locale: zhCn,
});

app.use(router);
app.use(pinia);

// 设置权限指令
setupPermissionDirective(app);

app.mount("#app");
