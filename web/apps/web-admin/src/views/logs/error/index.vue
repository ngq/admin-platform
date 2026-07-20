<!-- 页面功能：错误日志 - 系统错误记录查询、处理与删除 -->
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
  Descriptions,
  DescriptionsItem,
} from 'ant-design-vue';
import { ref } from 'vue';

import { getErrorLogListApi, getErrorLogApi, resolveErrorLogApi, deleteErrorLogApi } from '#/api/core/error-log';

const tableData = ref<any[]>([]);
const total = ref(0);
const loading = ref(false);

const pagination = ref({ page: 1, pageSize: 20 });

const searchForm = ref({
  path: '',
  startTime: '',
  endTime: '',
});

const detailVisible = ref(false);
const detailData = ref<any>(null);

async function fetchList() {
  loading.value = true;
  try {
    const { page, pageSize } = pagination.value;
    const res = await getErrorLogListApi({ ...searchForm.value, page, pageSize });
    tableData.value = res.list || [];
    total.value = res.total || 0;
  } catch {
    message.error('获取列表失败');
  } finally {
    loading.value = false;
  }
}

async function openDetailModal(id: string) {
  try {
    detailData.value = await getErrorLogApi(id);
    detailVisible.value = true;
  } catch {
    message.error('获取详情失败');
  }
}

async function handleResolve(id: string) {
  try {
    await resolveErrorLogApi(id);
    message.success('标记成功');
    await fetchList();
  } catch {
    message.error('标记失败');
  }
}

async function handleDelete(id: string) {
  Modal.confirm({
    title: '确认删除',
    content: '确定要删除这条错误日志吗？',
    async onOk() {
      try {
        await deleteErrorLogApi(id);
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
  searchForm.value = { path: '', startTime: '', endTime: '' };
  pagination.value.page = 1;
  fetchList();
}

function handlePageChange(page: number, pageSize: number) {
  pagination.value = { page, pageSize };
  fetchList();
}

const columns = [
  { title: '错误信息', dataIndex: 'errorMessage', key: 'errorMessage', ellipsis: true },
  { title: 'IP地址', dataIndex: 'ip', key: 'ip', width: 150 },
  { title: '发生时间', dataIndex: 'occurTime', key: 'occurTime', width: 180 },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    width: 100,
    customRender: ({ text }: { text: number }) =>
      h(Tag, { color: text === 1 ? 'green' : 'orange' }, () => (text === 1 ? '已解决' : '未解决')),
  },
  { title: '创建时间', dataIndex: 'createdAt', key: 'createdAt', width: 180 },
  { title: '操作', key: 'action', width: 200 },
];

columns.splice(1, 0, { title: '请求URL', dataIndex: 'path', key: 'path', ellipsis: true });

fetchList();
</script>

<template>
  <Page>
    <Form layout="inline" class="mb-4">
      <FormItem label="请求URL">
        <Input v-model:value="searchForm.path" placeholder="请输入请求 URL" allow-clear />
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
        <template v-if="column.key === 'path'">
          <span>{{ [record.method, record.path].filter(Boolean).join(' ') || '-' }}</span>
        </template>
        <template v-else-if="column.key === 'action'">
          <Space>
            <Button size="small" type="link" @click="openDetailModal(record.id)">详情</Button>
            <Button v-if="record.status === 0" size="small" type="link" @click="handleResolve(record.id)">标记已解决</Button>
            <Button size="small" danger type="link" @click="handleDelete(record.id)">删除</Button>
          </Space>
        </template>
      </template>
    </Table>

    <Modal v-model:open="detailVisible" title="错误详情" :footer="null">
      <Descriptions v-if="detailData" :column="1" bordered size="small">
        <DescriptionsItem label="错误信息">{{ detailData.errorMessage }}</DescriptionsItem>
        <DescriptionsItem label="请求URL">{{ [detailData.method, detailData.path].filter(Boolean).join(' ') || '-' }}</DescriptionsItem>
        <DescriptionsItem label="IP地址">{{ detailData.ip }}</DescriptionsItem>
        <DescriptionsItem label="发生时间">{{ detailData.occurTime }}</DescriptionsItem>
        <DescriptionsItem label="状态">{{ detailData.status === 1 ? '已解决' : '未解决' }}</DescriptionsItem>
        <DescriptionsItem label="创建时间">{{ detailData.createdAt }}</DescriptionsItem>
      </Descriptions>
    </Modal>
  </Page>
</template>
