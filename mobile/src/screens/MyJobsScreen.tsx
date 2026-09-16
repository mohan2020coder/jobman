import { useQuery, keepPreviousData } from '@tanstack/react-query'
import { useState } from 'react'
import { FlatList, RefreshControl, StyleSheet, View } from 'react-native'
import type { NativeStackScreenProps } from '@react-navigation/native-stack'
import type { RootStackParamList } from '../navigation/types'
import { myJobsRequest } from '../api/client'
import { JobCard } from '../components/JobCard'
import { Chip, EmptyState, ErrorView, Icon, Input, LoadingView } from '../components/ui'
import { colors, spacing } from '../theme'

type Props = NativeStackScreenProps<RootStackParamList, 'MyJobs'>
type Filter = '' | 'today' | 'upcoming' | 'completed'

const tabs: { key: Filter; label: string; icon: 'grid-outline' | 'today-outline' | 'calendar-outline' | 'checkmark-done-outline' }[] = [
  { key: '', label: 'All', icon: 'grid-outline' },
  { key: 'today', label: 'Today', icon: 'today-outline' },
  { key: 'upcoming', label: 'Upcoming', icon: 'calendar-outline' },
  { key: 'completed', label: 'Completed', icon: 'checkmark-done-outline' },
]

export function MyJobsScreen({ navigation, route }: Props) {
  const [filter, setFilter] = useState<Filter>(route.params?.initialFilter ?? '')
  const [search, setSearch] = useState('')

  const jobs = useQuery({
    queryKey: ['techjobs', filter],
    queryFn: () => myJobsRequest(filter),
    placeholderData: keepPreviousData,
    staleTime: 15_000,
  })

  function selectTab(key: Filter) {
    setFilter(key)
    setSearch('')
  }

  const q = search.trim().toLowerCase()
  const data =
    jobs.data && q
      ? jobs.data.filter(
          (j) =>
            j.customer?.name?.toLowerCase().includes(q) ||
            j.job_number.toLowerCase().includes(q) ||
            j.service_type.toLowerCase().includes(q) ||
            j.problem_description?.toLowerCase().includes(q) ||
            j.address?.toLowerCase().includes(q),
        )
      : jobs.data

  return (
    <View style={styles.flex}>
      <FlatList
        data={data ?? []}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.content}
        showsVerticalScrollIndicator={false}
        keyboardShouldPersistTaps="handled"
        refreshControl={
          <RefreshControl
            refreshing={jobs.isRefetching}
            onRefresh={() => jobs.refetch()}
            tintColor={colors.primary}
          />
        }
        ListHeaderComponent={
          <View>
            <View style={styles.searchWrap}>
              <Icon name="search" size={18} color={colors.muted} />
              <Input
                style={styles.searchInput}
                placeholder="Search customer, job no, service…"
                value={search}
                onChangeText={setSearch}
                autoCapitalize="none"
                autoCorrect={false}
              />
              {search ? (
                <Icon name="close-circle" size={18} color={colors.muted} />
              ) : null}
            </View>
            <View style={styles.tabs}>
              {tabs.map((t) => (
                <Chip
                  key={t.key}
                  label={t.label}
                  icon={t.icon}
                  active={t.key === filter}
                  onPress={() => selectTab(t.key)}
                />
              ))}
            </View>
          </View>
        }
        renderItem={({ item }) => (
          <JobCard job={item} onPress={() => navigation.navigate('JobDetail', { id: item.id })} />
        )}
        ListEmptyComponent={
          jobs.isLoading ? (
            <LoadingView label="Loading jobs…" />
          ) : jobs.isError ? (
            <ErrorView message="Could not load jobs." onRetry={() => jobs.refetch()} />
          ) : q ? (
            <EmptyState
              icon="search-outline"
              title="No matches"
              message={`Nothing found for "${search.trim()}" in ${filterName(filter)}.`}
            />
          ) : (
            <EmptyState
              icon="folder-open-outline"
              title="No jobs here"
              message={`You have no ${filterName(filter).toLowerCase()} jobs right now.`}
            />
          )
        }
      />
    </View>
  )
}

function filterName(f: Filter): string {
  switch (f) {
    case 'today':
      return 'Today'
    case 'upcoming':
      return 'Upcoming'
    case 'completed':
      return 'Completed'
    default:
      return 'All'
  }
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl },
  searchWrap: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: 14,
    paddingHorizontal: spacing.md,
  },
  searchInput: { borderWidth: 0, backgroundColor: 'transparent', paddingVertical: 13, flex: 1 },
  tabs: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    gap: 8,
    marginTop: spacing.md,
    marginBottom: spacing.lg,
  },
})