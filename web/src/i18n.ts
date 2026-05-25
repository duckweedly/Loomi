import type { RuntimeEventGroup } from './domain'
import type { BackendCapabilityStatus } from './runtime/backendCapabilityStatus'

export type Locale = 'zh' | 'en'

type Dictionary = {
  app: {
    chat: string
    work: string
    collapseSidebar: string
    openSidebar: string
    search: string
    openRunDetails: string
    openRightTools: string
  }
  sidebar: {
    newChat: string
    projects: string
    scheduled: string
    threads: string
    settings: string
    theme: string
    update: string
    current: string
    open: string
    light: string
    dark: string
    archiveThread: string
    renameThread: string
    loadingThreads: string
    retry: string
    emptyThreads: (mode: string) => string
  }
  runtime: {
    eventGroups: Record<RuntimeEventGroup, string>
    workerJob: {
      queued: string
      claimed: string
      running: string
      retrying: string
      recovered: string
      exhausted: string
      cancelled: string
      failed: string
      jobClaimed: string
      leaseRenewed: string
      jobRecovering: string
      retryScheduled: string
      attemptFailed: string
      retryExhausted: string
      cancellationRequested: string
      diagnostics: string
      unknownWorkerEvent: string
      currentRunJob: string
      noTaskRunning: string
      runRealMessage: string
      readOnlyObserver: string
      readOnlyNoControls: string
      latestEvents: string
      noEventsYet: string
      statusQueued: string
      statusLeased: string
      statusRetrying: string
      statusRecovering: string
      statusCompleted: string
      statusFailed: string
      statusCancelled: string
      statusDead: string
    }
  }
  backendCapability: Record<BackendCapabilityStatus, { title: string; detail: string }>
  chatCanvas: {
    noThreadTitle: string
    noThreadDetail: string
    emptyThreadTitle: string
    emptyThreadDetail: string
    loadingTitle: string
    loadingDetail: string
    errorTitle: string
    errorDetail: string
    waitingRunTitle: string
    waitingRunDetail: string
    runningTitle: string
    runningDetail: string
    completedTitle: string
    completedDetail: string
    failedTitle: string
    failedDetail: string
    stoppedTitle: string
    stoppedDetail: string
    recoveringTitle: string
    recoveringDetail: string
    stoppingTitle: string
    stoppingDetail: string
    backendUnavailableTitle: string
    backendUnavailableDetail: string
    assistant: string
    user: string
    modelGateway: string
    toolBoundaryNotice: string
    context: string
    localSimulated: string
    stop: string
    retry: string
    regenerate: string
    attach: string
    messageLoomi: string
    providerUnavailableWarning: string
    openProviderSettings: string
    stoppedDraft: string
    recoveringDraft: string
    modelDrafting: string
    generating: string
  }
  settings: {
    title: string
    back: string
    language: string
    languageHelper: string
    chinese: string
    english: string
    working: string
    readOnly: string
    previewOnly: string
    disabled: string
    mixed: string
    available: string
    unavailable: string
    misconfigured: string
    notConnected: string
    noThreadSelected: string
    noActiveRun: string
    general: string
    generalDescription: string
    workspaceDefaults: string
    workspaceDefaultsDescription: string
    defaultWorkspaceMode: string
    defaultWorkspaceModeHelper: string
    mockRuntimeScenario: string
    mockRuntimeScenarioHelper: string
    success: string
    failure: string
    runtimeStatus: string
    runtimeStatusDescription: string
    dataSourceMode: string
    dataSourceModeHelper: string
    backendCapability: string
    backendCapabilityHelper: string
    streamState: string
    streamStateHelper: string
    selectedThread: string
    selectedThreadHelper: string
    selectedRunStatus: string
    selectedRunStatusHelper: string
    providerCapability: string
    providerCapabilityHelper: string
    providerConsoleTitle: string
    providerConsoleDescription: string
    providerConfiguredProviders: string
    providerConfiguredProvidersHelper: string
    providerConsoleEmpty: string
    providerConsoleEnvGuide: string
    providerTestConnection: string
    providerChecking: string
    providerCheckResult: (status: string, message?: string) => string
    providerLocalDraftTitle: string
    providerLocalDraftDescription: string
    providerBaseUrl: string
    providerBaseUrlHelper: string
    providerModel: string
    providerModelHelper: string
    providerApiKey: string
    providerApiKeyHelper: string
    providerConfigured: string
    providerNotConfigured: string
    aboutLocalApp: string
    aboutLocalAppDescription: string
    appName: string
    appNameHelper: string
    appVersion: string
    appVersionHelper: string
    appStatus: string
    appStatusHelper: string
    previewControl: string
    previewControlHelper: string
    connectionState: string
    connectionStateHelper: string
    categoryPreview: (label: string) => string
  }
}

