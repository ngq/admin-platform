<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { h, ref } from 'vue';
import {
  Button,
  Form,
  FormItem,
  Input,
  message,
  Modal,
  Select,
  SelectOption,
  Space,
  Table,
  Tag,
  Textarea,
} from 'ant-design-vue';

import {
  batchDeleteApiResourceApi,
  createApiResourceApi,
  deleteApiResourceApi,
  getApiResourceApi,
  getApiResourceListApi,
  updateApiResourceApi,
} from '#/api/core/api-resource';
import { getMenuTreeApi } from '#/api/core/menu';
import { formatDateTime } from '#/utils/time';

const methodColorMap: Record<string, string> = {
  DELETE: 'red',
  GET: 'green',
  PATCH: 'purple',
  POST: 'blue',
  PUT: 'orange',
};

const methodOptions = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].map((value) => ({ label: value, value }));

const actionTypeOptions = [
  { label: '列表/查询', value: 'list' },
  { label: '详情/查看', value: 'view' },
  { label: '新增', value: 'create' },
  { label: '编辑', value: 'update' },
  { label: '删除', value: 'delete' },
  { label: '导入', value: 'import' },
  { label: '导出', value: 'export' },
  { label: '审核', value: 'approve' },
  { label: '测试', value: 'test' },
  { label: '公共读取', value: 'auth' },
  { label: '其他', value: 'other' },
];

const permissionModeOptions = [
  { label: '角色权限（Casbin）', value: 'casbin' },
  { label: '登录即可访问', value: 'auth' },
  { label: '忽略鉴权', value: 'ignored' },
];

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);
const selectedRowKeys = ref<string[]>([]);
const menuOptions = ref<Array<{ label: string; value: string }>>([]);
const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  group: '',
  method: '',
  path: '',
});

const modalVisible = ref(false);
const modalTitle = ref('新增 API');
const isEdit = ref(false);
const currentId = ref('');

const formData = ref({
  actionType: 'other',
  description: '',
  group: '',
  menuId: '',
  method: 'GET',
  path: '',
  permissionMode: 'casbin',
});

function flattenMenuOptions(menus: any[], level = 0): Array<{ label: string; value: string }> {
  return menus.flatMap((menu) => [
    { label: `${'　'.repeat(level)}${menu.name}`, value: String(menu.id) },
    ...flattenMenuOptions(menu.children || [], level + 1),
  ]);
}

async function fetchMenus() {
  try {
    menuOptions.value = flattenMenuOptions(await getMenuTreeApi());
  } catch {
    menuOptions.value = [];
  }
}

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getApiResourceListApi({ ...searchForm.value, page, pageSize });
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
  modalTitle.value = '新增 API';
  formData.value = {
    actionType: 'other',
    description: '',
    group: '',
    menuId: '',
    method: 'GET',
    path: '',
    permissionMode: 'casbin',
  };
  modalVisible.value = true;
}

async function openEditModal(id: string) {
  isEdit.value = true;
  modalTitle.value = '编辑 API';
  currentId.value = id;
  try {
    const data = await getApiResourceApi(id);
    formData.value = {
      actionType: data.actionType || 'other',
      description: data.description || '',
      group: data.group,
      menuId: data.menuId || '',
      method: data.method,
      path: data.path,
      permissionMode: data.permissionMode || 'casbin',
    };
    modalVisible.value = true;
  } catch {
    message.error('获取详情失败');
  }
}

