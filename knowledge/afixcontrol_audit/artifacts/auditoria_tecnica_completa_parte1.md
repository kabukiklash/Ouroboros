# Auditoria Técnica AfixControl — Parte 1: DetalhesCliente + API OTK

## 1. Todos os locais de chamada `DetalhesCliente()`

Evidência: `grep -rn "DetalhesCliente" app/` — 30 call-sites reais (excluindo definição e logs).

| # | Arquivo | Linha | Finalidade | Impacta |
|---|---------|-------|------------|---------|
| 1 | `classes/otkweb.php` | 403 | **Definição do método** — GET `cliente/{id}` | — |
| 2 | `classes/otkweb.php` | 515 | Chamado dentro de `ReivindicarCliente()` | Cliente |
| 3 | `classes/otkweb.php` | 1147 | Chamado dentro de `InserirClienteAC()` | Cliente |
| 4 | `classes/propostas.php` | 49 | `Proposta::Insert()` — busca razão social para `proposta_nome_cliente` | Proposta |
| 5 | `classes/proposta_view_model.php` | 33 | `GetDadosView()` — monta dados p/ PDF/view | Proposta |
| 6 | `classes/relatorios.php` | 70 | Relatório — busca nome do cliente por pedido | Faturamento |
| 7 | `classes/RelatoriosV2Service.php` | 195 | Relatório V2 — busca dados do cliente | Faturamento |
| 8 | `classes/afixcontrol.php` | 116 | Método genérico de busca de cliente | Cliente |
| 9 | `_clientes/cliente.php` | 53 | **Fallback** quando `ChecarDadosAC` não encontra no banco local | Cliente |
| 10 | `_clientes/alterar-cliente.php` | 16 | Carrega dados do cliente para form de edição | Cliente |
| 11 | `_propostas/proposta.php` | 97 | Exibição da proposta individual | Proposta |
| 12 | `_propostas/orcamento.php` | 46 | Exibição do orçamento público | Proposta |
| 13 | `pedido-venda.php` | 131 | Carrega contato financeiro no pedido de venda | Pedido |
| 14 | `modelo_layout.php` | 18 | Geração de layout/modelo da proposta | Proposta |
| 15 | `index.php` | 205 | Dashboard — lista propostas com nome do cliente | Proposta |
| 16 | `index.php` | 301 | Dashboard — lista pedidos com nome do cliente | Pedido |
| 17 | `control.php` | 199 | Dashboard admin — propostas | Proposta |
| 18 | `control.php` | 295 | Dashboard admin — pedidos | Pedido |
| 19 | `etiqueta-simples.php` | 57 | Impressão de etiqueta simples | Expedição |
| 20 | `cron/processar_fila_sincronizacao.php` | 210 | Worker — busca contatos atuais antes de refazer | Fila/Sync |
| 21 | `ajax/clientes_otk.php` | 412 | `deletar_cliente` — verifica vendedor | Cliente |
| 22 | `ajax/clientes_otk.php` | 439 | `reivindicar-cliente` — busca vendedor anterior | Cliente |
| 23 | `ajax/clientes_otk.php` | 731 | `procurar_cliente` — busca para nova proposta/orçamento | Proposta |
| 24 | `ajax/clientes_otk.php` | 810 | `checar_pendencias` — verifica pendências financeiras | Proposta |
| 25 | `ajax/pv.php` | 45 | `abrir-pedido` — monta payload do pedido OTK | Pedido |
| 26 | `ajax/pedidos-venda.php` | 162 | `criarPedidoVenda` — busca endereço para PV local | Pedido |
| 27 | `ajax/pedidos-venda.php` | 1016 | `atualizarEstadosPedidos` — busca dados ao atualizar | Pedido |
| 28 | `ajax/pedidos-venda.php` | 1459 | `getOtkClientData` — endpoint dedicado busca cliente | Pedido |
| 29 | `ajax/nfe.php` | 645 | Emissão de NFe — busca dados do cliente | Faturamento |
| 30 | `ajax/nfe.php` | 2175 | Emissão NFe via PV — busca dados do cliente | Faturamento |
| 31 | `ajax/propostas.php` | 440 | `alterar-estado` (aprovação) — busca p/ gerar pedido produção | Proposta/Pedido |
| 32 | `ajax/propostas.php` | 512 | `confirmar-pagamento` — gera pedido após pgto antecipado | Proposta/Pedido |
| 33 | `ajax/propostas.php` | 1468 | Outro ponto de geração de proposta | Proposta |
| 34 | `ajax/propostas.php` | 2721 | Outro ponto de consulta de cliente | Proposta |
| 35 | `ajax/rdstation.php` | 163 | Sincronização RD — busca dados do cliente | Cliente/CRM |
| 36 | `ajax/rdstation.php` | 577 | Sincronização RD — busca dados do cliente | Cliente/CRM |
| 37 | `ajax/sincronizar_clientes.php` | 53 | Sync em massa — busca detalhes por código | Cliente |
| 38 | `ajax/sincronizar_clientes.php` | 114 | Sync em massa — segundo ponto | Cliente |
| 39 | `ajax/solicitacao.php` | 18 | Solicitação — busca dados do cliente | Cliente |
| 40 | `ajax/prepostagem.php` | 41 | Pré-postagem Correios — busca endereço | Expedição |
| 41 | `ajax/prepostagem.php` | 278 | Pré-postagem — busca cliente do pedido | Expedição |
| 42 | `ajax/prepostagem.php` | 295 | Pré-postagem — fallback busca | Expedição |
| 43 | `ajax/prepostagem.php` | 435 | Pré-postagem — etiqueta | Expedição |
| 44 | `ajax/orcamento.php` | 45 | Ação de orçamento — busca cliente | Proposta |
| 45 | `ajax/orcamento.php` | 136 | Ação de orçamento — segunda chamada | Proposta |
| 46 | `ajax/orcamento.php` | 220 | Ação de orçamento — terceira chamada | Proposta |

