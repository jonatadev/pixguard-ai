# Manual Operacional Simulado — PIX & Mecanismo Especial de Devolução (MED)

> Documento didático de referência regulatória simulada para o **PixGuard AI**, inspirado nos normativos públicos do Banco Central do Brasil (Resolução BCB nº 1/2020 e atualizações do MED).

---

## 1. Elegibilidade ao MED (Art. 41-B a 41-J)

O Mecanismo Especial de Devolução (MED) aplica-se exclusivamente quando há **fundada suspeita de fraude** ou **falha operacional** nos sistemas das instituições participantes do arranjo Pix.

### Casos Elegíveis
1. **Fraude / Golpe com Engenharia Social:** Golpe da Falsa Central Bancária, falso parente no WhatsApp, phishing com captura de credenciais, invasão de conta (account takeover) ou coação.
2. **Falha Operacional do PSP:** Duplicidade indevida de lançamento ou erro técnico de liquidação atribuível à instituição.

### Casos NÃO Elegíveis
1. **Desacordo Comercial:** Produto não entregue, defeito de mercadoria, arrependimento de compra ou divergência contratual entre comprador e vendedor legítimo (Art. 41-C, § 2º).
2. **Erro de Digitação do Pagador:** Envio voluntário para chave Pix incorreta sem indício de fraude do recebedor (deve seguir fluxo de solicitação amigável de devolução).

---

## 2. Prazos Operacionais Obrigatórios

| Regra | Prazo | Implicação na Decisão |
| :--- | :--- | :--- |
| **Janela de Solicitação pelo Pagador** | Até **80 dias** corridos da data da transação Pix | Contestações após 80 dias devem ser **indeferidas automaticamente por decurso de prazo**. |
| **Bloqueio dos Recursos na Conta Recebedora** | Imediato após abertura da notificação de infração | Instituição recebedora retém saldo disponível até conclusão da análise. |
| **Prazo de Análise pela Instituição Recebedora** | Até **7 dias** corridos | Aprovação total/parcial ou rejeição fundamentada. |
| **Devolução Efetiva (se procedente)** | Até **96 horas** após conclusão positiva | Valores bloqueados retornam ao pagador; saldo insuficiente gera acompanhamento contínuo. |
| **Monitoramento de Saldo (Devoluções Parciais)** | Até **90 dias** | Varreduras periódicas na conta recebedora enquanto houver saldo pendente. |

---

## 3. Critérios de Priorização de Risco

- **Risco ALTO:** Indício de falsa central, conta recebedora recém-aberta (< 15 dias), pulverização imediata (cash-out rápido) ou marcação prévia de fraude na chave Pix.
- **Risco MÉDIO:** Transação atípica fora do horário habitual ou pedido próximo ao limite da janela de 80 dias sem histórico prévio na conta recebedora.
- **Risco BAIXO:** Desacordo comercial declarado pelo próprio cliente ou transferência recorrente para favorecido habitual.
