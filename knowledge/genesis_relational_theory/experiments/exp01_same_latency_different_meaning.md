# Experimento 01: Latência igual, significado diferente

## Objetivo
Validar a hipótese central do PER: "o dado isolado não carrega significado suficiente; o significado nasce da relação entre dado, histórico, regra, contexto e impacto."

## Setup Experimental
Foram criados dois serviços simulados rodando em paralelo no Ouroboros (Gate 01):
- **SERVICE_FAST_NORMAL:** Baseline de 30ms a 50ms.
- **SERVICE_SLOW_NORMAL:** Baseline de 400ms a 450ms.

Após um período de *warmup* de 20 segundos para estabilização da memória (Alpha), injetou-se uma latência artificial fixa de **500ms** em ambos os serviços simultaneamente.

## Resultados Coletados (Snapshots Reais)

### Durante o Warmup (Normalidade)
- **SERVICE_FAST_NORMAL (Latência ~40ms):**
  - CkScore: ~12.8
  - Fricção: ~12.8
  - Estado: **SAFE**
- **SERVICE_SLOW_NORMAL (Latência ~425ms):**
  - CkScore: ~84.5
  - Fricção: ~84.5
  - Estado: **DENY** *(Observação Crítica 1)*

### Após a Injeção do Pico (500ms)
- **SERVICE_FAST_NORMAL (Latência 500ms):**
  - Alpha Histórico: ~12.8 (em transição)
  - CkScore: 96.48
  - Divergência: ~83.6
  - Fricção: 100.0 (Cap Máximo)
  - Estado: **DENY**
- **SERVICE_SLOW_NORMAL (Latência 500ms):**
  - Alpha Histórico: ~84.5 (em transição)
  - CkScore: 96.48
  - Divergência: ~11.9
  - Fricção: 100.0 (Cap Máximo)
  - Estado: **DENY**

## Conclusão e Veredito

**Hipótese: REJEITADA. Teoria PER: FORTALECIDA.**

Ao contrário da expectativa teórica de que o `SERVICE_SLOW_NORMAL` absorveria os 500ms como um "WARN leve", **ambos os serviços colapsaram para o estado DENY instantaneamente**.

### ACHADO 01 — Dominância Absoluta do Dado Cru

**Definição:**
Um sistema aparenta ser relacional porque considera histórico e divergência, mas continua sendo dominado por um valor bruto absoluto quando esse valor é somado diretamente ao estado final. A fórmula atual do Ouroboros possui intenção relacional, mas o `CkScore` bruto domina a fricção final. Como `Fricção = min(100, CkScore + Divergência^1.5)`, qualquer `CkScore` acima do threshold de `DENY` (70) torna a memória histórica completamente irrelevante.

A prova irrefutável disso é que o `SERVICE_SLOW_NORMAL` já estava em estado `DENY` crônico mesmo durante o seu warmup normal (latência de 400ms = CkScore de 80.0 > Limiar de 70).

### Impacto para a Teoria da Computação Relacional
O experimento rejeitou a hipótese operacional inicial (a fórmula atual falhou), mas confirmou a necessidade premente da Teoria PER. O Gate 01 cumpriu sua função científica: fotografou a limitação sem alterar o core Go. O Ouroboros provou que adicionar memória a uma fórmula não basta. Se o dado cru continuar dominando o estado final, o sistema ainda não é verdadeiramente relacional.

**Próxima Etapa Autorizada:** Gate 02 — Relational Formula Sandbox (Read-Only). Vamos testar fórmulas alternativas offline usando os snapshots coletados.
