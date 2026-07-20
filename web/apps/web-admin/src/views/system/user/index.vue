<!-- 页面功能：用户管理 - 用户列表展示、创建、编辑、删除、角色分配、密码重置 -->
<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { h } from 'vue';
import {
  message,
  Tag,
  Button,
  Form,
  FormItem,
  Input,
  InputPassword,
  Select,
  SelectOption,
  Table,
  Modal,
  Space,
} from 'ant-design-vue';
import { ref } from 'vue';

import {
  getDepartmentTreeApi,
} from '#/api/core/department';
import {
  assignUserRolesApi,
  createUserApi,
  deleteUserApi,
  getRoleOptionsApi,
  getUserApi,
  getUserListApi,
  updateUserApi,
} from '#/api/core/user';
import { formatDateTime } from '#/utils/time';

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 0 },
];

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);

const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  username: '',
  deptId: undefined as string | undefined,
  status: undefined as number | undefined,
});

// 弹框
const modalVisible = ref(false);
const modalTitle = ref('新增用户');
const isEdit = ref(false);
const currentId = ref('');

const formData = ref({
  username: '',
  nickname: '',
  deptId: '0',
  password: '',
  status: 1,
  roleIds: [] as string[],
});

// 分配角色弹框
const roleModalVisible = ref(false);
const currentUserId = ref('');
const roleOptions = ref<any[]>([]);
const selectedRoleIds = ref<string[]>([]);
const departmentOptions = ref<any[]>([]);

function flattenDepartments(items: any[], level = 0): any[] {
  return items.flatMap((item) => [
    { label: `${'　'.repeat(level)}${item.name}`, value: item.id },
    ...flattenDepartments(item.children || [], level + 1),
  ]);
}

async function fetchDepartmentOptions() {
  try {
    const depts = await getDepartmentTreeApi();
    departmentOptions.value = [
      { label: '未分配部门', value: '0' },
      ...flattenDepartments(depts),
    ];
  } catch {
    departmentOptions.value = [{ label: '未分配部门', value: '0' }];
  }
}

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getUserListApi({ ...searchForm.value, page, pageSize });
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
  modalTitle.value = '新增用户';
  formData.value = { username: '', nickname: '', deptId: '0', password: '', status: 1, roleIds: [] };
  fetchDepartmentOptions();
  modalVisible.value = true;
}

async function openEditModal(id: string) {
  isEdit.value = true;
  modalTitle.value = '编辑用户';
  currentId.value = id;
  try {
    const data = await getUserApi(id);
    formData.value = {
      username: data.username,
      nickname: data.nickname,
      deptId: data.deptId || '0',
      password: '',
      status: data.status,
      roleIds: data.roles?.map((r: any) => r.id) || [],
    };
    await fetchDepartmentOptions();
    modalVisible.value = true;
  } catch {
    message.error('获取详情失败');
  }
}

async function handleSave() {
  try {
    if (isEdit.value) {
      await updateUserApi(currentId.value, formData.value);
      message.success('更新成功');
    } else {
      await createUserApi(formData.value);
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
    await deleteUserApi(id);
    message.success('删除成功');
    await fetchList();
  } catch {
    message.error('删除失败');
  }
}

async function openRoleModal(id: string) {
  currentUserId.value = id;
  let data: any;
  try {
    data = await getUserApi(id);
  } catch {
    message.error('获取用户详情失败');
    return;
  }
  try {
    const res = await getRoleOptionsApi();
    selectedRoleIds.value = data.roles?.map((r: any) => String(r.id)) || [];
    roleOptions.value = res.list || [];
    roleModalVisible.value = true;
  } catch {
    message.error('获取角色列表失败');
  }
}

async function handleSaveRoles() {
  try {
    await assignUserRolesApi({ id: currentUserId.value, roleIds: selectedRoleIds.value });
    message.success('分配成功');
    roleModalVisible.value = false;
  } catch {
    message.error('分配失败');
  }
}

function handleSearch() {
  pagination.value.page = 1;
  fetchList();
}

function handleReset() {
  searchForm.value = { username: '', deptId: undefined, status: undefined };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

const columns = [
  { title: '用户名', dataIndex: 'username', key: 'username' },
  { title: '昵称', dataIndex: 'nickname', key: 'nickname' },
  { title: '所属部门', dataIndex: 'deptName', key: 'deptName', width: 140, customRender: ({ text }: { text: string }) => text || '-' },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }: { text: number }) =>
      h(Tag, { color: text === 1 ? 'green' : 'red' }, () => (text === 1 ? '正常' : '禁用')),
  },
  {
    title: '角色',
    dataIndex: 'roles',
    key: 'roles',
    customRender: ({ text }: { text: any[] }) =>
      h('div', {}, (text || []).map((r: any) => h(Tag, { key: r.id }, () => r.name))),
  },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180, customRender: ({ text }: { text: string }) => formatDateTime(text) },
  { title: '操作', key: 'action', width: 200 },
];

fetchList();
fetchDepartmentOptions();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="用户名">
        <Input v-model:value="searchForm.username" placeholder="请输入用户名" />
      </FormItem>
      <FormItem label="所属部门">
        <Select v-model:value="searchForm.deptId" allow-clear placeholder="请选择部门" style="width: 160px">
          <SelectOption v-for="opt in departmentOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </SelectOption>
        </Select>
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
            <Button size="small" type="link" @click="openRoleModal(record.id)">分配角色</Button>
            <Button size="small" type="link" @click="openEditModal(record.id)">编辑</Button>
            <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="modalVisible" :title="modalTitle" @ok="handleSave">
      <Form :label-width="80">
        <FormItem label="用户名" required>
          <Input v-model:value="formData.username" placeholder="请输入用户名" />
        </FormItem>
        <FormItem label="昵称" required>
          <Input v-model:value="formData.nickname" placeholder="请输入昵称" />
        </FormItem>
        <FormItem label="所属部门">
          <Select v-model:value="formData.deptId" placeholder="请选择所属部门">
            <SelectOption v-for="opt in departmentOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem v-if="!isEdit" label="密码" required>
          <InputPassword v-model:value="formData.password" placeholder="请输入密码" />
        </FormItem>
        <FormItem label="状态" required>
          <Select v-model:value="formData.status" placeholder="请选择状态">
            <SelectOption v-for="opt in statusOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
        <FormItem label="角色">
          <Select v-model:value="formData.roleIds" mode="multiple" placeholder="请选择角色">
            <SelectOption v-for="opt in roleOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
        </FormItem>
      </Form>
    </Modal>

    <Modal v-model:open="roleModalVisible" title="分配角色" @ok="handleSaveRoles">
      <Select v-model:value="selectedRoleIds" mode="multiple" placeholder="请选择角色" class="w-full">
        <SelectOption v-for="opt in roleOptions" :key="opt.value" :value="opt.value">
          {{ opt.label }}
        </SelectOption>
      </Select>
    </Modal>
  </Page>
</template>
