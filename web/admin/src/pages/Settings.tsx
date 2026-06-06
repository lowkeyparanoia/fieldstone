import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { Save, Database, HardDrive, Key, RefreshCw } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

function SettingRow({
  label,
  description,
  checked,
  onCheckedChange,
}: {
  label: string
  description: string
  checked: boolean
  onCheckedChange: (v: boolean) => void
}) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="space-y-0.5">
        <Label>{label}</Label>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <Switch checked={checked} onCheckedChange={onCheckedChange} />
    </div>
  )
}

export function SettingsPage() {
  const [isSaving, setIsSaving] = useState(false)
  const [flags, setFlags] = useState({
    reg: true,
    verify: true,
    publicApi: false,
    jwt: true,
    rate: true,
    passkey: false,
    autobackup: true,
    notifReg: false,
    notifFail: true,
    notifSys: true,
  })
  const set = (k: keyof typeof flags) => (v: boolean) => setFlags((f) => ({ ...f, [k]: v }))

  const handleSave = () => {
    setIsSaving(true)
    setTimeout(() => setIsSaving(false), 900)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Settings</h1>
          <p className="text-muted-foreground">Manage your Fieldstone instance configuration</p>
        </div>
        <Button onClick={handleSave} loading={isSaving}>
          <Save className="mr-2 h-4 w-4" />
          {isSaving ? 'Saving…' : 'Save Changes'}
        </Button>
      </div>

      <Tabs defaultValue="general" className="space-y-4">
        <TabsList>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="database">Database</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Instance Settings</CardTitle>
              <CardDescription>Configure basic instance information</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid gap-2">
                <Label htmlFor="instance-name">Instance Name</Label>
                <Input id="instance-name" defaultValue="Fieldstone Production" />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="admin-email">Admin Email</Label>
                <Input id="admin-email" type="email" defaultValue="admin@fieldstone.io" />
              </div>
              <div className="grid gap-2">
                <Label>Default Timezone</Label>
                <Select defaultValue="utc">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="utc">UTC</SelectItem>
                    <SelectItem value="est">Eastern Time</SelectItem>
                    <SelectItem value="pst">Pacific Time</SelectItem>
                    <SelectItem value="gmt">GMT</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Features</CardTitle>
              <CardDescription>Enable or disable platform features</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <SettingRow label="User Registration" description="Allow new users to register" checked={flags.reg} onCheckedChange={set('reg')} />
              <Separator />
              <SettingRow label="Email Verification" description="Require email verification for new users" checked={flags.verify} onCheckedChange={set('verify')} />
              <Separator />
              <SettingRow label="Public API Access" description="Allow unauthenticated API requests" checked={flags.publicApi} onCheckedChange={set('publicApi')} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="database" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Database Configuration</CardTitle>
              <CardDescription>Manage database connections and settings</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center gap-4 rounded-md border p-3">
                <Database className="h-5 w-5 text-muted-foreground" />
                <div className="flex-1">
                  <p className="text-sm font-medium">SQLite (Default)</p>
                  <p className="font-mono text-[13px] text-muted-foreground">file:data/fieldstone.db</p>
                </div>
                <Badge variant="success">Connected</Badge>
              </div>

              <Separator />

              <div className="grid gap-2">
                <Label>Backup Schedule</Label>
                <Select defaultValue="daily">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="hourly">Every hour</SelectItem>
                    <SelectItem value="daily">Daily</SelectItem>
                    <SelectItem value="weekly">Weekly</SelectItem>
                    <SelectItem value="manual">Manual only</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              <SettingRow label="Auto-backup" description="Automatically backup database" checked={flags.autobackup} onCheckedChange={set('autobackup')} />

              <div className="flex gap-2 pt-2">
                <Button variant="outline">
                  <HardDrive className="mr-2 h-4 w-4" />
                  Backup Now
                </Button>
                <Button variant="outline">
                  <RefreshCw className="mr-2 h-4 w-4" />
                  Restore
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="security" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Authentication</CardTitle>
              <CardDescription>Configure authentication settings</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <SettingRow label="JWT Authentication" description="Enable JWT-based authentication" checked={flags.jwt} onCheckedChange={set('jwt')} />
              <Separator />
              <SettingRow label="Rate Limiting" description="Limit requests per IP address" checked={flags.rate} onCheckedChange={set('rate')} />
              <Separator />
              <SettingRow label="WebAuthn / Passkeys" description="Enable passwordless authentication" checked={flags.passkey} onCheckedChange={set('passkey')} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-lg">API Keys</CardTitle>
              <CardDescription>Manage API keys for external access</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center justify-between rounded-md border p-3">
                <div className="flex items-center gap-3">
                  <Key className="h-4 w-4 text-muted-foreground" />
                  <div>
                    <p className="text-sm font-medium">Production API Key</p>
                    <p className="text-xs text-muted-foreground">Created: Jan 1, 2026</p>
                  </div>
                </div>
                <Button variant="outline" size="sm">Revoke</Button>
              </div>
              <Button variant="outline" className="mt-4 w-full">
                <Key className="mr-2 h-4 w-4" />
                Generate New API Key
              </Button>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="notifications" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle className="text-lg">Email Settings</CardTitle>
              <CardDescription>Configure email notifications</CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <SettingRow label="New User Registration" description="Notify when a new user registers" checked={flags.notifReg} onCheckedChange={set('notifReg')} />
              <Separator />
              <SettingRow label="Failed Login Attempts" description="Alert on suspicious login activity" checked={flags.notifFail} onCheckedChange={set('notifFail')} />
              <Separator />
              <SettingRow label="System Alerts" description="Receive critical system notifications" checked={flags.notifSys} onCheckedChange={set('notifSys')} />
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
