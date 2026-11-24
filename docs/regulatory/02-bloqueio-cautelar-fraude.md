# Normativo Simulado — Bloqueio Cautelar Preventivo em Transações Pix

> Documento de referência simulada baseado nas diretrizes de prevenção a fraudes no arranjo Pix (Bloqueio Cautelar — até 72 horas).

---

## 1. Objetivo do Bloqueio Cautelar

Permitir que a instituição detentora da conta **recebedora** retenha preventivamente recursos oriundos de Pix sob suspeita de fraude antes da dispersão do saldo para contas laranja ("mule accounts").

## 2. Gatilhos Objetivos de Suspeita

1. **Conta Recebedora Recém-Criada:** Conta aberta há menos de 15 dias recebendo transferência de valor elevado (> R$ 1.500,00) seguida de tentativa imediata de transferência ou saque.
2. **Anomalia Comportamental / Velocidade:** Múltiplos créditos Pix de pagadores distintos sem histórico prévio em janela inferior a 30 minutos.
3. **Reputação da Chave / Dispositivo:** Chave Pix (CPF, telefone, e-mail ou EVP) ou device fingerprint vinculado a notificações de infração anteriores.

## 3. Limites Temporais e Comunicação

- **Duração máxima do Bloqueio Cautelar:** **72 horas** corridas.
- **Notificação obrigatória:** O titular da conta recebedora deve ser notificado imediatamente sobre a retenção cautelar e os canais para apresentação de justificativa.
- **Liberação ou Conversão em MED:** Ao final das 72 horas, caso não haja confirmação de fraude nem abertura de MED, os recursos devem ser liberados ao recebedor.
