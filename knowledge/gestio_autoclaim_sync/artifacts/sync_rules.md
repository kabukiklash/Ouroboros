# Autocura (Claim-on-Conflict) e Prevenção de Timeouts na Fila de Sincronização

## O Problema do "Veneno de Fila" na Criação de Clientes
Quando a Fila de Sincronização tentava disparar a tarefa `criar_cliente_otk_rd_v1` para um CNPJ que já havia sido criado por outro usuário na nuvem (ERP Gestio), a API retornava **HTTP 400 - CNPJ já cadastrado**.
Isso causava um bloqueio definitivo:
1. O ID Oficial da Gestio não era devolvido.
2. O sistema local preservava o ID temporário (`queue_rd_...`).
3. O freio "Anti-Órfão" (que checa se o cliente tem ID válido antes de enviar Propostas pro RD Station) barrava o envio das propostas atreladas a esse cliente.

## Solução Arquitetural: Claim-on-Conflict (Upsert)
Foi injetada a lógica de "Reivindicação" no arquivo `app/cron/processar_fila_sincronizacao.php`.

**Fluxo de Autocura:**
1. A API retorna HTTP 400 e a mensagem de erro da Gestio contém a string `'já cadastrado'`.
2. O script imediatamente dispara um `GET /clientes?cnpj=...` usando o método `BuscarClientePorCnpjCpf` da classe `OtkWeb`.
3. O script extrai o ID Oficial da Gestio.
4. **CASCATA OBRIGATÓRIA:** Como o AfixControl local utilizava a chave de relacionamento `cliente_codigo` (o próprio ID da Gestio) como chave primária comportamental nas outras tabelas, **você não pode atualizar apenas a tabela clientes**.
   - `UPDATE clientes SET cliente_codigo = :otkid WHERE cliente_codigo = :codigo`
   - `UPDATE propostas SET proposta_cliente = :otkid WHERE proposta_cliente = :codigo`
   - `UPDATE fila_sincronizacao SET fila_alvo_id = :otkid WHERE fila_alvo_id = :codigo`
5. O fluxo continua normalmente para inserir contatos e criar o registro no RD Station, destravando a fila.

## Rate Limiting e Prevenção de ConnectException (Guzzle)
A Fila processa dezenas de tarefas em milissegundos num laço `foreach`. A API de Desenvolvimento da Gestio bloqueia conexões em massa gerando Timeouts de Rede, que o Guzzle interpreta como `ConnectException` e aborta o worker inteiro de forma fatal.

**A Regra de Ouro da Fila Sincrona:**
Toda iteração do Cron que se comunica com a Gestio em lote **DEVE** possuir um freio de *rate-limiting*.
Um simples `sleep(2);` no final do laço `foreach` do arquivo `processar_fila_sincronizacao.php` reduziu os Timeouts a zero sem afetar a percepção do usuário. Nunca processe integração de terceiros em massa sem espaçamento local.
