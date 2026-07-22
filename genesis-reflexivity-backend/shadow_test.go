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
