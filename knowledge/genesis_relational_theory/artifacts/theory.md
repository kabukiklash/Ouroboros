# Computação Relacional e o Processador de Estados Relacionais (PER)

## 1. Princípio Fundamental
A computação tradicional trata o dado como unidade central. O **PER** propõe que o significado não nasce do dado isolado, mas do **estado relacional verificável**.

Um estado relacional contém: valor, origem, contexto temporal, histórico, regras aplicadas, dependências, conflitos, impacto e validade.

## 2. Ouroboros como Laboratório (Organismo Mínimo)
O Ouroboros é governado pelo GenesisCore e serve como a plataforma mínima para validar a Teoria da Computação Relacional. Ele já processa um estado relacional mínimo: a latência (dado cru) não vale sozinha; ela é contextualizada contra o histórico (Alpha), a divergência e a fricção matemática, culminando em uma ação defensiva (SAFE/WARN/DENY).

### Separação de Papéis Arquiteturais
- **Core Go (Tronco Encefálico):** Simples, rápido, matemático e defensivo. Calcula fricção instantânea e executa gatilhos defensivos.
- **Sandbox Python/WASM (Córtex Analítico):** Ambiente fora do loop crítico focado na análise de padrões, grafos e geração de hipóteses (operando em read-only inicialmente).

## 3. Estágios de Validação do PER no Ouroboros
- **Alpha:** Semente do PER-M (Memória de normalidade operacional).
- **Divergência:** Semente do PER-D (Dinâmica temporal instantânea, apontando a distância do normal).
- **Padrões Estruturais (PER-R):** Reservado para ressonância estrutural entre domínios diferentes, e não apenas sazonalidade horária.

## 4. O Mapa de Evolução (Gates de Implementação)
- **Gate 00:** Memória Persistente da Teoria (este documento).
- **Gate 01:** Relational Snapshot Layer (Apenas observação e registro em modelo híbrido: colunas fortes + JSON flexível). Sem alteração matemática, sem IA no core, sem mudança de comportamento.
- **Gate 02:** Sandbox analisando e sugerindo.
- **Gate 03:** Ajuste de comportamento com autorização.
- **Gate 04:** Ajuste automático governado.
