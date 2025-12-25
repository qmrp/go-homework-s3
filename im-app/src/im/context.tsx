import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useReducer,
  useRef,
} from "react";

import {
  apiCreateTopic,
  apiJoinTopic,
  apiLogin,
  apiLogout,
  apiQuitTopic,
  apiSendMessage,
  apiTopics,
  apiUsers,
  getSID,
  setSID,
} from "@/im/api";
import type {
  ChatMessage,
  Conversation,
  ConversationKey,
  ConversationKind,
} from "@/im/types";
import { convKey, titleFor } from "@/im/types";
import { wsURLWithSID } from "@/im/ws";

type AuthState = "logged_out" | "logging_in" | "logged_in";
type WSState = "disconnected" | "connecting" | "connected";

type State = {
  auth: AuthState;
  ws: WSState;
  me: string | null;
  error: string | null;

  conversations: Record<ConversationKey, Conversation>;
  order: ConversationKey[];
  selected: ConversationKey | null;

  draftText: Record<ConversationKey, string>;
  draftMentions: Record<ConversationKey, string>;
};

type Action =
  | { type: "AUTH_START" }
  | { type: "AUTH_OK"; me: string }
  | { type: "AUTH_LOGOUT" }
  | { type: "AUTH_ERR"; error: string }
  | { type: "WS_STATE"; ws: WSState }
  | { type: "ERR"; error: string | null }
  | { type: "SELECT"; key: ConversationKey | null }
  | { type: "UPSERT_CONV"; kind: ConversationKind; id: string }
  | { type: "CONV_TOUCHED"; key: ConversationKey; at: number; preview: string }
  | {
      type: "ADD_MSG";
      key: ConversationKey;
      msg: ChatMessage;
      unreadInc: boolean;
    }
  | {
      type: "INBOUND_MSG";
      kind: ConversationKind;
      id: string;
      msg: ChatMessage;
    }
  | { type: "SET_DRAFT_TEXT"; key: ConversationKey; text: string }
  | { type: "SET_DRAFT_MENTIONS"; key: ConversationKey; text: string };

const initialState: State = {
  auth: "logged_out",
  ws: "disconnected",
  me: null,
  error: null,
  conversations: {},
  order: [],
  selected: null,
  draftText: {},
  draftMentions: {},
};

function ensureConversation(state: State, kind: ConversationKind, id: string) {
  const key = convKey(kind, id);
  if (state.conversations[key]) return { state, key };

  const conv: Conversation = {
    key,
    kind,
    id,
    title: titleFor(kind, id),
    lastAt: 0,
    lastPreview: "",
    unread: 0,
    messages: [],
  };

  const next: State = {
    ...state,
    conversations: { ...state.conversations, [key]: conv },
    order: state.order.includes(key) ? state.order : [...state.order, key],
  };
  return { state: next, key };
}

function sortOrder(
  convs: Record<ConversationKey, Conversation>,
  order: ConversationKey[],
) {
  return [...order].sort(
    (a, b) => (convs[b]?.lastAt ?? 0) - (convs[a]?.lastAt ?? 0),
  );
}

function reduce(state: State, action: Action): State {
  switch (action.type) {
    case "AUTH_START":
      return { ...state, auth: "logging_in", error: null };
    case "AUTH_OK":
      return { ...initialState, auth: "logged_in", me: action.me };
    case "AUTH_LOGOUT":
      return { ...initialState };
    case "AUTH_ERR":
      return { ...state, auth: "logged_out", error: action.error };
    case "WS_STATE":
      return { ...state, ws: action.ws };
    case "ERR":
      return { ...state, error: action.error };
    case "UPSERT_CONV": {
      const { state: next } = ensureConversation(state, action.kind, action.id);
      return next;
    }
    case "CONV_TOUCHED": {
      const conv = state.conversations[action.key];
      if (!conv) return state;
      const nextConv: Conversation = {
        ...conv,
        lastAt: action.at,
        lastPreview: action.preview,
      };
      const nextConvs = { ...state.conversations, [action.key]: nextConv };
      return {
        ...state,
        conversations: nextConvs,
        order: sortOrder(nextConvs, state.order),
      };
    }
    case "ADD_MSG": {
      const conv = state.conversations[action.key];
      if (!conv) return state;
      const limit = 200;
      const nextMsgs = [...conv.messages, action.msg].slice(-limit);
      const nextConv: Conversation = {
        ...conv,
        messages: nextMsgs,
        lastAt: action.msg.at,
        lastPreview: action.msg.content,
        unread: action.unreadInc ? conv.unread + 1 : conv.unread,
      };
      const nextConvs = { ...state.conversations, [action.key]: nextConv };
      return {
        ...state,
        conversations: nextConvs,
        order: sortOrder(nextConvs, state.order),
      };
    }
    case "INBOUND_MSG": {
      const { state: next, key } = ensureConversation(
        state,
        action.kind,
        action.id,
      );
      const unreadInc = next.selected !== key;
      return reduce(next, { type: "ADD_MSG", key, msg: action.msg, unreadInc });
    }
    case "SELECT": {
      if (!action.key) return { ...state, selected: null };
      const conv = state.conversations[action.key];
      if (!conv) return state;
      const nextConv = { ...conv, unread: 0 };
      const nextConvs = { ...state.conversations, [action.key]: nextConv };
      return { ...state, selected: action.key, conversations: nextConvs };
    }
    case "SET_DRAFT_TEXT":
      return {
        ...state,
        draftText: { ...state.draftText, [action.key]: action.text },
      };
    case "SET_DRAFT_MENTIONS":
      return {
        ...state,
        draftMentions: { ...state.draftMentions, [action.key]: action.text },
      };
    default:
      return state;
  }
}

