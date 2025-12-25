export function wsURL(path: string) {
  const proto = window.location.protocol === "https:" ? "wss" : "ws";
  return `${proto}://${window.location.host}${path}`;
}

export function wsURLWithSID(path: string, sid: string) {
  const u = new URL(wsURL(path));
  u.searchParams.set("sid", sid);
  return u.toString();
}
