import { http } from "@/utils/http";
import { formatApiError } from "@/utils/apiError";

// -------------------------------------------------------------
// WebVPN 服务 (WebvpnService) Types & APIs
// -------------------------------------------------------------

export type WebvpnServiceItem = {
  id?: number;
  name?: string;
  hostname?: string;
  port?: string;
  tls?: boolean;
  h2?: boolean;
  certificate?: string;
  login_url?: string;
  fallback?: string; // "404" | "login"
  status?: number; // 1: enabled, 0: disabled
  remark?: string;
  root_domain?: string;
  site_count?: number;
  created_at?: string;
  updated_at?: string;
};

export const getWebvpnServiceList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/webvpn-service", { params });
    const list = Array.isArray(res) ? res : (res?.data || res?.list || []);
    return {
      code: 0,
      message: "success",
      data: {
        list,
        total: list.length,
        pageSize: 10,
        currentPage: 1
      }
    };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "获取WebVPN服务列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createWebvpnService = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/webvpn-service", { data });
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "创建WebVPN服务失败")
    };
  }
};

export const updateWebvpnService = async (id: number, data?: object) => {
  try {
    const res = await http.request<any>("put", `/api/webvpn-service/${id}`, {
      data
    });
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "更新WebVPN服务失败")
    };
  }
};

export const deleteWebvpnService = async (id: number) => {
  try {
    const res = await http.request<any>("delete", `/api/webvpn-service/${id}`);
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "删除WebVPN服务失败")
    };
  }
};

// -------------------------------------------------------------
// WebVPN 站点 (WebvpnSite) Types & APIs
// -------------------------------------------------------------

export type WebvpnSiteItem = {
  id?: number;
  name?: string;
  service_id?: number;
  service_name?: string;
  service_hostname?: string;
  http_proxy_id?: number;
  http_proxy_name?: string;
  http_proxy_hostname?: string;
  target_url?: string;
  prefix?: string;
  hosts?: string;
  replace?: string;
  allowed_group_ids?: string; // JSON array string e.g. "[1, 2]"
  is_protected?: number; // 1: protected, 0: public
  status?: number; // 1: enabled, 0: disabled
  full_access_url?: string;
  remark?: string;
  created_at?: string;
  updated_at?: string;
};

export const getWebvpnList = async (params?: object) => {
  try {
    const res = await http.request<any>("get", "/api/webvpn-site", { params });
    const list = Array.isArray(res) ? res : (res?.data || res?.list || []);
    return {
      code: 0,
      message: "success",
      data: {
        list,
        total: list.length,
        pageSize: 10,
        currentPage: 1
      }
    };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "获取WebVPN应用列表失败"),
      data: { list: [], total: 0, pageSize: 10, currentPage: 1 }
    };
  }
};

export const createWebvpn = async (data?: object) => {
  try {
    const res = await http.request<any>("post", "/api/webvpn-site", { data });
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "创建WebVPN应用失败")
    };
  }
};

export const updateWebvpn = async (id: number, data?: object) => {
  try {
    const res = await http.request<any>("put", `/api/webvpn-site/${id}`, {
      data
    });
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "更新WebVPN应用失败")
    };
  }
};

export const deleteWebvpn = async (id: number) => {
  try {
    const res = await http.request<any>("delete", `/api/webvpn-site/${id}`);
    if (res && typeof res.code === "number" && res.code !== 0) return res;
    return { code: 0, message: "success", data: res };
  } catch (err: any) {
    return {
      code: 1,
      message: formatApiError(err, "webvpn", "删除WebVPN应用失败")
    };
  }
};

// Aliases for clear semantic naming
export const getWebvpnSiteList = getWebvpnList;
export const createWebvpnSite = createWebvpn;
export const updateWebvpnSite = updateWebvpn;
export const deleteWebvpnSite = deleteWebvpn;

