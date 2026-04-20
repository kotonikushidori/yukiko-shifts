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
// silent: true のときはローディングスピナーを出さずにデータだけ更新する
export async function loadBoard({ silent = false } = {}) {
  if (!silent) {
    st.loading = true;
    renderAll();
  }

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

  // 表示する現場: siteList の順を基準に固定し、末尾にアサインのみの現場を追加
  let sites;
  if (st.siteList.length > 0) {
    const withAssign = new Set(Object.keys(grouped).map(Number));
    // siteList 順を優先（active + アサインあり）
    const inList = st.siteList.filter(s => s.status === 'active' || withAssign.has(s.id));
    const inListIds = new Set(inList.map(s => s.id));
    // siteList にない現場（削除済み等）を末尾に追加
    const extra = [...withAssign]
      .filter(id => !inListIds.has(id))
      .map(id => ({ id, name: grouped[String(id)]?.name ?? `現場#${id}` }));
    sites = [...inList, ...extra];
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

  // 週に表示する現場: siteList の順を基準に固定し、末尾にアサインのみの現場を追加
  // ※ grouped を先に入れると DnD 後にアサインの変化でソート順が変わってしまうため
  const siteIdSet = new Set([
    ...st.siteList.filter(s => s.status === 'active').map(s => s.id),
    ...Object.keys(grouped).map(Number),
  ]);
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
    // 翌日の日付を計算（週の範囲外でも可）
    const nextDay = new Date(d);
    nextDay.setDate(d.getDate() + 1);
    const nextDs = fmtDate(nextDay);
    const copyBtn = !st.locked
      ? `<button class="btn-copy-date" data-from="${ds}" data-to="${nextDs}"
             title="${fmtMonthDay(d)} のシフトを ${fmtMonthDay(nextDay)} にコピー">翌日→</button>`
      : '';
    return `<th class="${cls}">
      <span class="day-date">${fmtMonthDay(d)}</span>
      <span class="day-dow">（${dow}）</span>
      ${holidayName ? `<span class="day-holiday">${escHtml(holidayName)}</span>` : ''}
      ${copyBtn}
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
  // 翌日コピーボタン
  document.querySelectorAll('.btn-copy-date').forEach(btn => {
    btn.addEventListener('click', e => {
      e.stopPropagation();
      copyDateShifts(btn.dataset.from, btn.dataset.to);
    });
  });

  // ＋ボタン → 一括設定モーダルを開く
  document.querySelectorAll('.btn-add-assign').forEach(btn => {
    btn.addEventListener('click', () =>
      openBulkModal(Number(btn.dataset.site), btn.dataset.date));
  });

  // 週表示バッジ・カンバンカード本体クリック → 同じく一括設定モーダル
  document.querySelectorAll('.week-badge, .kanban-card').forEach(el => {
    el.addEventListener('click', e => {
      if (e.target.closest('.badge-del, .kcard-del')) return; // 削除ボタンは別処理
      openBulkModal(Number(el.dataset.siteId), el.dataset.date);
    });
  });

  // 削除ボタン（バッジ内）
  document.querySelectorAll('.badge-del, .kcard-del').forEach(btn => {
    btn.addEventListener('click', async e => {
      e.stopPropagation();
      const id = Number(btn.dataset.id);
      if (!confirm('このアサインを削除しますか？')) return;
      try {
        await apiDeleteAssign(id);
        showToast('削除しました', 'success');
        await loadBoard({ silent: true });
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
        await loadBoard({ silent: true });
      } catch (err) {
        showToast(err.message, 'error');
      }
    });
  });
}

// ─── Bulk Assignment Modal ───────────────────────────────────
// 現場×日付に対して複数作業者を一括設定するモーダル
let _bulkSlot = 'AM';

function dateFmt(date) {
  const [y, m, d] = date.split('-');
  return `${y}年${parseInt(m)}月${parseInt(d)}日`;
}

// 指定セルの現在のアサイン一覧を返す
function cellAssigns(siteId, date) {
  return st.assignments.filter(a =>
    Number(a.site_id) === Number(siteId) && parseWorkDate(a.work_date) === date
  );
}

// モーダルのメインコンテンツ HTML を組み立てる（再描画にも使用）
function buildBulkBody(siteId, date) {
  const current = cellAssigns(siteId, date);
  const assignedIds = new Set(current.map(a => Number(a.user_id)));

  // 現在のアサイン
  const currentHTML = current.length === 0
    ? `<p class="bulk-empty">まだアサインがありません</p>`
    : current.map(a => {
        const name = escHtml(st.workerDispMap[a.user_id] ?? a.user_name ?? `ID:${a.user_id}`);
        const cls  = SLOT_CLS[a.time_slot] ?? 'badge-am';
        return `
          <div class="bulk-assign-row">
            <span class="badge ${cls} bulk-badge">${a.time_slot}</span>
            <span class="bulk-assign-name">${name}</span>
            <button class="bulk-del" data-assign-id="${a.id}" title="削除">×</button>
          </div>`;
      }).join('');

  // 追加できる作業者（いずれかのスロットでアサイン済みの人を除く）
  const available = st.workers.filter(w => !assignedIds.has(Number(w.id)));
  const workerListHTML = available.length === 0
    ? `<p class="bulk-empty">全員アサイン済みです</p>`
    : available.map(w => {
        const name = escHtml(st.workerDispMap[w.id] ?? w.name);
        return `
          <label class="bulk-worker-row">
            <input type="checkbox" class="bulk-check" value="${w.id}">
            <span class="bulk-worker-name">${name}</span>
          </label>`;
      }).join('');

  const slotBtns = ['AM', 'PM', 'ALL'].map(s =>
    `<button class="slot-opt${s === _bulkSlot ? ' active' : ''}" data-slot="${s}">${s}</button>`
  ).join('');

  return `
    <div class="bulk-section">
      <div class="bulk-section-label">現在のアサイン</div>
      <div class="bulk-current-list" id="bulk-current">${currentHTML}</div>
    </div>
    <div class="bulk-section">
      <div class="bulk-section-label">作業者を追加</div>
      <div class="bulk-slot-row">
        <span class="bulk-slot-label">時間帯</span>
        <div class="slot-group" id="bulk-slot-group">${slotBtns}</div>
      </div>
      <div class="bulk-worker-list" id="bulk-worker-list">${workerListHTML}</div>
    </div>
    <div class="modal-error" id="bulk-err"></div>`;
}

function openBulkModal(siteId, date) {
  // 既存モーダルがあれば閉じる
  document.getElementById('bulk-modal')?.remove();

  const siteName = st.siteMap[siteId] ?? `現場#${siteId}`;

  const el = document.createElement('div');
  el.className = 'modal-overlay';
  el.id = 'bulk-modal';
  el.innerHTML = `
    <div class="modal bulk-modal" role="dialog" aria-modal="true">
      <div class="modal-header">
        <span class="modal-title">${escHtml(siteName)}</span>
        <span class="bulk-modal-date">${dateFmt(date)}</span>
        <button class="modal-close" id="bulk-close">×</button>
      </div>
      <div class="modal-body" id="bulk-body">
        ${buildBulkBody(siteId, date)}
      </div>
      <div class="modal-footer">
        <button class="btn btn-secondary" id="bulk-cancel">閉じる</button>
        <button class="btn btn-primary" id="bulk-submit">選択した作業者を追加</button>
      </div>
    </div>`;

  document.body.appendChild(el);
  bindBulkModal(el, siteId, date, siteName);
}

function bindBulkModal(el, siteId, date, siteName) {
  const close = () => { el.remove(); };

  el.querySelector('#bulk-close').addEventListener('click', close);
  el.querySelector('#bulk-cancel').addEventListener('click', close);
  el.addEventListener('click', e => { if (e.target === el) close(); });

  // 時間帯切替
  el.querySelectorAll('.slot-opt').forEach(btn => {
    btn.addEventListener('click', () => {
      _bulkSlot = btn.dataset.slot;
      el.querySelectorAll('.slot-opt').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
    });
  });

  // 個別削除ボタン
  bindBulkDeletes(el, siteId, date, siteName);

  // 一括追加ボタン
  el.querySelector('#bulk-submit').addEventListener('click', () =>
    bulkSubmit(el, siteId, date, siteName));
}

// 個別削除ボタンのバインド（再描画後も呼ばれる）
function bindBulkDeletes(el, siteId, date, siteName) {
  el.querySelectorAll('.bulk-del').forEach(btn => {
    btn.addEventListener('click', async e => {
      e.stopPropagation();
      const id = Number(btn.dataset.assignId);
      btn.disabled = true;
      try {
        await apiDeleteAssign(id);
        st.assignments = st.assignments.filter(a => a.id !== id);
        refreshBulkBody(el, siteId, date, siteName);
        renderAll(); // ボード側も更新
      } catch (err) {
        showToast(err.message, 'error');
        btn.disabled = false;
      }
    });
  });
}

// モーダルのボディのみ再描画（モーダルを閉じない）
function refreshBulkBody(el, siteId, date, siteName) {
  el.querySelector('#bulk-body').innerHTML = buildBulkBody(siteId, date);
  // 再バインド
  el.querySelectorAll('.slot-opt').forEach(btn => {
    btn.addEventListener('click', () => {
      _bulkSlot = btn.dataset.slot;
      el.querySelectorAll('.slot-opt').forEach(b => b.classList.remove('active'));
      btn.classList.add('active');
    });
  });
  bindBulkDeletes(el, siteId, date, siteName);
  el.querySelector('#bulk-submit').addEventListener('click', () =>
    bulkSubmit(el, siteId, date, siteName));
}

async function bulkSubmit(el, siteId, date, siteName) {
  const errEl     = el.querySelector('#bulk-err');
  const submitBtn = el.querySelector('#bulk-submit');

  const checked = [...el.querySelectorAll('.bulk-check:checked')];
  if (checked.length === 0) {
    errEl.textContent = '作業者を1人以上選んでください';
    errEl.classList.add('visible');
    return;
  }

  submitBtn.disabled    = true;
  submitBtn.textContent = '追加中…';
  errEl.classList.remove('visible');

  const results = await Promise.allSettled(
    checked.map(cb => apiCreateAssign(Number(siteId), date, Number(cb.value), _bulkSlot))
  );

  const failed  = results.filter(r => r.status === 'rejected');
  const success = results.filter(r => r.status === 'fulfilled').length;

  if (failed.length > 0) {
    errEl.textContent = failed.map(r => r.reason.message).join(' / ');
    errEl.classList.add('visible');
  }
  if (success > 0) {
    showToast(`${success}名を追加しました`, 'success');
    await loadBoard({ silent: true });          // データ再取得
    refreshBulkBody(el, siteId, date, siteName); // モーダル内を更新
  }

  submitBtn.disabled    = false;
  submitBtn.textContent = '選択した作業者を追加';
}

// ─── Copy Date Shifts ────────────────────────────────────────
async function copyDateShifts(fromDate, toDate) {
  const assigns = st.assignments.filter(a => parseWorkDate(a.work_date) === fromDate);

  if (assigns.length === 0) {
    showToast('コピー元にアサインがありません', 'error');
    return;
  }

  const [fy, fm, fd] = fromDate.split('-');
  const [ty, tm, td] = toDate.split('-');
  const fromDisp = `${parseInt(fm)}/${parseInt(fd)}`;
  const toDisp   = `${parseInt(tm)}/${parseInt(td)}`;

  if (!confirm(`${fromDisp} の全シフト（${assigns.length}件）を ${toDisp} にコピーしますか？`)) return;

  const results = await Promise.allSettled(
    assigns.map(a => apiCreateAssign(a.site_id, toDate, a.user_id, a.time_slot))
  );

  const success = results.filter(r => r.status === 'fulfilled').length;
  const skipped = results.filter(r => r.status === 'rejected').length; // 重複など

  if (success > 0) {
    const msg = skipped > 0
      ? `${success}件コピーしました（${skipped}件は重複のためスキップ）`
      : `${success}件コピーしました`;
    showToast(msg, 'success');
    await loadBoard({ silent: true });
  } else {
    showToast('コピーできませんでした（全件が重複または重複エラー）', 'error');
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
