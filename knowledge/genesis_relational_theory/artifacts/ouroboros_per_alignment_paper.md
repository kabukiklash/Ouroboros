# Paper Mestre de Alinhamento: PER × Ouroboros

## 2. Propósito
Explicar que o Ouroboros será usado como plataforma simples, local-first e mensurável para validar experimentalmente a Teoria da Computação Relacional.

## 3. Hipótese central
A computação tradicional trata dados isolados como unidade de significado. A Computação Relacional propõe que a menor unidade de significado é o estado relacional verificável.

## 4. Definição do PER
PER = motor que transforma sinais/dados isolados em estados relacionais verificáveis, contendo valor, origem, contexto, histórico, regra, dependências, impacto, validade e explicação.

## 5. Ouroboros como Proto-PER
Registrar que o Ouroboros já processa sintomas, não apenas dados: latência + entropia + densidade + CkScore + Alpha + divergência + fricção + estado + ação defensiva.

## 6. O que o Ouroboros já valida parcialmente
- Alpha como semente do PER-M, não PER-M completo.
- Divergência como semente do PER-D, não PER-D completo.
- SAFE/WARN/DENY como estado relacional de saúde.
- Circuit Breaker/Sandbox como ação defensiva derivada do estado.

## 7. Lacunas atuais
- Ausência de snapshot completo.
- Ausência de topologia relacional.
- Ausência de ressonância estrutural.
- Ausência de emergência relacional.
- Ausência de gravidade relacional.
- Ausência de memória relacional completa.

## 8. Separação arquitetural obrigatória
**Go/Ouroboros Core = tronco encefálico:**
- rápido
- local-first
- matemático
- defensivo
- sem IA pesada
- sem análise complexa no loop crítico

**Python/WASM Sandbox = córtex analítico:**
- read-only no Gate 01
- leitura em batch dos snapshots
- geração de hipóteses
- análise futura de dinâmica, topologia, padrões e emergência
- sem alterar constantes do core Go no Gate 01

## 9. Gate 01 — Relational Snapshot Layer
Objetivo: fotografar o estado relacional completo sem alterar a matemática central.
Regra: O Gate 01 só observa e registra. Não interpreta, não prevê, não adapta e não decide além do que o Ouroboros já decide hoje.

## 10. Modelo mínimo do snapshot
Registrar campos obrigatórios:
- id
- service_id
- observed_at_ms
- inserted_at
- latency_ms
- sample_ok
- error_kind
- entropy_source
- entropy
- density
- ck_score
- alpha_at_time
- divergence
- friction
- state
- action_taken
- reason_json
- relations_json
- impact_json
- schema_version
- previous_hash
- current_hash

## 11. Política de hash
current_hash deve selar o snapshot completo.
previous_hash deve ser calculado por service_id.
O hash não pode usar apenas timestamp + service_id + ck_score.

## 12. Política de retenção
A tabela é append-only no fluxo normal de escrita, mas possui retenção governada para sobrevivência local-first.
Configuração recomendada:
```env
OUROBOROS_RELATIONAL_SNAPSHOTS_ENABLED=true
OUROBOROS_RELATIONAL_MAX_ROWS=500000
OUROBOROS_RELATIONAL_RETENTION_HOURS=24
OUROBOROS_RELATIONAL_PURGE_INTERVAL_SECONDS=300
```
Regra: expurgar por tempo ou por limite de linhas, o que chegar primeiro.

## 13. Registro de expurgo
Criar `relational_retention_events` ou log equivalente contendo:
- timestamp
- rows_deleted
- policy_json
- last_deleted_id
- last_deleted_hash
Objetivo: garantir que remoções sejam governadas e não mutações silenciosas.

## 14. Regras de segurança do Gate 01
- Não alterar fórmulas centrais.
- Não refatorar main.go além do necessário.
- Não implementar PER-T, PER-R, PER-E, PER-G ou PER-M completo agora.
- Não permitir que Sandbox altere o Go.
- Escrita SQLite deve ser best-effort.
- Falha ao gravar snapshot não pode afetar SAFE/WARN/DENY.
- Endpoint, se existir, deve ser read-only, paginado e limitado.
- Usar WAL, synchronous=NORMAL e busy_timeout.

## 15. Critério de sucesso
Para cada estado SAFE/WARN/DENY, o Ouroboros deve conseguir registrar:
- o que foi observado;
- qual era o normal esperado;
- qual foi a divergência;
- qual regra matemática produziu a fricção;
- qual estado defensivo foi ativado;
- qual ação foi tomada;
- qual impacto foi previsto;
- quais relações estavam envolvidas.

## 16. Frase canônica
“O Gate 01 não implementa a família PER. Ele cria o solo observável, auditável e mensurável onde a família PER poderá ser estudada posteriormente.”

## 17. Status
Este documento deve ser salvo como referência obrigatória antes da implementação do Gate 01. Atualizar também o ANTIGRAVITY_RECOVERY_MAP.md com uma âncora curta apontando para este paper.
