<!-- 页面功能：系统参数管理 - 参数列表展示、创建、编辑、删除 -->
<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { h } from 'vue';
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
import { ref } from 'vue';

import {
  createParamApi,
  deleteParamApi,
  getParamApi,
  getParamListApi,
  updateParamApi,
} from '#/api/core/param';
import { formatDateTime } from '#/utils/time';

const statusOptions = [
  { label: '正常', value: 1 },
  { label: '禁用', value: 0 },
];

const categoryOptions = [
  { label: '系统', value: 'system' },
  { label: '业务', value: 'business' },
  { label: '其他', value: 'other' },
];

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);

const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  paramKey: '',
  category: '',
  status: undefined as number | undefined,
});

const modalVisible = ref(false);
const modalTitle = ref('新增参数');
const isEdit = ref(false);
const currentId = ref('');

const formData = ref({
  paramKey: '',
  paramValue: '',
  category: 'system',
  description: '',
  status: 1,
});

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getParamListApi({ ...searchForm.value, page, pageSize });
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
  modalTitle.value = '新增参数';
  formData.value = { paramKey: '', paramValue: '', category: 'system', description: '', status: 1 };
  modalVisible.value = true;
}

async function openEditModal(id: string) {
  isEdit.value = true;
  modalTitle.value = '编辑参数';
  currentId.value = id;
  try {
    const data = await getParamApi(id);
    formData.value = {
      paramKey: data.paramKey,
      paramValue: data.paramValue,
      category: data.category,
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
      await updateParamApi(currentId.value, formData.value);
      message.success('更新成功');
    } else {
      await createParamApi(formData.value);
      message.success('新增成功');
    }
    modalVisible.value = false;
    await fetchList();
  } catch {
    message.error('操作失败');
  }
}

async function handleDelete(id: string) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这条参数吗？',
    async onOk() {
      try {
        await deleteParamApi(id);
        message.success('删除成功');
        await fetchList();
      } catch {
        message.error('删除失败');
      }
    },
  });
}

function handleSearch() {
  pagination.value.page = 1;
  fetchList();
}

function handleReset() {
  searchForm.value = { paramKey: '', category: '', status: undefined };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

const columns = [
  { title: '参数键', dataIndex: 'paramKey', key: 'paramKey' },
  { title: '参数值', dataIndex: 'paramValue', key: 'paramValue', ellipsis: true },
  {
    title: '分类',
    dataIndex: 'category',
    key: 'category',
    width: 100,
    customRender: ({ text }: { text: string }) => {
      const item = categoryOptions.find((o) => o.value === text);
      return item?.label || text;
    },
  },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 80,
    customRender: ({ text }: { text: number }) =>
      h(Tag, { color: text === 1 ? 'green' : 'red' }, () => (text === 1 ? '正常' : '禁用')),
  },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180, customRender: ({ text }: { text: string }) => formatDateTime(text) },
  { title: '操作', key: 'action', width: 150 },
];

fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="参数键">
        <Input v-model:value="searchForm.paramKey" placeholder="请输入参数键" />
      </FormItem>
      <FormItem label="分类">
        <Select v-model:value="searchForm.category" placeholder="请选择分类" style="width: 120px" allow-clear>
          <SelectOption v-for="opt in categoryOptions" :key="opt.value" :value="opt.value">
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
            <Button size="small" type="link" @click="openEditModal(record.id)">编辑</Button>
            <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="modalVisible" :title="modalTitle" @ok="handleSave">
      <Form :label-width="80">
        <FormItem label="参数键" required>
          <Input v-model:value="formData.paramKey" placeholder="请输入参数键" />
        </FormItem>
        <FormItem label="参数值" required>
          <Input v-model:value="formData.paramValue" placeholder="请输入参数值" />
        </FormItem>
        <FormItem label="分类" required>
          <Select v-model:value="formData.category" placeholder="请选择分类">
            <SelectOption v-for="opt in categoryOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </SelectOption>
          </Select>
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
  </Page>
</template>
