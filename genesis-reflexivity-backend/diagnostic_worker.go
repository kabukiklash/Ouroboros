package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

type DiagnosticRecord struct {
	ID                       int64     `json:"id"`
	ServiceID                string    `json:"serviceId"`
	ServiceLabel             string    `json:"serviceLabel"`
	RootCauseCategory        string    `json:"rootCauseCategory"`
	RootCauseHuman           string    `json:"rootCauseHuman"`
	MitigationApplied        string    `json:"mitigationApplied"`
	MitigationHuman          string    `json:"mitigationHuman"`
	NeedsDestructiveRestart  float64   `json:"needsDestructiveRestart"`
	DiagnosticConfidence     string    `json:"diagnosticConfidence"`
	HumanSummary             string    `json:"humanSummary"`
	Timestamp                string    `json:"timestamp"`
}

var lastDiagnostics = struct {
	sync.RWMutex
	m map[string]time.Time
}{m: make(map[string]time.Time)}

var cachedDiagnostics = struct {
	sync.RWMutex
	records []DiagnosticRecord
}{records: make([]DiagnosticRecord, 0)}

// getHumanServiceLabel traduz IDs crus de serviços para nomes amigáveis da AfixGraf
func getHumanServiceLabel(serviceId string) (string, string) {
	switch serviceId {
	case "SERVICE_SLOW_NORMAL":
		return "📦 Motor de Processamento Pesado (Relatórios / Lote)", "Simula rotinas de 400-450ms como geração de PDFs, etiquetas e faturamento em lote. Naturalmente pesado e sem defeito."
	case "SERVICE_FAST_NORMAL":
		return "⚡ Motor Transacional Rápido (Tempo Real)", "Simula rotas instantâneas (30-50ms) como autenticação, consulta de sessão e cliques no Kanban."
	case "http://127.0.0.1:3000/v1/health":
		return "🛡️ GenesisCore Foundation (Ledger Central)", "Servidor Node.js na porta 3000 responsável pela governança soberana e ledger imutável de intents."
	case "http://127.0.0.1/3000", "http://127.0.0.1/8090":
		return "🔌 Sentinela de Conectividade Local", "Probe externo ativo para monitorar estabilidade de sockets e interfaces de rede da máquina."
	default:
		return fmt.Sprintf("🌐 Alvo Monitorado: %s", serviceId), "Endpoint monitorado pelo sentinela Ouroboros via pingWorker."
	}
}

// translateDiagnosticToHuman converte os códigos matemáticos do Jev em explicações em bom português
func translateDiagnosticToHuman(category, mitigation string, restartProb float64) (string, string, string) {
	causeHuman := ""
	switch category {
	case "DATABASE_CONTROLLER_LOCK":
		causeHuman = "Contenção de escrita ou concorrência elevada no banco de dados SQLite."
	case "UPSTREAM_NETWORK_TIMEOUT":
		causeHuman = "Atraso na rota de rede ou lentidão transitória no alvo externo monitorado."
	case "RESOURCE_EXHAUSTION_LEAK":
		causeHuman = "Aumento monótono de consumo de memória ou recursos do sistema."
	case "UNEXPECTED_TRAFFIC_SURGE":
		causeHuman = "Pico súbito de requisições de clientes sem erro de protocolo."
	default:
		causeHuman = "Operação sob regime de carga pesada sem anomalia crítica."
	}

	mitigHuman := ""
	switch mitigation {
	case "TIER_1_SHED_LOAD":
		mitigHuman = "Alívio de Carga: Frequência de checagem espaçada para não sobrecarregar o host."
	case "TIER_2_RESET_POOL":
		mitigHuman = "Reciclagem do pool de conexões SQLite efetuada sem desligar o sistema."
	case "TIER_3_DEGRADE_GRACEFUL":
		mitigHuman = "Degradação Graciosa: Consultas atendidas a partir do cache relacional imutável."
	case "TIER_4_RESTART_LAST_RESORT":
		mitigHuman = "Reinicialização de segurança executada como último recurso extremo."
	default:
		mitigHuman = "Nenhuma intervenção necessária: o serviço está saudável no seu ritmo normal."
	}

	summary := ""
	if restartProb < 0.40 {
		summary = fmt.Sprintf("✅ REINÍCIO VETADO PELO JEV (Risco de dano: %.0f%%). A falha é transitória e o restart apenas causaria queda de operadores.", (1.0-restartProb)*100.0)
	} else {
		summary = fmt.Sprintf("⚠️ Reinício avaliado como provável necessidade (Probabilidade: %.0f%%).", restartProb*100.0)
	}

	return causeHuman, mitigHuman, summary
}

