/** 错误日志 API */
import { requestClient } from '#/api/request';
import { formatDateTime } from '#/utils/time';

export namespace ErrorLogApi {
  export interface ErrorLog {
    id: string;
    errorMessage: string;
    method: string;
    path: string;
    ip: string;
    occurTime: string;
    status: number;
    createdAt: string;
  }

  export interface ListParams {
    path?: string;
    startTime?: string;
    endTime?: string;
    status?: number;
    page?: number;
    pageSize?: number;
  }
}

function normalizeLogTime(value: any) {
  return formatDateTime(value ?? '');
}

function normalizeDateParam(value?: string, endOfDay = false) {
  if (!value) return undefined;
  if (value.includes('T')) return value;
  return `${value}T${endOfDay ? '23:59:59' : '00:00:00'}+08:00`;
}

function normalizeLog(data: any): ErrorLogApi.ErrorLog {
  const createdAt = normalizeLogTime(data.createdAt ?? data.created_at);
  return {
    ...data,
    id: String(data.id ?? ''),
    errorMessage: data.errorMessage ?? data.error_message ?? data.message ?? '',
    method: String(data.method ?? '').toUpperCase(),
    path: data.path ?? '',
    ip: data.ip ?? data.IP ?? '',
    occurTime: normalizeLogTime(data.occurTime ?? data.createdAt ?? data.created_at),
    status: data.status ?? 0,
    createdAt,
  };
}

/** 获取错误日志列表 */
export async function getErrorLogListApi(params: ErrorLogApi.ListParams) {
  const data = await requestClient.get<{ items?: any[]; list?: any[]; total?: number }>('/error-logs', {
    params: {
      page: params.page,
      pageSize: params.pageSize,
      path: params.path,
      status: params.status,
      start_time: normalizeDateParam(params.startTime),
      end_time: normalizeDateParam(params.endTime, true),
    },
  });
  const rawList = data.list ?? data.items ?? [];
  return {
    list: rawList.map(normalizeLog),
    total: data.total ?? 0,
  };
}

/** 获取错误日志详情 */
export async function getErrorLogApi(id: string) {
  const data = await requestClient.get<{ items?: any[]; list?: any[]; total?: number }>('/error-logs', {
    params: { page: 1, pageSize: 500 },
  });
  const item = (data.list ?? data.items ?? []).find((entry) => String(entry.id) === id);
  if (!item) {
    throw new Error('未找到，请求的资源不存在。');
  }
  return normalizeLog(item);
}

/** 标记错误日志为已解决 */
export async function resolveErrorLogApi(id: string) {
  return requestClient.post(`/error-logs/${id}`, { status: 1 });
}

/** 删除错误日志 */
export async function deleteErrorLogApi(id: string) {
  return requestClient.post(`/error-logs/delete/${id}`);
}
