#!/usr/bin/env pwsh
# Audio Steganography LSB API - Development Scripts (Swagger Version)

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("build", "run", "test", "docs", "clean", "swagger")]
    [string]$Action
)

$projectRoot = $PSScriptRoot
$backendDir = $projectRoot

switch ($Action) {
    "build" {
        Write-Host "🔨 Building Audio Steganography LSB API..." -ForegroundColor Cyan
        Set-Location $backendDir
        go build -o bin/audio-steganography-api.exe .
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Build successful!" -ForegroundColor Green
        } else {
            Write-Host "❌ Build failed!" -ForegroundColor Red
            exit 1
        }
    }
    
    "run" {
        Write-Host "🚀 Starting Audio Steganography LSB API Server..." -ForegroundColor Cyan
        Set-Location $backendDir
        go run .
    }
    
    "test" {
        Write-Host "🧪 Running API Tests..." -ForegroundColor Cyan
        Set-Location $backendDir
        
        # Start server in background
        $serverJob = Start-Job -ScriptBlock {
            Set-Location $using:backendDir
            go run .
        }
        
        # Wait for server to start
        Start-Sleep -Seconds 3
        
        # Run tests
        & "$backendDir/test-api.ps1"
        
        # Stop server
        Stop-Job $serverJob
        Remove-Job $serverJob
    }
    
    "docs" {
        Write-Host "🔄 Regenerating Swagger documentation from annotations..." -ForegroundColor Cyan
        Set-Location $backendDir
        
        swag init
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Swagger documentation generation successful!" -ForegroundColor Green
        } else {
            Write-Host "❌ Swagger documentation generation failed!" -ForegroundColor Red
            exit 1
        }
    }
    
    "clean" {
        Write-Host "🧹 Cleaning build artifacts..." -ForegroundColor Cyan
        Set-Location $backendDir
        
        if (Test-Path "bin") {
            Remove-Item -Recurse -Force "bin"
        }
        
        if (Test-Path "audio-steganography-api.exe") {
            Remove-Item "audio-steganography-api.exe"
        }
        
        Write-Host "✅ Clean complete!" -ForegroundColor Green
    }
    
    "swagger" {
        Write-Host "📚 Opening Swagger Documentation..." -ForegroundColor Cyan
        Start-Process "http://localhost:8080/swagger/index.html"
    }
}

Write-Host "`n🎯 Available commands:" -ForegroundColor Blue
Write-Host "  .\dev.ps1 build     - Build the application" -ForegroundColor White
Write-Host "  .\dev.ps1 run       - Run the development server" -ForegroundColor White
Write-Host "  .\dev.ps1 test      - Run API tests" -ForegroundColor White
Write-Host "  .\dev.ps1 docs      - Regenerate Swagger documentation" -ForegroundColor White
Write-Host "  .\dev.ps1 clean     - Clean build artifacts" -ForegroundColor White
Write-Host "  .\dev.ps1 swagger   - Open Swagger documentation" -ForegroundColor White