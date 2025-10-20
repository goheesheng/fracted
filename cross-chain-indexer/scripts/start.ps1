# Cross-Chain Indexer - Windows 启动脚本
# 功能: 编译、启动服务、健康检查

Write-Host "=====================================" -ForegroundColor Cyan
Write-Host " Cross-Chain Indexer - 启动服务" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# 检查是否在项目根目录
if (-not (Test-Path "main.go")) {
    Write-Host "错误: 请在项目根目录执行此脚本" -ForegroundColor Red
    exit 1
}

# 1. 停止现有进程
Write-Host "1. 检查现有进程..." -ForegroundColor Yellow
$existing = Get-Process | Where-Object {$_.ProcessName -eq "cross-chain-indexer"}
if ($existing) {
    Write-Host "   发现运行中的进程，正在停止..." -ForegroundColor Gray
    $existing | Stop-Process -Force
    Start-Sleep -Seconds 1
    Write-Host "   已停止" -ForegroundColor Green
} else {
    Write-Host "   无运行中的进程" -ForegroundColor Gray
}

# 2. 编译项目
Write-Host "`n2. 编译项目..." -ForegroundColor Yellow
$compileResult = go build -o cross-chain-indexer.exe . 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "   编译失败:" -ForegroundColor Red
    Write-Host $compileResult -ForegroundColor Red
    exit 1
}
Write-Host "   编译成功" -ForegroundColor Green

# 3. 启动服务
Write-Host "`n3. 启动服务..." -ForegroundColor Yellow
Start-Process -FilePath ".\cross-chain-indexer.exe" -WindowStyle Hidden
Start-Sleep -Seconds 3

# 4. 健康检查
Write-Host "`n4. 健康检查..." -ForegroundColor Yellow
$maxRetries = 5
$retryCount = 0
$healthy = $false

while ($retryCount -lt $maxRetries) {
    try {
        $health = Invoke-WebRequest -Uri "http://localhost:8080/health" -Method GET -UseBasicParsing -TimeoutSec 3 | ConvertFrom-Json
        if ($health.ok) {
            $healthy = $true
            break
        }
    } catch {
        $retryCount++
        if ($retryCount -lt $maxRetries) {
            Write-Host "   等待服务启动... ($retryCount/$maxRetries)" -ForegroundColor Gray
            Start-Sleep -Seconds 2
        }
    }
}

if ($healthy) {
    Write-Host "   服务运行正常" -ForegroundColor Green
} else {
    Write-Host "   服务启动超时，请检查日志" -ForegroundColor Red
    exit 1
}

# 5. 显示信息
Write-Host "`n=====================================" -ForegroundColor Cyan
Write-Host " 服务启动成功！" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Dashboard:" -ForegroundColor Yellow
Write-Host "  http://localhost:8080/dashboard/" -ForegroundColor Cyan
Write-Host ""
Write-Host "健康检查:" -ForegroundColor Yellow
Write-Host "  http://localhost:8080/health" -ForegroundColor Cyan
Write-Host ""
Write-Host "日志文件:" -ForegroundColor Yellow
Write-Host "  solana_log.txt (Solana 交易日志)" -ForegroundColor Gray
Write-Host ""
Write-Host "停止服务:" -ForegroundColor Yellow
Write-Host "  .\scripts\stop.ps1" -ForegroundColor Gray
Write-Host ""
Write-Host "查看 Solana 日志:" -ForegroundColor Yellow
Write-Host "  Get-Content solana_log.txt -Wait -Tail 20" -ForegroundColor Gray
Write-Host ""

