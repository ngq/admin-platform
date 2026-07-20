<!-- 页面功能：部门管理 - 维护组织部门树，供角色数据范围过滤使用 -->
<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { h, ref } from 'vue';
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
  createDepartmentApi,
  deleteDepartmentApi,
  getDepartmentApi,
  getDepartmentTreeApi,
  updateDepartmentApi,
} from '#/api/core/department';
import { useAccessCodes } from '#/composables/useAccessCodes';
import { formatDateTime } from '#/utils/time';

const { hasAccessCode } = useAccessCodes();

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 0 },
];

const tableData = ref<any[]>([]);
const loading = ref(false);
const searchForm = ref({
  name: '',
  status: undefined as number | undefined,
});

const modalVisible = ref(false);
const modalTitle = ref('新增部门');
const isEdit = ref(false);
const currentId = ref('');
const formData = ref({
  parentId: '0',
  name: '',
  leaderUserId: '0',
  sort: 0,
  status: 1,
});

function flattenDepartments(items: any[], level = 0): any[] {
  return items.flatMap((item) => [
    { ...item, optionLabel: `${'　'.repeat(level)}${item.name}` },
    ...flattenDepartments(item.children || [], level + 1),
  ]);
}

const departmentOptions = ref<any[]>([]);

async function fetchList() {
  loading.value = true;
  try {
    const data = await getDepartmentTreeApi(searchForm.value);
    tableData.value = data;
    departmentOptions.value = flattenDepartments(data);
  } catch {
    message.error('获取部门列表失败');
  } finally {
    loading.value = false;
  }
}

function openAddModal(parentId = '0') {
  isEdit.value = false;
  modalTitle.value = '新增部门';
  currentId.value = '';
  formData.value = { parentId, name: '', leaderUserId: '0', sort: 0, status: 1 };
  modalVisible.value = true;
}

async function openEditModal(id: string) {
  isEdit.value = true;
  modalTitle.value = '编辑部门';
  currentId.value = id;
  try {
    const data = await getDepartmentApi(id);
    formData.value = {
      parentId: data.parentId,
      name: data.name,
      leaderUserId: data.leaderUserId || '0',
      sort: data.sort,
      status: data.status,
    };
    modalVisible.value = true;
  } catch {
    message.error('获取部门详情失败');
  }
}

async function handleSave() {
  try {
    if (isEdit.value) {
      await updateDepartmentApi(currentId.value, formData.value);
      message.success('更新成功');
    } else {
      await createDepartmentApi(formData.value);
      message.success('新增成功');
    }
    modalVisible.value = false;
    await fetchList();
  } catch (error: any) {
    message.error(error?.message || '操作失败');
  }
}

function handleDelete(id: string) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这个部门吗？删除前请确认没有用户继续归属于该部门。',
    async onOk() {
      try {
        await deleteDepartmentApi(id);
        message.success('删除成功');
        await fetchList();
      } catch {
        message.error('删除失败');
      }
    },
  });
}

function handleSearch() {
  fetchList();
}

function handleReset() {
  searchForm.value = { name: '', status: undefined };
  fetchList();
}

const columns = [
  { title: '部门名称', dataIndex: 'name', key: 'name' },
  { title: '负责人用户ID', dataIndex: 'leaderUserId', key: 'leaderUserId', width: 140 },
  { title: '排序', dataIndex: 'sort', key: 'sort', width: 90 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 90,
    customRender: ({ text }: { text: number }) =>
      h(Tag, { color: text === 1 ? 'green' : 'red' }, () => (text === 1 ? '正常' : '禁用')),
  },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180, customRender: ({ text }: { text: string }) => formatDateTime(text) },
  { title: '操作', key: 'action', width: 220 },
];

fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="部门名称">
        <Input v-model:value="searchForm.name" placeholder="请输入部门名称" />
      </FormItem>
      <FormItem label="状态">
        <Select v-model:value="searchForm.status" allow-clear placeholder="请选择状态" style="width: 120px">
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
      <Button v-if="hasAccessCode('department:create')" type="primary" @click="openAddModal()">新增</Button>
    </div>

    <Table
      :columns="columns"
      :data-source="tableData"
      :loading="loading"
      :pagination="false"
      :row-key="(record: any) => record.id"
      :scroll="{ x: 900 }"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'action'">
          <Space>
            <Button v-if="hasAccessCode('department:create')" size="small" type="link" @click="openAddModal(record.id)">
              新增下级
            </Button>
            <Button v-if="hasAccessCode('department:update')" size="small" type="link" @click="openEditModal(record.id)">
              编辑
            </Button>
            <Button v-if="hasAccessCode('department:delete')" size="small" danger type="link" @click="handleDelete(record.id)">
              删除
            </Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="modalVisible" :title="modalTitle" @ok="handleSave">
      <Form :label-width="100">
        <FormItem label="上级部门">
          <Select v-model:value="formData.parentId" placeholder="请选择上级部门">
            <SelectOption value="0">顶级部门</SelectOption>
            <SelectOption
              v-for="dept in departmentOptions.filter((item) => item.id !== currentId)"
              :key="dept.id"
              :value="dept.id"
            >
              {{ dept.optionLabel }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="部门名称" required>
          <Input v-model:value="formData.name" placeholder="请输入部门名称" />
        </FormItem>
        <FormItem label="负责人用户ID">
          <Input v-model:value="formData.leaderUserId" placeholder="请输入负责人用户ID" />
        </FormItem>
        <FormItem label="排序">
          <InputNumber v-model:value="formData.sort" :min="0" />
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
  </Page>
</template>
