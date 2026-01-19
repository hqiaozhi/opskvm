# One-KVM 前端API文档

## 1. WebSocket 接口

### 1.1 连接信息
- **URL**: `ws://{host}:{port}/ws`
- **协议**: WebSocket

### 1.2 消息格式

所有WebSocket消息都使用JSON格式，包含以下字段：

```json
{
  "type": <消息类型>,
  "data": <消息数据>
}
```

### 1.3 消息类型

| 类型值 | 类型名称 | 描述 |
|--------|----------|------|
| 1 | WSMessageTypeKeyboard | 键盘事件 |
| 2 | WSMessageTypeMouse | 鼠标事件 |
| 3 | WSMessageTypeMouseMode | 鼠标模式切换 |

## 2. 键盘控制

### 2.1 键盘事件消息格式

```json
{
  "type": 1,
  "data": {
    "modifier": <修饰键>,
    "keys": [<按键1>, <按键2>, ..., <按键6>]
  }
}
```

### 2.2 修饰键定义

| 修饰键 | 值 | 描述 |
|--------|-----|------|
| ModifierLeftCtrl | 0x01 | 左Ctrl键 |
| ModifierLeftShift | 0x02 | 左Shift键 |
| ModifierLeftAlt | 0x04 | 左Alt键 |
| ModifierLeftGUI | 0x08 | 左GUI键（Windows键/Command键） |
| ModifierRightCtrl | 0x10 | 右Ctrl键 |
| ModifierRightShift | 0x20 | 右Shift键 |
| ModifierRightAlt | 0x40 | 右Alt键 |
| ModifierRightGUI | 0x80 | 右GUI键（Windows键/Command键） |

### 2.3 按键映射表

#### 2.3.1 字母键

| 按键 | 值 |
|------|-----|
| KeyA | 0x04 |
| KeyB | 0x05 |
| KeyC | 0x06 |
| KeyD | 0x07 |
| KeyE | 0x08 |
| KeyF | 0x09 |
| KeyG | 0x0A |
| KeyH | 0x0B |
| KeyI | 0x0C |
| KeyJ | 0x0D |
| KeyK | 0x0E |
| KeyL | 0x0F |
| KeyM | 0x10 |
| KeyN | 0x11 |
| KeyO | 0x12 |
| KeyP | 0x13 |
| KeyQ | 0x14 |
| KeyR | 0x15 |
| KeyS | 0x16 |
| KeyT | 0x17 |
| KeyU | 0x18 |
| KeyV | 0x19 |
| KeyW | 0x1A |
| KeyX | 0x1B |
| KeyY | 0x1C |
| KeyZ | 0x1D |

#### 2.3.2 数字键

| 按键 | 值 |
|------|-----|
| Key1 | 0x1E |
| Key2 | 0x1F |
| Key3 | 0x20 |
| Key4 | 0x21 |
| Key5 | 0x22 |
| Key6 | 0x23 |
| Key7 | 0x24 |
| Key8 | 0x25 |
| Key9 | 0x26 |
| Key0 | 0x27 |

#### 2.3.3 功能键

| 按键 | 值 |
|------|-----|
| KeyF1 | 0x3A |
| KeyF2 | 0x3B |
| KeyF3 | 0x3C |
| KeyF4 | 0x3D |
| KeyF5 | 0x3E |
| KeyF6 | 0x3F |
| KeyF7 | 0x40 |
| KeyF8 | 0x41 |
| KeyF9 | 0x42 |
| KeyF10 | 0x43 |
| KeyF11 | 0x44 |
| KeyF12 | 0x45 |
| KeyF13 | 0x68 |
| KeyF14 | 0x69 |
| KeyF15 | 0x6A |
| KeyF16 | 0x6B |
| KeyF17 | 0x6C |
| KeyF18 | 0x6D |
| KeyF19 | 0x6E |
| KeyF20 | 0x6F |
| KeyF21 | 0x70 |
| KeyF22 | 0x71 |
| KeyF23 | 0x72 |
| KeyF24 | 0x73 |

#### 2.3.4 特殊键

| 按键 | 值 |
|------|-----|
| KeyEnter | 0x28 |
| KeyEscape | 0x29 |
| KeyBackspace | 0x2A |
| KeyTab | 0x2B |
| KeySpace | 0x2C |
| KeyMinus | 0x2D |
| KeyEqual | 0x2E |
| KeyLeftBrace | 0x2F |
| KeyRightBrace | 0x30 |
| KeyBackslash | 0x31 |
| KeySemicolon | 0x33 |
| KeyApostrophe | 0x34 |
| KeyGrave | 0x35 |
| KeyComma | 0x36 |
| KeyDot | 0x37 |
| KeySlash | 0x38 |
| KeyCapsLock | 0x39 |

#### 2.3.5 导航键

