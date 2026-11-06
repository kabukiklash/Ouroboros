package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	_ "modernc.org/sqlite"
)

const OUROBOROS_RELATIONAL_SNAPSHOTS_ENABLED = true
const OUROBOROS_RELATIONAL_MAX_ROWS = 500000
const OUROBOROS_RELATIONAL_RETENTION_HOURS = 24
const OUROBOROS_RELATIONAL_PURGE_INTERVAL_SECONDS = 300

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

var serviceAlphas = struct {
	sync.RWMutex
	m map[string]float64
}{m: make(map[string]float64)}

var db *sql.DB
var dbMutex sync.Mutex

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "./temporal_audit.db")
	if err != nil {
		log.Fatal(err)
	}

	// Otimização Best-Effort e Sobrevivência
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA synchronous=NORMAL;")
	db.Exec("PRAGMA busy_timeout=5000;")

	db.Exec(`CREATE TABLE IF NOT EXISTS telemetry_history (id INTEGER PRIMARY KEY AUTOINCREMENT, service_id TEXT, ck_score REAL, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)`)

	// Gate 01: Relational Snapshot Layer
	db.Exec(`CREATE TABLE IF NOT EXISTS relational_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id TEXT NOT NULL,
		observed_at_ms INTEGER NOT NULL,
		inserted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		latency_ms REAL,
		sample_ok INTEGER NOT NULL DEFAULT 1,
		error_kind TEXT,
		entropy_source TEXT,
		entropy REAL NOT NULL,
		density REAL NOT NULL,
		ck_score REAL NOT NULL,
		alpha_at_time REAL NOT NULL,
		divergence REAL NOT NULL,
		friction REAL NOT NULL,
		state TEXT NOT NULL,
		action_taken TEXT,
		reason_json TEXT, 
		relations_json TEXT,
		impact_json TEXT, 
		schema_version INTEGER DEFAULT 1,
		previous_hash TEXT, 
		current_hash TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS relational_retention_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		rows_deleted INTEGER NOT NULL,
		policy_json TEXT,
		last_deleted_id INTEGER,
		last_deleted_hash TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS relational_shadow_evaluations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		snapshot_id INTEGER NOT NULL,
		service_id TEXT NOT NULL,
		snapshot_hash TEXT NOT NULL,
		formula_version TEXT NOT NULL,
		absolute_pressure REAL NOT NULL,
		absolute_pressure_level TEXT NOT NULL,
		relational_shock REAL NOT NULL,
		relational_shock_level TEXT NOT NULL,
		directional_drift REAL NOT NULL,
		fault_signal TEXT NOT NULL,
		memory_confidence TEXT NOT NULL,
		experimental_interpretation TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	
	db.Exec(`CREATE TABLE IF NOT EXISTS relational_shadow_retention_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		rows_deleted INTEGER NOT NULL,
		policy_json TEXT,
		last_deleted_id INTEGER,
		last_deleted_snapshot_id INTEGER,
		last_deleted_snapshot_hash TEXT
	)`)

	db.Exec(`CREATE TABLE IF NOT EXISTS relational_diagnostics_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service_id TEXT NOT NULL,
		snapshot_id INTEGER NOT NULL,
		root_cause TEXT NOT NULL,
		mitigation_applied TEXT NOT NULL,
		needs_restart_noul REAL NOT NULL,
		confidence_level TEXT NOT NULL,
		diagnostics_payload_json TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)

	db.Exec(`CREATE INDEX IF NOT EXISTS idx_telemetry_service_id ON telemetry_history(service_id, id DESC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_snapshots_service_id ON relational_snapshots(service_id, id DESC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_shadow_service_id ON relational_shadow_evaluations(service_id, id DESC)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_diag_service ON relational_diagnostics_events(service_id, id DESC)`)
	
	setAlpha("INTERNAL_KERNEL", 5.0)
}

type RelationalState struct {
	ServiceID     string
	ObservedAtMs  int64
	LatencyMs     float64
	SampleOk      int
	ErrorKind     string
	EntropySource string
	Entropy       float64
	Density       float64
	CkScore       float64
	AlphaAtTime   float64
	Divergence    float64
	Friction      float64
	State         string
	ActionTaken   string
	ReasonJSON    string
}

