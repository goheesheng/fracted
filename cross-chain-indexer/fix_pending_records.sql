-- 修复历史 Pending 记录的 SQL 脚本
-- 将所有超过 2 分钟的 Pending 交易标记为 Delivered

-- 1. 首先查看受影响的记录（安全检查）
SELECT 
    tx_hash,
    status,
    datetime(timestamp) as tx_time,
    ROUND((julianday('now') - julianday(timestamp)) * 24 * 60, 2) as age_minutes,
    dst_eid,
    payer,
    merchant
FROM payouts 
WHERE status = 'Pending' 
    AND datetime(timestamp) < datetime('now', '-2 minutes')
ORDER BY timestamp DESC;

-- 2. 显示统计信息
SELECT 
    status,
    COUNT(*) as count,
    MIN(datetime(timestamp)) as oldest,
    MAX(datetime(timestamp)) as newest
FROM payouts 
GROUP BY status;

-- 3. 执行更新（将超过 2 分钟的 Pending 改为 Delivered）
-- 注意：执行前请先运行上面的查询确认
UPDATE payouts 
SET status = 'Delivered' 
WHERE status = 'Pending' 
    AND datetime(timestamp) < datetime('now', '-2 minutes');

-- 4. 验证更新结果
SELECT 
    status,
    COUNT(*) as count
FROM payouts 
GROUP BY status;

-- 5. 查看最近的交易状态
SELECT 
    tx_hash,
    status,
    datetime(timestamp) as tx_time,
    dst_eid
FROM payouts 
ORDER BY timestamp DESC 
LIMIT 20;

