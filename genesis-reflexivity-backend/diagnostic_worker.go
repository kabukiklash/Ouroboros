package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type DiagnosticRecord struct {
	ID                       int64   `json:"id"`
	ServiceID                string  `json:"serviceId"`
	ServiceLabel             string  `json:"serviceLabel"`
	SnapshotID               int64   `json:"snapshotId"`
	SampleOrigin             string  `json:"sampleOrigin"`      // "PROBE" ou "SIMULATION"
	SampleOriginHuman        string  `json:"sampleOriginHuman"` // "📡 Sonda Real" ou "🧪 Simulação"
	JevSource                string  `json:"jevSource"`          // "typesafe_jev_api", "cache_sha256", "fail_open_advisory", etc.
	Status                   string  `json:"status"`             // "AVAILABLE" ou "UNAVAILABLE"
	UnavailableReason        string  `json:"unavailableReason"`  // Detalhe se status == UNAVAILABLE
	RootCauseCategory        string  `json:"rootCauseCategory"`
	RootCauseHuman           string  `json:"rootCauseHuman"`
	MitigationSuggested      string  `json:"mitigationSuggested"`
	MitigationSuggestedHuman string  `json:"mitigationSuggestedHuman"`
	MitigationExecuted       string  `json:"mitigationExecuted"`
	MitigationExecutedHuman  string  `json:"mitigationExecutedHuman"`
	MitigationApplied        string  `json:"mitigationApplied"` // Legado para retrocompatibilidade
	MitigationHuman          string  `json:"mitigationHuman"`   // Legado para retrocompatibilidade
	NeedsDestructiveRestart  float64 `json:"needsDestructiveRestart"`
	DiagnosticConfidence     string  `json:"diagnosticConfidence"`
	HumanSummary             string  `json:"humanSummary"`
	Timestamp                string  `json:"timestamp"`
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

// translateReasonToHuman traduz razões técnicas de indisponibilidade para operadores humanos
func translateReasonToHuman(reason string) string {
	if reason == "" {
		return "Indisponibilidade momentânea na camada de inferência"
	}
	if reason == "NO_TYPESAFE_API_KEY" {
		return "Chave TypeSafe AI não configurada no ambiente"
	}
	if strings.Contains(reason, "fail_open_advisory") || strings.Contains(reason, "FAIL_OPEN_ADVISORY") {
		return "Modo Fail-Open de Proteção (Advisory ativo sem bloqueio)"
	}
	if strings.HasPrefix(reason, "API_ERROR:") {
		return fmt.Sprintf("Falha no endpoint da API Jev (%s)", strings.TrimPrefix(reason, "API_ERROR: "))
	}
	if reason == "RESPOSTA_INCOMPLETA_JEV" || strings.Contains(reason, "RESPOSTA_INCOMPLETA") {
		return "Resposta parcial ou campos essenciais ausentes no retorno do JEV"
	}
	return reason
}

// translateMitigationToHuman converte os tiers de mitigação para explicações claras
func translateMitigationToHuman(mitigation string) string {
	switch mitigation {
	case "TIER_1_SHED_LOAD":
		return "Alívio de Carga: Frequência de checagem espaçada para não sobrecarregar o host."
	case "TIER_2_RESET_POOL":
		return "Reciclagem do pool de conexões SQLite efetuada sem desligar o sistema."
	case "TIER_3_DEGRADE_GRACEFUL":
		return "Degradação Graciosa: Consultas atendidas a partir do cache relacional imutável."
	case "TIER_4_RESTART_LAST_RESORT":
		return "Reinicialização de segurança executada como último recurso extremo."
	case "NONE":
		return "Nenhuma mitigação aplicável (diagnóstico indisponível)."
	default:
		return "Nenhuma intervenção necessária: o serviço está saudável no seu ritmo normal."
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
	case "DIAGNOSTIC_UNAVAILABLE":
		causeHuman = "Diagnóstico Indisponível pelo Jev."
	default:
		causeHuman = "Operação sob regime de carga pesada sem anomalia crítica."
	}

	mitigHuman := translateMitigationToHuman(mitigation)

	summary := ""
	if restartProb < 0 {
		summary = "⚠️ Diagnóstico indisponível: o sistema mantém proteção passiva sem reinício."
	} else if restartProb < 0.40 {
		summary = fmt.Sprintf("✅ REINÍCIO VETADO PELO JEV (Risco de dano: %.0f%%). A falha é transitória e o restart apenas causaria queda de operadores.", (1.0-restartProb)*100.0)
	} else {
		summary = fmt.Sprintf("⚠️ Reinício avaliado como provável necessidade (Probabilidade: %.0f%%).", restartProb*100.0)
	}

	return causeHuman, mitigHuman, summary
}

// loadHistoricalDiagnostics restaura os últimos 50 eventos do SQLite na inicialização
func loadHistoricalDiagnostics() {
	dbMutex.Lock()
	defer dbMutex.Unlock()

	rows, err := db.Query(`SELECT id, service_id, snapshot_id, root_cause, mitigation_applied, needs_restart_noul, confidence_level, diagnostics_payload_json, 
		COALESCE(sample_origin, 'SIMULATION'), COALESCE(jev_source, 'typesafe_jev_api'), COALESCE(status, 'AVAILABLE'), 
		COALESCE(unavailable_reason, ''), COALESCE(mitigation_suggested, ''), COALESCE(mitigation_executed, ''), strftime('%H:%M:%S', created_at)
		FROM relational_diagnostics_events ORDER BY id DESC LIMIT 50`)
	if err != nil {
		log.Println("Gate 04 - Aviso ao carregar histórico inicial:", err)
		return
	}
	defer rows.Close()

	var records []DiagnosticRecord
	for rows.Next() {
		var r DiagnosticRecord
		var timeStr string
		var rawJson string
		err := rows.Scan(&r.ID, &r.ServiceID, &r.SnapshotID, &r.RootCauseCategory, &r.MitigationApplied, &r.NeedsDestructiveRestart,
			&r.DiagnosticConfidence, &rawJson, &r.SampleOrigin, &r.JevSource, &r.Status, &r.UnavailableReason, &r.MitigationSuggested, &r.MitigationExecuted, &timeStr)
		if err != nil {
			continue
		}
		r.Timestamp = timeStr
		label, _ := getHumanServiceLabel(r.ServiceID)
		r.ServiceLabel = label
		if r.SampleOrigin == "PROBE" {
			r.SampleOriginHuman = "📡 Sonda Real"
		} else {
			r.SampleOriginHuman = "🧪 Simulação"
		}

		if r.Status == "UNAVAILABLE" {
			r.RootCauseHuman = fmt.Sprintf("⚠️ Diagnóstico Indisponível (%s)", translateReasonToHuman(r.UnavailableReason))
			r.MitigationSuggestedHuman = "Nenhuma sugestão disponível"
			r.MitigationExecutedHuman = "Executado: Cooldown passivo ativo (sem restart)"
			r.HumanSummary = fmt.Sprintf("⚠️ DIAGNÓSTICO INDISPONÍVEL: %s", translateReasonToHuman(r.UnavailableReason))
		} else {
			causeHuman, _, summary := translateDiagnosticToHuman(r.RootCauseCategory, r.MitigationSuggested, r.NeedsDestructiveRestart)
			r.RootCauseHuman = causeHuman
			r.MitigationSuggestedHuman = translateMitigationToHuman(r.MitigationSuggested)
			if r.NeedsDestructiveRestart < 0.40 {
				r.MitigationExecutedHuman = "Executado: Reinício vetado pelo sistema imune (cooldown 60s)"
			} else {
				r.MitigationExecutedHuman = "Executado: Alerta em modo sombra emitido ao operador"
			}
			r.HumanSummary = summary
		}
		r.MitigationHuman = r.MitigationSuggestedHuman
		records = append(records, r)
	}

	cachedDiagnostics.Lock()
	cachedDiagnostics.records = records
	cachedDiagnostics.Unlock()
}

// dispatchGate04Diagnostics invoca o Jev System-One em modo assíncrono via ponte isolada
func dispatchGate04Diagnostics(rs RelationalState, snapshotID int64, shockLevel string, relShock float64, dirDrift float64) {
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

		// Identificação precisa da origem da amostra
		sampleOrigin := "SIMULATION"
		sampleOriginHuman := "🧪 Simulação"
		if rs.EntropySource == "REAL_NETWORK" {
			sampleOrigin = "PROBE"
			sampleOriginHuman = "📡 Sonda Real"
		}

		statePayload := map[string]interface{}{
			"service_id":        rs.ServiceID,
			"snapshot_id":       snapshotID,
			"sample_origin":     sampleOrigin,
			"latency_ms":        rs.LatencyMs,
			"friction":          rs.Friction,
			"ck_score":          rs.CkScore,
			"alpha":             rs.AlphaAtTime,
			"relational_shock":  relShock,
			"directional_drift": dirDrift,
			"error_kind":        rs.ErrorKind,
			"sample_ok":         rs.SampleOk,
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
					"TIER_1_SHED_LOAD":          "Espacar intervalo de checagem e reduzir carga sintetica.",
					"TIER_2_RESET_POOL":         "Executar checkpoint passivo no banco de dados SQLite.",
					"TIER_3_DEGRADE_GRACEFUL":   "Servir respostas pelo cache relacional local.",
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
		var errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		cmdErr := cmd.Run()

		var bridgeResp struct {
			Ok             bool                   `json:"ok"`
			Source         string                 `json:"source"`
			FallbackReason string                 `json:"fallback_reason"`
			Error          string                 `json:"error"`
			Answers        map[string]interface{} `json:"answers"`
			CacheHit       bool                   `json:"cache_hit"`
		}

		status := "AVAILABLE"
		unavailableReason := ""
		jevSource := "typesafe_jev_api"

		if cmdErr != nil {
			status = "UNAVAILABLE"
			unavailableReason = fmt.Sprintf("FALHA_EXECUCAO_CLI_BRIDGE: %v", cmdErr)
			jevSource = "cli_bridge_error"
		} else if err := json.Unmarshal(outBuf.Bytes(), &bridgeResp); err != nil {
			status = "UNAVAILABLE"
			unavailableReason = "ERRO_DESERIALIZACAO_JSON"
			jevSource = "raw_output_error"
		} else if !bridgeResp.Ok {
			status = "UNAVAILABLE"
			unavailableReason = bridgeResp.Error
			if unavailableReason == "" {
				unavailableReason = "BRIDGE_ERROR_RESPONSE"
			}
			jevSource = "bridge_rejected"
		} else if bridgeResp.Source == "fail_open_advisory" {
			// Protocolo de Honestidade: Tratar fail_open_advisory como diagnóstico indisponível, sem inventar causa raiz!
			status = "UNAVAILABLE"
			unavailableReason = bridgeResp.FallbackReason
			if unavailableReason == "" {
				unavailableReason = "FAIL_OPEN_ADVISORY"
			}
			jevSource = "fail_open_advisory"
		} else {
			jevSource = bridgeResp.Source
			if bridgeResp.CacheHit {
				jevSource = "cache_sha256"
			}
		}

		// Checar completude das respostas do JEV
		var rcObj, nrObj, mtObj map[string]interface{}
		var rootCause, mitigationTier string
		var needsRestart float64 = -1.0

		if status == "AVAILABLE" {
			var okRc, okNr, okMt bool
			rcObj, okRc = bridgeResp.Answers["root_cause_category"].(map[string]interface{})
			nrObj, okNr = bridgeResp.Answers["needs_destructive_restart"].(map[string]interface{})
			mtObj, okMt = bridgeResp.Answers["mitigation_tier"].(map[string]interface{})

			if !okRc || !okNr || !okMt {
				status = "UNAVAILABLE"
				unavailableReason = "RESPOSTA_INCOMPLETA_JEV (campos obrigatórios ausentes)"
			} else {
				if ch, ok := rcObj["choice"].(string); ok && ch != "" {
					rootCause = ch
				} else {
					status = "UNAVAILABLE"
					unavailableReason = "RESPOSTA_INCOMPLETA_JEV (campo root_cause_category inválido)"
				}

				if nl, ok := nrObj["noul"].(float64); ok {
					needsRestart = nl
				} else {
					status = "UNAVAILABLE"
					unavailableReason = "RESPOSTA_INCOMPLETA_JEV (campo needs_destructive_restart inválido)"
				}

				if ch, ok := mtObj["choice"].(string); ok && ch != "" {
					mitigationTier = ch
				} else {
					status = "UNAVAILABLE"
					unavailableReason = "RESPOSTA_INCOMPLETA_JEV (campo mitigation_tier inválido)"
				}
			}
		}

		label, _ := getHumanServiceLabel(rs.ServiceID)
		var causeHuman, mitigationSuggested, mitigationSuggestedHuman, mitigationExecuted, mitigationExecutedHuman, summary string

		if status == "UNAVAILABLE" {
			rootCause = "DIAGNOSTIC_UNAVAILABLE"
			causeHuman = fmt.Sprintf("⚠️ Diagnóstico Indisponível (%s)", translateReasonToHuman(unavailableReason))
			mitigationSuggested = "NONE"
			mitigationSuggestedHuman = "Nenhuma recomendação disponível pelo JEV"
			mitigationExecuted = "DEFENSIVE_COOLDOWN"
			mitigationExecutedHuman = "Executado: Cooldown defensivo ativo (modo passivo sem reinício cego)"
			summary = fmt.Sprintf("⚠️ DIAGNÓSTICO INDISPONÍVEL: %s. Operando em regime de proteção passiva.", translateReasonToHuman(unavailableReason))
		} else {
			mitigationSuggested = mitigationTier
			mitigationSuggestedHuman = translateMitigationToHuman(mitigationTier)

			if needsRestart < 0.40 {
				mitigationExecuted = "VETO_RESTART_AND_PACE"
				mitigationExecutedHuman = "Executado: Reinício vetado pelo sistema imune (cooldown 60s aplicado)"
			} else {
				mitigationExecuted = "ALERT_OPERATOR_SHADOW"
				mitigationExecutedHuman = "Executado: Alerta emitido em modo sombra (sem reinício automático)"
			}

			ch, _, sum := translateDiagnosticToHuman(rootCause, mitigationSuggested, needsRestart)
			causeHuman = ch
			summary = sum
		}

		// Persiste no SQLite com rastreabilidade completa (Snapshot ID real, Origem da Amostra, Origem do Jev, Status)
		dbMutex.Lock()
		db.Exec(`INSERT INTO relational_diagnostics_events 
			(service_id, snapshot_id, root_cause, mitigation_applied, needs_restart_noul, confidence_level, diagnostics_payload_json, sample_origin, jev_source, status, unavailable_reason, mitigation_suggested, mitigation_executed) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			rs.ServiceID, snapshotID, rootCause, mitigationSuggested, needsRestart, "HIGH", summary, sampleOrigin, jevSource, status, unavailableReason, mitigationSuggested, mitigationExecuted)
		dbMutex.Unlock()

		// Mantém histórico em memória para o endpoint HTTP da UI
		rec := DiagnosticRecord{
			ID:                       time.Now().UnixNano(),
			ServiceID:                rs.ServiceID,
			ServiceLabel:             label,
			SnapshotID:               snapshotID,
			SampleOrigin:             sampleOrigin,
			SampleOriginHuman:        sampleOriginHuman,
			JevSource:                jevSource,
			Status:                   status,
			UnavailableReason:        unavailableReason,
			RootCauseCategory:        rootCause,
			RootCauseHuman:           causeHuman,
			MitigationSuggested:      mitigationSuggested,
			MitigationSuggestedHuman: mitigationSuggestedHuman,
			MitigationExecuted:       mitigationExecuted,
			MitigationExecutedHuman:  mitigationExecutedHuman,
			MitigationApplied:        mitigationSuggested,
			MitigationHuman:          mitigationSuggestedHuman,
			NeedsDestructiveRestart:  needsRestart,
			DiagnosticConfidence:     "HIGH",
			HumanSummary:             summary,
			Timestamp:                time.Now().Format("15:04:05"),
		}

		cachedDiagnostics.Lock()
		cachedDiagnostics.records = append([]DiagnosticRecord{rec}, cachedDiagnostics.records...)
		if len(cachedDiagnostics.records) > 50 {
			cachedDiagnostics.records = cachedDiagnostics.records[:50]
		}
		cachedDiagnostics.Unlock()

		log.Printf("Gate 04 [JEV]: Snapshot=#%d (%s) | Serviço=%s | Status=%s | Origem=%s | Causa=%s | Sugerida=%s | Executada=%s",
			snapshotID, sampleOrigin, rs.ServiceID, status, jevSource, rootCause, mitigationSuggested, mitigationExecuted)
	}()
}
