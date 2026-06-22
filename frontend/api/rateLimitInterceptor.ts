// api/rateLimitInterceptor.ts
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import { 
  loginLimiter, 
  registerLimiter, 
  uploadLimiter, 
  listLimiter, 
  defaultLimiter 
} from './rateLimiter'

// 根据请求获取对应的限流器
function getLimiterForRequest(method: string, url: string) {
  const path = url.split('?')[0]
  const key = `${method}:${path}`
  
  switch (key) {
    case 'POST:/api/v1/auth/login':
      return loginLimiter
    case 'POST:/api/v1/devices/register':
      return registerLimiter
    case 'POST:/api/v1/upload/chunk':
      return uploadLimiter
    case 'GET:/api/v1/screenshots':
    case 'GET:/api/v1/recordings':
    case 'GET:/api/v1/activities':
      return listLimiter
    default:
      return defaultLimiter
  }
}

export function setupRateLimitInterceptor(axiosInstance: AxiosInstance) {
  axiosInstance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const method = config.method?.toUpperCase() || 'GET'
    const url = config.url || ''
    const limiter = getLimiterForRequest(method, url)
    
    if (!limiter.canExecute()) {
      throw new Error('请求过于频繁，请稍后再试')
    }
    
    return config
  })
}