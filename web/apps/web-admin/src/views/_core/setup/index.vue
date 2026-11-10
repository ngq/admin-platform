<!-- 页面功能：数据库初始化指引向导 - 支持常见数据库类型手动填写连接、测试连通性、自动建库迁移与写入配置文件 -->
<script lang="ts" setup>
import type { FormInstance, Rule } from 'ant-design-vue/es/form';

import type { SetupApi } from '#/api/core/setup';

import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { useAccessStore } from '@vben/stores';

import {
  Alert,
  Button,
  Card,
  Descriptions,
  DescriptionsItem,
  Form,
  FormItem,
  Input,
  InputNumber,
  InputPassword,
  Result,
  Select,
  SelectOption,
  Spin,
  Steps,
  Tag,
} from 'ant-design-vue';

import {
  getSetupStatusApi,
  initializeDatabaseApi,
  probeSetupStatusSilentApi,
  testDatabaseConnectionApi,
} from '#/api/core/setup';
import { markDatabaseInitialized } from '#/router/guard';

defineOptions({ name: 'DatabaseSetupWizard' });

interface SetupFormModel {
  database: string;
  driver: string | undefined;
  host: string;
  password: string;
  port: number | undefined;
  username: string;
}

const DATABASE_OPTIONS = [
  { label: 'MySQL', value: 'mysql' },
  { label: 'MariaDB（兼容 MySQL）', value: 'mariadb' },
  { label: 'PostgreSQL', value: 'postgres' },
  { label: 'SQLite（本地文件数据库）', value: 'sqlite' },
];

const router = useRouter();
const accessStore = useAccessStore();

const formRef = ref<FormInstance>();
const pageLoading = ref(true);
const testing = ref(false);
const initializing = ref(false);
const currentStep = ref(0);

const setupStatus = ref<null | SetupApi.SetupStatus>(null);
const testResult = ref<null | SetupApi.TestConnectionResult>(null);
const testError = ref('');
const initResult = ref<null | SetupApi.InitResult>(null);
const serviceReady = ref(false);

// 所有连接参数均保持空白，要求用户手动选择与输入，不预填或写死连接信息
const formState = reactive<SetupFormModel>({
  driver: undefined,
  host: '',
  port: undefined,
  database: '',
  username: '',
  password: '',
});

const isSQLite = computed(() => formState.driver === 'sqlite');

const portPlaceholder = computed(() =>
  formState.driver === 'postgres' ? '例如：5432' : '例如：3306',
);
const usernamePlaceholder = computed(() =>
  formState.driver === 'postgres' ? '例如：postgres' : '例如：root',
);

const rules = computed<Record<string, Rule[]>>(() => {
  if (isSQLite.value) {
    const sqliteRules: Record<string, Rule[]> = {
      driver: [{ required: true, message: '请选择数据库类型', trigger: 'change' }],
      database: [
        {
          required: true,
          whitespace: true,
          message: '请输入 SQLite 数据库文件路径或名称',
          trigger: 'blur',
        },
      ],
    };
    return sqliteRules;
  }
  const serverRules: Record<string, Rule[]> = {
    driver: [{ required: true, message: '请选择数据库类型', trigger: 'change' }],
    host: [
      {
        required: true,
        whitespace: true,
        message: '请输入数据库主机或 IP',
        trigger: 'blur',
      },
    ],
    port: [
      {
        required: true,
        type: 'number',
        min: 1,
        max: 65_535,
        message: '请输入有效端口（1 - 65535）',
        trigger: 'change',
      },
    ],
    database: [
      {
        required: true,
        whitespace: true,
        message: '请输入数据库名称',
        trigger: 'blur',
      },
      {
        pattern: /^[a-zA-Z0-9_]{1,64}$/,
        message: '数据库名称仅支持字母、数字和下划线（1-64位）',
        trigger: 'blur',
      },
    ],
    username: [
      {
        required: true,
        whitespace: true,
        message: '请输入数据库账号',
        trigger: 'blur',
      },
    ],
  };
  return serverRules;
});

const appName = computed(() => preferences.app.name || 'Admin Platform');

