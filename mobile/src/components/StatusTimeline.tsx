import { StyleSheet, Text, View } from 'react-native'
import type { Job } from '../api/types'
import { colors, radius, spacing } from '../theme'
import { formatDateTime } from '../lib/format'
import { Icon } from './ui'

const ORDER = ['PENDING', 'ACCEPTED', 'ON_THE_WAY', 'STARTED', 'COMPLETED'] as const

export function StatusTimeline({ job }: { job: Job }) {
  const currentIdx = ORDER.indexOf(job.status as (typeof ORDER)[number])
  const cancelled = job.status === 'CANCELLED'
  const completed = job.status === 'COMPLETED'

  const pillLabel: Record<(typeof ORDER)[number] | 'CANCELLED', string> = {
    PENDING: 'Pending',
    ACCEPTED: 'Accepted',
    ON_THE_WAY: 'On the way',
    STARTED: 'In progress',
    COMPLETED: 'Done',
    CANCELLED: 'Cancelled',
  }

  const stages: { key: (typeof ORDER)[number]; label: string; time: string | null | undefined }[] = [
    { key: 'PENDING', label: 'Job created', time: job.accepted_at ?? job.updated_at ?? job.created_at },
    { key: 'ACCEPTED', label: 'Accepted', time: job.accepted_at },
    { key: 'ON_THE_WAY', label: 'On the way', time: job.on_the_way_at },
    { key: 'STARTED', label: 'Work started', time: job.started_at },
    { key: 'COMPLETED', label: 'Completed', time: job.completed_at },
  ]

  return (
    <View style={styles.wrap}>
      {stages.map((stage, i) => {
        const done = !cancelled && i <= currentIdx
        const isCurrent = !cancelled && i === currentIdx
        const isLast = i === stages.length - 1
        return (
          <View key={stage.key} style={styles.stage}>
            <View style={styles.markerCol}>
              <View
                style={[
                  styles.marker,
                  done && styles.markerDone,
                  isCurrent && styles.markerCurrent,
                ]}
              >
                {done && !isCurrent ? <Icon name="checkmark" size={13} color={colors.white} /> : null}
                {isCurrent ? <View style={styles.markerInner} /> : null}
              </View>
              {!isLast ? <View style={styles.line} /> : null}
            </View>
            <View style={styles.body}>
              <Text style={[styles.label, done && !isCurrent && styles.labelDone]}>{stage.label}</Text>
              {isCurrent && !completed ? (
                <View style={styles.currentPill}>
                  <Text style={styles.currentPillText}>{pillLabel[job.status as keyof typeof pillLabel]}</Text>
                </View>
              ) : null}
              {done && stage.time ? (
                <Text style={styles.time}>{formatDateTime(stage.time)}</Text>
              ) : null}
            </View>
          </View>
        )
      })}
      {cancelled ? (
        <View style={styles.stage}>
          <View style={styles.markerCol}>
            <View style={[styles.marker, styles.markerCancelled]}>
              <Icon name="close" size={13} color={colors.white} />
            </View>
          </View>
          <View style={styles.body}>
            <Text style={[styles.label, styles.labelCancelled]}>Cancelled</Text>
            <Text style={styles.time}>{formatDateTime(job.updated_at)}</Text>
          </View>
        </View>
      ) : null}
    </View>
  )
}

const styles = StyleSheet.create({
  wrap: { gap: 0 },
  stage: { flexDirection: 'row' },
  markerCol: { alignItems: 'center', width: 24, marginRight: spacing.md },
  marker: {
    width: 22,
    height: 22,
    borderRadius: 11,
    borderWidth: 2,
    borderColor: colors.borderStrong,
    backgroundColor: colors.surface,
    alignItems: 'center',
    justifyContent: 'center',
  },
  markerDone: {
    backgroundColor: colors.success,
    borderColor: colors.success,
  },
  markerCurrent: {
    borderColor: colors.primary,
    backgroundColor: colors.surface,
  },
  markerInner: {
    width: 8,
    height: 8,
    borderRadius: 4,
    backgroundColor: colors.primary,
  },
  markerCancelled: {
    backgroundColor: colors.danger,
    borderColor: colors.danger,
  },
  line: {
    flex: 1,
    width: 2,
    minHeight: 26,
    backgroundColor: colors.border,
  },
  body: { flex: 1, paddingBottom: spacing.xl, gap: 2 },
  label: { fontSize: 14, fontWeight: '700', color: colors.text },
  labelDone: { color: colors.faint },
  labelCancelled: { color: colors.danger },
  time: { fontSize: 12, color: colors.muted },
  currentPill: {
    alignSelf: 'flex-start',
    backgroundColor: colors.primarySoft,
    borderRadius: radius.pill,
    paddingHorizontal: 8,
    paddingVertical: 2,
    marginTop: 2,
  },
  currentPillText: { fontSize: 10, fontWeight: '700', color: colors.primaryDark, textTransform: 'uppercase', letterSpacing: 0.4 },
})