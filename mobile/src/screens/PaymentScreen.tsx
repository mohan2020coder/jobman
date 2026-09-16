import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native'
import type { NativeStackScreenProps } from '@react-navigation/native-stack'
import type { RootStackParamList } from '../navigation/types'
import { addPayment, apiErrorMessage, jobRequest, listPayments } from '../api/client'
import {
  Button,
  Card,
  Chip,
  ErrorView,
  FormLabel,
  Icon,
  Input,
  Muted,
  SectionTitle,
  Spinner,
} from '../components/ui'
import { formatDateTime, formatMoney } from '../lib/format'
import { colors, methodColor, radius, spacing, tint } from '../theme'

type Props = NativeStackScreenProps<RootStackParamList, 'Payment'>

const methods = ['CASH', 'UPI', 'CARD', 'BANK_TRANSFER', 'OTHER']

export function PaymentScreen({ route }: Props) {
  const { id } = route.params
  const qc = useQueryClient()

  const job = useQuery({
    queryKey: ['job', id],
    queryFn: () => jobRequest(id),
  })

  const payments = useQuery({
    queryKey: ['job', id, 'payments'],
    queryFn: () => listPayments(id),
  })

  const [amount, setAmount] = useState('')
  const [method, setMethod] = useState('CASH')
  const [txnRef, setTxnRef] = useState('')
  const [error, setError] = useState('')

  const add = useMutation({
    mutationFn: async () => {
      const n = Number(amount)
      if (!n || n <= 0) throw new Error('Enter a valid amount.')
      const jobData = job.data
      const basis = jobData ? jobData.final_amount || jobData.estimated_amount : 0
      const paid = payments.data?.reduce((s, p) => s + (p.status === 'PAID' ? p.amount : 0), 0) ?? 0
      const remaining = Math.max(0, basis - paid)
      if (n > remaining) {
        throw new Error(`Payment exceeds the remaining balance of ${formatMoney(remaining)}.`)
      }
      return addPayment(id, {
        amount: n,
        method,
        transaction_reference: txnRef.trim() || undefined,
      })
    },
    onSuccess: () => {
      setAmount('')
      setTxnRef('')
      setError('')
      qc.invalidateQueries({ queryKey: ['job', id] })
      qc.invalidateQueries({ queryKey: ['job', id, 'payments'] })
      qc.invalidateQueries({ queryKey: ['job', id, 'receipt'] })
      qc.invalidateQueries({ queryKey: ['techjobs'] })
    },
  })

  if (job.isLoading || payments.isLoading) return <Spinner label="Loading payments…" />
  if (job.isError || !job.data) {
    return <ErrorView message="Unable to load this job." onRetry={() => job.refetch()} />
  }

  const totalPaid =
    payments.data?.reduce((s, p) => s + (p.status === 'PAID' ? p.amount : 0), 0) ?? 0
  const basis = job.data.final_amount || job.data.estimated_amount
  const remaining = Math.max(0, basis - totalPaid)
  const pct = basis > 0 ? Math.min(100, (totalPaid / basis) * 100) : 0
  const balanceColor = remaining <= 0 ? colors.success : colors.warning
  const paidFull = remaining <= 0

  return (
    <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView
        style={styles.flex}
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        <Card>
          <SectionTitle>Collection</SectionTitle>
          <View style={styles.summaryRow}>
            <View style={styles.summaryItem}>
              <Text style={styles.summaryLabel}>Job total</Text>
              <Text style={styles.summaryValue}>{formatMoney(basis)}</Text>
            </View>
            <View style={styles.summaryItem}>
              <Text style={styles.summaryLabel}>Paid</Text>
              <Text style={[styles.summaryValue, { color: colors.success }]}>{formatMoney(totalPaid)}</Text>
            </View>
            <View style={styles.summaryItem}>
              <Text style={styles.summaryLabel}>Balance</Text>
              <Text style={[styles.summaryValue, { color: balanceColor }]}>{formatMoney(remaining)}</Text>
            </View>
          </View>

          <View style={styles.progressTrack}>
            <View style={[styles.progressFill, { width: `${pct}%`, backgroundColor: paidFull ? colors.success : colors.primary }]} />
          </View>
          <Text style={styles.progressText}>
            {paidFull ? 'Fully paid' : `${pct.toFixed(0)}% collected · ${formatMoney(remaining)} due`}
          </Text>
        </Card>

        {job.data.customer ? (
          <View style={styles.customerRow}>
            <Text style={styles.customerText}>
              Collecting from {job.data.customer.name} · {job.data.service_type}
            </Text>
          </View>
        ) : null}

        <Card>
          <SectionTitle style={styles.cardTitle}>History</SectionTitle>
          {payments.data && payments.data.length > 0 ? (
            payments.data.map((p) => (
              <View key={p.id} style={styles.payRow}>
                <View style={[styles.methodIcon, { backgroundColor: methodColor[p.method] ?? colors.muted }]}>
                  <Text style={styles.methodIconText}>{p.method.slice(0, 1)}</Text>
                </View>
                <View style={styles.payMain}>
                  <Text style={styles.payMethod}>{p.method.replace(/_/g, ' ')}</Text>
                  <Muted>{formatDateTime(p.paid_at)}</Muted>
                  {p.transaction_reference ? (
                    <Muted>Ref · {p.transaction_reference}</Muted>
                  ) : null}
                </View>
                <Text style={styles.payAmount}>{formatMoney(p.amount)}</Text>
              </View>
            ))
          ) : (
            <Muted style={styles.noPayments}>No payments recorded yet.</Muted>
          )}
        </Card>

        {remaining > 0 ? (
          <Card>
            <SectionTitle style={styles.cardTitle}>Record payment</SectionTitle>

            <FormLabel>Amount</FormLabel>
            <View style={styles.amountRow}>
              <Input
                style={styles.amountInput}
                keyboardType="decimal-pad"
                placeholder="Enter amount"
                value={amount}
                onChangeText={setAmount}
              />
              <Button
                title="Full balance"
                size="sm"
                variant="soft"
                onPress={() => setAmount(String(remaining))}
              />
            </View>

            <FormLabel>Method</FormLabel>
            <View style={styles.methods}>
              {methods.map((m) => (
                <Chip
                  key={m}
                  label={m.replace(/_/g, ' ')}
                  active={method === m}
                  onPress={() => setMethod(m)}
                />
              ))}
            </View>

            <FormLabel>Transaction ref (optional)</FormLabel>
            <Input
              placeholder="e.g. UPI123456"
              value={txnRef}
              onChangeText={setTxnRef}
            />

            {error ? <Text style={styles.error}>{error}</Text> : null}
            {add.isError ? <Text style={styles.error}>{apiErrorMessage(add.error)}</Text> : null}

            <Button
              title="Save payment"
              size="lg"
              icon="checkmark-circle"
              full
              loading={add.isPending}
              disabled={add.isPending}
              style={{ marginTop: spacing.sm }}
              onPress={() => add.mutate()}
            />
          </Card>
        ) : (
          <View style={styles.paidBanner}>
            <Icon name="checkmark-circle" size={20} color={colors.success} />
            <Text style={styles.paidBannerText}>This job is fully paid.</Text>
          </View>
        )}

        <Muted style={styles.footNote}>
          Payments appear on the owner dashboard and on the public receipt.
        </Muted>
      </ScrollView>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl, gap: spacing.md },
  summaryRow: { flexDirection: 'row', justifyContent: 'space-between', marginTop: spacing.md },
  summaryItem: { flex: 1 },
  summaryLabel: {
    fontSize: 11,
    fontWeight: '700',
    color: colors.muted,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  summaryValue: { fontSize: 19, fontWeight: '800', color: colors.text, marginTop: 3 },
  progressTrack: {
    height: 8,
    borderRadius: radius.pill,
    backgroundColor: colors.surfaceAlt,
    overflow: 'hidden',
    marginTop: spacing.lg,
  },
  progressFill: { height: 8, borderRadius: radius.pill },
  progressText: {
    fontSize: 12,
    color: colors.muted,
    marginTop: 6,
    fontWeight: '600',
  },
  customerRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: tint(colors.primary, 0.06),
    borderRadius: radius.md,
    padding: spacing.md,
  },
  customerText: { color: colors.primaryDark, fontSize: 13, fontWeight: '700' },
  cardTitle: { marginBottom: spacing.md },
  payRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.md,
    paddingVertical: spacing.md,
  },
  methodIcon: {
    width: 36,
    height: 36,
    borderRadius: radius.sm,
    alignItems: 'center',
    justifyContent: 'center',
  },
  methodIconText: { color: colors.white, fontWeight: '800', fontSize: 13 },
  payMain: { flex: 1, gap: 1 },
  payMethod: { fontSize: 14, fontWeight: '700', color: colors.text },
  payAmount: { fontWeight: '800', color: colors.text, fontSize: 16 },
  noPayments: { paddingVertical: spacing.sm },
  amountRow: { flexDirection: 'row', gap: spacing.sm, alignItems: 'center' },
  amountInput: { flex: 1 },
  methods: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
  error: { color: colors.danger, fontSize: 13, marginTop: spacing.sm, fontWeight: '600' },
  paidBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
    backgroundColor: colors.successSoft,
    borderRadius: radius.md,
    padding: spacing.lg,
  },
  paidBannerText: { color: colors.success, fontWeight: '800', fontSize: 14 },
  footNote: { textAlign: 'center' },
})