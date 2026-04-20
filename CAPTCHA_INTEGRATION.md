# 驗證碼集成文檔 (CAPTCHA Integration)

## 後端實現 (Backend Implementation)

已在 **admin** 和 **user** 服務中添加驗證碼功能，使用 [base64Captcha](https://github.com/mojocn/base64Captcha) 庫。

### Admin 服務 API 端點 (Admin Service API Endpoints)

#### 1. 獲取驗證碼 (Get Captcha)
```
GET /api/v1/admin/captcha
```

**響應 (Response):**
```json
{
  "captcha_id": "string",
  "captcha_image": "data:image/png;base64,..."
}
```

#### 2. 管理員登錄 (Admin Login)
```
POST /api/v1/admin/login
```

**請求體 (Request Body):**
```json
{
  "username": "string",
  "password": "string",
  "captcha_id": "string",
  "captcha_code": "string"
}
```

**響應 (Response):**
```json
{
  "admin": {
    "id": 1,
    "username": "admin",
    "nickname": "管理員",
    "roles": ["admin"]
  },
  "token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

### User 服務 API 端點 (User Service API Endpoints)

#### 1. 獲取驗證碼 (Get Captcha)
```
GET /api/v1/user/captcha
```

**響應 (Response):**
```json
{
  "captcha_id": "string",
  "captcha_image": "data:image/png;base64,..."
}
```

#### 2. 用戶登錄 (User Login)
```
POST /api/v1/user/login
```

**請求體 (Request Body):**
```json
{
  "phone": "string",
  "password": "string",
  "captcha_id": "string",
  "captcha_code": "string"
}
```

**響應 (Response):**
```json
{
  "user": {
    "id": 1,
    "username": "user123",
    "phone": "1234567890",
    "invite_code": "ABC123"
  },
  "token": "eyJhbGc...",
  "refresh_token": "eyJhbGc..."
}
```

## 前端集成示例 (Frontend Integration Examples)

### React/TypeScript 示例

```typescript
import { useState, useEffect } from 'react';
import axios from 'axios';

interface CaptchaData {
  captcha_id: string;
  captcha_image: string;
}

interface LoginForm {
  username: string;
  password: string;
  captcha_code: string;
}

export const LoginPage = () => {
  const [captcha, setCaptcha] = useState<CaptchaData | null>(null);
  const [form, setForm] = useState<LoginForm>({
    username: '',
    password: '',
    captcha_code: ''
  });

  // 獲取驗證碼
  const fetchCaptcha = async () => {
    try {
      const response = await axios.get('/api/v1/admin/captcha');
      setCaptcha(response.data);
    } catch (error) {
      console.error('獲取驗證碼失敗:', error);
    }
  };

  useEffect(() => {
    fetchCaptcha();
  }, []);

  // 登錄處理
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const response = await axios.post('/api/v1/admin/login', {
        username: form.username,
        password: form.password,
        captcha_id: captcha?.captcha_id,
        captcha_code: form.captcha_code
      });
      
      // 保存 token
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
      
      // 跳轉到管理頁面
      window.location.href = '/admin/dashboard';
    } catch (error) {
      console.error('登錄失敗:', error);
      // 刷新驗證碼
      fetchCaptcha();
      setForm({ ...form, captcha_code: '' });
    }
  };

  return (
    <div className="login-container">
      <form onSubmit={handleLogin}>
        <input
          type="text"
          placeholder="用戶名"
          value={form.username}
          onChange={(e) => setForm({ ...form, username: e.target.value })}
        />
        
        <input
          type="password"
          placeholder="密碼"
          value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })}
        />
        
        <div className="captcha-container">
          {captcha && (
            <img 
              src={captcha.captcha_image} 
              alt="驗證碼"
              onClick={fetchCaptcha}
              style={{ cursor: 'pointer' }}
            />
          )}
          <input
            type="text"
            placeholder="驗證碼"
            value={form.captcha_code}
            onChange={(e) => setForm({ ...form, captcha_code: e.target.value })}
          />
        </div>
        
        <button type="submit">登錄</button>
      </form>
    </div>
  );
};
```

### Vue 3 示例

```vue
<template>
  <div class="login-container">
    <form @submit.prevent="handleLogin">
      <input
        v-model="form.username"
        type="text"
        placeholder="用戶名"
      />
      
      <input
        v-model="form.password"
        type="password"
        placeholder="密碼"
      />
      
      <div class="captcha-container">
        <img 
          v-if="captcha"
          :src="captcha.captcha_image" 
          alt="驗證碼"
          @click="fetchCaptcha"
          style="cursor: pointer"
        />
        <input
          v-model="form.captcha_code"
          type="text"
          placeholder="驗證碼"
        />
      </div>
      
      <button type="submit">登錄</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import axios from 'axios';

