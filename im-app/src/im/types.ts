export type ConversationKind = 'p2p' | 'topic'
export type ConversationKey = `${ConversationKind}:${string}`

export type MessageDirection = 'in' | 'out' | 'system'

export interface ChatMessage {
  id: string
  at: number
  direction: MessageDirection
  from: string
  to?: string[]
  topic?: string
  content: string
}

export interface Conversation {
  key: ConversationKey
  kind: ConversationKind
  id: string
  title: string
  lastAt: number
  lastPreview: string
  unread: number
  messages: ChatMessage[]
}

export interface WsPing {
  'message-type': 'ping'
  'message-id': number
  from: string
}

export interface WsPong {
  'message-type': 'pong'
  'message-id': number
  to: string
}

export interface WsAck {
  'message-type': 'ack'
  'ack-id': number
}

export interface DownMessage {
  'message-type': 'message'
  from: string
  to: string[]
  topic?: string
  'content-type': 'text/plain'
  content: string
}

export type WsInbound = WsPong | WsAck | DownMessage

export function convKey(kind: ConversationKind, id: string): ConversationKey {
  return `${kind}:${id}`
}

export function titleFor(kind: ConversationKind, id: string) {
  return kind === 'topic' ? `#${id}` : id
}
