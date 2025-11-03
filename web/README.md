# RateFlow Web

Modern React 18 frontend for the RateFlow exchange rate tracking platform.

## Tech Stack

- **React 18** - UI library
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server
- **Material-UI (MUI)** - UI component library
- **TanStack Query** - Data fetching and caching
- **Recharts** - Data visualization
- **Axios** - HTTP client
- **Day.js** - Date formatting

## Getting Started

### Prerequisites

- Node.js 20+
- npm or yarn

### Installation

```bash
# Install dependencies
npm install

# Copy environment variables
cp .env.example .env
```

### Development

```bash
# Start development server (with API proxy)
npm run dev

# The app will be available at http://localhost:5173
```

### Building

```bash
# Type check
npm run type-check

# Build for production
npm run build

# Preview production build
npm run preview
```

### Linting

```bash
# Run ESLint
npm run lint
```

## Project Structure

```
src/
├── api/                # API client and hooks
│   ├── client.ts       # Axios instance and API methods
│   └── hooks.ts        # React Query hooks
├── components/         # Reusable components
│   ├── Header.tsx
│   ├── LoadingSpinner.tsx
│   └── ErrorAlert.tsx
├── features/           # Feature-based modules
│   └── Dashboard/      # Main dashboard feature
├── types/              # TypeScript type definitions
├── utils/              # Utility functions
├── theme.ts            # MUI theme configuration
├── App.tsx             # Root component
└── main.tsx            # Application entry point
```

## Features

- **Real-time Rate Display** - Shows the latest exchange rate with auto-refresh
- **Historical Charts** - Interactive line charts with multiple time ranges (7, 14, 30, 60, 90 days)
- **Data Table** - Paginated historical data with sorting
- **Responsive Design** - Works on desktop, tablet, and mobile devices
- **Error Handling** - Graceful error states with retry functionality
- **Loading States** - Skeleton loaders for better UX
- **Health Monitoring** - Backend health status indicator

## API Integration

The frontend communicates with the backend API through a proxy configuration in development:

- `/api/*` → `http://localhost:8080/api/*`
- `/health` → `http://localhost:8080/health`

For production, set `VITE_API_BASE_URL` to your API server URL.

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run type-check` - Type check without emitting files

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

MIT
# 前端开发指南

## 快速开始

### 1. 安装依赖

```bash
npm install
```

### 2. 启动开发服务器

确保后端 API 服务已经在 `http://localhost:8080` 运行，然后：

```bash
npm run dev
```

前端将在 `http://localhost:5173` 启动，所有 `/api` 请求会自动代理到后端。

### 3. 或者使用 Make 命令（从项目根目录）

```bash
# 安装前端依赖
make web-install

# 启动前端开发服务器
make web-dev

# 一键启动全栈（后端 + 前端）
make fullstack
```

## 项目结构

```
src/
├── api/                    # API 客户端和 React Query hooks
│   ├── client.ts          # Axios 实例和 API 方法
│   └── hooks.ts           # useLatestRate, useHistoricalRates 等
│
├── components/            # 通用可复用组件
│   ├── Header.tsx         # 顶部导航栏
│   ├── LoadingSpinner.tsx # 加载状态组件
│   └── ErrorAlert.tsx     # 错误提示组件
│
├── features/              # 功能模块
│   └── Dashboard/         # 主仪表板
│       ├── index.tsx                  # Dashboard 主组件
│       ├── CurrentRateCard.tsx        # 当前汇率卡片
│       ├── CurrencyPairSelector.tsx   # 货币对选择器
│       ├── RateChart.tsx              # 汇率走势图
│       └── RateHistoryTable.tsx       # 历史数据表格
│
├── types/                 # TypeScript 类型定义
│   └── index.ts           # Rate, ApiResponse 等类型
│
├── utils/                 # 工具函数
│   └── formatters.ts      # 格式化函数（日期、数字等）
│
├── theme.ts               # Material-UI 主题配置
├── App.tsx                # 根组件
└── main.tsx               # 应用入口
```

## 技术栈