interface CaptchaData {
  captcha_id: string;
  captcha_image: string;
}

const captcha = ref<CaptchaData | null>(null);
const form = ref({
  username: '',
  password: '',
  captcha_code: ''
});

const fetchCaptcha = async () => {
  try {
    const response = await axios.get('/api/v1/admin/captcha');
    captcha.value = response.data;
  } catch (error) {
    console.error('獲取驗證碼失敗:', error);
  }
};

const handleLogin = async () => {
  try {
    const response = await axios.post('/api/v1/admin/login', {
      username: form.value.username,
      password: form.value.password,
      captcha_id: captcha.value?.captcha_id,
      captcha_code: form.value.captcha_code
    });
    
    localStorage.setItem('token', response.data.token);
    localStorage.setItem('refresh_token', response.data.refresh_token);
    
    window.location.href = '/admin/dashboard';
  } catch (error) {
    console.error('登錄失敗:', error);
    fetchCaptcha();
    form.value.captcha_code = '';
  }
};

onMounted(() => {
  fetchCaptcha();
});
</script>
```

### 原生 JavaScript 示例

```html
<!DOCTYPE html>
<html>
<head>
  <title>管理員登錄</title>
</head>
<body>
  <div class="login-container">
    <form id="loginForm">
      <input type="text" id="username" placeholder="用戶名" required />
      <input type="password" id="password" placeholder="密碼" required />
      
      <div class="captcha-container">
        <img id="captchaImage" alt="驗證碼" style="cursor: pointer" />
        <input type="text" id="captchaCode" placeholder="驗證碼" required />
      </div>
      
      <button type="submit">登錄</button>
    </form>
  </div>

  <script>
    let captchaId = '';

    // 獲取驗證碼
    async function fetchCaptcha() {
      try {
        const response = await fetch('/api/v1/admin/captcha');
        const data = await response.json();
        
        captchaId = data.captcha_id;
        document.getElementById('captchaImage').src = data.captcha_image;
      } catch (error) {
        console.error('獲取驗證碼失敗:', error);
      }
    }

    // 點擊驗證碼刷新
    document.getElementById('captchaImage').addEventListener('click', fetchCaptcha);

    // 登錄處理
    document.getElementById('loginForm').addEventListener('submit', async (e) => {
      e.preventDefault();
      
      const username = document.getElementById('username').value;
      const password = document.getElementById('password').value;
      const captchaCode = document.getElementById('captchaCode').value;
      
      try {
        const response = await fetch('/api/v1/admin/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            username,
            password,
            captcha_id: captchaId,
            captcha_code: captchaCode
          })
        });
        
        if (!response.ok) {
          throw new Error('登錄失敗');
        }
        
        const data = await response.json();
        
        localStorage.setItem('token', data.token);
        localStorage.setItem('refresh_token', data.refresh_token);
        
        window.location.href = '/admin/dashboard';
      } catch (error) {
        console.error('登錄失敗:', error);
        alert('登錄失敗，請檢查用戶名、密碼和驗證碼');
        fetchCaptcha();
        document.getElementById('captchaCode').value = '';
      }
    });

    // 頁面加載時獲取驗證碼
    fetchCaptcha();
  </script>
