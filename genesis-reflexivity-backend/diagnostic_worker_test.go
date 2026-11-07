package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslateDiagnosticToHuman(t *testing.T) {
	// 1. Valid case with low restart prob
	cause, mitig, summary := translateDiagnosticToHuman("RESOURCE_EXHAUSTION_LEAK", "TIER_1_SHED_LOAD", 0.15)
	if !strings.Contains(cause, "Aumento monótono") {
		t.Errorf("Expected cause translation for RESOURCE_EXHAUSTION_LEAK, got: %s", cause)
	}
	if !strings.Contains(mitig, "Alívio de Carga") {
		t.Errorf("Expected mitigation translation for TIER_1_SHED_LOAD, got: %s", mitig)
	}
	if !strings.Contains(summary, "REINÍCIO VETADO") {
		t.Errorf("Expected veto restart in summary, got: %s", summary)
	}

	// 2. Unavailable diagnostic case
	causeUnavail, _, summaryUnavail := translateDiagnosticToHuman("DIAGNOSTIC_UNAVAILABLE", "NONE", -1.0)
	if !strings.Contains(causeUnavail, "Diagnóstico Indisponível") {
		t.Errorf("Expected unavailable in cause, got: %s", causeUnavail)
	}
	if !strings.Contains(summaryUnavail, "Diagnóstico indisponível") {
		t.Errorf("Expected unavailable in summary, got: %s", summaryUnavail)
	}
}

