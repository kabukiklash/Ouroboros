package main

import (
	"math"
)

type ShadowEvaluation struct {
	SnapshotID                 int64
	ServiceID                  string
	SnapshotHash               string
	FormulaVersion             string
	AbsolutePressure           float64
	AbsolutePressureLevel      string
	RelationalShock            float64
	RelationalShockLevel       string
	DirectionalDrift           float64
	FaultSignal                string
	MemoryConfidence           string
	ExperimentalInterpretation string
}

func evaluateRelationalShadow_v01(state RelationalState, snapshotID int64, sampleCount int) ShadowEvaluation {
	// 1. Memory Confidence
	memConf := "COLD"
	if sampleCount >= 50 {
		memConf = "STABLE"
	} else if sampleCount >= 10 {
		memConf = "WARM"
	}

	// 2. Absolute Pressure (CkScore)
	// Como a entropia pune internamente ou reflete a latencia, o CkScore é a Pressao Absoluta.
	entropy := state.Entropy
	if entropy < 1.0 {
		entropy = 1.0
	}
	de := entropy / math.Log(entropy+2.0)
	ckScore := de * 1.2
	absPress := ckScore

	pressLevel := "LOW"
	if absPress > 70 {
		pressLevel = "CRITICAL"
	} else if absPress > 40 {
		pressLevel = "HIGH"
	} else if absPress > 20 {
		pressLevel = "MEDIUM"
	}

	// 3. Relational Shock & Directional Drift
	alpha := state.AlphaAtTime
	alphaSafe := math.Max(alpha, 0.0001)

	relShock := (math.Abs(ckScore-alpha) / alphaSafe) * 100.0
	dirDrift := ((ckScore - alpha) / alphaSafe) * 100.0

	shockLevel := "LOW"
	if relShock > 200 {
		shockLevel = "CRITICAL"
	} else if relShock > 50 {
		shockLevel = "HIGH"
	} else if relShock > 20 {
		shockLevel = "MEDIUM"
	}

	// 4. Fault Signal
	faultSignal := "NONE"
	if state.SampleOk == 0 {
		if state.ErrorKind != "" {
			faultSignal = state.ErrorKind
		} else {
			faultSignal = "UNKNOWN_FAULT"
		}
	}

	// 5. Experimental Interpretation
	interpretation := ""
	if faultSignal != "NONE" {
		interpretation = "Falha categórica real. Deve acionar defesa independente da métrica."
	} else {
		if dirDrift < -30 {
			interpretation = "Recuperação relacional (melhorando)."
		} else if (shockLevel == "CRITICAL" || shockLevel == "HIGH") && dirDrift > 0 {
			interpretation = "Degradação relacional severa (Choque de Piora)."
		} else if (pressLevel == "CRITICAL" || pressLevel == "HIGH") && shockLevel == "LOW" {
			interpretation = "Serviço naturalmente pesado, mas sem choque relevante."
		} else if pressLevel == "LOW" && shockLevel == "LOW" {
			interpretation = "Serviço rápido e estável."
		} else {
			interpretation = "Instabilidade relacional leve/moderada."
		}
	}

	return ShadowEvaluation{
		SnapshotID:                 snapshotID,
		ServiceID:                  state.ServiceID,
		SnapshotHash:               "", // Será injetado depois
		FormulaVersion:             "relational_formula_v0.1",
		AbsolutePressure:           absPress,
		AbsolutePressureLevel:      pressLevel,
		RelationalShock:            relShock,
		RelationalShockLevel:       shockLevel,
		DirectionalDrift:           dirDrift,
		FaultSignal:                faultSignal,
		MemoryConfidence:           memConf,
		ExperimentalInterpretation: interpretation,
	}
}
