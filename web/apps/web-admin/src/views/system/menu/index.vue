<!-- 页面功能：菜单管理 - 菜单树展示、创建、编辑、删除 -->
<script lang="ts" setup>
import { IconPicker, Page } from '@vben/common-ui';

import { computed, h, onMounted, ref } from 'vue';
import {
  Button,
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  Modal,
  Select,
  SelectOption,
  Space,
  Table,
  Tag,
} from 'ant-design-vue';

import {
  createButtonApi,
  createMenuApi,
  deleteMenuApi,
  getMenuApi,
  getMenuTreeApi,
  updateButtonApi,
  updateMenuApi,
  type BackendMenu,
} from '#/api/core/menu';
import { useAccessCodes } from '#/composables/useAccessCodes';

const { hasAccessCode, hasAnyAccessCode } = useAccessCodes();
const iconPickerInput = h(Input);

const loading = ref(false);
const tableData = ref<BackendMenu[]>([]);

const modalVisible = ref(false);
const modalTitle = ref('新增菜单');
const isEdit = ref(false);
const editingButton = ref(false);
const currentId = ref('');

const formData = ref({
  component: '',
  description: '',
  icon: '',
  name: '',
  parentId: '0',
  path: '',
  permissionCode: '',
  redirect: '',
  sort: 0,
  status: 1,
  type: 2,
  visible: 1,
});

const typeOptions = [
  { label: '目录', value: 1 },
  { label: '页面菜单', value: 2 },
  { label: '功能按钮', value: 3 },
];

const visibleOptions = [
  { label: '显示', value: 1 },
  { label: '隐藏', value: 0 },
];

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 0 },
];

const isDirectoryType = computed(() => Number(formData.value.type) === 1);
const isPageMenuType = computed(() => Number(formData.value.type) === 2);
const isButtonType = computed(() => Number(formData.value.type) === 3);

const columns = computed(() => {
  const baseColumns: any[] = [
    {
      dataIndex: 'name',
      key: 'name',
      title: '菜单名',
      width: 200,
    },
    {
      customRender: ({ text }: { text: string }) => (text ? h(Tag, {}, () => text) : '-'),
      dataIndex: 'icon',
      key: 'icon',
      title: '图标',
      width: 80,
    },
    {
      dataIndex: 'path',
      key: 'path',
      title: '路径',
    },
    {
      dataIndex: 'component',
      ellipsis: true,
      key: 'component',
      title: '组件',
    },
    {
      dataIndex: 'sort',
      key: 'sort',
      title: '排序',
      width: 60,
    },
    {
      customRender: ({ text }: { text: number }) =>
        h(Tag, { color: text === 1 ? 'green' : 'orange' }, () => (text === 1 ? '显示' : '隐藏')),
      dataIndex: 'visible',
      key: 'visible',
      title: '显示',
      width: 80,
    },
    {
      customRender: ({ text }: { text: number }) =>
        h(Tag, { color: text === 1 ? 'green' : 'red' }, () => (text === 1 ? '正常' : '禁用')),
      dataIndex: 'status',
      key: 'status',
      title: '状态',
      width: 80,
    },
  ];

  if (hasAnyAccessCode(['menu:create', 'menu:update', 'menu:delete'])) {
    baseColumns.push({
      fixed: 'right',
      key: 'action',
      title: '操作',
      width: 210,
    });
  }

  return baseColumns;
});

const parentMenuOptions = computed(() => {
  const disabledIds = isEdit.value && currentId.value ? collectMenuIds(findMenuById(tableData.value, currentId.value)) : [];
  return [
    { disabled: false, label: '顶级菜单', value: '0' },
    ...flattenMenuOptions(tableData.value, disabledIds),
  ];
});

const pageMenuOptions = computed(() =>
  flattenMenuOptions(tableData.value, [], 0).filter((option) => {
    const menu = findMenuById(tableData.value, option.value);
    return Number(menu?.type ?? 2) === 2;
  }),
);

async function fetchMenuTree() {
  loading.value = true;
  try {
    tableData.value = await getMenuTreeApi();
  } catch {
    message.error('获取菜单失败');
  } finally {
    loading.value = false;
  }
}

