/** 登录日志 API */
import { requestClient } from '#/api/request';

export namespace LoginLogApi {
  export interface LoginLog {
    id: string;
    username: string;
    status: number;
    ip: string;
    loginTime: string;
    description?: string;
    createdAt: string;
  }

  export interface ListParams {
    username?: string;
    status?: number;
    startTime?: string;
    endTime?: string;
    page?: number;
    pageSize?: number;
  }
}

function normalizeLog(data: any): LoginLogApi.LoginLog {
  return {
    ...data,
    id: String(data.id ?? ''),
    loginTime: data.loginTime ?? data.createdAt ?? '',
    createdAt: data.createdAt ?? data.created_at ?? '',
  };
}

/** 获取登录日志列表 */
export async function getLoginLogListApi(params: LoginLogApi.ListParams) {
  const data = await requestClient.get<{ items?: any[]; list?: any[]; total?: number }>('/login-logs', {
    params: {
      page: params.page,
      pageSize: params.pageSize,
      username: params.username,
      status: params.status,
      start_time: params.startTime,
      end_time: params.endTime,
    },
  });
  const rawList = data.list ?? data.items ?? [];
  return {
    list: rawList.map(normalizeLog),
    total: data.total ?? 0,
  };
}

/** 删除登录日志 */
export async function deleteLoginLogApi(id: string) {
  return requestClient.post(`/login-logs/delete/${id}`);
}
