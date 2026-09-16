import { NavigationContainer } from '@react-navigation/native'
import { createNativeStackNavigator } from '@react-navigation/native-stack'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { StatusBar } from 'expo-status-bar'
import { SafeAreaProvider } from 'react-native-safe-area-context'
import { AuthProvider, useAuth } from './src/auth/AuthContext'
import type { RootStackParamList } from './src/navigation/types'
import { navigationRef } from './src/navigation/ref'
import { RealtimeProvider } from './src/realtime/RealtimeContext'
import { HomeScreen } from './src/screens/HomeScreen'
import { JobDetailScreen } from './src/screens/JobDetailScreen'
import { LoginScreen } from './src/screens/LoginScreen'
import { MyJobsScreen } from './src/screens/MyJobsScreen'
import { PaymentScreen } from './src/screens/PaymentScreen'
import { ProfileScreen } from './src/screens/ProfileScreen'
import { WorkDetailsScreen } from './src/screens/WorkDetailsScreen'
import { Spinner, SafeScreen } from './src/components/ui'
import { colors } from './src/theme'

const queryClient = new QueryClient()
const Stack = createNativeStackNavigator<RootStackParamList>()

function AppNavigation() {
  const { user, loading } = useAuth()

  if (loading) {
    return (
      <SafeScreen>
        <Spinner label="Loading your account…" />
      </SafeScreen>
    )
  }

  return (
    <Stack.Navigator
      screenOptions={{
        headerStyle: { backgroundColor: colors.bg },
        headerTintColor: colors.text,
        headerTitleStyle: { fontWeight: '700' },
        contentStyle: { backgroundColor: colors.bg },
      }}
    >
      {user ? (
        <>
          <Stack.Screen name="Home" component={HomeScreen} options={{ headerShown: false }} />
          <Stack.Screen
            name="MyJobs"
            component={MyJobsScreen}
            options={{ title: 'My Jobs', headerBackTitle: 'Back' }}
          />
          <Stack.Screen
            name="JobDetail"
            component={JobDetailScreen}
            options={{ title: 'Job', headerBackTitle: 'Back' }}
          />
          <Stack.Screen
            name="WorkDetails"
            component={WorkDetailsScreen}
            options={{ title: 'Work Details', headerBackTitle: 'Back' }}
          />
          <Stack.Screen
            name="Payment"
            component={PaymentScreen}
            options={{ title: 'Payment', headerBackTitle: 'Back' }}
          />
          <Stack.Screen name="Profile" component={ProfileScreen} options={{ title: 'Profile' }} />
        </>
      ) : (
        <Stack.Screen name="Login" component={LoginScreen} options={{ headerShown: false }} />
      )}
    </Stack.Navigator>
  )
}

export default function App() {
  return (
    <SafeAreaProvider>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <StatusBar style="dark" />
          <NavigationContainer ref={navigationRef}>
            <RealtimeProvider>
              <AppNavigation />
            </RealtimeProvider>
          </NavigationContainer>
        </AuthProvider>
      </QueryClientProvider>
    </SafeAreaProvider>
  )
}