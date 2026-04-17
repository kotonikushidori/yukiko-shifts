// push.js — Web Push サブスクリプション管理

import { apiGetVapidKey, apiSubscribePush, apiUnsubscribePush } from './api.js';

function urlBase64ToUint8Array(base64) {
  const padding = '='.repeat((4 - (base64.length % 4)) % 4);
  const b64 = (base64 + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(b64);
  return Uint8Array.from([...raw].map(c => c.charCodeAt(0)));
}

// Service Worker 登録 + プッシュ購読
// ブラウザが非対応・許可拒否のときは静かに終了する
export async function initPush() {
  if (!('serviceWorker' in navigator) || !('PushManager' in window)) return;

  try {
    const { public_key: publicKey } = await apiGetVapidKey();
    if (!publicKey) return; // サーバー側でVAPID未設定

    const reg = await navigator.serviceWorker.register('/static/sw.js');

    // 既存サブスクリプションがあれば再登録（キーローテーション対応）
    const existing = await reg.pushManager.getSubscription();
    if (existing) {
      await apiSubscribePush(existing.toJSON()).catch(() => {});
      return;
    }

    // 通知許可を求める
    const permission = await Notification.requestPermission();
    if (permission !== 'granted') return;

    const sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(publicKey),
    });
    await apiSubscribePush(sub.toJSON());
  } catch (err) {
    console.warn('[push] 初期化エラー:', err);
  }
}

// ログアウト時などにサブスクリプションを削除する
export async function clearPush() {
  if (!('serviceWorker' in navigator)) return;
  try {
    const reg = await navigator.serviceWorker.getRegistration('/static/sw.js');
    if (!reg) return;
    const sub = await reg.pushManager.getSubscription();
    if (!sub) return;
    await apiUnsubscribePush(sub.endpoint);
    await sub.unsubscribe();
  } catch (err) {
    console.warn('[push] 解除エラー:', err);
  }
}
