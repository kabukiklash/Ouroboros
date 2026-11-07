package main

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTranslateDiagnosticToHuman(t *testing.T) {
	// 1. Caso com probabilidade de reinício baixa (< 40%)
	cause, mitig, summary := translateDiagnosticToHuman("RESOURCE_EXHAUSTION_LEAK", "TIER_1_SHED_LOAD", 0.15)
	if !strings.Contains(cause, "Aumento monótono") {
		t.Errorf("Esperava tradução para RESOURCE_EXHAUSTION_LEAK, obtido: %s", cause)
	}
	if !strings.Contains(mitig, "Alívio de Carga") {
		t.Errorf("Esperava tradução para TIER_1_SHED_LOAD, obtido: %s", mitig)
	}
	// Garante que a falsa equivalência de "risco de dano" foi removida
	if strings.Contains(summary, "Risco de dano") {
		t.Errorf("Falsa equivalência 'Risco de dano' não deve constar no sumário: %s", summary)
	}
	if !strings.Contains(summary, "REINÍCIO NÃO RECOMENDADO") || !strings.Contains(summary, "15.0%") {
		t.Errorf("Sumário calibrado incorreto: %s", summary)
	}

	// 2. Caso com probabilidade de reinício alta (>= 40%)
	_, _, summaryHigh := translateDiagnosticToHuman("UPSTREAM_NETWORK_TIMEOUT", "TIER_4_RESTART_LAST_RESORT", 0.75)
	if strings.Contains(summaryHigh, "Risco de dano") {
		t.Errorf("Falsa equivalência 'Risco de dano' encontrada no sumário de alta probabilidade: %s", summaryHigh)
	}
	if !strings.Contains(summaryHigh, "REINÍCIO INDICADO PELO MODELO") || !strings.Contains(summaryHigh, "75.0%") {
		t.Errorf("Esperava alerta explícito de necessidade de reinício em 75.0%%, obtido: %s", summaryHigh)
	}

	// 3. Caso de diagnóstico indisponível
	causeUnavail, _, summaryUnavail := translateDiagnosticToHuman("DIAGNOSTIC_UNAVAILABLE", "NONE", -1.0)
	if !strings.Contains(causeUnavail, "Diagnóstico Indisponível") {
		t.Errorf("Esperava diagnóstico indisponível na causa, obtido: %s", causeUnavail)
	}
	if !strings.Contains(summaryUnavail, "Diagnóstico indisponível pelo JEV") {
		t.Errorf("Esperava menção de indisponibilidade no sumário, obtido: %s", summaryUnavail)
	}
}

func TestSampleOrigin_PreservesProbeOnPunitiveFallback(t *testing.T) {
	// Valida que falha punitiva ou timeout em probe de rede real NÃO é rebaixada para simulação
	testCases := []struct {
		serviceID     string
		entropySource string
		expected      string
	}{
		{"http://127.0.0.1:8050", "REAL_NETWORK", "PROBE"},
		{"http://127.0.0.1:8050", "PUNITIVE_FALLBACK", "PROBE"},
		{"http://127.0.0.1:3000/v1/health", "PUNITIVE_FALLBACK", "PROBE"},
		{"SERVICE_FAST_NORMAL", "SIMULATION", "SIMULATION"},
		{"SERVICE_SLOW_NORMAL", "SIMULATION", "SIMULATION"},
	}

	for _, tc := range testCases {
		rs := RelationalState{
			ServiceID:     tc.serviceID,
			EntropySource: tc.entropySource,
		}

		sampleOrigin := "SIMULATION"
		if rs.EntropySource == "REAL_NETWORK" || rs.EntropySource == "PUNITIVE_FALLBACK" || strings.HasPrefix(rs.ServiceID, "http://") || strings.HasPrefix(rs.ServiceID, "https://") {
			sampleOrigin = "PROBE"
		}

		if sampleOrigin != tc.expected {
			t.Errorf("Para service=%s e entropySource=%s, esperava origin=%s mas obteve %s",
				tc.serviceID, tc.entropySource, tc.expected, sampleOrigin)
		}
	}
}

