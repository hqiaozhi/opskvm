## 调整目标
根据ch9329版本的实现，调整otgm版本的键盘鼠标控制，使其适配前端控制，并去掉水平滚轮支持。

## 主要差异分析
1. **鼠标报告格式**：ch9329版本不支持水平滚轮，otgm版本支持
2. **事件处理方法**：ch9329版本有ProcessButton、ProcessMove、ProcessRelativeMove、ProcessWheel等方法，otgm版本没有
3. **键盘事件处理**：ch9329版本有ProcessKey方法，otgm版本没有
4. **HID设备接口**：两个版本的接口定义不完全一致

## 调整步骤
1. **调整OTGKMHIDControl结构体**：
   - 添加mouseButtons、mouseX、mouseY、mouseDeltaX、mouseDeltaY、mouseWheel字段
   - 添加modifiers、activeKeys字段

2. **调整HIDDevice接口**：
   - 修改SendMouseReport方法签名，去掉wheelX参数
   - 添加ProcessButton、ProcessMove、ProcessRelativeMove、ProcessWheel方法
   - 添加ProcessKey方法

3. **修改鼠标报告相关方法**：
   - 修改sendMouseReportInternal方法，去掉wheelX参数
   - 修改SendRelativeMouseReport方法，去掉wheelX参数和水平滚轮字节
   - 修改SendAbsoluteMouseReport方法，去掉wheelX参数和水平滚轮字节

4. **添加事件处理方法**：
   - 添加ProcessButton方法，处理鼠标按键事件
   - 添加ProcessMove方法，处理鼠标绝对移动事件
   - 添加ProcessRelativeMove方法，处理鼠标相对移动事件
   - 添加ProcessWheel方法，处理鼠标滚轮事件
   - 添加ProcessKey方法，处理键盘按键事件

5. **调整鼠标报告格式**：
   - 相对鼠标报告格式：去掉水平滚轮字节（从5字节改为4字节）
   - 绝对鼠标报告格式：去掉水平滚轮字节（从7字节改为6字节）

6. **确保与ch9329版本保持一致**：
   - 方法签名和功能与ch9329版本保持一致
   - 事件处理逻辑与ch9329版本保持一致
   - 报告格式与ch9329版本保持一致

## 预期效果
调整后的otgm版本将能够适配前端控制，支持键盘和鼠标事件处理，并且去掉了水平滚轮的相关支持，与ch9329版本的实现保持一致。