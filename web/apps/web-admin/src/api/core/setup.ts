import { baseRequestClient, requestClient } from '#/api/request';

export namespace SetupApi {
  export interface SetupStatus {
    config_file?: string;
    database?: string;
    driver: string;
    host?: string;
    initialized: boolean;
    port?: number;
    reason?: string;
  }

  export interface DatabaseInitParams {
    database: string;
    driver: string;
    host: string;
    password: string;
    port: number;
    username: string;
  }

  export interface TestConnectionResult {
    connected: boolean;
    database_exists: boolean;
    message: string;
    server_version?: string;
  }

  export interface InitResult {
    config_file: string;
    database: string;
    default_username: string;
    driver: string;
    host: string;
    initialized: boolean;
    message: string;
    port: number;
  }
}

/** 获取数据库初始化状态 */
export async function getSetupStatusApi(): Promise<SetupApi.SetupStatus> {
  return requestClient.get<SetupApi.SetupStatus>('/setup/status');
}

/** 静默探测初始化状态（热切换重启期间不触发错误弹窗） */
export async function probeSetupStatusSilentApi(): Promise<null | SetupApi.SetupStatus> {
  try {
    const resp = await baseRequestClient.get<{
      data?: {
        code: number;
        data?: SetupApi.SetupStatus;
      };
    }>('/setup/status', { timeout: 2500 });
    const body = resp?.data;
    if (body?.code === 0 && body.data) {
      return body.data;
    }
    return null;
  } catch {
    return null;
  }
}

/** 测试数据库连接 */
export async function testDatabaseConnectionApi(
  data: SetupApi.DatabaseInitParams,
): Promise<SetupApi.TestConnectionResult> {
  return requestClient.post<SetupApi.TestConnectionResult>(
    '/setup/test-connection',
    data,
  );
}

/** 初始化数据库并写入配置文件 */
export async function initializeDatabaseApi(
  data: SetupApi.DatabaseInitParams,
): Promise<SetupApi.InitResult> {
  return requestClient.post<SetupApi.InitResult>('/setup/database', data, {
    timeout: 30_000,
  });
}
