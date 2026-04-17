// board.js — 管理者シフトボード

import { apiGetBoard, apiGetSites, apiCreateAssign, apiDeleteAssign,
         apiGetLockStatus, apiLockMonth, apiUnlockMonth, apiGetWorkers } from './api.js';
import { HOLIDAYS } from './holidays.js';

// ─── State ───────────────────────────────────────────────────
const st = {
  viewMode: 'week',
  currentDate: new Date(),
  assignments: [],
  siteList: [],     // GET /api/sites から取得した全現場
  siteMap: {},      // { siteId: siteName }
  workerMap: {},    // { userId: userName (フルネーム) }
  workerDispMap: {},// { userId: 表示名（苗字 or 苗字+頭文字） }
  workers: [],      // 作業者マスタ全件
  locked: false,
  loading: false,
};

// ─── Worker Display Names ────────────────────────────────────
// 苗字が重複するときだけ名前の頭1文字を付加する
function buildWorkerDisplayNames(workers) {
  const lastNameCount = {};
  for (const w of workers) {
    const ln = w.last_name || w.name;
    lastNameCount[ln] = (lastNameCount[ln] || 0) + 1;
  }
  const disp = {};
  for (const w of workers) {
    const ln = w.last_name || w.name;
    if (lastNameCount[ln] > 1 && w.first_name) {
      disp[w.id] = ln + w.first_name[0];   // 例: 田中一, 田中二
    } else {
      disp[w.id] = ln;                      // 例: 佐藤
    }
  }
  return disp;
}

// DnD 転送中の情報
let _drag = null; // { assignId, userId, slot, fromSiteId, fromDate }

// ─── Date Utilities ──────────────────────────────────────────
const DOW_JA = ['日', '月', '火', '水', '木', '金', '土'];

export function fmtDate(d) {
  const y  = d.getFullYear();
  const m  = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${dd}`;
}

export function parseWorkDate(isoStr) {
  return String(isoStr).substring(0, 10);
}

export function getWeekDates(ref) {
  const day  = ref.getDay();
  const diff = day === 0 ? -6 : 1 - day;
  const mon  = new Date(ref);
  mon.setDate(ref.getDate() + diff);
  return Array.from({ length: 7 }, (_, i) => {
    const d = new Date(mon);
    d.setDate(mon.getDate() + i);
    return d;
  });
}

function fmtMonthDay(d) { return `${d.getMonth() + 1}/${d.getDate()}`; }
function fmtFull(d)     { return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日`; }

