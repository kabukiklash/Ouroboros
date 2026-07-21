# Auditoria Técnica AfixControl — Parte 2: Cadastro, Edição, Proposta, Pedido, Fila, Classificação

## 3. Auditoria do Cadastro de Cliente

**Tela origem:** `app/_clientes/novo-cliente.php`
**Arquivo que recebe:** `app/ajax/clientes_otk.php`, action=`salvar_cliente` (linhas 140-374)

### Campos enviados pela tela (POST)
`natureza`, `cnpj_cpf`, `razao_social`, `nome_fantasia`, `inscricao_estadual`, `inscricao_municipal`, `cep_entrega`, `logradouro_entrega`, `numero_entrega`, `complemento_entrega`, `bairro_entrega`, `cidade_entrega`, `uf_entrega`, `codigo_ibge_entrega`, `cep_faturamento`, `logradouro_faturamento`, `numero_faturamento`, `complemento_faturamento`, `bairro_faturamento`, `cidade_faturamento`, `uf_faturamento`, `codigo_ibge_faturamento`, `contatoFinanceiro`, `emailContatoFinanceiro`, `telefoneFinanceiro`, `emailsFinanceiro[]`, `contato_nome[]`, `contato_email[]`, `contato_telefone[]`, `contato_celular[]`, `contato_departamento[]`, `contato_observacao[]`, `cliente_temporario`

### Campos salvos localmente (tabela `clientes`)
| Coluna | Conteúdo | Evidência |
|--------|----------|-----------|
| `cliente_codigo` | Gerado sequencialmente (>=195000) | L251-277 |
| `cliente_info` | JSON: `{cnpj_cpf, razaoSocial, nomeComercial, natureza, inscricaoEstadual, inscricaoMunicipal}` | L341-343 |
| `cliente_endereco_entrega` | JSON: `{entrega: {cep, logradouro, numero, complemento, bairro, cidade, uf, codigoIbge}}` | L344 |
| `cliente_endereco_faturamento` | JSON: `{endereco: {cep, logradouro, numero, ...}}` | L345 |
| `cliente_contatos` | JSON: `{contatos: [{nome, email, telefone, celular, departamento, observacao}]}` | L346 |
| `cliente_representante` | `$meu_codigo_vendedor` (int) | L347 |
| `cliente_rdid` | Placeholder: `queue_rd_comp_xxxxx` | L332, 348 |
| `cliente_temporario` | 0 ou 1 (do POST) | L349 |

**Campos NÃO salvos localmente:** `contatoFinanceiro`, `emailContatoFinanceiro`, `telefoneFinanceiro`, `emailsFinanceiro[]`, `categoria_de_cliente`. Esses campos são enviados para a Gestio via `dados_otkweb` (L283-305) mas **não têm coluna** na tabela `clientes`.

### Campos enviados à Gestio (via fila)
Payload do `dados_otkweb` (L283-305): `codigoDoCliente`, `razaoSocial`, `nomeComercial`, `natureza`, `inscricaoEstadual`, `inscricaoMunicipal`, `endereco` (entrega), `telefone`, `email`, `receberEmailBoleto`, `codigoDoVendedor`, `cnpj`/`cpf`, `contatoFinanceiro`, `emailContatoFinanceiro`, `telefoneFinanceiro`.

### Campos que entram na fila
Tipo: `criar_cliente_otk_rd_v1`. Payload (L321-327): `dados_otkweb`, `contatos_otk`, `razao_social`, `cnpj_cpf`, `codigo_vendedor`.

### Tratamento de erro
- Se falhar gerar código: retorna JSON com `success: false` (L279)
- Se falhar INSERT/UPDATE local: retorna `success: false` (L370)
- A fila não tem validação de retorno — `FilaSincronizacao::Inserir` pode falhar silenciosamente (retorna false, logado em error_log — L37 fila_sincronizacao.php)

### Risco de perda de dados
- **RISCO ALTO:** `contatoFinanceiro`, `emailContatoFinanceiro`, `telefoneFinanceiro` e `emailsFinanceiro[]` **não são persistidos localmente**. Se o Worker falhar e a fila atingir 3 retries, esses dados são PERDIDOS. Eles só existem no payload da fila.
- **RISCO MÉDIO:** O `cliente_rdid` recebe um placeholder `queue_rd_comp_*`. Se o Worker falhar no RD Station, o placeholder nunca é substituído pelo ID real.