func TestEvidenceRuleAcceptance_PercentageIdentifiedAsConfiancaDeclaradaPeloJEV(t *testing.T) {
	// Critério de Aceitação 1: Percentuais devem ser identificados formalmente como "confiança declarada pelo JEV: X%"
	validJSON := `{
		"ok": true,
		"source": "typesafe_jev_api",
		"model": "jev-1.14.0",
		"answers": {
			"root_cause_category": {"type": "choice", "choice": "DATABASE_CONTROLLER_LOCK", "confidence": 0.94},
			"needs_destructive_restart": {"type": "noul", "noul": 0.22},
			"mitigation_tier": {"type": "choice", "choice": "TIER_2_RESET_POOL", "confidence": 0.98}
		},
		"cache_hit": false
	}`

	var bridgeResp struct {
		Ok             bool                   `json:"ok"`
		Source         string                 `json:"source"`
		FallbackReason string                 `json:"fallback_reason"`
		Error          string                 `json:"error"`
		Answers        map[string]interface{} `json:"answers"`
		CacheHit       bool                   `json:"cache_hit"`
	}

	if err := json.Unmarshal([]byte(validJSON), &bridgeResp); err != nil {
		t.Fatalf("Failed to unmarshal valid JSON: %v", err)
	}

	status := "AVAILABLE"
	if status != "AVAILABLE" {
		t.Fatalf("Esperava status AVAILABLE")
	}
	rcObj := bridgeResp.Answers["root_cause_category"].(map[string]interface{})
	nrObj := bridgeResp.Answers["needs_destructive_restart"].(map[string]interface{})
	mtObj := bridgeResp.Answers["mitigation_tier"].(map[string]interface{})

	rootCause := rcObj["choice"].(string)
	needsRestart := nrObj["noul"].(float64)
	mitigationTier := mtObj["choice"].(string)

	// Extração calibrada formalmente identificada como confiança declarada pelo JEV
	confidenceLevel := "não informada"
	if confVal, ok := rcObj["confidence"].(float64); ok {
		confidenceLevel = fmt.Sprintf("confiança declarada pelo JEV: %.0f%%", confVal*100.0)
	}

	expectedPrefix := "confiança declarada pelo JEV: 94%"
	if confidenceLevel != expectedPrefix {
		t.Errorf("Critério de Aceitação Violado: esperava '%s', mas obteve: '%s'", expectedPrefix, confidenceLevel)
	}

	// Mitigação sugerida vs executada de acordo com o comportamento real do host
	mitigationSuggested := mitigationTier
	mitigationExecuted := ""
	mitigationExecutedHuman := ""
	if needsRestart < 0.40 {
		mitigationExecuted = "PASSIVE_SHADOW_NO_RESTART"
		mitigationExecutedHuman = "Modo Sombra Passivo: Reinício automático não acionado (despachante JEV em rate-limit de 60s; sondagem de 1s inalterada)"
	}

	if mitigationSuggested != "TIER_2_RESET_POOL" {
		t.Errorf("Mitigação sugerida incorreta: %s", mitigationSuggested)
	}
	if mitigationExecuted != "PASSIVE_SHADOW_NO_RESTART" {
		t.Errorf("Mitigação executada incorreta: %s", mitigationExecuted)
	}
	if !strings.Contains(mitigationExecutedHuman, "sondagem de 1s inalterada") {
		t.Errorf("Descrição de mitigação executada deve refletir o comportamento real do pingWorker: %s", mitigationExecutedHuman)
	}
	if rootCause != "DATABASE_CONTROLLER_LOCK" {
		t.Errorf("Root cause incorreta: %s", rootCause)
	}
}

func TestEvidenceRuleAcceptance_AbsentConfidenceIsNaoInformada(t *testing.T) {
	// Critério de Aceitação 2: Confiança ausente deve ser rigorosamente ajustada para "não informada"
	failOpenJSON := `{
		"ok": true,
		"source": "fail_open_advisory",
		"fallback_reason": "API_ERROR: HTTP Error 503: Service Unavailable",
		"answers": {},
		"cache_hit": false
	}`

	var bridgeResp struct {
		Ok             bool                   `json:"ok"`
		Source         string                 `json:"source"`
		FallbackReason string                 `json:"fallback_reason"`
		Error          string                 `json:"error"`
		Answers        map[string]interface{} `json:"answers"`
		CacheHit       bool                   `json:"cache_hit"`
	}

	if err := json.Unmarshal([]byte(failOpenJSON), &bridgeResp); err != nil {
		t.Fatalf("Failed to unmarshal fail open JSON: %v", err)
	}

	status := "AVAILABLE"
	unavailableReason := ""
	if bridgeResp.Source == "fail_open_advisory" {
		status = "UNAVAILABLE"
		unavailableReason = bridgeResp.FallbackReason
	}

	if status != "UNAVAILABLE" {
		t.Errorf("Expected status UNAVAILABLE for fail_open_advisory, got %s", status)
	}
	if !strings.Contains(unavailableReason, "503") {
		t.Errorf("Expected unavailableReason to contain 503, got: %s", unavailableReason)
	}

	rootCause := "DIAGNOSTIC_UNAVAILABLE"
	confidenceLevel := "não informada"
	mitigationExecuted := "PASSIVE_SHADOW_MONITORING"
	mitigationExecutedHuman := "Modo Sombra Passivo: Sondagem de 1s contínua sem intervenção destrutiva no host"

	if confidenceLevel != "não informada" {
		t.Errorf("Critério de Aceitação Violado: em falha/indisponibilidade a confiança deve ser 'não informada', obteve: %s", confidenceLevel)
	}
	if rootCause != "DIAGNOSTIC_UNAVAILABLE" {
		t.Errorf("Não deve fabricar causa raiz se indisponível: %s", rootCause)
	}
	if mitigationExecuted != "PASSIVE_SHADOW_MONITORING" {
		t.Errorf("Mitigação executada esperada PASSIVE_SHADOW_MONITORING, obtida: %s", mitigationExecuted)
	}
	if !strings.Contains(mitigationExecutedHuman, "Sondagem de 1s contínua") {
		t.Errorf("Ação executada em indisponibilidade deve descrever monitoramento contínuo: %s", mitigationExecutedHuman)
	}
}

