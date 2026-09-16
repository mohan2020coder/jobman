import { useState } from 'react'
import {
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native'
import { SafeAreaView } from 'react-native-safe-area-context'
import type { NativeStackScreenProps } from '@react-navigation/native-stack'
import type { RootStackParamList } from '../navigation/types'
import { useAuth } from '../auth/AuthContext'
import { apiErrorMessage } from '../api/client'
import { Button, Card, Icon, Input } from '../components/ui'
import { colors, radius, shadows, spacing } from '../theme'

type Props = NativeStackScreenProps<RootStackParamList, 'Login'>

export function LoginScreen({}: Props) {
  const { signIn } = useAuth()
  const [phone, setPhone] = useState('9876543211')
  const [password, setPassword] = useState('password123')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  async function handleLogin() {
    setError('')
    setBusy(true)
    try {
      await signIn(phone.trim(), password)
    } catch (err) {
      setError(apiErrorMessage(err))
    } finally {
      setBusy(false)
    }
  }

  return (
    <KeyboardAvoidingView
      style={styles.flex}
      behavior={Platform.OS === 'ios' ? 'padding' : undefined}
    >
      <SafeAreaView style={styles.flex} edges={['top', 'bottom', 'left', 'right']}>
        <ScrollView
          contentContainerStyle={styles.wrap}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          <View style={styles.brand}>
            <View style={styles.logo}>
              <Icon name="construct" size={36} color={colors.white} />
            </View>
            <Text style={styles.appName}>Jobman Technician</Text>
            <Text style={styles.tagline}>Accept jobs, update status, collect payment — all in one place.</Text>
          </View>

          <Card style={styles.card}>
            <Text style={styles.cardTitle}>Sign in</Text>
            <Text style={styles.cardSub}>Use your technician phone number</Text>

            <Text style={styles.fieldLabel}>Phone number</Text>
            <View style={styles.inputWrap}>
              <Icon name="phone-portrait-outline" size={18} color={colors.muted} />
              <Input
                style={styles.input}
                value={phone}
                onChangeText={setPhone}
                placeholder="9876543211"
                keyboardType="phone-pad"
                autoCapitalize="none"
                inputMode="tel"
              />
            </View>

            <Text style={styles.fieldLabel}>Password</Text>
            <View style={styles.inputWrap}>
              <Icon name="lock-closed-outline" size={18} color={colors.muted} />
              <Input
                style={styles.input}
                value={password}
                onChangeText={setPassword}
                placeholder="password123"
                secureTextEntry
                autoCapitalize="none"
              />
            </View>

            {error ? (
              <View style={styles.errorBanner}>
                <Icon name="alert-circle-outline" size={16} color={colors.danger} />
                <Text style={styles.errorText}>{error}</Text>
              </View>
            ) : null}

            <Button
              title="Sign in"
              size="lg"
              icon="arrow-forward"
              iconPosition="right"
              full
              onPress={handleLogin}
              loading={busy}
              disabled={busy || !phone.trim() || !password}
              style={styles.signIn}
            />

            <View style={styles.demoChip}>
              <Icon name="flask-outline" size={13} color={colors.warning} />
              <Text style={styles.demoText}>Demo: Kumar · 9876543211 · password123</Text>
            </View>
          </Card>
        </ScrollView>
      </SafeAreaView>
    </KeyboardAvoidingView>
  )
}

const styles = StyleSheet.create({
  flex: { flex: 1, backgroundColor: colors.bg },
  wrap: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: spacing.xl,
    paddingVertical: spacing.xxl + 8,
  },
  brand: { alignItems: 'center', marginBottom: spacing.xl + 4 },
  logo: {
    width: 76,
    height: 76,
    borderRadius: 22,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.lg,
    ...shadows.raised,
  },
  appName: { fontSize: 26, fontWeight: '800', color: colors.text, letterSpacing: -0.5 },
  tagline: {
    fontSize: 13.5,
    color: colors.muted,
    textAlign: 'center',
    marginTop: 6,
    lineHeight: 19,
    maxWidth: 280,
  },
  card: { padding: spacing.xl, gap: 2 },
  cardTitle: { fontSize: 20, fontWeight: '800', color: colors.text },
  cardSub: { color: colors.muted, fontSize: 13, marginBottom: spacing.sm + 2 },
  fieldLabel: {
    fontSize: 12,
    fontWeight: '700',
    color: colors.textSoft,
    marginTop: spacing.md,
    marginBottom: 6,
  },
  inputWrap: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 10,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.md,
    backgroundColor: colors.surface,
    paddingHorizontal: spacing.md,
  },
  input: { borderWidth: 0, backgroundColor: 'transparent', paddingVertical: 13, flex: 1 },
  errorBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    backgroundColor: colors.dangerSoft,
    borderRadius: radius.md,
    padding: spacing.md,
    marginTop: spacing.md,
  },
  errorText: { color: colors.danger, fontSize: 13, flex: 1, fontWeight: '600' },
  signIn: { marginTop: spacing.xl },
  demoChip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    backgroundColor: colors.warningSoft,
    borderRadius: radius.pill,
    paddingHorizontal: 12,
    paddingVertical: 7,
    alignSelf: 'center',
    marginTop: spacing.lg,
  },
  demoText: { color: colors.warning, fontSize: 12, fontWeight: '700' },
})