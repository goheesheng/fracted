#!/bin/bash
# 修复数据库中历史 Pending 记录的脚本

SSH_KEY="~/.ssh/id_rsa_new"
SERVER="azureuser@85.211.176.154"

echo "================================================"
echo "修复数据库历史 Pending 记录"
echo "================================================"
echo ""

# 1. 查看当前状态
echo "==== 1. 当前数据库状态 ===="
ssh -i $SSH_KEY $SERVER << 'EOF'
cd cross-chain-indexer
sqlite3 indexer.db << 'SQL'
.headers on
.mode column
SELECT 
    status,
    COUNT(*) as count
FROM payouts 
GROUP BY status;
SQL
EOF

echo ""

# 2. 查看需要更新的记录
echo "==== 2. 需要更新的 Pending 记录（超过2分钟）===="
ssh -i $SSH_KEY $SERVER << 'EOF'
cd cross-chain-indexer
sqlite3 indexer.db << 'SQL'
.headers on
.mode column
SELECT 
    COUNT(*) as pending_count,
    MIN(datetime(timestamp)) as oldest_tx,
    MAX(datetime(timestamp)) as newest_tx
FROM payouts 
WHERE status = 'Pending' 
    AND datetime(timestamp) < datetime('now', '-2 minutes');
SQL
EOF

echo ""
read -p "是否继续更新这些记录？(y/n): " confirm

if [ "$confirm" != "y" ]; then
    echo "取消更新"
    exit 0
fi

# 3. 执行更新
echo ""
echo "==== 3. 执行更新 ===="
ssh -i $SSH_KEY $SERVER << 'EOF'
cd cross-chain-indexer
sqlite3 indexer.db << 'SQL'
UPDATE payouts 
SET status = 'Delivered' 
WHERE status = 'Pending' 
    AND datetime(timestamp) < datetime('now', '-2 minutes');

SELECT changes() as 'Updated Records';
SQL
EOF

echo ""

# 4. 验证结果
echo "==== 4. 更新后的状态 ===="
ssh -i $SSH_KEY $SERVER << 'EOF'
cd cross-chain-indexer
sqlite3 indexer.db << 'SQL'
.headers on
.mode column
SELECT 
    status,
    COUNT(*) as count
FROM payouts 
GROUP BY status;
SQL
EOF

echo ""

# 5. 显示最近的交易
echo "==== 5. 最近 10 笔交易 ===="
ssh -i $SSH_KEY $SERVER << 'EOF'
cd cross-chain-indexer
sqlite3 indexer.db << 'SQL'
.headers on
.mode column
SELECT 
    substr(tx_hash, 1, 16) || '...' as tx_hash,
    status,
    datetime(timestamp) as time,
    dst_eid
FROM payouts 
ORDER BY timestamp DESC 
LIMIT 10;
SQL
EOF

echo ""
echo "================================================"
echo "✅ 数据库修复完成！"
echo "================================================"