---

## 4. Auditoria da Edição de Cliente

**Tela origem:** `app/_clientes/alterar-cliente.php`
**Arquivo que recebe:** `app/ajax/clientes_otk.php`, action=`alterar_cliente` (linhas 518-725)

### Campos enviados pela tela (POST)
Mesmos do cadastro, acrescido de `cliente` (código existente) e `rdid`.

### Campos que NÃO são enviados pela tela
- `codigoDoVendedor` — **NÃO é enviado pela tela de edição**.

### Preservação de valores ausentes

| Campo | Campo ausente preserva valor antigo? | Evidência |
|-------|--------------------------------------|-----------|
| `cliente_info` | **NÃO** — é sobrescrito inteiramente com os dados do POST | L545-552 → L685 |
| `cliente_endereco_entrega` | **NÃO** — sobrescrito | L555-566 → L686 |
| `cliente_endereco_faturamento` | **NÃO** — sobrescrito | L568-579 → L687 |
| `cliente_contatos` | **NÃO** — sobrescrito | L604-613 → L688 |
| `cliente_representante` | **SIM** — busca o valor original do banco (L702-705) antes de salvar | L708 |
| `cliente_rdid` | **PARCIAL** — usa `$_POST['rdid'] ?? $rd_id` | L698, L708 |
| `cliente_temporario` | **NÃO** — usa valor do POST (L537-543). Se o checkbox não for enviado, vira 0 | L537-543 |
| `contatoFinanceiro` | Não salvo localmente — só vai para a fila | — |

### Vendedor (representante)
**Comportamento correto**: L701-705 faz `SELECT cliente_representante FROM clientes WHERE cliente_codigo = :codigo` e preserva o valor original. Se o cliente não existir no banco (caso raro), usa `$meu_codigo_vendedor` como fallback.

### cliente_temporario
**RISCO:** Se o vendedor editar um cliente e o checkbox "Temporário" não estiver marcado, o campo é sobrescrito para 0 (L537-543). Isso é intencional (o vendedor controla), mas se a tela não enviar o campo (bug de JS), o cliente perde a flag silenciosamente.

### Contatos
**Bug corrigido (L707-708):** Comentário no código confirma que havia um bug onde `$contatos_otk` (formato OTK) sobrescrevia `$contatos` (formato local). Agora salva `$contatos` corretamente.

### Dados fiscais (contatoFinanceiro)
**NÃO persistidos localmente** — mesma situação do cadastro. Somente na fila (payload `dados_otkweb`).

---

## 5. Auditoria da Geração de Proposta

### Arquivos que geram proposta/orçamento
1. `ajax/propostas.php` — action diversas (salvar, aprovar, etc.)
2. `classes/propostas.php` — `Proposta::Insert()` (L44-86)
3. `_propostas/nova-proposta.php` — front-end

### Fluxo de criação — `Proposta::Insert()` (propostas.php L44-86)
```
$_otk = new OtkWeb($this->pdo);
$_cliente = $_otk->DetalhesCliente($cliente);  // ← CHAMADA SÍNCRONA DIRETA À API
```
**Finalidade:** Buscar `razaoSocial` para gravar em `proposta_nome_cliente`.

### Fluxo de busca de cliente na proposta — `procurar_cliente` (clientes_otk.php L727-804)
```
$cliente = $_otkweb->DetalhesCliente($codigo);   // ← SÍNCRONO
$contasReceber = $_otkweb->GetContasReceber();    // ← SÍNCRONO (lista TODOS os inadimplentes)
$pagante = $_otkweb->GetEstadoPagante($codigo);   // ← SÍNCRONO
```
**São 3 chamadas síncronas à API** antes de retornar ao JS. Se qualquer uma falhar, o vendedor não consegue selecionar o cliente.

**Mitigação parcial (L763-795):** Os contatos locais são priorizados e injetados em `$cliente['data']['contatos']`. Mas o corpo principal da resposta (`razaoSocial`, `endereco`, etc.) **ainda depende 100% da API**.

### Proposta pode ser criada com Gestio fora do ar?
**NÃO.** Duas dependências bloqueantes:
1. `Proposta::Insert()` chama `DetalhesCliente()` para preencher `proposta_nome_cliente`
2. `procurar_cliente` precisa da API para retornar dados ao JS

