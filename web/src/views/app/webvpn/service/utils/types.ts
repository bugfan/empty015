export interface ServiceFormItemProps {
  id?: number;
  name: string;
  hostname: string;
  port: string;
  tls: boolean;
  h2: boolean;
  certificate: string;
  login_url: string;
  fallback: string;
  status: number;
  remark: string;
}

export interface ServiceFormProps {
  formInline: ServiceFormItemProps;
}
