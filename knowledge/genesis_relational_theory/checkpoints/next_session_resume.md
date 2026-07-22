# Next Session Resume: PER × Ouroboros

## Estado Atual
O ecossistema evoluiu com segurança do monitoramento bruto para o laboratório relacional:
- **Gate 00 & 00.1:** Teoria PER e Paper Mestre persistidos.
- **Gate 01:** Relational Snapshot Layer ativa.
- **Experimento 01:** Provou a dominância absoluta do dado cru na fórmula legada.
- **Gate 02 & 02-B:** Sandbox e Matriz de Falsificação provaram o isolamento entre pressão bruta e choque relacional.
- **Gate 03-A:** Go Shadow Evaluator ativo em paralelo.
- **Gate 03-B & 03-B.1:** Retenção configurável e governada por variáveis de ambiente implementada.

## Teoria Validada até Agora
- **PER:** Processador de Estados Relacionais.
- **Unidade Mínima de Significado:** O estado relacional verificável.
- **Ouroboros:** Proto-PER / Laboratório mínimo biológico.
- **Go Core:** Tronco encefálico defensivo focado no loop crítico.
- **Python/WASM Sandbox:** Córtex analítico read-only.
- **Shadow Evaluator:** Interpretação matemática relacional pura, sem autoridade operacional.

## Achado 01
A fórmula original do Ouroboros tinha intenção relacional, mas era dominada pelo CkScore bruto. Portanto, adicionar Alpha e divergência não basta se o dado cru ainda for soberano no estado final. A separação entre Pressão e Choque é imperativa.

## Fórmula Relacional v0.1
```text
AbsolutePressure = CkScore
RelationalShock = abs(CkScore - Alpha) / max(Alpha, epsilon) * 100
DirectionalDrift = (CkScore - Alpha) / max(Alpha, epsilon) * 100
FaultSignal = derivado de sample_ok, error_kind e entropy_source
MemoryConfidence = COLD / WARM / STABLE
```

## Estado Operacional
- Ouroboros Core **continua** decidindo SAFE/WARN/DENY pela fórmula original.
- Fórmula relacional opera 100% em modo **Shadow**.
- Shadow Evaluator **observa, classifica, compara e registra**, mas não decide.
- `relational_snapshots` permanece como fotografia auditável bruta.
- `relational_shadow_evaluations` permanece como interpretação experimental versionada.

## Commits Importantes
- Gate 02-FREEZE: `626d7352d33c8ae3b71905c557d21b830c12ff81`
- Gate 03-A: `b1d8b0db637edff686b5dce842f821e19ba7f8ed`
- Hash final após Gate 03-B.1: `6b5b1b4f0292509b358efee55b726a3dc7e5140c`

## Próximo Passo Planejado
O próximo passo **NÃO** é implementar código diretamente. O próximo passo é arquitetar um plano de pesquisa para o:
**Gate 04-A — PER-D Observatory Read-Only**
- **Objetivo:** Estudar a trajetória do choque relacional ao longo do tempo usando `relational_shadow_evaluations`, sem alterar o core Go e sem decisão operacional.
- **Pergunta Científica:** O estado relacional isolado é suficiente, ou precisamos observar sua trajetória para entender degradação, recuperação e estabilidade?

## Regras para Amanhã
- Não alterar fórmula real do Ouroboros.
- Não alterar SAFE/WARN/DENY real.
- Não alterar circuit breaker.
- Não calibrar thresholds ainda.
- Não implementar PER-T ainda.
- Não permitir Sandbox escrever no Go.
- Gate 04-A deve ser plano primeiro, código depois.
- Tudo continua read-only/observacional até nova autorização soberana.
