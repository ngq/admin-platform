<!-- 页面功能：角色管理 - 角色列表展示、创建、编辑、删除、菜单/API/按钮权限分配 -->
<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { computed, h, ref } from 'vue';
import {
  message,
  Tag,
  Button,
  Form,
  FormItem,
  Input,
  Select,
  SelectOption,
  Table,
  Modal,
  Space,
  Tree,
  Textarea,
  Spin,
  Checkbox,
  Popover,
} from 'ant-design-vue';

import {
  createRoleApi,
  deleteRoleApi,
  getRoleApi,
  getRoleListApi,
  getRolePermissionConfigApi,
  saveRolePermissionConfigApi,
  updateRoleApi,
} from '#/api/core/role';
import { formatDateTime } from '#/utils/time';

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 0 },
];

const dataScopeOptions = [
  { label: '全部数据', value: 1 },
  { label: '本部门数据', value: 2 },
  { label: '本部门及下级数据', value: 3 },
  { label: '本人数据', value: 4 },
];

const methodColorMap: Record<string, string> = {
  GET: 'green',
  POST: 'blue',
  PUT: 'orange',
  DELETE: 'red',
  PATCH: 'purple',
};

const apiTableColumns = [
  { title: '功能', dataIndex: 'description', key: 'description', width: 240 },
  { title: '方法', dataIndex: 'method', key: 'method', width: 90 },
  { title: '路径', dataIndex: 'path', key: 'path', ellipsis: true },
  { title: '分组', dataIndex: 'group', key: 'group', width: 120 },
];

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);

const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  name: '',
  code: '',
  status: undefined as number | undefined,
});

// 弹框
const modalVisible = ref(false);
const modalTitle = ref('新增角色');
const isEdit = ref(false);
const currentId = ref('');

const formData = ref({
  name: '',
  code: '',
  dataScope: 1,
  description: '',
  status: 1,
});

const permissionModalVisible = ref(false);
const permissionLoading = ref(false);
const permissionTreeData = ref<any[]>([]);
const permissionApiList = ref<any[]>([]);
const permissionButtonList = ref<any[]>([]);
const selectedPermissionTreeKeys = ref<string[]>([]);
const selectedPermissionApiIds = ref<string[]>([]);
const selectedPermissionButtonIds = ref<string[]>([]);
const permissionApiKeyword = ref('');
const permissionButtonKeyword = ref('');

const buttonTableColumns = [
  { title: '按钮', dataIndex: 'name', key: 'name', width: 140 },
  { title: '权限码', dataIndex: 'code', key: 'code', width: 180 },
  { title: '所属菜单', dataIndex: 'menuId', key: 'menuId', width: 160 },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
];

const permissionMenuNameMap = computed(() => {
  const map = new Map<string, string>();
  flattenMenus(permissionTreeData.value).forEach((menu) => {
    map.set(String(menu.id), menu.name);
  });
  return map;
});

const filteredPermissionApiList = computed(() => {
  const keyword = permissionApiKeyword.value.trim().toLowerCase();
  if (!keyword) return permissionApiList.value;
  return permissionApiList.value.filter((api) =>
    [api.description, api.path, api.method, api.group]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword)),
  );
});

const filteredPermissionButtonList = computed(() => {
  const keyword = permissionButtonKeyword.value.trim().toLowerCase();
  if (!keyword) return permissionButtonList.value;
  return permissionButtonList.value.filter((button) =>
    [button.name, button.code, button.description, permissionMenuNameMap.value.get(String(button.menuId))]
      .filter(Boolean)
      .some((value) => String(value).toLowerCase().includes(keyword)),
  );
});

const permissionTreeNodes = computed(() =>
  buildPermissionTreeNodes(permissionTreeData.value, permissionButtonList.value, permissionApiList.value),
);

const permissionSummary = computed(() => {
  const keys = selectedPermissionTreeKeys.value;
  const permissionIds = splitPermissionKeys(keys);
  return {
    menus: keys.filter((key) => key.startsWith('menu:')).length,
    buttons: keys.filter((key) => key.startsWith('button:')).length,
    apis: inferSelectedApiIds(permissionIds.menuIds, permissionIds.buttonIds).length,
  };
});

function toPermissionKeys(menuIds: string[], _apiIds: string[], buttonIds: string[]) {
  return [
    ...menuIds.map((id) => `menu:${id}`),
    ...buttonIds.map((id) => `button:${id}`),
  ];
}

