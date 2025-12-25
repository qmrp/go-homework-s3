import { Hash, MessageSquare } from 'lucide-react'
import { useMemo, useState } from 'react'

import { NewConversationDialog } from '@/app/NewConversationDialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import { cn } from '@/lib/utils'
import { useIm } from '@/im/context'

export function ChatSidebar() {
  const { state, select } = useIm()
  const [q, setQ] = useState('')

  const items = useMemo(() => {
    const query = q.trim().toLowerCase()
    const keys = state.order
    if (!query) return keys
    return keys.filter((k) => state.conversations[k]?.title.toLowerCase().includes(query))
  }, [q, state.conversations, state.order])

  return (
    <div className="flex h-[calc(100vh-57px)] flex-col gap-3 p-4">
      <NewConversationDialog />
      <Input
        value={q}
        onChange={(e) => setQ(e.target.value)}
        placeholder="Search…"
      />
      <ScrollArea className="min-h-0 flex-1">
        <div className="space-y-1 pb-2">
          {items.length === 0 ? (
            <div className="rounded-md border bg-muted/30 p-3 text-sm text-muted-foreground">
              No conversations yet.
            </div>
          ) : null}
          {items.map((key) => {
            const conv = state.conversations[key]
            if (!conv) return null
            const active = state.selected === key
            const Icon = conv.kind === 'topic' ? Hash : MessageSquare
            return (
              <Button
                key={key}
                variant="ghost"
                className={cn(
                  'h-auto w-full justify-start gap-3 px-3 py-2',
                  active && 'bg-accent',
                )}
                onClick={() => select(key)}
              >
                <div className="mt-0.5 flex h-8 w-8 items-center justify-center rounded-md border bg-background">
                  <Icon className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1 text-left">
                  <div className="flex items-center gap-2">
                    <div className="truncate text-sm font-medium">{conv.title}</div>
                    {conv.unread > 0 ? (
                      <Badge variant="secondary" className="ml-auto">
                        {conv.unread}
                      </Badge>
                    ) : null}
                  </div>
                  <div className="truncate text-xs text-muted-foreground">
                    {conv.lastPreview || (conv.kind === 'topic' ? 'Topic' : 'Direct chat')}
                  </div>
                </div>
              </Button>
            )
          })}
        </div>
      </ScrollArea>
    </div>
  )
}