function flattenMenuOptions(
  menus: BackendMenu[],
  disabledIds: string[],
  level = 0,
): Array<{ disabled: boolean; label: string; value: string }> {
  return menus.flatMap((menu) => [
    {
      disabled: disabledIds.includes(String(menu.id)),
      label: `${'　'.repeat(level)}${menu.name}`,
      value: String(menu.id),
    },
    ...flattenMenuOptions(menu.children || [], disabledIds, level + 1),
  ]);
}

function collectMenuIds(menu?: BackendMenu): string[] {
  if (!menu) return [];
  return [String(menu.id), ...(menu.children || []).flatMap(collectMenuIds)];
}

function findMenuById(menus: BackendMenu[], id: string): BackendMenu | undefined {
  for (const menu of menus) {
    if (String(menu.id) === id) return menu;
    const child = findMenuById(menu.children || [], id);
    if (child) return child;
  }
  return undefined;
}

function resetForm(parentId = '0') {
  formData.value = {
    component: '',
    description: '',
    icon: '',
    name: '',
    parentId,
    path: '',
    permissionCode: '',
    redirect: '',
    sort: 0,
    status: 1,
    type: parentId === '0' ? 1 : 2,
    visible: 1,
  };
}

function openAddModal(parent?: any) {
  isEdit.value = false;
  editingButton.value = false;
  currentId.value = '';
  modalTitle.value = parent ? `新增子菜单 - ${parent.name}` : '新增菜单';
  resetForm(parent ? String(parent.id) : '0');
  modalVisible.value = true;
}

function openAddButtonModal(parent: any) {
  isEdit.value = false;
  editingButton.value = true;
  currentId.value = '';
  modalTitle.value = `新增功能按钮 - ${parent.name}`;
  resetForm(String(parent.id));
  formData.value.type = 3;
  formData.value.icon = '';
  formData.value.path = '';
  formData.value.component = '';
  formData.value.redirect = '';
  modalVisible.value = true;
}

async function openEditModal(record: any) {
  isEdit.value = true;
  editingButton.value = false;
  currentId.value = String(record.id);
  modalTitle.value = '编辑菜单';
  try {
    const data = await getMenuApi(String(record.id));
    formData.value = {
      component: data.component || '',
      description: '',
      icon: data.icon || '',
      name: data.name || '',
      parentId: String(data.parentId ?? data.parent_id ?? 0),
      path: data.path || '',
      permissionCode: '',
      redirect: data.redirect || '',
      sort: data.sort || 0,
      status: data.status || 1,
      type: data.type || 2,
      visible: data.visible ?? 1,
    };
    modalVisible.value = true;
  } catch {
    message.error('获取菜单详情失败');
  }
}

async function openEditButtonModal(menu: any, button: any) {
  isEdit.value = true;
  editingButton.value = true;
  currentId.value = String(button.id);
  modalTitle.value = '编辑功能按钮';
  formData.value = {
    component: '',
    description: button.description || '',
    icon: '',
    name: button.name || '',
    parentId: String(button.menuId ?? button.menu_id ?? menu.id),
    path: '',
    permissionCode: button.code || '',
    redirect: '',
    sort: 0,
    status: 1,
    type: 3,
    visible: 1,
  };
  modalVisible.value = true;
}

void openEditButtonModal;

async function saveCurrentForm() {
  if (isButtonType.value) {
    const payload = {
      code: formData.value.permissionCode,
      description: formData.value.description,
      menuId: formData.value.parentId,
      name: formData.value.name,
    };
    if (isEdit.value && editingButton.value) {
      await updateButtonApi(currentId.value, payload);
    } else {
      await createButtonApi(payload);
    }
    return;
  }

  const payload = {
    ...formData.value,
    component: isDirectoryType.value ? '' : formData.value.component,
  };
  if (isEdit.value) {
    await updateMenuApi(currentId.value, payload);
  } else {
    await createMenuApi(payload);
  }
}

async function handleSave() {
  if (!formData.value.name.trim()) {
    message.warning('请输入菜单名');
    return;
  }
  if (isPageMenuType.value && (!formData.value.path.trim() || !formData.value.component.trim())) {
    message.warning('页面菜单需要填写前端路径和组件');
    return;
  }
  if (isButtonType.value && (!String(formData.value.parentId || '').trim() || formData.value.parentId === '0')) {
    message.warning('功能按钮需要选择所属页面');
    return;
  }
  if (isButtonType.value && !formData.value.permissionCode.trim()) {
    message.warning('功能按钮需要填写权限码');
    return;
  }
  try {
    await saveCurrentForm();
    message.success(isEdit.value ? '更新成功' : '新增成功');
    modalVisible.value = false;
    await fetchMenuTree();
  } catch {
    message.error('保存失败');
  }
}