function splitPermissionKeys(keys: string[]) {
  const getIds = (prefix: string) =>
    keys
      .filter((key) => key.startsWith(prefix))
      .map((key) => key.slice(prefix.length))
      .filter(Boolean);

  return {
    menuIds: getIds('menu:'),
    buttonIds: getIds('button:'),
  };
}

function buildPermissionTreeNodes(menus: any[], buttons: any[], _apis: any[]) {
  const buttonMap = new Map<string, any[]>();
  buttons.forEach((button) => {
    const menuId = String(button.menuId);
    const list = buttonMap.get(menuId) || [];
    list.push(button);
    buttonMap.set(menuId, list);
  });

  const walk = (items: any[]): any[] =>
    items.map((menu) => {
      const menuId = String(menu.id);
      const menuButtons = buttonMap.get(menuId) || [];
      const children = menu.children?.length ? walk(menu.children) : [];
      return {
        buttons: menuButtons,
        children,
        id: `menu:${menuId}`,
        key: `menu:${menuId}`,
        name: `${menu.name}${menu.path ? ` (${menu.path})` : ''}`,
        rawId: menuId,
        title: `${menu.name}${menu.path ? ` (${menu.path})` : ''}`,
      };
    });

  return walk(menus);
}

function isButtonChecked(buttonId: number | string) {
  return selectedPermissionTreeKeys.value.includes(`button:${buttonId}`);
}

async function copyPermissionCode(code?: string) {
  if (!code) return;
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(code);
    } else {
      const input = document.createElement('textarea');
      input.value = code;
      input.style.position = 'fixed';
      input.style.opacity = '0';
      document.body.appendChild(input);
      input.select();
      document.execCommand('copy');
      document.body.removeChild(input);
    }
    message.success('权限码已复制');
  } catch {
    message.error('复制失败');
  }
}

function collectNodeButtonKeys(node: any): string[] {
  const data = node?.dataRef || node || {};
  const ownButtonKeys = (data.buttons || []).map((button: any) => `button:${button.id}`);
  const childButtonKeys = (data.children || []).flatMap((child: any) => collectNodeButtonKeys(child));
  return [...ownButtonKeys, ...childButtonKeys];
}

function collectNodeMenuKeys(node: any): string[] {
  const data = node?.dataRef || node || {};
  const ownKey = data.id ? [String(data.id)] : [];
  const childKeys = (data.children || []).flatMap((child: any) => collectNodeMenuKeys(child));
  return [...ownKey, ...childKeys];
}

function findPermissionNodeByKey(nodeKey: string) {
  const walk = (nodes: any[]): any | undefined => {
    for (const node of nodes) {
      if (node.id === nodeKey) return node;
      const matched = walk(node.children || []);
      if (matched) return matched;
    }
    return undefined;
  };
  return walk(permissionTreeNodes.value);
}

function normalizeMenuKey(menuIdOrKey: string | number) {
  const value = String(menuIdOrKey);
  return value.startsWith('menu:') ? value : `menu:${value}`;
}

function findMenuAncestors(menuIdOrKey: string | number) {
  const targetKey = normalizeMenuKey(menuIdOrKey);
  const walk = (nodes: any[], parents: string[]): string[] => {
    for (const node of nodes) {
      if (node.id === targetKey) return [...parents, node.id];
      const matched = walk(node.children || [], [...parents, node.id]);
      if (matched.length) return matched;
    }
    return [];
  };
  return walk(permissionTreeNodes.value, []);
}

function hasSelectedPermissionBelow(node: any, selectedKeys: Set<string>) {
  if (collectNodeButtonKeys(node).some((key) => selectedKeys.has(key))) return true;
  return (node.children || []).some(
    (child: any) => selectedKeys.has(child.id) || hasSelectedPermissionBelow(child, selectedKeys),
  );
}

function pruneEmptyMenuKeys(next: Set<string>, menuIdOrKey: string | number) {
  const ancestors = findMenuAncestors(menuIdOrKey).reverse();
  ancestors.forEach((menuKey) => {
    const node = findPermissionNodeByKey(menuKey);
    if (node && !hasSelectedPermissionBelow(node, next)) {
      next.delete(menuKey);
    }
  });
}

