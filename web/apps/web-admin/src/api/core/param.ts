/** 系统参数管理 API */
import { requestClient } from '#/api/request';

export namespace ParamApi {
  export interface Param {
    id: string;
    paramKey: string;
    paramValue: string;
    category: string;
    description?: string;
    status: number;
    createdAt: string;
    updatedAt?: string;
  }

  export interface ListParams {
    paramKey?: string;
    category?: string;
    status?: number;
    page?: number;
    pageSize?: number;
  }
}

function normalizeParam(data: any): ParamApi.Param {
  return {
    ...data,
    id: String(data.id ?? data.ID ?? ''),
    paramKey: data.paramKey ?? data.key ?? data.Key ?? '',
    paramValue: data.paramValue ?? data.value ?? data.Value ?? '',
    category: data.category ?? data.Category ?? '',
    description: data.description ?? data.Description ?? '',
    status: data.status === 2 ? 0 : (data.status ?? data.Status ?? 1),
    createdAt: data.createdAt ?? data.created_at ?? data.CreatedAt ?? '',
    updatedAt: data.updatedAt ?? data.updated_at ?? data.UpdatedAt ?? '',
  };
}

/** 获取参数列表 */
export async function getParamListApi(params: ParamApi.ListParams) {
  const data = await requestClient.get<{ items?: any[]; list?: any[]; total?: number }>('/params', {
    params: {
      page: params.page,
      pageSize: params.pageSize,
      category: params.category,
      key: params.paramKey,
    },
  });
  const rawList = data.list ?? data.items ?? [];
  return {
    list: rawList.map(normalizeParam),
    total: data.total ?? 0,
  };
}

/** 获取参数详情 */
export async function getParamApi(id: string) {
  const data = await requestClient.get<any>(`/params/${id}`);
  return normalizeParam(data);
}

/** 创建参数 */
export async function createParamApi(data: Partial<ParamApi.Param>) {
  return requestClient.post('/params', {
    key: data.paramKey,
    value: data.paramValue,
    category: data.category,
    description: data.description,
    status: data.status === 0 ? 2 : data.status,
  });
}

/** 更新参数 */
export async function updateParamApi(id: string, data: Partial<ParamApi.Param>) {
  return requestClient.post(`/params/${id}`, {
    key: data.paramKey,
    value: data.paramValue,
    category: data.category,
    description: data.description,
    status: data.status === 0 ? 2 : data.status,
  });
}

/** 删除参数 */
export async function deleteParamApi(id: string) {
  return requestClient.post(`/params/delete/${id}`);
}
