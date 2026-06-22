// api/rateLimiter.ts
export interface RateLimitConfig {
  maxRequests: number
  timeWindow: number
  enabled?: boolean
}

class RateLimiter {
  private requests: number[] = []
  private config: RateLimitConfig

  constructor(config: RateLimitConfig) {
    this.config = config
  }

  canExecute(): boolean {
    if (this.config.enabled === false) return true
    
    const now = Date.now()
    this.requests = this.requests.filter(t => now - t < this.config.timeWindow)
    
    if (this.requests.length >= this.config.maxRequests) {
      return false
    }
    
    this.requests.push(now)
    return true
  }

  reset(): void {
    this.requests = []
  }

  getRemainingRequests(): number {
    if (this.config.enabled === false) return Infinity
    const now = Date.now()
    const validRequests = this.requests.filter(t => now - t < this.config.timeWindow)
    return Math.max(0, this.config.maxRequests - validRequests.length)
  }
}

// 预定义的限流器实例
export const loginLimiter = new RateLimiter({ maxRequests: 5, timeWindow: 60000 })
export const registerLimiter = new RateLimiter({ maxRequests: 3, timeWindow: 60000 })
export const uploadLimiter = new RateLimiter({ maxRequests: 100, timeWindow: 60000 })
export const listLimiter = new RateLimiter({ maxRequests: 30, timeWindow: 60000 })
export const defaultLimiter = new RateLimiter({ maxRequests: 60, timeWindow: 60000 })