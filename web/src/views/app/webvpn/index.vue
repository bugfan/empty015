<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useWebvpn } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createWebvpn, updateWebvpn } from "@/api/webvpn";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";

defineOptions({
  name: "AppWebvpn"
});

const { t } = useI18n();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);

const {
  form,
  loading,
  columns,
  dataList,
  groupList,
  proxyList,
  pagination,
  onSearch,
  resetForm,
  handleDelete
} = useWebvpn(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("webvpn.addTitle", "添加 WebVPN 资源应用"),
    id: undefined,
    name: "",
    http_proxy_id: undefined,
    target_url: "",
    prefix: "",
    hosts: "",
    replace: "{}",
    replaceList: [],
    group_ids: [],
    is_protected: 1,
    status: 1,
    remark: ""
  };
}

function handleAddPage() {
  formInline.value = getDefaultFormInline();
  // Auto-select first wildcard proxy if available
  const defaultWildcardProxy = proxyList.value.find((p: any) =>
    (p.Hostname || p.hostname || "").includes("*")
  );
  if (defaultWildcardProxy) {
    formInline.value.http_proxy_id = defaultWildcardProxy.Id || defaultWildcardProxy.id;
  }
  showView.value = "new";
}

function handleEditPage(row: any) {
  let gIds: number[] = [];
  try {
    const raw = row.AllowedGroupIds || row.allowed_group_ids || "[]";
    gIds = typeof raw === "string" ? JSON.parse(raw) : raw;
  } catch (e) {}

  let replacePairs: { k: string; v: string }[] = [];
  try {
    const rawRep = row.Replace || row.replace || "{}";
    const repObj = typeof rawRep === "string" ? JSON.parse(rawRep) : rawRep;
    if (repObj && typeof repObj === "object") {
      replacePairs = Object.entries(repObj).map(([k, v]) => ({ k, v: String(v) }));
    }
  } catch (e) {}

  formInline.value = {
    title: `${t("webvpn.editTitle", "编辑 WebVPN 资源应用")} (${row.Name || row.name || row.Id || row.id})`,
    id: row.Id || row.id,
    name: row.Name || row.name,
    http_proxy_id: row.HttpProxyId || row.http_proxy_id,
    target_url: row.TargetURL || row.target_url,
    prefix: row.Prefix || row.prefix,
    hosts: row.Hosts || row.hosts || "",
    replace: row.Replace || row.replace || "{}",
    replaceList: replacePairs,
    group_ids: Array.isArray(gIds) ? gIds : [],
    is_protected: row.IsProtected ?? row.is_protected ?? 1,
    status: row.Status ?? row.status ?? 1,
    remark: row.Remark || row.remark || ""
  };
  showView.value = "edit";
}

function handleCancelPage() {
  showView.value = "list";
}

async function handleSaveSubmit() {
  const childComp = createEditFormRef.value;
  if (!childComp) return;
  const formRef = childComp.getRef();
  if (!formRef) return;

  await formRef.validate(async (valid: boolean) => {
    if (!valid) return;

    saving.value = true;
    try {
      const currentForm = childComp.newFormInline || formInline.value;
      const replaceMap: Record<string, string> = {};
      if (currentForm.replaceList && Array.isArray(currentForm.replaceList)) {
        for (const item of currentForm.replaceList) {
          if (item.k && item.k.trim()) {
            replaceMap[item.k.trim()] = item.v || "";
          }
        }
      }

      const payload: any = {
        name: currentForm.name,
        http_proxy_id: currentForm.http_proxy_id,
        target_url: currentForm.target_url,
        prefix: currentForm.prefix,
        hosts: currentForm.hosts || "",
        replace: JSON.stringify(replaceMap),
        allowed_group_ids: JSON.stringify(currentForm.group_ids || []),
        is_protected: currentForm.is_protected ?? 1,
        status: currentForm.status,
        remark: currentForm.remark
      };

      if (showView.value === "new") {
        const res = await createWebvpn(payload);
        if (res.code === 0) {
          message(t("webvpn.addSuccess", "创建成功"), { type: "success" });
          showView.value = "list";
          onSearch();
        } else {
          message(res.message || t("common.failed", "操作失败"), { type: "error" });
        }
      } else {
        const res = await updateWebvpn(formInline.value.id, payload);
        if (res.code === 0) {
          message(t("webvpn.updateSuccess", "更新成功"), { type: "success" });
          showView.value = "list";
          onSearch();
        } else {
          message(res.message || t("common.failed", "操作失败"), { type: "error" });
        }
      }
    } catch (e: any) {
      message(e.message || t("common.failed", "操作失败"), { type: "error" });
    } finally {
      saving.value = false;
    }
  });
}
</script>

