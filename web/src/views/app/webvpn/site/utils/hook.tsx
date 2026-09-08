import { reactive, ref, onMounted } from "vue";
import type { PaginationProps } from "@pureadmin/table";
import { message } from "@/utils/message";
import {
  getWebvpnList,
  createWebvpn,
  updateWebvpn,
  deleteWebvpn,
  getWebvpnServiceList,
  type WebvpnSiteItem,
  type WebvpnServiceItem
} from "@/api/webvpn";
import { getUserGroupList, type UserGroupItem } from "@/api/user-group";

export function useWebvpnSite(t: Function, tableRef: any) {
  const form = reactive({
    name: "",
    service_id: ""
  });
  const dataList = ref<WebvpnSiteItem[]>([]);
  const groupList = ref<UserGroupItem[]>([]);
  const groupMap = ref<Record<number, string>>({});
  const serviceList = ref<WebvpnServiceItem[]>([]);
  const loading = ref(true);

  const pagination = reactive<PaginationProps>({
    total: 0,
    pageSize: 10,
    currentPage: 1,
    background: true
  });

  async function fetchGroups() {
    try {
      const res = await getUserGroupList();
      groupList.value = res.data.list;
      const gMap: Record<number, string> = {};
      res.data.list.forEach((g: any) => {
        const id = g.Id || g.id;
        gMap[id] = g.Name || g.name;
      });
      groupMap.value = gMap;
    } catch (e) {}
  }

  async function fetchServices() {
    try {
      const res = await getWebvpnServiceList();
      serviceList.value = res.data.list || [];
    } catch (e) {}
  }

  const columns: TableColumnList = [
    {
      label: "ID",
      align: "center",
      prop: "Id",
      width: 70,
      formatter: row => row.Id || row.id
    },
    {
      label: t("webvpn.name", "名称"),
      align: "center",
      prop: "Name",
      minWidth: 140,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.name", "名称")}</span>
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
      label: t("webvpn.targetUrl", "站点地址"),
      align: "center",
      prop: "TargetURL",
      minWidth: 180,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.targetUrl", "站点地址")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const targetUrl = row.TargetURL || row.target_url || "-";
        return (
          <el-tooltip content={targetUrl} placement="top" show-after={500}>
            <el-tag
              type="primary"
              effect="light"
              class="font-mono font-bold truncate max-w-[240px] inline-block"
            >
              {targetUrl}
            </el-tag>
          </el-tooltip>
        );
      }
    },
    {
      label: t("webvpn.accessUrl", "访问地址"),
      align: "center",
      minWidth: 200,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.accessUrl", "访问地址")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const fullUrl = row.full_access_url || "";
        if (fullUrl) {
          return (
            <el-tooltip content={fullUrl} placement="top" show-after={500}>
              <a
                href={fullUrl}
                target="_blank"
                rel="noreferrer"
                class="font-mono text-xs text-primary hover:underline truncate max-w-[240px] inline-block"
              >
                {fullUrl}
              </a>
            </el-tooltip>
          );
        }
        return <span class="text-xs text-gray-400">-</span>;
      }
    },
    {
      label: t("webvpn.service", "所属基础域"),
      align: "center",
      minWidth: 160,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.service", "所属基础域")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const sName = row.service_name || row.http_proxy_name || "-";
        const sHost = row.service_hostname || row.http_proxy_hostname || "";
        return (
          <div class="flex flex-col items-center">
            <span class="text-xs font-medium text-(--el-text-color-primary)">{sName}</span>
            {sHost ? (
              <span class="text-xs text-gray-400 font-mono">{sHost}</span>
            ) : null}
          </div>
        );
      }
    },
    {
      label: t("webvpn.accessMode", "公开访问"),
      align: "center",
      width: 100,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.accessMode", "公开访问")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const isProtected = (row.IsProtected ?? row.is_protected ?? 1) === 1;
        return isProtected ? (
          <el-tag size="small" type="warning" effect="light">
            {t("webvpn.modeProtected", "不公开")}
          </el-tag>
        ) : (
          <el-tag size="small" type="success" effect="light">
            {t("webvpn.modePublic", "公开")}
          </el-tag>
        );
      }
    },
    {
      label: t("webvpn.allowedGroups", "用户组"),
      align: "center",
      prop: "allowed_group_ids",
      minWidth: 150,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.allowedGroups", "用户组")}</span>
      ),
      cellRenderer: scope => {
        const row = scope.row;
        const isProtected = (row.IsProtected ?? row.is_protected ?? 1) === 1;
        if (!isProtected) {
          return (
            <span class="text-xs text-gray-400">
              {t("webvpn.publicGroupTip", "公开免登录开放，无需指定用户组")}
            </span>
          );
        }
        let gIds: number[] = [];
        try {
          const raw = row.AllowedGroupIds || row.allowed_group_ids;
          if (raw) gIds = JSON.parse(raw);
        } catch (e) {}

        if (!gIds || gIds.length === 0) {
          return (
            <el-tag size="small" type="info" effect="plain">
              {t("webvpn.allUsersAllowed", "全员开放")}
            </el-tag>
          );
        }

        return (
          <div class="flex flex-wrap gap-1 justify-center">
            {gIds.map((id: number) => (
              <el-tag key={id} size="small" type="primary" effect="light">
                {groupMap.value[id] || `#${id}`}
              </el-tag>
            ))}
          </div>
        );
      }
    },
    {
      label: t("webvpn.status", "启用"),
      align: "center",
      prop: "Status",
      width: 90,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.status", "启用")}</span>
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
      label: t("webvpn.remark", "备注"),
      align: "center",
      prop: "Remark",
      minWidth: 120,
      headerRenderer: () => (
        <span class="whitespace-nowrap">{t("webvpn.remark", "备注")}</span>
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
      const res = await getWebvpnList({
        name: form.name,
        service_id: form.service_id
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
      const res = await updateWebvpn(id, { status });
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
      const res = await deleteWebvpn(id);
      if (res.code === 0) {
        message(t("webvpn.delSuccess", "删除 WebVPN 应用成功"), { type: "success" });
        onSearch();
      } else {
        message(res.message || t("common.deleteFailed", "删除失败"), { type: "error" });
      }
    } catch (err: any) {
      message(err.message || t("common.deleteFailed", "删除失败"), { type: "error" });
    }
  }

  onMounted(() => {
    fetchServices();
    fetchGroups();
    onSearch();
  });

  return {
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
  };
}
