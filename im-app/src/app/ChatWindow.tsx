import { Hash, Send, User } from "lucide-react";
import { useEffect, useMemo, useRef } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Separator } from "@/components/ui/separator";
import { Textarea } from "@/components/ui/textarea";
import { useIm } from "@/im/context";
import type { ChatMessage } from "@/im/types";
import { cn } from "@/lib/utils";

function fmtTime(ts: number) {
  return new Date(ts).toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export function ChatWindow() {
  const { state } = useIm();
  const conv = state.selected ? state.conversations[state.selected] : null;

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {state.error ? (
        <div className="border-b bg-destructive/10 px-4 py-2 text-sm text-destructive">
          {state.error}
        </div>
      ) : null}

      {conv ? <ActiveConversation /> : <EmptyConversation />}
    </div>
  );
}

function EmptyConversation() {
  return (
    <div className="flex min-h-0 flex-1 items-center justify-center p-6">
      <div className="max-w-md rounded-xl border bg-muted/30 p-6 text-center">
        <div className="text-lg font-semibold">No chat selected</div>
        <div className="mt-2 text-sm text-muted-foreground">
          Use the sidebar to start a direct chat or join a topic.
        </div>
      </div>
    </div>
  );
}

function ActiveConversation() {
  const { state, joinTopic, quitTopic, send, setDraftMentions, setDraftText } =
    useIm();
  const key = state.selected!;
  const conv = state.conversations[key]!;

  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [conv.messages.length, key]);

  const draft = state.draftText[key] ?? "";
  const mentions = state.draftMentions[key] ?? "";

  const titleIcon = conv.kind === "topic" ? Hash : User;

  const headerMeta = useMemo(() => {
    if (conv.kind === "topic") return `Topic: ${conv.id}`;
    return `Direct: ${conv.id}`;
  }, [conv.id, conv.kind]);

  return (
    <>
      <div className="flex items-center justify-between gap-3 px-4 py-3">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-md border bg-background">
            {(() => {
              const Icon = titleIcon;
              return <Icon className="h-4 w-4" />;
            })()}
          </div>
          <div className="min-w-0">
            <div className="truncate text-sm font-semibold">{conv.title}</div>
            <div className="truncate text-xs text-muted-foreground">
              {headerMeta}
            </div>
          </div>
        </div>
        {conv.kind === "topic" ? (
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => void joinTopic(conv.id)}
            >
              Join
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void quitTopic(conv.id)}
            >
              Quit
            </Button>
          </div>
        ) : null}
      </div>
      <Separator />

      <div className="min-h-0 flex-1">
        <ScrollArea className="h-full">
          <div className="space-y-3 p-4">
            {conv.messages.map((m) => (
              <MessageRow key={m.id} convKind={conv.kind} msg={m} />
            ))}
            <div ref={bottomRef} />
          </div>
        </ScrollArea>
      </div>

      <Separator />

      <div className="space-y-2 p-4">
        {conv.kind === "topic" ? (
          <div className="grid gap-2 sm:grid-cols-2">
            <div className="text-sm text-muted-foreground sm:col-span-2">
              Mentions (comma-separated)
            </div>
            <Input
              value={mentions}
              onChange={(e) => setDraftMentions(key, e.target.value)}
              placeholder="e.g. Ron,Hermione"
              className="sm:col-span-2"
            />
          </div>
        ) : null}

        <div className="flex items-end gap-2">
          <Textarea
            value={draft}
            onChange={(e) => setDraftText(key, e.target.value)}
            placeholder="Type a message…"
            className="min-h-[44px] resize-none"
            onKeyDown={(e) => {
              if (e.key !== "Enter" || e.shiftKey) return;
              e.preventDefault();
              void send(key, draft, mentions);
            }}
          />
          <Button
            size="icon"
            onClick={() => void send(key, draft, mentions)}
            title="Send"
          >
            <Send />
          </Button>
        </div>
        <div className="text-xs text-muted-foreground">
          Enter 发送，Shift+Enter 换行。
        </div>
      </div>
    </>
  );
}

function MessageRow({
  convKind,
  msg,
}: {
  convKind: "p2p" | "topic";
  msg: ChatMessage;
}) {
  const isOut = msg.direction === "out";
  const name = isOut ? "You" : msg.from;
  const showFrom = convKind === "topic" && !isOut;

  return (
    <div className={cn("flex", isOut ? "justify-end" : "justify-start")}>
      <div
        className={cn(
          "max-w-[80%] rounded-xl px-3 py-2 text-sm shadow-sm",
          isOut
            ? "bg-primary text-primary-foreground"
            : "bg-muted text-foreground",
        )}
      >
        {showFrom ? (
          <div
            className={cn(
              "mb-1 text-xs",
              isOut ? "text-primary-foreground/70" : "text-muted-foreground",
            )}
          >
            {name}
          </div>
        ) : null}
        <div className="whitespace-pre-wrap break-words">{msg.content}</div>
        <div
          className={cn(
            "mt-1 text-[10px]",
            isOut ? "text-primary-foreground/70" : "text-muted-foreground",
          )}
        >
          {fmtTime(msg.at)}
        </div>
      </div>
    </div>
  );
}