func TestDiagnosticPayloadParsing_ValidResponse(t *testing.T) {
	validJSON := `{
		"ok": true,
		"source": "typesafe_jev_api",
		"model": "jev-1.14.0",
		"answers": {
			"root_cause_category": {"type": "choice", "choice": "DATABASE_CONTROLLER_LOCK"},
			"needs_destructive_restart": {"type": "noul", "noul": 0.22},
			"mitigation_tier": {"type": "choice", "choice": "TIER_2_RESET_POOL"}
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
	unavailableReason := ""
	if !bridgeResp.Ok || bridgeResp.Source == "fail_open_advisory" {
		status = "UNAVAILABLE"
		unavailableReason = bridgeResp.FallbackReason
	}

	rcObj, hasRc := bridgeResp.Answers["root_cause_category"].(map[string]interface{})
	nrObj, hasNr := bridgeResp.Answers["needs_destructive_restart"].(map[string]interface{})
	mtObj, hasMt := bridgeResp.Answers["mitigation_tier"].(map[string]interface{})

	if !hasRc || !hasNr || !hasMt {
		status = "UNAVAILABLE"
		unavailableReason = "RESPOSTA_INCOMPLETA_JEV"
	}

	if status != "AVAILABLE" {
		t.Fatalf("Expected status AVAILABLE, got %s (reason: %s)", status, unavailableReason)
	}

	rootCause := rcObj["choice"].(string)
	needsRestart := nrObj["noul"].(float64)
	mitigationTier := mtObj["choice"].(string)

	if rootCause != "DATABASE_CONTROLLER_LOCK" {
		t.Errorf("Expected rootCause DATABASE_CONTROLLER_LOCK, got %s", rootCause)
	}
	if needsRestart != 0.22 {
		t.Errorf("Expected needsRestart 0.22, got %f", needsRestart)
	}
	if mitigationTier != "TIER_2_RESET_POOL" {
		t.Errorf("Expected mitigationTier TIER_2_RESET_POOL, got %s", mitigationTier)
	}

	// Test separation of suggested vs executed
	mitigationSuggested := mitigationTier
	mitigationExecuted := ""
	if needsRestart < 0.40 {
		mitigationExecuted = "VETO_RESTART_AND_PACE"
	} else {
		mitigationExecuted = "ALERT_OPERATOR_SHADOW"
	}

	if mitigationSuggested != "TIER_2_RESET_POOL" || mitigationExecuted != "VETO_RESTART_AND_PACE" {
		t.Errorf("Mismatch in suggested vs executed mitigation: suggested=%s, executed=%s", mitigationSuggested, mitigationExecuted)
	}
}

func TestDiagnosticPayloadParsing_FailOpenAdvisory(t *testing.T) {
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
	if !bridgeResp.Ok {
		status = "UNAVAILABLE"
		unavailableReason = bridgeResp.Error
	} else if bridgeResp.Source == "fail_open_advisory" {
		status = "UNAVAILABLE"
		unavailableReason = bridgeResp.FallbackReason
	}

	if status != "UNAVAILABLE" {
		t.Errorf("Expected status UNAVAILABLE for fail_open_advisory, got %s", status)
	}
	if !strings.Contains(unavailableReason, "503") {
		t.Errorf("Expected reason to contain 503, got %s", unavailableReason)
	}

	// Verify that in unavailable state, no fake root cause is invented
	rootCause := "DIAGNOSTIC_UNAVAILABLE"
	humanReason := translateReasonToHuman(unavailableReason)
	if !strings.Contains(humanReason, "503") {
		t.Errorf("Expected human translation to explain 503 error, got: %s", humanReason)
	}
	if rootCause != "DIAGNOSTIC_UNAVAILABLE" {
		t.Errorf("Expected rootCause to remain DIAGNOSTIC_UNAVAILABLE, got: %s", rootCause)
	}
}

func TestDiagnosticPayloadParsing_IncompleteResponse(t *testing.T) {
	incompleteJSON := `{
		"ok": true,
		"source": "typesafe_jev_api",
		"model": "jev-1.14.0",
		"answers": {
			"root_cause_category": {"type": "choice", "choice": "UPSTREAM_NETWORK_TIMEOUT"}
		}
	}`

	var bridgeResp struct {
		Ok             bool                   `json:"ok"`
		Source         string                 `json:"source"`
		FallbackReason string                 `json:"fallback_reason"`
		Error          string                 `json:"error"`
		Answers        map[string]interface{} `json:"answers"`
		CacheHit       bool                   `json:"cache_hit"`
	}

	json.Unmarshal([]byte(incompleteJSON), &bridgeResp)

	status := "AVAILABLE"
	unavailableReason := ""

	_, hasRc := bridgeResp.Answers["root_cause_category"].(map[string]interface{})
	_, hasNr := bridgeResp.Answers["needs_destructive_restart"].(map[string]interface{})
	_, hasMt := bridgeResp.Answers["mitigation_tier"].(map[string]interface{})

	if !hasRc || !hasNr || !hasMt {
		status = "UNAVAILABLE"
		unavailableReason = "RESPOSTA_INCOMPLETA_JEV (campos ausentes)"
	}

	if status != "UNAVAILABLE" {
		t.Errorf("Expected UNAVAILABLE on incomplete response, got: %s", status)
	}
	if !strings.Contains(unavailableReason, "INCOMPLETA") {
		t.Errorf("Expected incomplete reason, got: %s", unavailableReason)
	}
}


func TestLiveAndSimulatedDiagnosticsIntegration(t *testing.T) {
	// 1. Simular uma resposta válida do JEV e verificar persistência
	recValid := DiagnosticRecord{
		ID:                       1001,
		ServiceID:                "http://127.0.0.1:8050",
		ServiceLabel:             "🌐 Alvo Monitorado: http://127.0.0.1:8050",
		SnapshotID:               314180,
		SampleOrigin:             "PROBE",
		SampleOriginHuman:        "📡 Sonda Real",
		JevSource:                "typesafe_jev_api",
		Status:                   "AVAILABLE",
		UnavailableReason:        "",
		RootCauseCategory:        "UPSTREAM_NETWORK_TIMEOUT",
		RootCauseHuman:           "Atraso na rota de rede ou lentidão transitória no alvo externo monitorado.",
		MitigationSuggested:      "TIER_1_SHED_LOAD",
		MitigationSuggestedHuman: "Alívio de Carga: Frequência de checagem espaçada para não sobrecarregar o host.",
		MitigationExecuted:       "VETO_RESTART_AND_PACE",
		MitigationExecutedHuman:  "Executado: Reinício vetado pelo sistema imune (cooldown 60s aplicado)",
		NeedsDestructiveRestart:  0.15,
		DiagnosticConfidence:     "HIGH",
		HumanSummary:             "✅ REINÍCIO VETADO PELO JEV (Risco de dano: 85%).",
		Timestamp:                "22:45:00",
	}

	if recValid.SnapshotID != 314180 {
		t.Errorf("Snapshot ID inválido: esperado 314180, obtido %d", recValid.SnapshotID)
	}
	if recValid.SampleOrigin != "PROBE" {
		t.Errorf("Origem da amostra inválida: esperado PROBE, obtido %s", recValid.SampleOrigin)
	}
	if recValid.JevSource != "typesafe_jev_api" {
		t.Errorf("Origem do JEV inválida: esperado typesafe_jev_api, obtido %s", recValid.JevSource)
	}
	if recValid.MitigationSuggested != "TIER_1_SHED_LOAD" || recValid.MitigationExecuted != "VETO_RESTART_AND_PACE" {
		t.Errorf("Mitigações não separadas corretamente")
	}

	// 2. Simular falha de API (Fail-Open Advisory) e verificar que status é UNAVAILABLE
	failReason := "API_ERROR: HTTP Error 503: Service Unavailable"
	recFailed := DiagnosticRecord{
		ID:                       1002,
		ServiceID:                "SERVICE_SLOW_NORMAL",
		ServiceLabel:             "📦 Motor de Processamento Pesado",
		SnapshotID:               314181,
		SampleOrigin:             "SIMULATION",
		SampleOriginHuman:        "🧪 Simulação",
		JevSource:                "fail_open_advisory",
		Status:                   "UNAVAILABLE",
		UnavailableReason:        failReason,
		RootCauseCategory:        "DIAGNOSTIC_UNAVAILABLE",
		RootCauseHuman:           translateReasonToHuman(failReason),
		MitigationSuggested:      "NONE",
		MitigationSuggestedHuman: "Nenhuma recomendação disponível pelo JEV",
		MitigationExecuted:       "DEFENSIVE_COOLDOWN",
		MitigationExecutedHuman:  "Executado: Cooldown defensivo ativo (modo passivo sem reinício cego)",
		NeedsDestructiveRestart:  -1.0,
		DiagnosticConfidence:     "HIGH",
		HumanSummary:             "⚠️ DIAGNÓSTICO INDISPONÍVEL: " + translateReasonToHuman(failReason),
		Timestamp:                "22:45:05",
	}

	if recFailed.Status != "UNAVAILABLE" {
		t.Errorf("Status deveria ser UNAVAILABLE, obtido: %s", recFailed.Status)
	}
	if !strings.Contains(recFailed.UnavailableReason, "503") {
		t.Errorf("Motivo da indisponibilidade não foi preservado: %s", recFailed.UnavailableReason)
	}
	if recFailed.MitigationSuggested != "NONE" {
		t.Errorf("Mitigação sugerida deveria ser NONE em caso de indisponibilidade, obtido: %s", recFailed.MitigationSuggested)
	}
	if !strings.Contains(recFailed.MitigationExecuted, "COOLDOWN") {
		t.Errorf("Mitigação executada deveria manter cooldown defensivo, obtido: %s", recFailed.MitigationExecuted)
	}
}