</body>
</html>
```

### 用戶登錄示例 (User Login Example)

```typescript
// React/TypeScript - User Login
import { useState, useEffect } from 'react';
import axios from 'axios';

interface CaptchaData {
  captcha_id: string;
  captcha_image: string;
}

interface UserLoginForm {
  phone: string;
  password: string;
  captcha_code: string;
}

export const UserLoginPage = () => {
  const [captcha, setCaptcha] = useState<CaptchaData | null>(null);
  const [form, setForm] = useState<UserLoginForm>({
    phone: '',
    password: '',
    captcha_code: ''
  });

  // 獲取驗證碼
  const fetchCaptcha = async () => {
    try {
      const response = await axios.get('/api/v1/user/captcha');
      setCaptcha(response.data);
    } catch (error) {
      console.error('獲取驗證碼失敗:', error);
    }
  };

  useEffect(() => {
    fetchCaptcha();
  }, []);

  // 登錄處理
  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    
    try {
      const response = await axios.post('/api/v1/user/login', {
        phone: form.phone,
        password: form.password,
        captcha_id: captcha?.captcha_id,
        captcha_code: form.captcha_code
      });
      
      // 保存 token
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
      
      // 跳轉到用戶頁面
      window.location.href = '/user/dashboard';
    } catch (error) {
      console.error('登錄失敗:', error);
      // 刷新驗證碼
      fetchCaptcha();
      setForm({ ...form, captcha_code: '' });
    }
  };

  return (
    <div className="login-container">
      <form onSubmit={handleLogin}>
        <input
          type="tel"
          placeholder="手機號碼"
          value={form.phone}
          onChange={(e) => setForm({ ...form, phone: e.target.value })}
        />
        
        <input
          type="password"
          placeholder="密碼"
          value={form.password}
          onChange={(e) => setForm({ ...form, password: e.target.value })}
        />
        
        <div className="captcha-container">
          {captcha && (
            <img 
              src={captcha.captcha_image} 
              alt="驗證碼"
              onClick={fetchCaptcha}
              style={{ cursor: 'pointer' }}
            />
          )}
          <input
            type="text"
            placeholder="驗證碼"
            value={form.captcha_code}
            onChange={(e) => setForm({ ...form, captcha_code: e.target.value })}
          />
        </div>
        
        <button type="submit">登錄</button>
      </form>
    </div>
  );
};
```

## 錯誤處理 (Error Handling)

### 常見錯誤碼

- `CAPTCHA_GENERATE_FAILED` (500): 驗證碼生成失敗
- `CAPTCHA_VERIFY_FAILED` (400): 驗證碼錯誤
- `LOGIN_FAILED` (400): 用戶名或密碼錯誤

### 建議的錯誤處理流程

1. 驗證碼錯誤時，自動刷新驗證碼
2. 登錄失敗時，清空驗證碼輸入框
3. 顯示友好的錯誤提示信息

## 安全建議 (Security Recommendations)

1. 驗證碼有效期為 10 分鐘
2. 驗證碼驗證後自動清除（clear=true）
3. 建議在登錄失敗後刷新驗證碼
4. 使用 HTTPS 傳輸敏感信息
5. Token 應存儲在 httpOnly cookie 或安全的本地存儲中

## 測試 (Testing)

### Admin 服務測試

使用 curl 測試 Admin API：

```bash
# 獲取驗證碼
curl -X GET http://localhost:8000/api/v1/admin/captcha

# 管理員登錄
curl -X POST http://localhost:8000/api/v1/admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password",
    "captcha_id": "your_captcha_id",
    "captcha_code": "12345"
  }'
```

### User 服務測試

使用 curl 測試 User API：

```bash
# 獲取驗證碼
curl -X GET http://localhost:8001/api/v1/user/captcha

# 用戶登錄
curl -X POST http://localhost:8001/api/v1/user/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "1234567890",
    "password": "password",
    "captcha_id": "your_captcha_id",
    "captcha_code": "12345"
  }'
```
