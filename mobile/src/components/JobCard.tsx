import { Linking, Pressable, StyleSheet, Text, View } from 'react-native'
import type { Job } from '../api/types'
import { colors, radius, spacing, statusColor } from '../theme'
import { formatMoney, formatSchedule } from '../lib/format'
import { Avatar, Badge, Icon, Pill } from './ui'

export function JobCard({ job, onPress }: { job: Job; onPress: () => void }) {
  const amount = job.final_amount || job.estimated_amount
  const payColor = statusColor[job.payment_status as keyof typeof statusColor]

  return (
    <Pressable
      onPress={onPress}
      style={({ pressed }) => [styles.card, pressed && styles.pressed]}
    >
      <View style={styles.header}>
        <Avatar name={job.customer?.name} />
        <View style={styles.headerMain}>
          <Text style={styles.customer} numberOfLines={1}>
            {job.customer?.name ?? 'Customer'}
          </Text>
          <Text style={styles.service} numberOfLines={1}>
            {job.service_type}
          </Text>
        </View>
        <Badge value={job.status} />
      </View>

      {job.problem_description ? (
        <Text style={styles.problem} numberOfLines={2}>
          {job.problem_description}
        </Text>
      ) : null}

      {job.address ? (
        <View style={styles.metaRow}>
          <Icon name="location-outline" size={14} color={colors.muted} />
          <Text style={styles.meta} numberOfLines={1}>
            {job.address}
          </Text>
        </View>
      ) : null}
      <View style={styles.metaRow}>
        <Icon name="time-outline" size={14} color={colors.muted} />
        <Text style={styles.meta}>
          {job.job_number} · {formatSchedule(job.scheduled_at)}
        </Text>
      </View>

      <View style={styles.footer}>
        <View style={styles.amountGroup}>
          <Text style={styles.amount}>{formatMoney(amount)}</Text>
          {job.payment_status && !['PENDING', 'ACCEPTED', 'ON_THE_WAY', 'STARTED'].includes(job.status) ? (
            <Pill label={job.payment_status} color={payColor} />
          ) : null}
        </View>
        <View style={styles.actions}>
          {job.customer?.phone ? (
            <Pressable
              style={styles.iconAction}
              hitSlop={6}
              onPress={(e) => {
                e.stopPropagation()
                Linking.openURL(`tel:${job.customer?.phone}`).catch(() => {})
              }}
            >
              <Icon name="call" size={17} color={colors.primary} />
            </Pressable>
          ) : null}
          <View style={styles.viewChip}>
            <Text style={styles.viewChipText}>View job</Text>
            <Icon name="chevron-forward" size={14} color={colors.primary} />
          </View>
        </View>
      </View>
    </Pressable>
  )
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.lg,
    padding: spacing.lg,
    marginBottom: spacing.md,
    gap: spacing.sm + 2,
  },
  pressed: { opacity: 0.85, transform: [{ scale: 0.992 }] },
  header: { flexDirection: 'row', alignItems: 'center', gap: spacing.md },
  headerMain: { flex: 1 },
  customer: { fontSize: 16, fontWeight: '800', color: colors.text },
  service: { fontSize: 13, color: colors.muted, marginTop: 1 },
  problem: { fontSize: 13.5, color: colors.text, lineHeight: 19 },
  metaRow: { flexDirection: 'row', alignItems: 'center', gap: 6 },
  meta: { fontSize: 12.5, color: colors.muted, flex: 1 },
  footer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 2,
  },
  amountGroup: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  amount: { fontSize: 17, fontWeight: '800', color: colors.text, letterSpacing: -0.3 },
  actions: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  iconAction: {
    width: 34,
    height: 34,
    borderRadius: radius.sm,
    backgroundColor: colors.primarySoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
  viewChip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 2,
    backgroundColor: colors.primarySoft,
    borderRadius: radius.sm,
    paddingHorizontal: 10,
    paddingVertical: 8,
  },
  viewChipText: { fontSize: 12.5, fontWeight: '700', color: colors.primary },
})