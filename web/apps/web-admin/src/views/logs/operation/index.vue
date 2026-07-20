<!-- 页面功能：操作日志 - 系统操作记录查询与删除 -->
<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import {
  message,
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

import { getOperationLogListApi, deleteOperationLogApi } from '#/api/core/operation-log';

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);

const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  operator: '',
  startTime: '',
  endTime: '',
});

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getOperationLogListApi({ ...searchForm.value, page, pageSize });
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
    content: '确定要删除这条操作日志吗？',
    async onOk() {
      try {
        await deleteOperationLogApi(id);
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
  searchForm.value = { operator: '', startTime: '', endTime: '' };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

function formatRequestBody(value?: string) {
  const body = String(value || '').trim();
  if (!body) return '-';

  try {
    return JSON.stringify(JSON.parse(body), null, 2);
  } catch {
    return body;
  }
}

async function copyRequestBody(value?: string) {
  const body = formatRequestBody(value);
  if (body === '-') return;

  try {
    await navigator.clipboard.writeText(body);
    message.success('复制成功');
  } catch {
    message.error('复制失败');
  }
}

const columns = [
  { title: '操作人', dataIndex: 'operator', key: 'operator', width: 120 },
  { title: '操作内容', dataIndex: 'operation', key: 'operation', ellipsis: true },
  { title: 'IP地址', dataIndex: 'ip', key: 'ip', width: 150 },
  { title: '耗时(ms)', dataIndex: 'duration', key: 'duration', width: 100 },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '操作', key: 'action', width: 100 },
];

columns.splice(2, 0, { title: '请求Body', dataIndex: 'requestBody', key: 'requestBody', width: 320 });

fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="操作人">
        <Input v-model:value="searchForm.operator" placeholder="请输入操作人" />
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
      :scroll="{ x: 1200 }"
      @change="(p: any) => handlePageChange(p.current, p.pageSize)"
    >
      <template #bodyCell="{ column, record }">
        <template v-if="column.key === 'requestBody'">
          <div v-if="record.requestBody" class="request-body-cell">
            <pre class="request-body-preview">{{ formatRequestBody(record.requestBody) }}</pre>
            <Button size="small" type="link" @click="copyRequestBody(record.requestBody)">复制</Button>
          </div>
          <span v-else>-</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
        </template>
      </template>
    </Table>
  </Page>
</template>

<style scoped>
.request-body-cell {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.request-body-preview {
  max-height: 72px;
  max-width: 240px;
  margin: 0;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 18px;
}
</style>