function handlePermissionTreeCheck(checkedKeys: any, info: any) {
  const rawMenuKeys = Array.isArray(checkedKeys) ? checkedKeys : checkedKeys.checked || [];
  const menuKeys = rawMenuKeys.map((key: string | number) => String(key));
  const buttonKeys = selectedPermissionTreeKeys.value.filter((key) => key.startsWith('button:'));
  const changedMenuKeys = collectNodeMenuKeys(info?.node || {});
  const changedButtonKeys = collectNodeButtonKeys(info?.node || {});
  const next = new Set([...menuKeys, ...buttonKeys]);
  changedMenuKeys.forEach((key) => {
    if (info?.checked) {
      next.add(key);
    } else {
      next.delete(key);
    }
  });
  changedButtonKeys.forEach((key) => {
    if (info?.checked) {
      next.add(key);
    } else {
      next.delete(key);
    }
  });
  selectedPermissionTreeKeys.value = [...next];
}

function handleButtonChecked(buttonId: number | string, menuIdOrKey: number | string, checked: boolean) {
  const key = `button:${buttonId}`;
  const next = new Set(selectedPermissionTreeKeys.value);
  if (checked) {
    next.add(key);
    findMenuAncestors(menuIdOrKey).forEach((menuKey) => next.add(menuKey));
  } else {
    next.delete(key);
    pruneEmptyMenuKeys(next, menuIdOrKey);
  }
  selectedPermissionTreeKeys.value = [...next];
}

function inferApiMenuId(api: any, menus: any[]) {
  if (api.permissionMode && api.permissionMode !== 'casbin') return '';
  if (api.menuId && String(api.menuId) !== '0') return String(api.menuId);
  const apiGroup = normalizeToken(api.group);
  const apiPath = normalizePath(api.path);
  let best: { id: string; score: number; depth: number } | undefined;

  menus.forEach((menu) => {
    const menuPath = normalizePath(menu.path);
    const component = normalizePath(menu.component);
    const lastSegment = getLastSegment(menuPath || component);
    const singularLast = singularize(lastSegment);
    let score = 0;

    if (apiPath && menuPath && apiPath.includes(`/admin${menuPath}`)) score += 8;
    if (apiPath && menuPath && apiPath.includes(menuPath)) score += 6;
    if (apiGroup && (apiGroup === lastSegment || apiGroup === singularLast)) score += 5;
    if (apiGroup && component.includes(apiGroup)) score += 4;
    if (apiPath && lastSegment && (apiPath.includes(`/${lastSegment}`) || apiPath.includes(`/${singularLast}`))) score += 3;

    if (score === 0) return;
    const depth = menuPath.split('/').filter(Boolean).length;
    if (!best || score > best.score || (score === best.score && depth > best.depth)) {
      best = { id: String(menu.id), score, depth };
    }
  });

  return best?.id || '';
}

function normalizePath(value: unknown) {
  return String(value || '').toLowerCase().replaceAll('_', '-');
}

function normalizeToken(value: unknown) {
  return normalizePath(value).replaceAll('-', '');
}

function getLastSegment(path: string) {
  return normalizeToken(path.split('/').filter(Boolean).pop() || '');
}

function singularize(value: string) {
  return value.endsWith('s') ? value.slice(0, -1) : value;
}

function inferSelectedApiIds(menuIds: string[], buttonIds: string[]) {
  const selectedMenuIds = new Set(menuIds);
  const selectedButtons = permissionButtonList.value.filter((button) => buttonIds.includes(String(button.id)));
  const flatMenus = flattenMenus(permissionTreeData.value);
  const selectedApiIds = new Set<string>();
  const managedApiIds = new Set<string>();

  permissionApiList.value.forEach((api) => {
    const apiId = String(api.id);
    const menuId = inferApiMenuId(api, flatMenus);
    if (!menuId) return;
    managedApiIds.add(apiId);
    if (selectedButtons.some((button) => isApiDependencyForButton(api, button))) {
      selectedApiIds.add(apiId);
      return;
    }
    if (!selectedMenuIds.has(menuId)) return;

    if (isReadApi(api) || selectedButtons.some((button) => isApiMatchedToButton(api, button))) {
      selectedApiIds.add(apiId);
    }
  });

  selectedPermissionApiIds.value.forEach((apiId) => {
    if (!managedApiIds.has(String(apiId))) {
      selectedApiIds.add(String(apiId));
    }
  });

  return [...selectedApiIds];
}

function isApiDependencyForButton(api: any, button: any) {
  const method = String(api.method || '').toUpperCase();
  const path = normalizePath(api.path);
  const buttonCode = normalizePath(button.code);
  return buttonCode === 'user:roles' && method === 'GET' && path === '/api/v1/roles';
}

