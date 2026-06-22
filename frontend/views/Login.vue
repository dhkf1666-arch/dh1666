<template>
  <div class="login-container">
    <!-- 背景动画效果 -->
    <div class="bg-animation">
      <div class="bg-shape shape-1"></div>
      <div class="bg-shape shape-2"></div>
      <div class="bg-shape shape-3"></div>
      <div class="bg-shape shape-4"></div>
    </div>

    <!-- 登录卡片 -->
    <div class="login-wrapper">
      <div class="login-card" :class="{ 'card-loaded': cardLoaded }">
        <!-- Logo区域 -->
        <div class="logo-section">
          <div class="logo-icon">
            <el-icon :size="48"><Monitor /></el-icon>
          </div>
          <h1 class="system-title">DHPG管理后台</h1>
          <p class="system-subtitle">所有的坚持，终将美好</p>
        </div>

        <!-- 表单区域 -->
        <div class="form-section">
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            @submit.prevent="handleLogin"
            @keyup.enter="handleLogin"
          >
            <el-form-item prop="username">
              <el-input
                ref="usernameInput"
                v-model="form.username"
                placeholder="请输入用户名"
                size="large"
                :prefix-icon="User"
                class="custom-input"
                :disabled="loading"
                autofocus
                @keyup.enter="handleLogin"
              />
            </el-form-item>

            <el-form-item prop="password">
              <el-input
                v-model="form.password"
                type="password"
                placeholder="请输入密码"
                size="large"
                :prefix-icon="Lock"
                show-password
                class="custom-input"
                :disabled="loading"
                @keyup.enter="handleLogin"
              />
            </el-form-item>

            <!-- 验证码 -->
            <el-form-item prop="captcha">
              <div class="captcha-container">
                <el-input
                  v-model="form.captcha"
                  placeholder="请输入验证码"
                  size="large"
                  :prefix-icon="CircleCheck"
                  class="captcha-input"
                  maxlength="4"
                  :disabled="loading"
                  @keyup.enter="handleLogin"
                />
                <div
                  class="captcha-code"
                  @click="refreshCaptcha"
                  @keydown.enter="refreshCaptcha"
                  tabindex="0"
                  role="button"
                  aria-label="刷新验证码"
                >
                  <canvas ref="captchaCanvas" width="120" height="42"></canvas>
                  <el-icon class="refresh-icon" :size="20"
                    ><RefreshRight
                  /></el-icon>
                </div>
              </div>
            </el-form-item>

            <!-- 记住密码和忘记密码 -->
            <div class="form-options">
              <el-checkbox v-model="rememberMe" :disabled="loading"
                >记住用户名</el-checkbox
              >
              <el-link
                type="primary"
                :underline="false"
                @click="handleForgotPassword"
              >
                忘记密码？
              </el-link>
            </div>

            <el-form-item>
              <el-button
                type="primary"
                size="large"
                :loading="loading"
                @click="handleLogin"
                class="login-btn"
              >
                {{ loading ? "登录中..." : "登录系统" }}
              </el-button>
            </el-form-item>
          </el-form>

          <!-- 底部信息 -->
          <div class="login-footer">
            <span>DHPG监控系统</span>
            <span class="version">天 道 酬 勤</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 粒子效果 -->
    <div class="particles" ref="particlesRef"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import {
  User,
  Lock,
  Monitor,
  CircleCheck,
  RefreshRight,
} from "@element-plus/icons-vue";
import { useUserStore } from "@store/user";

const router = useRouter();
const userStore = useUserStore();

// 如果已经登录，直接跳转
if (userStore.isAuthenticated) {
  router.push("/");
}

const formRef = ref();
const usernameInput = ref();
const loading = ref(false);
const cardLoaded = ref(false);
const rememberMe = ref(false);
const captchaCanvas = ref<HTMLCanvasElement | null>(null);
const particlesRef = ref<HTMLDivElement | null>(null);

let currentCaptcha = "";
let captchaExpireTime = 0;
let animationId: number | null = null;
let particleAnimationId: number | null = null;
let stopParticlesCleanup: (() => void) | null = null;

const form = reactive({
  username: "",
  password: "",
  captcha: "",
});

// 验证码验证规则
const validateCaptcha = (_: any, value: string, callback: Function) => {
  if (!value) {
    callback(new Error("请输入验证码"));
  } else if (Date.now() > captchaExpireTime) {
    callback(new Error("验证码已过期，请刷新"));
    refreshCaptcha();
  } else if (value.toLowerCase() !== currentCaptcha.toLowerCase()) {
    callback(new Error("验证码错误"));
  } else {
    callback();
  }
};

