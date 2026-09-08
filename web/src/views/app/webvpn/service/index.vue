<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useWebvpnService } from "./utils/hook";
import editForm from "./form/index.vue";
import PageHeader from "@/components/PageHeader/index.vue";
import { PureTableBar } from "@/components/RePureTableBar";
import { useRenderIcon } from "@/components/ReIcon/src/hooks";
import { message } from "@/utils/message";
import { createWebvpnService, updateWebvpnService } from "@/api/webvpn";

import Delete from "~icons/ep/delete";
import EditPen from "~icons/ep/edit-pen";
import AddFill from "~icons/ri/add-circle-line";
import CheckIcon from "~icons/ep/check";
import CloseIcon from "~icons/ep/close";
import Refresh from "~icons/ep/refresh";
import Search from "~icons/ep/search";

defineOptions({
  name: "AppWebvpnService"
});

const { t } = useI18n();
const searchFormRef = ref();
const tableRef = ref();
const createEditFormRef = ref();

// View Mode: 'list' | 'new' | 'edit'
const showView = ref<"list" | "new" | "edit">("list");
const formInline = ref<any>({});
const saving = ref(false);

const {
  form,
  loading,
  columns,
  dataList,
  pagination,
  onSearch,
  resetForm,
  handleDelete
} = useWebvpnService(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("webvpnService.addTitle", "添加基础域"),
    id: undefined,
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
  };
}

function getFormInlineFromRow(row: any) {
  return {
    title: `${t("webvpnService.editTitle", "编辑基础域")} (${row.Name || row.name})`,
    id: row.Id || row.id,
    name: row.Name || row.name || "",
    hostname: row.Hostname || row.hostname || "",
    port: String(row.Port || row.port || "443"),
    tls: (row.TLS ?? row.tls ?? true) !== false,
    h2: (row.H2 ?? row.h2 ?? true) !== false,
    certificate: row.Certificate || row.certificate || "",
    login_url: row.LoginURL || row.login_url || "",
    fallback: row.Fallback || row.fallback || "404",
    status: row.Status ?? row.status ?? 1,
    remark: row.Remark || row.remark || ""
  };
}

function handleAddPage() {
  formInline.value = getDefaultFormInline();
  showView.value = "new";
}

function handleEditPage(row: any) {
  formInline.value = getFormInlineFromRow(row);
  showView.value = "edit";
}

function handleCancelPage() {
  showView.value = "list";
}

async function handleSaveSubmit() {
  if (!createEditFormRef.value) return;
  const FormRef = createEditFormRef.value.getRef();
  if (!FormRef) return;

  FormRef.validate(async (valid: boolean) => {
    if (valid) {
      saving.value = true;
      try {
        const formData = createEditFormRef.value.newFormInline;
        const payload = {
          name: formData.name,
          hostname: formData.hostname,
          port: formData.port,
          tls: formData.tls,
          h2: formData.h2,
          certificate: formData.certificate,
          login_url: formData.login_url,
          fallback: formData.fallback,
          status: formData.status,
          remark: formData.remark
        };

        let res;
        if (showView.value === "edit") {
          res = await updateWebvpnService(formData.id, payload);
        } else {
          res = await createWebvpnService(payload);
        }

        if (res.code === 0) {
          message(
            showView.value === "edit"
              ? t("webvpnService.updateSuccess", "更新基础域成功")
              : t("webvpnService.addSuccess", "创建基础域成功"),
            { type: "success" }
          );
          showView.value = "list";
          onSearch();
        } else {
          message(res.message || t("common.failed", "操作失败"), {
            type: "error"
          });
        }
      } catch (err: any) {
        message(err.message || t("common.failed", "操作失败"), {
          type: "error"
        });
      } finally {
        saving.value = false;
      }
    }
  });
}
</script>

<template>
  <div class="main">
    <!-- List View -->
    <div v-if="showView === 'list'">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('webvpnService.name', '基础域名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('webvpnService.namePlaceholder', '请输入基础域名称')"
            clearable
            class="w-full sm:w-50!"
            @keyup.enter="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('webvpnService.hostname', '泛域名')" prop="hostname">
          <el-input
            v-model="form.hostname"
            :placeholder="t('webvpnService.hostnamePlaceholder', '请输入泛域名')"
            clearable
            class="w-full sm:w-50!"
            @keyup.enter="onSearch"
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            :icon="useRenderIcon(Search)"
            :loading="loading"
            @click="onSearch"
          >
            {{ t("buttons.pureSearch", "搜索") }}
          </el-button>
          <el-button :icon="useRenderIcon(Refresh)" @click="resetForm(searchFormRef)">
            {{ t("buttons.pureReset", "重置") }}
          </el-button>
        </el-form-item>
      </el-form>

      <PureTableBar
        :title="t('webvpnService.title', '基础域')"
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
            align-whole="center"
            table-layout="auto"
            :loading="loading"
            :size="size"
            :adaptive="true"
            :data="dataList"
            :columns="dynamicColumns"
            :pagination="pagination"
            :paginationSmall="size === 'small'"
            :header-cell-style="{
              background: 'var(--el-fill-color-light)',
              color: 'var(--el-text-color-primary)',
              fontWeight: 'bold'
            }"
            @page-size-change="onSearch"
            @page-current-change="onSearch"
          >
            <template #operation="{ row }">
              <div class="flex items-center justify-center space-x-2 whitespace-nowrap">
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
                  :title="t('webvpnService.deleteConfirm', { name: row.Name || row.name })"
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

    <!-- Create / Edit Full Page Mode -->
    <div
      v-else-if="showView === 'new' || showView === 'edit'"
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <PageHeader
        :title="formInline.title"
        :description="t('webvpnService.headerDesc', '配置 WebVPN 底座泛域名网关、监听端口、SSL 证书及安全策略')"
        :backTitle="t('webvpnService.backToList', '返回基础域列表')"
        @back="handleCancelPage"
      >
        <template #actions>
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
        </template>
      </PageHeader>

      <!-- Full-width form view -->
      <editForm ref="createEditFormRef" :formInline="formInline" />

      <!-- Bottom Actions -->
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

@media (max-width: 640px) {
  .search-form :deep(.el-form-item) {
    margin-right: 0;
    width: 100%;
  }
}
</style>