### Dados que precisariam estar no banco local
`razaoSocial`, `nomeComercial`, `cnpj/cpf`, `natureza`, `codigoDoVendedor`, `endereco` — **todos já existem** em `clientes.cliente_info` e `clientes.cliente_endereco_faturamento`.

---

## 6. Auditoria da Geração e Processamento do Pedido

### 6.1 Pedido de Produção (via aprovação de proposta)
**Arquivo:** `ajax/propostas.php`, action=`alterar-estado` caso `aprovado` (L436-469)

```php
$cliente = $_otkweb->DetalhesCliente($proposta_cliente);  // L440 ← SÍNCRONO
// Extrai: codigoDoCliente, razaoSocial, nomeComercial, natureza, cnpj, cpf, codigoDoVendedor
$insert_pedido = $_pedido->InsertPedido(..., $info);      // L456
```

**Consulta Gestio?** SIM — bloqueante.
**Grava snapshot?** SIM — grava `$info` como JSON no pedido de produção.
**Pode travar com Gestio fora?** SIM — se `DetalhesCliente` falhar, o pedido **não é gerado**, mesmo a proposta já estando aprovada.
**Mesmo fluxo em `confirmar-pagamento`** (L509-512): mesmo problema.

### 6.2 Pedido de Venda (PV local) — `pedidos-venda.php`
**Arquivo:** `ajax/pedidos-venda.php`, function `criarPedidoVenda()` (L124-333)

```php
$cliente = $_otkweb->DetalhesCliente($proposta['proposta_cliente']);  // L162 ← SÍNCRONO
```

**Usa banco local?** PARCIALMENTE — L179-212 busca endereços de entrega/faturamento e contatos do banco local (`clientes`). Mas os dados principais do cliente (`razaoSocial`, `endereco`, `cnpj`, etc.) vêm de `$cliente['data']` (API).

**Grava snapshot?** SIM — L229-292 monta um array `$dados` com todos os dados e grava como JSON em `pedidos_venda.pedido_venda_dados`.

**Divergência proposta vs pedido?** POSSÍVEL — os dados gravados no snapshot vêm da API no momento da criação do PV. Se o cliente foi editado localmente mas o Worker não sincronizou, os dados do PV terão os dados antigos da Gestio.

### 6.3 Pedido OTK (legado) — `pv.php`
**Arquivo:** `ajax/pv.php`, action=`abrir-pedido` (L33-176)

```php
$cliente = $_otkweb->DetalhesCliente(intval($_POST['cliente']));  // L45 ← SÍNCRONO
// ... monta payload com dados do cliente ...
$resultado = $_otkweb->PostPedidoDeVenda($data, $_POST['proposta']);  // L173 ← SÍNCRONO
```

**100% dependente da Gestio:** Tanto a busca do cliente quanto o POST do pedido são síncronos. Nenhum dado local é usado. Nenhuma fila é utilizada. **Se a Gestio travar, é perda total do fluxo.**

---

## 7. Auditoria da Fila/Worker

### Tabela da fila: `fila_sincronizacao`
**Classe:** `app/classes/fila_sincronizacao.php` (145 linhas)

Colunas inferidas do código: `fila_id`, `fila_tipo`, `fila_alvo_id`, `fila_status`, `fila_payload`, `fila_retries`, `fila_erro_log`, `fila_criado_em`.

### Status existentes
| Status | Significado | Evidência |
|--------|-------------|-----------|
| `pending` | Aguardando processamento | L23 (Inserir), L127 (MarcarErro volta p/ pending se retries < 3) |
| `processing` | Sendo processado (lock) | L77 (PegarPendentes) |
| `success` | Concluído com sucesso | L102 (MarcarSucesso) |
| `error` | Falhou 3x, abandonado | L127 (MarcarErro) |

### Quando entra item
1. `salvar_cliente` → tipo `criar_cliente_otk_rd_v1` (clientes_otk.php L330)
2. `alterar_cliente` → tipo `alterar_cliente_otk_rd_v1` (clientes_otk.php L674)
3. Propostas → tipo `oportunidade_rd_v1` (em outros pontos do código)

### Quando sai item
Processado pelo Worker (`cron/processar_fila_sincronizacao.php`). Marcado `success` (L259) ou `error` (L263).

