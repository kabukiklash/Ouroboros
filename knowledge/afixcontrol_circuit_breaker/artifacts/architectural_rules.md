# Arquitetura de Circuit Breaker (Gestio & RD Station)

## O Problema do Monothread
O servidor de desenvolvimento do AfixControl pode se comportar como single-thread. Quando as APIs da Gestio (`OtkWeb`) ou RD Station demoram (e.g. 30s) para responder, elas bloqueiam a thread inteira. 
As próximas chamadas (CSS, JS, AJAX secundários) ficam presas na fila e sofrem `net::ERR_CONNECTION_RESET`, resultando na **Paralisia da Tela** (CSS não carrega, UI congela no estado `Carregando...`, Sessão cai).

## Regra de Ouro: Separação por Tipo de Operação
Para estabilizar a rede sem corromper transações, adotamos uma abordagem de duplo-timeout:

1. **LEITURAS (GET - Visuais)**: Timeout Curto (Circuit Breaker)
   - *Onde*: `CarregarContatos`, `DetalhesCliente`, `SelecionarOportunidade`, etc.
   - *Tempo*: `timeout => 8`, `connect_timeout => 3`
   - *Por quê*: Se a leitura falhar, o usuário vê apenas um aviso na UI ("Erro de Conexão com a Gestio") sem congelar o navegador inteiro.

2. **ESCRITAS (POST/PUT - Críticas)**: Timeout Longo (Tolerante)
   - *Onde*: `CriarCliente`, `EmitirNfe`, `SalvarPedido`, `sincronizar_clientes.php`.
   - *Tempo*: Default Guzzle (`timeout => 30`).
   - *Por quê*: Cortar uma emissão de NF-e no meio do caminho gera "falso negativo". O sistema acusaria "Timeout" para o usuário, mas a SEFAZ processaria a nota. O usuário clicaria novamente, gerando notas em duplicidade.

## Tratamento de Exceções
Sempre capture o timeout sem disparar `Fatal Error` 500 para proteger o AJAX.
Exemplo no `OtkGuzzleWrapper::__call`:
```php
catch (\GuzzleHttp\Exception\ConnectException $e) {
    return new \GuzzleHttp\Psr7\Response(
        504, 
        ['Content-Type' => 'application/json'], 
        json_encode(['success' => false, 'error' => true, 'message' => 'Timeout'])
    );
}
```
Isso garante que `$request->getBody()->getContents()` nos arquivos de classe não quebre.
