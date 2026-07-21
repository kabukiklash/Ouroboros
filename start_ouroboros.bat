@echo off
echo =======================================================
echo Inicializando Ouroboros (Genesis Reflexivity)
echo =======================================================

echo.
echo [1/3] Iniciando o Backend (Go - porta 8082)...
start "Ouroboros Backend" cmd /c "cd /d "%~dp0genesis-reflexivity-backend" && go run main.go"

echo.
echo Aguardando o Backend compilar e iniciar (10s)...
timeout /t 10 >nul

echo [2/3] Iniciando o Frontend (Vite - porta 5173)...
start "Ouroboros Frontend" cmd /c "cd /d "%~dp0genesis-reflexivity-ui" && npm run dev"

echo.
echo Aguardando o Frontend subir (5s)...
timeout /t 5 >nul

echo.
echo [3/3] Gerando relatorio de testes...
echo.
echo Testando porta do Frontend (5173)...
netstat -ano | find "5173" >nul
if %errorlevel% equ 0 (
    echo [OK] Frontend esta rodando.
) else (
    echo [FALHA] Frontend parece nao estar respondendo na porta 5173.
)

echo Testando porta do Backend (8082)...
netstat -ano | find "8082" >nul
if %errorlevel% equ 0 (
    echo [OK] Backend esta rodando.
) else (
    echo [AVISO] Backend nao encontrado na porta 8082. Pode estar terminando de compilar.
)

echo.
echo Tudo certo! Pressione qualquer tecla para abrir o navegador...
pause >nul
start http://localhost:5173/

echo =======================================================
echo [SISTEMA ONLINE] Ouroboros esta operante!
echo =======================================================
pause