func calculateLocalRelationalState(serviceId string, latencyMs float64, sampleOk int, errorKind, entropySource string, currentFriction float64, historicalAlpha float64) RelationalState {
	entropy := latencyMs
	if entropy < 1.0 {
		entropy = 1.0
	}
	baseEntropyDensity := entropy / math.Log(entropy+2.0)
	ckScore := baseEntropyDensity * currentFriction
	divergence := math.Abs(ckScore - historicalAlpha)
	finalFriction := math.Min(100.0, ckScore+math.Pow(divergence, 1.5))

	status := "SAFE"
	if finalFriction >= 70.0 {
		status = "DENY"
	} else if finalFriction >= 20.0 {
		status = "WARN"
	}

	reasonJSON := fmt.Sprintf(`{"status":"%s","ck_score":%.4f,"divergence":%.4f,"alpha":%.4f}`, status, ckScore, divergence, historicalAlpha)

	return RelationalState{
		ServiceID:     serviceId,
		ObservedAtMs:  time.Now().UnixNano() / 1e6,
		LatencyMs:     latencyMs,
		SampleOk:      sampleOk,
		ErrorKind:     errorKind,
		EntropySource: entropySource,
		Entropy:       entropy,
		Density:       baseEntropyDensity,
		CkScore:       ckScore,
		AlphaAtTime:   historicalAlpha,
		Divergence:    divergence,
		Friction:      finalFriction,
		State:         status,
		ReasonJSON:    reasonJSON,
	}
}

