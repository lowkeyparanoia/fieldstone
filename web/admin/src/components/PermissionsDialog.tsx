import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Check, Save } from 'lucide-react'
import { cn } from '@/lib/utils'
import { MOCK_COLLECTIONS, type AdminUser } from '@/lib/mock'

/**
 * Per-collection access matrix grounded in the BaaS access model: a role plus a
 * matrix over the CRUD verbs the API exposes (list/view/create/update/delete).
 */
const PERM_VERBS = ['List', 'View', 'Create', 'Update', 'Delete'] as const
type Verb = (typeof PERM_VERBS)[number]
type Role = AdminUser['role']

const ROLE_PRESET: Record<Role, Record<Verb, boolean>> = {
  admin: { List: true, View: true, Create: true, Update: true, Delete: true },
  editor: { List: true, View: true, Create: true, Update: true, Delete: false },
  viewer: { List: true, View: true, Create: false, Update: false, Delete: false },
}

export function PermissionsDialog({
  user,
  onClose,
}: {
  user: AdminUser
  onClose: () => void
}) {
  const cols = MOCK_COLLECTIONS.map((c) => c.name)
  const [role, setRole] = useState<Role>(user.role)
  const [verified, setVerified] = useState(user.verified)
  const [active, setActive] = useState(user.status === 'active')
  const [matrix, setMatrix] = useState<Record<string, Record<Verb, boolean>>>(() => {
    const m: Record<string, Record<Verb, boolean>> = {}
    cols.forEach((c) => {
      m[c] = { ...ROLE_PRESET[user.role] }
    })
    return m
  })

  const applyRole = (r: Role) => {
    setRole(r)
    const m: Record<string, Record<Verb, boolean>> = {}
    cols.forEach((c) => {
      m[c] = { ...ROLE_PRESET[r] }
    })
    setMatrix(m)
  }
  const toggle = (c: string, v: Verb) =>
    setMatrix((m) => ({ ...m, [c]: { ...m[c], [v]: !m[c][v] } }))

  return (
    <Dialog open onOpenChange={(o) => !o && onClose()}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Manage Permissions</DialogTitle>
          <DialogDescription>
            Role and per-collection access for {user.email}
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-5">
          <div className="flex flex-wrap items-end gap-4">
            <div className="flex w-[180px] flex-col gap-2">
              <Label>Role</Label>
              <Select value={role} onValueChange={(v) => applyRole(v as Role)}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="editor">Editor</SelectItem>
                  <SelectItem value="viewer">Viewer</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-2.5">
              <Switch checked={verified} onCheckedChange={setVerified} />
              <Label>Verified</Label>
            </div>
            <div className="flex items-center gap-2.5">
              <Switch checked={active} onCheckedChange={setActive} />
              <Label>Active</Label>
            </div>
          </div>

          <div>
            <div className="mb-2 flex items-center gap-2">
              <Label>Collection access</Label>
              <span className="text-xs text-muted-foreground">— overrides from the role preset</span>
            </div>
            <div className="overflow-hidden rounded-md border">
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <th className="border-b px-2.5 py-2 text-left font-mono text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
                      Collection
                    </th>
                    {PERM_VERBS.map((v) => (
                      <th
                        key={v}
                        className="border-b px-2.5 py-2 text-center font-mono text-[11px] font-medium uppercase tracking-wider text-muted-foreground"
                      >
                        {v}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {cols.map((c, i) => (
                    <tr key={c} className={cn(i % 2 === 1 && 'bg-muted/30')}>
                      <td className="border-b px-2.5 py-2 text-left font-mono text-[13px]">{c}</td>
                      {PERM_VERBS.map((v) => (
                        <td key={v} className="border-b px-2.5 py-2 text-center">
                          <button
                            type="button"
                            title={`${v} ${c}`}
                            onClick={() => toggle(c, v)}
                            className={cn(
                              'inline-flex h-[22px] w-[22px] items-center justify-center rounded-sm',
                              matrix[c][v]
                                ? 'bg-primary text-primary-foreground'
                                : 'border border-input'
                            )}
                          >
                            {matrix[c][v] && <Check className="h-3.5 w-3.5" strokeWidth={3.5} />}
                          </button>
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={onClose}>
            <Save className="mr-2 h-4 w-4" />
            Save Permissions
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
