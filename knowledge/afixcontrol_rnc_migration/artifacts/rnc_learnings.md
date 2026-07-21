# Aprendizados da Migração de RNC e Pedidos de Retrabalho

Durante a integração da feature de Relatório de Não-Conformidade (RNC) na branch `dev`, identificamos falhas silenciosas de schema de banco de dados que causavam o travamento do frontend (botão "Salvando..." infinito).

## O Problema Identificado
Os arquivos oficiais de despejo de dados (`app/prodsis/database.sql` e `rnc_tabelas_predefinidas.sql`) **estavam desatualizados** em relação aos últimos commits do backend. As queries SQL de `INSERT` no PHP tentavam salvar dados em colunas que ainda não existiam no ambiente local.

### Colunas Ocultas Descobertas:
1. **`pedido_retrabalho`** (INT): Adicionada à tabela `pedidos`. Foi identificada rastreando a chamada de duplicação do pedido em `app/ajax/pedidos.php`.
2. **`produtos_selecionados`** (TEXT): Adicionada à tabela `nao_conformidades`. Foi descoberta apenas quando o salvamento do formulário falhou e o `app/ajax/rnc.php` revelou um campo JSON embutido na requisição que não estava em nenhum script `.sql`.
3. **`pedido_id`** (INT) e **`classificacao`** (ENUM): Também não estavam nos CREATE TABLE originais exportados.

## Protocolo Preventivo Adotado
Sempre que puxarmos novos módulos grandes para um ambiente legível sem ORM, o procedimento padrão deve ser:
1. Extrair os `.sql` providenciados.
2. Fazer `grep` reverso de `INSERT INTO {tabela_alvo}` nos arquivos `.php`.
3. Comparar o número de colunas da query contra as colunas do `CREATE TABLE` do `.sql` para evitar surpresas 500s de PDO no AJAX.

## Resolução
Criado e versionado um script de migração complementar: `app/prodsis/migrations/add_rnc_missing_columns.sql` contendo `IF NOT EXISTS` para imunizar o banco.