func generateHash(state RelationalState, prevHash string, schemaVersion int) string {
	payload := fmt.Sprintf("%d|%s|%s|%d|%.2f|%d|%s|%s|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%s|%s|%s|%s|%s",
		schemaVersion, prevHash, state.ServiceID, state.ObservedAtMs, state.LatencyMs, state.SampleOk, state.ErrorKind, state.EntropySource, state.Entropy, state.Density, state.CkScore, state.AlphaAtTime, state.Divergence, state.Friction, state.State, state.ActionTaken, state.ReasonJSON, "{}", "{}")
	h := sha256.New()
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func insertSnapshot(state RelationalState) {
	if !OUROBOROS_RELATIONAL_SNAPSHOTS_ENABLED {
		return
	}
	go func() { // Best-effort, assíncrono
		dbMutex.Lock()
		defer dbMutex.Unlock()

		var prevHash sql.NullString
		db.QueryRow(`SELECT current_hash FROM relational_snapshots WHERE service_id = ? ORDER BY id DESC LIMIT 1`, state.ServiceID).Scan(&prevHash)

		ph := ""
		if prevHash.Valid {
			ph = prevHash.String
		}

		schemaVersion := 1
		currHash := generateHash(state, ph, schemaVersion)
		relationsJSON := fmt.Sprintf(`{"service_id":"%s","entropy_source":"%s","previous_hash":"%s"}`, state.ServiceID, state.EntropySource, ph)
		impactJSON := fmt.Sprintf(`{"state":"%s","friction":%.4f,"action_taken":"%s"}`, state.State, state.Friction, state.ActionTaken)

		res, err := db.Exec(`INSERT INTO relational_snapshots 
		(service_id, observed_at_ms, latency_ms, sample_ok, error_kind, entropy_source, entropy, density, ck_score, alpha_at_time, divergence, friction, state, action_taken, reason_json, relations_json, impact_json, schema_version, previous_hash, current_hash) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			state.ServiceID, state.ObservedAtMs, state.LatencyMs, state.SampleOk, state.ErrorKind, state.EntropySource, state.Entropy, state.Density, state.CkScore, state.AlphaAtTime, state.Divergence, state.Friction, state.State, state.ActionTaken, state.ReasonJSON, relationsJSON, impactJSON, schemaVersion, ph, currHash)

		if err != nil {
			log.Println("Gate 01 - Falha best-effort ao inserir snapshot:", err)
			return
		}

		lastId, errId := res.LastInsertId()
		if errId == nil {
			go executeShadowEvaluation(state, lastId, currHash)
		}
	}()
}

func executeShadowEvaluation(state RelationalState, snapshotID int64, currHash string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Gate 03-A - Panic recuperado na avaliacao shadow:", r)
		}
	}()

	var sampleCount int
	dbMutex.Lock()
	db.QueryRow(`SELECT count(*) FROM relational_snapshots WHERE service_id = ?`, state.ServiceID).Scan(&sampleCount)
	dbMutex.Unlock()

	eval := evaluateRelationalShadow_v01(state, snapshotID, sampleCount)
	eval.SnapshotHash = currHash

	dbMutex.Lock()
	defer dbMutex.Unlock()

	_, err := db.Exec(`INSERT INTO relational_shadow_evaluations 
		(snapshot_id, service_id, snapshot_hash, formula_version, absolute_pressure, absolute_pressure_level, relational_shock, relational_shock_level, directional_drift, fault_signal, memory_confidence, experimental_interpretation)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		eval.SnapshotID, eval.ServiceID, eval.SnapshotHash, eval.FormulaVersion, eval.AbsolutePressure, eval.AbsolutePressureLevel, eval.RelationalShock, eval.RelationalShockLevel, eval.DirectionalDrift, eval.FaultSignal, eval.MemoryConfidence, eval.ExperimentalInterpretation)

	if err != nil {
		log.Println("Gate 03-A - Falha ao inserir shadow evaluation:", err)
	} else {
		go dispatchGate04Diagnostics(state, eval.RelationalShockLevel, eval.RelationalShock, eval.DirectionalDrift)
	}
}

func getAlpha(id string) float64 {
	serviceAlphas.RLock()
	defer serviceAlphas.RUnlock()
	if val, ok := serviceAlphas.m[id]; ok {
		return val
	}
	return 50.0
}

func setAlpha(id string, val float64) {
	serviceAlphas.Lock()
	defer serviceAlphas.Unlock()
	serviceAlphas.m[id] = val
}

func temporalAuditWorker() {
	for {
		time.Sleep(5 * time.Second)
		serviceAlphas.RLock()
		var services []string
		for k := range serviceAlphas.m {
			services = append(services, k)
		}
		serviceAlphas.RUnlock()

		for _, srv := range services {
			var avg sql.NullFloat64
			dbMutex.Lock()
			err := db.QueryRow(`SELECT AVG(ck_score) FROM (SELECT ck_score FROM telemetry_history WHERE service_id = ? ORDER BY id DESC LIMIT 10)`, srv).Scan(&avg)
			dbMutex.Unlock()
			if err != nil {
				log.Println("DB Audit Error:", err)
			}
			if avg.Valid {
				curr := getAlpha(srv)
				next := (curr * 0.3) + (avg.Float64 * 0.7)
				setAlpha(srv, next)
			}
		}
		dbMutex.Lock()
		db.Exec(`DELETE FROM telemetry_history WHERE id NOT IN (SELECT id FROM telemetry_history ORDER BY id DESC LIMIT 500)`)
		dbMutex.Unlock()
	}
}

func temporalRetentionWorker() {
	for {
		time.Sleep(time.Duration(OUROBOROS_RELATIONAL_PURGE_INTERVAL_SECONDS) * time.Second)
		if !OUROBOROS_RELATIONAL_SNAPSHOTS_ENABLED {
			continue
		}

		dbMutex.Lock()
		cutoff := time.Now().Add(-time.Duration(OUROBOROS_RELATIONAL_RETENTION_HOURS) * time.Hour)
		cutoffStr := cutoff.Format("2006-01-02 15:04:05")

		var lastId sql.NullInt64
		var lastHash sql.NullString
		db.QueryRow(`SELECT id, current_hash FROM relational_snapshots WHERE inserted_at < ? OR id NOT IN (SELECT id FROM relational_snapshots ORDER BY id DESC LIMIT ?) ORDER BY id DESC LIMIT 1`, cutoffStr, OUROBOROS_RELATIONAL_MAX_ROWS).Scan(&lastId, &lastHash)

		resTime, _ := db.Exec(`DELETE FROM relational_snapshots WHERE inserted_at < ?`, cutoffStr)
		rowsTime, _ := resTime.RowsAffected()

		resCount, _ := db.Exec(`DELETE FROM relational_snapshots WHERE id NOT IN (SELECT id FROM relational_snapshots ORDER BY id DESC LIMIT ?)`, OUROBOROS_RELATIONAL_MAX_ROWS)
		rowsCount, _ := resCount.RowsAffected()

		totalDeleted := rowsTime + rowsCount

		if totalDeleted > 0 {
			policy := fmt.Sprintf(`{"max_rows":%d,"retention_hours":%d}`, OUROBOROS_RELATIONAL_MAX_ROWS, OUROBOROS_RELATIONAL_RETENTION_HOURS)
			db.Exec(`INSERT INTO relational_retention_events (rows_deleted, policy_json, last_deleted_id, last_deleted_hash) VALUES (?, ?, ?, ?)`, totalDeleted, policy, lastId.Int64, lastHash.String)
			log.Printf("Gate 01 - Expurgo Relacional executado: %d linhas removidas.", totalDeleted)
		}
		dbMutex.Unlock()
	}
}

func getShadowRetentionEnabled() bool {
	val := os.Getenv("OUROBOROS_SHADOW_RETENTION_ENABLED")
	if val == "false" {
		return false
	}
	return true
}

func getShadowMaxRows() int {
	val := os.Getenv("OUROBOROS_SHADOW_MAX_ROWS")
	if val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
		log.Println("Gate 03-B.1 - Aviso: OUROBOROS_SHADOW_MAX_ROWS invalido, usando default 500000.")
	}
	return 500000
}

func getShadowRetentionHours() int {
	val := os.Getenv("OUROBOROS_SHADOW_RETENTION_HOURS")
	if val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
		log.Println("Gate 03-B.1 - Aviso: OUROBOROS_SHADOW_RETENTION_HOURS invalido, usando default 24.")
	}
	return 24
}

