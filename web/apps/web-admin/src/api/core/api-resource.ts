/** API 资源管理 API */
import { requestClient } from '#/api/request';

export namespace ApiResourceApi {
  export interface ApiResource {
    id: string;
    path: string;
    method: string;
    description?: string;
    group: string;
    menuId?: string;
    actionType?: string;
    permissionMode?: string;
    ignored?: number;
    createdAt: string;
    updatedAt?: string;
  }

  export interface ListParams {
    path?: string;
    group?: string;
    method?: string;
    page?: number;
    pageSize?: number;
  }
}

function normalizeApiResource(data: any): ApiResourceApi.ApiResource {
  return {
    ...data,
    id: String(data.id ?? data.ID ?? ''),
    path: data.path ?? data.Path ?? '',
    method: data.method ?? data.Method ?? '',
    description: data.description ?? data.Description ?? '',
    group: data.group ?? data.Group ?? '',
    menuId: String(data.menuId ?? data.menu_id ?? data.MenuID ?? ''),
    actionType: data.actionType ?? data.action_type ?? data.ActionType ?? 'other',
    permissionMode: data.ignored === 1 || data.Ignored === 1 ? 'ignored' : (data.permissionMode ?? data.permission_mode ?? data.PermissionMode ?? 'casbin'),
    ignored: data.ignored ?? data.Ignored ?? 0,
    createdAt: data.createdAt ?? data.created_at ?? data.CreatedAt ?? '',
    updatedAt: data.updatedAt ?? data.updated_at ?? data.UpdatedAt ?? '',
  };
}

function normalizePagedList<T>(data: { items?: T[]; list?: T[]; total?: number }) {
  return {
    list: data.list ?? data.items ?? [],
    total: data.total ?? 0,
  };
}

function serializeApiResource(data: Partial<ApiResourceApi.ApiResource>) {
  return {
    path: data.path,
    method: data.method,
    group: data.group,
    description: data.description,
    menu_id: data.menuId ? Number(data.menuId) : 0,
    action_type: data.actionType || 'other',
    permission_mode: data.permissionMode || 'casbin',
    ignored: data.permissionMode === 'ignored' ? 1 : (data.ignored ?? 0),
  };
}

/** 获取 API 资源列表 */
export async function getApiResourceListApi(params: ApiResourceApi.ListParams) {
  const data = await requestClient.get<{ items?: ApiResourceApi.ApiResource[]; list?: ApiResourceApi.ApiResource[]; total?: number }>('/apis', {
    params,
  });
  const normalized = normalizePagedList(data);
  return {
    list: normalized.list.map(normalizeApiResource),
    total: normalized.total,
  };
}

/** 获取 API 资源详情 */
export async function getApiResourceApi(id: string) {
  const data = await requestClient.get<{ items?: ApiResourceApi.ApiResource[]; list?: ApiResourceApi.ApiResource[]; total?: number }>('/apis', {
    params: { page: 1, pageSize: 500 },
  });
  const item = (data.list ?? data.items ?? []).map(normalizeApiResource).find((entry) => String(entry.id) === id);
  if (!item) {
    throw new Error('未找到，请求的资源不存在。');
  }
  return item;
}

/** 创建 API 资源 */
export async function createApiResourceApi(data: Partial<ApiResourceApi.ApiResource>) {
  return requestClient.post('/apis', serializeApiResource(data));
}

/** 更新 API 资源 */
export async function updateApiResourceApi(id: string, data: Partial<ApiResourceApi.ApiResource>) {
  return requestClient.post(`/apis/${id}`, serializeApiResource(data));
}

/** 删除 API 资源 */
export async function deleteApiResourceApi(id: string) {
  return requestClient.post(`/apis/delete/${id}`);
}

/** 批量删除 API 资源 */
export async function batchDeleteApiResourceApi(ids: string[]) {
  return requestClient.post('/apis/batch-delete', { ids });
}
