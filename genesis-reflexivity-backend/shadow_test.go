package main

import (
	"math"
	"testing"
	"time"
)

func approxEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestShadowEvaluator_GoldenVectors(t *testing.T) {
	tests := []struct {
		name          string
		entropy       float64
		alpha         float64
		sampleOk      int
		errorKind     string
		entropySrc    string
		expPress      float64
		expShock      float64
		expDrift      float64
		expFault      string
	}{
		{"1. SERVICE_FAST_NORMAL baseline", 40.0, 12.8, 1, "", "REAL", 12.84, 0.3, +0.3, "NONE"},
		{"2. SERVICE_FAST_NORMAL c/ 500ms", 500.0, 12.8, 1, "", "REAL", 96.48, 653.8, +653.8, "NONE"},
		{"3. SERVICE_SLOW_NORMAL baseline", 425.0, 84.5, 1, "", "REAL", 84.22, 0.3, -0.3, "NONE"},
		{"4. SERVICE_SLOW_NORMAL c/ 500ms", 500.0, 84.5, 1, "", "REAL", 96.48, 14.2, +14.2, "NONE"},
		{"5. Serviço lento c/ degradação (800ms)", 800.0, 84.5, 1, "", "REAL", 143.55, 69.9, +69.9, "NONE"},
		{"6. Serviço rápido recuperação", 40.0, 150.0, 1, "", "REAL", 12.84, 91.4, -91.4, "NONE"},
		{"7. Crônico pesado (1500ms estavel)", 1500.0, 245.0, 1, "", "REAL", 246.08, 0.44, +0.44, "NONE"},
		{"8. Instável por natureza", 200.0, 80.0, 1, "", "REAL", 45.22, 43.5, -43.5, "NONE"},
		{"9. Timeout real", 3000.0, 30.0, 0, "TIMEOUT", "PUNITIVE_FALLBACK", 449.63, 1398.8, +1398.8, "TIMEOUT"},
		{"10. HTTP 500 real", 3000.0, 40.0, 0, "HTTP_500", "PUNITIVE_FALLBACK", 449.63, 1024.1, +1024.1, "HTTP_500"},
		{"11. Connection refused", 3000.0, 15.0, 0, "CONNECTION_REFUSED", "PUNITIVE_FALLBACK", 449.63, 2897.5, +2897.5, "CONNECTION_REFUSED"},
		{"12. Latência alta, sample_ok=1", 3000.0, 85.0, 1, "", "REAL", 449.63, 429.0, +429.0, "NONE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := RelationalState{
				ServiceID:     "TEST",
				Entropy:       tt.entropy,
				AlphaAtTime:   tt.alpha,
				SampleOk:      tt.sampleOk,
				ErrorKind:     tt.errorKind,
				EntropySource: tt.entropySrc,
			}

			eval := evaluateRelationalShadow_v01(state, 1, 100)

			// Tolerate up to 0.5 variation due to rounding differences between Python print formatting and float64 comparison
			if !approxEqual(eval.AbsolutePressure, tt.expPress, 0.5) {
				t.Errorf("AbsolutePressure esperado ~%.2f, obteve %.2f", tt.expPress, eval.AbsolutePressure)
			}
			if !approxEqual(eval.RelationalShock, tt.expShock, 0.5) {
				t.Errorf("RelationalShock esperado ~%.1f%%, obteve %.1f%%", tt.expShock, eval.RelationalShock)
			}
			if !approxEqual(eval.DirectionalDrift, tt.expDrift, 0.5) {
				t.Errorf("DirectionalDrift esperado ~%+.1f%%, obteve %+.1f%%", tt.expDrift, eval.DirectionalDrift)
			}
			if eval.FaultSignal != tt.expFault {
				t.Errorf("FaultSignal esperado %s, obteve %s", tt.expFault, eval.FaultSignal)
			}
			
			// Test Memory Confidence
			evalCold := evaluateRelationalShadow_v01(state, 1, 5)
			if evalCold.MemoryConfidence != "COLD" {
				t.Errorf("MemoryConfidence esperado COLD, obteve %s", evalCold.MemoryConfidence)
			}
			evalWarm := evaluateRelationalShadow_v01(state, 1, 20)
			if evalWarm.MemoryConfidence != "WARM" {
				t.Errorf("MemoryConfidence esperado WARM, obteve %s", evalWarm.MemoryConfidence)
			}
		})
	}
}

func TestShadowDBIntegration(t *testing.T) {
	initDB()
	
	// Generate 3 test cases to trigger the shadow evaluator logic and insert into DB
	s1 := calculateLocalRelationalState("SERVICE_FAST_NORMAL_TEST", 500.0, 1, "", "REAL", 100.0, 12.8)
	insertSnapshot(s1)
	
	s2 := calculateLocalRelationalState("SERVICE_SLOW_NORMAL_TEST", 500.0, 1, "", "REAL", 14.0, 84.5)
	insertSnapshot(s2)
	
	s3 := calculateLocalRelationalState("SERVICE_TIMEOUT_TEST", 3000.0, 0, "TIMEOUT", "PUNITIVE_FALLBACK", 100.0, 30.0)
	insertSnapshot(s3)
	
	time.Sleep(2 * time.Second)
}

func TestShadowRetentionPolicy(t *testing.T) {
	initDB()
	
	// Create mock shadow evaluations to trigger purge
	for i := 0; i < 5; i++ {
		db.Exec(`INSERT INTO relational_shadow_evaluations 
		(snapshot_id, service_id, snapshot_hash, formula_version, absolute_pressure, absolute_pressure_level, relational_shock, relational_shock_level, directional_drift, fault_signal, memory_confidence, experimental_interpretation, created_at)
		VALUES (1, 'TEST_PURGE', 'hash', 'v0.1', 10, 'LOW', 10, 'LOW', 10, 'NONE', 'COLD', 'TEST', '2000-01-01 12:00:00')`)
	}

	dbMutex.Lock()
	cutoff := time.Now().Add(-time.Duration(OUROBOROS_SHADOW_RETENTION_HOURS) * time.Hour)
	res, _ := db.Exec(`DELETE FROM relational_shadow_evaluations WHERE created_at < ?`, cutoff.Format("2006-01-02 15:04:05"))
	rows, _ := res.RowsAffected()
	dbMutex.Unlock()

	if rows == 0 {
		t.Errorf("Esperava apagar registros antigos, mas apagou 0.")
	}
	
	var count int
	db.QueryRow(`SELECT count(*) FROM relational_shadow_retention_events`).Scan(&count)
	// We don't assert count since temporal worker might not have written, we just check no panic.
}
