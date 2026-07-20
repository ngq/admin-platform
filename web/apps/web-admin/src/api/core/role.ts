import { requestClient } from '#/api/request';

/**
 * 角色管理 API
 */

export namespace RoleApi {
  export interface Role {
    id: string;
    name: string;
    code: string;
    dataScope: number;
    description?: string;
    status: number;
    createdAt: string;
    updatedAt?: string;
  }

  export interface ListParams {
    name?: string;
    code?: string;
    status?: number;
    page?: number;
    pageSize?: number;
  }

  export interface AssignMenusParams {
    id: string;
    menuIds: string[];
  }

  export interface AssignApisParams {
    id: string;
    apiIds: string[];
  }

  export interface AssignButtonsParams {
    id: string;
    menuId: string;
    buttonIds: string[];
  }

  export interface PermissionMenu {
    id: string;
    parentId: number;
    name: string;
    path: string;
    component?: string;
    children?: PermissionMenu[];
  }

  export interface PermissionApi {
    id: string;
    path: string;
    method: string;
    description?: string;
    group?: string;
    menuId?: string;
    actionType?: string;
    permissionMode?: string;
  }

  export interface PermissionButton {
    id: string;
    menuId: string;
    name: string;
    code: string;
    description?: string;
  }

  export interface PermissionConfig {
    menus: PermissionMenu[];
    apis: PermissionApi[];
    buttons: PermissionButton[];
    selectedMenuIds: string[];
    selectedApiIds: string[];
    selectedButtonIds: string[];
  }

  export interface SavePermissionConfigParams {
    id: string;
    menuIds: string[];
    apiIds: string[];
    buttonIds: string[];
  }
}

function normalizeRole(data: any): RoleApi.Role {
  return {
    ...data,
    id: String(data.id ?? data.ID ?? ''),
    name: data.name ?? data.Name ?? '',
    code: data.code ?? data.Code ?? '',
    dataScope: data.dataScope ?? data.data_scope ?? data.DataScope ?? 1,
    description: data.description ?? data.Description ?? '',
    createdAt: data.createdAt ?? data.created_at ?? data.CreatedAt ?? '',
    updatedAt: data.updatedAt ?? data.updated_at ?? data.UpdatedAt ?? '',
    status: (data.status ?? data.Status) === 2 ? 0 : ((data.status ?? data.Status) ?? 1),
  };
}

function normalizePagedList<T>(data: { items?: T[]; list?: T[]; total?: number }) {
  return {
    list: data.list ?? data.items ?? [],
    total: data.total ?? 0,
  };
}

function normalizePermissionMenu(menu: any): RoleApi.PermissionMenu {
  return {
    ...menu,
    id: String(menu.id ?? menu.ID ?? ''),
    parentId: menu.parentId ?? menu.parent_id ?? menu.ParentID ?? 0,
    name: menu.name ?? menu.Name ?? '',
    path: menu.path ?? menu.Path ?? '',
    component: menu.component ?? menu.Component ?? '',
    children: (menu.children ?? []).map(normalizePermissionMenu),
  };
}

function normalizePermissionApi(api: any): RoleApi.PermissionApi {
  return {
    ...api,
    id: String(api.id ?? api.ID ?? ''),
    path: api.path ?? api.Path ?? '',
    method: api.method ?? api.Method ?? '',
    description: api.description ?? api.Description ?? '',
    group: api.group ?? api.Group ?? '',
    menuId: String(api.menuId ?? api.menu_id ?? api.MenuID ?? ''),
    actionType: api.actionType ?? api.action_type ?? api.ActionType ?? 'other',
    permissionMode: api.permissionMode ?? api.permission_mode ?? api.PermissionMode ?? (api.ignored === 1 || api.Ignored === 1 ? 'ignored' : 'casbin'),
  };
}

function normalizePermissionButton(button: any): RoleApi.PermissionButton {
  return {
    ...button,
    id: String(button.id ?? button.ID ?? ''),
    menuId: String(button.menuId ?? button.menu_id ?? button.MenuID ?? ''),
    name: button.name ?? button.Name ?? '',
    code: button.code ?? button.Code ?? '',
    description: button.description ?? button.Description ?? '',
  };
}

/** 获取角色列表 */
export async function getRoleListApi(params: RoleApi.ListParams) {
  const data = await requestClient.get<{ items?: RoleApi.Role[]; list?: RoleApi.Role[]; total?: number }>('/roles', {
    params,
  });
  const normalized = normalizePagedList(data);
  return {
    list: normalized.list.map(normalizeRole),
    total: normalized.total,
  };
}

