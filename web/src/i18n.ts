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
    newWork: string
    projects: string
    scheduled: string
    threads: string
    settings: string
    theme: string
    update: string
    current: string
    open: string
    threadActions: string
    light: string
    dark: string
    archiveThread: string
    renameThread: string
    loadingThreads: string
    retry: string
    searchThreads: string
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
    copy: string
    model: string
    modelUnavailable: string
    attach: string
    pasteImage: string
    attachmentPending: string
    chooseWorkspaceFolder: string
    workspaceRootHome: string
    workspaceRootSelected: (name: string) => string
    messageLoomi: string
    describeTask: string
    workReadOnlyComposer: string
    providerUnavailableWarning: string
    localCodexUnsupportedWarning: string
    localCodexUnavailableWarning: string
    openProviderSettings: string
    stoppedDraft: string
    recoveringDraft: string
    modelDrafting: string
    thinkingHints: string[]
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
    theme: string
    themeHelper: string
    light: string
    dark: string
    providerConsoleTitle: string
    providerConsoleDescription: string
    providerConfiguredProviders: string
    providerConfiguredProvidersHelper: string
    providerConsoleEmpty: string
    providerConsoleEnvGuide: string
    providerTestConnection: string
    providerChecking: string
    providerAdd: string
    providerSearchPlaceholder: string
    providerFilterLabel: string
    providerFilterAll: string
    providerFilterEnabled: string
    providerFilterLocal: string
    providerFilterCloud: string
    providerName: string
    providerType: string
    providerAdvancedOptions: string
    providerHeaders: string
    providerHeaderName: string
    providerHeaderValue: string
    providerAddHeader: string
    cancel: string
    close: string
    remove: string
    providerCheckResult: (status: string, message?: string) => string
    providerLocalDraftTitle: string
    providerLocalDraftDescription: string
    providerSaveTitle: string
    providerSaveDescription: string
    providerSave: string
    providerSaving: string
    providerSaveResult: (status: string, message?: string) => string
    providerBaseUrl: string
    providerBaseUrlHelper: string
    providerModel: string
    providerModelHelper: string
    providerApiKey: string
    providerApiKeyHelper: string
    providerConfigured: string
    providerNotConfigured: string
    localProviderAutodetectTitle: string
    localProviderAutodetectDescription: string
    localProviderDetected: string
    localProviderNotDetected: string
    localProviderNeedsLogin: string
    localProviderUnsupported: string
    localProviderExplicitOptIn: string
    localProviderNoSecrets: string
    localProviderDetectAction: string
    localProviderDetectionIdle: string
    localProviderEnableForSession: string
    localProviderDisableForSession: string
    localProviderSessionLocal: string
    localProviderCredentialRedacted: string
    localProviderExecutionUnsupported: string
    localProviderExecutionSupported: string
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
      newChat: '新会话',
      newWork: '新工作',
      projects: '项目',
      scheduled: '计划任务',
      threads: '会话',
      settings: '设置',
      theme: '主题',
      update: '更新',
      current: '当前',
      open: '打开',
      threadActions: '会话操作',
      light: '浅色',
      dark: '深色',
      archiveThread: '删除会话',
      renameThread: '重命名会话',
      loadingThreads: '加载会话中',
      retry: '重试',
      searchThreads: '搜索会话',
      emptyThreads: () => '暂无会话',
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
      copy: '复制',
      model: '模型',
      modelUnavailable: '无可用模型',
      attach: '附件',
      pasteImage: '剪贴板图片',
      attachmentPending: '随本条消息发送',
      chooseWorkspaceFolder: '选择目录',
      workspaceRootHome: '默认 Home',
      workspaceRootSelected: (name: string) => `目录：${name}`,
      messageLoomi: '给 Loomi 发消息',
      describeTask: '描述你要 Loomi 完成的任务',
      workReadOnlyComposer: 'M16 Work mode 只读展示计划和进度',
      providerUnavailableWarning: '模型 Provider 未配置或不可用',
      localCodexUnsupportedWarning: 'Local Codex 已启用，但暂不支持执行',
      localCodexUnavailableWarning: 'Local Codex 登录态不可用，请重新检测或配置 OpenAI-compatible provider',
      openProviderSettings: '打开设置',
      stoppedDraft: '已停止生成，保留已生成内容',
      recoveringDraft: '恢复中…',
      modelDrafting: '模型正在生成回复',
      thinkingHints: ['组织回复', '梳理线索', '核对上下文', '提炼重点', '推敲答案', '收束思路', '准备回答', '再看一眼'],
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
      generalDescription: '当前会话的基础偏好。',
      workspaceDefaults: '工作区默认值',
      workspaceDefaultsDescription: '只影响后续本地工作区动作的当前会话偏好。',
      defaultWorkspaceMode: '默认工作区模式',
      defaultWorkspaceModeHelper: '影响之后从侧边栏创建的本地会话。',
      theme: '主题',
      themeHelper: '切换当前工作区界面的明暗显示。',
      light: '浅色',
      dark: '深色',
      providerConsoleTitle: 'Provider Test Console',
      providerConsoleDescription: '读取后端已配置的 provider，并可安全触发一次连接测试。',
      providerConfiguredProviders: '已配置 Provider',
      providerConfiguredProvidersHelper: '这些 provider 来自后端和本地设置；页面只显示脱敏能力，不显示密钥。',
      providerConsoleEmpty: '暂无已配置 provider',
      providerConsoleEnvGuide: '保存一个本地 OpenAI-compatible provider 后即可测试连接。',
      providerTestConnection: '测试连接',
      providerChecking: '测试中',
      providerAdd: '添加供应商',
      providerSearchPlaceholder: '搜索供应商...',
      providerFilterLabel: '供应商筛选',
      providerFilterAll: '全部',
      providerFilterEnabled: '已启用',
      providerFilterLocal: '本地',
      providerFilterCloud: '云端',
      providerName: '供应商名称',
      providerType: '类型',
      providerAdvancedOptions: '高级选项',
      providerHeaders: 'Headers',
      providerHeaderName: 'Header 名称',
      providerHeaderValue: 'Header 值',
      providerAddHeader: '添加 Header',
      cancel: '取消',
      close: '关闭',
      remove: '移除',
      providerCheckResult: (status, message) => `${status === 'success' ? '成功' : status === 'failed' ? '失败' : '测试中'}${message ? ` · ${message}` : ''}`,
      providerLocalDraftTitle: '本地 Provider',
      providerLocalDraftDescription: '保存 OpenAI-compatible provider 到本地 API 进程，用于真实模型调用；密钥只提交给后端，不会在界面回显。',
      providerSaveTitle: '保存本地 provider',
      providerSaveDescription: '保存后会立即刷新已配置 provider，并用于下一次模型调用。',
      providerSave: '保存',
      providerSaving: '保存中',
      providerSaveResult: (status, message) => `${status === 'success' ? '已保存' : status === 'failed' ? '保存失败' : '保存中'}${message ? ` · ${message}` : ''}`,
      providerBaseUrl: 'Base URL',
      providerBaseUrlHelper: '填写 OpenAI-compatible 中转站地址。',
      providerModel: '模型 ID',
      providerModelHelper: '填写下一次真实模型调用使用的模型名称。',
      providerApiKey: 'API Key',
      providerApiKeyHelper: '密钥只发送给本地 API 保存到当前进程，不会从接口返回或在界面回显。',
      providerConfigured: '已填写',
      providerNotConfigured: '未填写',
      localProviderAutodetectTitle: '本地 Provider 自动检测',
      localProviderAutodetectDescription: '只检测 Claude Code / Codex 登录态或密钥是否可能可用；不会启用、读取明文或发起模型调用。',
      localProviderDetected: 'detected',
      localProviderNotDetected: 'not detected',
      localProviderNeedsLogin: 'needs login',
      localProviderUnsupported: 'unsupported',
      localProviderExplicitOptIn: 'requires explicit opt-in before use',
      localProviderNoSecrets: 'no secrets shown',
      localProviderDetectAction: '检测本机 Provider',
      localProviderDetectionIdle: '点击检测后才会读取本机 Provider 可用性摘要。',
      localProviderEnableForSession: '本会话启用',
      localProviderDisableForSession: '禁用本会话',
      localProviderSessionLocal: 'session-local',
      localProviderCredentialRedacted: 'credential redacted',
      localProviderExecutionUnsupported: 'Local Codex is enabled but execution is not supported yet',
      localProviderExecutionSupported: 'execution supported',
      aboutLocalApp: '本地应用状态',
      aboutLocalAppDescription: '显示已知本地状态；构建和账号信息仍是占位。',
      appName: '应用',
      appNameHelper: '当前运行的 Loomi web shell。',
      appVersion: '版本',
      appVersionHelper: 'M5.5 暂不读取构建版本。',
      appStatus: '状态',
      appStatusHelper: '真实 API 和后端能力的组合状态。',
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
      newChat: 'New thread',
      newWork: 'New Work',
      projects: 'Projects',
      scheduled: 'Scheduled',
      threads: 'Threads',
      settings: 'Settings',
      theme: 'Theme',
      update: 'Update',
      current: 'Current',
      open: 'Open',
      threadActions: 'Thread actions',
      light: 'Light',
      dark: 'Dark',
      archiveThread: 'Delete thread',
      renameThread: 'Rename thread',
      loadingThreads: 'Loading threads',
      retry: 'Retry',
      searchThreads: 'Search threads',
      emptyThreads: () => 'No threads',
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
      copy: 'Copy',
      model: 'Model',
      modelUnavailable: 'No model',
      attach: 'Attach',
      pasteImage: 'Clipboard image',
      attachmentPending: 'queued for this message',
      chooseWorkspaceFolder: 'Choose folder',
      workspaceRootHome: 'Default Home',
      workspaceRootSelected: (name: string) => `Folder: ${name}`,
      messageLoomi: 'Message Loomi',
      describeTask: 'Describe the task for Loomi',
      workReadOnlyComposer: 'M16 Work mode is read-only for plan and progress',
      providerUnavailableWarning: 'Model provider is not configured or unavailable',
      localCodexUnsupportedWarning: 'Local Codex is enabled but execution is not supported yet',
      localCodexUnavailableWarning: 'Local Codex login is unavailable. Detect again or configure an OpenAI-compatible provider',
      openProviderSettings: 'Provider Settings',
      stoppedDraft: 'Generation stopped; generated content was preserved',
      recoveringDraft: 'Recovering…',
      modelDrafting: 'Model is drafting a reply',
      thinkingHints: ['Organizing reply', 'Tracing context', 'Checking details', 'Sharpening answer', 'Gathering signal', 'Preparing reply'],
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
      generalDescription: 'Basic preferences for the current session.',
      workspaceDefaults: 'Workspace defaults',
      workspaceDefaultsDescription: 'Session-local preferences for future local workspace actions.',
      defaultWorkspaceMode: 'Default workspace mode',
      defaultWorkspaceModeHelper: 'Applies to future local conversations created from the sidebar.',
      theme: 'Theme',
      themeHelper: 'Switches the current workspace between light and dark appearance.',
      light: 'Light',
      dark: 'Dark',
      providerConsoleTitle: 'Provider Test Console',
      providerConsoleDescription: 'Reads configured providers and safely triggers a connection test.',
      providerConfiguredProviders: 'Configured providers',
      providerConfiguredProvidersHelper: 'These providers come from backend configuration and local Settings; only redacted capability is shown.',
      providerConsoleEmpty: 'No configured providers',
      providerConsoleEnvGuide: 'Save a local OpenAI-compatible provider to test the connection.',
      providerTestConnection: 'Test connection',
      providerChecking: 'Checking',
      providerAdd: 'Add provider',
      providerSearchPlaceholder: 'Search providers...',
      providerFilterLabel: 'Provider filters',
      providerFilterAll: 'All',
      providerFilterEnabled: 'Enabled',
      providerFilterLocal: 'Local',
      providerFilterCloud: 'Cloud',
      providerName: 'Provider name',
      providerType: 'Type',
      providerAdvancedOptions: 'Advanced options',
      providerHeaders: 'Headers',
      providerHeaderName: 'Header name',
      providerHeaderValue: 'Header value',
      providerAddHeader: 'Add Header',
      cancel: 'Cancel',
      close: 'Close',
      remove: 'Remove',
      providerCheckResult: (status, message) => `${status === 'success' ? 'Success' : status === 'failed' ? 'Failed' : 'Checking'}${message ? ` · ${message}` : ''}`,
      providerLocalDraftTitle: 'Local provider',
      providerLocalDraftDescription: 'Save an OpenAI-compatible provider into the local API process for real model calls. Keys are submitted to the backend and never echoed in the UI.',
      providerSaveTitle: 'Save local provider',
      providerSaveDescription: 'After saving, the provider list refreshes and the next model call can use it.',
      providerSave: 'Save',
      providerSaving: 'Saving',
      providerSaveResult: (status, message) => `${status === 'success' ? 'Saved' : status === 'failed' ? 'Save failed' : 'Saving'}${message ? ` · ${message}` : ''}`,
      providerBaseUrl: 'Base URL',
      providerBaseUrlHelper: 'Enter the OpenAI-compatible gateway URL.',
      providerModel: 'Model ID',
      providerModelHelper: 'Enter the model name used for the next real model call.',
      providerApiKey: 'API Key',
      providerApiKeyHelper: 'The key is sent only to the local API and is never returned or echoed in the UI.',
      providerConfigured: 'Set',
      providerNotConfigured: 'Not set',
      localProviderAutodetectTitle: 'Local provider autodetect',
      localProviderAutodetectDescription: 'Detects whether Claude Code or Codex may be usable locally; it does not enable providers, show secrets, or start model calls.',
      localProviderDetected: 'detected',
      localProviderNotDetected: 'not detected',
      localProviderNeedsLogin: 'needs login',
      localProviderUnsupported: 'unsupported',
      localProviderExplicitOptIn: 'requires explicit opt-in before use',
      localProviderNoSecrets: 'no secrets shown',
      localProviderDetectAction: 'Detect local providers',
      localProviderDetectionIdle: 'Run detection to inspect local provider availability.',
      localProviderEnableForSession: 'Enable for this session',
      localProviderDisableForSession: 'Disable for this session',
      localProviderSessionLocal: 'session-local',
      localProviderCredentialRedacted: 'credential redacted',
      localProviderExecutionUnsupported: 'Local Codex is enabled but execution is not supported yet',
      localProviderExecutionSupported: 'execution supported',
      aboutLocalApp: 'Local app status',
      aboutLocalAppDescription: 'Shows known local state; build and account metadata remain placeholders.',
      appName: 'Application',
      appNameHelper: 'The current Loomi web shell.',
      appVersion: 'Version',
      appVersionHelper: 'M5.5 does not read build metadata yet.',
      appStatus: 'Status',
      appStatusHelper: 'Combined Real API and backend capability status.',
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
