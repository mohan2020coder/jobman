import Ionicons from '@expo/vector-icons/Ionicons'
import type { ComponentProps, ReactNode } from 'react'
import {
  ActivityIndicator,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from 'react-native'
import type { TextInputProps } from 'react-native'
import { SafeAreaView } from 'react-native-safe-area-context'
import { colors, radius, shadows, spacing, statusColor, statusSoft, tint } from '../theme'
import { initials } from '../lib/format'

export type IconName = ComponentProps<typeof Ionicons>['name']

// ---------------- Badge / Pill ----------------

export function Badge({ value }: { value: string }) {
  const key = value.replace(/\s/g, '_') as keyof typeof statusColor
  const fg = statusColor[key] ?? colors.muted
  const bg = statusSoft[key] ?? colors.border
  return (
    <View style={[styles.badge, { backgroundColor: bg }]}>
      <View style={[styles.badgeDot, { backgroundColor: fg }]} />
      <Text style={[styles.badgeText, { color: fg }]}>{value.replace(/_/g, ' ')}</Text>
    </View>
  )
}

export function Pill({ label, color }: { label: string; color?: string }) {
  const c = color ?? colors.muted
  return (
    <View style={[styles.pill, { backgroundColor: tint(c, 0.12) }]}>
      <Text style={[styles.pillText, { color: c }]}>{label}</Text>
    </View>
  )
}

// ---------------- Cards & layout ----------------

export function Card({
  children,
  style,
  pad = true,
}: {
  children: ReactNode
  style?: object
  pad?: boolean
}) {
  return <View style={[styles.card, pad && styles.cardPad, style]}>{children}</View>
}

export function Divider({ style }: { style?: object }) {
  return <View style={[styles.divider, style]} />
}

export function SectionHeader({
  title,
  action,
  onAction,
}: {
  title: string
  action?: string
  onAction?: () => void
}) {
  return (
    <View style={styles.sectionHeader}>
      <Text style={styles.sectionTitle}>{title}</Text>
      {action ? (
        <Pressable onPress={onAction} hitSlop={8} style={styles.sectionAction}>
          <Icon name="chevron-forward" size={16} color={colors.primary} />
          <Text style={styles.sectionActionText}>{action}</Text>
        </Pressable>
      ) : null}
    </View>
  )
}

export function StatCard({
  label,
  value,
  icon,
  accent = colors.primary,
}: {
  label: string
  value: string | number
  icon: IconName
  accent?: string
}) {
  return (
    <Card style={[styles.statCard, { borderTopColor: accent, borderTopWidth: 3 }]}>
      <View style={[styles.statIcon, { backgroundColor: tint(accent, 0.12) }]}>
        <Icon name={icon} size={18} color={accent} />
      </View>
      <Text style={styles.statValue}>{value}</Text>
      <Text style={styles.statLabel}>{label}</Text>
    </Card>
  )
}

// ---------------- Buttons ----------------

type ButtonVariant = 'primary' | 'soft' | 'outline' | 'ghost' | 'danger'
type ButtonSize = 'sm' | 'md' | 'lg'

export function Button({
  title,
  onPress,
  variant = 'primary',
  size = 'md',
  icon,
  iconPosition = 'left',
  disabled,
  loading,
  full,
  flex1,
  style,
}: {
  title: string
  onPress?: () => void
  variant?: ButtonVariant
  size?: ButtonSize
  icon?: IconName
  iconPosition?: 'left' | 'right'
  disabled?: boolean
  loading?: boolean
  full?: boolean
  flex1?: boolean
  style?: object
}) {
  const base =
    variant === 'danger'
      ? { backgroundColor: colors.dangerSoft, borderColor: '#fecaca' }
      : variant === 'soft'
        ? { backgroundColor: colors.primarySoft, borderColor: colors.primarySoft }
        : variant === 'outline'
          ? { backgroundColor: colors.surface, borderColor: colors.borderStrong }
          : variant === 'ghost'
            ? { backgroundColor: 'transparent', borderColor: 'transparent' }
            : { backgroundColor: colors.primary, borderColor: colors.primary }
  const fg =
    variant === 'danger'
      ? colors.danger
      : variant === 'soft'
        ? colors.primaryDark
        : variant === 'outline' || variant === 'ghost'
          ? colors.textSoft
          : colors.white

  const height = size === 'sm' ? 36 : size === 'lg' ? 54 : 48
  const fontSize = size === 'sm' ? 13 : size === 'lg' ? 16 : 15

  return (
    <Pressable
      onPress={onPress}
      disabled={disabled || loading}
      style={({ pressed }) => [
        styles.button,
        base,
        { height },
        full && styles.buttonFull,
        flex1 && styles.buttonFlex,
        variant === 'primary' && shadows.raised,
        (disabled || loading) && styles.buttonDisabled,
        pressed && !disabled && !loading && styles.buttonPressed,
        style,
      ]}
    >
      {loading ? (
        <ActivityIndicator size="small" color={fg} />
      ) : (
        <View style={styles.buttonContent}>
          {icon && iconPosition === 'left' ? <Icon name={icon} size={fontSize + 2} color={fg} /> : null}
          <Text style={[styles.buttonText, { color: fg, fontSize }]}>{title}</Text>
          {icon && iconPosition === 'right' ? <Icon name={icon} size={fontSize + 2} color={fg} /> : null}
        </View>
      )}
    </Pressable>
  )
}

export function IconButton({
  name,
  onPress,
  size = 20,
  color = colors.textSoft,
  bg,
  style,
}: {
  name: IconName
  onPress?: () => void
  size?: number
  color?: string
  bg?: string
  style?: object
}) {
  return (
    <Pressable
      onPress={onPress}
      hitSlop={8}
      style={({ pressed }) => [
        styles.iconButton,
        bg ? { backgroundColor: bg } : null,
        pressed && styles.buttonPressed,
        style,
      ]}
    >
      <Icon name={name} size={size} color={color} />
    </Pressable>
  )
}

export function Chip({
  label,
  active,
  onPress,
  accent = colors.primary,
  icon,
}: {
  label: string
  active?: boolean
  onPress?: () => void
  accent?: string
  icon?: IconName
}) {
  return (
    <Pressable
      onPress={onPress}
      style={({ pressed }) => [
        styles.chip,
        active ? { backgroundColor: accent, borderColor: accent } : styles.chipInactive,
        pressed && styles.buttonPressed,
      ]}
    >
      {icon && active ? <Icon name={icon} size={14} color={colors.white} /> : null}
      <Text style={[styles.chipText, active ? { color: colors.white } : { color: colors.textSoft }]}>
        {label}
      </Text>
    </Pressable>
  )
}

// ---------------- Form ----------------

export function FormLabel({ children }: { children: ReactNode }) {
  return <Text style={styles.formLabel}>{children}</Text>
}

export function Input({
  style,
  ...props
}: TextInputProps & { style?: object }) {
  return <TextInput style={[styles.input, style]} placeholderTextColor={colors.faint} {...props} />
}

// ---------------- Avatar ----------------

export function Avatar({ name, size = 44 }: { name: string | null | undefined; size?: number }) {
  const initial = initials(name)
  const text = name ?? '?'
  let hash = 0
  for (let i = 0; i < text.length; i++) hash = (hash * 31 + text.charCodeAt(i)) | 0
  const palette = [colors.primary, colors.accent, '#0891b2', '#059669', '#c26a18']
  const color = palette[Math.abs(hash) % palette.length]
  return (
    <View
      style={[
        styles.avatar,
        { width: size, height: size, borderRadius: size / 2, backgroundColor: tint(color, 0.15) },
      ]}
    >
      <Text style={[styles.avatarText, { color, fontSize: size * 0.42 }]}>{initial}</Text>
    </View>
  )
}

// ---------------- State views ----------------

export function Spinner({ label = 'Loading…' }: { label?: string }) {
  return (
    <View style={styles.stateView}>
      <ActivityIndicator size="large" color={colors.primary} />
      {label ? <Muted>{label}</Muted> : null}
    </View>
  )
}

export function EmptyState({
  icon = 'file-tray-outline',
  title = 'Nothing here yet',
  message,
  actionLabel,
  onAction,
}: {
  icon?: IconName
  title?: string
  message?: string
  actionLabel?: string
  onAction?: () => void
}) {
  return (
    <View style={styles.emptyState}>
      <View style={styles.emptyIcon}>
        <Icon name={icon} size={30} color={colors.faint} />
      </View>
      <Text style={styles.emptyTitle}>{title}</Text>
      {message ? <Muted style={styles.emptyMsg}>{message}</Muted> : null}
      {actionLabel && onAction ? (
        <Button title={actionLabel} variant="soft" size="sm" onPress={onAction} style={{ marginTop: 14 }} />
      ) : null}
    </View>
  )
}

export function ErrorView({ message, onRetry }: { message: string; onRetry?: () => void }) {
  return (
    <View style={styles.emptyState}>
      <View style={[styles.emptyIcon, { backgroundColor: colors.dangerSoft }]}>
        <Icon name="cloud-offline-outline" size={30} color={colors.danger} />
      </View>
      <Text style={styles.emptyTitle}>Something went wrong</Text>
      <Muted style={styles.emptyMsg}>{message}</Muted>
      {onRetry ? (
        <Button title="Try again" variant="soft" size="sm" icon="refresh" onPress={onRetry} style={{ marginTop: 14 }} />
      ) : null}
    </View>
  )
}

export function LoadingView({ label }: { label?: string }) {
  return <View style={styles.stateViewSpaced}><Spinner label={label} /></View>
}

export function SafeScreen({ children }: { children: ReactNode }) {
  return (
    <SafeAreaView style={styles.safeScreen} edges={['top', 'left', 'right']}>
      {children}
    </SafeAreaView>
  )
}

// ---------------- Text helpers ----------------

export function Label({ children }: { children: ReactNode }) {
  return <Text style={styles.label}>{children}</Text>
}

export function Muted({ children, style }: { children: ReactNode; style?: object }) {
  return <Text style={[styles.muted, style]}>{children}</Text>
}

export function SectionTitle({ children, style }: { children: ReactNode; style?: object }) {
  return <Text style={[styles.sectionTitle, style]}>{children}</Text>
}

// ---------------- Icon ----------------

export function Icon({
  name,
  size = 20,
  color = colors.textSoft,
  style,
}: {
  name: IconName
  size?: number
  color?: string
  style?: object
}) {
  return <Ionicons name={name} size={size} color={color} style={style} />
}

const styles = StyleSheet.create({
  safeScreen: { flex: 1, backgroundColor: colors.bg },
  badge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    paddingHorizontal: 9,
    paddingVertical: 4,
    borderRadius: radius.pill,
  },
  badgeDot: { width: 6, height: 6, borderRadius: 3 },
  badgeText: { fontSize: 11, fontWeight: '700', textTransform: 'uppercase', letterSpacing: 0.4 },
  pill: {
    alignSelf: 'flex-start',
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: radius.pill,
  },
  pillText: { fontSize: 11, fontWeight: '700' },
  card: {
    backgroundColor: colors.surface,
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.lg,
    ...shadows.card,
  },
  cardPad: { padding: spacing.lg },
  divider: { height: 1, backgroundColor: colors.border, marginVertical: spacing.md },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginBottom: spacing.md,
  },
  sectionTitle: { fontSize: 18, fontWeight: '800', color: colors.text },
  sectionAction: { flexDirection: 'row', alignItems: 'center', gap: 2 },
  sectionActionText: { fontSize: 13, fontWeight: '700', color: colors.primary },
  statCard: { flex: 1, position: 'relative', borderTopLeftRadius: radius.sm, borderTopRightRadius: radius.sm },
  statIcon: { width: 34, height: 34, borderRadius: radius.sm, alignItems: 'center', justifyContent: 'center', marginBottom: 10 },
  statValue: { fontSize: 24, fontWeight: '800', color: colors.text, letterSpacing: -0.4 },

  statLabel: { fontSize: 12, fontWeight: '600', color: colors.muted, marginTop: 2 },
  button: {
    borderRadius: radius.md,
    borderWidth: 1,
    paddingHorizontal: spacing.lg - 2,
    alignItems: 'center',
    justifyContent: 'center',
  },
  buttonContent: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  buttonText: { fontWeight: '700' },
  buttonFull: { alignSelf: 'stretch' },
  buttonFlex: { flex: 1 },
  buttonDisabled: { opacity: 0.55 },
  buttonPressed: { opacity: 0.85 },
  iconButton: { width: 40, height: 40, borderRadius: radius.md, alignItems: 'center', justifyContent: 'center' },
  chip: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 5,
    paddingHorizontal: 14,
    paddingVertical: 8,
    borderRadius: radius.pill,
    borderWidth: 1,
  },
  chipInactive: { backgroundColor: colors.surface, borderColor: colors.border },
  chipText: { fontSize: 13, fontWeight: '700' },
  formLabel: { fontSize: 12, fontWeight: '700', color: colors.textSoft, marginBottom: 6 },
  input: {
    borderWidth: 1,
    borderColor: colors.border,
    borderRadius: radius.md,
    paddingHorizontal: spacing.md,
    paddingVertical: 12,
    fontSize: 15,
    backgroundColor: colors.surface,
    color: colors.text,
  },
  avatar: { alignItems: 'center', justifyContent: 'center' },
  avatarText: { fontWeight: '800' },
  label: {
    fontSize: 11,
    color: colors.muted,
    fontWeight: '700',
    textTransform: 'uppercase',
    letterSpacing: 0.6,
  },
  muted: { color: colors.muted, fontSize: 13 },
  stateView: { alignItems: 'center', justifyContent: 'center', paddingVertical: spacing.xl, gap: 10 },
  stateViewSpaced: { alignItems: 'center', justifyContent: 'center', paddingVertical: spacing.xxl, gap: 10 },
  emptyState: { alignItems: 'center', paddingVertical: spacing.xxl, paddingHorizontal: spacing.xl },
  emptyIcon: {
    width: 68,
    height: 68,
    borderRadius: 34,
    backgroundColor: colors.surfaceAlt,
    borderWidth: 1,
    borderColor: colors.border,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.md,
  },
  emptyTitle: { fontSize: 16, fontWeight: '800', color: colors.text },
  emptyMsg: { textAlign: 'center', marginTop: 4, lineHeight: 19 },
})