<template>
  <div class="main">
    <!-- 1. 列表视图 (直接展示搜索表单和表格，顶部无冗余 PageHeader) -->
    <div v-if="showView === 'list'">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('webvpn.name', '名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('webvpn.name', '名称')"
            clearable
            class="w-full sm:w-50!"
            @keyup.enter="onSearch"
          />
        </el-form-item>
        <el-form-item :label="t('webvpn.boundSite', '泛域名')" prop="http_proxy_id">
          <el-select
            v-model="form.http_proxy_id"
            clearable
            :placeholder="t('webvpn.allSites', '全部站点')"
            class="w-full sm:w-55!"
            @change="onSearch"
          >
            <el-option
              v-for="item in proxyList"
              :key="item.Id || item.id"
              :label="item.Name || item.name"
              :value="item.Id || item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="onSearch">
            {{ t("common.search", "搜索") }}
          </el-button>
          <el-button @click="resetForm(searchFormRef)">
            {{ t("common.reset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('webvpn.title', 'WebVPN')"
        :columns="columns"
        @refresh="onSearch"
      >
        <template #buttons>
          <el-button
            type="primary"
            :icon="useRenderIcon(AddFill)"
            @click="handleAddPage"
          >
            {{ t("buttons.pureAdd", "添加") }}
          </el-button>
        </template>
        <template v-slot="{ size, dynamicColumns }">
          <pure-table
            ref="tableRef"
            row-key="id"
            adaptive
            :adaptiveConfig="{ offsetBottom: 108 }"
            align-whole="center"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="pagination"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)',
              fontWeight: 'bold'
            }"
            @page-size-change="onSearch"
            @page-current-change="onSearch"
          >
            <template #operation="{ row }">
              <div class="flex items-center justify-center gap-2 whitespace-nowrap">
                <el-button
                  class="reset-margin"
                  link
                  type="primary"
                  :size="size"
                  :icon="useRenderIcon(EditPen)"
                  @click="handleEditPage(row)"
                >
                  {{ t("common.edit", "编辑") }}
                </el-button>
                <el-popconfirm
                  :title="t('webvpn.deleteConfirm', { name: row.Name || row.name })"
                  @confirm="handleDelete(row)"
                >
                  <template #reference>
                    <el-button
                      class="reset-margin"
                      link
                      type="danger"
                      :size="size"
                      :icon="useRenderIcon(Delete)"
                    >
                      {{ t("common.delete", "删除") }}
                    </el-button>
                  </template>
                </el-popconfirm>
              </div>
            </template>
          </pure-table>
        </template>
      </PureTableBar>
    </div>

    <!-- 2. 新增 / 编辑全屏视图 (右上角和右下角均有取消与保存按钮) -->
    <div
      v-else-if="showView === 'new' || showView === 'edit'"
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <!-- 顶部 Header 操作栏 -->
      <PageHeader
        :title="formInline.title"
        :description="t('webvpn.headerDesc', '配置 WebVPN 资源应用目标地址、专属子域名及授权用户组')"
        :backTitle="t('webvpn.backToList', '返回 WebVPN 列表')"
        @back="handleCancelPage"
      >
        <template #actions>
          <div class="flex items-center space-x-2">
            <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
              {{ t("common.cancel", "取消") }}
            </el-button>
            <el-button
              type="primary"
              :loading="saving"
              :icon="useRenderIcon(CheckIcon)"
              @click="handleSaveSubmit"
            >
              {{ t("common.save", "保存") }}
            </el-button>
          </div>
        </template>
      </PageHeader>

      <!-- 表单主体 -->
      <editForm
        ref="createEditFormRef"
        :formInline="formInline"
        :groupList="groupList"
        :proxyList="proxyList"
      />

      <!-- 底部操作按钮栏 (保持全局统一，提供便捷提交) -->
      <div
        class="flex items-center justify-end space-x-3 pt-4 mt-4 border-t border-(--el-border-color-lighter)"
      >
        <el-button :icon="useRenderIcon(CloseIcon)" @click="handleCancelPage">
          {{ t("common.cancel", "取消") }}
        </el-button>
        <el-button
          type="primary"
          :loading="saving"
          :icon="useRenderIcon(CheckIcon)"
          @click="handleSaveSubmit"
        >
          {{ t("common.save", "保存") }}
        </el-button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-form :deep(.el-form-item) {
  margin-bottom: 12px;
}

:deep(.el-table .el-table__header th.el-table__cell .cell) {
  white-space: nowrap !important;
  word-break: keep-all !important;
}
</style>
