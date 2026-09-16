export type RootStackParamList = {
  Login: undefined
  Home: undefined
  MyJobs: { initialFilter?: 'today' | 'upcoming' | 'completed' } | undefined
  JobDetail: { id: string }
  WorkDetails: { id: string }
  Payment: { id: string }
  Profile: undefined
}