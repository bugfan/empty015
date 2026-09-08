import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getWebvpnServiceList,
  createWebvpnService,
  updateWebvpnService,
  deleteWebvpnService,
  type WebvpnServiceItem
} from "@/api/webvpn";

export function useWebvpnService(t: Function, tableRef: any) {
  const form = reactive({
    name: "",
    hostname: ""
  });
  const dataList = ref<WebvpnServiceItem[]>([]);
  const loading = ref(true);

  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  const columns: TableColumnList = [
    {
      label: "ID",
      align: "center",
      prop: "Id",
      width: 70,
      formatter: row => row.Id || row.id
    },
    {
      label: t("webvpnService.name", "基础域名称"),
      align: "center",
      prop: "Name",
      minWidth: 140,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.name", "基础域名称")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        return (
          <span class="font-semibold text-sm text-(--el-text-color-primary)">
            {row.Name || row.name}
          </span>
        );
      }
    },
    {
      label: t("webvpnService.hostname", "泛域名"),
      align: "center",
      prop: "Hostname",
      minWidth: 180,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.hostname", "泛域名")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const host = row.Hostname || row.hostname || "-";
        return (
          <el-tag
            type="primary"
            effect="light"
            class="font-mono font-bold whitespace-nowrap"
          >
            {host}
          </el-tag>
        );
      }
    },
    {
      label: t("webvpnService.port", "端口"),
      align: "center",
      prop: "Port",
      width: 85,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.port", "端口")}</span>
      ),
      formatter: row => row.Port || row.port || "443"
    },
    {
      label: t("webvpnService.protocol", "安全协议"),
      align: "center",
      width: 130,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.protocol", "安全协议")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const tls = row.TLS ?? row.tls;
        const h2 = row.H2 ?? row.h2;
        return (
          <div class="flex items-center justify-center gap-1">
            <el-tag size="small" type={tls ? "success" : "info"} effect="light">
              {tls ? "TLS" : "No-TLS"}
            </el-tag>
            {h2 ? (
              <el-tag size="small" type="warning" effect="light">
                H2
              </el-tag>
            ) : null}
          </div>
        );
      }
    },
    {
      label: t("webvpnService.certificate", "SSL 证书"),
      align: "center",
      prop: "Certificate",
      minWidth: 140,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.certificate", "SSL 证书")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const cert = row.Certificate || row.certificate;
        return cert ? (
          <span class="text-xs text-gray-600 dark:text-gray-300">{cert}</span>
        ) : (
          <span class="text-xs text-gray-400">{t("common.auto", "自动匹配")}</span>
        );
      }
    },
    {
      label: t("webvpnService.fallback", "未命中策略"),
      align: "center",
      prop: "Fallback",
      width: 120,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.fallback", "未命中策略")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const fb = row.Fallback || row.fallback;
        return fb === "login" ? (
          <el-tag size="small" type="warning">
            {t("webvpnService.fallbackLoginShort", "重定向登录")}
          </el-tag>
        ) : (
          <el-tag size="small" type="danger">
            {t("webvpnService.fallback404Short", "404 阻断")}
          </el-tag>
        );
      }
    },
    {
      label: t("webvpnService.siteCount", "站点数"),
      align: "center",
      width: 90,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.siteCount", "站点数")}</span>
      ),
      cellRenderer: scope => {
        const count = scope.row.site_count ?? 0;
        return (
          <el-tag size="small" effect="plain" round>
            {count}
          </el-tag>
        );
      }
    },
    {
      label: t("webvpnService.status", "状态"),
      align: "center",
      prop: "Status",
      width: 90,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.status", "状态")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const id = row.Id || row.id;
        const isEnabled = (row.Status ?? row.status) === 1;
        return (
          <el-switch
            modelValue={isEnabled}
            active-value={true}
            inactive-value={false}
            onChange={(val: boolean) => handleStatusChange(id, val ? 1 : 0)}
          />
        );
      }
    },
    {
      label: t("webvpnService.remark", "备注"),
      align: "center",
      prop: "Remark",
      minWidth: 120,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpnService.remark", "备注")}</span>
      ),
      formatter: row => row.Remark || row.remark || "-"
    },
    {
      label: t("common.operations", "操作"),
      fixed: "right",
      width: 160,
      align: "center",
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("common.operations", "操作")}</span>
      ),
      slot: "operation"
    }
  ];

  async function onSearch() {
    loading.value = true;
    try {
      const res = await getWebvpnServiceList({
        name: form.name,
        hostname: form.hostname
      });
      dataList.value = res.data.list;
      pagination.total = res.data.total;
    } catch (e) {
    } finally {
      loading.value = false;
    }
  }

  function resetForm(formEl: any) {
    if (!formEl) return;
    formEl.resetFields();
    onSearch();
  }

  async function handleStatusChange(id: number, status: number) {
    try {
      const res = await updateWebvpnService(id, { status });
      if (res.code === 0) {
        message(t("common.updateSuccess", "状态更新成功"), { type: "success" });
        onSearch();
      } else {
        message(res.message || t("common.updateFailed", "更新失败"), { type: "error" });
      }
    } catch (err: any) {
      message(err.message || t("common.updateFailed", "更新失败"), { type: "error" });
    }
  }

  async function handleDelete(row: any) {
    const id = row.Id || row.id;
    try {
      const res = await deleteWebvpnService(id);
      if (res.code === 0) {
        message(t("webvpnService.delSuccess", "删除基础域成功"), { type: "success" });
        onSearch();
      } else {
        message(res.message || t("common.deleteFailed", "删除失败"), { type: "error" });
      }
    } catch (err: any) {
      message(err.message || t("common.deleteFailed", "删除失败"), { type: "error" });
    }
  }

  onMounted(() => {
    onSearch();
  });

  return {
    form,
    loading,
    columns,
    dataList,
    pagination,
    onSearch,
    resetForm,
    handleDelete
  };
}