function isReadApi(api: any) {
  const method = String(api.method || '').toUpperCase();
  const text = `${api.description || ''} ${api.path || ''}`.toLowerCase();
  return (
    method === 'GET' ||
    text.includes('list') ||
    text.includes('tree') ||
    text.includes('all') ||
    text.includes('detail') ||
    text.includes('current') ||
    text.includes('codes') ||
    text.includes('列表') ||
    text.includes('获取') ||
    text.includes('查看') ||
    text.includes('详情')
  );
}

function isApiMatchedToButton(api: any, button: any) {
  const action = getButtonAction(button);
  const apiAction = normalizeToken(api.actionType);
  if (apiAction && apiAction !== 'other' && action) {
    return apiAction === action || action.includes(apiAction) || apiAction.includes(action);
  }
  const text = normalizeToken(`${api.description || ''} ${api.path || ''} ${api.group || ''}`);
  if (!action) return false;
  if (text.includes(action)) return true;

  const method = String(api.method || '').toUpperCase();
  const path = normalizePath(api.path);
  if (action === 'create') return method === 'POST' && !path.includes(':id') && !path.includes('delete');
  if (action === 'update') return method !== 'GET' && path.includes(':id') && !path.includes('delete');
  if (action === 'delete') return method !== 'GET' && path.includes('delete');
  if (action === 'view') return isReadApi(api);

  const aliases: Record<string, string[]> = {
    permissions: ['permissionconfig', 'permission'],
    password: ['password'],
    roles: ['roles'],
    channel: ['channel', 'channels'],
    test: ['test'],
    generate: ['generate'],
    run: ['run'],
    approve: ['approve'],
    reject: ['reject'],
    takedown: ['takedown'],
    feature: ['feature', 'featured'],
    pin: ['pin'],
    mute: ['mute'],
    grant: ['grant'],
    archive: ['archive'],
    toggle: ['toggle', 'enable', 'disable'],
    config: ['config'],
  };
  return (aliases[action] || []).some((alias) => text.includes(alias));
}

function getButtonAction(button: any) {
  const code = String(button.code || '');
  const parts = code.split(':').filter(Boolean);
  return normalizeToken(parts[parts.length - 1] || '');
}

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getRoleListApi({ ...searchForm.value, page, pageSize });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } catch {
    message.error('获取列表失败');
  } finally {
    loading.value = false;
  }
}

function openAddModal() {
  isEdit.value = false;
  modalTitle.value = '新增角色';
  formData.value = { name: '', code: '', dataScope: 1, description: '', status: 1 };
  modalVisible.value = true;
}

async function openEditModal(id: string) {
  isEdit.value = true;
  modalTitle.value = '编辑角色';
  currentId.value = id;
  try {
    const data = await getRoleApi(id);
    formData.value = {
      name: data.name,
      code: data.code,
      dataScope: data.dataScope,
      description: data.description || '',
      status: data.status,
    };
    modalVisible.value = true;
  } catch {
    message.error('获取详情失败');
  }
}

async function handleSave() {
  try {
    if (isEdit.value) {
      await updateRoleApi(currentId.value, formData.value);
      message.success('更新成功');
    } else {
      await createRoleApi(formData.value);
      message.success('新增成功');
    }
    modalVisible.value = false;
    await fetchList();
  } catch {
    message.error('操作失败');
  }
}

async function handleDelete(id: string) {
  try {
    await deleteRoleApi(id);
    message.success('删除成功');
    await fetchList();
  } catch {
    message.error('删除失败');
  }
}

async function openPermissionModal(id: string) {
  currentId.value = id;
  permissionLoading.value = true;
  permissionApiKeyword.value = '';
  permissionButtonKeyword.value = '';
  try {
    const config = await getRolePermissionConfigApi(id);
    permissionTreeData.value = config.menus || [];
    permissionApiList.value = config.apis || [];
    permissionButtonList.value = config.buttons || [];
    selectedPermissionApiIds.value = config.selectedApiIds || [];
    selectedPermissionTreeKeys.value = toPermissionKeys(
      config.selectedMenuIds || [],
      config.selectedApiIds || [],
      config.selectedButtonIds || [],
    );
    permissionModalVisible.value = true;
  } catch {
    message.error('获取权限配置失败');
  } finally {
    permissionLoading.value = false;
  }
}

