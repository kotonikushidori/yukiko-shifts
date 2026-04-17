// api.js — API クライアント

function authHeaders() {
  const token = localStorage.getItem('shift_token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

async function request(url, options = {}) {
  const res = await fetch(url, { headers: authHeaders(), ...options });
  if (res.status === 204) return null;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}

// POST /api/auth/login
export async function apiLogin(employeeId, password) {
  const res = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ employee_id: employeeId, password }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'ログインに失敗しました');
  return data; // { token, user }
}

// GET /api/shifts/board?from=YYYY-MM-DD&to=YYYY-MM-DD
export function apiGetBoard(from, to) {
  return request(`/api/shifts/board?from=${from}&to=${to}`);
}

// POST /api/sites/{siteID}/shifts/{date}/assign
export function apiCreateAssign(siteId, date, userId, timeSlot) {
  return request(`/api/sites/${siteId}/shifts/${date}/assign`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, time_slot: timeSlot }),
  });
}

// DELETE /api/shifts/assign/{id}
export function apiDeleteAssign(id) {
  return request(`/api/shifts/assign/${id}`, { method: 'DELETE' });
}

// GET /api/workers
export async function apiGetWorkers() {
  try {
    return await request('/api/workers');
  } catch {
    return null;
  }
}

// POST /api/admin/workers
export function apiCreateWorker(data) {
  return request('/api/admin/workers', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// PUT /api/admin/workers/{id}
export function apiUpdateWorker(id, data) {
  return request(`/api/admin/workers/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// ─── 現場マスタ ──────────────────────────────────────────────

// GET /api/sites
export function apiGetSites() {
  return request('/api/sites');
}

// GET /api/sites/{id}
export function apiGetSite(id) {
  return request(`/api/sites/${id}`);
}

// POST /api/sites
export function apiCreateSite(data) {
  return request('/api/sites', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

// PUT /api/sites/{id}
export function apiUpdateSite(id, data) {
  return request(`/api/sites/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// ─── 日報 ─────────────────────────────────────────────────────

// GET /api/reports/my?year=Y&month=M
export function apiGetMyReports(year, month) {
  return request(`/api/reports/my?year=${year}&month=${month}`);
}

// PUT /api/reports/{date}
export function apiUpsertReport(date, data) {
  return request(`/api/reports/${date}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
}

// PUT /api/reports/site-client
export function apiUpdateSiteClient(year, month, siteId, clientName) {
  return request('/api/reports/site-client', {
    method: 'PUT',
    body: JSON.stringify({ year, month, site_id: siteId, client_name: clientName }),
  });
}

// ─── Web Push ────────────────────────────────────────────────

// GET /api/push/vapid-key
export function apiGetVapidKey() {
  return request('/api/push/vapid-key');
}

// POST /api/push/subscribe
export function apiSubscribePush(subJson) {
  // subJson は PushSubscription.toJSON() の結果
  // { endpoint, keys: { p256dh, auth } } の形式
  return request('/api/push/subscribe', {
    method: 'POST',
    body: JSON.stringify({
      endpoint: subJson.endpoint,
      p256dh:   subJson.keys?.p256dh ?? '',
      auth:     subJson.keys?.auth   ?? '',
    }),
  });
}

// DELETE /api/push/subscribe
export function apiUnsubscribePush(endpoint) {
  return request('/api/push/subscribe', {
    method: 'DELETE',
    body: JSON.stringify({ endpoint }),
  });
}

// POST /api/push/hope-submit
export function apiHopeSubmit(year, month) {
  return request('/api/push/hope-submit', {
    method: 'POST',
    body: JSON.stringify({ year, month }),
  });
}

// ─── シフトロック ─────────────────────────────────────────────

// GET /api/shifts/lock?year=Y&month=M
export function apiGetLockStatus(year, month) {
  return request(`/api/shifts/lock?year=${year}&month=${month}`);
}

// POST /api/admin/shifts/lock
export function apiLockMonth(year, month) {
  return request('/api/admin/shifts/lock', {
    method: 'POST',
    body: JSON.stringify({ year, month }),
  });
}

// DELETE /api/admin/shifts/lock
export function apiUnlockMonth(year, month) {
  return request('/api/admin/shifts/lock', {
    method: 'DELETE',
    body: JSON.stringify({ year, month }),
  });
}