func getShadowPurgeInterval() int {
	val := os.Getenv("OUROBOROS_SHADOW_PURGE_INTERVAL_SECONDS")
	if val != "" {
		if parsed, err := strconv.Atoi(val); err == nil && parsed > 0 {
			return parsed
		}
		log.Println("Gate 03-B.1 - Aviso: OUROBOROS_SHADOW_PURGE_INTERVAL_SECONDS invalido, usando default 300.")
	}
	return 300
}

func temporalShadowRetentionWorker() {
	for {
		time.Sleep(time.Duration(getShadowPurgeInterval()) * time.Second)
		if !getShadowRetentionEnabled() {
			continue
		}

		dbMutex.Lock()
		maxRows := getShadowMaxRows()
		retentionHours := getShadowRetentionHours()
		
		cutoff := time.Now().Add(-time.Duration(retentionHours) * time.Hour)
		cutoffStr := cutoff.Format("2006-01-02 15:04:05")

		var lastId, lastSnapId sql.NullInt64
		var lastHash sql.NullString
		db.QueryRow(`SELECT id, snapshot_id, snapshot_hash FROM relational_shadow_evaluations WHERE created_at < ? OR id NOT IN (SELECT id FROM relational_shadow_evaluations ORDER BY id DESC LIMIT ?) ORDER BY id DESC LIMIT 1`, cutoffStr, maxRows).Scan(&lastId, &lastSnapId, &lastHash)

		resTime, _ := db.Exec(`DELETE FROM relational_shadow_evaluations WHERE created_at < ?`, cutoffStr)
		rowsTime, _ := resTime.RowsAffected()

		resCount, _ := db.Exec(`DELETE FROM relational_shadow_evaluations WHERE id NOT IN (SELECT id FROM relational_shadow_evaluations ORDER BY id DESC LIMIT ?)`, maxRows)
		rowsCount, _ := resCount.RowsAffected()

		totalDeleted := rowsTime + rowsCount

		if totalDeleted > 0 {
			policy := fmt.Sprintf(`{"max_rows":%d,"retention_hours":%d}`, maxRows, retentionHours)
			db.Exec(`INSERT INTO relational_shadow_retention_events (rows_deleted, policy_json, last_deleted_id, last_deleted_snapshot_id, last_deleted_snapshot_hash) VALUES (?, ?, ?, ?, ?)`, totalDeleted, policy, lastId.Int64, lastSnapId.Int64, lastHash.String)
			log.Printf("Gate 03-B.1 - Expurgo Shadow executado: %d linhas removidas.", totalDeleted)
		}
		dbMutex.Unlock()
	}
}

// ---- STATES ----
type InternalMetrics struct {
	Time      string  `json:"time"`
	Friction  float64 `json:"friction"`
	CkScore   float64 `json:"ckScore"`
	Alpha     float64 `json:"alpha"`
	Entropy   float64 `json:"entropy"`
	Status    string  `json:"status"`
	ActionMsg string  `json:"actionMsg"`
}