function formatTargetSummary(
  driver?: string,
  host?: string,
  port?: number,
  database?: string,
) {
  if (driver === 'sqlite' || (!host && !port)) {
    return database || '-';
  }
  return `${host || '-'}:${port || '-'} / ${database || '-'}`;
}

function handleDriverChange() {
  testResult.value = null;
  testError.value = '';
  formRef.value?.clearValidate();
}

function buildSubmitPayload(): SetupApi.DatabaseInitParams {
  return {
    driver: formState.driver || '',
    host: isSQLite.value ? '' : formState.host.trim(),
    port: isSQLite.value ? 0 : Number(formState.port),
    database: formState.database.trim(),
    username: isSQLite.value ? '' : formState.username.trim(),
    password: isSQLite.value ? '' : formState.password,
  };
}

async function loadStatus() {
  pageLoading.value = true;
  try {
    const status = await getSetupStatusApi();
    setupStatus.value = status;
    if (status.initialized && !initResult.value) {
      markDatabaseInitialized(true);
      serviceReady.value = true;
      await router.replace(LOGIN_PATH);
    }
  } finally {
    pageLoading.value = false;
  }
}

async function handleTestConnection() {
  testResult.value = null;
  testError.value = '';
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  testing.value = true;
  try {
    testResult.value = await testDatabaseConnectionApi(buildSubmitPayload());
  } catch (error: any) {
    testError.value =
      error?.message ||
      error?.response?.data?.message ||
      '无法连接到数据库，请检查连接参数是否正确';
  } finally {
    testing.value = false;
  }
}

async function waitForNormalServerReady() {
  for (let i = 0; i < 20; i++) {
    await new Promise((resolve) => setTimeout(resolve, 400));
    const status = await probeSetupStatusSilentApi();
    if (status?.initialized) {
      serviceReady.value = true;
      markDatabaseInitialized(true);
      return;
    }
  }
  serviceReady.value = true;
  markDatabaseInitialized(true);
}

async function handleInitialize() {
  testError.value = '';
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  initializing.value = true;
  currentStep.value = 1;
  try {
    const res = await initializeDatabaseApi(buildSubmitPayload());
    initResult.value = res;
    accessStore.setAccessToken(null);
    accessStore.setRefreshToken(null);
    await waitForNormalServerReady();
    currentStep.value = 2;
  } catch (error: any) {
    currentStep.value = 0;
    testError.value =
      error?.message ||
      error?.response?.data?.message ||
      '初始化数据库失败，请检查权限与配置后重试';
  } finally {
    initializing.value = false;
  }
}

function goToLogin() {
  router.replace(LOGIN_PATH);
}

onMounted(() => {
  void loadStatus();
});
</script>

