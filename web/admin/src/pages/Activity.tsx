import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useDashboardActivities } from '@/hooks/useApi'
import { 
  Activity,
  UserPlus,
  Database,
  FileEdit,
  Trash2,
  Save,
  Bell
} from 'lucide-react'
import { formatRelativeTime } from '@/lib/utils'

const activityIcons: Record<string, React.ElementType> = {
  user_created: UserPlus,
  collection_modified: Database,
  record_created: FileEdit,
  record_updated: FileEdit,
  record_deleted: Trash2,
  backup_completed: Save,
  webhook_triggered: Bell,
}

const activityColors: Record<string, string> = {
  user_created: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-100',
  collection_modified: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-100',
  record_created: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900 dark:text-emerald-100',
  record_updated: 'bg-amber-100 text-amber-800 dark:bg-amber-900 dark:text-amber-100',
  record_deleted: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-100',
  backup_completed: 'bg-cyan-100 text-cyan-800 dark:bg-cyan-900 dark:text-cyan-100',
  webhook_triggered: 'bg-pink-100 text-pink-800 dark:bg-pink-900 dark:text-pink-100',
}

export function ActivityPage() {
  const { data: activities, isLoading } = useDashboardActivities(50)

  // Group activities by date
  const groupedActivities = activities?.reduce((groups, activity) => {
    const date = new Date(activity.createdAt).toLocaleDateString('en-US', {
      weekday: 'long',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })
    if (!groups[date]) {
      groups[date] = []
    }
    groups[date].push(activity)
    return groups
  }, {} as Record<string, typeof activities>)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Activity Log</h1>
        <p className="text-muted-foreground">
          Track all actions performed in the system
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>
            A chronological list of all system events
          </CardDescription>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : (
            <div className="space-y-8">
              {groupedActivities && Object.entries(groupedActivities).length > 0 ? (
                Object.entries(groupedActivities).map(([date, dayActivities]) => (
                  <div key={date}>
                    <h3 className="text-sm font-semibold text-muted-foreground mb-3">
                      {date}
                    </h3>
                    <div className="space-y-3">
                      {dayActivities?.map((activity) => {
                        const Icon = activityIcons[activity.type] || Activity
                        const colorClass = activityColors[activity.type] || 'bg-gray-100 text-gray-800'
                        
                        return (
                          <div
                            key={activity.id}
                            className="flex items-start gap-4 p-3 rounded-lg border hover:bg-muted/50 transition-colors"
                          >
                            <div className={`p-2 rounded-full ${colorClass}`}>
                              <Icon className="h-4 w-4" />
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-medium">
                                {activity.message}
                              </p>
                              <div className="flex items-center gap-2 mt-1">
                                <Badge variant="outline" className="text-xs">
                                  {activity.type.replace(/_/g, ' ')}
                                </Badge>
                                {activity.userId && (
                                  <span className="text-xs text-muted-foreground">
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
                ))
              ) : (
                <div className="text-center py-12">
                  <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No activity recorded yet</p>
                </div>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