type ExternalTarget struct {
	URL       string  `json:"url"`
	Latency   float64 `json:"latency"`
	Alpha     float64 `json:"alpha"`
	Friction  float64 `json:"friction"`
	Status    string  `json:"status"`
	ActionMsg string  `json:"actionMsg"`
}

type MasterPayload struct {
	Internal     InternalMetrics  `json:"internal"`
	External     []ExternalTarget `json:"external"`
	EngineActive bool             `json:"engineActive"`
}

var currentScenario = "NORMAL"
var scenarioMutex sync.RWMutex

var EngineActive = false
var engineMutex sync.RWMutex

var internalState InternalMetrics
var internalStateMutex sync.RWMutex

var externalStates = struct {
	sync.RWMutex
	m map[string]ExternalTarget
}{m: make(map[string]ExternalTarget)}

// ---- WORKERS ----
func simulationWorker() {
	go runExperiment("SERVICE_FAST_NORMAL", 30.0, 50.0)
	go runExperiment("SERVICE_SLOW_NORMAL", 400.0, 450.0)
}

func runExperiment(srv string, min, max float64) {
	denyCount := 0
	setAlpha(srv, (min+max)/2) // Pre-seed Alpha for faster warmup
	for {
		engineMutex.RLock()
		active := EngineActive
		engineMutex.RUnlock()
		if !active {
			time.Sleep(1 * time.Second)
			continue
		}

		scenarioMutex.RLock()
		sc := currentScenario
		scenarioMutex.RUnlock()

		latency := min + rand.Float64()*(max-min)
		if sc == "HIGH_LOAD_EXPECTED" {
			latency = max*1.8 + rand.Float64()*20.0
		} else if sc == "ATTACK" {
			latency = 500.0
		}

		sampleOk := 1
		errorKind := ""
		entropySrc := "SIMULATION"

		alpha := getAlpha(srv)
		rs := calculateLocalRelationalState(srv, latency, sampleOk, errorKind, entropySrc, 1.2, alpha)

		if rs.State == "DENY" {
			denyCount++
			if denyCount >= 3 {
				rs.ActionTaken = "GATILHO EXECUTADO: Reiniciando Kernel..."
				denyCount = 0
			}
		} else {
			denyCount = 0
		}

		dbMutex.Lock()
		db.Exec("INSERT INTO telemetry_history (service_id, ck_score) VALUES (?, ?)", srv, rs.CkScore)
		dbMutex.Unlock()

		insertSnapshot(rs)
		
		// Atualiza o estado interno para o gráfico APENAS se for o serviço mais rápido (para não sobrepor no UI)
		if srv == "SERVICE_FAST_NORMAL" {
			internalStateMutex.Lock()
			internalState = InternalMetrics{
				Time:      time.Now().Format("15:04:05"),
				Friction:  math.Round(rs.Friction*100) / 100,
				CkScore:   math.Round(rs.CkScore*100) / 100,
				Alpha:     math.Round(alpha*100) / 100,
				Entropy:   math.Round(rs.Entropy*100) / 100,
				Status:    rs.State,
				ActionMsg: rs.ActionTaken,
			}
			internalStateMutex.Unlock()
		}

		time.Sleep(1 * time.Second)
	}
}

func pingWorker(url string) {
	setAlpha(url, 50.0)
	denyCount := 0
	log.Printf("Iniciando monitoramento ativo externo em: %s", url)

	for {
		externalStates.RLock()
		_, exists := externalStates.m[url]
		externalStates.RUnlock()
		if !exists {
			log.Printf("Parando sentinela ativo em: %s", url)
			return
		}

		engineMutex.RLock()
		active := EngineActive
		engineMutex.RUnlock()
		if !active {
			time.Sleep(1 * time.Second)
			continue
		}

		start := time.Now()
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url)
		latency := float64(time.Since(start).Milliseconds())

		sampleOk := 1
		errorKind := ""
		entropySrc := "REAL_NETWORK"

		if err != nil || resp.StatusCode >= 500 {
			latency = 3000.0 // Entropia Punitiva
			sampleOk = 0
			errorKind = "TIMEOUT_OR_5XX"
			entropySrc = "PUNITIVE_FALLBACK"
		}
		if resp != nil {
			resp.Body.Close()
		}

		alpha := getAlpha(url)
		rs := calculateLocalRelationalState(url, latency, sampleOk, errorKind, entropySrc, 1.2, alpha)

		if rs.State == "DENY" {
			denyCount++
			if denyCount >= 3 {
				rs.ActionTaken = "GATILHO EXECUTADO: Drop/Restart acionado!"
				denyCount = 0
			}
		} else {
			denyCount = 0
		}

		dbMutex.Lock()
		db.Exec("INSERT INTO telemetry_history (service_id, ck_score) VALUES (?, ?)", url, rs.CkScore)
		dbMutex.Unlock()

		insertSnapshot(rs)

		state := ExternalTarget{
			URL:       url,
			Latency:   math.Round(latency),
			Alpha:     math.Round(alpha*100) / 100,
			Friction:  math.Round(rs.Friction*100) / 100,
			Status:    rs.State,
			ActionMsg: rs.ActionTaken,
		}

		externalStates.Lock()
		externalStates.m[url] = state
		externalStates.Unlock()

		time.Sleep(1 * time.Second)
	}
}

