# Roadmap: Implementação do Mistral AI como Shadow Evaluator (Gate 03)

**Status:** Rascunho de Visão (Pendente de execução após a estabilização matemática do PER).
**Conceito:** Utilizar modelos Open-Weights da Mistral (Local-First) como motores de inferência autônomos dentro do ambiente Ouroboros, operando de forma assíncrona para não onerar o fluxo em tempo real do sistema.
**Objetivo:** Garantir Soberania de Dados (Zero Cloud APIs) e permitir auditoria semântica profunda sobre anomalias no sistema AfixControl (NFe, Integrações).

---

## Fase 1: Setup da Infraestrutura Base Local (A fundação)
**Objetivo:** Levantar a API LLM no ambiente host (Dell G3 - 40GB RAM) sem tocar no código Go ainda.
- [ ] **1.1. Instalação do Motor de Inferência:** Instalar localmente o **Ollama** (ou vLLM) no Windows. O Ollama atua como um servidor HTTP leve para gerenciar os pesos dos LLMs.
- [ ] **1.2. Download do Modelo Ágil (Testes Iniciais):** Executar `ollama pull mistral` (Mistral 7B) para baixar o modelo base quantizado (~4.1GB).
- [ ] **1.3. Teste de Stress Simples:** Fazer um cURL manual para `http://localhost:11434/api/generate` com um JSON contendo uma payload falsa de erro do ERP Gestio para validar o tempo de resposta da CPU (i7-9750H) + GPU.

## Fase 2: Integração Assíncrona no Ouroboros (O Parasita se conecta ao Cérebro)
**Objetivo:** Conectar o backend Go (Sentinela) ao servidor Ollama de forma segura e não bloqueante.
- [ ] **2.1. Criação do Módulo Go (shadow_evaluator.go):** Criar um novo package interno no Ouroboros responsável apenas por orquestrar chamadas HTTP POST para o `localhost:11434`.
- [ ] **2.2. Fila Assíncrona (Go Routines/Channels):** Como a inferência local é mais lenta (pode levar 10-15s no Mixtral), a chamada para a IA *nunca* pode travar a thread principal. Implementar uma Worker Pool simples que lê anomalias de um canal (`chan rel.AnomalyPayload`) e processa sequencialmente.
- [ ] **2.3. Circuit Breaker Local:** Implementar timeout de 30s-60s. Se a IA local travar, o Ouroboros apenas ignora e continua rodando o Gate 01 e Gate 02.

## Fase 3: Engenharia Semântica para o PER (O "System Prompt")
**Objetivo:** Ensinar a IA a "falar" a linguagem da Teoria Computacional Relacional (PER).
- [ ] **3.1. Design do System Prompt:** Criar um prompt rígido que assuma o papel de **Auditor Sistêmico Relacional**.
    * *Regra de Injeção:* "Você é um classificador relacional. Sua função é receber um JSON de erro (ex: Gestio/Correios), avaliar sob a ótica do sistema AfixControl, e devolver ESTRITAMENTE um objeto JSON contendo: `causa_raiz`, `impacto_logistico`, e `sugestao_correcao`."
- [ ] **3.2. Formato JSON Restrito:** Configurar a API do Ollama (`format: "json"`) para forçar o Mistral a sempre responder em JSON estruturado, impedindo respostas "tagarelas" que quebram o parser do Go.

## Fase 4: Persistência Observacional (O Log Inteligente)
**Objetivo:** Salvar o insight gerado pela IA no repositório de histórico (SQLite), consolidando o Gate 03.
- [ ] **4.1. Atualização de Schema SQLite:** Confirmar que a tabela `relational_shadow_evaluations` no Ouroboros está apta a receber as colunas: `id_evento`, `payload_enviado`, `analise_ia_json`, `tempo_inferencia_ms`, `modelo_utilizado`.
- [ ] **4.2. Commit no Banco Local:** Após o Classificador Relacional em Go receber a resposta JSON da IA local, persistir o resultado no SQLite e disparar um evento (WebSocket) para que o Dashboard React mostre o insight em tempo real (ex: "Robô identificou anomalia NFe e sugeriu a correção").

## Fase 5: O Teste com o Gigante (Mixtral 8x7B)
**Objetivo:** Maximizar o uso dos 40GB de RAM do Dell G3 para aumentar brutalmente a capacidade analítica.
- [ ] **5.1. Download do Mixtral MoE:** Executar `ollama pull mixtral` (Baixará ~26GB correspondentes ao modelo Mixtral 8x7B Quantizado).
- [ ] **5.2. Benchmarking de RAM/CPU:** Observar como o i7 lida com a ativação de 13B parâmetros (MoE). O tempo de resposta saltará de 1-2 segundos (Mistral 7B) para possivelmente 15-20 segundos. 
- [ ] **5.3. Validação de Caso Real:** Lançar 5 anomalias pesadas (ex: 5 logs de erro de integração) para a fila do Ouroboros e validar a capacidade do modelo MoE de encontrar correlações invisíveis sem comprometer a estabilidade do sistema host.

---
*Documento forjado sob os protocolos Antigravity. Soberania arquitetural preservada.*
