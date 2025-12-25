import { useMemo, useState } from 'react'

import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useIm } from '@/im/context'
import type { ConversationKind } from '@/im/types'

function validName(kind: ConversationKind, value: string) {
  const re = /^[a-zA-Z0-9_-]{4,30}$/
  if (kind === 'p2p') return re.test(value)
  return re.test(value)
}

export function NewConversationDialog() {
  const { startConversation } = useIm()
  const [open, setOpen] = useState(false)
  const [kind, setKind] = useState<ConversationKind>('p2p')
  const [value, setValue] = useState('')
  const canSubmit = useMemo(() => validName(kind, value), [kind, value])

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" className="w-full">
          New chat / topic
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Start a conversation</DialogTitle>
          <DialogDescription>单聊输入 username，群聊输入 topic（4-30 字符）。</DialogDescription>
        </DialogHeader>

        <div className="space-y-3">
          <Tabs value={kind} onValueChange={(v) => setKind(v as ConversationKind)}>
            <TabsList className="grid w-full grid-cols-2">
              <TabsTrigger value="p2p">Direct</TabsTrigger>
              <TabsTrigger value="topic">Topic</TabsTrigger>
            </TabsList>
          </Tabs>
          <Input
            value={value}
            onChange={(e) => setValue(e.target.value)}
            placeholder={kind === 'topic' ? 'e.g. golang' : 'e.g. Ron'}
            autoFocus
            onKeyDown={(e) => {
              if (e.key !== 'Enter') return
              if (!canSubmit) return
              void startConversation(kind, value).then(() => {
                setOpen(false)
                setValue('')
              })
            }}
          />
          <div className="text-xs text-muted-foreground">
            {canSubmit ? ' ' : '只允许字母、数字、下划线、连字符；长度 4-30。'}
          </div>
        </div>

        <DialogFooter>
          <Button
            disabled={!canSubmit}
            onClick={() => {
              void startConversation(kind, value).then(() => {
                setOpen(false)
                setValue('')
              })
            }}
          >
            Start
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
