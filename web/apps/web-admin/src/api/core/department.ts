import { requestClient } from '#/api/request';

export namespace DepartmentApi {
  export interface Department {
    id: string;
    parentId: string;
    name: string;
    leaderUserId: string;
    sort: number;
    status: number;
    createdAt: string;
    updatedAt?: string;
    children?: Department[];
  }

  export interface ListParams {
    name?: string;
    status?: number;
  }
}

function normalizeDepartment(data: any): DepartmentApi.Department {
  return {
    ...data,
    id: String(data.id ?? data.ID ?? ''),
    parentId: String(data.parentId ?? data.parent_id ?? data.ParentID ?? 0),
    name: data.name ?? data.Name ?? '',
    leaderUserId: String(data.leaderUserId ?? data.leader_user_id ?? data.LeaderUserID ?? 0),
    sort: data.sort ?? data.Sort ?? 0,
    status: (data.status ?? data.Status) === 2 ? 0 : ((data.status ?? data.Status) ?? 1),
    createdAt: data.createdAt ?? data.created_at ?? data.CreatedAt ?? '',
    updatedAt: data.updatedAt ?? data.updated_at ?? data.UpdatedAt ?? '',
    children: (data.children ?? data.Children ?? []).map(normalizeDepartment),
  };
}

function normalizePayload(data: Partial<DepartmentApi.Department>) {
  return {
    parent_id: Number(data.parentId ?? 0),
    name: data.name,
    leader_user_id: Number(data.leaderUserId ?? 0),
    sort: data.sort ?? 0,
    status: data.status === 0 ? 2 : data.status,
  };
}

export async function getDepartmentTreeApi(params: DepartmentApi.ListParams = {}) {
  const data = await requestClient.get<any[]>('/departments', { params });
  return (data ?? []).map(normalizeDepartment);
}

export async function getDepartmentApi(id: string) {
  const data = await requestClient.get<any>(`/departments/${id}`);
  return normalizeDepartment(data);
}

export async function createDepartmentApi(data: Partial<DepartmentApi.Department>) {
  return requestClient.post('/departments', normalizePayload(data));
}

export async function updateDepartmentApi(id: string, data: Partial<DepartmentApi.Department>) {
  return requestClient.post(`/departments/${id}`, normalizePayload(data));
}

export async function deleteDepartmentApi(id: string) {
  return requestClient.post(`/departments/delete/${id}`);
}
