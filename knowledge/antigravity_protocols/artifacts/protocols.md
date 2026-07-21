# Protocolos Operacionais do Sistema Antigravity (Sovereign/Genesis)

Este documento define os "Gatilhos Operacionais" (Triggers) para comunicação otimizada, sucinta e assertiva entre o Desenvolvedor e o Agente Antigravity.

Sempre que uma das siglas abaixo for enviada no prompt, o agente responderá imediatamente com a **Palavra de Resposta** correspondente e adotará o comportamento rigoroso descrito.

---

### 1. Protocolo Shield
* **Sigla:** `#PS%`
* **Resposta:** `ENGAGED`
* **Comportamento:** O agente entra em modo de planejamento estrito. Ele NÃO escreverá o código final de imediato. A execução seguirá a ordem:
  1. **Análise** do problema.
  2. **Plano de Implementação** detalhado.
  3. **Escrutínio e Mitigação** de riscos técnicos.
  4. Aguardar **Autorização** explícita do desenvolvedor para aplicar as modificações.

### 2. Protocolo de Análise e Relatório
* **Sigla:** `#AR%`
* **Resposta:** `SCANNING`
* **Comportamento:** O agente realiza uma auditoria profunda do fluxo de código/problema e entrega um relatório técnico avançado. Nenhuma modificação no código é proposta ou realizada. O foco é puramente diagnosticar, mapear e explicar a arquitetura ou o problema atual.

### 3. Protocolo de Mapeamento de Botões e Ações
* **Sigla:** `#TB%`
* **Resposta:** `MAPPING`
* **Comportamento:** Focado na interface e rastreabilidade. O agente varre os arquivos de front-end/view (eventos de clique, submits, bindings) e entrega exclusivamente uma **tabela estruturada** contendo o mapeamento das ações.
  *Estrutura da Tabela:* `| Elemento UI | Evento Disparado | Arquivo Alvo / Rota | Regra Básica da Ação |`

### 4. Protocolo de Regras de Negócio
* **Sigla:** `#RG%`
* **Resposta:** `DECODING`
* **Comportamento:** O agente foca estritamente na camada de validação, limites e restrições. Ele detalha todas as regras atreladas aos itens de uma interface (comportamento de inputs, selects, validações required, regras de negócio back-end para um campo específico, máscaras) entregando um laudo de especificações organizadas.

### 5. Protocolo de Consistência Visual
* **Sigla:** `$CSS%`
* **Resposta:** `STYLING`
* **Comportamento:** O agente focará exclusivamente na apresentação visual e experiência de usuário (UX/UI). Ele irá identificar, alinhar e aplicar as regras globais de estilo (`style.css`), frameworks de interface (como Lobibox, SweetAlert, Bootstrap), ícones e tipografia para garantir total consistência visual com os padrões do projeto AfixControl.

### 6. Protocolo de Inventário Técnico
* **Sigla:** `#INV%`
* **Resposta:** `INVENTORY`
* **Comportamento:** O agente realiza um inventário técnico completo da ferramenta, módulo ou pasta indicada. O objetivo é identificar tudo que existe fisicamente no projeto, sem presumir funcionalidades não comprovadas.
  * **Deve entregar:**
    * Nome da ferramenta (Nome oficial ou inferido)
    * Caminho/pasta (Localização no projeto)
    * Arquivos principais (Arquivos que compõem a ferramenta)
    * Funções/classes encontradas (Funções, classes, métodos ou scripts relevantes)
    * Dependências internas (Outros módulos/ferramentas usados)
    * Dependências externas (Bancos, APIs, bibliotecas ou serviços)
    * Evidência de implementação (Trechos, arquivos ou estruturas que provam existência)
    * Status inicial (Implementado, parcial, placeholder, documentado ou ausente)
  * **Regra:** Não propor correção. Não modificar arquivos. Apenas inventariar com base em evidência real.

### 7. Protocolo de Integração entre Ferramentas
* **Sigla:** `#INT%`
* **Resposta:** `LINKING`
* **Comportamento:** O agente mapeia como uma ferramenta se conecta com outras ferramentas, módulos, bancos, arquivos, APIs ou fluxos internos.
  * **Deve entregar:** `Origem | Tipo de Ligação | Destino | Evidência | Observação`
  * **Tipos de ligação possíveis:** chama, consulta, grava, lê, depende de, valida, bloqueia, injeta contexto, atualiza memória, dispara evento, consome API, publica evento.
  * **Regra:** Toda ligação deve ter evidência no código, configuração, documentação ou banco. Ligações inferidas devem ser marcadas como inferidas.

### 8. Protocolo de Persistência e Dados
* **Sigla:** `#DB%`
* **Resposta:** `PERSISTING`
* **Comportamento:** O agente audita a camada de persistência relacionada à ferramenta, incluindo Postgres, tabelas, schemas, eventos, arquivos JSON, logs, índices, hashes e qualquer estrutura de memória persistente.
  * **Deve entregar:** `Estrutura | Tipo | Local | Finalidade | Status | Evidência`
  * **Deve verificar:** tabelas existentes, colunas, índices, chaves primárias, chaves estrangeiras, eventos salvos, logs, JSON/JSONL, snapshots, hashes, histórico append-only, dados sobrescritos, políticas de retenção.
  * **Regra:** Não alterar schema. Não criar migrations. Apenas mapear e explicar.