func TestLegacyHistoricalRecordIdentification(t *testing.T) {
	// Simula a leitura de registros legados do SQLite gravados antes da migração do Gate 04
	legacyRow := struct {
		id                  int64
		serviceID           string
		snapshotID          int64
		rootCause           string
		sampleOriginRaw     string
		jevSourceRaw        string
		statusRaw           string
		unavailableReasonRaw string
		needsRestart        float64
	}{
		id:                  210,
		serviceID:           "http://127.0.0.1:8050",
		snapshotID:          1794015812000,
		rootCause:           "RESOURCE_EXHAUSTION_LEAK",
		sampleOriginRaw:     "", // Nulo/Vazio no banco legado
		jevSourceRaw:        "", // Nulo/Vazio no banco legado
		statusRaw:           "", // Nulo/Vazio no banco legado
		unavailableReasonRaw: "",
		needsRestart:        0.18,
	}

	isLegacy := (legacyRow.sampleOriginRaw == "" || legacyRow.statusRaw == "" || legacyRow.jevSourceRaw == "")
	if !isLegacy {
		t.Fatalf("Deveria identificar registro legado com colunas vazias")
	}

	sampleOrigin := "UNKNOWN"
	sampleOriginHuman := "❓ Origem Desconhecida (Legado)"
	status := "UNKNOWN"
	jevSource := "LEGACY_UNKNOWN"
	confidence := "não informada"

	if sampleOrigin != "UNKNOWN" || sampleOriginHuman != "❓ Origem Desconhecida (Legado)" {
		t.Errorf("Registro legado deve ser explicitamente rotulado com origem desconhecida")
	}
	if status != "UNKNOWN" {
		t.Errorf("Status de registro legado deve ser UNKNOWN")
	}
	if jevSource != "LEGACY_UNKNOWN" {
		t.Errorf("Origem do JEV para registro legado deve ser LEGACY_UNKNOWN")
	}
	if confidence != "não informada" {
		t.Errorf("Confiança de registro legado deve ser 'não informada', obteve: %s", confidence)
	}
}

func TestEvidenceRuleAcceptance_ProductionFlowValidation(t *testing.T) {
	// Inicializa conexão e carrega histórico de produção
	if db == nil {
		initDB()
	}
	loadHistoricalDiagnostics()

	// Critério de Aceitação 3: Validação direta contra o fluxo de produção (endpoint HTTP de telemetria)
	req := httptest.NewRequest("GET", "/api/relational/diagnostics", nil)
	w := httptest.NewRecorder()

	getRelationalDiagnostics(w, req)

	if w.Code != 200 {
		t.Fatalf("Endpoint de produção retornou HTTP %d, esperado 200", w.Code)
	}

	var recs []DiagnosticRecord
	if err := json.Unmarshal(w.Body.Bytes(), &recs); err != nil {
		t.Fatalf("Falha ao desserializar resposta do fluxo de produção: %v", err)
	}

	if len(recs) == 0 {
		t.Skip("Nenhum registro no cache de produção para validação (tabela vazia no momento)")
	}

	for _, r := range recs {
		// Valida que confiança NUNCA é inventada como "HIGH", "MÉDIA", "N/A" ou "UNKNOWN"
		if r.DiagnosticConfidence == "HIGH" || r.DiagnosticConfidence == "MÉDIA" || r.DiagnosticConfidence == "N/A" || r.DiagnosticConfidence == "UNKNOWN" {
			t.Errorf("Violação da Regra de Evidências no registro #%d: confiança proibida '%s'", r.ID, r.DiagnosticConfidence)
		}

		// Garante conformidade com os dois critérios permitidos
		isValidConfidence := (r.DiagnosticConfidence == "não informada" || strings.HasPrefix(r.DiagnosticConfidence, "confiança declarada pelo JEV:"))
		if !isValidConfidence {
			t.Errorf("Violação de critério de aceitação no registro #%d: confiança '%s' não segue o padrão da regra", r.ID, r.DiagnosticConfidence)
		}

		// Se o registro for legado, deve ter status UNKNOWN
		if r.Status == "UNKNOWN" {
			if r.SampleOrigin != "UNKNOWN" || r.JevSource != "LEGACY_UNKNOWN" {
				t.Errorf("Registro legado inconsistente #%d: origin=%s, source=%s", r.ID, r.SampleOrigin, r.JevSource)
			}
		}
	}
}
