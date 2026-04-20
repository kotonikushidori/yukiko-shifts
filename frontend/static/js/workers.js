// workers.js — 作業者管理（管理者用）

import { apiGetWorkers, apiCreateWorker, apiUpdateWorker } from './api.js';

function escHtml(s) {
  return String(s ?? '').replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

let _workers = [];
let _editId  = null; // 編集中のworker ID (null = 新規)

// ─── Init ────────────────────────────────────────────────────
export async function initWorkers() {
  const root = document.getElementById('workers-root');
  if (!root) return;
  root.innerHTML = `<div class="wm-loading"><div class="spinner"></div></div>`;
  _workers = await apiGetWorkers().catch(() => []) ?? [];
  render(root);
}

// ─── Render ─────────────────────────────────────────────────
function render(root) {
  const rows = _workers.map(w => {
    const ln = w.last_name  ?? '';
    const fn = w.first_name ?? '';
    const foremanBadge = w.is_foreman_qualified
      ? `<span class="wm-foreman-badge">職長</span>` : '';
    return `
      <tr>
        <td>${escHtml(w.employee_id)}</td>
        <td>${escHtml(ln)}</td>
        <td>${escHtml(fn)}</td>
        <td>${foremanBadge}</td>
        <td>${escHtml(w.phone ?? '—')}</td>
        <td>
          <button class="wm-edit-btn btn btn-sm" data-id="${w.id}">編集</button>
        </td>
      </tr>`;
  }).join('');

  root.innerHTML = `
    <div class="wm-page">
      <div class="wm-header">
        <h2 class="wm-title">作業者管理</h2>
        <button class="btn btn-primary" id="wm-add-btn">＋ 追加</button>
      </div>
      <table class="wm-table">
        <thead>
          <tr><th>社員ID</th><th>苗字</th><th>名前</th><th>職長</th><th>電話</th><th></th></tr>
        </thead>
        <tbody>${rows}</tbody>
      </table>
    </div>
    <div class="wm-modal-overlay" id="wm-overlay" style="display:none">
      <div class="wm-modal" id="wm-modal"></div>
    </div>`;

  document.getElementById('wm-add-btn').addEventListener('click', () => openModal(null));
  root.querySelectorAll('.wm-edit-btn').forEach(btn => {
    btn.addEventListener('click', () => {
      const w = _workers.find(w => w.id === parseInt(btn.dataset.id, 10));
      if (w) openModal(w);
    });
  });
}

// ─── Modal ───────────────────────────────────────────────────
function openModal(worker) {
  _editId = worker?.id ?? null;
  const isNew = _editId === null;
  const overlay = document.getElementById('wm-overlay');
  const modal   = document.getElementById('wm-modal');

  modal.innerHTML = `
    <div class="wm-modal-header">
      <span>${isNew ? '作業者を追加' : '作業者を編集'}</span>
      <button class="wm-modal-close" id="wm-close">✕</button>
    </div>
    <div class="wm-modal-body">
      <div class="wm-field">
        <label>社員ID <span class="wm-required">*</span></label>
        <input id="wm-empid" class="form-control" type="text"
          value="${escHtml(worker?.employee_id ?? '')}"
          ${isNew ? '' : 'readonly'}>
      </div>
      <div class="wm-field wm-field-row">
        <div class="wm-field">
          <label>苗字 <span class="wm-required">*</span></label>
          <input id="wm-last" class="form-control" type="text"
            value="${escHtml(worker?.last_name ?? '')}">
        </div>
        <div class="wm-field">
          <label>名前 <span class="wm-required">*</span></label>
          <input id="wm-first" class="form-control" type="text"
            value="${escHtml(worker?.first_name ?? '')}">
        </div>
      </div>
      <div class="wm-field">
        <label>電話番号</label>
        <input id="wm-phone" class="form-control" type="tel"
          value="${escHtml(worker?.phone ?? '')}" placeholder="例: 090-1234-5678">
      </div>
      <div class="wm-field">
        <label>${isNew ? 'パスワード' : 'パスワード変更（空欄で変更なし）'}
          ${isNew ? '<span class="wm-required">*</span>' : ''}
        </label>
        <input id="wm-pw" class="form-control" type="password"
          placeholder="${isNew ? 'パスワードを入力' : '変更する場合のみ入力'}">
      </div>
      <div class="wm-field">
        <label class="wm-check-label">
          <input type="checkbox" id="wm-foreman"
            ${worker?.is_foreman_qualified ? 'checked' : ''}>
          職長資格あり
        </label>
      </div>
      <div class="wm-error" id="wm-err" style="display:none"></div>
    </div>
    <div class="wm-modal-footer">
      <button class="wm-cancel btn" id="wm-cancel-btn">キャンセル</button>
      <button class="btn btn-primary" id="wm-save-btn">保存</button>
    </div>`;

  overlay.style.display = 'flex';

  document.getElementById('wm-close').addEventListener('click', closeModal);
  document.getElementById('wm-cancel-btn').addEventListener('click', closeModal);
  overlay.addEventListener('click', e => { if (e.target === overlay) closeModal(); });
  document.getElementById('wm-save-btn').addEventListener('click', saveWorker);
}

function closeModal() {
  document.getElementById('wm-overlay').style.display = 'none';
}

async function saveWorker() {
  const empId    = document.getElementById('wm-empid').value.trim();
  const lastName  = document.getElementById('wm-last').value.trim();
  const firstName = document.getElementById('wm-first').value.trim();
  const phone     = document.getElementById('wm-phone').value.trim() || null;
  const password  = document.getElementById('wm-pw').value;
  const errEl     = document.getElementById('wm-err');
  const saveBtn   = document.getElementById('wm-save-btn');
  const isNew     = _editId === null;

  if (!lastName || !firstName || (isNew && (!empId || !password))) {
    errEl.textContent = isNew
      ? '社員ID・苗字・名前・パスワードは必須です'
      : '苗字・名前は必須です';
    errEl.style.display = 'block';
    return;
  }

  saveBtn.disabled    = true;
  saveBtn.textContent = '保存中…';
  errEl.style.display = 'none';

  try {
    const isForemanQualified = document.getElementById('wm-foreman')?.checked ?? false;
    const data = {
      employee_id: empId, last_name: lastName, first_name: firstName,
      phone, is_foreman_qualified: isForemanQualified,
    };
    if (password) data.password = password;

    if (isNew) {
      await apiCreateWorker(data);
    } else {
      await apiUpdateWorker(_editId, data);
    }

    _workers = await apiGetWorkers().catch(() => []) ?? [];
    closeModal();
    const root = document.getElementById('workers-root');
    if (root) render(root);
  } catch (err) {
    errEl.textContent    = err.message;
    errEl.style.display  = 'block';
    saveBtn.disabled     = false;
    saveBtn.textContent  = '保存';
  }
}
