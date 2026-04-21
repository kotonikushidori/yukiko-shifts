-- 007_seed_foreman_assignments.sql
-- サンプル職長アサイン（シードシフトデータ対応: 4月14日〜18日）
-- ※ INSERT OR IGNORE なので再実行でもエラーにならない
-- ※ 4/15 南平岸４条 は w007(伊藤浩二) を職長に設定（動作確認用）

WITH
w AS (
  SELECT id, employee_id FROM users WHERE tenant_id = 1
),
s(site_id, emp_id, work_date) AS (
  SELECT 1,'w001','2026-04-14' UNION ALL
  SELECT 1,'w007','2026-04-15' UNION ALL
  SELECT 1,'w001','2026-04-16' UNION ALL
  SELECT 1,'w001','2026-04-17' UNION ALL
  SELECT 1,'w001','2026-04-18' UNION ALL
  SELECT 2,'w020','2026-04-16' UNION ALL
  SELECT 3,'w030','2026-04-18' UNION ALL
  SELECT 4,'w036','2026-04-14' UNION ALL
  SELECT 4,'w036','2026-04-16' UNION ALL
  SELECT 5,'w046','2026-04-14' UNION ALL
  SELECT 5,'w046','2026-04-16' UNION ALL
  SELECT 5,'w046','2026-04-17' UNION ALL
  SELECT 6,'w057','2026-04-15'
)
INSERT OR IGNORE INTO foreman_assignments (tenant_id, site_id, work_date, user_id, is_manual)
SELECT 1, s.site_id, s.work_date, w.id, 1
FROM s JOIN w ON w.employee_id = s.emp_id;
