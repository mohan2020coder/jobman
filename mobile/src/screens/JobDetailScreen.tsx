import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Alert, Linking, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native'
import type { NativeStackScreenProps } from '@react-navigation/native-stack'
import type { RootStackParamList } from '../navigation/types'
import {
  API_BASE,
  apiErrorMessage,
  jobAccept,
  jobOnTheWay,
  jobRequest,
  jobStart,
  receiptUrlRequest,
} from '../api/client'
import {
  Avatar,
  Badge,
  Button,
  Card,
  Divider,
  ErrorView,
  Icon,
  IconButton,
  Label,
  Muted,
  Pill,
  SectionTitle,
  Spinner,
} from '../components/ui'
import { StatusTimeline } from '../components/StatusTimeline'
import { formatDateTime, formatMoney, formatSchedule } from '../lib/format'
import { colors, methodColor, radius, spacing, statusColor } from '../theme'

type Props = NativeStackScreenProps<RootStackParamList, 'JobDetail'>

export function JobDetailScreen({ navigation, route }: Props) {
  const { id } = route.params
  const qc = useQueryClient()

  const job = useQuery({
    queryKey: ['job', id],
    queryFn: () => jobRequest(id),
    staleTime: 15_000,
  })

  const refetchAll = () => {
    qc.invalidateQueries({ queryKey: ['job', id] })
    qc.invalidateQueries({ queryKey: ['job', id, 'payments'] })
    qc.invalidateQueries({ queryKey: ['job', id, 'receipt'] })
    qc.invalidateQueries({ queryKey: ['techjobs'] })
  }

  const action = useMutation({
    mutationFn: async () => {
      const j = job.data
      if (!j) throw new Error('Job not loaded')
      switch (j.status) {
        case 'PENDING':
          return jobAccept(id)
        case 'ACCEPTED':
          return jobOnTheWay(id)
        case 'ON_THE_WAY':
          return jobStart(id)
        default:
          throw new Error('No action available')
      }
    },
    onSuccess: refetchAll,
  })

  const receipt = useQuery({
    queryKey: ['job', id, 'receipt'],
    queryFn: () => receiptUrlRequest(id),
    enabled: job.data?.status === 'COMPLETED',
  })

  if (job.isLoading) {
    return (
      <View style={styles.flex}>
        <Spinner label="Loading job…" />
      </View>
    )
  }
  if (job.isError || !job.data) {
    return (
      <View style={styles.flex}>
        <ErrorView message="Unable to load this job." onRetry={() => job.refetch()} />
      </View>
    )
  }

  const j = job.data
  const status = j.status
  const amount = j.final_amount || j.estimated_amount
  const paid = j.total_paid ?? 0
  const remaining = Math.max(0, amount - paid)
  const payColor = statusColor[j.payment_status as keyof typeof statusColor]

  const actionLabel =
    status === 'PENDING'
      ? 'Accept job'
      : status === 'ACCEPTED'
        ? 'On the way'
        : status === 'ON_THE_WAY'
          ? 'Start job'
          : null
  const actionIcon =
    status === 'PENDING'
      ? 'checkmark-done'
      : status === 'ACCEPTED'
        ? 'navigate'
        : status === 'ON_THE_WAY'
          ? 'play'
          : undefined

  function confirmAnd(actionName: string, fn: () => void) {
    Alert.alert(`Confirm ${actionName.toLowerCase()}?`, `Mark this job as "${actionName}"?`, [
      { text: 'Cancel', style: 'cancel' },
      { text: actionName, onPress: fn },
    ])
  }

  function openMaps() {
    const q = j.latitude && j.longitude ? `${j.latitude},${j.longitude}` : (j.address ?? j.customer?.address ?? '')
    Linking.openURL(`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(q)}`).catch(() => {})
  }

  function callCustomer() {
    const phone = j.customer?.phone
    if (phone) Linking.openURL(`tel:${phone}`).catch(() => {})
  }

  const detailRows: { label: string; value: string; icon?: 'time-outline' | 'location-outline' | 'pricetag-outline' | 'wallet-outline' | 'document-text-outline' }[] = [
    { label: 'Scheduled', value: formatSchedule(j.scheduled_at), icon: 'time-outline' },
    { label: 'Problem', value: j.problem_description ?? '—', icon: 'document-text-outline' },
    { label: 'Address', value: j.address ?? j.customer?.address ?? '—', icon: 'location-outline' },
    { label: 'Estimated', value: formatMoney(j.estimated_amount), icon: 'pricetag-outline' },
  ]
  if (j.final_amount > 0) detailRows.push({ label: 'Final', value: formatMoney(j.final_amount), icon: 'pricetag-outline' })
  if (paid > 0) detailRows.push({ label: 'Paid', value: formatMoney(paid), icon: 'wallet-outline' })

  return (
    <View style={styles.flex}>
      <ScrollView
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl refreshing={job.isRefetching} onRefresh={() => job.refetch()} tintColor={colors.primary} />
        }
      >
        {/* Hero */}
        <Card style={styles.hero}>
          <View style={styles.heroTop}>
            <Avatar name={j.customer?.name} size={52} />
            <View style={styles.heroMain}>
              <Text style={styles.customer}>{j.customer?.name ?? 'Customer'}</Text>
              <View style={styles.customerRow}>
                <Text style={styles.phone}>{j.customer?.phone ?? ''}</Text>
                <Pill label={j.job_number} color={colors.primary} />
              </View>
            </View>
            <IconButton name="call" color={colors.primary} bg={colors.primarySoft} onPress={callCustomer} />
          </View>

          <Divider />

          <View style={styles.serviceRow}>
            <Text style={styles.service}>{j.service_type}</Text>
            <View style={styles.badgeWrap}>
              <Badge value={status} />
              {j.payment_status ? <Pill label={j.payment_status} color={payColor} /> : null}
            </View>
          </View>

          {j.notes ? (
            <Text style={styles.notes}>{j.notes}</Text>
          ) : null}
        </Card>

        {/* Action buttons */}
        <View style={styles.actionArea}>
          {actionLabel ? (
            <Button
              title={actionLabel}
              size="lg"
              icon={actionIcon}
              full
              onPress={() =>
                confirmAnd(actionLabel, () => action.mutate())
              }
              loading={action.isPending}
              disabled={action.isPending}
            />
          ) : status === 'STARTED' ? (
            <Button
              title="Work details + Complete"
              size="lg"
              icon="create-outline"
              full
              onPress={() => navigation.navigate('WorkDetails', { id })}
            />
          ) : null}

          {action.isError ? (
            <View style={styles.actionError}>
              <Icon name="alert-circle-outline" size={15} color={colors.danger} />
              <Text style={styles.actionErrorText}>{apiErrorMessage(action.error)}</Text>
            </View>
          ) : null}

          <View style={styles.quickRow}>
            <Button title="Call" variant="outline" icon="call" flex1 onPress={callCustomer} disabled={!j.customer?.phone} />
            <Button title="Directions" variant="outline" icon="navigate-outline" flex1 onPress={openMaps} />
            <Button
              title="Payment"
              variant={['PENDING', 'ACCEPTED', 'ON_THE_WAY'].includes(status) ? 'outline' : 'soft'}
              icon="card-outline"
              flex1
              onPress={() => navigation.navigate('Payment', { id })}
            />
          </View>
        </View>

        {/* Details */}
        <Card>
          <SectionTitle style={styles.cardTitle}>Job details</SectionTitle>
          {detailRows.map((row) => (
            <View key={row.label} style={styles.detailRow}>
              <View style={styles.detailLabelWrap}>
                <Icon name={row.icon ?? 'information-circle-outline'} size={15} color={colors.muted} />
                <Muted>{row.label}</Muted>
              </View>
              <Text style={styles.detailValue}>{row.value}</Text>
            </View>
          ))}
        </Card>

        {/* Progress timeline */}
        <Card>
          <SectionTitle style={styles.cardTitle}>Progress</SectionTitle>
          <StatusTimeline job={j} />
        </Card>

        {/* Items */}
        {j.items && j.items.length > 0 ? (
          <Card>
            <SectionTitle style={styles.cardTitle}>Charges</SectionTitle>
            {j.items.map((item) => (
              <View key={item.id} style={styles.detailRow}>
                <Muted style={styles.itemDesc}>
                  {item.description} × {formatMoney(item.unit_price)}{item.quantity !== 1 ? ` (${item.quantity})` : ''}
                </Muted>
                <Text style={styles.detailValue}>{formatMoney(item.total_price)}</Text>
              </View>
            ))}
            <Divider />
            <View style={styles.totalRow}>
              <Text style={styles.totalLabel}>Total</Text>
              <Text style={styles.totalValue}>{formatMoney(j.final_amount || amount)}</Text>
            </View>
          </Card>
        ) : null}

        {/* Payments */}
        {j.payments && j.payments.length > 0 ? (
          <Card>
            <View style={styles.cardsHeader}>
              <SectionTitle style={styles.cardTitle}>Payments</SectionTitle>
              <Text style={styles.paidTotal}>{formatMoney(paid)}</Text>
            </View>
            {j.payments.map((p) => (
              <View key={p.id} style={styles.payRow}>
                <View style={styles.payLeft}>
                  <View style={[styles.methodIcon, { backgroundColor: methodColor[p.method] ?? colors.muted }]}>
                    <Text style={styles.methodIconText}>{p.method.slice(0, 1)}</Text>
                  </View>
                  <View>
                    <Text style={styles.payMethod}>{p.method.replace(/_/g, ' ')}</Text>
                    <Muted>{formatDateTime(p.paid_at)}</Muted>
                  </View>
                </View>
                <Text style={styles.payAmount}>{formatMoney(p.amount)}</Text>
              </View>
            ))}
          </Card>
        ) : null}

        {/* Receipt */}
        {status === 'COMPLETED' && j.receipt ? (
          <Card>
            <View style={styles.receiptHead}>
              <View style={styles.receiptIcon}>
                <Icon name="receipt-outline" size={20} color={colors.success} />
              </View>
              <View style={styles.flex1}>
                <SectionTitle style={styles.cardTitle}>Receipt issued</SectionTitle>
                <Muted>{j.receipt.receipt_number} · {formatDateTime(j.receipt.issued_at)}</Muted>
              </View>
            </View>
            {receipt.data ? (
              <Button
                title="Open public receipt"
                variant="soft"
                icon="open-outline"
                full
                style={{ marginTop: spacing.md }}
                onPress={() => Linking.openURL(API_BASE + receipt.data).catch(() => {})}
              />
            ) : null}
          </Card>
        ) : null}

        {remaining > 0 && status === 'COMPLETED' ? (
          <Button
            title={`Collect balance ${formatMoney(remaining)}`}
            variant="soft"
            icon="card-outline"
            size="lg"
            full
            onPress={() => navigation.navigate('Payment', { id })}
          />
        ) : null}
      </ScrollView>
    </View>
  )
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl, gap: spacing.md },
  hero: { padding: spacing.lg },
  heroTop: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  heroMain: { flex: 1 },
  customer: { fontSize: 20, fontWeight: '800', color: colors.text, letterSpacing: -0.3 },
  customerRow: { flexDirection: 'row', alignItems: 'center', gap: 8, marginTop: 3 },
  phone: { color: colors.muted, fontSize: 13 },
  serviceRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: spacing.sm,
  },
  service: { flex: 1, fontSize: 17, fontWeight: '800', color: colors.text },
  badgeWrap: { flexDirection: 'row', alignItems: 'center', gap: 6, flexShrink: 1 },
  notes: { color: colors.textSoft, fontSize: 13.5, marginTop: spacing.sm, lineHeight: 19 },
  actionArea: { gap: spacing.md },
  quickRow: { flexDirection: 'row', gap: spacing.sm },
  actionError: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: colors.dangerSoft,
    borderRadius: radius.md,
    padding: spacing.md,
  },
  actionErrorText: { color: colors.danger, fontSize: 13, fontWeight: '600', flex: 1 },
  cardTitle: { marginBottom: spacing.md },
  detailRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    gap: spacing.md,
    paddingVertical: 7,
  },
  detailLabelWrap: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  detailValue: { color: colors.text, fontWeight: '700', flexShrink: 1, textAlign: 'right', fontSize: 14 },
  itemDesc: { flex: 1, flexShrink: 1 },
  totalRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  totalLabel: { fontWeight: '800', color: colors.text, fontSize: 15 },
  totalValue: { fontWeight: '800', color: colors.text, fontSize: 17 },
  cardsHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  paidTotal: { fontWeight: '800', color: colors.success, fontSize: 16 },
  payRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.sm,
  },
  payLeft: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  methodIcon: {
    width: 30,
    height: 30,
    borderRadius: radius.sm,
    alignItems: 'center',
    justifyContent: 'center',
  },
  methodIconText: { color: colors.white, fontWeight: '800', fontSize: 12 },
  payMethod: { fontSize: 13.5, fontWeight: '700', color: colors.text },
  payAmount: { fontWeight: '800', color: colors.text },
  receiptHead: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  receiptIcon: {
    width: 44,
    height: 44,
    borderRadius: radius.md,
    backgroundColor: colors.successSoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  flex1: { flex: 1 },
})