function parseMentions(text: string) {
  return text
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);
}

function safeJSONParse(text: string): unknown | null {
  try {
    return JSON.parse(text);
  } catch {
    return null;
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function getString(
  obj: Record<string, unknown>,
  key: string,
): string | undefined {
  const v = obj[key];
  return typeof v === "string" ? v : undefined;
}

function getNumber(
  obj: Record<string, unknown>,
  key: string,
): number | undefined {
  const v = obj[key];
  return typeof v === "number" ? v : undefined;
}

function getStringArray(
  obj: Record<string, unknown>,
  key: string,
): string[] | undefined {
  const v = obj[key];
  if (!Array.isArray(v)) return undefined;
  const out: string[] = [];
  for (const item of v) {
    if (typeof item !== "string") return undefined;
    out.push(item);
  }
  return out;
}

function asErrorMessage(err: unknown): { status?: number; message: string } {
  if (
    isRecord(err) &&
    typeof err.status === "number" &&
    typeof err.message === "string"
  ) {
    return { status: err.status, message: err.message };
  }
  if (err instanceof Error) return { message: err.message };
  return { message: "unknown error" };
}

type API = {
  state: State;
  login: (username: string) => Promise<void>;
  logout: () => Promise<void>;
  select: (key: ConversationKey) => void;
  startConversation: (kind: ConversationKind, id: string) => Promise<void>;
  send: (
    key: ConversationKey,
    content: string,
    mentions?: string,
  ) => Promise<void>;
  joinTopic: (topic: string) => Promise<void>;
  quitTopic: (topic: string) => Promise<void>;
  setDraftText: (key: ConversationKey, text: string) => void;
  setDraftMentions: (key: ConversationKey, text: string) => void;
};

const Ctx = createContext<API | null>(null);

export function ImProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(reduce, initialState);

  const wsRef = useRef<WebSocket | null>(null);
  const pingIdRef = useRef(0);
  const pendingPingRef = useRef(new Set<number>());
  const lastPingAckAtRef = useRef<number>(0);
  const reconnectTimerRef = useRef<number | null>(null);
  const pingTimerRef = useRef<number | null>(null);
  const watchdogTimerRef = useRef<number | null>(null);
  const attemptRef = useRef(0);
  const meRef = useRef<string | null>(null);
  const connectWSRef = useRef<() => void>(() => {});

  useEffect(() => {
    meRef.current = state.me;
  }, [state.me]);

  const cleanupSocket = useCallback(() => {
    if (reconnectTimerRef.current)
      window.clearTimeout(reconnectTimerRef.current);
    if (pingTimerRef.current) window.clearInterval(pingTimerRef.current);
    if (watchdogTimerRef.current)
      window.clearInterval(watchdogTimerRef.current);
    reconnectTimerRef.current = null;
    pingTimerRef.current = null;
    watchdogTimerRef.current = null;

    const ws = wsRef.current;
    wsRef.current = null;
    if (
      ws &&
      (ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING)
    ) {
      ws.close();
    }
  }, []);

  const handleInbound = useCallback((raw: MessageEvent) => {
    const data = typeof raw.data === "string" ? raw.data : "";
    const parsed = safeJSONParse(data);
    if (!isRecord(parsed)) return;

    const mt = getString(parsed, "message-type");
    if (mt === "pong") {
      const ackId = getNumber(parsed, "message-id");
      if (typeof ackId === "number") {
        wsRef.current?.send(
          JSON.stringify({ "message-type": "ack", "ack-id": ackId }),
        );
      }
      return;
    }
    if (mt === "ack") {
      const ackId = getNumber(parsed, "ack-id");
      if (typeof ackId === "number" && pendingPingRef.current.has(ackId)) {
        pendingPingRef.current.delete(ackId);
        lastPingAckAtRef.current = Date.now();
      }
      return;
    }
    if (mt === "message") {
      const from = getString(parsed, "from") ?? "";
      const content = getString(parsed, "content") ?? "";
      const topic = getString(parsed, "topic") ?? "";
      const to = getStringArray(parsed, "to");
      const kind: ConversationKind = topic ? "topic" : "p2p";
      const id = topic ? topic : from;
      const chatMsg: ChatMessage = {
        id: crypto.randomUUID(),
        at: Date.now(),
        direction: "in",
        from,
        to,
        topic: topic || undefined,
        content,
      };
      dispatch({ type: "INBOUND_MSG", kind, id, msg: chatMsg });
    }
  }, []);

  const scheduleReconnect = useCallback(() => {
    if (state.auth !== "logged_in") return;
    if (reconnectTimerRef.current) return;
    attemptRef.current++;
    const delay = Math.min(10_000, 500 + attemptRef.current * 800);
    reconnectTimerRef.current = window.setTimeout(() => {
      reconnectTimerRef.current = null;
      connectWSRef.current();
    }, delay);
  }, [state.auth]);

  const connectWS = useCallback(() => {
    if (state.auth !== "logged_in" || !meRef.current) return;
    if (wsRef.current && wsRef.current.readyState !== WebSocket.CLOSED) return;
    const sid = getSID();
    if (!sid) return;

    dispatch({ type: "WS_STATE", ws: "connecting" });
    const ws = new WebSocket(wsURLWithSID("/api/ws", sid));
    wsRef.current = ws;

    ws.onopen = () => {
      attemptRef.current = 0;
      pendingPingRef.current.clear();
      lastPingAckAtRef.current = Date.now();
      dispatch({ type: "WS_STATE", ws: "connected" });

      pingTimerRef.current = window.setInterval(() => {
        if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN)
          return;
        const me = meRef.current;
        if (!me) return;
        pingIdRef.current++;
        const id = pingIdRef.current;
        pendingPingRef.current.add(id);
        wsRef.current.send(
          JSON.stringify({
            "message-type": "ping",
            "message-id": id,
            from: me,
          }),
        );
      }, 5000);

      watchdogTimerRef.current = window.setInterval(async () => {
        const lastAck = lastPingAckAtRef.current;
        if (lastAck && Date.now() - lastAck > 65_000) {
          wsRef.current?.close();
          return;
        }
        if (state.auth !== "logged_in") return;
        try {
          await apiUsers(true);
        } catch (err: unknown) {
          const e = asErrorMessage(err);
          if (e.status === 401) {
            dispatch({ type: "AUTH_LOGOUT" });
          }
        }
      }, 30_000);
    };

    ws.onmessage = handleInbound;

    ws.onerror = () => {
      dispatch({ type: "WS_STATE", ws: "disconnected" });
      scheduleReconnect();
    };

    ws.onclose = () => {
      dispatch({ type: "WS_STATE", ws: "disconnected" });
      scheduleReconnect();
    };
  }, [handleInbound, scheduleReconnect, state.auth]);

  useEffect(() => {
    connectWSRef.current = connectWS;
  }, [connectWS]);

  useEffect(() => {
    if (state.auth !== "logged_in") {
      cleanupSocket();
      dispatch({ type: "WS_STATE", ws: "disconnected" });
      return;
    }
    connectWS();
    return () => cleanupSocket();
  }, [cleanupSocket, connectWS, state.auth]);

  useEffect(() => {
    const saved = window.localStorage.getItem("huayi_im_username");
    const sid = getSID();
    if (!saved || !sid) return;
    (async () => {
      try {
        await apiUsers(true);
        dispatch({ type: "AUTH_OK", me: saved });
        try {
          const topics = await apiTopics();
          for (const t of topics)
            dispatch({ type: "UPSERT_CONV", kind: "topic", id: t });
        } catch {
          // ignore
        }
      } catch {
        setSID(null);
        window.localStorage.removeItem("huayi_im_username");
      }
    })();
  }, []);

  const login = useCallback(async (username: string) => {
    dispatch({ type: "AUTH_START" });
    try {
      await apiLogin(username);
      if (!getSID()) {
        throw new Error("missing sid");
      }
      window.localStorage.setItem("huayi_im_username", username);
      dispatch({ type: "AUTH_OK", me: username });
      try {
        const topics = await apiTopics();
        for (const t of topics)
          dispatch({ type: "UPSERT_CONV", kind: "topic", id: t });
      } catch {
        // ignore
      }
    } catch (err: unknown) {
      const e = asErrorMessage(err);
      setSID(null);
      dispatch({ type: "AUTH_ERR", error: e.message ?? "login failed" });
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await apiLogout();
    } catch {
      // ignore
    }
    setSID(null);
    window.localStorage.removeItem("huayi_im_username");
    cleanupSocket();
    dispatch({ type: "AUTH_LOGOUT" });
  }, [cleanupSocket]);

  const select = useCallback((key: ConversationKey) => {
    dispatch({ type: "SELECT", key });
  }, []);

  const startConversation = useCallback(
    async (kind: ConversationKind, id: string) => {
      const key = convKey(kind, id);
      dispatch({ type: "UPSERT_CONV", kind, id });
      dispatch({ type: "CONV_TOUCHED", key, at: Date.now(), preview: "" });
      dispatch({ type: "SELECT", key });

      if (kind === "topic") {
        try {
          await apiCreateTopic(id);
        } catch {
          // ignore
        }
        try {
          await apiJoinTopic(id);
        } catch {
          // ignore
        }
      }
    },
    [],
  );

  const send = useCallback(
    async (key: ConversationKey, content: string, mentionsText?: string) => {
      const me = state.me;
      if (!me) return;
      const conv = state.conversations[key];
      if (!conv) return;

      const trimmed = content.trim();
      if (!trimmed) return;

      dispatch({ type: "ERR", error: null });
      try {
        await apiSendMessage({
          from: me,
          kind: conv.kind,
          id: conv.id,
          content: trimmed,
          mentions:
            conv.kind === "topic"
              ? parseMentions(mentionsText ?? "")
              : undefined,
        });
        const outMsg: ChatMessage = {
          id: crypto.randomUUID(),
          at: Date.now(),
          direction: "out",
          from: me,
          to:
            conv.kind === "p2p" ? [conv.id] : parseMentions(mentionsText ?? ""),
          topic: conv.kind === "topic" ? conv.id : undefined,
          content: trimmed,
        };
        dispatch({ type: "ADD_MSG", key, msg: outMsg, unreadInc: false });
        dispatch({ type: "SET_DRAFT_TEXT", key, text: "" });
      } catch (err: unknown) {
        const e = asErrorMessage(err);
        if (e.status === 401) {
          await logout();
          return;
        }
        dispatch({ type: "ERR", error: e.message ?? "send failed" });
      }
    },
    [logout, state.conversations, state.me],
  );

  const joinTopic = useCallback(async (topic: string) => {
    dispatch({ type: "ERR", error: null });
    try {
      await apiJoinTopic(topic);
    } catch (err: unknown) {
      const e = asErrorMessage(err);
      dispatch({ type: "ERR", error: e.message ?? "join failed" });
    }
  }, []);

  const quitTopic = useCallback(async (topic: string) => {
    dispatch({ type: "ERR", error: null });
    try {
      await apiQuitTopic(topic);
    } catch (err: unknown) {
      const e = asErrorMessage(err);
      dispatch({ type: "ERR", error: e.message ?? "quit failed" });
    }
  }, []);

  const setDraftText = useCallback((key: ConversationKey, text: string) => {
    dispatch({ type: "SET_DRAFT_TEXT", key, text });
  }, []);

  const setDraftMentions = useCallback((key: ConversationKey, text: string) => {
    dispatch({ type: "SET_DRAFT_MENTIONS", key, text });
  }, []);

  const api = useMemo<API>(
    () => ({
      state,
      login,
      logout,
      select,
      startConversation,
      send,
      joinTopic,
      quitTopic,
      setDraftText,
      setDraftMentions,
    }),
    [
      joinTopic,
      login,
      logout,
      quitTopic,
      select,
      send,
      setDraftMentions,
      setDraftText,
      startConversation,
      state,
    ],
  );

  return <Ctx.Provider value={api}>{children}</Ctx.Provider>;
}

export function useIm() {
  const ctx = useContext(Ctx);
  if (!ctx) throw new Error("useIm must be used within ImProvider");
  return ctx;
}
