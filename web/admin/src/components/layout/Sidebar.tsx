import { NavLink, useLocation } from 'react-router-dom'
import {
  LayoutDashboard,
  Database,
  Users,
  Settings,
  Webhook,
  Puzzle,
  Activity,
  ChevronLeft,
  ChevronRight,
  Box,
  Sun,
  Moon,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useTheme } from '@/context/theme'

interface SidebarProps {
  isCollapsed: boolean
  onToggle: () => void
}

const navigation = [
  { name: 'Dashboard', href: '/', icon: LayoutDashboard },
  { name: 'Collections', href: '/collections', icon: Database },
  { name: 'Users', href: '/users', icon: Users },
  { name: 'Activity', href: '/activity', icon: Activity },
  { name: 'Plugins', href: '/plugins', icon: Puzzle },
  { name: 'Webhooks', href: '/webhooks', icon: Webhook },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar({ isCollapsed, onToggle }: SidebarProps) {
  const location = useLocation()
  const { theme, setTheme } = useTheme()
  const isDark =
    theme === 'dark' ||
    (theme === 'system' &&
      typeof window !== 'undefined' &&
      window.matchMedia('(prefers-color-scheme: dark)').matches)

  return (
    <aside
      className={cn(
        'fixed left-0 top-0 z-40 h-screen border-r bg-background transition-all duration-300',
        isCollapsed ? 'w-16' : 'w-64'
      )}
    >
      {/* Header */}
      <div className="flex h-16 items-center border-b px-4">
        <div className={cn('flex items-center gap-2', isCollapsed && 'justify-center w-full')}>
          <Box className="h-6 w-6 text-primary flex-shrink-0" />
          {!isCollapsed && (
            <span className="text-lg font-bold">Fieldstone</span>
          )}
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex flex-col gap-1 p-2">
        {navigation.map((item) => {
          const isActive = location.pathname === item.href || 
            (item.href !== '/' && location.pathname.startsWith(item.href))
          
          return (
            <NavLink
              key={item.name}
              to={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-muted hover:text-foreground',
                isCollapsed && 'justify-center px-2'
              )}
              title={isCollapsed ? item.name : undefined}
            >
              <item.icon className="h-5 w-5 flex-shrink-0" />
              {!isCollapsed && <span>{item.name}</span>}
            </NavLink>
          )
        })}
      </nav>

      {/* Footer: account + theme + collapse */}
      <div className="absolute bottom-0 left-0 right-0 flex flex-col gap-1 border-t p-2">
        <div
          className={cn(
            'flex items-center gap-2.5 rounded-md px-2.5 py-2',
            isCollapsed && 'justify-center px-0'
          )}
          title="admin@fieldstone.io"
        >
          <span className="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
            A
          </span>
          {!isCollapsed && (
            <div className="min-w-0 flex-1">
              <div className="text-[13px] font-semibold leading-tight">Admin</div>
              <div className="truncate font-mono text-[11px] text-muted-foreground">
                admin@fieldstone.io
              </div>
            </div>
          )}
        </div>

        <div className="flex gap-1">
          <Button
            variant="ghost"
            size="icon"
            className="h-9 flex-1"
            onClick={() => setTheme(isDark ? 'light' : 'dark')}
            title="Toggle theme"
          >
            {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-9 flex-1"
            onClick={onToggle}
            title={isCollapsed ? 'Expand' : 'Collapse'}
          >
            {isCollapsed ? (
              <ChevronRight className="h-4 w-4" />
            ) : (
              <ChevronLeft className="h-4 w-4" />
            )}
          </Button>
        </div>
      </div>
    </aside>
  )
}
