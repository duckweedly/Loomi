import type { Message, Run, Thread } from './domain'

export const threads: Thread[] = [
  {
    id: 'thread-brief',
    title: 'M1 interface pass',
    project: 'Loomi',
    mode: 'work',
    updatedAt: '2m',
    lifecycleStatus: 'active',
    runStatus: 'running',
  },
  {
    id: 'thread-docs',
    title: 'Roadmap alignment',
    project: 'Planning',
    mode: 'work',
    updatedAt: '18m',
    lifecycleStatus: 'active',
    runStatus: 'completed',
  },
  {
    id: 'thread-agent',
    title: 'Agent profile sketch',
    project: 'Lab',
    mode: 'chat',
    updatedAt: '1h',
    lifecycleStatus: 'active',
    runStatus: 'stopped',
  },
]

export const messages: Message[] = [
  {
    id: 'msg-1',
    threadId: 'thread-brief',
    role: 'user',
    content: 'Create a polished M1 shell for Loomi. Keep it desktop-like, compact, and glassy.',
    createdAt: '10:24',
  },
  {
    id: 'msg-2',
    threadId: 'thread-brief',
    role: 'assistant',
    content: 'Set the shell as a three-panel workspace: navigation, cowork canvas, and run rail. Use mock events only.',
    createdAt: '10:25',
    toolCalls: [
      {
        id: 'tool-1',
        name: 'scan_reference',
        status: 'completed',
        summary: 'Mapped Arkloop work mode, timeline, right panel, and artifacts.',
        input: 'Arkloop web workspace files',
        output: 'Work mode, timeline, context compact bar, right panel tabs',
      },
    ],
  },
  {
    id: 'msg-3',
    threadId: 'thread-docs',
    role: 'user',
    content: 'Keep M1 clean. No backend, no auth, no Electron.',
    createdAt: '09:48',
  },
  {
    id: 'msg-4',
    threadId: 'thread-docs',
    role: 'assistant',
    content: 'M1 stays browser-only with a replaceable API boundary and local mock data.',
    createdAt: '09:49',
  },
  {
    id: 'msg-5',
    threadId: 'thread-agent',
    role: 'user',
    content: 'Sketch a small agent profile surface.',
    createdAt: 'Yesterday',
  },
  {
    id: 'msg-6',
    threadId: 'thread-agent',
    role: 'assistant',
    content: 'Use a compact profile card later. For M1, show agent metadata in the right rail only.',
    createdAt: 'Yesterday',
  },
]

export const runs: Run[] = [
  {
    id: 'run-1',
    threadId: 'thread-brief',
    status: 'running',
    model: 'Claude Sonnet',
    context: '42k / 128k',
    events: [
      { id: 'evt-1', type: 'run.started', label: 'Started', detail: 'Work mode run', time: '10:25', status: 'completed' },
      { id: 'evt-2', type: 'context.loaded', label: 'Context', detail: 'Docs and reference map', time: '10:25', status: 'completed' },
      { id: 'evt-3', type: 'tool.completed', label: 'Tool', detail: 'Reference scan complete', time: '10:26', status: 'completed' },
      { id: 'evt-4', type: 'message.drafting', label: 'Drafting', detail: 'Preparing UI shell', time: 'Now', status: 'running' },
    ],
  },
  {
    id: 'run-2',
    threadId: 'thread-docs',
    status: 'completed',
    model: 'Claude Sonnet',
    context: '18k / 128k',
    events: [
      { id: 'evt-5', type: 'run.started', label: 'Started', detail: 'Roadmap pass', time: '09:48', status: 'completed' },
      { id: 'evt-6', type: 'run.completed', label: 'Completed', detail: 'M1 boundary confirmed', time: '09:49', status: 'completed' },
    ],
  },
  {
    id: 'run-3',
    threadId: 'thread-agent',
    status: 'stopped',
    model: 'Claude Sonnet',
    context: '9k / 128k',
    events: [
      { id: 'evt-7', type: 'run.started', label: 'Started', detail: 'Agent profile sketch', time: 'Yesterday', status: 'completed' },
      { id: 'evt-8', type: 'run.stopped', label: 'Stopped', detail: 'Paused by user', time: 'Yesterday', status: 'stopped' },
    ],
  },
]