function handleDelete(record: any) {
  Modal.confirm({
    content: `删除「${record.name}」会同时删除其子菜单，确定继续？`,
    okText: '删除',
    okType: 'danger',
    onOk: async () => {
      try {
        await deleteMenuApi(String(record.id));
        message.success('删除成功');
        await fetchMenuTree();
      } catch {
        message.error('删除失败');
      }
    },
    title: '确认删除菜单',
  });
}

onMounted(fetchMenuTree);
</script>

<template>
  <Page>
    <div class="mb-4">
      <Button v-if="hasAccessCode('menu:create')" type="primary" @click="openAddModal()">新增菜单</Button>
    </div>

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="false"
      :row-key="(record: any) => record.id"
      childrenColumnName="children"
      default-expand-all-rows
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Space>
            <Button v-if="hasAccessCode('menu:create')" size="small" type="link" @click="openAddModal(record)">
              新增子菜单
            </Button>
            <Button v-if="hasAccessCode('button:create')" size="small" type="link" @click="openAddButtonModal(record)">
              新增按钮
            </Button>
            <Button v-if="hasAccessCode('menu:update')" size="small" type="link" @click="openEditModal(record)">
              编辑
            </Button>
            <Button v-if="hasAccessCode('menu:delete')" danger size="small" type="link" @click="handleDelete(record)">
              删除
            </Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="modalVisible" :title="modalTitle" :width="720" @ok="handleSave">
      <Form :label-col="{ span: 5 }" :model="formData" :wrapper-col="{ span: 17 }">
        <FormItem :label="isButtonType ? '所属页面' : '上级菜单'">
          <Select v-if="!isButtonType" v-model:value="formData.parentId" show-search option-filter-prop="label">
            <SelectOption
              v-for="opt in parentMenuOptions"
              :key="opt.value"
              :disabled="opt.disabled"
              :label="opt.label"
              :value="opt.value"
            >
              {{ opt.label }}
            </SelectOption>
          </Select>
          <Select v-else v-model:value="formData.parentId" show-search option-filter-prop="label" :options="pageMenuOptions" />
        </FormItem>
        <FormItem :label="isButtonType ? '按钮名称' : '菜单名'" required>
          <Input v-model:value="formData.name" :placeholder="isButtonType ? '例如：新增用户' : '请输入菜单名'" />
        </FormItem>
        <FormItem label="类型" required>
          <Select v-model:value="formData.type" :options="typeOptions" />
        </FormItem>
        <FormItem v-if="!isButtonType" :label="isDirectoryType ? '目录路径' : '前端路径'" :required="isPageMenuType">
          <Input v-model:value="formData.path" :placeholder="isDirectoryType ? '/system' : '/system/menu'" />
        </FormItem>
        <FormItem v-if="isPageMenuType" label="页面组件" required>
          <Input v-model:value="formData.component" placeholder="system/menu/index" />
        </FormItem>
        <FormItem v-if="isDirectoryType" label="默认跳转">
          <Input v-model:value="formData.redirect" placeholder="/system/user" />
        </FormItem>
        <FormItem v-if="isButtonType" label="权限码" required>
          <Input v-model:value="formData.permissionCode" placeholder="user:create" />
        </FormItem>
        <FormItem v-if="isButtonType" label="说明">
          <Input v-model:value="formData.description" placeholder="用于按钮显隐和权限分配" />
        </FormItem>
        <FormItem v-if="!isButtonType" label="图标">
          <IconPicker
            v-model="formData.icon"
            prefix="lucide"
            :input-component="iconPickerInput"
            icon-slot="addonAfter"
            model-value-prop="value"
            placeholder="请选择图标"
          />
        </FormItem>
        <FormItem v-if="!isButtonType" label="排序">
          <InputNumber v-model:value="formData.sort" class="w-full" :min="0" />
        </FormItem>
        <FormItem v-if="!isButtonType" label="显示">
          <Select v-model:value="formData.visible" :options="visibleOptions" />
        </FormItem>
        <FormItem v-if="!isButtonType" label="状态">
          <Select v-model:value="formData.status" :options="statusOptions" />
        </FormItem>
      </Form>
    </Modal>
  </Page>
</template>
