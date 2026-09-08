<script setup lang="ts">
import { ref, onMounted, watch } from "vue";
import { useI18n } from "vue-i18n";
import ReCol from "@/components/ReCol";
import { formRules } from "../utils/rule";
import { ServiceFormProps } from "../utils/types";
import { getCertList } from "@/api/certificate";

const props = withDefaults(defineProps<ServiceFormProps>(), {
  formInline: () => ({
    name: "",
    hostname: "",
    port: "443",
    tls: true,
    h2: true,
    certificate: "",
    login_url: "",
    fallback: "404",
    status: 1,
    remark: ""
  })
});

const { t } = useI18n();
const ruleFormRef = ref();
const newFormInline = ref(props.formInline);

watch(
  () => props.formInline,
  val => {
    newFormInline.value = val;
  },
  { deep: true }
);
const certOptions = ref<Array<{ label: string; value: string }>>([]);

async function fetchCertificates() {
  try {
    const res = await getCertList();
    if (res?.code === 0 && res?.data?.list) {
      certOptions.value = res.data.list.map((c: any) => {
        const idVal = c.CertId || c.cert_id || `id-${c.Id || c.id}`;
        const cnVal = c.SubjectCN || c.subject_cn || c.Name || c.name || idVal;
        return {
          label: `${cnVal} (${idVal})`,
          value: idVal
        };
      });
    }
  } catch (e) {}
}

onMounted(() => {
  fetchCertificates();
});

function getRef() {
  return ruleFormRef.value;
}

defineExpose({ getRef, newFormInline });
</script>

<template>
  <el-form
    ref="ruleFormRef"
    :model="newFormInline"
    :rules="formRules"
    label-width="140px"
    class="space-y-6"
  >
    <!-- Section 1: 基本信息 -->
    <el-card shadow="never" class="border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("webvpnService.basicSection", "基本信息") }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 1. 基础域名称 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.name', '基础域名称')" prop="name">
            <el-input
              v-model="newFormInline.name"
              clearable
              :placeholder="t('webvpnService.namePlaceholder', '如：主校区 WebVPN 网关')"
            />
          </el-form-item>
        </re-col>

        <!-- 2. 泛域名 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.hostname', '泛域名')" prop="hostname">
            <div class="flex flex-col w-full">
              <el-input
                v-model="newFormInline.hostname"
                clearable
                :placeholder="t('webvpnService.hostnamePlaceholder', '如：*.webvpn.example.com')"
              />
              <p class="text-xs text-gray-400 mt-2 leading-relaxed">
                {{ t("webvpnService.hostnameHint", "WebVPN 底座泛域名，必须以 *. 开头，例如 *.webvpn.example.com。") }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 3. 监听端口 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.port', '监听端口')" prop="port">
            <el-input
              v-model="newFormInline.port"
              clearable
              :placeholder="t('webvpnService.portPlaceholder', '443')"
            />
          </el-form-item>
        </re-col>

        <!-- 4. 安全协议 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.protocol', '安全协议')" :for="''">
            <div class="flex items-center gap-8">
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-600 dark:text-gray-300">TLS:</span>
                <el-switch v-model="newFormInline.tls" />
              </div>
              <div class="flex items-center gap-2">
                <span class="text-sm text-gray-600 dark:text-gray-300">HTTP/2:</span>
                <el-switch v-model="newFormInline.h2" />
              </div>
            </div>
          </el-form-item>
        </re-col>

        <!-- 5. SSL 证书 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.certificate', 'SSL 证书')" prop="certificate">
            <div class="flex flex-col w-full">
              <el-select
                v-model="newFormInline.certificate"
                filterable
                clearable
                class="w-full"
                :placeholder="t('webvpnService.certPlaceholder', '选择匹配的通配符 SSL 证书')"
              >
                <el-option
                  v-for="c in certOptions"
                  :key="c.value"
                  :label="c.label"
                  :value="c.value"
                />
              </el-select>
              <p class="text-xs text-gray-400 mt-2 leading-relaxed">
                {{ t("webvpnService.certHint", "请选择已在系统中颁发且涵盖该泛域名的通配符证书；留空时将尝试自动匹配。") }}
              </p>
            </div>
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>

    <!-- Section 2: 访问与安全策略 -->
    <el-card shadow="never" class="border-(--el-border-color-lighter)! rounded-xl">
      <template #header>
        <div class="flex items-center space-x-2">
          <div class="w-1.5 h-4 bg-primary rounded-full" />
          <span class="font-bold text-(--el-text-color-primary) text-sm sm:text-base">
            {{ t("webvpnService.policySection", "访问与安全策略") }}
          </span>
        </div>
      </template>

      <el-row :gutter="24">
        <!-- 6. 认证中心地址 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.loginUrl', '认证中心地址')" prop="login_url">
            <div class="flex flex-col w-full">
              <el-input
                v-model="newFormInline.login_url"
                clearable
                :placeholder="t('webvpnService.loginUrlPlaceholder', '选填，如 https://auth.example.com')"
              />
              <p class="text-xs text-gray-400 mt-2 leading-relaxed">
                {{ t("webvpnService.loginUrlHint", "用户未登录时跳转的认证登录页面；留空时自动关联系统内已配置的认证中心。") }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 7. 兜底策略 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.fallback', '未命中策略')" prop="fallback">
            <div class="flex flex-col w-full">
              <el-select v-model="newFormInline.fallback" class="w-full">
                <el-option
                  value="404"
                  :label="t('webvpnService.fallback404', '404 页面阻断（推荐，严防未知请求穿透）')"
                />
                <el-option
                  value="login"
                  :label="t('webvpnService.fallbackLogin', '重定向至认证中心登录页')"
                />
              </el-select>
              <p class="text-xs text-gray-400 mt-2 leading-relaxed">
                {{ t("webvpnService.fallbackHint", "当外部请求的子域名未在站点列表中注册或已被停用时的安全防护策略。") }}
              </p>
            </div>
          </el-form-item>
        </re-col>

        <!-- 8. 启用 -->
        <re-col :value="24">
          <el-form-item
            :label="t('webvpnService.status', '启用')"
            prop="status"
            :for="''"
          >
            <el-switch
              v-model="newFormInline.status"
              :active-value="1"
              :inactive-value="0"
              :active-text="t('webvpnService.statusEnabled', '启用')"
              :inactive-text="t('webvpnService.statusDisabled', '禁用')"
            />
          </el-form-item>
        </re-col>

        <!-- 9. 备注 -->
        <re-col :value="24">
          <el-form-item :label="t('webvpnService.remark', '备注')" prop="remark">
            <el-input
              v-model="newFormInline.remark"
              type="textarea"
              :rows="2"
              :placeholder="t('webvpnService.remarkPlaceholder', '选填，关于该 WebVPN 网关基础域的说明')"
            />
          </el-form-item>
        </re-col>
      </el-row>
    </el-card>
  </el-form>
</template>
