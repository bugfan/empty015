<script setup lang="ts">
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { useWebvpnSite } from "./utils/hook";
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
import Refresh from "~icons/ep/refresh";
import Search from "~icons/ep/search";

defineOptions({
  name: "AppWebvpnSite"
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
  serviceList,
  groupList,
  fetchServices,
  fetchGroups,
  onSearch,
  resetForm,
  handleDelete
} = useWebvpnSite(t, tableRef);

function getDefaultFormInline() {
  return {
    title: t("webvpn.addTitle", "添加 WebVPN 站点"),
    id: undefined,
    name: "",
    service_id:
      serviceList.value[0]?.Id ||
      serviceList.value[0]?.id ||
      undefined,
    target_url: "",
    hosts: "",
    replaceList: [],
    is_protected: 1,
    group_ids: [],
    status: 1,
    remark: ""
  };
}

function getFormInlineFromRow(row: any) {
  let initialGroupIds: number[] = [];
  if (row?.AllowedGroupIds || row?.allowed_group_ids) {
    try {
      initialGroupIds = JSON.parse(
        row.AllowedGroupIds || row.allowed_group_ids
      );
    } catch (e) {}
  }

  let initialReplaceList: Array<{ k: string; v: string }> = [];
  if (row?.Replace || row?.replace) {
    try {
      const obj = JSON.parse(row.Replace || row.replace);
      initialReplaceList = Object.keys(obj).map(k => ({ k, v: obj[k] }));
    } catch (e) {}
  }

  return {
    title: `${t("webvpn.editTitle", "编辑 WebVPN 站点")} (${row.Name || row.name})`,
    id: row?.Id || row?.id,
    name: row?.Name || row?.name || "",
    service_id:
      row?.ServiceId ||
      row?.service_id ||
      row?.HttpProxyId ||
      row?.http_proxy_id ||
      (serviceList.value[0]?.Id || serviceList.value[0]?.id),
    target_url: row?.TargetURL || row?.target_url || "",
    hosts: row?.Hosts || row?.hosts || "",
    replaceList: initialReplaceList,
    is_protected: row?.IsProtected ?? row?.is_protected ?? 1,
    group_ids: initialGroupIds,
    status: row?.Status ?? row?.status ?? 1,
    remark: row?.Remark || row?.remark || ""
  };
}

async function handleAddPage() {
  await Promise.all([fetchServices(), fetchGroups()]);
  formInline.value = getDefaultFormInline();
  showView.value = "new";
}

async function handleEditPage(row: any) {
  await Promise.all([fetchServices(), fetchGroups()]);
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

        const replaceMap: Record<string, string> = {};
        if (formData.replaceList && formData.replaceList.length > 0) {
          formData.replaceList.forEach((item: any) => {
            const k = (item.k || "").trim();
            const v = item.v || "";
            if (k) {
              replaceMap[k] = v;
            }
          });
        }

        const payload = {
          name: formData.name,
          service_id: formData.service_id,
          target_url: formData.target_url,
          hosts: formData.hosts,
          replace: JSON.stringify(replaceMap),
          is_protected: formData.is_protected,
          allowed_group_ids: JSON.stringify(formData.group_ids || []),
          status: formData.status,
          remark: formData.remark
        };

        let res;
        if (showView.value === "edit") {
          res = await updateWebvpn(formData.id, payload);
        } else {
          res = await createWebvpn(payload);
        }

        if (res.code === 0) {
          message(
            showView.value === "edit"
              ? t("webvpn.updateSuccess", "更新 WebVPN 应用成功")
              : t("webvpn.addSuccess", "创建 WebVPN 应用成功"),
            { type: "success" }
          );
          showView.value = "list";
          onSearch();
        } else {
          message(
            res.message ||
              (showView.value === "edit"
                ? t("common.updateFailed", "更新失败")
                : t("common.addFailed", "创建失败")),
            {
              type: "error"
            }
          );
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
    <!-- List View Mode -->
    <div v-if="showView === 'list'">
      <el-form
        ref="searchFormRef"
        :inline="true"
        :model="form"
        class="search-form bg-bg_color w-full px-3 sm:px-6 pt-3 pb-1 overflow-auto mb-3 rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
      >
        <el-form-item :label="t('webvpn.name', '应用名称')" prop="name">
          <el-input
            v-model="form.name"
            :placeholder="t('webvpn.namePlaceholder', '请输入应用名称')"
            clearable
            class="w-full sm:w-50!"
            @keyup.enter="onSearch"
          />
        </el-form-item>

        <el-form-item :label="t('webvpn.service', '所属基础域')" prop="service_id">
          <el-select
            v-model="form.service_id"
            :placeholder="t('webvpn.servicePlaceholder', '选择所属基础域')"
            clearable
            class="w-full sm:w-56!"
          >
            <el-option
              v-for="s in serviceList"
              :key="s.Id || s.id"
              :label="`${s.Name || s.name} (${s.Hostname || s.hostname})`"
              :value="s.Id || s.id"
            />
          </el-select>
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
        :title="t('webvpn.title', 'WebVPN 站点')"
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

    <!-- Create / Edit Full Page Mode -->
    <div
      v-else-if="showView === 'new' || showView === 'edit'"
      class="p-3 sm:p-5 bg-bg_color rounded-xl border border-(--el-border-color-lighter) shadow-2xs"
    >
      <PageHeader
        :title="formInline.title"
        :description="t('webvpn.headerDesc', '配置 WebVPN 站点目标地址、所属服务、关联域名与用户组权限')"
        :backTitle="t('webvpn.backToList', '返回站点列表')"
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

      <!-- Form Component Embedded Directly (Full Width) -->
      <editForm
        ref="createEditFormRef"
        :formInline="formInline"
        :groupList="groupList"
        :serviceList="serviceList"
      />

      <!-- Bottom Action Bar -->
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
