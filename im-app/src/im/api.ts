import type { ConversationKind } from "@/im/types";

export type ApiError = {
  status: number;
  message: string;
};

const SID_KEY = "huayi_im_sid";

export function getSID(): string | null {
  return window.localStorage.getItem(SID_KEY);
}

export function setSID(sid: string | null) {
  if (!sid) window.localStorage.removeItem(SID_KEY);
  else window.localStorage.setItem(SID_KEY, sid);
}

async function readErrorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string };
    return data.error || res.statusText;
  } catch {
    return res.statusText;
  }
}

async function request(path: string, init?: RequestInit): Promise<Response> {
  const sid = getSID();
  const headers = new Headers(init?.headers ?? undefined);
  if (sid) headers.set("Authorization", `Bearer ${sid}`);
  return fetch(path, {
    credentials: "omit",
    ...init,
    headers,
  });
}

export async function apiLogin(username: string): Promise<void> {
  const res = await request("/api/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username }),
  });
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
  try {
    const data = (await res.json()) as { sid?: string; username?: string };
    if (data.sid) setSID(data.sid);
  } catch {
    // ignore
  }
}

export async function apiLogout(): Promise<void> {
  const res = await request("/api/logout", { method: "POST" });
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
  setSID(null);
}

export async function apiUsers(online?: boolean): Promise<string[]> {
  const url = new URL("/api/users", window.location.origin);
  if (online !== undefined) url.searchParams.set("online", String(online));
  const res = await request(url.pathname + url.search);
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
  const data = (await res.json()) as { list: Array<{ username: string }> };
  return data.list.map((x) => x.username);
}

export async function apiTopics(): Promise<string[]> {
  const res = await request("/api/topics");
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
  const data = (await res.json()) as { list: Array<{ topic: string }> };
  return data.list.map((x) => x.topic);
}

export async function apiCreateTopic(topic: string): Promise<void> {
  const res = await request("/api/topics", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ topic }),
  });
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
}

export async function apiDeleteTopic(topic: string): Promise<void> {
  const res = await request(`/api/topics/${encodeURIComponent(topic)}`, {
    method: "DELETE",
  });
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
}

export async function apiJoinTopic(topic: string): Promise<void> {
  const res = await request(
    `/api/topics/${encodeURIComponent(topic)}/actions/join`,
    { method: "POST" },
  );
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
}

export async function apiQuitTopic(topic: string): Promise<void> {
  const res = await request(
    `/api/topics/${encodeURIComponent(topic)}/actions/quit`,
    { method: "POST" },
  );
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
}

export async function apiSendMessage(params: {
  from: string;
  kind: ConversationKind;
  id: string;
  content: string;
  mentions?: string[];
}): Promise<void> {
  const body =
    params.kind === "topic"
      ? {
          "message-type": "message",
          from: params.from,
          to: params.mentions ?? [],
          topic: params.id,
          "content-type": "text/plain",
          content: params.content,
        }
      : {
          "message-type": "message",
          from: params.from,
          to: [params.id],
          "content-type": "text/plain",
          content: params.content,
        };
  const res = await request("/api/messages", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!res.ok) {
    throw {
      status: res.status,
      message: await readErrorMessage(res),
    } satisfies ApiError;
  }
}