export const dictionaries: Record<Locale, Dictionary> = {
  zh: {
    app: {
      chat: '聊天',
      work: '工作',
      collapseSidebar: '收起侧边栏',
      openSidebar: '打开侧边栏',
      search: '搜索',
      openRunDetails: '打开运行详情',
      openRightTools: '打开右侧工具',
    },
    sidebar: {
      newChat: '新对话',
      projects: '项目',
      scheduled: '计划任务',
      threads: '会话',
      settings: '设置',
      theme: '主题',
      update: '更新',
      current: '当前',
      open: '打开',
      light: '浅色',
      dark: '深色',
      archiveThread: '归档会话',
      renameThread: '重命名会话',
      loadingThreads: '加载会话中',
      retry: '重试',
      emptyThreads: (mode) => `暂无 ${mode} 会话`,
    },
    runtime: {
      eventGroups: {
        'run-lifecycle': '运行生命周期',
        'model-stream': '模型流',
        'worker-job': 'Worker/Job',
        'tool-call': '工具调用',
        error: '错误',
      },
      workerJob: {
        queued: '已排队',
        claimed: '已领取',
        running: '运行中',
        retrying: '重试中',
        recovered: '已恢复',
        exhausted: '已耗尽',
        cancelled: '已取消',
        failed: '失败',
        jobClaimed: 'Worker 已领取任务',
        leaseRenewed: 'Lease 已续期',
        jobRecovering: '任务恢复中',
        retryScheduled: '已安排重试',
        attemptFailed: '本次尝试失败',
        retryExhausted: '重试已耗尽',
        cancellationRequested: '已请求取消',
        diagnostics: 'Worker 诊断',
        unknownWorkerEvent: '未知 Worker 事件',
        currentRunJob: '当前 Run Job',
        noTaskRunning: '当前没有后台任务',
        runRealMessage: '发送一条真实模型消息来观察排队任务和 Worker 事件。',
        readOnlyObserver: '只读观察面板',
        readOnlyNoControls: '只读 · 不提供重试、恢复或取消控件',
        latestEvents: '最新 Worker/Job 事件',
        noEventsYet: '暂无 Worker/Job 事件',
        statusQueued: '排队中',
        statusLeased: 'Worker 已领取',
        statusRetrying: '重试中',
        statusRecovering: '恢复中',
        statusCompleted: '已完成',
        statusFailed: '失败',
        statusCancelled: '已取消',
        statusDead: '已终止',
      },
    },
    backendCapability: {
      mock: { title: 'Mock', detail: '确定性的本地行为；不是真实模型输出。' },
      'local-simulated': { title: '本地模拟', detail: 'Real API 路径已连接，但生成仍是模拟。' },
      'model-gateway': { title: '模型网关', detail: '真实 provider 执行可用。' },
      'backend-unavailable': { title: '后端不可用', detail: '后端无法提供 runtime 执行能力。' },
      'model-setup-missing': { title: '模型设置缺失', detail: '生成前需要配置模型设置或凭证。' },
      'provider-unavailable': { title: 'Provider 不可用', detail: 'Provider 拒绝或未能完成生成。' },
      'stream-disconnected': { title: '流已断开', detail: '事件流在终态确认前断开。' },
      'run-recovering': { title: '运行恢复中', detail: '界面正在恢复最近已知的运行状态。' },
    },
    chatCanvas: {
      noThreadTitle: '未选择会话',
      noThreadDetail: '创建新对话',
      emptyThreadTitle: '新对话',
      emptyThreadDetail: '输入第一条消息',
      loadingTitle: '加载中',
      loadingDetail: '同步会话',
      errorTitle: '加载失败',
      errorDetail: '重试',
      waitingRunTitle: '等待执行',
      waitingRunDetail: '消息已发送',
      runningTitle: '执行中',
      runningDetail: '查看右侧时间线',
      completedTitle: '已完成',
      completedDetail: '回复已生成',
      failedTitle: '执行失败',
      failedDetail: '未生成成功回复',
      stoppedTitle: '已停止',
      stoppedDetail: '保留已生成内容',
      recoveringTitle: '恢复中',
      recoveringDetail: '正在恢复运行状态',
      stoppingTitle: '停止中',
      stoppingDetail: '等待后台 worker 确认',
      backendUnavailableTitle: '后端能力未接入',
      backendUnavailableDetail: '等待 M4/M5 run/event',
      assistant: 'Loomi',
      user: '你',
      modelGateway: '模型网关',
      toolBoundaryNotice: '工具调用未执行：M5 只记录边界事件，不执行外部动作。',
      context: '上下文',
      localSimulated: '本地模拟',
      stop: '停止',
      retry: '重试',
      regenerate: '重新生成',
      attach: '附件',
      messageLoomi: '给 Loomi 发消息',
      providerUnavailableWarning: '模型 Provider 未配置或不可用',
      openProviderSettings: '打开设置',
      stoppedDraft: '已停止生成，保留已生成内容',
      recoveringDraft: '恢复中…',
      modelDrafting: '模型正在生成回复',
      generating: '生成中',
    },
    settings: {
      title: '设置',
      back: '返回工作区',
      language: '语言',
      languageHelper: '当前会话内切换界面语言。默认中文，不写入持久化设置。',
      chinese: '中文',
      english: 'English',
      working: '可用',
      readOnly: '只读',
      previewOnly: '预览',
      disabled: '禁用',
      mixed: '混合',
      available: '可用',
      unavailable: '不可用',
      misconfigured: '配置错误',
      notConnected: '未连接',
      noThreadSelected: '未选择会话',
      noActiveRun: '无活动运行',
      general: '通用',
      generalDescription: '当前会话的工作区默认值和运行状态可见性。',
      workspaceDefaults: '工作区默认值',
      workspaceDefaultsDescription: '只影响后续本地工作区动作的当前会话偏好。',
      defaultWorkspaceMode: '默认工作区模式',
      defaultWorkspaceModeHelper: '影响之后从侧边栏创建的本地会话。',
      mockRuntimeScenario: 'Mock 运行场景',
      mockRuntimeScenarioHelper: '只影响之后的 mock 发送，不修改正在运行的任务。',
      success: '成功',
      failure: '失败',
      runtimeStatus: '运行状态',
      runtimeStatusDescription: '当前工作区运行状态的只读视图。',
      dataSourceMode: '数据源模式',
      dataSourceModeHelper: '显示前端正在使用 mock 数据还是本地 API。',
      backendCapability: '后端能力',
      backendCapabilityHelper: '显示运行能力可用性，不暴露凭证。',
      streamState: '流状态',
      streamStateHelper: '显示当前运行事件流状态。',
      selectedThread: '当前会话',
      selectedThreadHelper: '显示打开设置时保留的工作区上下文。',
      selectedRunStatus: '当前运行状态',
      selectedRunStatusHelper: '显示当前运行状态，不修改运行。',
      providerCapability: 'Provider 能力',
      providerCapabilityHelper: '可用时只显示已脱敏的 provider id、family、model 和 status。',
      providerConsoleTitle: 'Provider Test Console',
      providerConsoleDescription: '读取后端已配置的 provider，并可安全触发一次连接测试。',
      providerConfiguredProviders: '已配置 Provider',
      providerConfiguredProvidersHelper: '这些 provider 来自后端环境变量；页面只显示脱敏能力，不显示密钥。',
      providerConsoleEmpty: '暂无已配置 provider',
      providerConsoleEnvGuide: '在本地 API 环境变量中配置 provider 后重启后端，再刷新 Settings。',
      providerTestConnection: '测试连接',
      providerChecking: '测试中',
      providerCheckResult: (status, message) => `${status === 'success' ? '成功' : status === 'failed' ? '失败' : '测试中'}${message ? ` · ${message}` : ''}`,
      providerLocalDraftTitle: '本地草稿',
      providerLocalDraftDescription: 'Base URL、模型和 API Key 输入只保存在当前浏览器会话，用于记录草稿；不会保存，也不会改变真实模型调用。',
      providerBaseUrl: 'Base URL',
      providerBaseUrlHelper: '填写 OpenAI-compatible 中转站地址，仅作为本地草稿。',
      providerModel: '模型 ID',
      providerModelHelper: '填写模型名称，仅作为本地草稿，不影响后端调用。',
      providerApiKey: 'API Key',
      providerApiKeyHelper: '仅显示是否已填写，不回显密钥内容，也不写入后端。',
      providerConfigured: '已填写',
      providerNotConfigured: '未填写',
      aboutLocalApp: '本地应用状态',
      aboutLocalAppDescription: '显示已知本地状态；构建和账号信息仍是占位。',
      appName: '应用',
      appNameHelper: '当前运行的 Loomi web shell。',
      appVersion: '版本',
      appVersionHelper: 'M5.5 暂不读取构建版本。',
      appStatus: '状态',
      appStatusHelper: '当前数据源和后端能力的组合状态。',
      previewControl: '预览控件',
      previewControlHelper: 'Mock only。该控件未连接 provider、工具、文件或后端写入。',
      connectionState: '连接状态',
      connectionStateHelper: 'M5.5 未连接。未来设置需要单独实现计划。',
      categoryPreview: (label) => `${label} 预览`,
    },
  },
  en: {
    app: {
      chat: 'Chat',
      work: 'Work',
      collapseSidebar: 'Collapse sidebar',
      openSidebar: 'Open sidebar',
      search: 'Search',
      openRunDetails: 'Open run details',
      openRightTools: 'Open right tools',
    },
    sidebar: {
      newChat: 'New Chat',
      projects: 'Projects',
      scheduled: 'Scheduled',
      threads: 'Threads',
      settings: 'Settings',
      theme: 'Theme',
      update: 'Update',
      current: 'Current',
      open: 'Open',
      light: 'Light',
      dark: 'Dark',
      archiveThread: 'Archive thread',
      renameThread: 'Rename thread',
      loadingThreads: 'Loading threads',
      retry: 'Retry',
      emptyThreads: (mode) => `No ${mode} threads`,
    },
    runtime: {
      eventGroups: {
        'run-lifecycle': 'Run lifecycle',
        'model-stream': 'Model stream',
        'worker-job': 'Worker/job',
        'tool-call': 'Tool call',
        error: 'Error',
      },
      workerJob: {
        queued: 'Queued',
        claimed: 'Claimed',
        running: 'Running',
        retrying: 'Retrying',
        recovered: 'Recovered',
        exhausted: 'Exhausted',
        cancelled: 'Cancelled',
        failed: 'Failed',
        jobClaimed: 'Job claimed by worker',
        leaseRenewed: 'Lease renewed',
        jobRecovering: 'Job recovering',
        retryScheduled: 'Retry scheduled',
        attemptFailed: 'Job attempt failed',
        retryExhausted: 'Retries exhausted',
        cancellationRequested: 'Cancellation requested',
        diagnostics: 'Worker diagnostics',
        unknownWorkerEvent: 'Unknown worker event',
        currentRunJob: 'Current run job',
        noTaskRunning: 'No background task is running',
        runRealMessage: 'Run a real model message to observe queued jobs and worker events.',
        readOnlyObserver: 'Read-only observer',
        readOnlyNoControls: 'Read-only · no retry, recover, or cancel controls',
        latestEvents: 'Latest worker/job events',
        noEventsYet: 'No worker/job events yet',
        statusQueued: 'Queued',
        statusLeased: 'Leased by worker',
        statusRetrying: 'Retrying',
        statusRecovering: 'Recovering',
        statusCompleted: 'Completed',
        statusFailed: 'Failed',
        statusCancelled: 'Cancelled',
        statusDead: 'Dead',
      },
    },
    backendCapability: {
      mock: { title: 'Mock', detail: 'Deterministic local behavior; not real model output.' },
      'local-simulated': { title: 'Local simulated', detail: 'Real API path is connected, but generation is simulated.' },
      'model-gateway': { title: 'Model gateway', detail: 'Real provider execution is available.' },
      'backend-unavailable': { title: 'Backend unavailable', detail: 'The backend cannot provide runtime execution.' },
      'model-setup-missing': { title: 'Model setup missing', detail: 'Model setup or credentials are required before generation.' },
      'provider-unavailable': { title: 'Provider unavailable', detail: 'The provider rejected or failed generation.' },
      'stream-disconnected': { title: 'Stream disconnected', detail: 'The event stream disconnected before terminal reconciliation.' },
      'run-recovering': { title: 'Run recovering', detail: 'The UI is recovering the latest known run state.' },
    },
    chatCanvas: {
      noThreadTitle: 'No thread selected',
      noThreadDetail: 'Create a new conversation',
      emptyThreadTitle: 'New conversation',
      emptyThreadDetail: 'Send the first message',
      loadingTitle: 'Loading',
      loadingDetail: 'Syncing workspace',
      errorTitle: 'Load failed',
      errorDetail: 'Retry',
      waitingRunTitle: 'Waiting to run',
      waitingRunDetail: 'Message sent',
      runningTitle: 'Running',
      runningDetail: 'View the timeline',
      completedTitle: 'Completed',
      completedDetail: 'Reply generated',
      failedTitle: 'Run failed',
      failedDetail: 'No successful reply generated',
      stoppedTitle: 'Stopped',
      stoppedDetail: 'Generated content was preserved',
      recoveringTitle: 'Recovering',
      recoveringDetail: 'Recovering the latest run state',
      stoppingTitle: 'Stopping',
      stoppingDetail: 'Waiting for the background worker to confirm',
      backendUnavailableTitle: 'Backend capability unavailable',
      backendUnavailableDetail: 'Waiting for M4/M5 run/event',
      assistant: 'Loomi',
      user: 'You',
      modelGateway: 'Model gateway',
      toolBoundaryNotice: 'Tool call not executed: M5 records the boundary event only and does not run external actions.',
      context: 'Context',
      localSimulated: 'Local simulated',
      stop: 'Stop',
      retry: 'Retry',
      regenerate: 'Regenerate',
      attach: 'Attach',
      messageLoomi: 'Message Loomi',
      providerUnavailableWarning: 'Model provider is not configured or unavailable',
      openProviderSettings: 'Open Settings',
      stoppedDraft: 'Generation stopped; generated content was preserved',
      recoveringDraft: 'Recovering…',
      modelDrafting: 'Model is drafting a reply',
      generating: 'Generating',
    },
    settings: {
      title: 'Settings',
      back: 'Back to workspace',
      language: 'Language',
      languageHelper: 'Switch the interface language for the current session. Chinese is the default and nothing is persisted.',
      chinese: '中文',
      english: 'English',
      working: 'Working',
      readOnly: 'Read-only',
      previewOnly: 'Preview only',
      disabled: 'Disabled',
      mixed: 'Mixed',
      available: 'Available',
      unavailable: 'Unavailable',
      misconfigured: 'Misconfigured',
      notConnected: 'Not connected',
      noThreadSelected: 'No thread selected',
      noActiveRun: 'No active run',
      general: 'General',
      generalDescription: 'Current-session workspace defaults and runtime visibility.',
      workspaceDefaults: 'Workspace defaults',
      workspaceDefaultsDescription: 'Session-local preferences for future local workspace actions.',
      defaultWorkspaceMode: 'Default workspace mode',
      defaultWorkspaceModeHelper: 'Applies to future local conversations created from the sidebar.',
      mockRuntimeScenario: 'Mock runtime scenario',
      mockRuntimeScenarioHelper: 'Applies only to future mock sends and does not mutate active runs.',
      success: 'Success',
      failure: 'Failure',
      runtimeStatus: 'Runtime status',
      runtimeStatusDescription: 'Read-only visibility for the currently selected workspace runtime.',
      dataSourceMode: 'Data source mode',
      dataSourceModeHelper: 'Shows whether the frontend is using mock data or the local API.',
      backendCapability: 'Backend capability',
      backendCapabilityHelper: 'Displays runtime availability without exposing credentials.',
      streamState: 'Stream state',
      streamStateHelper: 'Shows the selected run event stream state.',
      selectedThread: 'Selected thread',
      selectedThreadHelper: 'Shows the workspace context preserved while Settings is open.',
      selectedRunStatus: 'Selected run status',
      selectedRunStatusHelper: 'Shows the current run state without changing the run.',
      providerCapability: 'Provider capability',
      providerCapabilityHelper: 'Shows redacted provider id, family, model, and status when available.',
      providerConsoleTitle: 'Provider Test Console',
      providerConsoleDescription: 'Reads backend-configured providers and safely triggers a connection test.',
      providerConfiguredProviders: 'Configured providers',
      providerConfiguredProvidersHelper: 'These providers come from backend environment variables; Settings only shows redacted capability and never shows keys.',
      providerConsoleEmpty: 'No configured providers',
      providerConsoleEnvGuide: 'Configure providers in the local API environment, restart the backend, then refresh Settings.',
      providerTestConnection: 'Test connection',
      providerChecking: 'Checking',
      providerCheckResult: (status, message) => `${status === 'success' ? 'Success' : status === 'failed' ? 'Failed' : 'Checking'}${message ? ` · ${message}` : ''}`,
      providerLocalDraftTitle: 'Local draft',
      providerLocalDraftDescription: 'Base URL, model, and API key inputs stay in the current browser session as notes only; they are not saved and do not change real model calls.',
      providerBaseUrl: 'Base URL',
      providerBaseUrlHelper: 'Enter the OpenAI-compatible gateway URL as a local draft only.',
      providerModel: 'Model ID',
      providerModelHelper: 'Enter a model name as a local draft only; it does not affect backend calls.',
      providerApiKey: 'API Key',
      providerApiKeyHelper: 'Only whether a key was entered is shown; it is not echoed back or written to the backend.',
      providerConfigured: 'Set',
      providerNotConfigured: 'Not set',
      aboutLocalApp: 'Local app status',
      aboutLocalAppDescription: 'Shows known local state; build and account metadata remain placeholders.',
      appName: 'Application',
      appNameHelper: 'The current Loomi web shell.',
      appVersion: 'Version',
      appVersionHelper: 'M5.5 does not read build metadata yet.',
      appStatus: 'Status',
      appStatusHelper: 'Combined current data source and backend capability.',
      previewControl: 'Preview control',
      previewControlHelper: 'Mock only. This control is not connected to providers, tools, files, or backend writes.',
      connectionState: 'Connection state',
      connectionStateHelper: 'Not connected in M5.5. Future settings will require a separate implementation plan.',
      categoryPreview: (label) => `${label} preview`,
    },
  },
}

export function getDictionary(locale: Locale) {
  return dictionaries[locale]
}
