import type { UserInfo } from '@vben/types';

import { requestClient } from '#/api/request';

/**
 * 用户管理 API
 */

/**
 * 后端用户信息响应格式
 */
interface BackendUserInfo {
  id: number;
  username: string;
  nickname: string;
  avatar: string;
  roles: Array<{ code: string }>;
}

function normalizeRole(role: any) {
  const id = String(role.id ?? '');
  const name = role.name ?? '';
  return {
    ...role,
    id,
    label: name || role.code || id,
    name,
    value: id,
    createdAt: role.createdAt ?? role.created_at ?? '',
    updatedAt: role.updatedAt ?? role.updated_at ?? '',
  };
}

function normalizeUser(data: any): UserApi.User {
  return {
    ...data,
    id: String(data.id ?? ''),
    deptId: String(data.deptId ?? data.dept_id ?? data.DeptID ?? 0),
    deptName: data.deptName ?? data.dept_name ?? data.DeptName ?? '',
    createdAt: data.createdAt ?? data.created_at ?? '',
    updatedAt: data.updatedAt ?? data.updated_at ?? '',
    status: data.status === 2 ? 0 : (data.status ?? 1),
    roles: (data.roles ?? []).map(normalizeRole),
  };
}

/**
 * 获取用户信息
 * 适配后端 /auth/current 接口
 */
export async function getUserInfoApi(): Promise<UserInfo> {
  const data = await requestClient.get<BackendUserInfo>('/auth/current');

  // 转换后端响应格式为前端 UserInfo 格式
  return {
    userId: String(data.id),
    username: data.username,
    realName: data.nickname,
    avatar: data.avatar,
    roles: data.roles?.map((r: { code: string }) => r.code) || [],
    desc: '',
    homePath: '/dashboard/analytics',
    token: '',
  };
}

export namespace UserApi {
  export interface User {
    id: string;
    username: string;
    nickname: string;
    deptId: string;
    deptName?: string;
    password?: string;
    status: number;
    roles?: Role[];
    roleIds?: string[];
    createdAt: string;
    updatedAt?: string;
  }

  export interface Role {
    id: string;
    name: string;
    code: string;
  }

  export interface ListParams {
    username?: string;
    status?: number;
    deptId?: string;
    page?: number;
    pageSize?: number;
  }

  export interface AssignRolesParams {
    id: string;
    roleIds: string[];
  }

  export interface ResetPasswordParams {
    id: string;
    password: string;
  }
}

/** 获取用户列表 */
export async function getUserListApi(params: UserApi.ListParams) {
  const data = await requestClient.get<{ items?: UserApi.User[]; list?: UserApi.User[]; total: number }>('/users', {
    params: {
      ...params,
      dept_id: params.deptId && params.deptId !== '0' ? Number(params.deptId) : undefined,
    },
  });
  return {
    list: (data.list ?? data.items ?? []).map(normalizeUser),
    total: data.total ?? 0,
  };
}

/** 获取用户详情 */
export async function getUserApi(id: string) {
  const data = await requestClient.get<UserApi.User>(`/users/${id}`);
  return normalizeUser(data);
}

/** 创建用户 */
export async function createUserApi(data: Partial<UserApi.User>) {
  return requestClient.post<UserApi.User>('/users', {
    username: data.username,
    nickname: data.nickname,
    password: data.password,
    dept_id: Number(data.deptId ?? 0),
    roleIds: (data.roleIds ?? []).map((id) => Number(id)),
    status: data.status === 0 ? 2 : data.status,
  });
}

/** 更新用户 */
export async function updateUserApi(id: string, data: Partial<UserApi.User>) {
  return requestClient.post<UserApi.User>(`/users/${id}`, {
    nickname: data.nickname,
    dept_id: Number(data.deptId ?? 0),
    status: data.status === 0 ? 2 : data.status,
  });
}

/** 删除用户 */
export async function deleteUserApi(id: string) {
  return requestClient.post(`/users/delete/${id}`);
}

/** 分配角色给用户 */
export async function assignUserRolesApi(params: UserApi.AssignRolesParams) {
  return requestClient.post(`/users/roles?id=${params.id}`, { roleIds: params.roleIds });
}

/** 重置用户密码（管理员操作） */
export async function resetUserPasswordApi(params: UserApi.ResetPasswordParams) {
  return requestClient.post(`/users/password?id=${params.id}`, { newPassword: params.password });
}

/** 获取角色选项列表（用于下拉选择） */
export async function getRoleOptionsApi() {
  const data = await requestClient.get<{ items?: UserApi.Role[]; list?: UserApi.Role[] }>('/roles', {
    params: { page: 1, pageSize: 100 },
  });
  return {
    list: (data.list ?? data.items ?? []).map(normalizeRole),
  };
}
