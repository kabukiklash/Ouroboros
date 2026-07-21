# Traps e Bugs de Integração (ViPP & Gestio)

## 1. ViPP / Correios: Validação de Serviços Numéricos
A API de geração de etiquetas (ViPP) tem regras estritas de validação.
**O Erro:** A ViPP retorna `{"CodigoErro": 1400, "Descricao": "O ServicoECT É Obrigatóriamente Numérico!"}` se você enviar uma *string* legível como `'PAC'` ou `'SEDEX'` no payload json `['Servico']['ServicoECT']`.
**A Solução (Regra de Ouro):**
Sempre traduza nomes textuais do front-end/banco de dados para os códigos numéricos oficiais dos Correios ANTES de compilar o JSON da ViPP:
- `'PAC'` obrigatoriamente se converte para `'03298'`
- `'SEDEX'` obrigatoriamente se converte para `'03220'`
*Ref:* `app/ajax/pedidos-venda.php -> gerarEtiquetaCorreios()`

## 2. API Gestio: O Perigo do Null Coalescing (??) com Strings Vazias
Ao enviar dados de emissão de Notas Fiscais para o Gestio ERP (OtkWeb), é comum utilizar a estrutura de *fallback* para campos opcionais como `Nome Fantasia`.
**O Erro:** Utilizar `"nomeFantasia" => $nomeFantasia ?? $razaoSocial`.
O banco de dados do AfixControl envia campos vazios como `""` (empty string). O operador de coalescência nula `??` do PHP **ignora** strings vazias porque ele verifica apenas por `null`. Resultado: Uma string `""` é tratada como um valor válido, e a Gestio recebe o campo em branco, não caindo no fallback da Razão Social.
**A Solução (Regra de Ouro):**
Sempre valide payloads da Gestio com verificações rígidas de vazio:
`"nomeFantasia" => !empty($nomeFantasia) ? $nomeFantasia : $razaoSocial`
*Ref:* `app/ajax/nfe.php -> Montar payloads da Gestio`
