# API 参数校验示例

## 创建用户接口校验

### 正确的请求格式
```json
POST /api/v1/users
Content-Type: application/json

{
  "email": "user@example.com",
  "name": "张三"
}
```

### 各种错误情况及返回

#### 1. 缺少必填字段
**请求:**
```json
{
  "email": "user@example.com"
}
```

**返回:**
```json
{
  "error": "请求参数验证失败",
  "message": "请检查以下字段",
  "details": [
    {
      "field": "name",
      "message": "姓名 是必填字段"
    }
  ]
}
```

#### 2. 邮箱格式错误
**请求:**
```json
{
  "email": "invalid-email",
  "name": "张三"
}
```

**返回:**
```json
{
  "error": "请求参数验证失败",
  "message": "请检查以下字段",
  "details": [
    {
      "field": "email",
      "message": "邮箱 格式不正确",
      "value": "invalid-email"
    }
  ]
}
```

#### 3. 姓名长度不符合要求
**请求:**
```json
{
  "email": "user@example.com",
  "name": "a"
}
```

**返回:**
```json
{
  "error": "请求参数验证失败",
  "message": "请检查以下字段",
  "details": [
    {
      "field": "name",
      "message": "姓名 长度不能少于 2 个字符",
      "value": "a"
    }
  ]
}
```

#### 4. 多个字段同时错误
**请求:**
```json
{
  "email": "invalid-email",
  "name": "这是一个非常非常非常非常非常非常非常非常非常非常长的名字超过了50个字符的限制"
}
```

**返回:**
```json
{
  "error": "请求参数验证失败",
  "message": "请检查以下字段",
  "details": [
    {
      "field": "email",
      "message": "邮箱 格式不正确",
      "value": "invalid-email"
    },
    {
      "field": "name",
      "message": "姓名 长度不能超过 50 个字符",
      "value": "这是一个非常非常非常非常非常非常非常非常非常非常长的名字超过了50个字符的限制"
    }
  ]
}
```

#### 5. JSON 格式错误
**请求:**
```
{
  "email": "user@example.com"
  "name": "张三"  // 缺少逗号
}
```

**返回:**
```json
{
  "error": "请求参数验证失败",
  "message": "请检查以下字段",
  "details": [
    {
      "field": "request_body",
      "message": "请求体格式错误，请确保发送有效的JSON"
    }
  ]
}
```

## 更新用户接口校验

### 正确的请求格式（字段可选）
```json
PUT /api/v1/users/123
Content-Type: application/json

{
  "email": "newemail@example.com"
}
```

或

```json
{
  "name": "新姓名"
}
```

或

```json
{
  "email": "newemail@example.com",
  "name": "新姓名"
}
```

### 错误情况
即使字段是可选的，如果提供了值，仍然会进行格式校验：

**请求:**
```json
{
  "email": "invalid-email"
}
```

**返回:**
```json
{
  "error": "请求参数验证失败",
  "message": "请检查以下字段",
  "details": [
    {
      "field": "email",
      "message": "邮箱 格式不正确",
      "value": "invalid-email"
    }
  ]
}
```

## 校验规则说明

### CreateUserRequest
- `email`: 必填，必须是有效的邮箱格式，最大长度100字符
- `name`: 必填，最小长度2字符，最大长度50字符

### UpdateUserRequest  
- `email`: 可选，如果提供必须是有效的邮箱格式，最大长度100字符
- `name`: 可选，如果提供最小长度2字符，最大长度50字符 