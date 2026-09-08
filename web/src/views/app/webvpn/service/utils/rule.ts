import { reactive } from "vue";
import type { FormRules } from "element-plus";
import { $t, transformI18n } from "@/plugins/i18n";

export const formRules = reactive<FormRules>({
  name: [
    {
      required: true,
      message: transformI18n($t("webvpnService.nameRequired", "服务名称不能为空")),
      trigger: "blur"
    }
  ],
  hostname: [
    {
      required: true,
      message: transformI18n($t("webvpnService.hostnameRequired", "泛域名不能为空")),
      trigger: "blur"
    }
  ],
  port: [
    {
      required: true,
      message: transformI18n($t("webvpnService.portRequired", "监听端口不能为空")),
      trigger: "blur"
    }
  ]
});
