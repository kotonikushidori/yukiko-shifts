-- 008_seed_foreman_priorities.sql
-- 現場ごとの職長優先順位を seed（第1位・第2位）
-- priority_order: 0 = 第1位、1 = 第2位
-- INSERT OR IGNORE なので再実行でもエラーにならない

WITH
w AS (
  SELECT id, employee_id FROM users WHERE tenant_id = 1
),
p(site_id, emp_id, priority_order) AS (
  -- 南平岸４条 (site_id=1): 伊藤浩二 → 田中一郎
  SELECT 1,'w007',0 UNION ALL
  SELECT 1,'w001',1 UNION ALL
  -- 島松隊舎 (site_id=2): 吉田修 → 小林拓也
  SELECT 2,'w020',0 UNION ALL
  SELECT 2,'w016',1 UNION ALL
  -- CR澄川 (site_id=3): 井上洋 → 佐々木勇
  SELECT 3,'w030',0 UNION ALL
  SELECT 3,'w024',1 UNION ALL
  -- すすきの駅前 (site_id=4): 斎藤茂 → 山本裕介
  SELECT 4,'w036',0 UNION ALL
  SELECT 4,'w011',1 UNION ALL
  -- トヨタ分室 (site_id=5): 池田敏雄 → 鈴木太郎
  SELECT 5,'w046',0 UNION ALL
  SELECT 5,'w004',1 UNION ALL
  -- 北広工業 (site_id=6): 後藤康夫 → 遠藤栄一
  SELECT 6,'w057',0 UNION ALL
  SELECT 6,'w070',1
)
INSERT OR IGNORE INTO site_foreman_priorities (site_id, user_id, priority_order)
SELECT p.site_id, w.id, p.priority_order
FROM p JOIN w ON w.employee_id = p.emp_id;
