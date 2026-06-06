import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { 
  Plus, 
  Upload, 
  Play, 
  Square, 
  Trash2, 
  Puzzle,
  FileCode,
  AlertTriangle,
  CheckCircle2,
  Clock,
  MemoryStick
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

// Mock WASM plugins data
const mockPlugins = [
  {
    id: '1',
    name: 'Data Validator',
    description: 'Validates incoming data against custom rules',
    version: '1.0.0',
    status: 'active',
    language: 'rust',
    size: 156000,
    memory: 64,
    events: ['record.create', 'record.update'],
    lastExecuted: '2024-01-15T10:30:00Z',
  },
  {
    id: '2',
    name: 'Email Processor',
    description: 'Processes and sanitizes email content',
    version: '0.5.2',
    status: 'active',
    language: 'go',
    size: 89000,
    memory: 32,
    events: ['record.create'],
    lastExecuted: '2024-01-15T09:15:00Z',
  },
  {
    id: '3',
    name: 'Image Optimizer',
    description: 'Optimizes uploaded images for web',
    version: '2.1.0',
    status: 'inactive',
    language: 'rust',
    size: 256000,
    memory: 128,
    events: ['file.upload'],
    lastExecuted: '2024-01-14T16:45:00Z',
  },
]

export function PluginsPage() {
  const [plugins, setPlugins] = useState(mockPlugins)
  const [activeTab, setActiveTab] = useState('installed')

  const togglePlugin = (id: string) => {
    setPlugins(plugins.map(p => 
      p.id === id ? { ...p, status: p.status === 'active' ? 'inactive' : 'active' } : p
    ))
  }

  const deletePlugin = (id: string) => {
    setPlugins(plugins.filter(p => p.id !== id))
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Plugins</h1>
          <p className="text-muted-foreground">
            Manage WASM plugins for custom business logic
          </p>
        </div>
        <Dialog>
          <DialogTrigger asChild>
            <Button>
              <Plus className="mr-2 h-4 w-4" />
              Install Plugin
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[525px]">
            <DialogHeader>
              <DialogTitle>Install Plugin</DialogTitle>
              <DialogDescription>
                Upload a WASM plugin file to extend Fieldstone functionality.
              </DialogDescription>
            </DialogHeader>
            <div className="grid gap-4 py-4">
              <div className="border-2 border-dashed border-muted-foreground/25 rounded-lg p-8 text-center">
                <Upload className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
                <p className="text-sm text-muted-foreground">
                  Drag and drop your .wasm file here, or click to browse
                </p>
                <Input type="file" accept=".wasm" className="hidden" />
                <Button variant="outline" size="sm" className="mt-2">
                  Browse Files
                </Button>
              </div>
              <div className="grid gap-2">
                <label>Plugin Name</label>
                <Input placeholder="e.g., Custom Validator" />
              </div>
              <div className="grid gap-2">
                <label>Memory Limit (MB)</label>
                <Select defaultValue="64">
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="32">32 MB</SelectItem>
                    <SelectItem value="64">64 MB</SelectItem>
                    <SelectItem value="128">128 MB</SelectItem>
                    <SelectItem value="256">256 MB</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button type="submit">Install Plugin</Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Alert>
        <AlertTriangle className="h-4 w-4" />
        <AlertTitle>Sandboxed Environment</AlertTitle>
        <AlertDescription>
          All plugins run in a sandboxed WebAssembly environment with restricted access to system resources.
        </AlertDescription>
      </Alert>

      <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-4">
        <TabsList>
          <TabsTrigger value="installed">Installed</TabsTrigger>
          <TabsTrigger value="marketplace">Marketplace</TabsTrigger>
          <TabsTrigger value="logs">Execution Logs</TabsTrigger>
        </TabsList>

        <TabsContent value="installed" className="space-y-4">
          <div className="grid gap-4">
            {plugins.map((plugin) => (
              <Card key={plugin.id}>
                <CardContent className="p-6">
                  <div className="flex items-start justify-between">
                    <div className="flex gap-4">
                      <div className="h-12 w-12 rounded-lg bg-primary/10 flex items-center justify-center">
                        <Puzzle className="h-6 w-6 text-primary" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold">{plugin.name}</h3>
                          <Badge variant="outline">v{plugin.version}</Badge>
                          <Badge variant={plugin.status === 'active' ? 'success' : 'secondary'}>
                            {plugin.status}
                          </Badge>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                          {plugin.description}
                        </p>
                        <div className="flex items-center gap-4 mt-3 text-xs text-muted-foreground">
                          <span className="flex items-center gap-1">
                            <FileCode className="h-3 w-3" />
                            {plugin.language}
                          </span>
                          <span className="flex items-center gap-1">
                            <MemoryStick className="h-3 w-3" />
                            {plugin.memory}MB limit
                          </span>
                          <span className="flex items-center gap-1">
                            <Clock className="h-3 w-3" />
                            Last run: {new Date(plugin.lastExecuted).toLocaleTimeString()}
                          </span>
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => togglePlugin(plugin.id)}
                      >
                        {plugin.status === 'active' ? (
                          <Square className="mr-1 h-3 w-3" />
                        ) : (
                          <Play className="mr-1 h-3 w-3" />
                        )}
                        {plugin.status === 'active' ? 'Stop' : 'Start'}
                      </Button>
                      <Button
                        variant="outline"
                        size="icon"
                        className="text-destructive hover:bg-destructive/10"
                        onClick={() => deletePlugin(plugin.id)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {plugins.length === 0 && (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Puzzle className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-muted-foreground">No plugins installed</p>
                <Button className="mt-4" variant="outline">
                  <Plus className="mr-2 h-4 w-4" />
                  Install your first plugin
                </Button>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        <TabsContent value="marketplace">
          <Card>
            <CardHeader>
              <CardTitle>Plugin Marketplace</CardTitle>
              <CardDescription>
                Discover and install plugins from the community
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Data Sanitizer</CardTitle>
                    <CardDescription>Clean and normalize user input</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Badge variant="secondary" className="mb-2">Free</Badge>
                    <Button size="sm" className="w-full">Install</Button>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">PDF Generator</CardTitle>
                    <CardDescription>Generate PDFs from templates</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Badge variant="secondary" className="mb-2">Free</Badge>
                    <Button size="sm" className="w-full">Install</Button>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle className="text-base">Email Validator</CardTitle>
                    <CardDescription>Advanced email validation</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Badge variant="secondary" className="mb-2">Free</Badge>
                    <Button size="sm" className="w-full">Install</Button>
                  </CardContent>
                </Card>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="logs">
          <Card>
            <CardHeader>
              <CardTitle>Execution Logs</CardTitle>
              <CardDescription>
                Recent plugin executions and their results
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-500" />
                  <div>
                    <p className="font-medium text-sm">Data Validator</p>
                    <p className="text-xs text-muted-foreground">Execution completed in 12ms</p>
                  </div>
                </div>
                <span className="text-xs text-muted-foreground">
                  2 minutes ago
                </span>
              </div>
              <div className="flex items-center justify-between p-3 border rounded-lg">
                <div className="flex items-center gap-3">
                  <CheckCircle2 className="h-5 w-5 text-emerald-500" />
                  <div>
                    <p className="font-medium text-sm">Email Processor</p>
                    <p className="text-xs text-muted-foreground">Execution completed in 8ms</p>
                  </div>
                </div>
                <span className="text-xs text-muted-foreground">
                  5 minutes ago
                </span>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