### Worker: processar_fila_sincronizacao.php (268 linhas)

#### Tipo `criar_cliente_otk_rd_v1` (L143-187)
**Payload enviado à Gestio:**
- `InserirClienteOtkWeb($dados_otkweb)` — POST `cliente`
- `CadastrarContato($contato)` para cada contato (com verificação de duplicata)
- `CriarEmpresa($razaoSocial, $cnpjCpf, $codigoClienteOtk)` — RD Station

**Validação:** Verifica `$inserir_otkweb['success']` (L156). Se vazio, lança Exception.
**Corpo da resposta:** Valida apenas `success`, não valida campos individuais.
**Atualiza banco local:** SIM — grava `cliente_rdid` se RD Station retornar `_id` (L181-186).

#### Tipo `alterar_cliente_otk_rd_v1` (L189-252)
**Payload enviado à Gestio:**
- `AlterarClienteOtkWeb($dados_otkweb)` — PUT `cliente`
- `DetalhesCliente($codigo_otk)` — busca contatos antigos para removê-los (L210)
- `RemoverContato` para cada contato existente
- `CadastrarContato` para cada contato novo
- `CriarEmpresa` no RD se não tiver `cliente_rdid`

**Validação:** Verifica `$alterar_cliente_otk['success']` (L204).

### ⚠️ BUG IDENTIFICADO NO WORKER (L244-252)
```php
$empresa = json_decode($criar_empresa, true);
if (!empty($rd_id)) {          // ← $rd_id é '' (L230), NUNCA entrará aqui
    $update_sql = "UPDATE clientes SET cliente_rdid = :rd_id ...";
```
A variável `$rd_id` é inicializada como `''` na L230. O resultado de `$criar_empresa` é decodificado em `$empresa` (L243), mas o código verifica `$rd_id` (que continua vazio) em vez de `$empresa['_id']`. **O RDID nunca é salvo para alterações de cliente que não tinham RDID anterior.** Além disso, falta o `}` de fechamento do `else` — a indentação sugere que as linhas 244-249 estão dentro do `else` da L237, mas o `break` da L252 fecha o `case` antes do `}` do `else`. Isso é um **bug de escopo** que pode causar comportamento inesperado.

### Tentativas e risco de duplicidade
- Máximo de 3 retries (L54, L127)
- **RISCO DE DUPLICIDADE:** Se o Worker executa `InserirClienteOtkWeb`, a API aceita, mas a conexão cai antes de receber a resposta, o Worker marca erro e tenta novamente. A API pode criar um segundo registro.
- **Sem idempotência:** Não há verificação se o cliente já foi criado na Gestio antes de tentar novamente.

---

## 8. Separação de Conceitos

### Cliente Temporário (regra comercial/manual)
- **O que é:** Flag `cliente_temporario` na tabela `clientes`. Valor 0 ou 1.
- **Quem controla:** O vendedor, via checkbox na tela de cadastro/edição.
- **Finalidade:** Indicar que o cadastro está incompleto (falta IE, endereço, etc.) e não deve ser considerado "definitivo" pela gestão.
- **NÃO tem relação** com fila de sincronização. Um cliente temporário=1 pode já ter sido sincronizado com sucesso na Gestio.

### Status de sincronização Gestio (regra técnica)
- **O que é:** O campo `fila_status` na tabela `fila_sincronizacao`.
- **Quem controla:** O sistema (Worker automático).
- **Finalidade:** Indicar se o envio dos dados para a API Gestio foi concluído.
- **NÃO confundir** com `cliente_temporario`. Um cliente pode ser temporário=0 e ter fila_status=error (dados completos mas API rejeitou).

### Origem dos dados
| Dado | Origem atual | Origem desejada (Local-First) |
|------|-------------|-------------------------------|
| Razão Social na proposta | API Gestio (`DetalhesCliente`) | Banco local (`clientes.cliente_info`) |
| Endereço no PV | API Gestio (`DetalhesCliente`) | Banco local (`clientes.cliente_endereco_*`) |
| Contatos na proposta | Híbrido (API + local se houver) | Banco local (`clientes.cliente_contatos`) |
| Pendências financeiras | API Gestio (`GetContasReceber`) | Sem equivalente local |
| Estado pagante | API Gestio (`GetEstadoPagante`) | Sem equivalente local |
| Dados do pedido OTK | API Gestio (POST + GET) | Sem equivalente local |
| Dados do pedido produção | API Gestio (`DetalhesCliente`) | Banco local |