- **React 18** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **Material-UI (MUI)** - UI 组件库
- **TanStack Query** - 数据获取和缓存
- **Recharts** - 图表库
- **Axios** - HTTP 客户端
- **Day.js** - 日期处理

## 主要功能

### 1. 实时汇率显示

`CurrentRateCard` 组件显示最新汇率，使用 `useLatestRate` hook 自动每 5 分钟刷新。

```typescript
const { data: rate } = useLatestRate('CNY/JPY')
```

### 2. 历史数据图表

`RateChart` 组件使用 Recharts 展示汇率走势，支持 7/14/30/60/90 天的时间范围选择。

### 3. 数据表格

`RateHistoryTable` 组件展示分页的历史数据，支持排序和分页。

### 4. 自动缓存

TanStack Query 自动缓存数据，减少不必要的 API 请求：
- 最新汇率缓存 5 分钟
- 历史数据根据查询参数缓存

## 开发规范

### 组件命名

- 使用 PascalCase：`CurrentRateCard.tsx`
- 一个文件一个组件
- 功能模块放在 `features/` 目录

### 类型定义

所有 API 返回的数据都应该有对应的 TypeScript 类型定义，放在 `types/index.ts`。

### API 调用

使用 React Query hooks 而不是直接调用 API：

```typescript
// ✅ 推荐
const { data, isLoading, error } = useLatestRate(pair)

// ❌ 不推荐
const [data, setData] = useState()
useEffect(() => {
  rateApi.getLatestRate(pair).then(setData)
}, [pair])
```

### 错误处理

使用 `ErrorAlert` 组件统一显示错误：

```typescript
if (error) {
  return <ErrorAlert message={error.message} onRetry={() => refetch()} />
}
```

### 加载状态

使用 `LoadingSpinner` 或 Material-UI 的 `Skeleton` 组件：

```typescript
if (isLoading) {
  return <LoadingSpinner message="加载中..." />
}
```

## 环境变量

在 `.env` 文件中配置：

```bash
# API 基础 URL（留空则使用代理）
VITE_API_BASE_URL=
```

## 构建部署

```bash
# 类型检查
npm run type-check

# 构建生产版本
npm run build

# 预览生产构建
npm run preview
```

构建产物在 `dist/` 目录。

## 代理配置

开发环境中，Vite 配置了代理将 API 请求转发到后端：

```typescript
// vite.config.ts
proxy: {
  '/api': {
    target: 'http://localhost:8080',
    changeOrigin: true,
  }
}
```

## 常见问题

### Q: API 请求失败？

A: 确保后端服务在 `http://localhost:8080` 运行。检查：
```bash
curl http://localhost:8080/health
```

### Q: 依赖安装失败？

A: 确保 Node.js 版本 >= 20，然后：
```bash
rm -rf node_modules package-lock.json
npm install
```

### Q: 类型错误？

A: 运行类型检查查看详细错误：
```bash
npm run type-check
```

## 调试

### React DevTools

安装浏览器扩展：
- Chrome: [React Developer Tools](https://chrome.google.com/webstore/detail/react-developer-tools/fmkadmapgofadopljbjfkapdkoienihi)
- Firefox: [React Developer Tools](https://addons.mozilla.org/en-US/firefox/addon/react-devtools/)

### React Query DevTools

已内置在开发模式，查看右下角的 React Query 图标。

### 网络请求

使用浏览器开发者工具的 Network 标签查看 API 请求。

## 性能优化

1. **代码分割**: Vite 自动分割 vendor、mui、charts 代码
2. **懒加载**: 使用 React.lazy() 延迟加载大组件
3. **缓存**: TanStack Query 自动缓存数据
4. **优化渲染**: 使用 React.memo 和 useMemo 避免不必要的重渲染

## 贡献指南

1. 创建功能分支
2. 开发新功能
3. 运行 `npm run type-check` 和 `npm run lint`
4. 提交 PR

## 参考资料

- [React 文档](https://react.dev/)
- [Vite 文档](https://vitejs.dev/)
- [Material-UI 文档](https://mui.com/)
- [TanStack Query 文档](https://tanstack.com/query/latest)
- [Recharts 文档](https://recharts.org/)