**Total: 44 call-sites reais** (excluindo a definição do método e linhas de log/error_log).

## 2. Todos os endpoints da API OTK/Gestio chamados via `otkweb.php`

| Método OtkWeb | HTTP | Endpoint API | Valida corpo? | Salva local? | Usa fila? | Fallback local? |
|---|---|---|---|---|---|---|
| `DetalhesCliente($id)` | GET | `cliente/{id}` | Sim (`success`) | Não | Não | Não |
| `CarregarClientes($date)` | GET | `cliente?dataSuperior={date}` | Sim | Não | Não | Não |
| `CarregarMeusClientes($cod)` | GET | `cliente?codigoDoVendedor={cod}` | Sim | Não | Não | Não |
| `BuscarClientePorCnpjCpf($doc)` | GET | `cliente?cnpj={doc}` ou `?cpf={doc}` | Sim | Não | Não | Não |
| `InserirClienteOtkWeb($dados)` | POST | `cliente` | Sim (`success`) | Não direto | **Sim** (via caller) | Não |
| `AlterarClienteOtkWeb($dados)` | PUT | `cliente` | Sim | Não direto | **Sim** (via caller) | Não |
| `ExcluirCliente($id)` | DELETE | `cliente/{id}` | Sim | Não | Não | Não |
| `CadastrarContato($c)` | POST | `cliente/contato` | Sim | Não | Não | Não |
| `AlterarContato($c)` | PUT | `cliente/contato` | Sim | Não | Não | Não |
| `RemoverContato($cod,$seq)` | DELETE | `cliente/{cod}/contato/{seq}` | Sim | Não | Não | Não |
| `CarregarContatos($id)` | GET | `cliente/{id}/contato` | Sim | Parcial* | Não | Sim* |
| `ListarContatosApiOtk($id)` | GET | `cliente/{id}/contato` | Sim | Não | Não | Não |
| `ReivindicarCliente($cod,$vend)` | PUT | `cliente` | Sim | Sim (representante) | Não | Não |
| `PostPedidoDeVenda($data)` | POST | `faturamento/pedido` | Sim (`success`) | Sim (numero_pedido) | Não | Não |
| `GetPedidoDeVenda($id)` | GET | `faturamento/pedido/{id}` | Sim | Não | Não | Não |
| `UpdatePedidoDeVenda($data)` | PUT | `faturamento/pedido` | Sim | Não | Não | Não |
| `ProcessarPedidoDeVenda($id)` | POST | `faturamento/pedido/{id}/processar` | Sim | Não | Não | Não |
| `DeletePedidoDeVenda($id)` | DELETE | `faturamento/pedido/{id}` | Sim | Não | Não | Não |
| `CarregarPedidos($req)` | GET | `faturamento/pedido/{req}` | Sim | Não | Não | Não |
| `GetContasReceber()` | GET | `financeiro/contasreceber/` | Sim | Não | Não | Não |
| `GetNotasFiscais(...)` | GET | `faturamento/nfe/{estado}?...` | Sim | Não | Não | Não |
| `GetNotaFiscal($id)` | GET | `faturamento/nfe/{id}` | Sim | Não | Não | Não |
| `PostNotaFiscal($data)` | POST | `faturamento/nfe` | Sim | Sim (numero_nfe) | Não | Não |
| `PutNotaFiscal($data)` | PUT | `faturamento/nfe` | Sim | Não | Não | Não |
| `ProcessarNotaFiscal($id)` | POST | `faturamento/nfe/{id}/processar` | Sim | Não | Não | Não |
| `CondicoesDePagamento()` | GET | `financeiro/condicaodepagamento` | Sim | Não | Não | Não |
| `FormasDePagamento()` | GET | `financeiro/formadepagamento` | Sim | Não | Não | Não |
| `EstadoPagante($cli,$est)` | PUT | `cliente/estadopagante` | Sim | Não | Não | Não |
| `GetEstadoPagante($cli)` | GET | `cliente/{cli}/estadopagante` | Sim | Não | Não | Não |
| `TodosOsProdutos()` | GET | `produto` | Sim | Não | Não | Não |
| `GetFiliais()` | GET | `filial` | Sim | Não | Não | Não |
| `CarregarFornecedores()` | GET | `fornecedor?...` | Sim | Não | Não | Não |

*`CarregarContatos`: Dentro do método (L851-923 otkweb.php), há lógica que tenta ler `cliente_contatos` do banco local primeiro. Se encontrar, retorna. Se não, chama a API e salva no banco. É o **único método com fallback local real**.