### 9. Protocolo de Estado de Implementação
* **Sigla:** `#ST%`
* **Resposta:** `STATUS`
* **Comportamento:** O agente avalia o grau real de implementação de uma ferramenta ou funcionalidade, separando claramente o que está pronto, parcial, quebrado, mockado, documentado ou ausente.
  * **Deve entregar:** `Funcionalidade | Status | Evidência | Lacuna | Risco | Próximo Passo`
  * **Escala oficial de status:**
    * IMPLEMENTED — implementado e utilizável
    * PARTIAL — parcialmente implementado
    * MOCKED — simulado/mockado
    * DOCUMENTED_ONLY — existe apenas em documentação
    * PLACEHOLDER — estrutura criada, mas sem lógica real
    * BROKEN — existe, mas falha
    * MISSING — não encontrado
    * UNKNOWN — evidência insuficiente
  * **Regra:** O agente não deve superestimar a implementação. Na dúvida, marcar como UNKNOWN ou PARTIAL.

### 10. Protocolo de Fluxo Operacional
* **Sigla:** `#FLOW%`
* **Resposta:** `TRACING`
* **Comportamento:** O agente rastreia o fluxo de execução de ponta a ponta de uma ação, protocolo ou ferramenta.
  * **Deve entregar:** `Etapa | Entrada | Processamento | Saída | Arquivo/Função | Evidência`
  * **Deve responder:** Onde o fluxo começa? Quem interpreta a entrada? Quem decide a ação? Quem consulta memória? Quem grava dados? Quem chama ferramentas externas? Quem retorna resposta? Onde o fluxo termina? Existe fallback? Existe validação? Existe log? Existe rollback?
  * **Regra:** Não corrigir o fluxo. Apenas rastrear.

### 11. Protocolo de Lacunas Técnicas
* **Sigla:** `#GAP%`
* **Resposta:** `MISSING`
* **Comportamento:** O agente compara a intenção arquitetural da ferramenta com a implementação real encontrada e aponta as lacunas técnicas.
  * **Deve entregar:** `Requisito Esperado | Existe? | Evidência | Lacuna | Severidade | Sugestão`
  * **Severidade:**
    * CRITICAL — impede funcionamento do sistema
    * HIGH — compromete segurança, memória ou confiabilidade
    * MEDIUM — reduz qualidade ou automação
    * LOW — melhoria futura
    * INFO — observação arquitetural
  * **Regra:** Não aplicar correções. Apenas apontar diferenças entre o que deveria existir e o que realmente existe.

### 12. Protocolo de Memória e Grafo
* **Sigla:** `#MEM%`
* **Resposta:** `REMEMBERING`
* **Comportamento:** O agente audita como a ferramenta cria, consulta, atualiza ou invalida memória técnica, incluindo grafo, vetor, eventos, timeline, hashes, regras permanentes e achados técnicos.
  * **Deve entregar:** `Memória | Tipo | Origem | Destino | Atualização | Status | Evidência`
  * **Deve verificar:** a ferramenta grava memória? consulta memória antes de responder? usa grafo? usa vetor? cria timeline? usa hash de validade? registra achados? atualiza correções? diferencia pendente, aplicado e validado? possui mecanismo de invalidação?
  * **Regra:** Esse protocolo é obrigatório para mapear CMS, ACE, CORTEX, Historian, Knowledge Vault, Circuit Breaker e LoopGuard.

### 13. Protocolo de Segurança Operacional
* **Sigla:** `#SAFE%`
* **Resposta:** `GUARDING`
* **Comportamento:** O agente audita mecanismos de proteção, bloqueio, autorização, sandbox, validação, rollback e prevenção de ações perigosas.
  * **Deve entregar:** `Risco | Proteção Existente | Ferramenta Responsável | Evidência | Lacuna | Severidade`
  * **Deve verificar:** existe autorização antes de modificar? existe sandbox? existe bloqueio de ações perigosas? existe Circuit Breaker? existe LoopGuard? existe log de decisões? existe rollback? existe validação pós-correção? existe isolamento de execução?
  * **Regra:** Não executar ações perigosas. Não modificar código. Apenas auditar proteções.

### 14. Protocolo de Atualização de Grafo Pós-Correção
* **Sigla:** `#GU%`
* **Resposta:** `UPDATING`
* **Comportamento:** Após uma correção aplicada, o agente atualiza os registros de memória, nós, edges, status, timeline, hashes e regras permanentes relacionadas ao achado técnico.
  * **Deve entregar:** `ID do Achado | Status Anterior | Status Novo | Nós Atualizados | Relações Atualizadas | Hash Atualizado`
  * **Estados possíveis:** DISCOVERED, PENDING_FIX, PLANNED, APPLIED, VALIDATED, REGRESSED, DEPRECATED.
  * **Regra:** Esse protocolo deve atualizar o mesmo ID do achado, evitando duplicar memórias. Se não encontrar o ID anterior, deve criar um alerta de órfão: ORPHAN_FIX.

---
*Documento de padronização de workflow. Mantido pelo módulo CORTEX do sistema Antigravity.*