async function handleSave() {
  try {
    if (isEdit.value) {
      await updateApiResourceApi(currentId.value, formData.value);
      message.success('更新成功');
    } else {
      await createApiResourceApi(formData.value);
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
    await deleteApiResourceApi(id);
    message.success('删除成功');
    await fetchList();
  } catch {
    message.error('删除失败');
  }
}

async function handleBatchDelete() {
  if (selectedRowKeys.value.length === 0) {
    message.warning('请选择要删除的项');
    return;
  }
  try {
    await batchDeleteApiResourceApi(selectedRowKeys.value);
    message.success('批量删除成功');
    selectedRowKeys.value = [];
    await fetchList();
  } catch {
    message.error('批量删除失败');
  }
}

function handleSearch() {
  pagination.value.page = 1;
  fetchList();
}

function handleReset() {
  searchForm.value = { group: '', method: '', path: '' };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

function handleRowSelectionChange(keys: Array<number | string>) {
  selectedRowKeys.value = keys.map(String);
}

function getMenuName(menuId?: string) {
  return menuOptions.value.find((item) => item.value === String(menuId || ''))?.label.trim() || '未绑定';
}

function getActionTypeLabel(value?: string) {
  return actionTypeOptions.find((item) => item.value === value)?.label || value || '-';
}

function getPermissionModeLabel(value?: string) {
  return permissionModeOptions.find((item) => item.value === value)?.label || value || '-';
}

const columns = [
  { title: 'API 路径', dataIndex: 'path', key: 'path', ellipsis: true },
  {
    title: '方法',
    dataIndex: 'method',
    key: 'method',
    width: 80,
    customRender: ({ text }: { text: string }) =>
      h(Tag, { color: methodColorMap[text] || 'default' }, () => text),
  },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  { title: '分组', dataIndex: 'group', key: 'group', width: 120 },
  {
    title: '所属菜单',
    dataIndex: 'menuId',
    key: 'menuId',
    width: 160,
    customRender: ({ text }: { text: string }) => getMenuName(text),
  },
  {
    title: '功能',
    dataIndex: 'actionType',
    key: 'actionType',
    width: 110,
    customRender: ({ text }: { text: string }) => getActionTypeLabel(text),
  },
  {
    title: '权限模式',
    dataIndex: 'permissionMode',
    key: 'permissionMode',
    width: 170,
    customRender: ({ text }: { text: string }) =>
      h('span', { class: 'whitespace-nowrap' }, getPermissionModeLabel(text)),
  },
  {
    title: '创建时间',
    dataIndex: 'createdAt',
    key: 'createdAt',
    width: 180,
    customRender: ({ text }: { text: string }) => formatDateTime(text),
  },
  { title: '操作', key: 'action', width: 150 },
];

fetchMenus();
fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="API 路径">
        <Input v-model:value="searchForm.path" placeholder="请输入 API 路径" />
      </FormItem>
      <FormItem label="分组">
        <Input v-model:value="searchForm.group" placeholder="请输入分组" />
      </FormItem>
      <FormItem label="方法">
        <Select v-model:value="searchForm.method" allow-clear placeholder="请选择方法" style="width: 120px">
          <SelectOption v-for="opt in methodOptions" :key="opt.value" :value="opt.value">
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

    <div class="mb-4 flex gap-2">
      <Button type="primary" @click="openAddModal">新增</Button>
      <Button danger :disabled="selectedRowKeys.length === 0" @click="handleBatchDelete">
        批量删除
      </Button>
    </div>

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="{ total, current: pagination.page, pageSize: pagination.pageSize, showTotal: (t: number) => `共 ${t} 条` }"
      :row-key="(record: any) => record.id"
      :row-selection="{ selectedRowKeys, onChange: handleRowSelectionChange }"
      :scroll="{ x: 1300 }"
      @change="(p: any) => handlePageChange(p.current, p.pageSize)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Space>
            <Button size="small" type="link" @click="openEditModal(record.id)">编辑</Button>
            <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="modalVisible" :title="modalTitle" :width="680" @ok="handleSave">
      <Form :label-col="{ style: { width: '96px' } }">
        <FormItem label="API 路径" required>
          <Input v-model:value="formData.path" placeholder="请输入 API 路径" />
        </FormItem>
        <FormItem label="请求方法" required>
          <Select v-model:value="formData.method" placeholder="请选择方法">
            <SelectOption v-for="opt in methodOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="分组" required>
          <Input v-model:value="formData.group" placeholder="请输入分组" />
        </FormItem>
        <FormItem label="所属菜单">
          <Select
            v-model:value="formData.menuId"
            allow-clear
            show-search
            option-filter-prop="label"
            placeholder="公共接口可不选菜单"
          >
            <SelectOption v-for="opt in menuOptions" :key="opt.value" :label="opt.label" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="功能类型">
          <Select v-model:value="formData.actionType" placeholder="请选择功能类型">
            <SelectOption v-for="opt in actionTypeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="权限模式">
          <Select v-model:value="formData.permissionMode" placeholder="请选择权限模式">
            <SelectOption v-for="opt in permissionModeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="描述">
          <Textarea v-model:value="formData.description" placeholder="请输入描述" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
