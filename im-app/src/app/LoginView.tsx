import { useMemo, useState } from 'react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { useIm } from '@/im/context'

function validUsername(username: string) {
  return /^[a-zA-Z0-9_-]{4,30}$/.test(username)
}

export function LoginView() {
  const { state, login } = useIm()
  const [username, setUsername] = useState(() => window.localStorage.getItem('huayi_im_username') ?? '')

  const hint = useMemo(() => {
    if (!username) return '4-30 个字符：字母/数字/下划线/连字符'
    if (!validUsername(username)) return '用户名不合法'
    return null
  }, [username])

  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="mx-auto flex min-h-screen w-full max-w-lg items-center p-6">
        <Card className="w-full">
          <CardHeader>
            <CardTitle>Huayi IM</CardTitle>
            <CardDescription>登录后通过 WebSocket 接收消息，通过 HTTP 发送消息。</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="text-sm font-medium">Username</div>
              <Input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="e.g. Harry"
                autoFocus
                onKeyDown={(e) => {
                  if (e.key !== 'Enter') return
                  if (!validUsername(username)) return
                  void login(username)
                }}
              />
              <div className="text-xs text-muted-foreground">{hint ?? ' '}</div>
            </div>

            {state.error ? (
              <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
                {state.error}
              </div>
            ) : null}

            <Button
              className="w-full"
              disabled={state.auth === 'logging_in' || !validUsername(username)}
              onClick={() => void login(username)}
            >
              {state.auth === 'logging_in' ? 'Logging in…' : 'Login'}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
