import { useQuery } from '@tanstack/react-query'
import { Alert, ScrollView, StyleSheet, Text, View } from 'react-native'
import { SafeAreaView } from 'react-native-safe-area-context'
import { useAuth } from '../auth/AuthContext'
import { myJobsRequest } from '../api/client'
import {
  Avatar,
  Button,
  Card,
  Divider,
  Icon,
  Muted,
  SectionTitle,
  Spinner,
} from '../components/ui'
import { formatMoney } from '../lib/format'
import { colors, radius, spacing, tint } from '../theme'

export function ProfileScreen() {
  const { user, businessName, signOut } = useAuth()

  const completed = useQuery({
    queryKey: ['techjobs', 'completed'],
    queryFn: () => myJobsRequest('completed'),
    staleTime: 30_000,
  })

  const today = useQuery({
    queryKey: ['techjobs', 'today'],
    queryFn: () => myJobsRequest('today'),
    staleTime: 30_000,
  })

  const done = completed.data ?? []
  const doneToday = (today.data ?? []).filter((j) => j.status === 'COMPLETED')
  const totalBilled = done.reduce((s, j) => s + (j.final_amount || j.estimated_amount), 0)
  const totalCollected = done.reduce((s, j) => s + (j.total_paid ?? 0), 0)
  const todayBilled = doneToday.reduce((s, j) => s + (j.final_amount || j.estimated_amount), 0)

  function confirmSignOut() {
    Alert.alert('Sign out?', 'You will need to sign in again to see your jobs.', [
      { text: 'Cancel', style: 'cancel' },
      { text: 'Sign out', style: 'destructive', onPress: () => signOut() },
    ])
  }

  return (
    <SafeAreaView style={styles.safe} edges={['left', 'right']}>
      <ScrollView
        style={styles.flex}
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.header}>
          <Avatar name={user?.name} size={72} />
          <Text style={styles.name}>{user?.name ?? 'Technician'}</Text>
          <Text style={styles.business}>{businessName ?? ''}</Text>
          <View style={styles.rolePill}>
            <Text style={styles.rolePillText}>{user?.role ?? ''}</Text>
          </View>
        </View>

        <View style={styles.stats}>
          <Card style={styles.statCard}>
            <Text style={styles.statValue}>{done.length}</Text>
            <Text style={styles.statLabel}>Jobs completed</Text>
          </Card>
          <Card style={styles.statCard}>
            <Text style={styles.statValue}>{formatMoney(totalBilled)}</Text>
            <Text style={styles.statLabel}>Total billed</Text>
          </Card>
          <Card style={styles.statCard}>
            <Text style={styles.statValue}>{formatMoney(totalCollected)}</Text>
            <Text style={styles.statLabel}>Collected</Text>
          </Card>
        </View>

        <Card style={{ marginTop: spacing.lg }}>
          <SectionTitle style={styles.cardTitle}>Account</SectionTitle>
          <View style={styles.infoRow}>
            <View style={styles.infoLabel}>
              <Icon name="business-outline" size={16} color={colors.muted} />
              <Muted>Business</Muted>
            </View>
            <Text style={styles.infoValue}>{businessName ?? '—'}</Text>
          </View>
          <Divider />
          <View style={styles.infoRow}>
            <View style={styles.infoLabel}>
              <Icon name="person-outline" size={16} color={colors.muted} />
              <Muted>Role</Muted>
            </View>
            <Text style={styles.infoValue}>{user?.role === 'TECHNICIAN' ? 'Technician' : user?.role}</Text>
          </View>
          <Divider />
          <View style={styles.infoRow}>
            <View style={styles.infoLabel}>
              <Icon name="trophy-outline" size={16} color={colors.muted} />
              <Muted>Today billed</Muted>
            </View>
            <Text style={styles.infoValue}>{formatMoney(todayBilled)}</Text>
          </View>
        </Card>

        {completed.isLoading ? (
          <Spinner />
        ) : (
          <View style={styles.insight}>
            <Icon name="sparkles-outline" size={18} color={colors.accent} />
            <Text style={styles.insightText}>
              {done.length === 0
                ? 'Complete your first job to start building your track record.'
                : `${done.length} job${done.length === 1 ? '' : 's'} completed · ${formatMoney(totalCollected)} collected in total.`}
            </Text>
          </View>
        )}

        <Muted style={styles.version}>Jobman Technician · v1.0</Muted>

        <Button
          title="Sign out"
          variant="danger"
          size="lg"
          icon="log-out-outline"
          full
          onPress={confirmSignOut}
        />
      </ScrollView>
    </SafeAreaView>
  )
}

const styles = StyleSheet.create({
  flex: { flex: 1 },
  safe: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl },
  header: { alignItems: 'center', paddingVertical: spacing.lg },
  name: { fontSize: 22, fontWeight: '800', color: colors.text, marginTop: spacing.md, letterSpacing: -0.4 },
  business: { color: colors.muted, fontSize: 14, marginTop: 2 },
  rolePill: {
    marginTop: spacing.md,
    backgroundColor: tint(colors.primary, 0.1),
    borderRadius: radius.pill,
    paddingHorizontal: 12,
    paddingVertical: 5,
  },
  rolePillText: {
    color: colors.primaryDark,
    fontSize: 12,
    fontWeight: '800',
    textTransform: 'uppercase',
    letterSpacing: 0.6,
  },
  stats: { flexDirection: 'row', gap: spacing.sm },
  statCard: { flex: 1, alignItems: 'center', paddingVertical: spacing.lg },
  statValue: { fontSize: 18, fontWeight: '800', color: colors.text },
  statLabel: { fontSize: 11, color: colors.muted, marginTop: 4, fontWeight: '600', textAlign: 'center' },
  cardTitle: { marginBottom: spacing.sm },
  infoRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.sm + 2,
  },
  infoLabel: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  infoValue: { color: colors.text, fontWeight: '700', flexShrink: 1, textAlign: 'right' },
  insight: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    backgroundColor: tint(colors.accent, 0.08),
    borderRadius: radius.md,
    padding: spacing.lg,
    marginVertical: spacing.lg,
  },
  insightText: { color: colors.textSoft, fontSize: 13, flex: 1, lineHeight: 19 },
  version: { textAlign: 'center', marginBottom: spacing.lg },
})