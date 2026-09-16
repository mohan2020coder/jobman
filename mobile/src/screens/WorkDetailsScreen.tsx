import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import {
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native'
import type { NativeStackScreenProps } from '@react-navigation/native-stack'
import type { RootStackParamList } from '../navigation/types'
import { apiErrorMessage, jobComplete, jobRequest } from '../api/client'
import type { Job } from '../api/types'
import {
  Badge,
  Button,
  Card,
  Chip,
  Divider,
  ErrorView,
  FormLabel,
  Icon,
  IconButton,
  Input,
  SectionTitle,
  Spinner,
} from '../components/ui'
import { formatMoney } from '../lib/format'
import { colors, radius, spacing } from '../theme'

type Props = NativeStackScreenProps<RootStackParamList, 'WorkDetails'>

interface ItemRow {
  key: string
  description: string
  quantity: string
  unitPrice: string
}

let rowSeq = 0
function nextKey() {
  rowSeq += 1
  return `r${rowSeq}`
}

const methods = ['CASH', 'UPI', 'CARD', 'BANK_TRANSFER', 'OTHER']

export function WorkDetailsScreen({ navigation, route }: Props) {
  const { id } = route.params
  const qc = useQueryClient()

  const job = useQuery({
    queryKey: ['job', id],
    queryFn: () => jobRequest(id),
  })

  const [notes, setNotes] = useState('')
  const [items, setItems] = useState<ItemRow[]>([
    { key: nextKey(), description: '', quantity: '1', unitPrice: '' },
  ])
  const [collectPayment, setCollectPayment] = useState(true)
  const [method, setMethod] = useState('CASH')
  const [txnRef, setTxnRef] = useState('')
  const [paymentAmount, setPaymentAmount] = useState('')

  function addItem() {
    setItems((prev) => [...prev, { key: nextKey(), description: '', quantity: '1', unitPrice: '' }])
  }

  function removeItem(key: string) {
    setItems((prev) => (prev.length > 1 ? prev.filter((i) => i.key !== key) : prev))
  }

  function updateItem(key: string, patch: Partial<ItemRow>) {
    setItems((prev) => prev.map((i) => (i.key === key ? { ...i, ...patch } : i)))
  }

  const total = items.reduce((sum, i) => {
    const qty = Number(i.quantity) || 1
    const unit = Number(i.unitPrice) || 0
    return sum + qty * unit
  }, 0)

  const complete = useMutation({
    mutationFn: async () => {
      const validItems = items
        .map((i) => ({
          description: i.description.trim(),
          quantity: Number(i.quantity) || 1,
          unit_price: Number(i.unitPrice) || 0,
        }))
        .filter((i) => i.description !== '')
      if (validItems.length === 0) {
        throw new Error('Add at least one item with a description before completing.')
      }
      const payment =
        collectPayment && paymentAmount !== ''
          ? {
              amount: Number(paymentAmount) || 0,
              method,
              transaction_reference: txnRef.trim() || undefined,
            }
          : null
      if (payment && payment.amount <= 0) {
        throw new Error('Payment amount must be greater than zero.')
      }
      if (payment && payment.amount > total) {
        throw new Error(`Payment cannot exceed the job total of ${formatMoney(total)}.`)
      }
      return jobComplete(id, { notes: notes.trim() || undefined, items: validItems, payment })
    },
    onSuccess: (updated: Job) => {
      qc.invalidateQueries({ queryKey: ['job', id] })
      qc.invalidateQueries({ queryKey: ['job', id, 'payments'] })
      qc.invalidateQueries({ queryKey: ['techjobs'] })
      navigation.navigate('JobDetail', { id })
    },
  })

  function confirmComplete() {
    const paymentNote = collectPayment && paymentAmount !== '' ? ` + ${formatMoney(Number(paymentAmount) || 0)} via ${method}` : ''
    Alert.alert(
      'Complete this job?',
      `Total charges ${formatMoney(total)}.${paymentNote.length ? `\n\nPayment: ${paymentNote.replace(' + ', '')}.` : '\n\nNo payment will be recorded.'}`,
      [
        { text: 'Review', style: 'cancel' },
        { text: 'Complete', onPress: () => complete.mutate() },
      ],
    )
  }

  if (job.isLoading) return <Spinner label="Loading job…" />
  if (job.isError || !job.data) {
    return <ErrorView message="Unable to load this job." onRetry={() => job.refetch()} />
  }

  const j = job.data

  return (
    <KeyboardAvoidingView style={styles.flex} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
      <ScrollView
        style={styles.flex}
        contentContainerStyle={styles.content}
        keyboardShouldPersistTaps="handled"
        showsVerticalScrollIndicator={false}
      >
        <Card style={styles.headCard}>
          <View style={styles.headRow}>
            <View style={styles.headMain}>
              <Text style={styles.customer}>{j.customer?.name ?? 'Customer'}</Text>
              <Text style={styles.sub}>{j.job_number} · {j.service_type}</Text>
            </View>
            <Badge value={j.status} />
          </View>
        </Card>

        <Card>
          <SectionTitle style={styles.cardTitle}>Work notes</SectionTitle>
          <Input
            style={styles.textArea}
            multiline
            numberOfLines={4}
            placeholder="What did you do? Parts replaced, diagnostics, etc."
            value={notes}
            onChangeText={setNotes}
          />
        </Card>

        <Card>
          <View style={styles.cardHeadRow}>
            <SectionTitle style={styles.cardTitle}>Items / charges</SectionTitle>
            <Button title="Add item" variant="soft" size="sm" icon="add" onPress={addItem} />
          </View>

          {items.map((item, index) => (
            <View key={item.key} style={styles.itemRow}>
              <View style={styles.itemIndex}>
                <Text style={styles.itemIndexText}>{index + 1}</Text>
              </View>
              <View style={styles.itemFields}>
                <Input
                  placeholder="Description (e.g. Gas refill)"
                  value={item.description}
                  onChangeText={(t) => updateItem(item.key, { description: t })}
                />
                <View style={styles.itemInline}>
                  <Input
                    style={styles.qty}
                    placeholder="Qty"
                    keyboardType="decimal-pad"
                    value={item.quantity}
                    onChangeText={(t) => updateItem(item.key, { quantity: t })}
                  />
                  <Input
                    style={styles.price}
                    placeholder="Unit price"
                    keyboardType="decimal-pad"
                    value={item.unitPrice}
                    onChangeText={(t) => updateItem(item.key, { unitPrice: t })}
                  />
                </View>
              </View>
              {items.length > 1 ? (
                <IconButton
                  name="trash-outline"
                  color={colors.danger}
                  bg={colors.dangerSoft}
                  onPress={() => removeItem(item.key)}
                />
              ) : null}
            </View>
          ))}

          <Divider />
          <View style={styles.totalRow}>
            <Text style={styles.totalLabel}>Total</Text>
            <Text style={styles.totalValue}>{formatMoney(total)}</Text>
          </View>
        </Card>

        <Card>
          <SectionTitle style={styles.cardTitle}>Payment</SectionTitle>
          <View style={styles.toggleRow}>
            <Chip label="Record payment" icon="cash-outline" active={collectPayment} onPress={() => setCollectPayment(true)} />
            <Chip label="No payment" icon="close" active={!collectPayment} onPress={() => setCollectPayment(false)} />
          </View>

          {collectPayment ? (
            <>
              <FormLabel>Amount paid</FormLabel>
              <View style={styles.amountRow}>
                <Input
                  style={styles.amountInput}
                  keyboardType="decimal-pad"
                  placeholder={`Up to ${formatMoney(total)}`}
                  value={paymentAmount}
                  onChangeText={setPaymentAmount}
                />
                <Button
                  title="Full"
                  size="sm"
                  variant="soft"
                  onPress={() => setPaymentAmount(String(total))}
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
            </>
          ) : null}
        </Card>

        {complete.isError ? (
          <View style={styles.errorBanner}>
            <Icon name="alert-circle-outline" size={15} color={colors.danger} />
            <Text style={styles.errorText}>{apiErrorMessage(complete.error)}</Text>
          </View>
        ) : null}

        <Button
          title="Complete job"
          size="lg"
          icon="checkmark-circle"
          full
          loading={complete.isPending}
          disabled={complete.isPending}
          onPress={confirmComplete}
        />
      </ScrollView>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl, gap: spacing.md },
  headCard: {},
  headRow: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  headMain: { flex: 1 },
  customer: { fontSize: 18, fontWeight: '800', color: colors.text },
  sub: { color: colors.muted, fontSize: 13, marginTop: 2 },
  cardTitle: { marginBottom: spacing.md },
  cardHeadRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  textArea: {
    minHeight: 90,
    textAlignVertical: 'top',
  },
  itemRow: { flexDirection: 'row', alignItems: 'flex-start', gap: spacing.sm, marginBottom: spacing.md },
  itemIndex: {
    width: 26,
    height: 26,
    borderRadius: 13,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 2,
  },
  itemIndexText: { color: colors.primaryDark, fontWeight: '800', fontSize: 12 },
  itemFields: { flex: 1, gap: spacing.sm },
  itemInline: { flexDirection: 'row', gap: spacing.sm },
  qty: { width: 90 },
  price: { flex: 1 },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  totalLabel: { fontWeight: '800', color: colors.text },
  totalValue: { fontWeight: '800', color: colors.primary, fontSize: 18 },
  toggleRow: { flexDirection: 'row', gap: spacing.sm, marginBottom: spacing.md },
  amountRow: { flexDirection: 'row', gap: spacing.sm, alignItems: 'center' },
  amountInput: { flex: 1 },
  methods: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.sm },
  errorBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: colors.dangerSoft,
    borderRadius: radius.md,
    padding: spacing.md,
  },
  errorText: { color: colors.danger, fontSize: 13, fontWeight: '600', flex: 1 },
})