// ─── Utilities ──────────────────────────────────────────────
function escHtml(str) {
  return String(str ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

// ─── Data Grouping ────────────────────────────────────────────
function buildMaps(assignments) {
  for (const a of assignments) {
    if (a.site_id && a.site_name) st.siteMap[a.site_id] = a.site_name;
    if (a.user_id && a.user_name) st.workerMap[a.user_id] = a.user_name;
  }
}

/** 週表示: { siteId: { name, days: { 'YYYY-MM-DD': [assign] } } } */
export function groupWeek(assignments) {
  const g = {};
  for (const a of assignments) {
    const sid  = String(a.site_id);
    const date = parseWorkDate(a.work_date);
    if (!g[sid]) g[sid] = { name: a.site_name ?? `現場#${sid}`, days: {} };
    if (!g[sid].days[date]) g[sid].days[date] = [];
    g[sid].days[date].push(a);
  }
  return g;
}

/** 日表示（カンバン）: { siteId: { name, cards: [assign...] } } */
export function groupDay(assignments, dateStr) {
  const g = {};
  for (const a of assignments) {
    if (parseWorkDate(a.work_date) !== dateStr) continue;
    const sid = String(a.site_id);
    if (!g[sid]) g[sid] = { name: a.site_name ?? `現場#${sid}`, cards: [] };
    g[sid].cards.push(a);
  }
  return g;
}

// ─── Load ────────────────────────────────────────────────────
export async function loadBoard() {
  st.loading = true;
  renderAll();

  let from, to;
  if (st.viewMode === 'week') {
    const dates = getWeekDates(st.currentDate);
    from = fmtDate(dates[0]);
    to   = fmtDate(dates[6]);
  } else {
    from = to = fmtDate(st.currentDate);
  }

  // ロック対象の年月（1日表示なら当月、週表示なら開始週の月）
  const lockY = st.viewMode === 'week'
    ? getWeekDates(st.currentDate)[0].getFullYear()
    : st.currentDate.getFullYear();
  const lockM = st.viewMode === 'week'
    ? getWeekDates(st.currentDate)[0].getMonth() + 1
    : st.currentDate.getMonth() + 1;

  try {
    const [assignments, sites, lockData, workers] = await Promise.all([
      apiGetBoard(from, to),
      apiGetSites().catch(() => []),
      apiGetLockStatus(lockY, lockM).catch(() => ({ locked: false })),
      apiGetWorkers().catch(() => []),
    ]);
    st.assignments = assignments ?? [];
    st.siteList    = sites ?? [];
    st.locked      = lockData?.locked ?? false;
    buildMaps(st.assignments);
    for (const s of st.siteList) {
      if (s.id) st.siteMap[s.id] = s.name;
    }
    // 作業者マスタから workerMap / workerDispMap を構築
    st.workers = workers ?? [];
    for (const w of st.workers) {
      if (w.id) st.workerMap[w.id] = w.last_name && w.first_name
        ? w.last_name + ' ' + w.first_name
        : w.name;
    }
    st.workerDispMap = buildWorkerDisplayNames(st.workers);
  } catch (e) {
    showToast('データ取得エラー: ' + e.message, 'error');
    st.assignments = [];
  }

  st.loading = false;
  renderAll();
}

// ─── Card / Badge Renders ────────────────────────────────────
const SLOT_CLS = { AM: 'badge-am', PM: 'badge-pm', ALL: 'badge-all' };

/** カンバン用カード（1日表示） */
function renderKanbanCard(a) {
  const cls  = SLOT_CLS[a.time_slot] ?? 'badge-am';
  const name = escHtml(st.workerDispMap[a.user_id] ?? a.user_name ?? `ID:${a.user_id}`);
  const date = parseWorkDate(a.work_date);
  return `
    <div class="kanban-card ${cls}" draggable="true"
         data-assign-id="${a.id}"
         data-user-id="${a.user_id}"
         data-slot="${a.time_slot}"
         data-site-id="${a.site_id}"
         data-date="${date}">
      <span class="kcard-slot">${a.time_slot}</span>
      <span class="kcard-name">${name}</span>
      <button class="kcard-del" data-id="${a.id}" title="削除">×</button>
    </div>`;
}

/** 週表示用バッジ（ドラッグ可能） */
function renderWeekBadge(a) {
  const cls  = SLOT_CLS[a.time_slot] ?? 'badge-am';
  const name = escHtml(st.workerDispMap[a.user_id] ?? a.user_name ?? `ID:${a.user_id}`);
  const date = parseWorkDate(a.work_date);
  return `
    <span class="badge week-badge ${cls}" draggable="true"
          data-assign-id="${a.id}"
          data-user-id="${a.user_id}"
          data-slot="${a.time_slot}"
          data-site-id="${a.site_id}"
          data-date="${date}"
          title="${name}（${a.time_slot}）">
      <span class="badge-slot-label">${a.time_slot}</span>
      ${name}
      <button class="badge-del" data-id="${a.id}" title="削除">×</button>
    </span>`;
}

// ─── Day Kanban ──────────────────────────────────────────────
function renderKanban() {
  const dateStr = fmtDate(st.currentDate);
  const grouped = groupDay(st.assignments, dateStr);

  // 表示する現場: siteList の稼働中 + アサインがある現場（削除済みなど）
  let sites;
  if (st.siteList.length > 0) {
    const withAssign = new Set(Object.keys(grouped).map(Number));
    sites = st.siteList.filter(s => s.status === 'active' || withAssign.has(s.id));
  } else {
    sites = Object.entries(grouped).map(([id, v]) => ({ id: Number(id), name: v.name }));
  }

  if (sites.length === 0) {
    return `
      <div class="empty-state">
        <div class="empty-state-icon">📋</div>
        <h2>この日のシフトデータがありません</h2>
        <p>現場マスタに稼働中の現場を登録してください</p>
      </div>`;
  }

  const cols = sites.map(site => {
    const sid   = String(site.id);
    const group = grouped[sid];
    const cards = group ? group.cards.map(renderKanbanCard).join('') : '';
    const count = group ? group.cards.length : 0;
    return `
      <div class="kanban-col">
        <div class="kanban-col-header">
          <span class="kanban-col-name">${escHtml(site.name)}</span>
          <span class="kanban-col-count">${count}名</span>
        </div>
        <div class="kanban-drop-zone" data-site-id="${sid}" data-date="${dateStr}">
          ${cards}
        </div>
        <div class="kanban-col-footer">
          <button class="btn-add-assign" data-site="${sid}" data-date="${dateStr}"
                  title="アサイン追加">＋ 追加</button>
        </div>
      </div>`;
  }).join('');

  return `<div class="kanban-board">${cols}</div>`;
}

// ─── Week View ────────────────────────────────────────────────
function renderWeekTable() {
  const dates   = getWeekDates(st.currentDate);
  const today   = fmtDate(new Date());
  const grouped = groupWeek(st.assignments);

  // 週に表示する現場: 稼働中の全現場 + アサインのある現場
  const siteIdSet = new Set(Object.keys(grouped).map(Number));
  st.siteList.filter(s => s.status === 'active').forEach(s => siteIdSet.add(s.id));
  const siteIds = [...siteIdSet].map(String);

  const thCols = dates.map(d => {
    const ds          = fmtDate(d);
    const dow         = DOW_JA[d.getDay()];
    const holidayName = HOLIDAYS[ds] ?? null;
    const cls         = [
      ds === today     ? 'col-today'   : '',
      d.getDay() === 6 ? 'col-sat'     : '',
      d.getDay() === 0 || holidayName  ? 'col-sun' : '',
      holidayName      ? 'col-holiday' : '',
    ].filter(Boolean).join(' ');
    return `<th class="${cls}">
      <span class="day-date">${fmtMonthDay(d)}</span>
      <span class="day-dow">（${dow}）</span>
      ${holidayName ? `<span class="day-holiday">${escHtml(holidayName)}</span>` : ''}
    </th>`;
  }).join('');

  let rows;
  if (siteIds.length === 0) {
    rows = `<tr><td colspan="${dates.length + 1}">
      <div class="empty-state">
        <div class="empty-state-icon">📋</div>
        <h2>シフトデータがありません</h2>
        <p>+ ボタンでアサインを追加してください</p>
      </div>
    </td></tr>`;
  } else {
    rows = siteIds.map(sid => {
      const siteName = st.siteMap[sid] ?? `現場#${sid}`;
      const siteData = grouped[sid];
      const cells = dates.map(d => {
        const ds      = fmtDate(d);
        const assigns = siteData?.days[ds] ?? [];
        const isToday = ds === today;
        const badges  = assigns.map(renderWeekBadge).join('');
        const addBtn  = `<button class="btn-add-assign" data-site="${sid}" data-date="${ds}" title="追加">+</button>`;
        return `<td class="${isToday ? 'col-today' : ''}">
          <div class="cell-content week-drop-zone" data-site-id="${sid}" data-date="${ds}">
            ${badges}${addBtn}
          </div>
        </td>`;
      }).join('');
      return `<tr>
        <td>
          <span class="site-cell-name">${escHtml(siteName)}</span>
          <span class="site-cell-id">#${sid}</span>
        </td>
        ${cells}
      </tr>`;
    }).join('');
  }

  return `
    <div class="shift-table-wrap">
      <table class="shift-table">
        <thead><tr><th>現場</th>${thCols}</tr></thead>
        <tbody>${rows}</tbody>
      </table>
    </div>`;
}

// ─── Toolbar ─────────────────────────────────────────────────
function renderToolbar() {
  let label;
  if (st.viewMode === 'week') {
    const dates = getWeekDates(st.currentDate);
    label = `${dates[0].getFullYear()}年 ${fmtMonthDay(dates[0])}（月）〜 ${fmtMonthDay(dates[6])}（日）`;
  } else {
    const d = st.currentDate;
    label = `${fmtFull(d)}（${DOW_JA[d.getDay()]}）`;
  }
  return `
    <div class="board-toolbar">
      <div class="view-toggle">
        <button id="btn-day"  class="${st.viewMode === 'day'  ? 'active' : ''}">1日表示</button>
        <button id="btn-week" class="${st.viewMode === 'week' ? 'active' : ''}">1週表示</button>
      </div>
      <div class="toolbar-sep"></div>
      <div class="date-nav">
        <button class="btn-icon" id="btn-prev" title="前へ">‹</button>
        <span class="date-label">${label}</span>
        <button class="btn-icon" id="btn-next" title="次へ">›</button>
      </div>
      <button class="btn-today" id="btn-today">今日</button>
      <div class="toolbar-space"></div>
      ${st.locked
        ? `<button class="btn-lock active" id="btn-lock" title="ロック解除">🔒 確定済み</button>`
        : `<button class="btn-lock" id="btn-lock" title="この月の希望入力を締め切る">🔓 ロック</button>`
      }
    </div>`;
}

// ─── Main Render ─────────────────────────────────────────────
function renderAll() {
  const root = document.getElementById('board-root');
  if (!root) return;

  if (st.loading) {
    root.innerHTML = `
      ${renderToolbar()}
      <div class="board-container">
        <div class="loading-screen"><div class="spinner"></div></div>
      </div>`;
    bindToolbar();
    return;
  }

  const content = st.viewMode === 'week' ? renderWeekTable() : renderKanban();
  root.innerHTML = `
    ${renderToolbar()}
    <div class="board-container">${content}</div>`;

  bindToolbar();
  bindCells();
  bindDrag();
}

function bindToolbar() {
  document.getElementById('btn-day')?.addEventListener('click', () => {
    if (st.viewMode === 'day') return;
    st.viewMode = 'day'; loadBoard();
  });
  document.getElementById('btn-week')?.addEventListener('click', () => {
    if (st.viewMode === 'week') return;
    st.viewMode = 'week'; loadBoard();
  });
  document.getElementById('btn-prev')?.addEventListener('click',  () => navDate(-1));
  document.getElementById('btn-next')?.addEventListener('click',  () => navDate(1));
  document.getElementById('btn-today')?.addEventListener('click', () => {
    st.currentDate = new Date(); loadBoard();
  });

  document.getElementById('btn-lock')?.addEventListener('click', async () => {
    const d = st.viewMode === 'week' ? getWeekDates(st.currentDate)[0] : st.currentDate;
    const y = d.getFullYear(), m = d.getMonth() + 1;
    try {
      if (st.locked) {
        if (!confirm(`${y}年${m}月のロックを解除しますか？`)) return;
        await apiUnlockMonth(y, m);
        showToast(`${m}月のロックを解除しました`, 'success');
      } else {
        if (!confirm(`${y}年${m}月の希望入力を締め切り（ロック）しますか？`)) return;
        await apiLockMonth(y, m);
        showToast(`${m}月をロックしました`, 'success');
      }
      loadBoard();
    } catch (err) {
      showToast(err.message, 'error');
    }
  });
}

function navDate(dir) {
  const d = new Date(st.currentDate);
  d.setDate(d.getDate() + (st.viewMode === 'week' ? 7 : 1) * dir);
  st.currentDate = d;
  loadBoard();
}

function bindCells() {
  document.querySelectorAll('.btn-add-assign').forEach(btn => {
    btn.addEventListener('click', () => openAddModal(Number(btn.dataset.site), btn.dataset.date, null));
  });

  document.querySelectorAll('.badge-del, .kcard-del').forEach(btn => {
    btn.addEventListener('click', async e => {
      e.stopPropagation();
      const id = Number(btn.dataset.id);
      if (!confirm('このアサインを削除しますか？')) return;
      try {
        await apiDeleteAssign(id);
        showToast('削除しました', 'success');
        await loadBoard();
      } catch (err) {
        showToast(err.message, 'error');
      }
    });
  });
}

// ─── Drag & Drop ─────────────────────────────────────────────
function bindDrag() {
  // カンバンカード
  document.querySelectorAll('.kanban-card[draggable]').forEach(card => {
    card.addEventListener('dragstart', e => {
      _drag = {
        assignId:   Number(card.dataset.assignId),
        userId:     Number(card.dataset.userId),
        slot:       card.dataset.slot,
        fromSiteId: Number(card.dataset.siteId),
        fromDate:   card.dataset.date,
      };
      e.dataTransfer.effectAllowed = 'move';
      card.classList.add('dragging');
    });
    card.addEventListener('dragend', () => card.classList.remove('dragging'));
  });

  // 週表示バッジ
  document.querySelectorAll('.week-badge[draggable]').forEach(badge => {
    badge.addEventListener('dragstart', e => {
      _drag = {
        assignId:   Number(badge.dataset.assignId),
        userId:     Number(badge.dataset.userId),
        slot:       badge.dataset.slot,
        fromSiteId: Number(badge.dataset.siteId),
        fromDate:   badge.dataset.date,
      };
      e.dataTransfer.effectAllowed = 'move';
      badge.classList.add('dragging');
    });
    badge.addEventListener('dragend', () => badge.classList.remove('dragging'));
  });

  document.querySelectorAll('.kanban-drop-zone, .week-drop-zone').forEach(zone => {
    zone.addEventListener('dragover', e => {
      if (!_drag) return;
      e.preventDefault();
      e.dataTransfer.dropEffect = 'move';
      zone.classList.add('drag-over');
    });
    zone.addEventListener('dragleave', e => {
      if (!zone.contains(e.relatedTarget)) zone.classList.remove('drag-over');
    });
    zone.addEventListener('drop', async e => {
      e.preventDefault();
      zone.classList.remove('drag-over');
      if (!_drag) return;

      const toSiteId = Number(zone.dataset.siteId);
      const toDate   = zone.dataset.date;

      if (toSiteId === _drag.fromSiteId && toDate === _drag.fromDate) {
        _drag = null; return;
      }

      const drag = _drag;
      _drag = null;

      try {
        await apiDeleteAssign(drag.assignId);
        await apiCreateAssign(toSiteId, toDate, drag.userId, drag.slot);
        showToast('移動しました', 'success');
        await loadBoard();
      } catch (err) {
        showToast(err.message, 'error');
      }
    });
  });
}

// ─── Add Assignment Modal ─────────────────────────────────────
let _modalCtx = null;

function openAddModal(siteId, date, presetSlot) {
  _modalCtx = { siteId, date, slot: presetSlot ?? 'AM' };

  const siteName = st.siteMap[siteId] ?? `現場 #${siteId}`;
  const [y, m, d] = date.split('-');
  const dateDisp  = `${y}年${parseInt(m)}月${parseInt(d)}日`;

  const workerEntries = Object.entries(st.workerMap);
  let workerInput;
  if (workerEntries.length > 0) {
    const opts = workerEntries.map(([id, name]) =>
      `<option value="${id}">${escHtml(name)}</option>`
    ).join('');
    workerInput = `
      <select class="form-select" id="modal-worker">
        <option value="">-- 作業者を選択 --</option>
        ${opts}
      </select>`;
  } else {
    workerInput = `
      <input type="number" class="form-control" id="modal-worker"
        placeholder="作業者 ID を入力（例: 4）" min="1">
      <p style="font-size:12px;color:#64748b;margin-top:4px">
        ※ボード上にデータがないため直接IDを入力してください
      </p>`;
  }

  const slotBtns = ['AM', 'PM', 'ALL'].map(s => {
    const active = s === _modalCtx.slot ? 'active' : '';
    return `<button class="slot-opt ${active}" data-slot="${s}">${s}</button>`;
  }).join('');

  const el = document.createElement('div');
  el.className = 'modal-overlay';
  el.id = 'add-modal';
  el.innerHTML = `
    <div class="modal" role="dialog" aria-modal="true">
      <div class="modal-header">
        <span class="modal-title">アサイン追加</span>
        <button class="modal-close" id="modal-close">×</button>
      </div>
      <div class="modal-body">
        <div class="modal-info-box">
          <strong>${escHtml(siteName)}</strong> &nbsp;/&nbsp; ${dateDisp}
        </div>
        <div class="form-group">
          <label>作業者</label>
          ${workerInput}
        </div>
        <div class="form-group">
          <label>時間帯</label>
          <div class="slot-group">${slotBtns}</div>
        </div>
        <div class="modal-error" id="modal-err"></div>
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" id="modal-cancel">キャンセル</button>
        <button class="btn btn-primary" id="modal-submit" style="width:auto">追 加</button>
      </div>
    </div>`;

  document.body.appendChild(el);

  const close = () => { document.getElementById('add-modal')?.remove(); _modalCtx = null; };

  el.querySelector('#modal-close').addEventListener('click', close);
  el.querySelector('#modal-cancel').addEventListener('click', close);
  el.addEventListener('click', e => { if (e.target === el) close(); });

  el.querySelectorAll('.slot-opt').forEach(btn => {
    btn.addEventListener('click', () => {
      el.querySelectorAll('.slot-opt').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
      _modalCtx.slot = btn.dataset.slot;
    });
  });

  el.querySelector('#modal-submit').addEventListener('click', () => submitAdd(close));
}

async function submitAdd(close) {
  const workerEl  = document.getElementById('modal-worker');
  const errEl     = document.getElementById('modal-err');
  const submitBtn = document.getElementById('modal-submit');

  errEl.classList.remove('visible');

  const userId = parseInt(workerEl.value, 10);
  if (!userId || Number.isNaN(userId)) {
    errEl.textContent = '作業者を選択または入力してください';
    errEl.classList.add('visible');
    return;
  }

  submitBtn.disabled = true;
  submitBtn.textContent = '追加中…';

  try {
    await apiCreateAssign(_modalCtx.siteId, _modalCtx.date, userId, _modalCtx.slot);
    close();
    showToast('アサインを追加しました', 'success');
    await loadBoard();
  } catch (err) {
    errEl.textContent = err.message;
    errEl.classList.add('visible');
    submitBtn.disabled = false;
    submitBtn.textContent = '追 加';
  }
}

// ─── Toast ───────────────────────────────────────────────────
export function showToast(message, type = 'success') {
  let container = document.querySelector('.toast-container');
  if (!container) {
    container = document.createElement('div');
    container.className = 'toast-container';
    document.body.appendChild(container);
  }
  const icon  = type === 'success' ? '✓' : '⚠';
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `<span class="toast-icon">${icon}</span><span>${escHtml(message)}</span>`;
  container.appendChild(toast);
  setTimeout(() => toast.remove(), 3500);
}

// ─── Init ─────────────────────────────────────────────────────
export function initBoard() {
  loadBoard();
}