type ClientAction struct {
	Action   string `json:"action"`
	URL      string `json:"url"`
	Scenario string `json:"scenario"`
	Enabled  bool   `json:"enabled"`
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	go func() {
		for {
			var msg ClientAction
			if err := ws.ReadJSON(&msg); err != nil {
				break
			}
			if msg.Action == "ADD_TARGET" && msg.URL != "" {
				externalStates.RLock()
				_, exists := externalStates.m[msg.URL]
				externalStates.RUnlock()
				if !exists {
					externalStates.Lock()
					externalStates.m[msg.URL] = ExternalTarget{URL: msg.URL, Status: "STARTING"}
					externalStates.Unlock()
					go pingWorker(msg.URL)
				}
			}
			if msg.Action == "REMOVE_TARGET" && msg.URL != "" {
				externalStates.Lock()
				delete(externalStates.m, msg.URL)
				externalStates.Unlock()
			}
			if msg.Action == "TOGGLE_ENGINE" {
				engineMutex.Lock()
				EngineActive = msg.Enabled
				engineMutex.Unlock()
				log.Printf("Motor Relacional alterado para: %v", msg.Enabled)
			}
			if msg.Scenario != "" {
				scenarioMutex.Lock()
				currentScenario = msg.Scenario
				scenarioMutex.Unlock()
			}
		}
	}()

	for {
		externalStates.RLock()
		var extArr []ExternalTarget
		for _, v := range externalStates.m {
			extArr = append(extArr, v)
		}
		externalStates.RUnlock()

		internalStateMutex.RLock()
		is := internalState
		internalStateMutex.RUnlock()

		engineMutex.RLock()
		eActive := EngineActive
		engineMutex.RUnlock()

		payload := MasterPayload{
			Internal:     is,
			External:     extArr,
			EngineActive: eActive,
		}

		if err := ws.WriteJSON(payload); err != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	// Mantido endpoint original apenas para compatibilidade legada com frontend, se necessário.
	
	dbMutex.Lock()
	rows, err := db.Query(`SELECT id, service_id, ck_score, timestamp FROM telemetry_history ORDER BY id DESC LIMIT 500`)
	dbMutex.Unlock()
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	type HistoryRow struct {
		ID        int     `json:"id"`
		ServiceID string  `json:"serviceId"`
		CkScore   float64 `json:"ckScore"`
		Timestamp string  `json:"timestamp"`
	}
	
	var history []HistoryRow
	for rows.Next() {
		var row HistoryRow
		rows.Scan(&row.ID, &row.ServiceID, &row.CkScore, &row.Timestamp)
		history = append(history, row)
	}
	if history == nil {
		history = []HistoryRow{}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func getRelationalSnapshots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}
	
	afterIdStr := r.URL.Query().Get("after_id")
	afterId := 0
	if afterIdStr != "" {
		if a, err := strconv.Atoi(afterIdStr); err == nil {
			afterId = a
		}
	}

	serviceId := r.URL.Query().Get("service_id")
	
	query := "SELECT id, service_id, observed_at_ms, state, friction, ck_score, current_hash FROM relational_snapshots WHERE id > ?"
	args := []interface{}{afterId}
	
	if serviceId != "" {
		query += " AND service_id = ?"
		args = append(args, serviceId)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)
	
	dbMutex.Lock()
	rows, err := db.Query(query, args...)
	dbMutex.Unlock()
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	type SnapRow struct {
		ID           int     `json:"id"`
		ServiceID    string  `json:"serviceId"`
		ObservedAtMs int64   `json:"observedAtMs"`
		State        string  `json:"state"`
		Friction     float64 `json:"friction"`
		CkScore      float64 `json:"ckScore"`
		CurrentHash  string  `json:"currentHash"`
	}
	
	var snaps []SnapRow
	for rows.Next() {
		var row SnapRow
		rows.Scan(&row.ID, &row.ServiceID, &row.ObservedAtMs, &row.State, &row.Friction, &row.CkScore, &row.CurrentHash)
		snaps = append(snaps, row)
	}
	if snaps == nil {
		snaps = []SnapRow{}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snaps)
}

func getRelationalShadowEvaluations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	serviceId := r.URL.Query().Get("service_id")
	query := `SELECT id, snapshot_id, service_id, snapshot_hash, formula_version, absolute_pressure, absolute_pressure_level, relational_shock, relational_shock_level, directional_drift, fault_signal, memory_confidence, experimental_interpretation, created_at FROM relational_shadow_evaluations`
	var args []interface{}
	if serviceId != "" {
		query += " WHERE service_id = ?"
		args = append(args, serviceId)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	dbMutex.Lock()
	rows, err := db.Query(query, args...)
	dbMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ShadowRow struct {
		ID                         int64   `json:"id"`
		SnapshotID                 int64   `json:"snapshotId"`
		ServiceID                  string  `json:"serviceId"`
		SnapshotHash               string  `json:"snapshotHash"`
		FormulaVersion             string  `json:"formulaVersion"`
		AbsolutePressure           float64 `json:"absolutePressure"`
		AbsolutePressureLevel      string  `json:"absolutePressureLevel"`
		RelationalShock            float64 `json:"relationalShock"`
		RelationalShockLevel       string  `json:"relationalShockLevel"`
		DirectionalDrift           float64 `json:"directionalDrift"`
		FaultSignal                string  `json:"faultSignal"`
		MemoryConfidence           string  `json:"memoryConfidence"`
		ExperimentalInterpretation string  `json:"experimentalInterpretation"`
		CreatedAt                  string  `json:"createdAt"`
	}

	var evals []ShadowRow
	for rows.Next() {
		var row ShadowRow
		rows.Scan(
			&row.ID,
			&row.SnapshotID,
			&row.ServiceID,
			&row.SnapshotHash,
			&row.FormulaVersion,
			&row.AbsolutePressure,
			&row.AbsolutePressureLevel,
			&row.RelationalShock,
			&row.RelationalShockLevel,
			&row.DirectionalDrift,
			&row.FaultSignal,
			&row.MemoryConfidence,
			&row.ExperimentalInterpretation,
			&row.CreatedAt,
		)
		evals = append(evals, row)
	}
	if evals == nil {
		evals = []ShadowRow{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evals)
}

func getRelationalDiagnostics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	cachedDiagnostics.RLock()
	recs := make([]DiagnosticRecord, len(cachedDiagnostics.records))
	copy(recs, cachedDiagnostics.records)
	cachedDiagnostics.RUnlock()

	json.NewEncoder(w).Encode(recs)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	initDB()
	
	go temporalAuditWorker()
	go temporalRetentionWorker()
	go temporalShadowRetentionWorker()
	go simulationWorker()

	defaultTarget := "http://127.0.0.1:3000/v1/health"
	externalStates.Lock()
	externalStates.m[defaultTarget] = ExternalTarget{URL: defaultTarget, Status: "STARTING"}
	externalStates.Unlock()
	go pingWorker(defaultTarget)

	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api/history", getHistory)
	http.HandleFunc("/api/relational/snapshots", getRelationalSnapshots)       // Gate 01 read-only endpoint
	http.HandleFunc("/api/relational/shadow", getRelationalShadowEvaluations) // Gate 03 read-only endpoint
	http.HandleFunc("/api/relational/diagnostics", getRelationalDiagnostics) // Gate 04 Jev Diagnostics & Mitigation endpoint

	log.Println("=============================================================")
	log.Println(" Genesis Reflexivity (Hybrid: Simulação + API Probing) ONLINE")
	log.Println(" Gate 01: Relational Snapshot Layer [ATIVO]")
	log.Println(" Gate 03: Relational Shadow Evaluator [ATIVO]")
	log.Println(" Escutando na porta :8082")
	log.Println("=============================================================")
	http.ListenAndServe(":8082", nil)
}
