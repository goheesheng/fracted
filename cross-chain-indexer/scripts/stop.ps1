# Cross-Chain Indexer - Windows 停止脚本

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host " Cross-Chain Indexer - 停止服务" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# 查找并停止进程
$processes = Get-Process | Where-Object {$_.ProcessName -eq "cross-chain-indexer"}

if ($processes) {
    Write-Host "发现 $($processes.Count) 个进程，正在停止..." -ForegroundColor Yellow
    $processes | Stop-Process -Force
    Start-Sleep -Seconds 1
    
    # 验证是否已停止
    $remaining = Get-Process | Where-Object {$_.ProcessName -eq "cross-chain-indexer"}
    if ($remaining) {
        Write-Host "警告: 仍有进程在运行" -ForegroundColor Red
    } else {
        Write-Host "服务已停止" -ForegroundColor Green
    }
} else {
    Write-Host "没有找到运行中的进程" -ForegroundColor Gray
}

Write-Host ""
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host " 停止完成" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "重新启动服务:" -ForegroundColor Yellow
Write-Host "  .\scripts\start.ps1" -ForegroundColor Gray
Write-Host ""

