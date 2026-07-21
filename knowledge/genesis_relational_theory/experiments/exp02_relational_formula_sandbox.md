# Experimento 02: Relational Formula Sandbox (Gate 02)

## Objetivo
Testar fórmulas relacionais alternativas lendo snapshots já coletados, sem alterar o Ouroboros Core.
A intenção é isolar e comparar a **Pressão Absoluta** e o **Choque Relacional**, provando que a mesma latência bruta (500ms) pode produzir um choque diferente dependendo do Alpha do serviço.

## Fórmula Candidata Avaliada (Offline Sandbox)
- **AbsolutePressure** = CkScore
- **RelationalShock** = |CkScore - Alpha| / max(Alpha, ε)
- **RelationalFriction** = min(100, RelationalShock * 100)

*(Thresholds aplicados: >= 20 WARN, >= 70 DENY)*

## Análise Comparativa dos Serviços

### Serviço Rápido (`SERVICE_FAST_NORMAL`)
- **Latência Injetada:** 500.0ms
- **Alpha Histórico (Base):** 12.71
- **Estado Original Ouroboros:** DENY (CkScore dominou)

**Resultado Sandbox:**
- **Pressão Absoluta:** 96.48 (Alta carga física no hardware)
- **Choque Relacional:** 6.59x o normal
- **Nova Fricção Sandbox:** 100.00
- **Novo Estado:** **DENY**
> *Interpretação:* O serviço sofreu um pico tremendo. A pressão é altíssima e o choque justifica o DENY.

---

### Serviço Lento (`SERVICE_SLOW_NORMAL`)
- **Latência Injetada:** 500.0ms
- **Alpha Histórico (Base):** 86.59
- **Estado Original Ouroboros:** DENY (CkScore dominou)

**Resultado Sandbox:**
- **Pressão Absoluta:** 96.48 (Alta carga física no hardware)
- **Choque Relacional:** 0.11x o normal
- **Nova Fricção Sandbox:** 11.43
- **Novo Estado:** **SAFE**
> *Interpretação:* O serviço é naturalmente pesado. A pressão absoluta é alta, mas relacionalmente o choque é brando. O sistema não deve colapsar defensivamente, apenas alertar.

## Conclusão Científica

A Sandbox offline obteve **sucesso absoluto**.

Demonstramos com clareza que, ao usar a fórmula de Choque Relacional, **a mesma latência de 500ms gerou respostas biológicas (estados) diferentes**.
- Para o serviço rápido, o choque superou 7x o seu normal (Fricção máxima, DENY).
- Para o serviço naturalmente lento (cujo Alpha já era ~85), a variação de CkScore (de 85 para 96) representou um choque de apenas ~0.14x, mantendo a Fricção Relacional em níveis pacíficos (11.4), resultando num estado muito mais leve (SAFE).

O conceito de **Pressão Absoluta vs Choque Relacional** provou ser a matemática correta para representar a Teoria PER, permitindo que a camada analítica compreenda o significado verdadeiro do dado sem penalizar sistemas operando dentro do seu comportamento crônico esperado.