async function handleSavePermissionConfig() {
  try {
    const permissionIds = splitPermissionKeys(selectedPermissionTreeKeys.value);
    await saveRolePermissionConfigApi({
      id: currentId.value,
      menuIds: permissionIds.menuIds,
      apiIds: inferSelectedApiIds(permissionIds.menuIds, permissionIds.buttonIds),
      buttonIds: permissionIds.buttonIds,
    });
    message.success('保存成功');
    permissionModalVisible.value = false;
  } catch {
    message.error('保存失败');
  }
}

function handlePermissionApiSelectionChange(keys: (number | string)[]) {
  selectedPermissionApiIds.value = keys.map((key) => String(key));
}

function handlePermissionButtonSelectionChange(keys: (number | string)[]) {
  selectedPermissionButtonIds.value = keys.map((key) => String(key));
}

function flattenMenus(menus: any[]): any[] {
  const result: any[] = [];
  const walk = (items: any[]) => {
    items.forEach((item) => {
      result.push(item);
      if (item.children?.length) {
        walk(item.children);
      }
    });
  };
  walk(menus);
  return result;
}

function handleSearch() {
  pagination.value.page = 1;
  fetchList();
}

function handleReset() {
  searchForm.value = { name: '', code: '', status: undefined };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

const columns = [
  { title: '角色名', dataIndex: 'name', key: 'name' },
  { title: '角色编码', dataIndex: 'code', key: 'code' },
  {
    title: '数据范围',
    dataIndex: 'dataScope',
    key: 'dataScope',
    width: 180,
    customRender: ({ text }: { text: number }) => {
      const item = dataScopeOptions.find((o) => o.value === text);
      return item?.label || '-';
    },
  },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }: { text: number }) =>
      h(Tag, { color: text === 1 ? 'green' : 'red' }, () => (text === 1 ? '正常' : '禁用')),
  },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180, customRender: ({ text }: { text: string }) => formatDateTime(text) },
  { title: '操作', key: 'action', width: 250 },
];

fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="角色名">
        <Input v-model:value="searchForm.name" placeholder="请输入角色名" />
      </FormItem>
      <FormItem label="角色编码">
        <Input v-model:value="searchForm.code" placeholder="请输入角色编码" />
      </FormItem>
      <FormItem label="状态">
        <Select v-model:value="searchForm.status" placeholder="请选择状态" style="width: 120px" allow-clear>
          <SelectOption v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </SelectOption>
        </Select>
      </FormItem>
      <FormItem>
        <Space>
          <Button type="primary" @click="handleSearch">搜索</Button>
          <Button @click="handleReset">重置</Button>
        </Space>
      </FormItem>
    </Form>

    <div class="mb-4">
      <Button type="primary" @click="openAddModal">新增</Button>
    </div>

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="{ total, current: pagination.page, pageSize: pagination.pageSize, showTotal: (t: number) => `共 ${t} 条` }"
      :row-key="(record: any) => record.id"
      @change="(p: any) => handlePageChange(p.current, p.pageSize)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Space>
            <Button size="small" type="link" @click="openPermissionModal(record.id)">权限配置</Button>
            <Button size="small" type="link" @click="openEditModal(record.id)">编辑</Button>
            <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="modalVisible" :title="modalTitle" @ok="handleSave">
      <Form :label-width="80">
        <FormItem label="角色名" required>
          <Input v-model:value="formData.name" placeholder="请输入角色名" />
        </FormItem>
        <FormItem label="角色编码" required>
          <Input v-model:value="formData.code" placeholder="请输入角色编码" />
        </FormItem>
        <FormItem label="数据范围" required>
          <Select v-model:value="formData.dataScope" placeholder="请选择数据范围">
            <SelectOption v-for="opt in dataScopeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
          <div class="form-help">
            用于后台数据范围过滤：本人、本部门、本部门及下级或全部数据。
          </div>
        </FormItem>
        <FormItem label="描述">
          <Textarea v-model:value="formData.description" placeholder="请输入描述" />
        </FormItem>
        <FormItem label="状态" required>
          <Select v-model:value="formData.status" placeholder="请选择状态">
            <SelectOption v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
      </Form>
    </Modal>

    <Modal v-model:open="permissionModalVisible" title="权限配置" :width="920" @ok="handleSavePermissionConfig">
      <Spin :spinning="permissionLoading">
        <div class="permission-config-body">
          <div class="permission-menu-panel">
            <div class="permission-summary">
              已选 {{ permissionSummary.menus }} 个页面 / {{ permissionSummary.buttons }} 个功能，自动同步
              {{ permissionSummary.apis }} 个接口权限
            </div>
            <Tree
              :checked-keys="selectedPermissionTreeKeys"
              :tree-data="permissionTreeNodes"
              checkable
              check-strictly
              default-expand-all
              :field-names="{ title: 'name', key: 'id', children: 'children' }"
              @check="handlePermissionTreeCheck"
            >
              <template #title="{ data }">
                <div class="permission-node-title">
                  <span class="permission-menu-title">{{ data.name }}</span>
                  <div v-if="data.buttons?.length" class="permission-function-row" @click.stop>
                    <Checkbox
                      v-for="button in data.buttons"
                      :key="button.id"
                      :checked="isButtonChecked(button.id)"
                      @change="(event: any) => handleButtonChecked(button.id, data.id, event.target.checked)"
                    >
                      <Popover trigger="hover">
                        <template #content>
                          <span v-if="button.code" class="permission-code-popover">
                            <span>权限码：</span>
                            <Button
                              class="permission-code-copy"
                              size="small"
                              type="link"
                              @click.stop.prevent="copyPermissionCode(button.code)"
                            >
                              {{ button.code }}
                            </Button>
                          </span>
                          <span v-else>未配置权限码</span>
                        </template>
                        <span class="permission-function-label">{{ button.name || button.code }}</span>
                      </Popover>
                    </Checkbox>
                  </div>
                </div>
              </template>
            </Tree>
          </div>
          <div v-if="false" class="permission-resource-panel">
            <Input
              v-model:value="permissionApiKeyword"
              allow-clear
              class="mb-3"
              placeholder="搜索接口功能、路径、方法或分组"
            />
            <Table
              :columns="apiTableColumns"
              :data-source="filteredPermissionApiList"
              :pagination="{ pageSize: 6, showSizeChanger: false }"
              :row-selection="{ selectedRowKeys: selectedPermissionApiIds, onChange: handlePermissionApiSelectionChange }"
              :scroll="{ y: 220 }"
              row-key="id"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'method'">
                  <Tag :color="methodColorMap[record.method] || 'default'">{{ record.method }}</Tag>
                </template>
                <template v-else-if="column.key === 'description'">
                  <div class="font-medium">{{ record.description || record.path }}</div>
                </template>
              </template>
            </Table>

            <Input
              v-model:value="permissionButtonKeyword"
              allow-clear
              class="mb-3 mt-4"
              placeholder="搜索按钮、权限码或所属菜单"
            />
            <Table
              :columns="buttonTableColumns"
              :data-source="filteredPermissionButtonList"
              :pagination="{ pageSize: 6, showSizeChanger: false }"
              :row-selection="{ selectedRowKeys: selectedPermissionButtonIds, onChange: handlePermissionButtonSelectionChange }"
              :scroll="{ y: 180 }"
              row-key="id"
              size="small"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'menuId'">
                  {{ permissionMenuNameMap.get(String(record.menuId)) || record.menuId }}
                </template>
              </template>
            </Table>
          </div>
        </div>
      </Spin>
    </Modal>
  </Page>
</template>

<style scoped>
.permission-modal-body {
  height: 420px;
  overflow: auto;
}

.permission-modal-body--wide {
  height: 500px;
}

.permission-config-body {
  display: block;
  height: 620px;
  min-width: 760px;
  overflow: auto;
}

.permission-menu-panel,
.permission-resource-panel {
  min-height: 0;
  overflow: auto;
}

.permission-menu-panel {
  padding-right: 8px;
}

.permission-resource-panel {
  display: none;
}

.permission-summary {
  margin-bottom: 12px;
  padding: 8px 12px;
  border: 1px solid #f0f0f0;
  border-radius: 6px;
  background: #fafafa;
  color: #595959;
}

.permission-node-title {
  display: inline-flex;
  flex-direction: column;
  min-width: 0;
  vertical-align: top;
}

.permission-menu-title {
  line-height: 24px;
}

.permission-function-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 18px;
  max-width: 100%;
  padding: 4px 0 8px;
}

.permission-function-row :deep(.ant-checkbox-wrapper) {
  align-items: center;
  margin-inline-start: 0;
  white-space: nowrap;
}

.permission-function-label {
  margin-left: 4px;
}

.permission-code-popover {
  display: inline-flex;
  align-items: center;
}

.permission-code-copy {
  height: 20px;
  padding: 0 2px;
  line-height: 20px;
}

.form-help {
  margin-top: 6px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 18px;
}
</style>
