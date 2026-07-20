import type { RouteRecordStringComponent } from '@vben/types';

import { requestClient } from '#/api/request';

/**
 * 后端菜单响应格式
 */
export interface BackendMenu {
  id: number | string;
  parent_id?: number;
  parentId?: number;
  name: string;
  path: string;
  component?: string;
  redirect?: string;
  icon?: string;
  sort?: number;
  type?: number;
  visible?: number;
  status?: number;
  children?: BackendMenu[];
}

export interface MenuFormData {
  parentId?: number | string;
  parent_id?: number | string;
  name: string;
  path?: string;
  component?: string;
  redirect?: string;
  icon?: string;
  sort?: number;
  type?: number;
  visible?: number;
  status?: number;
}

export interface ButtonFormData {
  id?: number | string;
  menuId?: number | string;
  menu_id?: number | string;
  name: string;
  code: string;
  description?: string;
}

function normalizeMenu(menu: BackendMenu): BackendMenu {
  return {
    ...menu,
    id: String(menu.id),
    parentId: menu.parentId ?? menu.parent_id ?? 0,
    component: menu.component ?? '',
    redirect: menu.redirect ?? '',
    icon: menu.icon ?? '',
    sort: menu.sort ?? 0,
    type: menu.type ?? 0,
    visible: menu.visible ?? 1,
    status: menu.status === 2 ? 0 : (menu.status ?? 1),
    children: (menu.children ?? []).map(normalizeMenu),
  };
}

const supportedMenuComponents = new Set([
  'dashboard/analytics/index',
  'system/user/index',
  'system/dept/index',
  'system/role/index',
  'system/menu/index',
  'system/api/index',
  'system/param/index',
  'logs/operation/index',
  'logs/login/index',
  'logs/error/index',
]);

function hasSupportedComponent(component?: string) {
  if (!component) return false;
  const normalized = component.replace(/^\/+/, '').replace(/\.vue$/, '');
  return supportedMenuComponents.has(normalized);
}

function filterRenderableMenus(menus: BackendMenu[]): BackendMenu[] {
  return menus
    .map((menu) => {
      const children = filterRenderableMenus(menu.children ?? []);
      return { ...menu, children };
    })
    .filter((menu) => {
      const hasChildren = (menu.children ?? []).length > 0;
      return hasChildren || hasSupportedComponent(menu.component);
    });
}

/**
 * 获取当前用户所有菜单
 * 适配后端 /menu/all 接口
 */
export async function getAllMenusApi(): Promise<RouteRecordStringComponent[]> {
  const menus = await requestClient.get<BackendMenu[]>('/menu/all');
  return transformMenus(filterRenderableMenus((menus ?? []).map(normalizeMenu)));
}

/**
 * 获取原始菜单树数据，供系统管理页和权限分配页使用
 */
export async function getMenuTreeApi(): Promise<BackendMenu[]> {
  const menus = await requestClient.get<BackendMenu[]>('/menus/tree');
  return (menus ?? []).map(normalizeMenu);
}

export async function getMenuApi(id: string) {
  const menu = await requestClient.get<BackendMenu>(`/menus/${id}`);
  return normalizeMenu(menu);
}

function normalizeMenuPayload(data: MenuFormData) {
  return {
    parent_id: Number(data.parentId ?? data.parent_id ?? 0),
    name: data.name,
    path: data.path ?? '',
    component: data.component ?? '',
    redirect: data.redirect ?? '',
    icon: data.icon ?? '',
    sort: Number(data.sort ?? 0),
    type: Number(data.type ?? 2),
    visible: Number(data.visible ?? 1),
    status: Number(data.status ?? 1) === 0 ? 2 : Number(data.status ?? 1),
  };
}

export async function createMenuApi(data: MenuFormData) {
  return requestClient.post('/menus', normalizeMenuPayload(data));
}

export async function updateMenuApi(id: string, data: MenuFormData) {
  return requestClient.post(`/menus/${id}`, normalizeMenuPayload(data));
}

export async function deleteMenuApi(id: string) {
  return requestClient.post(`/menus/delete/${id}`);
}

export async function getMenuButtonsApi(menuId: number | string): Promise<ButtonFormData[]> {
  const buttons = await requestClient.get<any[]>('/menus/buttons', {
    params: { menuId },
  });
  return (buttons ?? []).map((button) => ({
    id: String(button.id ?? button.ID ?? ''),
    menuId: String(button.menuId ?? button.menu_id ?? button.MenuID ?? ''),
    name: button.name ?? button.Name ?? '',
    code: button.code ?? button.Code ?? '',
    description: button.description ?? button.Description ?? '',
  }));
}

function normalizeButtonPayload(data: ButtonFormData) {
  return {
    menu_id: Number(data.menuId ?? data.menu_id ?? 0),
    name: data.name,
    code: data.code,
    description: data.description ?? '',
  };
}

export async function createButtonApi(data: ButtonFormData) {
  return requestClient.post('/buttons', normalizeButtonPayload(data));
}

export async function updateButtonApi(id: string, data: ButtonFormData) {
  return requestClient.post(`/buttons/${id}`, normalizeButtonPayload(data));
}

export async function deleteButtonApi(id: string) {
  return requestClient.post(`/buttons/delete/${id}`);
}

/**
 * 将后端菜单格式转换为前端路由格式
 */
function transformMenus(menus: BackendMenu[]): RouteRecordStringComponent[] {
  return menus.map((menu) => {
    const route: RouteRecordStringComponent = {
      name: menu.name,
      path: menu.path,
      component: menu.component ?? '',
      meta: {
        icon: menu.icon || undefined,
        order: menu.sort,
        hideInMenu: menu.visible !== 1,
        title: menu.name,
      },
      children: menu.children ? transformMenus(menu.children) : undefined,
    };

    if (menu.redirect) {
      route.redirect = menu.redirect;
    }

    return route;
  });
}