/** 获取角色详情 */
export async function getRoleApi(id: string) {
  const data = await requestClient.get<RoleApi.Role>(`/roles/${id}`);
  return normalizeRole(data);
}

/** 创建角色 */
export async function createRoleApi(data: Partial<RoleApi.Role>) {
  return requestClient.post('/roles', {
    name: data.name,
    code: data.code,
    data_scope: data.dataScope,
    description: data.description,
    status: data.status === 0 ? 2 : data.status,
  });
}

/** 更新角色 */
export async function updateRoleApi(id: string, data: Partial<RoleApi.Role>) {
  return requestClient.post(`/roles/${id}`, {
    name: data.name,
    code: data.code,
    data_scope: data.dataScope,
    description: data.description,
    status: data.status === 0 ? 2 : data.status,
  });
}

/** 删除角色 */
export async function deleteRoleApi(id: string) {
  return requestClient.post(`/roles/delete/${id}`);
}

/** 获取所有角色（用于下拉选项） */
export async function getAllRolesApi() {
  const data = await requestClient.get<{ items?: RoleApi.Role[]; list?: RoleApi.Role[]; total?: number }>('/roles', {
    params: { page: 1, pageSize: 500 },
  });
  return {
    list: (data.list ?? data.items ?? []).map(normalizeRole),
  };
}

export async function getRolePermissionConfigApi(id: string): Promise<RoleApi.PermissionConfig> {
  const data = await requestClient.get<any>(`/roles/permission-config?id=${id}`);
  return {
    menus: (data.menus ?? []).map(normalizePermissionMenu),
    apis: (data.apis ?? data.APIs ?? []).map(normalizePermissionApi),
    buttons: (data.buttons ?? []).map(normalizePermissionButton),
    selectedMenuIds: (data.selectedMenuIds ?? data.selected_menu_ids ?? []).map((item: number | string) => String(item)),
    selectedApiIds: (data.selectedApiIds ?? data.selected_api_ids ?? []).map((item: number | string) => String(item)),
    selectedButtonIds: (data.selectedButtonIds ?? data.selected_button_ids ?? []).map((item: number | string) => String(item)),
  };
}

export async function saveRolePermissionConfigApi(params: RoleApi.SavePermissionConfigParams) {
  return requestClient.post(`/roles/permission-config?id=${params.id}`, {
    menu_ids: params.menuIds.map((id) => Number(id)),
    api_ids: params.apiIds.map((id) => Number(id)),
    button_ids: params.buttonIds.map((id) => Number(id)),
  });
}

/** 分配菜单权限给角色 */
export async function assignRoleMenusApi(params: RoleApi.AssignMenusParams) {
  return requestClient.post(`/roles/menus?id=${params.id}`, {
    menu_ids: params.menuIds,
  });
}

/** 获取角色已分配的菜单 ID 列表 */
export async function getRoleMenusApi(id: string) {
  const data = await requestClient.get<{ menu_ids?: string[]; menuIds?: string[] }>(`/roles/menus?id=${id}`);
  return {
    menuIds: data.menuIds ?? data.menu_ids ?? [],
  };
}

/** 分配 API 权限给角色 */
export async function assignRoleApisApi(params: RoleApi.AssignApisParams) {
  return requestClient.post(`/roles/apis?id=${params.id}`, {
    api_ids: params.apiIds.map((id) => Number(id)),
  });
}

/** 获取角色已分配的 API ID 列表 */
export async function getRoleApisApi(id: string) {
  const data = await requestClient.get<{ api_ids?: Array<number | string>; apiIds?: Array<number | string> }>(`/roles/apis?id=${id}`);
  return {
    apiIds: (data.apiIds ?? data.api_ids ?? []).map((apiId) => String(apiId)),
  };
}

/** 分配按钮权限给角色 */
export async function assignRoleButtonsApi(params: RoleApi.AssignButtonsParams) {
  return requestClient.post(`/roles/buttons?id=${params.id}`, {
    button_ids: params.buttonIds,
  });
}

/** 获取角色已分配的按钮 ID 列表 */
export async function getRoleButtonsApi(id: string, _menuId: string) {
  const data = await requestClient.get<{ button_ids?: string[]; buttonIds?: string[] }>(`/roles/buttons?id=${id}`);
  return {
    buttonIds: data.buttonIds ?? data.button_ids ?? [],
  };
}
