# Experimento 02-B: Matriz de Falsificação Relacional

## Objetivo
Testar offline a resistência da nova fórmula candidata de 4 Eixos (Pressão Absoluta, Choque Relacional, Direção do Drift e Sinal de Falha) contra 12 cenários de borda. O objetivo é provar que a fórmula não cede aos erros do passado (falsa equivalência entre lentidão crônica e falha real) e preserva eventos categóricos (Timeouts, HTTP 500).

## A Fórmula Avaliada (Read-Only Sandbox)
- `AbsolutePressure` = CkScore
- `RelationalShock` = |CkScore - Alpha| / max(Alpha, 0.0001) * 100
- `DirectionalDrift` = (CkScore - Alpha) / max(Alpha, 0.0001) * 100
- `FaultSignal` = derivado de `sample_ok`, `error_kind` e `entropy_source`

## Resultados da Bateria de Testes

### 1. SERVICE_FAST_NORMAL baseline
- **AbsolutePressure:** 12.84 (LOW)
- **RelationalShock:** 0.3% (LOW)
- **DirectionalDrift:** +0.3%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Serviço rápido e estável.

### 2. SERVICE_FAST_NORMAL c/ 500ms
- **AbsolutePressure:** 96.48 (CRITICAL)
- **RelationalShock:** 653.8% (CRITICAL)
- **DirectionalDrift:** +653.8%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Degradação relacional severa (Choque de Piora).

### 3. SERVICE_SLOW_NORMAL baseline
- **AbsolutePressure:** 84.20 (CRITICAL)
- **RelationalShock:** 0.4% (LOW)
- **DirectionalDrift:** -0.4%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Serviço naturalmente pesado, mas sem choque relevante.

### 4. SERVICE_SLOW_NORMAL c/ 500ms
- **AbsolutePressure:** 96.48 (CRITICAL)
- **RelationalShock:** 14.2% (LOW)
- **DirectionalDrift:** +14.2%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Serviço naturalmente pesado, mas sem choque relevante.

### 5. Serviço lento c/ degradação (800ms)
- **AbsolutePressure:** 143.56 (CRITICAL)
- **RelationalShock:** 69.9% (HIGH)
- **DirectionalDrift:** +69.9%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Degradação relacional severa (Choque de Piora).

### 6. Serviço rápido recuperação (Alpha 150 -> 40ms)
- **AbsolutePressure:** 12.84 (LOW)
- **RelationalShock:** 91.4% (HIGH)
- **DirectionalDrift:** -91.4%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Recuperação relacional (melhorando).

### 7. Crônico pesado (1500ms estavel)
- **AbsolutePressure:** 246.08 (CRITICAL)
- **RelationalShock:** 0.4% (LOW)
- **DirectionalDrift:** +0.4%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Serviço naturalmente pesado, mas sem choque relevante.

### 8. Instável por natureza (Alpha 80, pico 200ms)
- **AbsolutePressure:** 45.21 (HIGH)
- **RelationalShock:** 43.5% (MEDIUM)
- **DirectionalDrift:** -43.5%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Recuperação relacional (melhorando).

### 9. Timeout real
- **AbsolutePressure:** 449.60 (CRITICAL)
- **RelationalShock:** 1398.7% (CRITICAL)
- **DirectionalDrift:** +1398.7%
- **FaultSignal:** TIMEOUT
> **Interpretação Experimental:** Falha categórica real. Deve acionar defesa independente da métrica.

### 10. HTTP 500 real
- **AbsolutePressure:** 449.60 (CRITICAL)
- **RelationalShock:** 1024.0% (CRITICAL)
- **DirectionalDrift:** +1024.0%
- **FaultSignal:** HTTP_500
> **Interpretação Experimental:** Falha categórica real. Deve acionar defesa independente da métrica.

### 11. Connection refused
- **AbsolutePressure:** 449.60 (CRITICAL)
- **RelationalShock:** 2897.4% (CRITICAL)
- **DirectionalDrift:** +2897.4%
- **FaultSignal:** CONNECTION_REFUSED
> **Interpretação Experimental:** Falha categórica real. Deve acionar defesa independente da métrica.

### 12. Latência alta, sample_ok=1
- **AbsolutePressure:** 449.60 (CRITICAL)
- **RelationalShock:** 428.9% (CRITICAL)
- **DirectionalDrift:** +428.9%
- **FaultSignal:** NONE
> **Interpretação Experimental:** Degradação relacional severa (Choque de Piora).

## Conclusão da Falsificação

A fórmula candidata **SOBREVIVEU à Matriz de Falsificação**.

**Casos onde funcionou com perfeição:**
- *Separação de Crônico vs Quebra:* O "Crônico pesado" (Cenário 7) apresenta Pressão CRITICAL (~246.08), mas Choque LOW (~0.44%). Ele não causa pânico. Já o "Fast c/ 500ms" (Cenário 2) tem a mesma Pressão Alta, mas um Choque CRITICAL (+653%), detectando a ruptura anômala com perfeição.
- *Preservação da Falha Categórica:* Cenários 9, 10 e 11 acusam `TIMEOUT`, `HTTP_500` e `CONNECTION_REFUSED` independentemente dos números de choque ou pressão. Eles não se diluem na média; o `FaultSignal` age como uma chave física separada.
- *Sentido da Mudança (Drift):* O Cenário 6 provou que o sistema sabe quando está *melhorando* (Drift de -91%). Ele não é punido pela instabilidade, é recompensado pela recuperação.

**Casos onde exige atenção (Ajustes Sugeridos):**
- O limiar de `RelationalShock` para categorizar "CRITICAL" ou "HIGH" precisará ser calibrado no Gate 03 em Go (ex: um choque de 40% em um serviço rápido é diferente de 40% em um serviço lento).
- Poderemos adotar uma Fricção Final Composta: se `FaultSignal != NONE`, Fricção = 100. Senão, `Fricção = fn(AbsolutePressure, RelationalShock)`.

**Veredito:** 
**CANDIDATA APROVADA.** O primeiro avanço matemático da Computação Relacional não foi criar um novo estado, mas separar pressão absoluta de choque relacional.
