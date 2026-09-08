export interface ReplaceItem {
  k: string;
  v: string;
}

export interface FormItemProps {
  id?: number;
  name: string;
  service_id?: number;
  http_proxy_id?: number;
  target_url: string;
  prefix: string;
  hosts: string;
  replace?: string;
  replaceList?: ReplaceItem[];
  is_protected: number;
  group_ids: number[];
  status: number;
  remark: string;
}

export interface FormProps {
  formInline: FormItemProps;
}
