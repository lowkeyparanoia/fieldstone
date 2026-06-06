import { useDashboardStats, useDashboardActivities, useHealthStatus } from '@/hooks/useApi'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  Database,
  Users,
  Activity,
  HardDrive,
  Zap,
  TrendingUp,
} from 'lucide-react'
import { cn, formatNumber, formatBytes, formatRelativeTime } from '@/lib/utils'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { MOCK_STATS, MOCK_HEALTH, MOCK_ACTIVITIES, MOCK_CHART } from '@/lib/mock'

function StatCard({
  title,
  value,
  description,
  icon: Icon,
  trend,
  isLoading,
}: {
  title: string
  value: string | number
  description: string
  icon: React.ElementType
  trend?: string
  isLoading?: boolean
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <>
            <div className="text-2xl font-bold">{value}</div>
            <p className="text-xs text-muted-foreground">{description}</p>
            {trend && (
              <div className="mt-1 flex items-center text-xs text-emerald-500">
                <TrendingUp className="mr-1 h-3 w-3" />
                {trend}
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}

function StatusDot({
  status,
  label,
}: {
  status: 'healthy' | 'unhealthy' | 'warning'
  label: string
}) {
  const colors = {
    healthy: 'bg-emerald-500',
    warning: 'bg-amber-500',
    unhealthy: 'bg-destructive',
  }
  return (
    <div className="flex items-center gap-2.5 text-sm">
      <span className="relative flex">
        {status === 'healthy' && (
          <span className="absolute inset-0 animate-ping rounded-full bg-emerald-500 opacity-60" />
        )}
        <span className={cn('h-2.5 w-2.5 rounded-full', colors[status])} />
      </span>
      <span>{label}</span>
    </div>
  )
}

export function DashboardPage() {
  const { data: statsData, isLoading: statsLoading } = useDashboardStats()
  const { data: activitiesData } = useDashboardActivities(10)
  const { data: healthData, isLoading: healthLoading } = useHealthStatus()

  const stats = statsData ?? MOCK_STATS
  const health = healthData ?? MOCK_HEALTH
  const activities = activitiesData && activitiesData.length > 0 ? activitiesData : MOCK_ACTIVITIES

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Welcome to Fieldstone Admin. Monitor your backend performance and activity.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard
          title="Total Records"
          value={formatNumber(stats.totalRecords)}
          description="Across all collections"
          icon={Database}
          trend="+12% from last month"
          isLoading={statsLoading}
        />
        <StatCard
          title="Total Users"
          value={formatNumber(stats.totalUsers)}
          description="Registered users"
          icon={Users}
          trend="+5 this week"
          isLoading={statsLoading}
        />
        <StatCard
          title="API Requests"
          value={`${stats.requestsPerMinute}/min`}
          description="Average in last hour"
          icon={Zap}
          isLoading={statsLoading}
        />
        <StatCard
          title="Storage Used"
          value={formatBytes(stats.storageUsed)}
          description={`of ${formatBytes(stats.storageLimit)}`}
          icon={HardDrive}
          isLoading={statsLoading}
        />
      </div>

      {/* Chart + System Health */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        <Card className="lg:col-span-4">
          <CardHeader>
            <CardTitle className="text-2xl">API Requests</CardTitle>
            <CardDescription>Request volume over the last 24 hours</CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={260}>
              <AreaChart data={MOCK_CHART} margin={{ top: 8, right: 8, left: -16, bottom: 0 }}>
                <defs>
                  <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                <XAxis dataKey="time" stroke="hsl(var(--muted-foreground))" fontSize={11} tickLine={false} axisLine={false} />
                <YAxis stroke="hsl(var(--muted-foreground))" fontSize={11} tickLine={false} axisLine={false} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'hsl(var(--popover))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '8px',
                    fontSize: '12px',
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="requests"
                  stroke="hsl(var(--primary))"
                  strokeWidth={2}
                  fillOpacity={1}
                  fill="url(#colorRequests)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card className="lg:col-span-3">
          <CardHeader>
            <CardTitle className="text-2xl">System Health</CardTitle>
            <CardDescription>Current status of system services</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {healthLoading ? (
              <>
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-full" />
                <Skeleton className="h-6 w-full" />
              </>
            ) : (
              <>
                <StatusDot
                  status={health.services.database === 'healthy' ? 'healthy' : 'unhealthy'}
                  label="Database Connection"
                />
                <StatusDot
                  status={health.services.api === 'healthy' ? 'healthy' : 'unhealthy'}
                  label="API Server"
                />
                <StatusDot
                  status={health.services.cache === 'healthy' ? 'healthy' : 'warning'}
                  label="Cache Layer"
                />
                <div className="mt-4 border-t pt-4">
                  <p className="text-xs text-muted-foreground">
                    Last checked: {health.timestamp ? formatRelativeTime(health.timestamp) : 'Never'}
                  </p>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Recent Activity */}
      <Card>
        <CardHeader>
          <CardTitle className="text-2xl">Recent Activity</CardTitle>
          <CardDescription>Latest actions performed in the system</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {activities.map((activity) => (
            <div key={activity.id} className="flex items-center justify-between gap-3">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-full bg-muted">
                  <Activity className="h-4 w-4 text-muted-foreground" />
                </div>
                <div>
                  <p className="text-sm font-medium">{activity.message}</p>
                  <p className="text-xs text-muted-foreground">
                    {activity.userId && (
                      <span className="font-mono">{activity.userId} · </span>
                    )}
                    {formatRelativeTime(activity.createdAt)}
                  </p>
                </div>
              </div>
              <Badge variant="outline">{activity.type.replace(/_/g, ' ')}</Badge>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
