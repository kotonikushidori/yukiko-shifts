-- 006_seed_workers.sql
-- 本番規模想定: 作業者 80 名追加（合計 84 名）
-- パスワード: worker1234
-- 職長資格あり: 12 名（※ is_foreman_qualified = 1）

-- テナントの上限人数を本番規模に更新
UPDATE tenants SET max_workers = 120 WHERE id = 1;

-- ──────────────────────────────────────
-- 作業者追加（w005 ～ w084）
-- ──────────────────────────────────────
INSERT OR IGNORE INTO users
  (tenant_id, employee_id, name, last_name, first_name, role, phone, is_foreman_qualified, password_hash)
VALUES
  -- 高橋グループ
  (1,'w005','高橋健太','高橋','健太','worker','090-1001-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w006','高橋誠二','高橋','誠二','worker','090-1001-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 伊藤グループ
  (1,'w007','伊藤浩二','伊藤','浩二','worker','090-1002-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w008','伊藤和也','伊藤','和也','worker','090-1002-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 渡辺グループ
  (1,'w009','渡辺誠一','渡辺','誠一','worker','090-1003-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w010','渡辺大輔','渡辺','大輔','worker','090-1003-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 山本グループ
  (1,'w011','山本裕介','山本','裕介','worker','090-1004-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w012','山本直樹','山本','直樹','worker','090-1004-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w013','山本翔太','山本','翔太','worker','090-1004-0003',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 中村グループ
  (1,'w014','中村健一','中村','健一','worker','090-1005-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w015','中村雄一','中村','雄一','worker','090-1005-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 小林グループ
  (1,'w016','小林拓也','小林','拓也','worker','090-1006-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w017','小林光雄','小林','光雄','worker','090-1006-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 加藤グループ
  (1,'w018','加藤博之','加藤','博之','worker','090-1007-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w019','加藤俊介','加藤','俊介','worker','090-1007-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 吉田グループ
  (1,'w020','吉田修','吉田','修','worker','090-1008-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w021','吉田豊','吉田','豊','worker','090-1008-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 山田グループ
  (1,'w022','山田哲也','山田','哲也','worker','090-1009-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w023','山田亮','山田','亮','worker','090-1009-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 佐々木グループ
  (1,'w024','佐々木勇','佐々木','勇','worker','090-1010-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w025','佐々木清','佐々木','清','worker','090-1010-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 山口グループ
  (1,'w026','山口翔','山口','翔','worker','090-1011-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w027','山口正','山口','正','worker','090-1011-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 松本グループ
  (1,'w028','松本博','松本','博','worker','090-1012-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w029','松本大樹','松本','大樹','worker','090-1012-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 井上グループ
  (1,'w030','井上洋','井上','洋','worker','090-1013-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w031','井上稔','井上','稔','worker','090-1013-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 木村グループ
  (1,'w032','木村進','木村','進','worker','090-1014-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w033','木村勝','木村','勝','worker','090-1014-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 林グループ
  (1,'w034','林義雄','林','義雄','worker','090-1015-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w035','林英雄','林','英雄','worker','090-1015-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 斎藤グループ
  (1,'w036','斎藤茂','斎藤','茂','worker','090-1016-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w037','斎藤宏','斎藤','宏','worker','090-1016-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 清水グループ
  (1,'w038','清水隆','清水','隆','worker','090-1017-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w039','清水誠','清水','誠','worker','090-1017-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 山崎グループ
  (1,'w040','山崎昭','山崎','昭','worker','090-1018-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w041','山崎和夫','山崎','和夫','worker','090-1018-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 阿部グループ
  (1,'w042','阿部浩','阿部','浩','worker','090-1019-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w043','阿部功','阿部','功','worker','090-1019-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 森グループ
  (1,'w044','森良夫','森','良夫','worker','090-1020-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w045','森文雄','森','文雄','worker','090-1020-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 池田グループ
  (1,'w046','池田敏雄','池田','敏雄','worker','090-1021-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w047','池田正夫','池田','正夫','worker','090-1021-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 橋本グループ
  (1,'w048','橋本幸雄','橋本','幸雄','worker','090-1022-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w049','橋本省一','橋本','省一','worker','090-1022-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 石川グループ
  (1,'w050','石川順一','石川','順一','worker','090-1023-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w051','石川利夫','石川','利夫','worker','090-1023-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 前田グループ
  (1,'w052','前田貞夫','前田','貞夫','worker','090-1024-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w053','前田義男','前田','義男','worker','090-1024-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 岡田グループ
  (1,'w054','岡田富雄','岡田','富雄','worker','090-1025-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w055','岡田孝雄','岡田','孝雄','worker','090-1025-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 後藤グループ
  (1,'w056','後藤寿男','後藤','寿男','worker','090-1026-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w057','後藤康夫','後藤','康夫','worker','090-1026-0002',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 長谷川グループ
  (1,'w058','長谷川浩','長谷川','浩','worker','090-1027-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w059','長谷川実','長谷川','実','worker','090-1027-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 村上グループ
  (1,'w060','村上功','村上','功','worker','090-1028-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w061','村上守','村上','守','worker','090-1028-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 近藤グループ
  (1,'w062','近藤登','近藤','登','worker','090-1029-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w063','近藤武','近藤','武','worker','090-1029-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 石井グループ
  (1,'w064','石井賢一','石井','賢一','worker','090-1030-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w065','石井信雄','石井','信雄','worker','090-1030-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 藤田グループ
  (1,'w066','藤田芳雄','藤田','芳雄','worker','090-1031-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w067','藤田靖夫','藤田','靖夫','worker','090-1031-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 坂本グループ
  (1,'w068','坂本政雄','坂本','政雄','worker','090-1032-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w069','坂本安夫','坂本','安夫','worker','090-1032-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 遠藤グループ
  (1,'w070','遠藤栄一','遠藤','栄一','worker','090-1033-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w071','遠藤洋二','遠藤','洋二','worker','090-1033-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 武田グループ
  (1,'w072','武田俊雄','武田','俊雄','worker','090-1034-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w073','武田昌夫','武田','昌夫','worker','090-1034-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 西村グループ
  (1,'w074','西村光男','西村','光男','worker','090-1035-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w075','西村哲男','西村','哲男','worker','090-1035-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 松田グループ
  (1,'w076','松田克己','松田','克己','worker','090-1036-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w077','松田正雄','松田','正雄','worker','090-1036-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 原グループ
  (1,'w078','原清一','原','清一','worker','090-1037-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w079','原茂雄','原','茂雄','worker','090-1037-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 中島グループ
  (1,'w080','中島豊','中島','豊','worker','090-1038-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w081','中島博文','中島','博文','worker','090-1038-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- 福田グループ
  (1,'w082','福田道男','福田','道男','worker','090-1039-0001',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1,'w083','福田辰雄','福田','辰雄','worker','090-1039-0002',0,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),

  -- その他
  (1,'w084','菅原浩二','菅原','浩二','worker','090-1040-0001',1,'$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm');

-- ──────────────────────────────────────
-- 既存4名の is_foreman_qualified を設定
-- ──────────────────────────────────────
UPDATE users SET is_foreman_qualified = 1 WHERE employee_id = 'w001' AND tenant_id = 1; -- 田中一郎
UPDATE users SET is_foreman_qualified = 1 WHERE employee_id = 'w004' AND tenant_id = 1; -- 鈴木太郎

-- ──────────────────────────────────────
-- 職長資格あり合計: 14 名
--   既存: 田中一郎(w001), 鈴木太郎(w004)
--   新規: 伊藤浩二(w007), 山本裕介(w011), 小林拓也(w016),
--         吉田修(w020), 佐々木勇(w024), 井上洋(w030),
--         斎藤茂(w036), 池田敏雄(w046), 後藤康夫(w057),
--         遠藤栄一(w070), 菅原浩二(w084)
-- ──────────────────────────────────────

-- ──────────────────────────────────────
-- サンプルシフトデータ（今週分 / 4月14日〜18日）
-- 現場ごとに担当グループを固定し、社員IDで直接引いて割り当て
-- ※ INSERT OR IGNORE なので再実行でもエラーにならない
-- ──────────────────────────────────────
WITH
-- ── 作業者 ID ルックアップ ──
w AS (
  SELECT id, employee_id FROM users WHERE tenant_id = 1
),
-- ── 現場×作業者×日付の割り当て定義 ──
-- columns: site_id, emp_id, work_date, time_slot
s(site_id, emp_id, work_date, time_slot) AS (
  -- 南平岸４条 (site_id=1) : w001〜w011 グループ
  SELECT 1,'w001','2026-04-14','ALL' UNION ALL SELECT 1,'w002','2026-04-14','ALL' UNION ALL SELECT 1,'w003','2026-04-14','ALL' UNION ALL
  SELECT 1,'w004','2026-04-14','ALL' UNION ALL SELECT 1,'w005','2026-04-14','ALL' UNION ALL SELECT 1,'w006','2026-04-14','ALL' UNION ALL
  SELECT 1,'w001','2026-04-15','ALL' UNION ALL SELECT 1,'w002','2026-04-15','ALL' UNION ALL SELECT 1,'w003','2026-04-15','ALL' UNION ALL
  SELECT 1,'w004','2026-04-15','ALL' UNION ALL SELECT 1,'w007','2026-04-15','ALL' UNION ALL SELECT 1,'w008','2026-04-15','ALL' UNION ALL
  SELECT 1,'w001','2026-04-16','ALL' UNION ALL SELECT 1,'w002','2026-04-16','ALL' UNION ALL
  SELECT 1,'w004','2026-04-16','ALL' UNION ALL SELECT 1,'w005','2026-04-16','ALL' UNION ALL SELECT 1,'w009','2026-04-16','ALL' UNION ALL
  SELECT 1,'w001','2026-04-17','ALL' UNION ALL SELECT 1,'w002','2026-04-17','ALL' UNION ALL SELECT 1,'w003','2026-04-17','ALL' UNION ALL
  SELECT 1,'w004','2026-04-17','ALL' UNION ALL SELECT 1,'w010','2026-04-17','ALL' UNION ALL SELECT 1,'w011','2026-04-17','ALL' UNION ALL
  SELECT 1,'w001','2026-04-18','ALL' UNION ALL SELECT 1,'w002','2026-04-18','ALL' UNION ALL
  SELECT 1,'w005','2026-04-18','ALL' UNION ALL SELECT 1,'w006','2026-04-18','ALL' UNION ALL
  -- 島松隊舎 (site_id=2) : w015〜w020 グループ
  SELECT 2,'w015','2026-04-14','AM'  UNION ALL SELECT 2,'w016','2026-04-14','PM'  UNION ALL
  SELECT 2,'w017','2026-04-14','ALL' UNION ALL SELECT 2,'w018','2026-04-14','ALL' UNION ALL
  SELECT 2,'w015','2026-04-15','ALL' UNION ALL SELECT 2,'w016','2026-04-15','ALL' UNION ALL SELECT 2,'w019','2026-04-15','ALL' UNION ALL
  SELECT 2,'w015','2026-04-16','ALL' UNION ALL SELECT 2,'w017','2026-04-16','ALL' UNION ALL SELECT 2,'w020','2026-04-16','ALL' UNION ALL
  SELECT 2,'w016','2026-04-17','AM'  UNION ALL SELECT 2,'w018','2026-04-17','PM'  UNION ALL SELECT 2,'w019','2026-04-17','ALL' UNION ALL
  SELECT 2,'w015','2026-04-18','ALL' UNION ALL SELECT 2,'w016','2026-04-18','ALL' UNION ALL
  -- CR澄川 (site_id=3) : w025〜w030 グループ
  SELECT 3,'w025','2026-04-14','ALL' UNION ALL SELECT 3,'w026','2026-04-14','ALL' UNION ALL
  SELECT 3,'w025','2026-04-15','ALL' UNION ALL SELECT 3,'w026','2026-04-15','ALL' UNION ALL SELECT 3,'w027','2026-04-15','ALL' UNION ALL
  SELECT 3,'w028','2026-04-16','ALL' UNION ALL SELECT 3,'w029','2026-04-16','ALL' UNION ALL
  SELECT 3,'w025','2026-04-17','ALL' UNION ALL SELECT 3,'w027','2026-04-17','ALL' UNION ALL
  SELECT 3,'w025','2026-04-18','ALL' UNION ALL SELECT 3,'w026','2026-04-18','ALL' UNION ALL SELECT 3,'w030','2026-04-18','ALL' UNION ALL
  -- すすきの駅前 (site_id=4) : w035〜w040 グループ
  SELECT 4,'w035','2026-04-14','ALL' UNION ALL SELECT 4,'w036','2026-04-14','ALL' UNION ALL
  SELECT 4,'w035','2026-04-15','ALL' UNION ALL SELECT 4,'w037','2026-04-15','ALL' UNION ALL SELECT 4,'w038','2026-04-15','ALL' UNION ALL
  SELECT 4,'w035','2026-04-16','ALL' UNION ALL SELECT 4,'w036','2026-04-16','ALL' UNION ALL SELECT 4,'w039','2026-04-16','ALL' UNION ALL
  SELECT 4,'w037','2026-04-17','ALL' UNION ALL SELECT 4,'w040','2026-04-17','ALL' UNION ALL
  SELECT 4,'w035','2026-04-18','ALL' UNION ALL SELECT 4,'w038','2026-04-18','ALL' UNION ALL
  -- トヨタ分室 (site_id=5) : w045〜w050 グループ
  SELECT 5,'w045','2026-04-14','ALL' UNION ALL SELECT 5,'w046','2026-04-14','ALL' UNION ALL SELECT 5,'w047','2026-04-14','ALL' UNION ALL
  SELECT 5,'w045','2026-04-15','ALL' UNION ALL SELECT 5,'w048','2026-04-15','ALL' UNION ALL
  SELECT 5,'w046','2026-04-16','ALL' UNION ALL SELECT 5,'w049','2026-04-16','ALL' UNION ALL
  SELECT 5,'w045','2026-04-17','ALL' UNION ALL SELECT 5,'w046','2026-04-17','ALL' UNION ALL SELECT 5,'w050','2026-04-17','ALL' UNION ALL
  SELECT 5,'w045','2026-04-18','ALL' UNION ALL SELECT 5,'w047','2026-04-18','ALL' UNION ALL
  -- 北広工業 (site_id=6) : w055〜w061 グループ
  SELECT 6,'w055','2026-04-14','ALL' UNION ALL SELECT 6,'w056','2026-04-14','ALL' UNION ALL
  SELECT 6,'w055','2026-04-15','ALL' UNION ALL SELECT 6,'w057','2026-04-15','ALL' UNION ALL SELECT 6,'w058','2026-04-15','ALL' UNION ALL
  SELECT 6,'w056','2026-04-16','ALL' UNION ALL SELECT 6,'w059','2026-04-16','ALL' UNION ALL
  SELECT 6,'w055','2026-04-17','ALL' UNION ALL SELECT 6,'w060','2026-04-17','ALL' UNION ALL
  SELECT 6,'w055','2026-04-18','ALL' UNION ALL SELECT 6,'w056','2026-04-18','ALL' UNION ALL SELECT 6,'w061','2026-04-18','ALL'
)
INSERT OR IGNORE INTO shift_assignments (tenant_id, site_id, user_id, work_date, time_slot, created_by)
SELECT 1, s.site_id, w.id, s.work_date, s.time_slot, 1
FROM s JOIN w ON w.employee_id = s.emp_id;