---

## 9. Classificação Final por Área

| Área | Status | Justificativa |
|------|--------|---------------|
| **Cadastro de cliente (local)** | ✅ Validado | Dados persistem no banco local antes de ir à fila. Código sequencial sem colisão local. |
| **Cadastro de cliente (fila→Gestio)** | ⚠️ Parcialmente validado | Fila funciona, mas contato financeiro não persiste localmente. Sem idempotência no retry. |
| **Edição de cliente (local)** | ✅ Validado | Preserva representante. Bug de contatos corrigido. Salva antes da fila. |
| **Edição de cliente (fila→Gestio)** | 🔴 Quebrado | Bug no Worker (L244): RDID nunca é salvo para clientes novos via edição. Bug de escopo no else/break. |
| **Visualização de cliente** | ⚠️ Parcialmente validado | Banco local é prioridade (`ChecarDadosAC`). Fallback para API existe mas é síncrono e bloqueante. |
| **Verificação de duplicidade (CNPJ)** | ✅ Validado | Usa banco local (`LIKE %cnpj%`). Não chama API. |
| **Procurar cliente (proposta/orçamento)** | 🔴 Risco alto | 3 chamadas síncronas à API (DetalhesCliente + GetContasReceber + GetEstadoPagante). Gestio fora = tela trava. |
| **Criação de proposta (Insert)** | 🔴 Risco alto | `Proposta::Insert()` chama `DetalhesCliente()` direto. Gestio fora = proposta não é criada. |
| **Aprovação de proposta → pedido produção** | 🔴 Risco alto | `DetalhesCliente()` síncrono no momento da aprovação. Gestio fora = proposta fica aprovada sem pedido. |
| **Criação de PV local (pedidos-venda.php)** | ⚠️ Parcialmente validado | Endereços vêm do banco local, mas dados principais do cliente vêm da API. |
| **Criação de PV OTK (pv.php)** | 🔴 Risco alto | 100% síncrono, sem banco local, sem fila. |
| **Processamento de PV OTK** | 🔴 Risco alto | `ProcessarPedidoDeVenda` síncrono, sem retry, sem fila. |
| **NFe (emissão)** | 🔴 Risco alto | Duas chamadas a `DetalhesCliente` (L645, L2175 de nfe.php). Síncrono. |
| **Pré-postagem/Expedição** | 🔴 Risco alto | 4 chamadas a `DetalhesCliente` (prepostagem.php). Síncrono. |
| **Relatórios** | ⚠️ Parcialmente validado | `DetalhesCliente` chamado dentro de loop em relatórios (relatorios.php L70, RelatoriosV2Service.php L195). Performance ruim. |
| **Fila/Worker (fila_sincronizacao)** | ⚠️ Parcialmente validado | Estrutura sólida (lock, retry 3x), mas bug no RDID, sem idempotência, sem notificação de falha ao usuário. |
| **Contatos (CarregarContatos)** | ✅ Validado | Único método com fallback local real (otkweb.php L851-923). |
| **Persistência de contato financeiro** | 🔴 Quebrado | Dados existem apenas no payload da fila. Se a fila falhar 3x, são perdidos. Não há coluna local. |
| **Reivindicação de cliente** | 🔴 Risco alto | `DetalhesCliente` + `ReivindicarCliente` síncronos. Sem fila. |
| **Exclusão de cliente** | 🔴 Risco alto | `DetalhesCliente` síncrono para verificar vendedor + `ExcluirCliente` síncrono. |

### Legenda
- ✅ **Validado** — Funciona corretamente com evidência no código
- ⚠️ **Parcialmente validado** — Funciona mas com ressalvas/riscos menores
- 🔴 **Risco alto** — Dependência síncrona da Gestio que pode causar falha em produção
- 🔴 **Quebrado** — Bug verificado no código que causa comportamento incorreto

---

## 10. Nenhuma implementação neste documento

Este documento é exclusivamente um mapa do funcionamento atual com evidências de código. Nenhuma alteração foi feita. Todas as referências de linha são do código atual em produção.
