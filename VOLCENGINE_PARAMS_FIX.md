# 火山引擎即梦AI参数修正

## 🔧 **参数修正说明**

根据火山引擎官方文档，对即梦AI的请求参数进行了重要修正。

## 📋 **主要修正内容**

### 1. **服务标识修正**
```go
// 修正前（错误）
"req_key": "high_aes"

// 修正后（正确）
"req_key": "jimeng_high_aes_general_v21_L"
```

### 2. **图像尺寸范围修正**
```go
// 修正前（超出范围）
Width:  1024,  // 默认1024
Height: 1024,  // 默认1024
// 支持范围：512x512, 768x768, 1024x1024, 1024x768, 768x1024, 1536x1024, 1024x1536

// 修正后（符合官方范围）
Width:  512,   // 默认512
Height: 512,   // 默认512
// 支持范围：[256, 768]，具体尺寸：256x256, 512x512, 768x768, 512x768, 768x512
```

### 3. **移除不支持的参数**
```go
// 移除的参数（即梦AI不支持）
Scale:            7.5,  // 引导系数
Steps:            25,   // 推理步数
Model:            "jimeng-1.4", // 模型名称
UseStyleTransfer: bool, // 风格迁移
```

### 4. **添加官方支持的参数**
```go
// 新增的官方参数
UsePreLLM: true,  // 开启文本扩写，默认true
UseSR:     true,  // 开启AIGC超分，默认true
ReturnURL: true,  // 返回图片链接，默认true
Seed:      -1,    // 随机种子，默认-1（随机）
```

## 📊 **参数对照表**

| 参数名 | 类型 | 必填 | 默认值 | 取值范围 | 说明 |
|--------|------|------|--------|----------|------|
| req_key | string | 是 | jimeng_high_aes_general_v21_L | 固定值 | 服务标识 |
| prompt | string | 是 | - | - | 生成图像的提示词 |
| width | int | 否 | 512 | [256, 768] | 生成图像的宽度 |
| height | int | 否 | 512 | [256, 768] | 生成图像的高度 |
| seed | int | 否 | -1 | - | 随机种子，-1为随机 |
| use_pre_llm | bool | 否 | true | - | 开启文本扩写 |
| use_sr | bool | 否 | true | - | 开启AIGC超分 |
| return_url | bool | 否 | true | - | 返回图片链接 |
| logo_info | LogoInfo | 否 | - | - | 水印信息 |

## 🎯 **支持的图像尺寸**

### 正方形尺寸
- `256x256` - 最小尺寸
- `512x512` - 默认尺寸
- `768x768` - 最大正方形

### 矩形尺寸
- `512x768` - 竖向矩形
- `768x512` - 横向矩形

## 🔄 **更新的请求结构**

```go
type JimengImageRequest struct {
    Prompt     string `json:"prompt"`               // 必填：文本描述
    Width      int    `json:"width,omitempty"`      // 图像宽度，默认512，范围[256, 768]
    Height     int    `json:"height,omitempty"`     // 图像高度，默认512，范围[256, 768]
    Seed       int64  `json:"seed,omitempty"`       // 随机种子，默认-1（随机）
    UsePreLLM  bool   `json:"use_pre_llm"`          // 开启文本扩写，默认true
    UseSR      bool   `json:"use_sr"`               // 开启AIGC超分，默认true
    ReturnURL  bool   `json:"return_url"`           // 返回图片链接，默认true
    LogoInfo   string `json:"logo_info,omitempty"`  // 水印信息
}
```

## 📝 **API使用示例**

### 创建图像生成任务（修正后）
```bash
curl -X POST http://localhost:8080/api/v1/ai/image/task \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "一只可爱的小猫咪在花园里玩耍",
    "user_id": "6832b85e6e61e12084cea725",
    "size": "512x512",
    "provider": "volcengine_jimeng"
  }'
```

### 支持的尺寸选项
```json
{
  "size": "256x256"  // 最小尺寸
}
{
  "size": "512x512"  // 默认尺寸
}
{
  "size": "768x768"  // 最大正方形
}
{
  "size": "512x768"  // 竖向矩形
}
{
  "size": "768x512"  // 横向矩形
}
```

## ⚠️ **重要说明**

1. **尺寸限制**：图像宽高必须在[256, 768]范围内
2. **文本扩写**：短提示词建议开启`use_pre_llm`，长提示词建议关闭
3. **超分功能**：`use_sr=true`会启用文生图+AIGC超分
4. **链接有效期**：返回的图片链接有效期为24小时
5. **随机种子**：相同种子和参数会生成相似图片

## ✅ **修正验证**

- ✅ 服务标识使用正确的`jimeng_high_aes_general_v21_L`
- ✅ 图像尺寸限制在官方范围[256, 768]内
- ✅ 移除了不支持的参数（scale、steps、model等）
- ✅ 添加了官方支持的参数（use_pre_llm、use_sr、return_url）
- ✅ 编译成功，无语法错误

---

**修正完成时间**：2025年1月
**参考文档**：火山引擎即梦AI官方文档
**兼容性**：API接口保持不变，内部参数已修正 