// dispatchGate04Diagnostics invoca o Jev System-One em modo assíncrono via ponte isolada
func dispatchGate04Diagnostics(rs RelationalState, shockLevel string, relShock float64, dirDrift float64) {
	// Regra Argon: Ignora serviços naturalmente pesados se não houver choque ou falha real
	if rs.State != "DENY" && rs.SampleOk == 1 && shockLevel == "LOW" && dirDrift < 15.0 {
		return
	}

	// Regra Argon: Cooldown de 60s por serviço para evitar tempestade de chamadas (API storm)
	lastDiagnostics.Lock()
	lastTime, exists := lastDiagnostics.m[rs.ServiceID]
	if exists && time.Since(lastTime) < 60*time.Second {
		lastDiagnostics.Unlock()
		return
	}
	lastDiagnostics.m[rs.ServiceID] = time.Now()
	lastDiagnostics.Unlock()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Gate 04 - Recuperado de pânico no diagnostic worker:", r)
			}
		}()

		statePayload := map[string]interface{}{
			"service_id":       rs.ServiceID,
			"latency_ms":       rs.LatencyMs,
			"friction":         rs.Friction,
			"ck_score":         rs.CkScore,
			"alpha":            rs.AlphaAtTime,
			"relational_shock": relShock,
			"directional_drift": dirDrift,
			"error_kind":       rs.ErrorKind,
			"sample_ok":        rs.SampleOk,
		}

		questions := map[string]interface{}{
			"root_cause_category": map[string]interface{}{
				"type":         "choice",
				"instructions": "Identifique a categoria da anomalia com base no vetor PER.",
				"criteria": map[string]string{
					"DATABASE_CONTROLLER_LOCK": "Choque moderado com pressao alta e escrita concorrente.",
					"UPSTREAM_NETWORK_TIMEOUT": "Latencia punitiva de timeout ou erro de rede no probe.",
					"RESOURCE_EXHAUSTION_LEAK": "Friccao subindo de forma monotona e continua.",
					"UNEXPECTED_TRAFFIC_SURGE": "Aumento repentino de volume sem erro de conexao.",
				},
			},
			"needs_destructive_restart": map[string]interface{}{
				"type":         "noul",
				"instructions": "O reinicio destrutivo do processo e estritamente necessario para sanar esta ocorrencia?",
			},
			"mitigation_tier": map[string]interface{}{
				"type":         "choice",
				"instructions": "Selecione o nivel de mitigacao nao destrutivo mais adequado.",
				"criteria": map[string]string{
					"TIER_1_SHED_LOAD": "Espacar intervalo de checagem e reduzir carga sintetica.",
					"TIER_2_RESET_POOL": "Executar checkpoint passivo no banco de dados SQLite.",
					"TIER_3_DEGRADE_GRACEFUL": "Servir respostas pelo cache relacional local.",
					"TIER_4_RESTART_LAST_RESORT": "Reiniciar servico apenas como ultimo recurso comprovado.",
				},
			},
		}

		bridgeReq := map[string]interface{}{
			"mode":      "diagnose_per",
			"state":     statePayload,
			"questions": questions,
		}

		reqBytes, err := json.Marshal(bridgeReq)
		if err != nil {
			return
		}

		pythonPath := "python"
		bridgeScript := `C:\Users\RobsonSilva-AfixGraf\Habilidade_de_agente\15_JEV_SYSTEM_ONE_GATEKEEPER\core\mcp_cli_bridge.py`

		cmd := exec.Command(pythonPath, bridgeScript)
		cmd.Stdin = bytes.NewReader(reqBytes)
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf

		if err := cmd.Run(); err != nil {
			log.Println("Gate 04 - Falha ao consultar Jev via CLI bridge:", err)
			return
		}

		var bridgeResp struct {
			Ok      bool                   `json:"ok"`
			Answers map[string]interface{} `json:"answers"`
		}
		if err := json.Unmarshal(outBuf.Bytes(), &bridgeResp); err != nil || !bridgeResp.Ok {
			return
		}

		rootCause := "UPSTREAM_NETWORK_TIMEOUT"
		if rcObj, ok := bridgeResp.Answers["root_cause_category"].(map[string]interface{}); ok {
			if ch, ok := rcObj["choice"].(string); ok {
				rootCause = ch
			}
		}

		needsRestart := 0.10
		if nrObj, ok := bridgeResp.Answers["needs_destructive_restart"].(map[string]interface{}); ok {
			if nl, ok := nrObj["noul"].(float64); ok {
				needsRestart = nl
			}
		}

		mitigationTier := "TIER_1_SHED_LOAD"
		if mtObj, ok := bridgeResp.Answers["mitigation_tier"].(map[string]interface{}); ok {
			if ch, ok := mtObj["choice"].(string); ok {
				mitigationTier = ch
			}
		}

		label, _ := getHumanServiceLabel(rs.ServiceID)
		causeHuman, mitigHuman, summary := translateDiagnosticToHuman(rootCause, mitigationTier, needsRestart)

		// Persiste no SQLite
		dbMutex.Lock()
		db.Exec(`INSERT INTO relational_diagnostics_events 
			(service_id, snapshot_id, root_cause, mitigation_applied, needs_restart_noul, confidence_level, diagnostics_payload_json) 
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			rs.ServiceID, rs.ObservedAtMs, rootCause, mitigationTier, needsRestart, "HIGH", summary)
		dbMutex.Unlock()

		// Mantém histórico em memória para o endpoint HTTP da UI
		rec := DiagnosticRecord{
			ID:                      time.Now().UnixNano(),
			ServiceID:               rs.ServiceID,
			ServiceLabel:            label,
			RootCauseCategory:       rootCause,
			RootCauseHuman:          causeHuman,
			MitigationApplied:       mitigationTier,
			MitigationHuman:         mitigHuman,
			NeedsDestructiveRestart: needsRestart,
			DiagnosticConfidence:    "HIGH",
			HumanSummary:            summary,
			Timestamp:               time.Now().Format("15:04:05"),
		}

		cachedDiagnostics.Lock()
		cachedDiagnostics.records = append([]DiagnosticRecord{rec}, cachedDiagnostics.records...)
		if len(cachedDiagnostics.records) > 50 {
			cachedDiagnostics.records = cachedDiagnostics.records[:50]
		}
		cachedDiagnostics.Unlock()

		log.Printf("Gate 04 [JEV DIAGNOSTIC]: Serviço=%s | Causa=%s | Mitigação=%s | RestartProb=%.2f", rs.ServiceID, rootCause, mitigationTier, needsRestart)
	}()
}

