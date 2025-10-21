# Test script for Cross-Chain Indexer

Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Cross-Chain Indexer Test Script" -ForegroundColor Cyan  
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# 1. Check service status
Write-Host "1. Checking service status..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
    Write-Host "   - Database: $($health.db)" -ForegroundColor Green
    Write-Host "   - WSS Status: $($health.wssStatus)" -ForegroundColor Green
} catch {
    Write-Host "   - ERROR: Service not responding" -ForegroundColor Red
    exit
}
Write-Host ""

# 2. Get admin token
Write-Host "2. Getting admin token..." -ForegroundColor Yellow
$loginBody = @{
    address = "0x27f9B6A7C1Fd66AC4D0e76a2d43B35e8590165f6"
    role = "admin"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "   - Token obtained successfully" -ForegroundColor Green
} catch {
    Write-Host "   - Login failed: $_" -ForegroundColor Red
    exit
}
Write-Host ""

# 3. Query recent transactions
Write-Host "3. Querying recent transactions..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer $token"
}

try {
    $payouts = Invoke-RestMethod -Uri "http://localhost:8080/admin/payouts?limit=20" -Method Get -Headers $headers
    
    Write-Host "   - Found: $($payouts.Count) transactions" -ForegroundColor Green
    Write-Host ""
    
    if ($payouts.Count -gt 0) {
        Write-Host "4. Recent transactions:" -ForegroundColor Yellow
        Write-Host "   ----------------------------------------------------------------" -ForegroundColor Gray
        
        foreach ($tx in $payouts) {
            $txHash = $tx.TxHash
            $shortHash = $txHash.Substring(0, 10) + "..." + $txHash.Substring($txHash.Length - 8)
            $dstChain = $tx.DstChain
            $value = $tx.NetAmountUSD
            $time = $tx.Timestamp
            
            Write-Host "   TX: $shortHash" -ForegroundColor Cyan
            Write-Host "      Chain: $dstChain | Amount: `$$value | Time: $time" -ForegroundColor White
            
            # Check if this is the user's transaction
            if ($txHash -match "5268b1a1c7283cfd") {
                Write-Host "      [SUCCESS] This is your Base->Arb transaction!" -ForegroundColor Green
            }
        }
        Write-Host "   ----------------------------------------------------------------" -ForegroundColor Gray
    }
    
    Write-Host ""
    Write-Host "5. Looking for specific transaction (0x5268b1a1...)..." -ForegroundColor Yellow
    $found = $false
    foreach ($tx in $payouts) {
        if ($tx.TxHash -match "5268b1a1c7283cfd") {
            $found = $true
            Write-Host ""
            Write-Host "   [SUCCESS] Found your transaction!" -ForegroundColor Green
            Write-Host "   Transaction details:" -ForegroundColor Cyan
            Write-Host "   Hash: $($tx.TxHash)" -ForegroundColor White
            Write-Host "   Destination: $($tx.DstChain)" -ForegroundColor White
            Write-Host "   Amount: `$$($tx.NetAmountUSD)" -ForegroundColor White
            Write-Host "   Payer: $($tx.Payer)" -ForegroundColor White
            Write-Host "   Merchant: $($tx.Merchant)" -ForegroundColor White
            Write-Host "   Token: SRC=$($tx.SrcToken), DST=$($tx.DstToken)" -ForegroundColor White
            Write-Host "   Status: $($tx.Status)" -ForegroundColor White
            Write-Host "   Time: $($tx.Timestamp)" -ForegroundColor White
            break
        }
    }
    
    if (-not $found) {
        Write-Host "   [WARNING] Transaction not found in recent 20 records" -ForegroundColor Yellow
        Write-Host "   Possible reasons:" -ForegroundColor Gray
        Write-Host "   - Still backfilling (Base listener is scanning 50000 blocks)" -ForegroundColor Gray
        Write-Host "   - Transaction is older, not in recent 20" -ForegroundColor Gray
        Write-Host ""
        Write-Host "   Suggestion: Wait 1-2 minutes and run again" -ForegroundColor Gray
    }
    
} catch {
    Write-Host "   - Query failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "==================================" -ForegroundColor Cyan
Write-Host "Test Complete!" -ForegroundColor Cyan
Write-Host "Dashboard: http://localhost:8080/dashboard/" -ForegroundColor Green
Write-Host "==================================" -ForegroundColor Cyan
