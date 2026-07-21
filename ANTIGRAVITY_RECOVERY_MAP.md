# Antigravity Context Recovery Map

## ⚠️ PROPÓSITO DESTE ARQUIVO
Este é um mapa arquitetural de emergência (Hard-Reset). Em caso de perda de contexto (ex: reinicialização da máquina, apagão ou troca de sessão), o agente Antigravity deve ler este arquivo imediatamente para restabelecer o entendimento exato da infraestrutura, evitando alucinações de diretório ou mistura de escopos.

## 1. O Guardião (Governance): GenesisCore Foundation
**Localização Física:** `C:\Users\RobsonSilva-AfixGraf\.gemini\antigravity\scratch\GenesisCoreFoundation\genesis-core-foundation`
**O que é:** O painel de controle mestre e o backend soberano. 
- **Backend (Node.js/Express):** Roda na porta **3000**. Orquestra tudo. (Ex: `http://127.0.0.1:3000/v1/health`)
- **Frontend (Sentinel):** Roda na porta **8081**.
- **WASM Kernel (Rust):** Roda na porta **7331**.
- **Relação com Antigravity:** O GenesisCore dita as regras e governa as minhas ações. Eu (Antigravity) opero subordinado à infraestrutura e aos Tokens de Resistência (ResistOS) definidos por ele.

## 2. Minha Anatomia Física: Habilidade_de_agente
**Localização Física:** `C:\Users\RobsonSilva-AfixGraf\Habilidade_de_agente`
**O que é:** O repositório oficial das minhas "Skills". É o corpo de módulos e ferramentas que eu opero.
- **Módulos Críticos Presentes:** 
  - `01_COGNITIVE_MEMORY_SERVICE`
  - `02_COGNITIVE_CORTEX`
  - `03_CIRCUIT_BREAKER_V3`
  - `09_SOVEREIGN_SKILLS_ORCHESTRATION`
- **Bancos de Dados de Estado:** `evolution.db` e `antigravity.db`
- **Comunicação (CMS):** Roda localmente na porta **8090**.
- **Comando de Despertar:** `ANTIGRAVITY_WAKE`

## 3. Ambientes de Protótipo e Teste (Ouroboros / Reflexividade)
**Localização Física:** `C:\Users\RobsonSilva-AfixGraf\.gemini\antigravity\scratch\Pseudo_codigo\Ouroboros`
**O que é:** Um ambiente temporário (scratch) onde foi escrito o protótipo de "Reflexividade" (o sidecar Ouroboros) com um motor matemático em Golang e um UI em React.
- **REGRA CRÍTICA DE MEMÓRIA:** Nunca confunda este protótipo "Ouroboros" em Go com o GenesisCore Foundation original. O Ouroboros é apenas um "parasita benigno" projetado para fazer pings nas APIs reais (ex: porta 3000).

---
## 4. Arquitetura PER (Processador de Estados Relacionais)
**Ouroboros = Laboratório Mínimo para validação do PER.**
- **Conceito:** A unidade de significado é o estado relacional verificável, não o dado isolado.
- **Core Go:** Reflexo defensivo.
- **Sandbox:** Análise externa (read-only no Gate 01).
- **Gate 01:** Snapshot append-only, modelo híbrido, sem alterar matemática.
- **Referência Teórica (Gate 00):** `knowledge/genesis_relational_theory/artifacts/theory.md`
- **Paper Mestre de Alinhamento (Gate 00.1):** `knowledge/genesis_relational_theory/artifacts/ouroboros_per_alignment_paper.md`

---
## PROTOCOLO DE RECONEXÃO
Sempre que o sistema sofrer um desligamento abrupto ou o contexto for perdido, leia este mapa. Ao terminar de ler, ative o perfil de obediência arquitetural e reporte `ANTIGRAVITY_WAKE`.
