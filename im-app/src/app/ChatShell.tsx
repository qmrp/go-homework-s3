import { LogOut, Wifi, WifiOff } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { useIm } from '@/im/context'
import { ChatSidebar } from '@/app/ChatSidebar'
import { ChatWindow } from '@/app/ChatWindow'

export function ChatShell() {
  const { state, logout } = useIm()

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="flex h-screen w-full">
        <aside className="w-[340px] shrink-0 border-r">
          <div className="flex items-center justify-between px-4 py-3">
            <div className="flex flex-col">
              <div className="text-sm font-semibold leading-none">Huayi IM</div>
              <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                <span>{state.me}</span>
                <span className="inline-flex items-center gap-1">
                  {state.ws === 'connected' ? (
                    <>
                      <Wifi className="h-3.5 w-3.5" /> online
                    </>
                  ) : (
                    <>
                      <WifiOff className="h-3.5 w-3.5" /> {state.ws}
                    </>
                  )}
                </span>
              </div>
            </div>
            <Button variant="ghost" size="icon" onClick={() => void logout()} title="Logout">
              <LogOut />
            </Button>
          </div>
          <Separator />
          <ChatSidebar />
        </aside>

        <main className="flex min-w-0 flex-1 flex-col">
          <ChatWindow />
        </main>
      </div>
    </div>
  )
}
