import { baseRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    username: string;
    password: string;
    captchaId?: string;
    captchaCode?: string;
  }

  /** 后端登录响应 */
  export interface BackendLoginResult {
    access_token: string;
    refresh_token: string;
    expires_at: string;
  }

  /** 登录接口返回值（前端格式） */
  export interface LoginResult {
    accessToken: string;
    refreshToken: string;
  }

  export interface RefreshTokenResult {
    access_token: string;
    refresh_token: string;
    expires_at: string;
  }
}

/**
 * 登录
 * 适配后端 /auth/login 接口
 */
export async function loginApi(data: AuthApi.LoginParams): Promise<AuthApi.LoginResult> {
  const result = await requestClient.post<AuthApi.BackendLoginResult>('/auth/login', data);

  // 转换后端响应格式为前端格式
  return {
    accessToken: result.access_token,
    refreshToken: result.refresh_token,
  };
}

/**
 * 刷新accessToken
 * 需要传入 refreshToken
 */
export async function refreshTokenApi(refreshToken: string) {
  return requestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh', {
    refreshToken,
  });
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return baseRequestClient.post('/auth/logout', {
    withCredentials: true,
  });
}

/**
 * 获取用户权限码
 */
export async function getAccessCodesApi(): Promise<string[]> {
  return requestClient.get<string[]>('/auth/codes');
}
