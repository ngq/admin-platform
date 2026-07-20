<!-- 页面功能：登录日志 - 用户登录记录查询与删除 -->
<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { h } from 'vue';
import {
  message,
  Tag,
  Modal,
  Button,
  Form,
  FormItem,
  Input,
  DatePicker,
  Table,
  Space,
} from 'ant-design-vue';
import { ref } from 'vue';

import { getLoginLogListApi, deleteLoginLogApi } from '#/api/core/login-log';

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);

const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  username: '',
  startTime: '',
  endTime: '',
});

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getLoginLogListApi({ ...searchForm.value, page, pageSize });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } catch {
    message.error('获取列表失败');
  } finally {
    loading.value = false;
  }
}

async function handleDelete(id: string) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这条登录日志吗？',
    async onOk() {
      try {
        await deleteLoginLogApi(id);
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
  searchForm.value = { username: '', startTime: '', endTime: '' };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

const columns = [
  { title: '用户名', dataIndex: 'username', key: 'username', width: 120 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: number }) =>
      h(Tag, { color: text === 1 ? 'green' : 'red' }, () => (text === 1 ? '成功' : '失败')),
  },
  { title: 'IP地址', dataIndex: 'ip', key: 'ip', width: 150 },
  { title: '登录时间', dataIndex: 'loginTime', key: 'loginTime', width: 180 },
  { title: '描述', dataIndex: 'description', key: 'description', ellipsis: true },
  { title: '操作', key: 'action', width: 100 },
];

fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="用户名">
        <Input v-model:value="searchForm.username" placeholder="请输入用户名" />
      </FormItem>
      <FormItem label="开始时间">
        <DatePicker value-format="YYYY-MM-DD" v-model:value="searchForm.startTime" placeholder="请选择开始时间" />
      </FormItem>
      <FormItem label="结束时间">
        <DatePicker value-format="YYYY-MM-DD" v-model:value="searchForm.endTime" placeholder="请选择结束时间" />
      </FormItem>
      <FormItem>
        <Space>
          <Button type="primary" @click="handleSearch">搜索</Button>
          <Button @click="handleReset">重置</Button>
        </Space>
      </FormItem>
    </Form>

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
          <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
        </template>
      </template>
    </Table>
  </Page>
</template>
