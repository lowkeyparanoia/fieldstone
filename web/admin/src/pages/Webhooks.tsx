import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { 
  Plus, 
  Play, 
  Pause, 
  Trash2, 
  Webhook,
  CheckCircle2,
  XCircle
} from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

// Mock webhooks data
const mockWebhooks = [
  {
    id: '1',
    name: 'User Created Notification',
    url: 'https://api.example.com/webhooks/users',
    events: ['user.created'],
    status: 'active',
    lastTriggered: '2026-06-06T10:30:00Z',
    successRate: 98.5,
  },
  {
    id: '2',
    name: 'Order Processing',
    url: 'https://api.example.com/webhooks/orders',
    events: ['record.created', 'record.updated'],
    status: 'active',
    lastTriggered: '2026-06-06T09:15:00Z',
    successRate: 99.2,
  },
  {
    id: '3',
    name: 'Failed Payment Alert',
    url: 'https://api.example.com/webhooks/alerts',
    events: ['record.updated'],
    status: 'paused',
    lastTriggered: '2026-06-05T16:45:00Z',
    successRate: 85.0,
  },
]

const mockLogs = [
  {
    id: '1',
    webhookId: '1',
    status: 'success',
    statusCode: 200,
    timestamp: '2026-06-06T10:30:00Z',
    duration: 120,
    response: '{"received": true}',
  },
  {
    id: '2',
    webhookId: '1',
    status: 'error',
    statusCode: 500,
    timestamp: '2026-06-06T10:25:00Z',
    duration: 5000,
    error: 'Internal Server Error',
  },
  {
    id: '3',
    webhookId: '2',
    status: 'success',
    statusCode: 204,
    timestamp: '2026-06-06T09:15:00Z',
    duration: 85,
  },
]

export function WebhooksPage() {
  const [webhooks, setWebhooks] = useState(mockWebhooks)

  const toggleWebhook = (id: string) => {
    setWebhooks(webhooks.map(w => 
      w.id === id ? { ...w, status: w.status === 'active' ? 'paused' : 'active' } : w
    ))
  }

  const deleteWebhook = (id: string) => {
    setWebhooks(webhooks.filter(w => w.id !== id))
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Webhooks</h1>
          <p className="text-muted-foreground">
            Manage real-time event notifications
          </p>
        </div>
        <Dialog>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Create Webhook
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[525px]">
            <DialogHeader>
              <DialogTitle>Create Webhook</DialogTitle>
              <DialogDescription>
                Set up a new webhook to receive real-time event notifications.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <label htmlFor="name">Webhook Name</label>
                <Input id="name" placeholder="e.g., Order Notifications" />
              </div>
              <div className="grid gap-2">
                <label htmlFor="url">Endpoint URL</label>
                <Input id="url" placeholder="https://api.example.com/webhook" />
              </div>
              <div className="grid gap-2">
                <label>Events</label>
                <Select>
                  <SelectTrigger>
                    <SelectValue placeholder="Select events" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="record.created">Record Created</SelectItem>
                    <SelectItem value="record.updated">Record Updated</SelectItem>
                    <SelectItem value="record.deleted">Record Deleted</SelectItem>
                    <SelectItem value="user.created">User Created</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button type="submit">Create Webhook</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Tabs defaultValue="webhooks" className="space-y-4">
        <TabsList>
          <TabsTrigger value="webhooks">Webhooks</TabsTrigger>
          <TabsTrigger value="logs">Recent Logs</TabsTrigger>
        </TabsList>

        <TabsContent value="webhooks" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {webhooks.map((webhook) => (
              <Card key={webhook.id}>
                <CardHeader className="pb-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-2">
                      <Webhook className="h-5 w-5 text-muted-foreground" />
                      <CardTitle className="text-base">{webhook.name}</CardTitle>
                    </div>
                    <Badge variant={webhook.status === 'active' ? 'success' : 'secondary'}>
                      {webhook.status}
                    </Badge>
                  </div>
                  <CardDescription className="text-xs truncate">
                    {webhook.url}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    <div className="flex flex-wrap gap-1">
                      {webhook.events.map((event) => (
                        <Badge key={event} variant="outline" className="text-xs">
                          {event}
                        </Badge>
                      ))}
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Success Rate</span>
                      <span className={`font-medium ${
                        webhook.successRate > 95 ? 'text-emerald-600' : 'text-amber-600'
                      }`}>
                        {webhook.successRate}%
                      </span>
                    </div>
                    <div className="flex gap-2 pt-2">
                      <Button
                        variant="outline"
                        size="sm"
                        className="flex-1"
                        onClick={() => toggleWebhook(webhook.id)}
                      >
                        {webhook.status === 'active' ? (
                          <Pause className="mr-1 h-3 w-3" />
                        ) : (
                          <Play className="mr-1 h-3 w-3" />
                        )}
                        {webhook.status === 'active' ? 'Pause' : 'Resume'}
                      </Button>
                      <Button
                        variant="outline"
                        size="sm"
                        className="text-destructive hover:bg-destructive/10"
                        onClick={() => deleteWebhook(webhook.id)}
                      >
                        <Trash2 className="h-3 w-3" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {webhooks.length === 0 && (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Webhook className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">No webhooks configured</p>
                <Button className="mt-4" variant="outline">
                  <Plus className="mr-2 h-4 w-4" />
                  Create your first webhook
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="logs">
          <Card>
            <CardHeader>
              <CardTitle>Recent Delivery Attempts</CardTitle>
              <CardDescription>
                Last 50 webhook delivery attempts
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              {mockLogs.map((log) => (
                <div
                  key={log.id}
                  className="flex items-center justify-between p-3 border rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    {log.status === 'success' ? (
                      <CheckCircle2 className="h-5 w-5 text-emerald-500" />
                    ) : (
                      <XCircle className="h-5 w-5 text-destructive" />
                    )}
                    <div>
                      <p className="font-medium text-sm">
                        {log.status === 'success' ? 'Success' : 'Failed'}
                        <span className="text-muted-foreground ml-2">
                          HTTP {log.statusCode}
                        </span>
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(log.timestamp).toLocaleString()}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="text-sm font-medium">{log.duration}ms</p>
                    <p className="text-xs text-muted-foreground">Response time</p>
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
