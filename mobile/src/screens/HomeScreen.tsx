import { useQuery } from '@tanstack/react-query'
import { Pressable, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native'
import type { NativeStackScreenProps } from '@react-navigation/native-stack'
import type { RootStackParamList } from '../navigation/types'
import { useAuth } from '../auth/AuthContext'
import { myJobsRequest } from '../api/client'
import { JobCard } from '../components/JobCard'
import {
  Avatar,
  Button,
  EmptyState,
  ErrorView,
  Icon,
  SectionHeader,
  Spinner,
  StatCard,
} from '../components/ui'
import { greeting, formatMoney, todayLabel } from '../lib/format'
import { colors, spacing } from '../theme'

type Props = NativeStackScreenProps<RootStackParamList, 'Home'>

const ACTIVE = ['PENDING', 'ACCEPTED', 'ON_THE_WAY']

export function HomeScreen({ navigation }: Props) {
  const { user } = useAuth()

  const today = useQuery({
    queryKey: ['techjobs', 'today'],
    queryFn: () => myJobsRequest('today'),
    staleTime: 30_000,
  })

  const upcoming = useQuery({
    queryKey: ['techjobs', 'upcoming'],
    queryFn: () => myJobsRequest('upcoming'),
    staleTime: 60_000,
  })

  const jobs = today.data ?? []
  const pending = jobs.filter((j) => ACTIVE.includes(j.status)).length
  const inProgress = jobs.filter((j) => j.status === 'STARTED').length
  const completed = jobs.filter((j) => j.status === 'COMPLETED').length
  const todayTotal = jobs.reduce(
    (sum, j) => sum + (j.status === 'COMPLETED' ? j.final_amount || j.estimated_amount : 0),
    0,
  )

  const refreshing = today.isRefetching || upcoming.isRefetching

  function refresh() {
    today.refetch()
    upcoming.refetch()
  }

  return (
    <View style={styles.flex}>
      <View style={styles.hero}>
        <View style={styles.heroTop}>
          <View style={styles.heroIdentity}>
            <Pressable onPress={() => navigation.navigate('Profile')} hitSlop={8}>
              <Avatar name={user?.name} size={46} />
            </Pressable>
            <View>
              <Text style={styles.greeting}>{greeting()},</Text>
              <Text style={styles.userName}>{user?.name ?? 'Technician'}</Text>
            </View>
          </View>
          <Button
            title="My Jobs"
            size="sm"
            variant="soft"
            icon="list"
            onPress={() => navigation.navigate('MyJobs', { initialFilter: 'today' })}
          />
        </View>
        <View style={styles.heroDate}>
          <Icon name="calendar-outline" size={14} color="rgba(255,255,255,0.85)" />
          <Text style={styles.heroDateText}>{todayLabel()}</Text>
        </View>
      </View>

      <ScrollView
        style={styles.scroll}
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
        refreshControl={<RefreshControl refreshing={refreshing} onRefresh={refresh} tintColor={colors.primary} />}
      >
        {today.isError && !jobs.length ? (
          <ErrorView message="Could not load your jobs." onRetry={() => today.refetch()} />
        ) : (
          <>
            <View style={styles.stats}>
              <StatCard label="Today's Jobs" value={jobs.length} icon="calendar" accent={colors.primary} />
              <StatCard label="Pending" value={pending} icon="hourglass-outline" accent={colors.warning} />
              <StatCard label="In Progress" value={inProgress} icon="construct-outline" accent={colors.accent} />
              <StatCard label="Completed" value={completed} icon="checkmark-done-outline" accent={colors.success} />
            </View>

            <SectionHeader
              title="Today's jobs"
              action="See all"
              onAction={() => navigation.navigate('MyJobs', { initialFilter: 'today' })}
            />

            {today.isLoading ? (
              <Spinner label="Loading today's jobs…" />
            ) : jobs.length === 0 ? (
              <EmptyState
                icon="sunny-outline"
                title="No jobs today"
                message="When a job is scheduled for today it will show up here."
              />
            ) : (
              jobs.slice(0, 8).map((job) => (
                <JobCard
                  key={job.id}
                  job={job}
                  onPress={() => navigation.navigate('JobDetail', { id: job.id })}
                />
              ))
            )}

            <SectionHeader
              title="Upcoming"
              action={upcoming.data && upcoming.data.length > 0 ? 'See all' : undefined}
              onAction={() => navigation.navigate('MyJobs', { initialFilter: 'upcoming' })}
            />

            {upcoming.isLoading ? (
              <Spinner label="Loading upcoming…" />
            ) : !upcoming.data || upcoming.data.length === 0 ? (
              <View style={styles.upcomingEmpty}>
                <Icon name="calendar-outline" size={18} color={colors.muted} />
                <Text style={styles.upcomingEmptyText}>Nothing scheduled ahead.</Text>
              </View>
            ) : (
              upcoming.data.slice(0, 3).map((job) => (
                <JobCard
                  key={job.id}
                  job={job}
                  onPress={() => navigation.navigate('JobDetail', { id: job.id })}
                />
              ))
            )}

            {todayTotal > 0 ? (
              <View style={styles.earned}>
                <Text style={styles.earnedLabel}>Billed today</Text>
                <Text style={styles.earnedValue}>{formatMoney(todayTotal)}</Text>
                <Text style={styles.earnedTotal}>across {completed} completed job{completed === 1 ? '' : 's'}</Text>
              </View>
            ) : null}
          </>
        )}
      </ScrollView>
    </View>
  )
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  hero: {
    backgroundColor: colors.primary,
    paddingTop: spacing.xl,
    paddingHorizontal: spacing.lg,
    paddingBottom: spacing.xl,
    borderBottomLeftRadius: 22,
    borderBottomRightRadius: 22,
  },
  heroTop: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  heroIdentity: { flexDirection: 'row', alignItems: 'center', gap: 12 },
  greeting: { color: 'rgba(255,255,255,0.8)', fontSize: 13, fontWeight: '600' },
  userName: { color: colors.white, fontSize: 19, fontWeight: '800', marginTop: 1, letterSpacing: -0.3 },
  heroDate: { flexDirection: 'row', alignItems: 'center', gap: 6, marginTop: 14 },
  heroDateText: { color: 'rgba(255,255,255,0.9)', fontSize: 13, fontWeight: '600' },
  scroll: { flex: 1 },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl, marginTop: 0 },
  stats: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: spacing.md,
    marginBottom: spacing.xl,
    marginTop: spacing.lg + 4,
  },
  upcomingEmpty: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 14,
    padding: spacing.lg,
  },
  upcomingEmptyText: { color: colors.muted, fontSize: 13 },
  earned: {
    alignItems: 'center',
    marginTop: spacing.xl,
    padding: spacing.xl,
    backgroundColor: colors.successSoft,
    borderRadius: 16,
  },
  earnedLabel: { fontSize: 12, fontWeight: '700', color: colors.success, textTransform: 'uppercase', letterSpacing: 0.5 },
  earnedValue: { fontSize: 26, fontWeight: '800', color: colors.text, marginTop: 4 },
  earnedTotal: { fontSize: 13, color: colors.muted, marginTop: 2 },
})