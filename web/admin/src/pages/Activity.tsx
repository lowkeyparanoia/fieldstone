import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { useDashboardActivities } from '@/hooks/useApi'
import {
  Activity as ActivityIcon,
  UserPlus,
  Database,
  FileEdit,
  Trash2,
  Save,
  Bell,
  type LucideIcon,
} from 'lucide-react'
import { cn, formatRelativeTime } from '@/lib/utils'
import { MOCK_ACTIVITIES } from '@/lib/mock'
import type { Activity } from '@/lib/api'

const ACT_TYPE: Record<string, { icon: LucideIcon; chip: string }> = {
  user_created: { icon: UserPlus, chip: 'bg-blue-500/15 text-blue-500' },
  collection_modified: { icon: Database, chip: 'bg-purple-500/15 text-purple-500' },
  record_created: { icon: FileEdit, chip: 'bg-emerald-500/15 text-emerald-500' },
  record_updated: { icon: FileEdit, chip: 'bg-amber-500/15 text-amber-500' },
  record_deleted: { icon: Trash2, chip: 'bg-red-500/15 text-red-500' },
  backup_completed: { icon: Save, chip: 'bg-cyan-500/15 text-cyan-500' },
  webhook_triggered: { icon: Bell, chip: 'bg-pink-500/15 text-pink-500' },
}

export function ActivityPage() {
  const { data } = useDashboardActivities(50)
  const activities: Activity[] = data && data.length ? data : MOCK_ACTIVITIES

  const grouped = activities.reduce<Record<string, Activity[]>>((groups, activity) => {
    const date = new Date(activity.createdAt).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
    ;(groups[date] ||= []).push(activity)
    return groups
  }, {})

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Activity Log</h1>
        <p className="text-muted-foreground">Track all actions performed in the system</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Recent Activity</CardTitle>
          <CardDescription>A chronological list of all system events</CardDescription>
        </CardHeader>
        <CardContent className="space-y-8">
          {Object.entries(grouped).map(([date, items]) => (
            <div key={date}>
              <h3 className="mb-3 text-sm font-semibold text-muted-foreground">{date}</h3>
              <div className="space-y-3">
                {items.map((activity) => {
                  const meta = ACT_TYPE[activity.type] || {
                    icon: ActivityIcon,
                    chip: 'bg-muted text-muted-foreground',
                  }
                  const Icon = meta.icon
                  return (
                    <div
                      key={activity.id}
                      className="flex items-start gap-4 rounded-lg border p-3 transition-colors hover:bg-muted/50"
                    >
                      <span
                        className={cn(
                          'flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full',
                          meta.chip
                        )}
                      >
                        <Icon className="h-4 w-4" />
                      </span>
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium">{activity.message}</p>
                        <div className="mt-1 flex flex-wrap items-center gap-2">
                          <Badge variant="outline" className="text-xs">
                            {activity.type.replace(/_/g, ' ')}
                          </Badge>
                          {activity.userId && (
                            <span className="font-mono text-xs text-muted-foreground">
                              by {activity.userId}
                            </span>
                          )}
                          <span className="text-xs text-muted-foreground">
                            {formatRelativeTime(activity.createdAt)}
                          </span>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
