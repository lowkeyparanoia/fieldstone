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
  CheckCircle2,
  XCircle,
  AlertCircle
} from 'lucide-react'
import { formatNumber, formatBytes, formatRelativeTime } from '@/lib/utils'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

// Mock data for the chart
const mockChartData = [
  { time: '00:00', requests: 120 },
  { time: '04:00', requests: 80 },
  { time: '08:00', requests: 340 },
  { time: '12:00', requests: 520 },
  { time: '16:00', requests: 680 },
  { time: '20:00', requests: 420 },
  { time: '23:59', requests: 280 },
]

function StatCard({ 
  title, 
  value, 
  description, 
  icon: Icon, 
  trend,
  isLoading 
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
        <CardTitle className="text-sm font-medium">
          {title}
        </CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Skeleton className="h-8 w-24" />
        ) : (
          <>
            <div className="text-2xl font-bold">{value}</div>
            <p className="text-xs text-muted-foreground">
              {description}
            </p>
            {trend && (
              <div className="flex items-center text-xs text-emerald-600 mt-1">
                <TrendingUp className="h-3 w-3 mr-1" />
                {trend}
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}

function StatusIndicator({ status, label }: { status: 'healthy' | 'unhealthy' | 'warning'; label: string }) {
  const icons = {
    healthy: CheckCircle2,
    unhealthy: XCircle,
    warning: AlertCircle,
  }
  
  const colors = {
    healthy: 'text-emerald-500',
    unhealthy: 'text-destructive',
    warning: 'text-amber-500',
  }
  
  const Icon = icons[status]
  
  return (
    <div className="flex items-center gap-2">
      <Icon className={`h-5 w-5 ${colors[status]}`} />
      <span className="text-sm">{label}</span>
    </div>
  )
}

export function DashboardPage() {
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: activities, isLoading: activitiesLoading } = useDashboardActivities(10)
  const { data: health, isLoading: healthLoading } = useHealthStatus()

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
          value={statsLoading ? '' : formatNumber(stats?.totalRecords || 0)}
          description="Across all collections"
          icon={Database}
          trend="+12% from last month"
          isLoading={statsLoading}
        />
        <StatCard
          title="Total Users"
          value={statsLoading ? '' : formatNumber(stats?.totalUsers || 0)}
          description="Registered users"
          icon={Users}
          trend="+5 this week"
          isLoading={statsLoading}
        />
        <StatCard
          title="API Requests"
          value={statsLoading ? '' : `${stats?.requestsPerMinute || 0}/min`}
          description="Average in last hour"
          icon={Zap}
          isLoading={statsLoading}
        />
        <StatCard
          title="Storage Used"
          value={statsLoading ? '' : formatBytes(stats?.storageUsed || 0)}
          description={`of ${formatBytes(stats?.storageLimit || 10737418240)}`}
          icon={HardDrive}
          isLoading={statsLoading}
        />
      </div>

      {/* Charts and Activity */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-7">
        {/* Requests Chart */}
        <Card className="col-span-4">
          <CardHeader>
            <CardTitle>API Requests</CardTitle>
            <CardDescription>
              Request volume over the last 24 hours
            </CardDescription>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={mockChartData}>
                <defs>
                  <linearGradient id="colorRequests" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                <XAxis 
                  dataKey="time" 
                  stroke="hsl(var(--muted-foreground))"
                  fontSize={12}
                />
                <YAxis 
                  stroke="hsl(var(--muted-foreground))"
                  fontSize={12}
                />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: 'hsl(var(--background))',
                    border: '1px solid hsl(var(--border))',
                    borderRadius: '6px'
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="requests"
                  stroke="hsl(var(--primary))"
                  fillOpacity={1}
                  fill="url(#colorRequests)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        {/* System Health */}
        <Card className="col-span-3">
          <CardHeader>
            <CardTitle>System Health</CardTitle>
            <CardDescription>
              Current status of system services
            </CardDescription>
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
                <StatusIndicator 
                  status={health?.services.database === 'healthy' ? 'healthy' : 'unhealthy'} 
                  label="Database Connection"
                />
                <StatusIndicator 
                  status={health?.services.api === 'healthy' ? 'healthy' : 'unhealthy'} 
                  label="API Server"
                />
                <StatusIndicator 
                  status={health?.services.cache === 'healthy' ? 'healthy' : 'warning'} 
                  label="Cache Layer"
                />
                <div className="mt-4 pt-4 border-t">
                  <p className="text-xs text-muted-foreground">
                    Last checked: {health?.timestamp ? formatRelativeTime(health.timestamp) : 'Never'}
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
          <CardTitle>Recent Activity</CardTitle>
          <CardDescription>
            Latest actions performed in the system
          </CardDescription>
        </CardHeader>
        <CardContent>
          {activitiesLoading ? (
            <div className="space-y-4">
              {[1, 2, 3, 4, 5].map((i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : (
            <div className="space-y-4">
              {activities && activities.length > 0 ? (
                activities.map((activity) => (
                  <div key={activity.id} className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
                        <Activity className="h-4 w-4 text-muted-foreground" />
                      </div>
                      <div>
                        <p className="text-sm font-medium">{activity.message}</p>
                        <p className="text-xs text-muted-foreground">
                          {activity.userId && `by ${activity.userId} • `}
                          {formatRelativeTime(activity.createdAt)}
                        </p>
                      </div>
                    </div>
                    <Badge variant="outline">
                      {activity.type.replace(/_/g, ' ')}
                    </Badge>
                  </div>
                ))
              ) : (
                <p className="text-center text-muted-foreground py-8">
                  No recent activity
                </p>
              )}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
