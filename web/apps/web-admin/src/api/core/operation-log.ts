/** 操作日志 API */
import { requestClient } from '#/api/request';
import { formatDateTime } from '#/utils/time';

export namespace OperationLogApi {
  export interface OperationLog {
    id: string;
    operator: string;
    operation: string;
    method: string;
    path: string;
    requestBody?: string;
    ip: string;
    duration: number;
    responseStatus?: number;
    createdAt: string;
  }

  export interface ListParams {
    operator?: string;
    startTime?: string;
    endTime?: string;
    page?: number;
    pageSize?: number;
  }
}

function getOperationAction(method: string, path: string) {
  const normalizedPath = path.toLowerCase();
  if (normalizedPath.includes('/batch-delete')) return '批量删除';
  if (normalizedPath.includes('/delete')) return '删除';
  if (normalizedPath.includes('/password')) return '修改密码';
  if (normalizedPath.includes('/roles') && method === 'POST') return '分配角色';
  if (normalizedPath.includes('/permission-config')) return method === 'POST' ? '保存权限配置' : '获取权限配置';
  if (method === 'GET') return '查询';
  if (method === 'POST' && /\/\d+$/.test(normalizedPath)) return '更新';
  if (method === 'POST') return '提交';
  if (method === 'PUT' || method === 'PATCH') return '更新';
  if (method === 'DELETE') return '删除';
  return '操作';
}

function buildOperationContent(data: any) {
  const method = String(data.method ?? '').toUpperCase();
  const path = String(data.path ?? '');
  const status = data.responseStatus ?? data.response_status;
  const action = getOperationAction(method, path);
  const target = [method, path].filter(Boolean).join(' ');
  return `${action}${target ? ` ${target}` : ''}${status ? `（${status}）` : ''}`;
}

function normalizeLogTime(value: any) {
  return formatDateTime(value ?? '');
}

function normalizeDateParam(value?: string, endOfDay = false) {
  if (!value) return undefined;
  if (value.includes('T')) return value;
  return `${value}T${endOfDay ? '23:59:59' : '00:00:00'}+08:00`;
}

function normalizeLog(data: any): OperationLogApi.OperationLog {
  const method = String(data.method ?? '').toUpperCase();
  const path = String(data.path ?? '');
  const responseStatus = data.responseStatus ?? data.response_status;
  return {
    ...data,
    id: String(data.id ?? ''),
    operator: data.operator ?? data.username ?? '',
    operation: data.operation || buildOperationContent(data),
    method,
    path,
    requestBody: data.requestBody ?? data.request_body ?? '',
    ip: data.ip ?? data.IP ?? '',
    duration: data.duration ?? data.latencyMs ?? data.latency_ms ?? 0,
    responseStatus,
    createdAt: normalizeLogTime(data.createdAt ?? data.created_at),
  };
}

/** 获取操作日志列表 */
export async function getOperationLogListApi(params: OperationLogApi.ListParams) {
  const data = await requestClient.get<{ items?: any[]; list?: any[]; total?: number }>('/operation-logs', {
    params: {
      page: params.page,
      pageSize: params.pageSize,
      username: params.operator,
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

/** 删除操作日志 */
export async function deleteOperationLogApi(id: string) {
  return requestClient.post(`/operation-logs/delete/${id}`);
}