| 按键 | 值 |
|------|-----|
| KeyPrintScreen | 0x46 |
| KeyScrollLock | 0x47 |
| KeyPause | 0x48 |
| KeyInsert | 0x49 |
| KeyHome | 0x4A |
| KeyPageUp | 0x4B |
| KeyDelete | 0x4C |
| KeyEnd | 0x4D |
| KeyPageDown | 0x4E |
| KeyRightArrow | 0x4F |
| KeyLeftArrow | 0x50 |
| KeyDownArrow | 0x51 |
| KeyUpArrow | 0x52 |

#### 2.3.6 数字小键盘

| 按键 | 值 |
|------|-----|
| KeyNumLock | 0x53 |
| KeyKeypadSlash | 0x54 |
| KeyKeypadAsterisk | 0x55 |
| KeyKeypadMinus | 0x56 |
| KeyKeypadPlus | 0x57 |
| KeyKeypadEnter | 0x58 |
| KeyKeypad1 | 0x59 |
| KeyKeypad2 | 0x5A |
| KeyKeypad3 | 0x5B |
| KeyKeypad4 | 0x5C |
| KeyKeypad5 | 0x5D |
| KeyKeypad6 | 0x5E |
| KeyKeypad7 | 0x5F |
| KeyKeypad8 | 0x60 |
| KeyKeypad9 | 0x61 |
| KeyKeypad0 | 0x62 |
| KeyKeypadDot | 0x63 |
| KeyKeypadEqual | 0x67 |

## 3. 鼠标控制

### 3.1 鼠标事件消息格式

```json
{
  "type": 2,
  "data": {
    "buttons": <按键状态>,
    "dx": <X轴偏移>,
    "dy": <Y轴偏移>,
    "wheel": <滚轮偏移>
  }
}
```

### 3.2 鼠标按键定义

| 按键 | 值 |
|------|-----|
| MouseLeft | 0x01 |
| MouseRight | 0x02 |
| MouseMiddle | 0x04 |
| MouseBack | 0x08 |
| MouseForward | 0x10 |

### 3.3 按键状态计算

按键状态是一个字节，通过按下按键的OR运算计算得出。例如：
- 左键按下：`0x01`
- 左键+右键按下：`0x01 | 0x02 = 0x03`

### 3.4 鼠标模式切换

```json
{
  "type": 3,
  "data": {
    "absolute": <绝对模式标志>
  }
}
```

- `absolute` 为 `true` 时使用绝对鼠标模式
- `absolute` 为 `false` 时使用相对鼠标模式

## 4. 使用示例

### 4.1 发送键盘事件

发送一个Shift+A按键事件：

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  // 按下Shift+A
  ws.send(JSON.stringify({
    type: 1,
    data: {
      modifier: 0x02, // Shift键
      keys: [0x04]    // A键
    }
  }));

  // 释放所有按键
  ws.send(JSON.stringify({
    type: 1,
    data: {
      modifier: 0x00,
      keys: []
    }
  }));
};
```

### 4.2 发送鼠标事件

发送一个鼠标移动事件：

```javascript
// 鼠标移动（相对模式）
ws.send(JSON.stringify({
  type: 2,
  data: {
    buttons: 0x00,
    dx: 10,
    dy: 10,
    wheel: 0
  }
}));

// 鼠标左键点击
ws.send(JSON.stringify({
  type: 2,
  data: {
    buttons: 0x01, // 左键按下
    dx: 0,
    dy: 0,
    wheel: 0
  }
}));

ws.send(JSON.stringify({
  type: 2,
  data: {
    buttons: 0x00, // 左键释放
    dx: 0,
    dy: 0,
    wheel: 0
  }
}));
```

### 4.3 切换鼠标模式

```javascript
// 切换到绝对鼠标模式
ws.send(JSON.stringify({
  type: 3,
  data: {
    absolute: true
  }
}));
```

## 5. 注意事项

1. 键盘事件最多支持同时按下6个按键
2. 鼠标事件的dx、dy和wheel值范围为-127到127
3. 建议在按键释放后发送一个空按键事件，确保所有按键都被释放
4. 鼠标模式切换会立即生效，建议在连接初始化时设置合适的鼠标模式

## 6. 错误处理

- 连接错误：WebSocket连接失败时，浏览器会触发onerror事件
- 消息格式错误：服务器会忽略格式错误的消息
- 无效按键：服务器会忽略无效的按键值

## 7. 浏览器兼容性

- Chrome 4+ 
- Firefox 4+ 
- Safari 5+ 
- Edge 12+ 

## 8. 安全建议

1. 建议使用HTTPS/WSS协议，特别是在生产环境中
2. 实现适当的身份验证机制，防止未授权访问
3. 限制WebSocket连接的速率，防止滥用
4. 在客户端实现适当的错误处理和重连机制