<template>
  <div
    class="bg-background flex min-h-screen flex-col items-center justify-center px-4 py-10"
  >
    <div class="w-full max-w-2xl">
      <div class="mb-6 text-center">
        <h1 class="text-foreground text-2xl font-semibold tracking-tight">
          {{ appName }} · 数据库初始化向导
        </h1>
        <p class="text-muted-foreground mt-2 text-sm">
          手动选择数据库类型并填写连接信息，系统将自动完成建库、Goose 迁移建表、初始权限写入与配置文件持久化
        </p>
      </div>

      <Card :bordered="true" class="shadow-sm">
        <Spin :spinning="pageLoading">
          <Steps
            :current="currentStep"
            :items="[
              { title: '填写连接参数' },
              { title: '创建库表与配置' },
              { title: '初始化完成' },
            ]"
            class="mb-8"
            size="small"
          />

          <template v-if="currentStep === 0 || currentStep === 1">
            <Alert
              v-if="setupStatus?.reason && !testError"
              :message="setupStatus.reason"
              class="mb-5"
              show-icon
              type="info"
            />

              <Alert
                v-if="testError"
                :message="testError"
                class="mb-5"
                closable
                show-icon
                type="error"
                @close="testError = ''"
              />

              <Alert
                v-if="testResult && !testError"
                :description="
                  testResult.server_version
                    ? `数据库服务版本：${testResult.server_version}`
                    : undefined
                "
                :message="testResult.message"
                class="mb-5"
                show-icon
                type="success"
              />

              <Form
                ref="formRef"
                :disabled="initializing"
                :model="formState"
                :rules="rules"
                layout="vertical"
              >
                <div
                  :class="
                    isSQLite
                      ? 'grid grid-cols-1 gap-x-4 md:grid-cols-2'
                      : 'grid grid-cols-1 gap-x-4 md:grid-cols-3'
                  "
                >
                  <FormItem label="数据库" name="driver">
                    <Select
                      v-model:value="formState.driver"
                      placeholder="请选择数据库"
                      @change="handleDriverChange"
                    >
                      <SelectOption
                        v-for="item in DATABASE_OPTIONS"
                        :key="item.value"
                        :value="item.value"
                      >
                        {{ item.label }}
                      </SelectOption>
                    </Select>
                  </FormItem>

                  <template v-if="!isSQLite">
                    <FormItem label="主机 / IP" name="host">
                      <Input
                        v-model:value="formState.host"
                        placeholder="例如：127.0.0.1"
                      />
                    </FormItem>

                    <FormItem label="端口" name="port">
                      <InputNumber
                        v-model:value="formState.port"
                        :max="65535"
                        :min="1"
                        class="w-full"
                        :placeholder="portPlaceholder"
                      />
                    </FormItem>
                  </template>

                  <FormItem
                    v-else
                    label="数据库文件路径 / 名称"
                    name="database"
                  >
                    <Input
                      v-model:value="formState.database"
                      placeholder="例如：data/admin_platform.db"
                    />
                  </FormItem>
                </div>

                <template v-if="!isSQLite">
                  <FormItem label="数据库名称" name="database">
                    <Input
                      v-model:value="formState.database"
                      placeholder="例如：admin_platform（不存在时将自动创建）"
                    />
                  </FormItem>

                  <div class="grid grid-cols-1 gap-x-4 md:grid-cols-2">
                    <FormItem label="数据库账号" name="username">
                      <Input
                        v-model:value="formState.username"
                        autocomplete="off"
                        :placeholder="usernamePlaceholder"
                      />
                    </FormItem>

                    <FormItem label="数据库密码" name="password">
                      <InputPassword
                        v-model:value="formState.password"
                        autocomplete="new-password"
                        placeholder="请输入数据库密码"
                      />
                    </FormItem>
                  </div>
                </template>

                <div
                  class="mt-4 flex flex-wrap items-center justify-between gap-3"
                >
                  <Button
                    :disabled="initializing"
                    :loading="testing"
                    @click="handleTestConnection"
                  >
                    测试连接
                  </Button>

                  <Button
                    :loading="initializing"
                    type="primary"
                    @click="handleInitialize"
                  >
                    {{
                      initializing
                        ? '正在初始化数据库与写入配置...'
                        : '初始化数据库并保存配置'
                    }}
                  </Button>
                </div>
              </Form>
            </template>

            <template v-else-if="currentStep === 2 && initResult">
              <Result
                status="success"
                sub-title="数据库建库、Goose 迁移建表、基础权限初始化及配置文件写入已全部完成。"
                title="数据库初始化成功"
              >
                <template #extra>
                  <Button
                    :loading="!serviceReady"
                    type="primary"
                    @click="goToLogin"
                  >
                    前往登录页
                  </Button>
                </template>
                <Descriptions :column="1" bordered size="small">
                  <DescriptionsItem label="数据库">
                    <Tag color="blue">{{ initResult.driver }}</Tag>
                    {{
                      formatTargetSummary(
                        initResult.driver,
                        initResult.host,
                        initResult.port,
                        initResult.database,
                      )
                    }}
                  </DescriptionsItem>
                  <DescriptionsItem label="已写入配置文件">
                    <code>{{ initResult.config_file }}</code>
                    <span class="text-muted-foreground ml-2 text-xs">
                      (数据库驱动与连接字符串已自动持久化)
                    </span>
                  </DescriptionsItem>
                  <DescriptionsItem label="默认管理员账号">
                    <code>{{ initResult.default_username }}</code> /
                    <code>admin123</code>
                    <Tag class="ml-2" color="orange">登录后请及时修改密码</Tag>
                  </DescriptionsItem>
                </Descriptions>
              </Result>
            </template>
        </Spin>
      </Card>
    </div>
  </div>
</template>