const rules = {
  username: [
    { required: true, message: "请输入用户名", trigger: "blur" },
    { min: 2, max: 20, message: "用户名长度在2-20位之间", trigger: "blur" },
  ],
  password: [
    { required: true, message: "请输入密码", trigger: "blur" },
    { min: 6, message: "密码长度不能小于6位", trigger: "blur" },
  ],
  captcha: [
    { required: true, message: "请输入验证码", trigger: "blur" },
    { validator: validateCaptcha, trigger: "blur" },
  ],
};

// 生成随机验证码
const generateCaptcha = () => {
  const canvas = captchaCanvas.value;
  if (!canvas) return;

  const ctx = canvas.getContext("2d");
  if (!ctx) return;

  // 清空画布
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  // 设置背景色
  ctx.fillStyle = "#f5f7fa";
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  // 生成随机验证码（4位数字和字母组合，排除易混淆字符）
  const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz0123456789";
  let captcha = "";
  for (let i = 0; i < 4; i++) {
    captcha += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  currentCaptcha = captcha;
  captchaExpireTime = Date.now() + 5 * 60 * 1000; // 5分钟有效期

  // 绘制验证码文字
  for (let i = 0; i < captcha.length; i++) {
    ctx.save();
    ctx.font = `bold ${24 + Math.random() * 8}px "Comic Sans MS", cursive`;
    ctx.fillStyle = `rgb(${Math.random() * 100 + 50}, ${Math.random() * 100 + 50}, ${Math.random() * 100 + 50})`;
    ctx.translate(20 + i * 22, 28);
    ctx.rotate((Math.random() - 0.5) * 0.5);
    ctx.fillText(captcha[i], 0, 0);
    ctx.restore();
  }

  // 添加干扰线
  for (let i = 0; i < 8; i++) {
    ctx.beginPath();
    ctx.strokeStyle = `rgb(${Math.random() * 255}, ${Math.random() * 255}, ${Math.random() * 255})`;
    ctx.lineWidth = 1;
    ctx.moveTo(Math.random() * canvas.width, Math.random() * canvas.height);
    ctx.lineTo(Math.random() * canvas.width, Math.random() * canvas.height);
    ctx.stroke();
  }

  // 添加干扰点
  for (let i = 0; i < 50; i++) {
    ctx.fillStyle = `rgb(${Math.random() * 255}, ${Math.random() * 255}, ${Math.random() * 255})`;
    ctx.fillRect(
      Math.random() * canvas.width,
      Math.random() * canvas.height,
      1,
      1,
    );
  }
};

// 刷新验证码
const refreshCaptcha = () => {
  generateCaptcha();
  form.captcha = "";
};

// 粒子效果类
class Particle {
  x: number;
  y: number;
  size: number;
  speedX: number;
  speedY: number;
  color: string;

  constructor(canvasWidth: number, canvasHeight: number) {
    this.x = Math.random() * canvasWidth;
    this.y = Math.random() * canvasHeight;
    this.size = Math.random() * 2 + 1;
    this.speedX = (Math.random() - 0.5) * 0.5;
    this.speedY = (Math.random() - 0.5) * 0.5 + 0.2;
    this.color = `rgba(255, 255, 255, ${Math.random() * 0.3 + 0.1})`;
  }

  update(canvasWidth: number, canvasHeight: number) {
    this.x += this.speedX;
    this.y += this.speedY;

    if (this.x < 0) this.x = canvasWidth;
    if (this.x > canvasWidth) this.x = 0;
    if (this.y < 0) this.y = canvasHeight;
    if (this.y > canvasHeight) this.y = 0;
  }

  draw(ctx: CanvasRenderingContext2D) {
    ctx.fillStyle = this.color;
    ctx.fillRect(this.x, this.y, this.size, this.size);
  }
}

// 初始化粒子效果
const initParticles = (): (() => void) | null => {
  const container = particlesRef.value;
  if (!container) return null;

  const canvas = document.createElement("canvas");
  canvas.style.position = "absolute";
  canvas.style.top = "0";
  canvas.style.left = "0";
  canvas.style.width = "100%";
  canvas.style.height = "100%";
  canvas.style.pointerEvents = "none";
  container.appendChild(canvas);

  const ctx = canvas.getContext("2d");
  if (!ctx) return null;

  let particles: Particle[] = [];
  let width = window.innerWidth;
  let height = window.innerHeight;

  const resizeHandler = () => {
    width = window.innerWidth;
    height = window.innerHeight;
    canvas.width = width;
    canvas.height = height;
    initParticlesArray();
  };

  const initParticlesArray = () => {
    particles = [];
    const particleCount = Math.min(Math.floor((width * height) / 15000), 100);
    for (let i = 0; i < particleCount; i++) {
      particles.push(new Particle(width, height));
    }
  };

  const animate = () => {
    if (!ctx) return;
    ctx.clearRect(0, 0, width, height);

    for (const particle of particles) {
      particle.update(width, height);
      particle.draw(ctx);
    }

    particleAnimationId = requestAnimationFrame(animate);
  };

  window.addEventListener("resize", resizeHandler);
  resizeHandler();
  animate();

  // 返回清理函数
  return () => {
    window.removeEventListener("resize", resizeHandler);
    if (particleAnimationId) {
      cancelAnimationFrame(particleAnimationId);
      particleAnimationId = null;
    }
    canvas.remove();
  };
};

// 背景动画
const initBackgroundAnimation = () => {
  const shapes = document.querySelectorAll(".bg-shape");
  let angle = 0;

  const animate = () => {
    angle += 0.002;
    shapes.forEach((shape, index) => {
      const offset = (index * Math.PI * 2) / shapes.length;
      const x = Math.sin(angle + offset) * 20;
      const y = Math.cos(angle + offset) * 15;
      (shape as HTMLElement).style.transform = `translate(${x}px, ${y}px)`;
    });
    animationId = requestAnimationFrame(animate);
  };

  animate();

  return () => {
    if (animationId) {
      cancelAnimationFrame(animationId);
      animationId = null;
    }
  };
};

// 加载记住的用户名（只记住用户名，不记住密码）
const loadRememberedUser = () => {
  const savedUsername = localStorage.getItem("remember_username");
  if (savedUsername) {
    form.username = savedUsername;
    rememberMe.value = true;
  }
};

// 保存登录信息（只保存用户名）
const saveLoginInfo = () => {
  if (rememberMe.value) {
    localStorage.setItem("remember_username", form.username);
  } else {
    localStorage.removeItem("remember_username");
  }
};

const handleLogin = async () => {
  if (!formRef.value || loading.value) return;

  // 1. 表单验证（包含验证码的前端验证）
  try {
    await formRef.value.validate();
  } catch {
    return;
  }

  loading.value = true;

  try {
    // 2. 登录（只传用户名和密码，验证码已在前端验证）
    await userStore.login(form.username, form.password);

    saveLoginInfo();
    ElMessage.success("登录成功");
    setTimeout(() => {
      window.location.href = "/";
    }, 150);
  } catch (error: any) {
    console.error("Login error:", error);

    // 统一刷新验证码
    refreshCaptcha();
    form.captcha = "";

    const message = error?.message || "登录失败，请稍后重试";
    if (message.includes("用户名或密码错误")) {
      ElMessage.error("用户名或密码错误");
      form.password = "";
    } else if (message.includes("账号已被禁用")) {
      ElMessage.error("账号已被禁用，请联系管理员");
    } else if (
      error.code === "ECONNABORTED" ||
      message.includes("timeout") ||
      message.includes("请求超时")
    ) {
      ElMessage.error("网络超时，请检查网络连接");
    } else {
      ElMessage.error(message);
    }
  } finally {
    loading.value = false;
  }
};

const handleForgotPassword = () => {
  ElMessage.info("请联系管理员重置密码");
};

// 卡片加载动画
setTimeout(() => {
  cardLoaded.value = true;
}, 100);

onMounted(() => {
  generateCaptcha();
  loadRememberedUser();

  // 延迟聚焦，确保DOM已渲染
  setTimeout(() => {
    if (usernameInput.value) {
      usernameInput.value.focus();
    }
  }, 200);

  const stopBgAnimation = initBackgroundAnimation();
  stopParticlesCleanup = initParticles();

  onBeforeUnmount(() => {
    stopBgAnimation();
    if (stopParticlesCleanup) {
      stopParticlesCleanup();
    }
  });
});

// 清理定时器
onBeforeUnmount(() => {
  if (animationId) {
    cancelAnimationFrame(animationId);
  }
  if (particleAnimationId) {
    cancelAnimationFrame(particleAnimationId);
  }
});
</script>

<style scoped>
.login-container {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  overflow: hidden;
  background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
}

/* 背景动画形状 */
.bg-animation {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 0;
}

.bg-shape {
  position: absolute;
  border-radius: 50%;
  opacity: 0.1;
  animation: float 20s infinite ease-in-out;
}

.shape-1 {
  width: 300px;
  height: 300px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  top: -150px;
  left: -150px;
  animation-delay: 0s;
}

.shape-2 {
  width: 400px;
  height: 400px;
  background: linear-gradient(135deg, #f093fb, #f5576c);
  bottom: -200px;
  right: -200px;
  animation-delay: -5s;
}

.shape-3 {
  width: 200px;
  height: 200px;
  background: linear-gradient(135deg, #4facfe, #00f2fe);
  top: 50%;
  left: -100px;
  animation-delay: -10s;
}

.shape-4 {
  width: 250px;
  height: 250px;
  background: linear-gradient(135deg, #43e97b, #38f9d7);
  bottom: 30%;
  right: -125px;
  animation-delay: -15s;
}

@keyframes float {
  0%,
  100% {
    transform: translate(0, 0) scale(1);
  }
  50% {
    transform: translate(30px, -30px) scale(1.1);
  }
}

/* 粒子效果容器 */
.particles {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

/* 登录包装器 */
.login-wrapper {
  position: relative;
  z-index: 10;
  width: 100%;
  display: flex;
  justify-content: center;
  padding: 20px;
}

/* 登录卡片 */
.login-card {
  width: 460px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 24px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  overflow: hidden;
  transform: translateY(20px);
  opacity: 0;
  transition: all 0.6s cubic-bezier(0.4, 0, 0.2, 1);
}

.login-card.card-loaded {
  transform: translateY(0);
  opacity: 1;
}

/* Logo区域 */
.logo-section {
  text-align: center;
  padding: 40px 40px 20px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: white;
}

.logo-icon {
  width: 80px;
  height: 80px;
  margin: 0 auto 20px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0%,
  100% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(255, 255, 255, 0.4);
  }
  50% {
    transform: scale(1.05);
    box-shadow: 0 0 0 20px rgba(255, 255, 255, 0);
  }
}

.system-title {
  font-size: 28px;
  font-weight: 700;
  margin: 0 0 8px;
  letter-spacing: 2px;
}

.system-subtitle {
  font-size: 14px;
  opacity: 0.9;
  margin: 0;
}

/* 表单区域 */
.form-section {
  padding: 32px 40px 40px;
}

.custom-input :deep(.el-input__wrapper) {
  border-radius: 12px;
  transition: all 0.3s;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.custom-input :deep(.el-input__wrapper:hover) {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

.custom-input :deep(.el-input__wrapper.is-focus) {
  border-color: #667eea;
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
}

/* 验证码容器 */
.captcha-container {
  display: flex;
  gap: 12px;
  align-items: center;
}

.captcha-input {
  flex: 1;
}

.captcha-input :deep(.el-input__wrapper) {
  border-radius: 12px;
}

.captcha-code {
  cursor: pointer;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 8px;
  background: #f5f7fa;
  transition: all 0.3s;
}

.captcha-code:focus-visible {
  outline: 2px solid #667eea;
  outline-offset: 2px;
}

.captcha-code canvas {
  display: block;
  border-radius: 8px;
}

.refresh-icon {
  margin-right: 8px;
  transition: transform 0.3s;
}

.captcha-code:hover .refresh-icon {
  transform: rotate(180deg);
}

/* 表单选项 */
.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.form-options :deep(.el-checkbox) {
  color: #666;
}

.form-options :deep(.el-link) {
  font-size: 14px;
}

/* 登录按钮 */
.login-btn {
  width: 100%;
  height: 48px;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 600;
  background: linear-gradient(135deg, #667eea, #764ba2);
  border: none;
  transition: all 0.3s;
}

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
}

/* 底部信息 */
.login-footer {
  text-align: center;
  margin-top: 32px;
  padding-top: 20px;
  border-top: 1px solid #eee;
  font-size: 12px;
  color: #999;
}

.login-footer .version {
  margin-left: 8px;
  padding-left: 8px;
  border-left: 1px solid #ddd;
}

/* 响应式 */
@media (max-width: 576px) {
  .login-card {
    width: 95%;
  }

  .logo-section {
    padding: 30px 20px 15px;
  }

  .form-section {
    padding: 24px 24px 32px;
  }

  .captcha-container {
    flex-direction: column;
  }

  .captcha-code {
    width: 100%;
    justify-content: center;
  }
}

/* 禁用状态样式 */
.custom-input :deep(.el-input__wrapper.is-disabled) {
  background-color: #f5f7fa;
  opacity: 0.7;
}

.login-btn :deep(.el-button.is-disabled) {
  opacity: 0.6;
}
</style>
